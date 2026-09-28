/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package distributedstorage

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	apiDistributed "github.com/huawei/csm/v2/storage/api/distributedstorage"
	"github.com/huawei/csm/v2/storage/client"
	"github.com/huawei/csm/v2/storage/constant"
)

func TestNormalizeFieldNames(t *testing.T) {
	// arrange
	input := map[string]interface{}{
		"id":               35,
		"name":             "ns1",
		"space_hard_quota": 10485760,
	}
	// act
	result := normalizeFieldNames(input)
	// assert
	assert.Equal(t, 35, result["ID"])
	assert.Equal(t, "ns1", result["NAME"])
	assert.Equal(t, 10485760, result["SPACE_HARD_QUOTA"])
}

func TestNormalizeFieldNamesSlice(t *testing.T) {
	// arrange
	input := []map[string]interface{}{
		{"id": 35, "name": "ns1"},
		{"id": 36, "name": "ns2"},
	}
	// act
	result := normalizeFieldNamesSlice(input)
	// assert
	assert.Equal(t, 35, result[0]["ID"])
	assert.Equal(t, "ns1", result[0]["NAME"])
	assert.Equal(t, 36, result[1]["ID"])
	assert.Equal(t, "ns2", result[1]["NAME"])
}

func TestGetResponseCode(t *testing.T) {
	// arrange
	resp := map[string]interface{}{
		"result": map[string]interface{}{"code": float64(0)},
	}
	// act
	code, err := checkResponseCode(resp)
	// assert
	assert.Equal(t, float64(0), code)
	assert.NoError(t, err)
}

func TestGetResponseCode_ErrorResponse(t *testing.T) {
	// arrange
	resp := map[string]interface{}{
		"result": map[string]interface{}{"code": float64(12345)},
	}
	// act
	code, err := checkResponseCode(resp)
	// assert
	assert.Equal(t, float64(12345), code)
	assert.ErrorContains(t, err, "response code is")
}

func TestGetResponseCode_MissingResult(t *testing.T) {
	// arrange
	resp := map[string]interface{}{
		"data": map[string]interface{}{"id": 1},
	}
	// act
	code, err := checkResponseCode(resp)
	// assert
	assert.Equal(t, float64(0), code)
	assert.ErrorContains(t, err, "can not convert resp")
}

// --- NewDistributedClient tests ---

func TestDistributedClient_NewDistributedClient_Success(t *testing.T) {
	// arrange
	config := &constant.StorageBackendConfig{
		Urls:                    []string{"https://192.168.1.1:8088"},
		User:                    "admin",
		SecretNamespace:         "default",
		SecretName:              "test-secret",
		StorageBackendNamespace: "huawei-csi",
		StorageBackendName:      "test-backend",
		ClientMaxThreads:        10,
	}

	patches := gomonkey.NewPatches()
	patches.ApplyMethod(&client.Client{}, "InitHttpClient",
		func(_ *client.Client, _ context.Context) error {
			return nil
		})
	defer patches.Reset()

	// action
	gotClient, gotErr := NewDistributedClient(context.Background(), config)

	// assert
	assert.Nil(t, gotErr)
	assert.NotNil(t, gotClient)
	assert.Equal(t, []string{"https://192.168.1.1:8088"}, gotClient.Urls)
	assert.Equal(t, "admin", gotClient.User)
	assert.NotNil(t, gotClient.SetAuthHeaders)
}

func TestDistributedClient_NewDistributedClient_InitHttpClientFailed(t *testing.T) {
	// arrange
	config := &constant.StorageBackendConfig{
		Urls:             []string{"https://192.168.1.1:8088"},
		User:             "admin",
		SecretNamespace:  "default",
		SecretName:       "test-secret",
		ClientMaxThreads: 10,
	}
	wantErr := errors.New("init http client failed")

	patches := gomonkey.NewPatches()
	patches.ApplyMethod(&client.Client{}, "InitHttpClient",
		func(_ *client.Client, _ context.Context) error {
			return wantErr
		})
	defer patches.Reset()

	// action
	gotClient, gotErr := NewDistributedClient(context.Background(), config)

	// assert
	assert.Nil(t, gotClient)
	assert.Equal(t, wantErr, gotErr)
}

// --- Login tests ---

func TestDistributedClient_Login_Success(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Urls:            []string{"https://192.168.1.1:8088"},
			User:            "admin",
			SecretNamespace: "default",
			SecretName:      "test-secret",
		},
	}
	wantToken := "test-auth-token"
	wantESN := "test-system-esn"

	patches := gomonkey.NewPatches()
	// mock GetSecretWithFallback
	patches.ApplyMethod(&client.Client{}, "GetSecretWithFallback",
		func(_ *client.Client, _ context.Context) (*v1.Secret, error) {
			return &v1.Secret{
				Data: map[string][]byte{"password": []byte("test-password")},
			}, nil
		})
	// mock GenerateUrl
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/aa/sessions", nil)
	// mock Call - return successful login response
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string,
			_ string, _ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data": map[string]interface{}{
					"x_auth_token": wantToken,
					"system_esn":   wantESN,
				},
			}, nil
		})
	defer patches.Reset()

	// action
	gotErr := dc.Login(context.Background())

	// assert
	assert.Nil(t, gotErr)
	assert.Equal(t, wantToken, dc.Token)
	assert.Equal(t, wantESN, dc.ESN)
	assert.Equal(t, "https://192.168.1.1:8088", dc.Curl)
}

