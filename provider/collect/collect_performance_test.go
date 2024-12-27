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

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/provider/utils"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
)

func TestDoNameMapping(t *testing.T) {
	// arrange
	data := []map[string]interface{}{
		{"ID": "ID-1", "NAME": "NAME-1"},
		{"ID": "ID-2", "NAME": "NAME-2"},
		{"ID": "ID-3", "NAME": "NAME-3"},
		{"ID": "ID-3", "NAME-NOT-EXIST": "NAME--NOT-EXIST"},
		{"ID-NOT-EXIST": "ID-NOT-EXIST", "NAME-NOT-EXIST": "NAME-NOT-EXIST"},
	}
	want := map[string]string{
		"ID-1": "NAME-1",
		"ID-2": "NAME-2",
		"ID-3": "NAME-3",
	}

	// action
	got := DoNameMapping(data)

	// assert
	if !reflect.DeepEqual(want, got) {
		t.Errorf("TestDoNameMapping() want = %v, but got = %v", want, got)
	}
}

func TestGetNameMappingWithPage(t *testing.T) {
	// arrange
	want := map[string]string{
		"ID-1": "NAME-1",
		"ID-2": "NAME-2",
		"ID-3": "NAME-3",
	}

	countFunc := func(ctx context.Context) (int, error) {
		return 1000, nil
	}

	pageFunc := func(ctx context.Context, start, end int) ([]map[string]interface{}, error) {
		return []map[string]interface{}{}, nil
	}

	// mock
	applyFunc := gomonkey.ApplyFunc(ConcurrentPaginate, func(context.Context, CountFunc,
		PageFunc) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{"ID": "ID-1", "NAME": "NAME-1"},
			{"ID": "ID-2", "NAME": "NAME-2"},
			{"ID": "ID-3", "NAME": "NAME-3"},
		}, nil
	})
	defer applyFunc.Reset()

	// action
	got, err := GetNameMappingWithPage(context.Background(), countFunc, pageFunc)

	// assert
	if err != nil {
		t.Errorf("TestGetNameMappingWithPage() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestGetNameMappingWithPage() want = %v, but got = %v", want, got)
	}
}

func TestGetNameMapping(t *testing.T) {
	// arrange
	want := map[string]string{
		"ID-1": "NAME-1",
		"ID-2": "NAME-2",
		"ID-3": "NAME-3",
	}

	queryFunc := func(ctx context.Context) ([]map[string]interface{}, error) {
		return []map[string]interface{}{
			{"ID": "ID-1", "NAME": "NAME-1"},
			{"ID": "ID-2", "NAME": "NAME-2"},
			{"ID": "ID-3", "NAME": "NAME-3"},
		}, nil
	}

	// action
	got, err := GetNameMapping(context.Background(), queryFunc)

	// assert
	if err != nil {
		t.Errorf("TestGetNameMapping() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestGetNameMapping() want = %v, but got = %v", want, got)
	}
}

func TestCollectPerformance_with_get_performance_data_error(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{}
	client := &centralizedstorage.CentralizedClient{}

	// mock
	applyFunc := gomonkey.ApplyFunc(GetPerformanceData, func(ctx context.Context,
		client *centralizedstorage.CentralizedClient, request *cmi.CollectRequest) ([]PerformanceIndicators, error) {
		return []PerformanceIndicators{}, errors.New("GetPerformanceData error")
	})
	defer applyFunc.Reset()

	// action
	_, err := CollectPerformance(context.Background(), client, request)

	// assert
	if err == nil {
		t.Error("TestGetNameMapping() want an GetPerformanceData error, but got nil")
	}
}

