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
	"reflect"
	"testing"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/backend"
)

func TestRegisterClient(t *testing.T) {
	// arrange
	var backendName = "mock-backend-name=register-client"
	var mockClient = backend.ClientInfo{StorageType: "test-collect"}

	// action
	RegisterClient(backendName, mockClient)

	// assert
	gotClient, ok := clientCache[backendName]
	if !ok {
		t.Errorf("RegisterClient() want = %v, but got = %v", mockClient, nil)
	}
	if !reflect.DeepEqual(mockClient, gotClient) {
		t.Errorf("RegisterClient() want = %v, but got = %v", mockClient, gotClient)
	}
}

func TestRemoveClient_success(t *testing.T) {
	//arrange
	var backendName = "mock-backend-name=register-client"
	var mockClient = backend.ClientInfo{StorageType: "test-collect"}
	RegisterClient(backendName, mockClient)

	// action
	RemoveClient(backendName)

	// assert
	_, ok := clientCache[backendName]
	if ok {
		t.Errorf("RemoveClient() failed")
	}
}

func TestGetClient_success(t *testing.T) {
	// arrange
	var backendName = "mock-backend-name-with-get-client-success"
	var mockClient = backend.ClientInfo{StorageType: "test-collect"}
	var discoverFunc = func(ctx context.Context, name string) (backend.ClientInfo, error) {
		return mockClient, nil
	}

	// action
	got, err := GetClient(context.Background(), backendName, discoverFunc)

	// assert
	if err != nil {
		t.Errorf("TestGetClient_success() error = %v", err)
		return
	}
	if !reflect.DeepEqual(got, mockClient) {
		t.Errorf("TestGetClient_success() got = %v, want %v", got, mockClient)
	}
}

func TestGetClient_with_discover_return_error(t *testing.T) {
	// arrange
	var backendName = "mock-backend-name-with-get-client"
	var discoverFunc = func(ctx context.Context, name string) (backend.ClientInfo, error) {
		return backend.ClientInfo{}, errors.New("discover error")
	}

	// action
	_, err := GetClient(context.Background(), backendName, discoverFunc)

	// assert
	if err == nil {
		t.Error("GetClient() want an error, but error is nil")
	}
}

func TestRegisterObjectHandler_success(t *testing.T) {
	// arrange
	var mockStorageType, mockCollectType = "test-storage", "test-collect"
	var mockObjectFunc = func(context.Context, interface{}, *cmi.CollectRequest) (*cmi.CollectResponse, error) {
		return nil, nil
	}

	// action
	RegisterObjectHandler(mockStorageType, mockCollectType, mockObjectFunc)

	// assert
	handlerMap, ok := (*objectHandlerCache)[mockStorageType]
	if !ok || len(handlerMap) == 0 {
		t.Error("RegisterObjectHandler() failed, want handlerMap is not empty")
		return
	}
	_, ok = handlerMap[mockCollectType]
	if !ok {
		t.Error("RegisterObjectHandler() failed, want handler is not nil")
	}
}

func TestRegisterPerformanceHandler_success(t *testing.T) {
	// arrange
	var mockStorageType, mockCollectType = "test-storage", "test-collect"
	var mockPerformanceFunc = func(context.Context, interface{}) (map[string]string, error) {
		return map[string]string{}, nil
	}

	// action
	RegisterPerformanceHandler(mockStorageType, mockCollectType, mockPerformanceFunc)

	// assert
	handlerMap, ok := (*performanceHandlerCache)[mockStorageType]
	if !ok || len(handlerMap) == 0 {
		t.Error("RegisterPerformanceHandler() failed, want handlerMap is not empty")
		return
	}
	_, ok = handlerMap[mockCollectType]
	if !ok {
		t.Error("RegisterPerformanceHandler() failed, want handler is not nil")
	}
}

func TestGetHandler_when_storage_type_is_not_exist(t *testing.T) {
	// arrange
	var storageType, collectType = "not_exist_storage_type", "collect"

	// action
	_, err := GetObjectHandler(storageType, collectType)

	// assert
	if err == nil {
		t.Error("TestGetHandler_when_storage_type_is_not_exist() want an error, but error is nil")
	}
}

func TestGetHandler_when_collect_type_is_not_exist(t *testing.T) {
	// arrange
	var storageType, collectType = "storage_type", "storage_type"
	var objectFunc = func(context.Context, interface{}, *cmi.CollectRequest) (*cmi.CollectResponse, error) {
		return nil, nil
	}

	// mock
	objectHandlerCache = &HandlerMap[ObjectHandler]{
		storageType: {
			collectType: objectFunc,
		},
	}

	// action
	_, err := GetObjectHandler(storageType, "not_exist_storage_type")

	// assert
	if err == nil {
		t.Error("TestGetHandler_when_collect_type_is_not_exist() want an error, but error is nil")
	}
}

func TestGetHandler_success(t *testing.T) {
	// arrange
	var storageType, collectType = "storage_type", "storage_type"
	var objectFunc = func(context.Context, interface{}, *cmi.CollectRequest) (*cmi.CollectResponse, error) {
		return nil, nil
	}

	// mock
	objectHandlerCache = &HandlerMap[ObjectHandler]{
		storageType: {
			collectType: objectFunc,
		},
	}

	// action
	got, err := GetObjectHandler(storageType, collectType)

	// assert
	if err != nil {
		t.Errorf("TestGetHandler_success() error = %v", err)
	}
	if !reflect.DeepEqual(reflect.ValueOf(got).Pointer(), reflect.ValueOf(objectFunc).Pointer()) {
		t.Errorf("TestGetHandler_success() got = %v, want %v",
			reflect.ValueOf(got), reflect.ValueOf(objectFunc))
	}
}
