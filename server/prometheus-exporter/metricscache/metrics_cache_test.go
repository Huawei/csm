/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

package metricscache

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	clientSet "github.com/huawei/csm/v2/server/prometheus-exporter/clientset"
)

func TestMetricsDataCache_GetMetricsData(t *testing.T) {
	// arrange
	mockCollectDetail := cmi.CollectDetail{
		Data: map[string]string{"fake_data": "test_data"},
	}
	mockCollectResponse := cmi.CollectResponse{
		BackendName: "fake_backend_name",
		CollectType: "fake_type",
		MetricsType: "fake_collector_name",
		Details:     []*cmi.CollectDetail{&mockCollectDetail},
	}
	mockMetricsData := BaseMetricsData{
		MetricsType:         "fake_collector_name",
		MetricsDataResponse: &mockCollectResponse,
	}
	mockpMetricsDataCache := &MetricsDataCache{
		BackendName: "fake_name",
		CacheDataMap: map[string]MetricsData{
			"fake_collector_name": &mockMetricsData},
	}

	// action
	got := mockpMetricsDataCache.GetMetricsData("fake_collector_name")

	// assert
	if !reflect.DeepEqual(got, &mockCollectResponse) {
		t.Errorf("parseStorageData() got = %v, want %v", got, "fake_data")
	}
}

func TestMetricsDataCache_SetBatchDataFromSource(t *testing.T) {
	// arrange
	mockStorageMetricsData := &BaseMetricsData{BackendName: "fake_backend_name"}
	mockMetricsDataCache := &MetricsDataCache{
		BackendName:  "fake_name",
		CacheDataMap: map[string]MetricsData{"fake_metrics": mockStorageMetricsData},
	}
	mockClientsSet := &clientSet.ClientsSet{
		StorageGRPCClientSet: &cmi.ClientSet{}}
	ctx := context.Background()
	called := false

	// mock
	mock := gomonkey.NewPatches()
	mock.ApplyFunc(clientSet.GetExporterClientSet, func() *clientSet.ClientsSet {
		return mockClientsSet
	}).ApplyPrivateMethod(mockStorageMetricsData, "GetMetricsDataResponse",
		func() *cmi.CollectResponse {
			return nil
		}).ApplyPrivateMethod(mockStorageMetricsData, "SetMetricsData",
		func(ctx context.Context, collectorName, monitorType string, metricsIndicators []string) error {
			called = true
			return nil
		})

	// action
	mockMetricsDataCache.SetBatchDataFromSource(ctx, "fake_type",
		map[string][]string{"fake_metrics": {"fake_data"}})
	// assert
	if called != true {
		t.Errorf("SetBatchDataFromSource() got = %v, want true", called)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestMetricsDataCache_buildPVBatchParams_PerformanceSuccess(t *testing.T) {
	// arrange
	metricsDataCache := &MetricsDataCache{}
	ctx := context.TODO()
	monitorType := "performance"
	params := map[string][]string{"pv": {"lun,filesystem"}}
	batchParams := make(map[string][]string)
	wantRes := pvPerformanceMap

	// action
	gotErr := metricsDataCache.buildPVBatchParams(ctx, monitorType, params, batchParams)

	// assert
	if gotErr != nil {
		t.Errorf("TestMetricsDataCache_buildPVBatchParams_PerformanceSuccess failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, nil)
	}
	if !reflect.DeepEqual(batchParams, wantRes) {
		t.Errorf("TestMetricsDataCache_buildPVBatchParams_PerformanceSuccess failed, "+
			"gotRes [%v], wantRes [%v]", batchParams, wantRes)
	}

}

func TestMetricsDataCache_buildPVBatchParams_ObjectSuccess(t *testing.T) {
	// arrange
	metricsDataCache := &MetricsDataCache{}
	ctx := context.TODO()
	monitorType := "object"
	params := map[string][]string{"pv": {}}
	batchParams := make(map[string][]string)
	wantRes := map[string][]string{"lun": {""}, "filesystem": {""}}

	// action
	gotErr := metricsDataCache.buildPVBatchParams(ctx, monitorType, params, batchParams)

	// assert
	if gotErr != nil {
		t.Errorf("TestMetricsDataCache_buildPVBatchParams_ObjectSuccess failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, nil)
	}
	if !reflect.DeepEqual(batchParams, wantRes) {
		t.Errorf("TestMetricsDataCache_buildPVBatchParams_ObjectSuccess failed, "+
			"gotRes [%v], wantRes [%v]", batchParams, wantRes)
	}

}

func TestMetricsDataCache_buildPVBatchParams_GetIndicatorsFail(t *testing.T) {
	// arrange
	metricsDataCache := &MetricsDataCache{}
	ctx := context.TODO()
	monitorType := "performance"
	params := map[string][]string{}
	batchParams := make(map[string][]string)

	// action
	gotErr := metricsDataCache.buildPVBatchParams(ctx, monitorType, params, batchParams)

	// assert
	if gotErr != nil {
		t.Errorf("TestMetricsDataCache_buildPVBatchParams_GetIndicatorsFail failed, "+
			"gotErr [%v], wantErr [nil]", gotErr)
	}

}

func TestMetricsDataCache_buildPVBatchParams_EmptyIndicatorsFail(t *testing.T) {
	// arrange
	metricsDataCache := &MetricsDataCache{}
	ctx := context.TODO()
	monitorType := "performance"
	params := map[string][]string{"pv": {}}
	batchParams := make(map[string][]string)
	wantErr := fmt.Errorf("pv indicators [%v] are invalid with performance metrics type", params["pv"])

	// action
	gotErr := metricsDataCache.buildPVBatchParams(ctx, monitorType, params, batchParams)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestMetricsDataCache_buildPVBatchParams_EmptyIndicatorsFail failed, "+
			"gotRes [%v], wantRes [%v]", gotErr, wantErr)
	}

}