func TestCollectPerformance_with_get_mapping_error(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{}
	client := &centralizedstorage.CentralizedClient{}

	// mock
	applyFunc := gomonkey.
		ApplyFunc(GetPerformanceData, func(ctx context.Context,
			client *centralizedstorage.CentralizedClient, request *cmi.CollectRequest) ([]PerformanceIndicators, error) {
			return []PerformanceIndicators{{}}, nil
		}).
		ApplyFunc(GetMapping, func(ctx context.Context, collectType string,
			client *centralizedstorage.CentralizedClient) (map[string]string, error) {
			return nil, errors.New("GetMapping error")
		})
	defer applyFunc.Reset()

	// action
	_, err := CollectPerformance(context.Background(), client, request)

	// assert
	if err == nil {
		t.Error("TestGetNameMapping() want an GetMapping error, but got nil")
	}
}

func TestCollectPerformance_with_success(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{}
	client := &centralizedstorage.CentralizedClient{}

	// mock
	applyFunc := gomonkey.
		ApplyFunc(GetPerformanceData, func(ctx context.Context,
			client *centralizedstorage.CentralizedClient, request *cmi.CollectRequest) ([]PerformanceIndicators, error) {
			return []PerformanceIndicators{}, nil
		}).
		ApplyFunc(GetMapping, func(ctx context.Context, collectType string,
			client *centralizedstorage.CentralizedClient) (map[string]string, error) {
			return map[string]string{}, nil
		}).
		ApplyFunc(MergePerformance, func(performances []PerformanceIndicators, nameMapping map[string]string,
			request *cmi.CollectRequest) *cmi.CollectResponse {
			return &cmi.CollectResponse{}
		})
	defer applyFunc.Reset()

	// action
	_, err := CollectPerformance(context.Background(), client, request)

	// assert
	if err != nil {
		t.Errorf("TestGetNameMapping() error = %v", err)
	}
}

func TestGetPerformanceData_with_storage_V7_success(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{CollectType: constants.Filesystem}
	client := &centralizedstorage.CentralizedClient{}

	// mock
	applyFunc := gomonkey.
		ApplyFunc(utils.MapStringToInt, func(sources []string) []int {
			return []int{1, 2}
		}).
		ApplyMethodFunc(client, "GetSystemInfo", func(ctx context.Context) (
			map[string]interface{}, error) {
			return map[string]interface{}{
				"pointRelease": "V700R001C00",
			}, nil
		}).
		ApplyMethodFunc(client, "GetPerformanceByPost", func(ctx context.Context,
			objectType int, indicators []int) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{
					"indicators":       []int{1, 2},
					"indicator_values": []float64{0.0, 1.0},
					"object_id":        "1",
				},
			}, nil
		})
	defer applyFunc.Reset()

	// action
	_, err := GetPerformanceData(context.Background(), client, request)

	// assert
	if err != nil {
		t.Errorf("TestGetPerformanceData_with_storage_V7_success() error = %v", err)
	}
}

func TestGetPerformanceData_with_storage_greater_V612_success(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{CollectType: constants.Filesystem}
	client := &centralizedstorage.CentralizedClient{}

	// mock
	applyFunc := gomonkey.
		ApplyFunc(utils.MapStringToInt, func(sources []string) []int {
			return []int{1, 2}
		}).
		ApplyMethodFunc(client, "GetSystemInfo", func(ctx context.Context) (
			map[string]interface{}, error) {
			return map[string]interface{}{
				"pointRelease": "6.1.7",
			}, nil
		}).
		ApplyMethodFunc(client, "GetPerformanceByPost", func(ctx context.Context,
			objectType int, indicators []int) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{
					"indicators":       []int{1, 2},
					"indicator_values": []float64{0.0, 1.0},
					"object_id":        "1",
				},
			}, nil
		})
	defer applyFunc.Reset()

	// action
	_, err := GetPerformanceData(context.Background(), client, request)

	// assert
	if err != nil {
		t.Errorf("TestGetPerformanceData_with_storage_greater_V612_success() error = %v", err)
	}
}

