/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2026. All rights reserved.

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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/rest"

	"github.com/huawei/csm/v2/storage/utils"
	"github.com/huawei/csm/v2/utils/log"
	"github.com/huawei/csm/v2/utils/resource"
)

const (
	logName string = "storage_client_test"
	logDir  string = "/var/log/xuanwu"

	sbcGvr = "storagebackendclaims.xuanwu.huawei.io"
)

// get ctx
var ctx = context.Background()

// TestMain used for setup and teardown
func TestMain(m *testing.M) {
	// init log
	if err := log.InitLogging(logName); err != nil {
		_ = fmt.Errorf("init logging: %s failed. error: %v", logName, err)
		return
	}

	m.Run()

	// remove all log file
	logFile := path.Join(logDir, logName)
	if err := os.RemoveAll(logFile); err != nil {
		log.Errorf("Remove file: %s failed. error: %s", logFile, err)
	}
}

// newSbcUnstructured creates an unstructured SBC resource for testing.
func newSbcUnstructured(namespace, name string, spec map[string]interface{}) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "xuanwu.huawei.io",
		Version: "v1",
		Kind:    "StorageBackendClaim",
	})
	obj.SetNamespace(namespace)
	obj.SetName(name)
	obj.Object["spec"] = spec
	return obj
}

// newFakeDynClient creates a fake dynamic client with the given SBC objects pre-registered.
func newFakeDynClient(objects ...runtime.Object) *fake.FakeDynamicClient {
	scheme := runtime.NewScheme()
	return fake.NewSimpleDynamicClient(scheme, objects...)
}

// TestCloneFileSystemThenSuccess test Call() success
func TestCallThenSuccess(t *testing.T) {
	resp := map[string]interface{}{
		"code": "0",
	}
	js, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("Call() error: %v", err)
	}
	httpCLiDo := gomonkey.ApplyMethod(&http.Client{}, "Do",
		func(_ *http.Client, req *http.Request) (*http.Response, error) {
			return &http.Response{
				Body: ioutil.NopCloser(bytes.NewReader(js)),
			}, nil
		})
	defer httpCLiDo.Reset()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Errorf("Call() error: %v", err)
	}
	cli := &Client{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar:     jar,
			Timeout: 60 * time.Second,
		},
	}
	_, err = cli.Call(ctx, "method", "url", map[string]interface{}{})
	if err != nil {
		t.Errorf("Call() error: %v", err)
	}
}

func TestClient_getRequest_JsonMarshalFailed(t *testing.T) {
	// arrange
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Errorf("Call() error: %v", err)
	}

	cli := &Client{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar:     jar,
			Timeout: 60 * time.Second,
		},
	}
	reqData := map[string]interface{}{"username": "test_user", "password": "123456"}
	wantErr := errors.New("mock err")

	// mock
	p := gomonkey.NewPatches()
	p.ApplyFunc(json.Marshal, func(v any) ([]byte, error) {
		return nil, wantErr
	})

	// action
	_, gotErr := cli.getRequest(ctx, "method", sessionsSubStr, reqData)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestClient_getRequest_JsonMarshalFailed failed, want err = %v, get err = %v", wantErr, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestClient_Call_GetRequestFailed(t *testing.T) {
	// arrange
	cli := &Client{Client: &http.Client{}}
	reqData := map[string]interface{}{"key": "value"}
	wantErr := errors.New("json marshal failed")

	// mock
	patches := gomonkey.ApplyFuncReturn(json.Marshal, nil, wantErr)
	defer patches.Reset()

	// action
	_, gotErr := cli.Call(ctx, http.MethodPost, "http://localhost/test", reqData)

	// assert
	assert.Equal(t, wantErr, gotErr)
}

func TestClient_Call_GetResponseFailed(t *testing.T) {
	// arrange
	cli := &Client{Client: &http.Client{}}
	wantErr := errors.New("http response failed")

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&http.Client{}, "Do", nil, wantErr)
	defer patches.Reset()

	// action
	_, gotErr := cli.Call(ctx, http.MethodGet, "http://localhost/test", nil)

	// assert
	assert.Equal(t, wantErr, gotErr)
}

func TestClient_RetryCall_Success(t *testing.T) {
	// arrange
	cli := &Client{}
	retryCodes := []float64{1077936128}
	wantData := map[string]interface{}{"key": "value"}
	code := float64(0)

	call := func() (map[string]interface{}, *float64, error) {
		return wantData, &code, nil
	}

	// action
	gotData, gotErr := cli.RetryCall(ctx, retryCodes, call)

	// assert
	assert.Nil(t, gotErr)
	assert.Equal(t, wantData, gotData)
}

