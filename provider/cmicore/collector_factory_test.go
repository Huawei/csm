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

	"github.com/huawei/csm/v2/provider/constants"
)

func TestNewCollectorFactory_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{}, nil
	}

	// action
	gotFactory := NewCollectorFactory(cache, discoverFunc)

	// assert
	assert.NotNil(t, gotFactory)
}

func TestCollectorFactory_GetCollector_ObjectSuccess(t *testing.T) {
	// arrange
	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{}, nil
	}
	factory := NewCollectorFactory(cache, discoverFunc)

	// action
	gotCollector, gotErr := factory.GetCollector(constants.Object)

	// assert
	assert.NoError(t, gotErr)
	assert.NotNil(t, gotCollector)
}

func TestCollectorFactory_GetCollector_PerformanceSuccess(t *testing.T) {
	// arrange
	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{}, nil
	}
	factory := NewCollectorFactory(cache, discoverFunc)

	// action
	gotCollector, gotErr := factory.GetCollector(constants.Performance)

	// assert
	assert.NoError(t, gotErr)
	assert.NotNil(t, gotCollector)
}

func TestCollectorFactory_GetCollector_InvalidType(t *testing.T) {
	// arrange
	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{}, nil
	}
	factory := NewCollectorFactory(cache, discoverFunc)
	wantErr := true

	// action
	gotCollector, gotErr := factory.GetCollector("invalid-type")

	// assert
	assert.Equal(t, wantErr, gotErr != nil)
	assert.Nil(t, gotCollector)
	assert.Contains(t, gotErr.Error(), "invalid-type")
}