func TestDistributedClient_Login_GetSecretFailed(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Urls:            []string{"https://192.168.1.1:8088"},
			User:            "admin",
			SecretNamespace: "default",
			SecretName:      "test-secret",
		},
	}
	wantErr := errors.New("secret not found")

	patches := gomonkey.NewPatches()
	patches.ApplyMethod(&client.Client{}, "GetSecretWithFallback",
		func(_ *client.Client, _ context.Context) (*v1.Secret, error) {
			return nil, wantErr
		})
	defer patches.Reset()

	// action
	gotErr := dc.Login(context.Background())

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "get login secret failed")
}

func TestDistributedClient_Login_GenerateUrlFailed(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Urls:            []string{"https://192.168.1.1:8088"},
			User:            "admin",
			SecretNamespace: "default",
			SecretName:      "test-secret",
		},
	}
	wantErr := errors.New("generate url failed")

	patches := gomonkey.NewPatches()
	patches.ApplyMethod(&client.Client{}, "GetSecretWithFallback",
		func(_ *client.Client, _ context.Context) (*v1.Secret, error) {
			return &v1.Secret{
				Data: map[string][]byte{"password": []byte("test-password")},
			}, nil
		})
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "", wantErr)
	defer patches.Reset()

	// action
	gotErr := dc.Login(context.Background())

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "generate login url failed")
}

func TestDistributedClient_Login_AllUrlsFailed(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Urls:            []string{"https://192.168.1.1:8088", "https://192.168.1.2:8088"},
			User:            "admin",
			SecretNamespace: "default",
			SecretName:      "test-secret",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyMethod(&client.Client{}, "GetSecretWithFallback",
		func(_ *client.Client, _ context.Context) (*v1.Secret, error) {
			return &v1.Secret{
				Data: map[string][]byte{"password": []byte("test-password")},
			}, nil
		})
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/aa/sessions", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string,
			_ string, _ map[string]interface{}) (map[string]interface{}, error) {
			// Return non-zero code for all URLs
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(1)},
			}, nil
		})
	defer patches.Reset()

	// action
	gotErr := dc.Login(context.Background())

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "login failed for all urls")
	assert.Equal(t, "", dc.Token)
	assert.Equal(t, "", dc.ESN)
	assert.Equal(t, "", dc.Curl)
}

// --- Logout tests ---

func TestDistributedClient_Logout_Success(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
		ESN: "test-esn",
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "/api/v2/aa/sessions", nil)
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string,
			_ string, _ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		})
	defer patches.Reset()

	// action
	dc.Logout(context.Background())

	// assert - Logout just calls DELETE, no error return to verify
	// The main assertion is that no panic occurred and Call was invoked
}

func TestDistributedClient_Logout_GenerateUrlFailed(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Curl:  "https://192.168.1.1:8088",
			Token: "test-token",
		},
		ESN: "test-esn",
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(apiDistributed.GenerateUrl, "", errors.New("url generation failed"))
	defer patches.Reset()

	// action
	dc.Logout(context.Background())

	// assert - should not panic, GenerateUrl error is logged and function returns
}

// --- setAuthHeaders tests ---

func TestDistributedClient_SetAuthHeaders_WithToken(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Token: "test-auth-token",
		},
	}
	req, err := http.NewRequest("GET", "https://example.com/test", nil)

	// action
	dc.setAuthHeaders(req)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "test-auth-token", req.Header.Get("X-Auth-Token"))
	assert.Equal(t, "X_Auth_Token=test-auth-token", req.Header.Get("Cookie"))
}

func TestDistributedClient_SetAuthHeaders_WithoutToken(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Token: "",
		},
	}
	req, err := http.NewRequest("GET", "https://example.com/test", nil)

	// action
	dc.setAuthHeaders(req)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "", req.Header.Get("X-Auth-Token"))
	assert.Equal(t, "", req.Header.Get("Cookie"))
}

// --- RegisterObjectType tests ---

func TestRegisterObjectType(t *testing.T) {
	// arrange
	typeID := 999
	listKey := "GetControllers"
	countKey := ""
	pageKey := ""
	idKey := "ID"

	// action
	RegisterObjectType(typeID, listKey, countKey, pageKey, idKey)

	// assert
	info, exists := objectTypeMapping[typeID]
	assert.True(t, exists)
	assert.Equal(t, listKey, info.ListKey)
	assert.Equal(t, idKey, info.IDKey)
}

// --- needRetry tests ---

func TestDistributedClient_NeedRetry(t *testing.T) {
	// arrange
	dc := &DistributedClient{}

	// act & assert
	assert.True(t, dc.needRetry(-401))
	assert.True(t, dc.needRetry(1077949069))
	assert.True(t, dc.needRetry(10000003))
	assert.False(t, dc.needRetry(0))
	assert.False(t, dc.needRetry(200))
}

