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
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockRestClient is a minimal RestClient implementation for testing
type mockRestClient struct{}

func (m *mockRestClient) Login(ctx context.Context) error { return nil }
func (m *mockRestClient) Logout(ctx context.Context)      {}
func (m *mockRestClient) GetSingleByUrlKey(ctx context.Context, s string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *mockRestClient) GetListByUrlKey(ctx context.Context, s string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *mockRestClient) GetCountByUrlKey(ctx context.Context, s string) (int, error) { return 0, nil }
func (m *mockRestClient) GetPageByUrlKey(ctx context.Context, s string, a, b int) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *mockRestClient) QueryPerformanceData(ctx context.Context,
	i int, s []string) ([]map[string]interface{}, error) {
	return nil, nil
}

func TestNewClientCache_Success(t *testing.T) {
	// action
	gotCache := NewClientCache()

	// assert
	assert.NotNil(t, gotCache)
	assert.Equal(t, 0, gotCache.Size())
}

func TestClientCache_RegisterClient_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	info := ClientInfo{StorageName: "backend1", StorageType: "oceanStorage", Client: &mockRestClient{}}

	// action
	cache.RegisterClient("backend1", info)

	// assert
	gotClient, gotOk := cache.GetClient("backend1")
	assert.True(t, gotOk)
	assert.Equal(t, info, gotClient)
}

func TestClientCache_RemoveClient_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	cache.RegisterClient("backend1", ClientInfo{StorageName: "backend1"})

	// action
	cache.RemoveClient("backend1")

	// assert
	_, exists := cache.GetClient("backend1")
	assert.False(t, exists)
}

func TestClientCache_RemoveClient_NotExist(t *testing.T) {
	// arrange
	cache := NewClientCache()

	// action
	cache.RemoveClient("non-existent")

	// assert - no panic
	assert.Equal(t, 0, cache.Size())
}

func TestClientCache_GetClient_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	info := ClientInfo{StorageName: "backend1"}
	cache.RegisterClient("backend1", info)

	// action
	gotClient, gotOk := cache.GetClient("backend1")

	// assert
	assert.True(t, gotOk)
	assert.Equal(t, info, gotClient)
}

func TestClientCache_GetClient_NotExist(t *testing.T) {
	// arrange
	cache := NewClientCache()

	// action
	_, gotOk := cache.GetClient("non-existent")

	// assert
	assert.False(t, gotOk)
}

func TestClientCache_DiscoverClient_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	wantInfo := ClientInfo{StorageName: "backend1"}
	discoverFunc := func(ctx context.Context, name string) (ClientInfo, error) {
		return wantInfo, nil
	}

	// action
	gotClient, gotErr := cache.DiscoverClient(context.Background(), "backend1", discoverFunc)

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantInfo, gotClient)

	// verify client is cached
	cachedClient, ok := cache.GetClient("backend1")
	assert.True(t, ok)
	assert.Equal(t, wantInfo, cachedClient)
}

func TestClientCache_DiscoverClient_AlreadyCached(t *testing.T) {
	// arrange
	cache := NewClientCache()
	wantInfo := ClientInfo{StorageName: "backend1"}
	cache.RegisterClient("backend1", wantInfo)
	discoverCalled := false
	discoverFunc := func(ctx context.Context, name string) (ClientInfo, error) {
		discoverCalled = true
		return ClientInfo{}, nil
	}

	// action
	gotClient, gotErr := cache.DiscoverClient(context.Background(), "backend1", discoverFunc)

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantInfo, gotClient)
	assert.False(t, discoverCalled)
}

func TestClientCache_DiscoverClient_DiscoverFailed(t *testing.T) {
	// arrange
	cache := NewClientCache()
	wantErr := errors.New("discover failed")
	discoverFunc := func(ctx context.Context, name string) (ClientInfo, error) {
		return ClientInfo{}, wantErr
	}

	// action
	_, gotErr := cache.DiscoverClient(context.Background(), "backend1", discoverFunc)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestClientCache_DiscoverClient_DoubleCheckConcurrent(t *testing.T) {
	// arrange
	cache := NewClientCache()
	wantInfo := ClientInfo{StorageName: "backend1"}
	discoverCount := int32(0)
	var mu sync.Mutex
	discoverFunc := func(ctx context.Context, name string) (ClientInfo, error) {
		mu.Lock()
		discoverCount++
		mu.Unlock()
		return wantInfo, nil
	}

	// action - concurrent discovery for the same backend
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = cache.DiscoverClient(context.Background(), "backend1", discoverFunc)
		}()
	}
	wg.Wait()

	// assert - double-check locking should limit discover calls to 1
	assert.Equal(t, int32(1), discoverCount)
	assert.Equal(t, 1, cache.Size())
}

func TestClientCache_ClearAll_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	cache.RegisterClient("backend1", ClientInfo{Client: &mockRestClient{}})
	cache.RegisterClient("backend2", ClientInfo{Client: &mockRestClient{}})

	// action
	cache.ClearAll()

	// assert
	assert.Equal(t, 0, cache.Size())
}

func TestClientCache_ClearAll_EmptyCache(t *testing.T) {
	// arrange
	cache := NewClientCache()

	// action
	cache.ClearAll()

	// assert - no panic
	assert.Equal(t, 0, cache.Size())
}

func TestClientCache_GetAllBackendNames_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	cache.RegisterClient("backend1", ClientInfo{})
	cache.RegisterClient("backend2", ClientInfo{})

	// action
	gotNames := cache.GetAllBackendNames()

	// assert
	assert.Len(t, gotNames, 2)
	assert.Contains(t, gotNames, "backend1")
	assert.Contains(t, gotNames, "backend2")
}

func TestClientCache_Size_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	cache.RegisterClient("backend1", ClientInfo{})
	cache.RegisterClient("backend2", ClientInfo{})

	// action
	gotSize := cache.Size()

	// assert
	assert.Equal(t, 2, gotSize)
}

func TestClientCache_ConcurrentRegister(t *testing.T) {
	// arrange
	cache := NewClientCache()
	var wg sync.WaitGroup

	// action
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("backend-%d", idx)
			cache.RegisterClient(name, ClientInfo{StorageName: name})
		}(i)
	}
	wg.Wait()

	// assert
	assert.Equal(t, 100, cache.Size())
}
