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
	"sync"
)

// LabelService defines the interface for storage label operations.
// CreateLabel and DeleteLabel both return error only (no response object needed).
type LabelService interface {
	// CreateLabel creates a label on the storage backend
	CreateLabel(ctx context.Context, req *LabelRequest) error

	// DeleteLabel deletes a label from the storage backend
	DeleteLabel(ctx context.Context, req *LabelRequest) error
}

// LabelServiceFactory is a function that creates a new LabelService
type LabelServiceFactory func(clientCache *ClientCache,
	discoverFunc func(context.Context, string) (ClientInfo, error)) LabelService

var (
	labelServiceMutex      sync.RWMutex
	labelServiceFactory    LabelServiceFactory
	registeredLabelService LabelService
)

// RegisterLabelService registers a factory for creating LabelService instances.
// This allows different storage types to provide their own implementations.
// Any previously cached instance is cleared so the next GetLabelService call
// will create a new instance from the new factory.
func RegisterLabelService(factory LabelServiceFactory) {
	labelServiceMutex.Lock()
	defer labelServiceMutex.Unlock()

	labelServiceFactory = factory
	registeredLabelService = nil
}

// GetLabelService returns the registered LabelService instance.
// If no factory is registered, returns nil.
// The service is created on first call and cached for subsequent calls.
func GetLabelService(clientCache *ClientCache,
	discoverFunc func(context.Context, string) (ClientInfo, error)) LabelService {
	labelServiceMutex.Lock()
	defer labelServiceMutex.Unlock()

	if registeredLabelService != nil {
		return registeredLabelService
	}

	if labelServiceFactory != nil {
		registeredLabelService = labelServiceFactory(clientCache, discoverFunc)
	}

	return registeredLabelService
}

// SetLabelService sets the LabelService instance directly (for testing)
func SetLabelService(service LabelService) {
	labelServiceMutex.Lock()
	defer labelServiceMutex.Unlock()

	registeredLabelService = service
}