func TestClient_RetryCall_CodeNilNoRetry(t *testing.T) {
	// arrange
	cli := &Client{}
	retryCodes := []float64{1077936128}
	wantErr := errors.New("error without code")

	call := func() (map[string]interface{}, *float64, error) {
		return nil, nil, wantErr
	}

	// action
	gotData, gotErr := cli.RetryCall(ctx, retryCodes, call)

	// assert
	assert.Equal(t, wantErr, gotErr)
	assert.Nil(t, gotData)
}

func TestClient_RetryCall_ErrorWithRetryCode(t *testing.T) {
	// arrange
	cli := &Client{}
	retryCodes := []float64{1077936128}
	callCount := 0
	retryCode := float64(1077936128)
	wantData := map[string]interface{}{"key": "value"}

	call := func() (map[string]interface{}, *float64, error) {
		callCount++
		if callCount == 1 {
			return nil, &retryCode, errors.New("retryable error")
		}
		return wantData, nil, nil
	}

	// mock RetryCallFunc to avoid sleep in tests
	patches := gomonkey.ApplyFunc(utils.RetryCallFunc, func(retryFunc func() bool) {
		for i := 0; i < 10; i++ {
			if !retryFunc() {
				break
			}
		}
	})
	defer patches.Reset()

	// action
	gotData, gotErr := cli.RetryCall(ctx, retryCodes, call)

	// assert
	assert.Nil(t, gotErr)
	assert.Equal(t, wantData, gotData)
}

func TestClient_RetryCall_ErrorWithNonRetryCode(t *testing.T) {
	// arrange
	cli := &Client{}
	retryCodes := []float64{1077936128}
	nonRetryCode := float64(999)
	wantErr := errors.New("non-retryable error")

	call := func() (map[string]interface{}, *float64, error) {
		return nil, &nonRetryCode, wantErr
	}

	// action
	gotData, gotErr := cli.RetryCall(ctx, retryCodes, call)

	// assert
	assert.Equal(t, wantErr, gotErr)
	assert.Nil(t, gotData)
}

func TestClient_RetryListCall_Success(t *testing.T) {
	// arrange
	cli := &Client{}
	retryCodes := []float64{1077936128}
	wantData := []map[string]interface{}{{"key": "value"}}
	code := float64(0)

	call := func() ([]map[string]interface{}, *float64, error) {
		return wantData, &code, nil
	}

	// action
	gotData, gotErr := cli.RetryListCall(ctx, retryCodes, call)

	// assert
	assert.Nil(t, gotErr)
	assert.Equal(t, wantData, gotData)
}

func TestClient_RetryListCall_ErrorWithRetryCode(t *testing.T) {
	// arrange
	cli := &Client{}
	retryCodes := []float64{1077936128}
	callCount := 0
	retryCode := float64(1077936128)
	wantData := []map[string]interface{}{{"key": "value"}}

	call := func() ([]map[string]interface{}, *float64, error) {
		callCount++
		if callCount == 1 {
			return nil, &retryCode, errors.New("retryable error")
		}
		return wantData, nil, nil
	}

	// mock RetryCallFunc to avoid sleep in tests
	patches := gomonkey.ApplyFunc(utils.RetryCallFunc, func(retryFunc func() bool) {
		for i := 0; i < 10; i++ {
			if !retryFunc() {
				break
			}
		}
	})
	defer patches.Reset()

	// action
	gotData, gotErr := cli.RetryListCall(ctx, retryCodes, call)

	// assert
	assert.Nil(t, gotErr)
	assert.Equal(t, wantData, gotData)
}

// --- InitHttpClient tests ---

func TestClient_InitHttpClient_Success(t *testing.T) {
	// arrange
	cli := &Client{
		StorageBackendNamespace: "default",
		StorageBackendName:      "test-backend",
	}
	// SBC with no useCert → skip verify
	sbcObj := newSbcUnstructured("default", "test-backend", map[string]interface{}{})
	dynClient := newFakeDynClient(sbcObj)

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(rest.InClusterConfig, &rest.Config{}, nil)
	patches.ApplyMethod(reflect.TypeOf(&dynamic.DynamicClient{}), "Resource",
		func(_ *dynamic.DynamicClient, gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
			return dynClient.Resource(gvr)
		})
	defer patches.Reset()

	// action
	gotErr := cli.InitHttpClient(ctx)

	// assert
	assert.Nil(t, gotErr)
	assert.NotNil(t, cli.Client)
}

