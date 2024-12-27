/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

// Package collect is a package that provides object and performance collect
package collect

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/backend"
	"github.com/huawei/csm/v2/utils/log"
)

var mutex sync.Mutex

// clientCache
// key is backend name
// values is a storage client,
// e.g.
//
//	|-----------------|---------------------------------------|
//	| backendName     | client                                |
//	|-----------------|---------------------------------------|
//	| test-backend    | centralizedstorage.CentralizedClient  |
//	|---------------------------------------------------------|
var clientCache = map[string]backend.ClientInfo{}

// objectHandlerCache is routing table with three-layer routing
// e.g.
//
//	|-------------------|-----------------|-------------------|
//	|   StorageType     |   collectType   |  handler          |
//	|-------------------|-----------------|-------------------|
//	|   oceanStorage    |   controller    |  CollectController|
//	|-------------------|-----------------|-------------------|
//
// The above table indicates that the object data of the controllers in the ocean storage will be collected using
// the CollectController function
var objectHandlerCache = &HandlerMap[ObjectHandler]{}

// performanceHandlerCache is routing table with three-layer routing
// e.g.
//
//	|-------------------|----------------|--------------------|
//	|   StorageType     |   collectType  |  handler           |
//	|-------------------|----------------|--------------------|
//	|   oceanStorage    |   controller   |  GetLunNameMapping |
//	|-------------------|----------------|--------------------|
//
// The above table indicates that GetLunNameMapping is specified to obtain the name mapping of the lun volume
var performanceHandlerCache = &HandlerMap[PerformanceHandler]{}

// HandlerMap cache format
type HandlerMap[T any] map[string]map[string]T

// ObjectHandler object handler format
type ObjectHandler TObjectHandler[interface{}]

// TObjectHandler When clients are different, we must use generics to represent different clients
type TObjectHandler[T any] func(context.Context, T, *cmi.CollectRequest) (*cmi.CollectResponse, error)

// PerformanceHandler performance handler format
type PerformanceHandler TPerformanceHandler[interface{}]

// TPerformanceHandler When clients are different, we must use generics to represent different clients
type TPerformanceHandler[T any] func(context.Context, T) (map[string]string, error)

// RegisterObjectHandler register a function to handle object data
func RegisterObjectHandler[T any](storageType, collectType string, tHandler TObjectHandler[T]) {
	registerHandler(objectHandlerCache, storageType, collectType, tHandler.ToObjectHandler())
}

// RegisterPerformanceHandler register a function to handle performance data
func RegisterPerformanceHandler[T any](storageType, collectType string, tHandler TPerformanceHandler[T]) {
	registerHandler(performanceHandlerCache, storageType, collectType, tHandler.ToPerformanceHandler())
}

// RegisterClient key is backend name, value is ClientInfo
func RegisterClient(backendName string, info backend.ClientInfo) {
	mutex.Lock()
	defer mutex.Unlock()

	clientCache[backendName] = info
}

// RemoveClient remove the client from cache
func RemoveClient(backendName string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(clientCache, backendName)
}

// GetObjectHandler get collect object data handler
func GetObjectHandler(storageType, collectType string) (ObjectHandler, error) {
	return getHandler(objectHandlerCache, storageType, collectType)
}

// GetPerformanceHandler get collect performance data handler
func GetPerformanceHandler(storageType, collectType string) (PerformanceHandler, error) {
	return getHandler(performanceHandlerCache, storageType, collectType)
}

// registerHandler register a handler with the specified key to the cache
func registerHandler[T any](cache *HandlerMap[T], storageType, collectType string, handler T) {
	mutex.Lock()
	defer mutex.Unlock()

	handlerMap, ok := (*cache)[storageType]
	if !ok {
		handlerMap = map[string]T{}
		(*cache)[storageType] = handlerMap
	}

	handlerMap[collectType] = handler
	(*cache)[storageType] = handlerMap
}

// getHandler query whether there is a handler in the specified cache based on the specified key.
// If so, return the handler. If not, return an error
func getHandler[T any](cache *HandlerMap[T], storageType, collectType string) (T, error) {
	handlers, ok := (*cache)[storageType]
	var t T
	if !ok {
		errMsg := fmt.Sprintf("not found handlers, storageType type is [%s] ", collectType)
		return t, errors.New(errMsg)
	}

	handler, ok := handlers[collectType]
	if ok {
		return handler, nil
	}

	errMsg := fmt.Sprintf("not found handlers, collect type is [%s] ", collectType)
	return t, errors.New(errMsg)
}

// GetClient get or register client
// This function needs two parameter: backendName and discover function.
// discover function should return an instance of client.
func GetClient(ctx context.Context, backendName string,
	discoverFunc func(context.Context, string) (backend.ClientInfo, error)) (backend.ClientInfo, error) {
	client, ok := clientCache[backendName]
	if ok {
		return client, nil
	}

	client, err := discoverFunc(ctx, backendName)
	if err != nil {
		log.AddContext(ctx).Errorf("discover client failed, backend name: [%s], error: [%v]", backendName, err)
		return backend.ClientInfo{}, err
	}
	RegisterClient(backendName, client)
	return client, nil
}

// ToObjectHandler convert TObjectHandler to ObjectHandler
func (receiver TObjectHandler[T]) ToObjectHandler() ObjectHandler {
	return func(ctx context.Context, param interface{}, request *cmi.CollectRequest) (*cmi.CollectResponse, error) {
		if param == nil {
			return nil, errors.New("ToObjectHandler IllegalArgumentError, handler function argument is nil")
		}
		if t, ok := param.(T); ok {
			return receiver(ctx, t, request)
		}
		errMsg := fmt.Sprintf("ToObjectHandler IllegalArgumentError, current param is [%s], "+
			"want is [%s]", reflect.TypeOf(param).Kind().String(), reflect.TypeOf((*T)(nil)).Kind().String())
		return nil, errors.New(errMsg)
	}
}

// ToPerformanceHandler convert TPerformanceHandler to PerformanceHandler
func (receiver TPerformanceHandler[T]) ToPerformanceHandler() PerformanceHandler {
	return func(ctx context.Context, param interface{}) (map[string]string, error) {
		if param == nil {
			return nil, errors.New("ToPerformanceHandler IllegalArgumentError, handler function argument is nil")
		}
		if t, ok := param.(T); ok {
			return receiver(ctx, t)
		}
		errMsg := fmt.Sprintf("ToPerformanceHandler IllegalArgumentError, current param is [%s], "+
			"want is [%s]", reflect.TypeOf(param).Kind().String(), reflect.TypeOf((*T)(nil)).Kind().String())
		return nil, errors.New(errMsg)
	}
}