func TestGetPerformanceData_with_storage_V610_success(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{CollectType: constants.Filesystem}
	client := &centralizedstorage.CentralizedClient{}

	// mock
	applyFunc := gomonkey.
		ApplyFunc(utils.MapStringToInt, func(sources []string) []int {
			return []int{1, 2}
		}).
		ApplyMethodFunc(client, "GetSystemInfo", func(ctx context.Context) (
			map[string]interface{}, error) {
			return map[string]interface{}{
				"pointRelease": "6.1.0",
			}, nil
		}).
		ApplyMethodFunc(client, "GetPerformance", func(ctx context.Context,
			objectType int, indicators []int) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{
					"indicators":       []int{1, 2},
					"indicator_values": []float64{0.0, 1.0},
					"object_id":        "1",
				},
			}, nil
		})
	defer applyFunc.Reset()

	// action
	_, err := GetPerformanceData(context.Background(), client, request)

	// assert
	if err != nil {
		t.Errorf("TestGetPerformanceData_with_storage_V610_success() error = %v", err)
	}
}

func TestGetPerformanceData_with_storage_V610_empty_return_success(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{CollectType: constants.Filesystem}
	client := &centralizedstorage.CentralizedClient{}

	// mock
	applyFunc := gomonkey.
		ApplyFunc(utils.MapStringToInt, func(sources []string) []int {
			return []int{1, 2}
		}).
		ApplyMethodFunc(client, "GetSystemInfo", func(ctx context.Context) (
			map[string]interface{}, error) {
			return map[string]interface{}{
				"pointRelease": "6.1.0",
			}, nil
		}).
		ApplyMethodFunc(client, "GetPerformance", func(ctx context.Context,
			objectType int, indicators []int) ([]map[string]interface{}, error) {
			return nil, nil
		})
	defer applyFunc.Reset()

	// action
	_, err := GetPerformanceData(context.Background(), client, request)

	// assert
	if err != nil {
		t.Errorf("TestGetPerformanceData_with_storage_v610_empty_return_success() error = %v", err)
	}
}

func TestGetPerformanceData_with_storage_V3orV5_success(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{CollectType: constants.Filesystem}
	client := &centralizedstorage.CentralizedClient{}

	// mock
	applyFunc := gomonkey.
		ApplyFunc(utils.MapStringToInt, func(sources []string) []int {
			return []int{1, 2}
		}).
		ApplyMethodFunc(client, "GetSystemInfo", func(ctx context.Context) (
			map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		}).
		ApplyMethodFunc(client, "GetPerformance", func(ctx context.Context,
			objectType int, indicators []int) ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{
					"indicators":       []int{1, 2},
					"indicator_values": []float64{0.0, 1.0},
					"object_id":        "1",
				},
			}, nil
		})
	defer applyFunc.Reset()

	// action
	_, err := GetPerformanceData(context.Background(), client, request)

	// assert
	if err != nil {
		t.Errorf("TestGetPerformanceData_with_storage_v610_empty_return_success() error = %v", err)
	}
}

func TestGetPerformanceData_with_collect_type_not_exist_error(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{CollectType: "not-exist"}
	client := &centralizedstorage.CentralizedClient{}

	// action
	_, err := GetPerformanceData(context.Background(), client, request)

	// assert
	if err == nil {
		t.Error("TestGetPerformanceData() want an error, but got nil")
	}
}

func TestMergePerformance(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{CollectType: "not-exist"}
	nameMapping := map[string]string{"1": "test-object"}
	performances := []PerformanceIndicators{
		{
			Indicators:      []int{1, 2},
			IndicatorValues: []float64{0.0, 1.1},
			ObjectId:        "1",
		},
	}

	// action
	response := MergePerformance(performances, nameMapping, request)

	// assert
	if len(response.GetDetails()) != 1 {
		t.Error("TestGetPerformanceData() want an error, but got nil")
	}
}

func TestPerformanceIndicators_ToMap(t *testing.T) {
	// arrange
	performances := PerformanceIndicators{
		Indicators:      []int{1, 2},
		IndicatorValues: []float64{0.0, 1.1},
		ObjectId:        "1",
	}

	want := map[string]string{
		"1": "0.0000",
		"2": "1.1000",
	}

	// action
	got := performances.ToMap()

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestGetPerformanceData() want = %v, but got = %v", want, got)
	}
}