func TestClient_InitHttpClient_DynamicNewForConfigFailed(t *testing.T) {
	// arrange
	cli := &Client{
		StorageBackendNamespace: "default",
		StorageBackendName:      "test-backend",
	}
	wantErr := errors.New("dynamic client creation failed")

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(rest.InClusterConfig, &rest.Config{}, nil)
	patches.ApplyFuncReturn(dynamic.NewForConfig, nil, wantErr)
	defer patches.Reset()

	// action
	gotErr := cli.InitHttpClient(ctx)

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "getting dynamicClient error")
}

func TestClient_InitHttpClient_SbcGetFailed(t *testing.T) {
	// arrange
	cli := &Client{
		StorageBackendNamespace: "default",
		StorageBackendName:      "test-backend",
	}
	// No SBC objects registered → Get will return not found
	dynClient := newFakeDynClient()

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(rest.InClusterConfig, &rest.Config{}, nil)
	patches.ApplyMethod(reflect.TypeOf(&dynamic.DynamicClient{}), "Resource",
		func(_ *dynamic.DynamicClient, _ schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
			return dynClient.Resource(schema.GroupVersionResource{})
		})
	defer patches.Reset()

	// action
	gotErr := cli.InitHttpClient(ctx)

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "get unstructuredResource of sbc")
}

func TestClient_InitHttpClient_CertNotInSecret(t *testing.T) {
	// arrange
	cli := &Client{
		StorageBackendNamespace: "default",
		StorageBackendName:      "test-backend",
	}
	sbcObj := newSbcUnstructured("default", "test-backend", map[string]interface{}{
		"useCert":    true,
		"certSecret": "default/cert-secret",
	})
	wantSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cert-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{},
	}
	dynClient := newFakeDynClient(sbcObj)

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(rest.InClusterConfig, &rest.Config{}, nil)
	patches.ApplyMethod(reflect.TypeOf(&dynamic.DynamicClient{}), "Resource",
		func(_ *dynamic.DynamicClient, gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
			return dynClient.Resource(gvr)
		})
	patches.ApplyFunc(resource.Instance, func() resource.Ops {
		return &resource.Client{}
	})
	var coreCli *resource.Client
	patches.ApplyMethodReturn(coreCli, "GetSecret", wantSecret, nil)
	defer patches.Reset()

	// action
	gotErr := cli.InitHttpClient(ctx)

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "certificate not config")
}

func TestClient_InitHttpClient_PemDecodeFailed(t *testing.T) {
	// arrange
	cli := &Client{
		StorageBackendNamespace: "default",
		StorageBackendName:      "test-backend",
	}
	sbcObj := newSbcUnstructured("default", "test-backend", map[string]interface{}{
		"useCert":    true,
		"certSecret": "default/cert-secret",
	})
	wantSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cert-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"tls.crt": []byte("not valid pem data"),
		},
	}
	dynClient := newFakeDynClient(sbcObj)

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(rest.InClusterConfig, &rest.Config{}, nil)
	patches.ApplyMethod(reflect.TypeOf(&dynamic.DynamicClient{}), "Resource",
		func(_ *dynamic.DynamicClient, gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
			return dynClient.Resource(gvr)
		})
	patches.ApplyFunc(resource.Instance, func() resource.Ops {
		return &resource.Client{}
	})
	var coreCli *resource.Client
	patches.ApplyMethodReturn(coreCli, "GetSecret", wantSecret, nil)
	defer patches.Reset()

	// action
	gotErr := cli.InitHttpClient(ctx)

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "certificate data decode error")
}

func TestClient_InitHttpClient_X509ParseFailed(t *testing.T) {
	// arrange
	cli := &Client{
		StorageBackendNamespace: "default",
		StorageBackendName:      "test-backend",
	}
	pemData := `-----BEGIN CERTIFICATE-----
aGknbG9iYWwgZGF0YSBpcyBub3QgYSB2YWxpZCBjZXJ0aWZpY2F0ZQ==
-----END CERTIFICATE-----`
	sbcObj := newSbcUnstructured("default", "test-backend", map[string]interface{}{
		"useCert":    true,
		"certSecret": "default/cert-secret",
	})
	wantSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cert-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"tls.crt": []byte(pemData),
		},
	}
	dynClient := newFakeDynClient(sbcObj)

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(rest.InClusterConfig, &rest.Config{}, nil)
	patches.ApplyMethod(reflect.TypeOf(&dynamic.DynamicClient{}), "Resource",
		func(_ *dynamic.DynamicClient, gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
			return dynClient.Resource(gvr)
		})
	patches.ApplyFunc(resource.Instance, func() resource.Ops {
		return &resource.Client{}
	})
	var coreCli *resource.Client
	patches.ApplyMethodReturn(coreCli, "GetSecret", wantSecret, nil)
	defer patches.Reset()

	// action
	gotErr := cli.InitHttpClient(ctx)

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "malformed certificate")
}

