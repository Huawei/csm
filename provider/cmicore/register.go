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
	"reflect"
	"sync"
)

var handlerMutex sync.RWMutex

// HandlerMap cache format for two-level routing: storageType -> collectType -> handler
type HandlerMap[T any] map[string]map[string]T

// ObjectHandler is a function that handles object data collection
type ObjectHandler func(ctx context.Context, client RestClient, req *CollectRequest) (*CollectResponse, error)

// TObjectHandler is a generic object handler that preserves client type information
type TObjectHandler[T any] func(ctx context.Context, client T, req *CollectRequest) (*CollectResponse, error)

// PerformanceHandler is a function that handles performance data collection.
// It returns a complete CollectResponse with performance data merged with object names.
type PerformanceHandler func(ctx context.Context, client RestClient, req *CollectRequest) (*CollectResponse, error)

// TPerformanceHandler is a generic performance handler that preserves client type information
type TPerformanceHandler[T any] func(ctx context.Context, client T, req *CollectRequest) (*CollectResponse, error)

// objectHandlerCache routes object collect requests to handler functions
var objectHandlerCache = &HandlerMap[ObjectHandler]{}

// performanceHandlerCache routes performance collect requests to handler functions
var performanceHandlerCache = &HandlerMap[PerformanceHandler]{}

// RegisterObjectHandler registers a function to handle object data collection.
// This is called from init() functions in storage backend implementations.
func RegisterObjectHandler[T any](storageType, collectType string, tHandler TObjectHandler[T]) {
	registerHandler(objectHandlerCache, storageType, collectType, tHandler.ToObjectHandler())
}

// RegisterObjectHandlerDirect registers an ObjectHandler directly without generic type erasure.
func RegisterObjectHandlerDirect(storageType, collectType string, handler ObjectHandler) {
	registerHandler(objectHandlerCache, storageType, collectType, handler)
}

// RegisterPerformanceHandler registers a function to handle performance data collection.
// This is called from init() functions in storage backend implementations.
func RegisterPerformanceHandler[T any](storageType, collectType string, tHandler TPerformanceHandler[T]) {
	registerHandler(performanceHandlerCache, storageType, collectType, tHandler.ToPerformanceHandler())
}

// RegisterPerformanceHandlerDirect registers a PerformanceHandler directly without generic type erasure.
func RegisterPerformanceHandlerDirect(storageType, collectType string, handler PerformanceHandler) {
	registerHandler(performanceHandlerCache, storageType, collectType, handler)
}

// GetObjectHandler returns the object handler for the given storage and collect type
func GetObjectHandler(storageType, collectType string) (ObjectHandler, error) {
	return getHandler(objectHandlerCache, storageType, collectType)
}

// GetPerformanceHandler returns the performance handler for the given storage and collect type
func GetPerformanceHandler(storageType, collectType string) (PerformanceHandler, error) {
	return getHandler(performanceHandlerCache, storageType, collectType)
}

// registerHandler registers a handler with the specified key to the cache
func registerHandler[T any](cache *HandlerMap[T], storageType, collectType string, handler T) {
	handlerMutex.Lock()
	defer handlerMutex.Unlock()

	handlerMap, ok := (*cache)[storageType]
	if !ok {
		handlerMap = map[string]T{}
		(*cache)[storageType] = handlerMap
	}

	handlerMap[collectType] = handler
	(*cache)[storageType] = handlerMap
}

// getHandler queries whether there is a handler in the specified cache based on the given keys.
// Returns the handler if found, or an error if not.
func getHandler[T any](cache *HandlerMap[T], storageType, collectType string) (T, error) {
	handlerMutex.RLock()
	defer handlerMutex.RUnlock()

	handlers, ok := (*cache)[storageType]
	var t T
	if !ok {
		errMsg := fmt.Sprintf("not found handlers, storage type is [%s] ", storageType)
		return t, errors.New(errMsg)
	}

	handler, ok := handlers[collectType]
	if ok {
		return handler, nil
	}

	errMsg := fmt.Sprintf("not found handlers, collect type is [%s] ", collectType)
	return t, errors.New(errMsg)
}

// ToObjectHandler converts TObjectHandler to ObjectHandler (type erasure)
func (receiver TObjectHandler[T]) ToObjectHandler() ObjectHandler {
	return func(ctx context.Context, param RestClient, req *CollectRequest) (*CollectResponse, error) {
		if param == nil {
			return nil, errors.New("ToObjectHandler IllegalArgumentError, handler function argument is nil")
		}
		if t, ok := param.(T); ok {
			return receiver(ctx, t, req)
		}
		errMsg := fmt.Sprintf("ToObjectHandler IllegalArgumentError, current param is [%s], "+
			"want is [%s]", reflect.TypeOf(param).Kind().String(), reflect.TypeOf((*T)(nil)).Kind().String())
		return nil, errors.New(errMsg)
	}
}

// ToPerformanceHandler converts TPerformanceHandler to PerformanceHandler (type erasure)
func (receiver TPerformanceHandler[T]) ToPerformanceHandler() PerformanceHandler {
	return func(ctx context.Context, param RestClient, req *CollectRequest) (*CollectResponse, error) {
		if param == nil {
			return nil, errors.New("ToPerformanceHandler IllegalArgumentError, handler function argument is nil")
		}
		if t, ok := param.(T); ok {
			return receiver(ctx, t, req)
		}
		errMsg := fmt.Sprintf("ToPerformanceHandler IllegalArgumentError, current param is [%s], "+
			"want is [%s]", reflect.TypeOf(param).Kind().String(), reflect.TypeOf((*T)(nil)).Kind().String())
		return nil, errors.New(errMsg)
	}
}
