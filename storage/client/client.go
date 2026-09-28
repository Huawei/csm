/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2024. All rights reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

// Package client is related with storage common client and operation
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/huawei/csm/v2/storage/constant"
	"github.com/huawei/csm/v2/storage/utils"
	"github.com/huawei/csm/v2/utils/log"
	"github.com/huawei/csm/v2/utils/resource"
)

const (
	// 20000 is the approximate characters number of one page
	// If a log exceeds one page, it will be compressed
	charLimit = 20000

	sessionsSubStr     = "/sessions"
	defaultHTTPTimeout = 60 * time.Second
	secretMetaLen      = 2
)

// Client is used to extract storage common attribute
type Client struct {
	Curl     string
	Urls     []string
	User     string
	DeviceId string
	Token    string
	VStore   string
	Client   HttpClient

	SecretNamespace string
	SecretName      string

	StorageBackendNamespace string
	StorageBackendName      string

	ReLoginMutex sync.Mutex
	Semaphore    *utils.Semaphore

	// SetAuthHeaders sets authentication headers on the request.
	// Must be set by the embedding client (e.g., iBaseToken for OceanStor, X-Auth-Token for FusionStorage).
	SetAuthHeaders func(req *http.Request)
}

// HttpClient is used to define http interface
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Call is used to remote call storage interfaces
func (c *Client) Call(ctx context.Context, method string, url string,
	reqData map[string]interface{}) (map[string]interface{}, error) {
	if c.Semaphore != nil {
		c.Semaphore.Acquire()
		defer c.Semaphore.Release()
	}

	if !strings.Contains(url, sessionsSubStr) {
		log.AddContext(ctx).Infof("call request %s %s, request: %v", method, url, reqData)
	}
	log.AddContext(ctx).Infof("call reloginLock: %v", c.ReLoginMutex)

	req, err := c.getRequest(ctx, method, url, reqData)
	if err != nil {
		log.AddContext(ctx).Errorf(
			"client http request error, method: %s, url: %s, error: %v", method, url, err)
		return nil, err
	}

	resp, err := c.getResponse(ctx, req)
	if err != nil {
		log.AddContext(ctx).Errorf("client http response error, method: %s, url: %s, error: %v",
			method, url, err)
		return nil, err
	}

	if !strings.Contains(url, sessionsSubStr) {
		responseStr := fmt.Sprintf("%v", resp)
		if len(responseStr) > charLimit {
			compressedStr, err := utils.CompressStr(responseStr)
			if err != nil {
				log.AddContext(ctx).Warningf("compress storage response fail, "+
					"the log will be printed without compression, err: %v", err)
			}
			log.AddContext(ctx).Infof("call response %s %s, response compressed by deflate algorithm: %s",
				method, url, compressedStr)
			log.AddContext(ctx).Debugf("call response %s %s, response: %s", method, url, responseStr)
		} else {
			log.AddContext(ctx).Infof("call response %s %s, response: %s", method, url, responseStr)
		}
	}
	return resp, nil
}

// RetryCall is used to retry remote call storage interfaces
func (c *Client) RetryCall(ctx context.Context, retryCodes []float64,
	call func() (map[string]interface{}, *float64, error)) (map[string]interface{}, error) {
	var err error
	var code *float64
	var respData map[string]interface{}

	retryFunc := func() bool {
		respData, code, err = call()
		// if code not exist, then do not retry
		if code == nil {
			return false
		}

		if err != nil {
			if !utils.IsFloat64InList(retryCodes, *code) {
				return false
			} else {
				log.AddContext(ctx).Infoln("storage client retry call...")
				return true
			}
		}

		return false
	}

	utils.RetryCallFunc(retryFunc)

	return respData, err
}

