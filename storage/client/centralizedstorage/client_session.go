/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2025. All rights reserved.

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
	"errors"
	"fmt"
	"strings"

	coreV1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/huawei/csm/v2/storage/constant"
	"github.com/huawei/csm/v2/utils/log"
	"github.com/huawei/csm/v2/utils/resource"
)

const (
	authenticationModeKey = "authenticationMode"
	passwordKey           = "password"
	authModeScopeLocal    = "0"
)

// backendLoginParams for login backend
type backendLoginParams struct {
	// password for log in backend
	password []byte
	// authentication, local:0, ldap:1
	scope string
}

// Login is used to log in storage client
func (c *CentralizedClient) Login(ctx context.Context) error {
	log.AddContext(ctx).Infof("storage client login start, urls: %v", c.Urls)
	params, err := c.getBackendLoginParamsFromSecret(ctx)
	if err != nil {
		log.AddContext(ctx).Errorf("get BackendLoginParams failed, err: %w", err)
		return err
	}

	reqData := map[string]interface{}{
		"username": c.User,
		"password": string(params.password),
		"scope":    params.scope,
	}

	for i := range params.password {
		params.password[i] = 0
	}

	resp, err := c.loginCall(ctx, reqData)
	reqData[passwordKey] = ""
	if err != nil {
		log.AddContext(ctx).Errorf("storage client login error: %v", err)
		return err
	}

	respData, _, err := c.getResultFromResponse(ctx, resp)
	if err != nil {
		return err
	}

	if err = c.checkLoginAccountState(ctx, respData); err != nil {
		return err
	}

	if err = c.setClientWithLoginResponseData(ctx, respData); err != nil {
		log.AddContext(ctx).Errorf("storage client login set client error: %v", err)
		return err
	}

	log.AddContext(ctx).Infof("storage client login success, url: %s", c.Curl)
	return nil
}

// ReLogin is used to reLogin storage client
func (c *CentralizedClient) ReLogin(ctx context.Context) error {
	log.AddContext(ctx).Infof("storage client reLogin start...")
	defer log.AddContext(ctx).Infof("storage client reLogin success...")

	oldToken := c.Token

	c.ReLoginMutex.Lock()
	defer c.ReLoginMutex.Unlock()
	if c.Token != "" && oldToken != c.Token {
		// other thread had already done relogin, so no need to relogin again
		return nil
	}

	c.Logout(ctx)
	err := c.Login(ctx)
	if err != nil {
		log.AddContext(ctx).Errorf("storage client try to relogin error: %v", err)
		return err
	}

	return nil
}

// Logout is used to logout storage client
func (c *CentralizedClient) Logout(ctx context.Context) {
	log.AddContext(ctx).Infof("storage client logout start...")
	defer log.AddContext(ctx).Infof("storage client logout success...")

	resp, err := c.delete(ctx, "/sessions", nil)
	if err != nil {
		log.AddContext(ctx).Errorf("storage client logout %s error: %v", c.Curl, err)
		return
	}

	_, err = c.checkResponseCode(ctx, resp)
	if err != nil {
		log.AddContext(ctx).Errorf("storage client logout %s error: %v", c.Curl, err)
		return
	}

	log.AddContext(ctx).Infof("storage client logout %s success", c.Curl)
}

func (c *CentralizedClient) setClientWithLoginResponseData(ctx context.Context, respData map[string]interface{}) error {
	var exist bool
	c.DeviceId, exist = respData["deviceid"].(string)
	if !exist {
		msg := fmt.Sprintf(
			"storage client login response deviceid: %v can not convert to string", respData["deviceid"])
		log.AddContext(ctx).Errorln(msg)
		return errors.New(msg)
	}

	c.Token, exist = respData["iBaseToken"].(string)
	if !exist {
		msg := fmt.Sprintf(
			"storage client login response iBaseToken can not convert to string")
		log.AddContext(ctx).Errorln(msg)
		return errors.New(msg)
	}

	c.VStore, exist = respData["vstoreName"].(string)
	if !exist {
		log.AddContext(ctx).Infof(
			"storage client login response vstoreName: %v can not convert to string", respData["vstoreName"])
	}

	return nil
}

func (c *CentralizedClient) loginCall(ctx context.Context, reqData map[string]interface{}) (*Response, error) {
	for _, url := range c.Urls {
		c.Curl = url + "/deviceManager/rest"
		log.AddContext(ctx).Infof("storage client try to login: %s", c.Curl)
		resp, err := c.post(ctx, "/xx/sessions", reqData)
		if err == nil {
			return resp, err
		}

		log.AddContext(ctx).Infof("storage client %s login error, going to try another url", c.Curl)
	}

	return nil, errors.New("storage client all url connect error")
}