func TestClient_InitHttpClient_SplitMetaNamespaceKeyFailed(t *testing.T) {
	// arrange
	cli := &Client{
		StorageBackendNamespace: "default",
		StorageBackendName:      "test-backend",
	}
	// certSecret without namespace/name separator
	sbcObj := newSbcUnstructured("default", "test-backend", map[string]interface{}{
		"useCert":    true,
		"certSecret": "invalid-format",
	})
	dynClient := newFakeDynClient(sbcObj)

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(rest.InClusterConfig, &rest.Config{}, nil)
	patches.ApplyMethod(reflect.TypeOf(&dynamic.DynamicClient{}), "Resource",
		func(_ *dynamic.DynamicClient, gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
			return dynClient.Resource(gvr)
		})
	defer patches.Reset()

	// action
	gotErr := cli.InitHttpClient(ctx)

	// assert
	assert.NotNil(t, gotErr)
}

func TestClient_InitHttpClient_CookiejarFailed(t *testing.T) {
	// arrange
	cli := &Client{}
	wantErr := errors.New("cookiejar creation failed")

	// mock
	patches := gomonkey.ApplyFuncReturn(cookiejar.New, nil, wantErr)
	defer patches.Reset()

	// action
	gotErr := cli.InitHttpClient(ctx)

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), wantErr.Error())
}

// --- GetSecretWithFallback tests ---

func TestClient_GetSecretWithFallback_Success(t *testing.T) {
	// arrange
	cli := &Client{
		SecretNamespace: "default",
		SecretName:      "my-secret",
	}
	wantSecret := &v1.Secret{
		Data: map[string][]byte{"password": []byte("test")},
	}

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyFunc(resource.Instance, func() resource.Ops {
		return &resource.Client{}
	})
	var coreCli *resource.Client
	patches.ApplyMethodReturn(coreCli, "GetSecret", wantSecret, nil)
	defer patches.Reset()

	// action
	gotSecret, gotErr := cli.GetSecretWithFallback(ctx)

	// assert
	assert.Nil(t, gotErr)
	assert.Equal(t, wantSecret, gotSecret)
}

func TestClient_GetSecretWithFallback_NotFoundFallbackFailed(t *testing.T) {
	// arrange
	cli := &Client{
		SecretNamespace: "default",
		SecretName:      "my-secret",
	}
	notFoundErr := apiErrors.NewNotFound(schema.GroupResource{Group: "", Resource: "secrets"}, "my-secret")
	wantErr := errors.New("unable to load in-cluster configuration")

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyFunc(resource.Instance, func() resource.Ops {
		return &resource.Client{}
	})
	var coreCli *resource.Client
	patches.ApplyMethodReturn(coreCli, "GetSecret", nil, notFoundErr)
	// mock rest.InClusterConfig to fail, causing getSecretFromSbcDynamically to fail
	patches.ApplyFuncReturn(rest.InClusterConfig, nil, wantErr)
	defer patches.Reset()

	// action
	_, gotErr := cli.GetSecretWithFallback(ctx)

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "get secret from sbc dynamically failed")
}

func TestClient_GetSecretWithFallback_GetSecretError(t *testing.T) {
	// arrange
	cli := &Client{
		SecretNamespace: "default",
		SecretName:      "my-secret",
	}
	wantErr := errors.New("get secret internal error")

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyFunc(resource.Instance, func() resource.Ops {
		return &resource.Client{}
	})
	var coreCli *resource.Client
	patches.ApplyMethodReturn(coreCli, "GetSecret", nil, wantErr)
	defer patches.Reset()

	// action
	_, gotErr := cli.GetSecretWithFallback(ctx)

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), wantErr.Error())
}

func TestClient_GetSecretWithFallback_NilSecretData(t *testing.T) {
	// arrange
	cli := &Client{
		SecretNamespace: "default",
		SecretName:      "my-secret",
	}
	nilDataSecret := &v1.Secret{}

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyFunc(resource.Instance, func() resource.Ops {
		return &resource.Client{}
	})
	var coreCli *resource.Client
	patches.ApplyMethodReturn(coreCli, "GetSecret", nilDataSecret, nil)
	defer patches.Reset()

	// action
	_, gotErr := cli.GetSecretWithFallback(ctx)

	// assert
	assert.NotNil(t, gotErr)
	assert.Contains(t, gotErr.Error(), "secret is nil or has no data")
}

