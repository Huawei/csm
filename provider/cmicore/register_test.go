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

// Package cmicore provides a shared library for storage backend operations
package cmicore

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testClient is a concrete type for testing generic handler registration
type testClient struct {
	Name string
}

func (c *testClient) Login(ctx context.Context) error { return nil }
func (c *testClient) Logout(ctx context.Context)      {}
func (c *testClient) GetSingleByUrlKey(ctx context.Context, s string) (map[string]interface{}, error) {
	return nil, nil
}
func (c *testClient) GetListByUrlKey(ctx context.Context, s string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (c *testClient) GetCountByUrlKey(ctx context.Context, s string) (int, error) { return 0, nil }
func (c *testClient) GetPageByUrlKey(ctx context.Context, s string, a, b int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (c *testClient) QueryPerformanceData(ctx context.Context, i int, s []string) ([]map[string]interface{}, error) {
	return nil, nil
}

// anotherRestClient is used for type mismatch testing
type anotherRestClient struct{}

func (c *anotherRestClient) Login(ctx context.Context) error { return nil }
func (c *anotherRestClient) Logout(ctx context.Context)      {}
func (c *anotherRestClient) GetSingleByUrlKey(ctx context.Context, s string) (map[string]interface{}, error) {
	return nil, nil
}
func (c *anotherRestClient) GetListByUrlKey(ctx context.Context, s string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (c *anotherRestClient) GetCountByUrlKey(ctx context.Context, s string) (int, error) {
	return 0, nil
}
func (c *anotherRestClient) GetPageByUrlKey(ctx context.Context, s string, a, b int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (c *anotherRestClient) QueryPerformanceData(ctx context.Context,
	i int, s []string) ([]map[string]interface{}, error) {
	return nil, nil
}

func TestRegisterObjectHandler_GetObjectHandler_Success(t *testing.T) {
	// arrange
	wantBackendName := "generic-test"
	handler := func(ctx context.Context, client *testClient, req *CollectRequest) (*CollectResponse, error) {
		return &CollectResponse{BackendName: wantBackendName}, nil
	}
	RegisterObjectHandler("gen-storage", "gen-collect", TObjectHandler[*testClient](handler))

	// action
	gotHandler, gotErr := GetObjectHandler("gen-storage", "gen-collect")

	// assert
	assert.NoError(t, gotErr)
	gotResp, err := gotHandler(context.Background(), &testClient{Name: "c1"}, &CollectRequest{})
	assert.NoError(t, err)
	assert.Equal(t, wantBackendName, gotResp.BackendName)
}

func TestRegisterObjectHandlerDirect_GetObjectHandler_Success(t *testing.T) {
	// arrange
	wantBackendName := "direct-test"
	handler := func(ctx context.Context, client RestClient, req *CollectRequest) (*CollectResponse, error) {
		return &CollectResponse{BackendName: wantBackendName}, nil
	}
	RegisterObjectHandlerDirect("direct-storage", "direct-collect", handler)

	// action
	gotHandler, gotErr := GetObjectHandler("direct-storage", "direct-collect")

	// assert
	assert.NoError(t, gotErr)
	gotResp, err := gotHandler(context.Background(), &testClient{Name: "c1"}, &CollectRequest{})
	assert.NoError(t, err)
	assert.Equal(t, wantBackendName, gotResp.BackendName)
}

func TestGetObjectHandler_StorageTypeNotFound(t *testing.T) {
	// arrange
	wantErr := true

	// action
	_, gotErr := GetObjectHandler("nonexistent-storage", "any-collect")

	// assert
	assert.Equal(t, wantErr, gotErr != nil)
	assert.Contains(t, gotErr.Error(), "nonexistent-storage")
}

func TestGetObjectHandler_CollectTypeNotFound(t *testing.T) {
	// arrange
	handler := func(ctx context.Context, client RestClient, req *CollectRequest) (*CollectResponse, error) {
		return nil, nil
	}
	RegisterObjectHandlerDirect("exist-storage", "exist-collect", handler)
	wantErr := true

	// action
	_, gotErr := GetObjectHandler("exist-storage", "nonexistent-collect")

	// assert
	assert.Equal(t, wantErr, gotErr != nil)
	assert.Contains(t, gotErr.Error(), "nonexistent-collect")
}

func TestRegisterPerformanceHandler_GetPerformanceHandler_Success(t *testing.T) {
	// arrange
	wantBackendName := "gen-perf-test"
	handler := func(ctx context.Context, client *testClient, req *CollectRequest) (*CollectResponse, error) {
		return &CollectResponse{BackendName: wantBackendName}, nil
	}
	RegisterPerformanceHandler("gen-perf-storage", "gen-perf-collect", handler)

	// action
	gotHandler, gotErr := GetPerformanceHandler("gen-perf-storage", "gen-perf-collect")

	// assert
	assert.NoError(t, gotErr)
	gotResp, err := gotHandler(context.Background(), &testClient{Name: "c1"}, &CollectRequest{})
	assert.NoError(t, err)
	assert.Equal(t, wantBackendName, gotResp.BackendName)
}

func TestRegisterPerformanceHandlerDirect_GetPerformanceHandler_Success(t *testing.T) {
	// arrange
	wantBackendName := "direct-perf-test"
	handler := func(ctx context.Context, client RestClient, req *CollectRequest) (*CollectResponse, error) {
		return &CollectResponse{BackendName: wantBackendName}, nil
	}
	RegisterPerformanceHandlerDirect("direct-perf-storage", "direct-perf-collect", handler)

	// action
	gotHandler, gotErr := GetPerformanceHandler("direct-perf-storage", "direct-perf-collect")

	// assert
	assert.NoError(t, gotErr)
	gotResp, err := gotHandler(context.Background(), &testClient{Name: "c1"}, &CollectRequest{})
	assert.NoError(t, err)
	assert.Equal(t, wantBackendName, gotResp.BackendName)
}

func TestGetPerformanceHandler_StorageTypeNotFound(t *testing.T) {
	// arrange
	wantErr := true

	// action
	_, gotErr := GetPerformanceHandler("nonexistent-perf-storage", "any-collect")

	// assert
	assert.Equal(t, wantErr, gotErr != nil)
}

func TestGetPerformanceHandler_CollectTypeNotFound(t *testing.T) {
	// arrange
	handler := func(ctx context.Context, client RestClient, req *CollectRequest) (*CollectResponse, error) {
		return nil, nil
	}
	RegisterPerformanceHandlerDirect("exist-perf-storage", "exist-perf-collect", handler)
	wantErr := true

	// action
	_, gotErr := GetPerformanceHandler("exist-perf-storage", "nonexistent-perf-collect")

	// assert
	assert.Equal(t, wantErr, gotErr != nil)
}

func TestToObjectHandler_Success(t *testing.T) {
	// arrange
	wantBackendName := "type-erasure-test"
	tHandler := TObjectHandler[*testClient](func(ctx context.Context,
		client *testClient, req *CollectRequest) (*CollectResponse, error) {
		return &CollectResponse{BackendName: wantBackendName}, nil
	})
	erasedHandler := tHandler.ToObjectHandler()

	// action
	gotResp, gotErr := erasedHandler(context.Background(), &testClient{Name: "c1"}, &CollectRequest{})

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantBackendName, gotResp.BackendName)
}

func TestToObjectHandler_NilParam(t *testing.T) {
	// arrange
	tHandler := TObjectHandler[*testClient](func(ctx context.Context,
		client *testClient, req *CollectRequest) (*CollectResponse, error) {
		return nil, nil
	})
	erasedHandler := tHandler.ToObjectHandler()

	// action
	var nilClient RestClient = nil
	_, gotErr := erasedHandler(context.Background(), nilClient, &CollectRequest{})

	// assert
	assert.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "IllegalArgumentError")
}

func TestToObjectHandler_TypeMismatch(t *testing.T) {
	// arrange
	tHandler := TObjectHandler[*testClient](func(ctx context.Context,
		client *testClient, req *CollectRequest) (*CollectResponse, error) {
		return nil, nil
	})
	erasedHandler := tHandler.ToObjectHandler()
	wantErr := true

	// action - pass a different RestClient implementation
	_, gotErr := erasedHandler(context.Background(), &anotherRestClient{}, &CollectRequest{})

	// assert
	assert.Equal(t, wantErr, gotErr != nil)
	assert.Contains(t, gotErr.Error(), "IllegalArgumentError")
}

func TestToPerformanceHandler_Success(t *testing.T) {
	// arrange
	wantBackendName := "perf-type-erasure-test"
	tHandler := TPerformanceHandler[*testClient](func(ctx context.Context,
		client *testClient, req *CollectRequest) (*CollectResponse, error) {
		return &CollectResponse{BackendName: wantBackendName}, nil
	})
	erasedHandler := tHandler.ToPerformanceHandler()

	// action
	gotResp, gotErr := erasedHandler(context.Background(), &testClient{Name: "c1"}, &CollectRequest{})

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantBackendName, gotResp.BackendName)
}

func TestToPerformanceHandler_NilParam(t *testing.T) {
	// arrange
	tHandler := TPerformanceHandler[*testClient](func(ctx context.Context,
		client *testClient, req *CollectRequest) (*CollectResponse, error) {
		return nil, nil
	})
	erasedHandler := tHandler.ToPerformanceHandler()

	// action
	var nilClient RestClient = nil
	_, gotErr := erasedHandler(context.Background(), nilClient, &CollectRequest{})

	// assert
	assert.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "IllegalArgumentError")
}

