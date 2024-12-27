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
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/backend"
)

func TestObjectCollector_Collect_with_client_not_exist(t *testing.T) {
	// arrange
	var mockCollector = &ObjectCollector{}

	// mock
	patches := gomonkey.
		ApplyFunc(GetClient, func(context.Context, string,
			func(context.Context, string) (backend.ClientInfo, error)) (backend.ClientInfo, error) {
			return backend.ClientInfo{}, errors.New("client not exist")
		})
	defer patches.Reset()

	// action
	_, err := mockCollector.Collect(context.Background(), &cmi.CollectRequest{})

	// assert
	if err == nil || err.Error() != "client not exist" {
		t.Errorf("testObjectCollector_Collect_client_not_exist() want an error with client not exist,"+
			" but got error = %s", err.Error())
	}
}

func TestObjectCollector_Collect_with_handler_not_exist(t *testing.T) {
	// arrange
	var mockCollector = &ObjectCollector{}

	// mock
	patches := gomonkey.
		ApplyFunc(GetClient, func(context.Context, string,
			func(context.Context, string) (backend.ClientInfo, error)) (backend.ClientInfo, error) {
			return backend.ClientInfo{}, nil
		}).
		ApplyFunc(GetObjectHandler, func(storageType, collectType string) (ObjectHandler, error) {
			return nil, errors.New("handler not exist")
		})
	defer patches.Reset()

	// action
	_, err := mockCollector.Collect(context.Background(), &cmi.CollectRequest{})

	// assert
	if err == nil || err.Error() != "handler not exist" {
		t.Errorf("testObjectCollector_Collect_with_handler_not_exist() want an error with handler not exist,"+
			" but got error = %s", err.Error())
	}
}

func TestObjectCollector_Collect_with_success(t *testing.T) {
	// arrange
	var mockCollector = &ObjectCollector{}
	type mockCorrectClient struct{}
	var mockHandler = func(context.Context, interface{}, *cmi.CollectRequest) (*cmi.CollectResponse, error) {
		return &cmi.CollectResponse{}, nil
	}

	//mock
	patches := gomonkey.
		ApplyFunc(GetClient, func(context.Context, string,
			func(context.Context, string) (backend.ClientInfo, error)) (backend.ClientInfo, error) {
			return backend.ClientInfo{Client: &mockCorrectClient{}}, nil
		}).
		ApplyFunc(GetObjectHandler, func(storageType, collectType string) (ObjectHandler, error) {
			return mockHandler, nil
		})
	defer patches.Reset()

	// action
	_, err := mockCollector.Collect(context.Background(), &cmi.CollectRequest{})

	// assert
	if err != nil {
		t.Errorf("testObjectCollector_Collect_with_success() error = %v", err.Error())
	}
}

func TestDoCollect(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{
		BackendName: "test-backend",
		CollectType: "test-collect",
		MetricsType: "test-metrics",
	}
	queryFunc := func(context.Context) ([]map[string]interface{}, error) {
		var result []map[string]interface{}
		for i := 0; i < 1000; i++ {
			result = append(result, map[string]interface{}{
				"ID":   "123",
				"NAME": "name-1",
			})
		}
		return result, nil
	}

	// action
	_, err := DoCollect[[]map[string]interface{}, LunObject](context.Background(), request, queryFunc)

	// assert
	if err != nil {
		t.Errorf("TestDoPageCollect() failed, error = %v", err)
	}
}

func TestDoPageCollect(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{
		BackendName: "test-backend",
		CollectType: "test-collect",
		MetricsType: "test-metrics",
	}

	countFunc := func(ctx context.Context) (int, error) {
		return 1000, nil
	}
	pageFunc := func(ctx context.Context, start, end int) ([]map[string]interface{}, error) {
		var result []map[string]interface{}
		total := end - start
		for i := 0; i < total; i++ {
			result = append(result, map[string]interface{}{
				"ID":   "123",
				"NAME": "name-1",
			})
		}
		return result, nil
	}

	// action
	_, err := DoPageCollect[LunObject](context.Background(), request, countFunc, pageFunc)

	// assert
	if err != nil {
		t.Errorf("TestDoPageCollect() failed, error = %v", err)
	}
}
