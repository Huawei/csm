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
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockLabelService struct {
	createCalled bool
	deleteCalled bool
}

func (m *mockLabelService) CreateLabel(ctx context.Context, req *LabelRequest) error {
	m.createCalled = true
	return nil
}

func (m *mockLabelService) DeleteLabel(ctx context.Context, req *LabelRequest) error {
	m.deleteCalled = true
	return nil
}

func TestRegisterLabelService_GetLabelService_Success(t *testing.T) {
	// arrange
	SetLabelService(nil)
	factory := func(clientCache *ClientCache,
		discoverFunc func(context.Context, string) (ClientInfo, error)) LabelService {
		return &mockLabelService{}
	}
	RegisterLabelService(factory)
	cache := NewClientCache()
	discoverFunc := func(ctx context.Context, name string) (ClientInfo, error) {
		return ClientInfo{}, nil
	}

	// action
	gotService := GetLabelService(cache, discoverFunc)

	// assert
	assert.NotNil(t, gotService)
}

func TestGetLabelService_NoRegistration(t *testing.T) {
	// arrange
	SetLabelService(nil)
	RegisterLabelService(nil)
	cache := NewClientCache()
	discoverFunc := func(ctx context.Context, name string) (ClientInfo, error) {
		return ClientInfo{}, nil
	}

	// action
	gotService := GetLabelService(cache, discoverFunc)

	// assert
	assert.Nil(t, gotService)
}

func TestSetLabelService_Success(t *testing.T) {
	// arrange
	mock := &mockLabelService{}
	SetLabelService(mock)
	cache := NewClientCache()
	discoverFunc := func(ctx context.Context, name string) (ClientInfo, error) {
		return ClientInfo{}, nil
	}

	// action
	gotService := GetLabelService(cache, discoverFunc)

	// assert
	assert.Equal(t, mock, gotService)
	SetLabelService(nil)
	RegisterLabelService(nil)
}

func TestRegisterLabelService_ReRegisterClearsCache(t *testing.T) {
	// arrange - register factory A and get a cached instance
	SetLabelService(nil)
	factoryA := func(clientCache *ClientCache,
		discoverFunc func(context.Context, string) (ClientInfo, error)) LabelService {
		return &mockLabelService{createCalled: false}
	}
	RegisterLabelService(factoryA)
	cache := NewClientCache()
	discoverFunc := func(ctx context.Context, name string) (ClientInfo, error) {
		return ClientInfo{}, nil
	}
	firstService := GetLabelService(cache, discoverFunc)
	assert.NotNil(t, firstService)

	// arrange - register factory B, which should clear the cached instance
	factoryB := func(clientCache *ClientCache,
		discoverFunc func(context.Context, string) (ClientInfo, error)) LabelService {
		return &mockLabelService{createCalled: true}
	}
	RegisterLabelService(factoryB)

	// action - GetLabelService should now use factory B, not return the old cached instance
	gotService := GetLabelService(cache, discoverFunc)

	// assert - gotService should be a new instance from factory B, not firstService
	assert.NotNil(t, gotService)
	assert.NotEqual(t, firstService, gotService)

	// cleanup
	SetLabelService(nil)
	RegisterLabelService(nil)
}
