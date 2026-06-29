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

package centralizedstorage

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	coreV1 "k8s.io/api/core/v1"

	"github.com/huawei/csm/v2/storage/client"
	"github.com/huawei/csm/v2/storage/utils"
	"github.com/huawei/csm/v2/utils/resource"
)

func TestLoginThenSuccess(t *testing.T) {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code": float64(0),
		},
		"data": map[string]interface{}{
			"deviceid":     "1",
			"iBaseToken":   "2",
			"accountstate": float64(1),
		},
	}
	var cli *client.Client
	call := gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return response, nil
		})
	defer call.Reset()

	secret := &coreV1.Secret{
		Data: map[string][]byte{
			passwordKey:           []byte{'1'},
			authenticationModeKey: []byte("1"),
		},
	}
	var coreCli *resource.Client
	getSecret := gomonkey.ApplyMethod(reflect.TypeOf(coreCli), "GetSecret",
		func(_ *resource.Client, name string, namespace string) (*coreV1.Secret, error) {
			return secret, nil
		})
	defer getSecret.Reset()

	centralizedCli := &CentralizedClient{
		Client: client.Client{
			Semaphore: utils.NewSemaphore(3),
		},
	}
	centralizedCli.Urls = []string{"url"}

	err := centralizedCli.Login(ctx)
	if err != nil {
		t.Errorf("Login() error: %v", err)
	}
}

func TestLoginWhenUnConnectedThenFailed(t *testing.T) {
	var cli *client.Client
	httpGet := gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return nil, errors.New("unconnected")
		})
	defer httpGet.Reset()

	secret := &coreV1.Secret{
		Data: map[string][]byte{
			"password": []byte{'1'},
		},
	}
	var coreCli *resource.Client
	getSecret := gomonkey.ApplyMethod(reflect.TypeOf(coreCli), "GetSecret",
		func(_ *resource.Client, name string, namespace string) (*coreV1.Secret, error) {
			return secret, nil
		})
	defer getSecret.Reset()

	centralizedCli := &CentralizedClient{
		Client: client.Client{
			Semaphore: utils.NewSemaphore(3),
		},
	}
	centralizedCli.Urls = []string{"url"}

	expectError := errors.New("storage client all url connect error")
	actualError := centralizedCli.Login(ctx)

	if actualError == nil || actualError.Error() != expectError.Error() {
		t.Errorf("Login() error: expect error: %v, actual error: %v", expectError, actualError)
	}
}

func TestLoginWhenIBaseTokenNotExistThenFailed(t *testing.T) {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code": float64(0),
		},
		"data": map[string]interface{}{
			"deviceid":     "1",
			"accountstate": float64(1),
		},
	}

	var cli *client.Client
	httpGet := gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return response, nil
		})
	defer httpGet.Reset()

	secret := &coreV1.Secret{
		Data: map[string][]byte{
			"password": []byte{'1'},
		},
	}
	var coreCli *resource.Client
	getSecret := gomonkey.ApplyMethod(reflect.TypeOf(coreCli), "GetSecret",
		func(_ *resource.Client, name string, namespace string) (*coreV1.Secret, error) {
			return secret, nil
		})
	defer getSecret.Reset()

	centralizedCli := &CentralizedClient{
		Client: client.Client{
			Semaphore: utils.NewSemaphore(3),
		},
	}
	centralizedCli.Urls = []string{"url"}

	expectErr := fmt.Errorf(
		"storage client login response iBaseToken can not convert to string")
	actualErr := centralizedCli.Login(ctx)
	if actualErr == nil || expectErr.Error() != actualErr.Error() {
		t.Errorf("Login() error, expect error: %v, actual error: %v", expectErr, actualErr)
	}
}

func TestLogoutThenSuccess(t *testing.T) {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code": float64(0),
		},
	}

	var cli *client.Client
	httpGet := gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return response, nil
		})
	defer httpGet.Reset()

	centralizedCli := &CentralizedClient{
		Client: client.Client{
			Semaphore: utils.NewSemaphore(3),
		},
	}
	centralizedCli.Logout(ctx)
}

func TestLogoutCheckResponseCodeError(t *testing.T) {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code": float64(1),
		},
	}

	var cli *client.Client
	httpGet := gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return response, nil
		})
	defer httpGet.Reset()

	centralizedCli := &CentralizedClient{
		Client: client.Client{
			Semaphore: utils.NewSemaphore(3),
		},
	}
	centralizedCli.Logout(ctx)
}

func TestLogin_GetBackendLoginParamsError(t *testing.T) {
	var coreCli *resource.Client
	getSecret := gomonkey.ApplyMethod(reflect.TypeOf(coreCli), "GetSecret",
		func(_ *resource.Client, name string, namespace string) (*coreV1.Secret, error) {
			return nil, errors.New("get secret error")
		})
	defer getSecret.Reset()

	centralizedCli := &CentralizedClient{
		Client: client.Client{
			Semaphore:       utils.NewSemaphore(3),
			SecretName:      "test",
			SecretNamespace: "test",
		},
	}

	err := centralizedCli.Login(ctx)
	if err == nil {
		t.Errorf("Login() expected error")
	}
}