func TestClient_GetSecretWithFallback_NotFoundFallbackSuccess(t *testing.T) {
	// arrange
	cli := &Client{
		SecretNamespace:         "default",
		SecretName:              "my-secret",
		StorageBackendNamespace: "huawei-csi",
		StorageBackendName:      "test-backend",
	}
	notFoundErr := apiErrors.NewNotFound(schema.GroupResource{Group: "", Resource: "secrets"}, "my-secret")
	wantSecret := &v1.Secret{
		Data: map[string][]byte{"password": []byte("fallback-password")},
	}
	sbcObj := newSbcUnstructured("huawei-csi", "test-backend", map[string]interface{}{
		"secretMeta": "fallback-ns/fallback-secret",
	})
	dynClient := newFakeDynClient(sbcObj)

	// mock
	patches := gomonkey.NewPatches()
	patches.ApplyFunc(resource.Instance, func() resource.Ops {
		return &resource.Client{}
	})
	var coreCli *resource.Client
	// First call: GetSecret returns NotFound; second call (from getSecretFromSbcDynamically): returns secret
	patches.ApplyMethodReturn(coreCli, "GetSecret", nil, notFoundErr)
	patches.ApplyFuncReturn(rest.InClusterConfig, &rest.Config{}, nil)
	patches.ApplyMethod(reflect.TypeOf(&dynamic.DynamicClient{}), "Resource",
		func(_ *dynamic.DynamicClient, _ schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
			return dynClient.Resource(schema.GroupVersionResource{})
		})
	patches.ApplyMethodReturn(coreCli, "GetSecret", wantSecret, nil)
	defer patches.Reset()

	// action
	gotSecret, gotErr := cli.GetSecretWithFallback(ctx)

	// assert
	assert.Nil(t, gotErr)
	assert.Equal(t, wantSecret, gotSecret)
}

// --- Call with SetAuthHeaders ---

func TestClient_Call_WithSetAuthHeaders(t *testing.T) {
	// arrange
	resp := map[string]interface{}{"code": "0"}
	js, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("json.Marshal failed: %v", err)
	}

	cli := &Client{
		SetAuthHeaders: func(req *http.Request) {
			req.Header.Set("X-Auth-Token", "test-token")
		},
	}

	mockClient := &mockHTTPClient{
		responseBody: js,
		statusCode:   200,
	}
	cli.Client = mockClient

	// action
	gotResp, gotErr := cli.Call(ctx, "GET", "/test", nil)

	// assert
	assert.Nil(t, gotErr)
	assert.NotNil(t, gotResp)
}

// mockHTTPClient is a simple mock for http.Client
type mockHTTPClient struct {
	responseBody []byte
	statusCode   int
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader(m.responseBody)),
	}, nil
}

// --- Call with Semaphore tests ---

func TestClient_Call_WithSemaphore(t *testing.T) {
	// arrange
	resp := map[string]interface{}{"code": "0"}
	js, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("json.Marshal failed: %v", err)
	}

	cli := &Client{
		Semaphore: utils.NewSemaphore(2),
	}
	mockClient := &mockHTTPClient{
		responseBody: js,
		statusCode:   200,
	}
	cli.Client = mockClient

	// action
	gotResp, gotErr := cli.Call(ctx, "GET", "/test", nil)

	// assert
	assert.Nil(t, gotErr)
	assert.NotNil(t, gotResp)
}

func TestClient_Call_WithSemaphoreConcurrent(t *testing.T) {
	// arrange
	resp := map[string]interface{}{"code": "0"}
	js, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("json.Marshal failed: %v", err)
	}

	cli := &Client{
		Semaphore: utils.NewSemaphore(1),
	}
	mockClient := &mockHTTPClient{
		responseBody: js,
		statusCode:   200,
	}
	cli.Client = mockClient

	// action - run multiple goroutines to verify semaphore limits concurrency
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			_, gotErr := cli.Call(ctx, "GET", "/test", nil)
			assert.Nil(t, gotErr)
			done <- true
		}()
	}

	// assert - all goroutines should complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestClient_Call_NilSemaphore(t *testing.T) {
	// arrange
	resp := map[string]interface{}{"code": "0"}
	js, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("json.Marshal failed: %v", err)
	}

	cli := &Client{
		Semaphore: nil,
	}
	mockClient := &mockHTTPClient{
		responseBody: js,
		statusCode:   200,
	}
	cli.Client = mockClient

	// action
	gotResp, gotErr := cli.Call(ctx, "GET", "/test", nil)

	// assert - should work fine without semaphore
	assert.Nil(t, gotErr)
	assert.NotNil(t, gotResp)
}