// getPasswordFromSecret is used to get password and authMode from secret
func (c *CentralizedClient) getBackendLoginParamsFromSecret(ctx context.Context) (*backendLoginParams, error) {
	secret, err := resource.Instance().GetSecret(c.SecretName, c.SecretNamespace)
	if err != nil && !apiErrors.IsNotFound(err) {
		return nil, fmt.Errorf("storage client get secret with name %s and namespace %s failed, error: %w",
			c.SecretName, c.SecretNamespace, err)
	}

	// when the sbc change the password by using oceanctl, the secret of sbc will be changed.
	// in this case, need to get the latest secret from sbc.
	if apiErrors.IsNotFound(err) {
		log.AddContext(ctx).Infof("secret [%s/%s] not found, try to get new one from sbc dynamically",
			c.SecretNamespace, c.SecretName)
		secret, err = c.getSecretFromSbcDynamically(ctx)
		if err != nil {
			return nil, fmt.Errorf("get secret from sbc dynamiclly failed, error is [%w]", err)
		}
		log.AddContext(ctx).Infof("get secret [%s/%s] from sbc dynamically", secret.Namespace, secret.Name)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret is nil or the data not exist in secret, namespace: %s, secret name: %s",
			c.SecretName, c.SecretNamespace)
	}

	password, exist := secret.Data[passwordKey]
	if !exist {
		return nil, fmt.Errorf("failed to query the password, namespace: %s, secret name: %s",
			c.SecretName, c.SecretNamespace)
	}

	scope := authModeScopeLocal
	authMode, exist := secret.Data[authenticationModeKey]
	if exist {
		scope = string(authMode)
	}

	return &backendLoginParams{password: password, scope: scope}, nil
}

func (c *CentralizedClient) getSecretFromSbcDynamically(ctx context.Context) (*coreV1.Secret, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("getting cluster config error, error is [%v]", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)

	gvr := schema.GroupVersionResource{
		Group:    "xuanwu.huawei.io",
		Version:  "v1",
		Resource: "storagebackendclaims",
	}
	unstructuredResource, err := dynamicClient.Resource(gvr).Namespace(c.StorageBackendNamespace).
		Get(context.TODO(), c.StorageBackendName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get unstructuredResource of sbc [%s/%s] failed, "+
			"error is [%v]", c.StorageBackendNamespace, c.StorageBackendName, err)
	}

	secretMeta, found, err := unstructured.NestedString(
		unstructuredResource.UnstructuredContent(), "spec", "secretMeta")
	if !found || err != nil {
		return nil, fmt.Errorf("get secret meta from sbc [%s/%s] failed, "+
			"error is [%v]", c.StorageBackendNamespace, c.StorageBackendName, err)
	}

	// secretMeta format is <namespace>/<name>
	secretNameSpace := strings.Split(secretMeta, "/")[0]
	secretName := strings.Split(secretMeta, "/")[1]
	return resource.Instance().GetSecret(secretName, secretNameSpace)
}

func (c *CentralizedClient) getCertParametersFromSbcDynamically(ctx context.Context) (bool, string, error) {
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
		Get(context.TODO(), c.StorageBackendName, metav1.GetOptions{})
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

func (c *CentralizedClient) checkLoginAccountState(ctx context.Context, respData map[string]interface{}) error {
	accountState, exist := respData["accountstate"].(float64)
	if !exist {
		msg := fmt.Sprintf("login response accountstate: %v can not convert to float64",
			respData["accountstate"])
		log.AddContext(ctx).Errorln(msg)
		return errors.New(msg)
	}

	// check accountstate
	if float64(constant.LoginNormal) == accountState ||
		float64(constant.LoginPasswordIsAboutToExpire) == accountState ||
		float64(constant.NextLoginPasswordMustBeChanged) == accountState ||
		float64(constant.LoginPasswordNeverExpires) == accountState {
		log.AddContext(ctx).Infof("login valid accountstate: %s", constant.LoginAccountStateMap[accountState])
		return nil
	}

	msg := fmt.Sprintf("login invalid accountstate: %s", constant.LoginAccountStateMap[accountState])
	log.AddContext(ctx).Errorln(msg)
	return errors.New(msg)
}