// --- reLogin tests ---

func TestDistributedClient_ReLogin_Success(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Token: "old-token",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&DistributedClient{}, "Logout")
	patches.ApplyMethodReturn(&DistributedClient{}, "Login", nil)
	defer patches.Reset()

	// action
	gotErr := dc.reLogin(context.Background())

	// assert
	assert.Nil(t, gotErr)
}

func TestDistributedClient_ReLogin_LoginFailed(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Token: "old-token",
		},
	}
	wantErr := errors.New("login failed")

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&DistributedClient{}, "Logout")
	patches.ApplyMethodReturn(&DistributedClient{}, "Login", wantErr)
	defer patches.Reset()

	// action
	gotErr := dc.reLogin(context.Background())

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "re-login failed")
}

func TestDistributedClient_ReLogin_EmptyToken(t *testing.T) {
	// arrange
	// When Token is empty (no active session), reLogin should proceed
	// normally: Logout + Login
	dc := &DistributedClient{
		Client: client.Client{
			Token: "",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&DistributedClient{}, "Logout")
	patches.ApplyMethodReturn(&DistributedClient{}, "Login", nil)
	defer patches.Reset()

	// action
	gotErr := dc.reLogin(context.Background())

	// assert
	assert.Nil(t, gotErr)
}

// --- callWithRetry tests ---

func TestDistributedClient_CallWithRetry_Success(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Token: "test-token",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string,
			_ string, _ map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data":   map[string]interface{}{"id": "obj1"},
			}, nil
		})
	defer patches.Reset()

	// action
	gotResp, gotErr := dc.callWithRetry(context.Background(), http.MethodGet, "https://example.com/api", nil)

	// assert
	assert.Nil(t, gotErr)
	assert.NotNil(t, gotResp)
}

func TestDistributedClient_CallWithRetry_NeedRetry(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Token: "expired-token",
		},
	}
	callCount := 0

	patches := gomonkey.NewPatches()
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string,
			_ string, _ map[string]interface{}) (map[string]interface{}, error) {
			callCount++
			if callCount == 1 {
				// First call returns unauthorized code
				return map[string]interface{}{
					"result": map[string]interface{}{"code": float64(-401)},
				}, nil
			}
			// After re-login, return success
			return map[string]interface{}{
				"result": map[string]interface{}{"code": float64(0)},
				"data":   map[string]interface{}{"id": "obj1"},
			}, nil
		})
	// mock Logout and Login (called by reLogin which is unexported)
	patches.ApplyMethodReturn(&DistributedClient{}, "Logout")
	patches.ApplyMethodReturn(&DistributedClient{}, "Login", nil)
	defer patches.Reset()

	// action
	gotResp, gotErr := dc.callWithRetry(context.Background(), http.MethodGet, "https://example.com/api", nil)

	// assert
	assert.Nil(t, gotErr)
	assert.NotNil(t, gotResp)
	assert.Equal(t, 2, callCount)
}

func TestDistributedClient_CallWithRetry_CallError(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			Token: "test-token",
		},
	}
	wantErr := errors.New("connection refused")

	patches := gomonkey.NewPatches()
	patches.ApplyMethod(&client.Client{}, "Call",
		func(_ *client.Client, _ context.Context, _ string,
			_ string, _ map[string]interface{}) (map[string]interface{}, error) {
			return nil, wantErr
		})
	defer patches.Reset()

	// action
	_, gotErr := dc.callWithRetry(context.Background(), http.MethodGet, "https://example.com/api", nil)

	// assert
	assert.Equal(t, wantErr, gotErr)
}

// --- getSecret tests ---

func TestDistributedClient_GetSecret_Success(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			SecretNamespace: "default",
			SecretName:      "test-secret",
		},
	}
	wantPassword := []byte("test-password")

	patches := gomonkey.NewPatches()
	patches.ApplyMethod(&client.Client{}, "GetSecretWithFallback",
		func(_ *client.Client, _ context.Context) (*v1.Secret, error) {
			return &v1.Secret{
				Data: map[string][]byte{"password": wantPassword},
			}, nil
		})
	defer patches.Reset()

	// action
	gotBytes, gotErr := dc.getSecret(context.Background())

	// assert
	assert.Nil(t, gotErr)
	assert.Equal(t, wantPassword, gotBytes)
}

func TestDistributedClient_GetSecret_PasswordNotFound(t *testing.T) {
	// arrange
	dc := &DistributedClient{
		Client: client.Client{
			SecretNamespace: "default",
			SecretName:      "test-secret",
		},
	}

	patches := gomonkey.NewPatches()
	patches.ApplyMethod(&client.Client{}, "GetSecretWithFallback",
		func(_ *client.Client, _ context.Context) (*v1.Secret, error) {
			return &v1.Secret{
				Data: map[string][]byte{},
			}, nil
		})
	defer patches.Reset()

	// action
	_, gotErr := dc.getSecret(context.Background())

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "password key not found in secret")
}