// RetryListCall is used to retry remote call storage interfaces
func (c *Client) RetryListCall(ctx context.Context, retryCodes []float64,
	call func() ([]map[string]interface{}, *float64, error)) ([]map[string]interface{}, error) {
	var err error
	var code *float64
	var respData []map[string]interface{}

	retryFunc := func() bool {
		respData, code, err = call()
		// if code not exist, then do not retry
		if code == nil {
			return false
		}

		if err != nil {
			if !utils.IsFloat64InList(retryCodes, *code) {
				return false
			} else {
				log.AddContext(ctx).Infoln("storage client retry list call...")
				return true
			}
		}

		return false
	}

	utils.RetryCallFunc(retryFunc)

	return respData, err
}

func (c *Client) getRequest(ctx context.Context, method string, url string,
	reqData map[string]interface{}) (*http.Request, error) {
	log.AddContext(ctx).Debugln("get request start...")
	defer log.AddContext(ctx).Debugln("get request end...")

	reqBody, err := c.getRequestBody(ctx, url, reqData)

	if err != nil {
		log.AddContext(ctx).Errorf("client http request body error, url: %s, error: %s", url, err.Error())
		return nil, err
	}

	req, err := c.newRequest(ctx, method, url, reqBody)
	if err != nil {
		log.AddContext(ctx).Errorf("client http new request error, url: %s, error: %s", url, err.Error())
		return nil, err
	}

	return req, nil
}

func (c *Client) getResponse(ctx context.Context, req *http.Request) (map[string]interface{}, error) {
	log.AddContext(ctx).Debugln("get response start...")
	defer log.AddContext(ctx).Debugln("get response end...")

	clientResp, err := c.Client.Do(req)
	if err != nil {
		log.AddContext(ctx).Errorf("client http response error, error: %v", err)
		return nil, err
	}
	defer clientResp.Body.Close()

	log.AddContext(ctx).Debugln("start read response body...")
	body, err := ioutil.ReadAll(clientResp.Body)
	if err != nil {
		log.AddContext(ctx).Errorf("client read response body error: %v", err)
		return nil, err
	}
	log.AddContext(ctx).Debugln("read response body success...")

	var resp map[string]interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.AddContext(ctx).Errorf("client response body convert response error, body: %s, error: %v", body, err)
		return nil, err
	}

	return resp, nil
}

func (c *Client) getRequestBody(ctx context.Context, url string, reqData map[string]interface{}) (io.Reader, error) {
	log.AddContext(ctx).Debugln("get request body start...")
	defer log.AddContext(ctx).Debugln("get request body end...")

	if reqData == nil {
		return nil, nil
	}

	reqBytes, err := json.Marshal(reqData)
	if err != nil {
		if strings.Contains(url, sessionsSubStr) {
			log.AddContext(ctx).Errorf("client http request body error: %v", err)
		} else {
			log.AddContext(ctx).Errorf("client http request body error, data: %v, error: %v", reqData, err)
		}

		return nil, err
	}

	return bytes.NewReader(reqBytes), nil
}

func (c *Client) newRequest(ctx context.Context, method string, reqUrl string,
	reqBody io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqUrl, reqBody)
	if err != nil {
		log.AddContext(ctx).Errorf("client http new request error: %s", err.Error())
		return req, err
	}

	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")

	if c.SetAuthHeaders != nil {
		c.SetAuthHeaders(req)
	} else if c.Token != "" {
		log.AddContext(ctx).Warningf("SetAuthHeaders is nil but Token is not empty, " +
			"authentication headers will not be set")
	}

	return req, nil
}

// InitHttpClient initializes the HTTP client with TLS certificate configuration.
// It reads certificate settings from the StorageBackendClaim CRD and configures
// the HTTP client accordingly. If no certificate is configured, it falls back
// to skipping TLS verification.
func (c *Client) InitHttpClient(ctx context.Context) error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.AddContext(ctx).Errorf("init http client cookiejar fail, error: %v", err)
		return err
	}

	certPool, skipVerify, err := c.getTlsCertConfig(ctx)
	if err != nil {
		return err
	}

	tlsConfig := tls.Config{
		InsecureSkipVerify: skipVerify,
	}

	if certPool != nil {
		tlsConfig.RootCAs = certPool
	}

	c.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tlsConfig,
		},
		Jar:     jar,
		Timeout: defaultHTTPTimeout,
	}

	log.AddContext(ctx).Infof("init http client success, skip verify certificate: %v", skipVerify)
	return nil
}