func TestToPerformanceHandler_TypeMismatch(t *testing.T) {
	// arrange
	tHandler := TPerformanceHandler[*testClient](func(ctx context.Context,
		client *testClient, req *CollectRequest) (*CollectResponse, error) {
		return nil, nil
	})
	erasedHandler := tHandler.ToPerformanceHandler()
	wantErr := true

	// action - pass a different RestClient implementation
	_, gotErr := erasedHandler(context.Background(), &anotherRestClient{}, &CollectRequest{})

	// assert
	assert.Equal(t, wantErr, gotErr != nil)
	assert.Contains(t, gotErr.Error(), "IllegalArgumentError")
}

func TestToObjectHandler_HandlerReturnsError(t *testing.T) {
	// arrange
	wantErr := errors.New("handler internal error")
	tHandler := TObjectHandler[*testClient](func(ctx context.Context,
		client *testClient, req *CollectRequest) (*CollectResponse, error) {
		return nil, wantErr
	})
	erasedHandler := tHandler.ToObjectHandler()

	// action
	_, gotErr := erasedHandler(context.Background(), &testClient{Name: "c1"}, &CollectRequest{})

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestToPerformanceHandler_HandlerReturnsError(t *testing.T) {
	// arrange
	wantErr := errors.New("perf handler error")
	tHandler := TPerformanceHandler[*testClient](func(ctx context.Context,
		client *testClient, req *CollectRequest) (*CollectResponse, error) {
		return nil, wantErr
	})
	erasedHandler := tHandler.ToPerformanceHandler()

	// action
	_, gotErr := erasedHandler(context.Background(), &testClient{Name: "c1"}, &CollectRequest{})

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}
