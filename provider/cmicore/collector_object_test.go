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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewObjectCollector_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{}, nil
	}

	// action
	gotCollector := NewObjectCollector(cache, discoverFunc)

	// assert
	assert.NotNil(t, gotCollector)
}

func TestObjectCollector_Collect_DiscoverClientFailed(t *testing.T) {
	// arrange
	wantErr := errors.New("discover client failed")
	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{}, wantErr
	}
	collector := NewObjectCollector(cache, discoverFunc)
	req := &CollectRequest{BackendName: "b1", CollectType: "lun", MetricsType: "object"}

	// action
	_, gotErr := collector.Collect(context.Background(), req)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestObjectCollector_Collect_HandlerNotFound(t *testing.T) {
	// arrange
	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{StorageType: "nonexistent-storage", Client: &mockRestClient{}}, nil
	}
	collector := NewObjectCollector(cache, discoverFunc)
	req := &CollectRequest{BackendName: "b1", CollectType: "lun", MetricsType: "object"}
	wantErr := true

	// action
	_, gotErr := collector.Collect(context.Background(), req)

	// assert
	assert.Equal(t, wantErr, gotErr != nil)
	assert.Contains(t, gotErr.Error(), "nonexistent-storage")
}

func TestObjectCollector_Collect_Success(t *testing.T) {
	// arrange
	wantBackendName := "obj-collect-test"
	handler := func(ctx context.Context, client RestClient, req *CollectRequest) (*CollectResponse, error) {
		return &CollectResponse{BackendName: wantBackendName}, nil
	}
	RegisterObjectHandlerDirect("obj-test-storage", "obj-test-collect", handler)

	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{StorageType: "obj-test-storage", Client: &mockRestClient{}}, nil
	}
	collector := NewObjectCollector(cache, discoverFunc)
	req := &CollectRequest{BackendName: "b1", CollectType: "obj-test-collect", MetricsType: "object"}

	// action
	gotResp, gotErr := collector.Collect(context.Background(), req)

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantBackendName, gotResp.BackendName)
}

func TestObjectCollector_Collect_HandlerReturnsError(t *testing.T) {
	// arrange
	wantErr := errors.New("handler error")
	handler := func(ctx context.Context, client RestClient, req *CollectRequest) (*CollectResponse, error) {
		return nil, wantErr
	}
	RegisterObjectHandlerDirect("obj-err-storage", "obj-err-collect", handler)

	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{StorageType: "obj-err-storage", Client: &mockRestClient{}}, nil
	}
	collector := NewObjectCollector(cache, discoverFunc)
	req := &CollectRequest{BackendName: "b1", CollectType: "obj-err-collect", MetricsType: "object"}

	// action
	_, gotErr := collector.Collect(context.Background(), req)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

// Suppress unused import warning for gomonkey (used in other test files in this package)
var _ = gomonkey.NewPatches