// getTlsCertConfig determines the TLS configuration by reading the
// StorageBackendClaim CRD. Returns the cert pool, whether to skip verification,
// and any error encountered.
func (c *Client) getTlsCertConfig(ctx context.Context) (*x509.CertPool, bool, error) {
	useCert, certSecret, err := c.getCertParametersFromSbcDynamically(ctx)
	if err != nil {
		log.AddContext(ctx).Errorf("get cert parameters from sbc error: %v", err)
		return nil, true, err
	}

	if !useCert {
		return nil, true, nil
	}

	// certSecret format is <namespace>/<name>
	certSecretNameSpace, certSecretName, err := cache.SplitMetaNamespaceKey(certSecret)
	if err != nil {
		log.AddContext(ctx).Errorf("split cert secret error: %v", err)
		return nil, true, err
	}

	secret, err := resource.Instance().GetSecret(certSecretName, certSecretNameSpace)
	if err != nil {
		log.AddContext(ctx).Errorf("get cert secret error: %v", err)
		return nil, true, err
	}

	certPool, err := c.getCertPool(ctx, secret)
	if err != nil {
		log.AddContext(ctx).Errorf("get certificate error: %v", err)
		return nil, true, err
	}

	return certPool, false, nil
}

// getCertPool parses the TLS certificate from a Kubernetes Secret and
// returns an x509.CertPool containing the certificate.
func (c *Client) getCertPool(ctx context.Context, secret *v1.Secret) (*x509.CertPool, error) {
	log.AddContext(ctx).Infof("start get cert from secret %s/%s", secret.Namespace, secret.Name)
	defer log.AddContext(ctx).Infof("end get cert from secret %s/%s", secret.Namespace, secret.Name)

	certData, exist := secret.Data[constant.CertificateKeyName]
	if !exist {
		msg := fmt.Sprintf("certificate not config in secret %s/%s", secret.Namespace, secret.Name)
		log.AddContext(ctx).Errorln(msg)
		return nil, errors.New(msg)
	}

	certBlock, _ := pem.Decode(certData)
	if certBlock == nil {
		msg := fmt.Sprintf("certificate data decode error in secret %s/%s", secret.Namespace, secret.Name)
		log.AddContext(ctx).Errorln(msg)
		return nil, errors.New(msg)
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		log.AddContext(ctx).Errorf("error parse certificate: %v", err)
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(cert)
	return certPool, nil
}

// getCertParametersFromSbcDynamically reads the StorageBackendClaim CRD to
// determine whether TLS certificate verification is enabled and obtains the
// certificate secret reference. Returns (useCert, certSecret, error).
func (c *Client) getCertParametersFromSbcDynamically(ctx context.Context) (bool, string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return false, "", fmt.Errorf("getting cluster config error, error is [%v]", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return false, "", fmt.Errorf("getting dynamicClient error, error is [%v]", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "xuanwu.huawei.io",
		Version:  "v1",
		Resource: "storagebackendclaims",
	}
	unstructuredResource, err := dynamicClient.Resource(gvr).Namespace(c.StorageBackendNamespace).
		Get(ctx, c.StorageBackendName, metav1.GetOptions{})
	if err != nil {
		return false, "", fmt.Errorf("get unstructuredResource of sbc [%s/%s] failed, "+
			"error is [%v]", c.StorageBackendNamespace, c.StorageBackendName, err)
	}

	useCert, found, err := unstructured.NestedBool(
		unstructuredResource.UnstructuredContent(), "spec", "useCert")
	if err != nil {
		return false, "", fmt.Errorf("get isUseCert parameter from sbc [%s/%s] failed, "+
			"error is [%v]", c.StorageBackendNamespace, c.StorageBackendName, err)
	}
	if !found {
		log.AddContext(ctx).Infof("useCert is not found, skip the cert")
		return false, "", nil
	}

	if !useCert {
		log.AddContext(ctx).Infof("useCert is false, skip the cert")
		return false, "", nil
	}

	certSecret, found, err := unstructured.NestedString(
		unstructuredResource.UnstructuredContent(), "spec", "certSecret")
	if err != nil {
		return false, "", fmt.Errorf("get certSecret parameter from sbc [%s/%s] failed, "+
			"error is [%v]", c.StorageBackendNamespace, c.StorageBackendName, err)
	}
	if !found {
		return false, "", fmt.Errorf("get certSecret parameter from sbc [%s/%s] failed, "+
			"certSecret parameter is not found", c.StorageBackendNamespace, c.StorageBackendName)
	}

	return true, certSecret, nil
}

// GetSecretWithFallback reads the K8s Secret specified by c.SecretNamespace/c.SecretName.
// If the Secret is not found, it falls back to reading the secret reference from
// the StorageBackendClaim CRD dynamically. Returns the full Secret so callers can
// extract multiple fields (e.g., password, authenticationMode).
func (c *Client) GetSecretWithFallback(ctx context.Context) (*v1.Secret, error) {
	secret, err := resource.Instance().GetSecret(c.SecretName, c.SecretNamespace)
	if err != nil && !apiErrors.IsNotFound(err) {
		return nil, fmt.Errorf("get secret [%s/%s] failed: %w", c.SecretNamespace, c.SecretName, err)
	}

	if apiErrors.IsNotFound(err) {
		log.AddContext(ctx).Infof("secret [%s/%s] not found, try to get from sbc dynamically",
			c.SecretNamespace, c.SecretName)
		secret, err = c.getSecretFromSbcDynamically(ctx)
		if err != nil {
			return nil, fmt.Errorf("get secret from sbc dynamically failed: %w", err)
		}
		log.AddContext(ctx).Infof("get secret [%s/%s] from sbc dynamically", secret.Namespace, secret.Name)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret is nil or has no data, namespace: %s, name: %s",
			c.SecretNamespace, c.SecretName)
	}

	return secret, nil
}

// getSecretFromSbcDynamically reads the StorageBackendClaim CRD to find the
// secret reference (spec.secretMeta), then fetches the actual K8s Secret.
func (c *Client) getSecretFromSbcDynamically(ctx context.Context) (*v1.Secret, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("getting cluster config error: %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("getting dynamicClient error: %v", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "xuanwu.huawei.io",
		Version:  "v1",
		Resource: "storagebackendclaims",
	}
	unstructuredResource, err := dynamicClient.Resource(gvr).Namespace(c.StorageBackendNamespace).
		Get(ctx, c.StorageBackendName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get unstructuredResource of sbc [%s/%s] failed: %v",
			c.StorageBackendNamespace, c.StorageBackendName, err)
	}

	secretMeta, found, err := unstructured.NestedString(
		unstructuredResource.UnstructuredContent(), "spec", "secretMeta")
	if !found || err != nil {
		return nil, fmt.Errorf("get secretMeta from sbc [%s/%s] failed: %v",
			c.StorageBackendNamespace, c.StorageBackendName, err)
	}

	// secretMeta format is <namespace>/<name>
	parts := strings.SplitN(secretMeta, "/", secretMetaLen)
	if len(parts) != secretMetaLen {
		return nil, fmt.Errorf("invalid secretMeta format %q from sbc [%s/%s]",
			secretMeta, c.StorageBackendNamespace, c.StorageBackendName)
	}
	return resource.Instance().GetSecret(parts[1], parts[0])
}
