/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

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

// Package centralizedstorage is related with storage client
package centralizedstorage

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/huawei/csm/v2/storage/client"
	"github.com/huawei/csm/v2/storage/constant"
	"github.com/huawei/csm/v2/storage/utils"
	"github.com/huawei/csm/v2/utils/log"
	"github.com/huawei/csm/v2/utils/resource"
)

const defaultTimeout = 60 * time.Second

// CentralizedClient is used to use centralized storage related functions
type CentralizedClient struct {
	client.Client
}

// NewCentralizedClient is used to new centralized storage client
func NewCentralizedClient(ctx context.Context, config *constant.StorageBackendConfig) (*CentralizedClient, error) {
	centralizedClient := &CentralizedClient{
		Client: client.Client{
			Urls:                    config.Urls,
			User:                    config.User,
			SecretNamespace:         config.SecretNamespace,
			SecretName:              config.SecretName,
			StorageBackendNamespace: config.StorageBackendNamespace,
			StorageBackendName:      config.StorageBackendName,
			Client:                  newHttpClient(),
			Semaphore:               utils.NewSemaphore(config.ClientMaxThreads),
		},
	}
	if err := centralizedClient.initHttpClient(ctx); err != nil {
		return nil, err
	}

	return centralizedClient, nil
}

func newHttpClient() *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Warningf("storage client init http client fail, error: %v", err)
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Jar:     jar,
		Timeout: defaultTimeout,
	}
}

func (c *CentralizedClient) initHttpClient(ctx context.Context) error {
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

	c.Client.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tlsConfig,
		},
		Jar:     jar,
		Timeout: defaultTimeout,
	}

	log.AddContext(ctx).Infof("init http client success, skip verify certificate: %v", skipVerify)
	return nil
}

func (c *CentralizedClient) getTlsCertConfig(ctx context.Context) (*x509.CertPool, bool, error) {
	useCert, certSecret, err := c.getCertParametersFromSbcDynamically(ctx)
	if err != nil {
		log.AddContext(ctx).Errorf("get cert parameters from sbc error: %v", err)
		return nil, true, err
	}

	// judge to skip certificate verification
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

func (c *CentralizedClient) getCertPool(ctx context.Context, secret *v1.Secret) (*x509.CertPool, error) {
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
