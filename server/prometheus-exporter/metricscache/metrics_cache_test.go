/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2026. All rights reserved.
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
	"github.com/stretchr/testify/assert"

	"github.com/huawei/csm/v2/provider/cmicore"
	"github.com/huawei/csm/v2/provider/constants"
	clientSet "github.com/huawei/csm/v2/server/prometheus-exporter/clientset"
)

func TestMetricsDataCache_GetMetricsData(t *testing.T) {
	// arrange
	mockCollectDetail := &cmicore.CollectDetail{
		Data: map[string]string{"fake_data": "test_data"},
	}
	mockCollectResponse := &cmicore.CollectResponse{
		BackendName: "fake_backend_name",
		CollectType: "fake_type",
		MetricsType: "fake_collector_name",
		Details:     []*cmicore.CollectDetail{mockCollectDetail},
	}
	mockMetricsData := &BaseMetricsData{
		BackendName:         "fake_backend_name",
		MetricsType:         "fake_collector_name",
		MetricsDataResponse: mockCollectResponse,
	}
	mockpMetricsDataCache := &MetricsDataCache{
		BackendName: "fake_name",
		CacheDataMap: map[string]MetricsData{
			"fake_collector_name": mockMetricsData},
	}

	// action
	got := mockpMetricsDataCache.GetMetricsData("fake_collector_name")

	// assert
	if !reflect.DeepEqual(got, mockCollectResponse) {
		t.Errorf("parseStorageData() got = %v, want %v", got, "fake_data")
	}
}

func TestMetricsDataCache_GetMetricsData_NotFound(t *testing.T) {
	// arrange
	mockpMetricsDataCache := &MetricsDataCache{
		BackendName:  "fake_name",
		CacheDataMap: map[string]MetricsData{},
	}

	// action
	got := mockpMetricsDataCache.GetMetricsData("non_existent")

	// assert
	if got != nil {
		t.Errorf("GetMetricsData() got = %v, want nil", got)
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
		Core: nil, // Not needed for this test since we're mocking SetMetricsData
	}
	ctx := context.Background()
	called := false

	// mock
	mock := gomonkey.NewPatches()
	mock.ApplyFuncReturn(clientSet.GetExporterClientSet, mockClientsSet)
	mock.ApplyPrivateMethod(mockStorageMetricsData, "GetMetricsDataResponse",
		func() *cmicore.CollectResponse {
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
	wantRes := map[string][]string{
		constants.Lun:        pvPerformanceMap[constants.Lun],
		constants.Filesystem: pvPerformanceMap[constants.Filesystem],
	}

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

func TestMetricsDataCache_buildPVBatchParams_ObjectOceanStorage(t *testing.T) {
	// arrange
	metricsDataCache := &MetricsDataCache{BackendName: "oceanstor-backend"}
	ctx := context.TODO()
	monitorType := "object"
	params := map[string][]string{"pv": {}}
	batchParams := make(map[string][]string)

	// mock getCollectTypesForBackend to return oceanStorage collect types
	mockCore := &cmicore.Core{}
	mockClientsSet := &clientSet.ClientsSet{Core: mockCore}
	wantRes := map[string][]string{"lun": {""}, "filesystem": {""}}

	p := gomonkey.NewPatches()
	p.ApplyFuncReturn(clientSet.GetExporterClientSet, mockClientsSet)
	p.ApplyMethodReturn(mockCore, "GetStorageType", constants.OceanStorage, nil)
	defer p.Reset()

	// action
	gotErr := metricsDataCache.buildPVBatchParams(ctx, monitorType, params, batchParams)

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantRes, batchParams)
}

func TestMetricsDataCache_buildPVBatchParams_ObjectFusionStorage(t *testing.T) {
	// arrange
	metricsDataCache := &MetricsDataCache{BackendName: "fusion-backend"}
	ctx := context.TODO()
	monitorType := "object"
	params := map[string][]string{"pv": {}}
	batchParams := make(map[string][]string)

	// mock getCollectTypesForBackend to return fusionStorage collect types
	mockCore := &cmicore.Core{}
	mockClientsSet := &clientSet.ClientsSet{Core: mockCore}
	wantRes := map[string][]string{"namespace": {""}}

	p := gomonkey.NewPatches()
	p.ApplyFuncReturn(clientSet.GetExporterClientSet, mockClientsSet)
	p.ApplyMethodReturn(mockCore, "GetStorageType", constants.FusionStorage, nil)
	defer p.Reset()

	// action
	gotErr := metricsDataCache.buildPVBatchParams(ctx, monitorType, params, batchParams)

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantRes, batchParams)
}

func TestMetricsDataCache_buildPVBatchParams_ObjectFallback(t *testing.T) {
	// arrange
	metricsDataCache := &MetricsDataCache{BackendName: "unknown-backend"}
	ctx := context.TODO()
	monitorType := "object"
	params := map[string][]string{"pv": {}}
	batchParams := make(map[string][]string)

	// mock getCollectTypesForBackend to return error, triggering fallback to all types
	p := gomonkey.NewPatches()
	p.ApplyFuncReturn(clientSet.GetExporterClientSet, &clientSet.ClientsSet{Core: nil})
	defer p.Reset()

	wantRes := map[string][]string{"lun": {""}, "filesystem": {""}, "namespace": {""}}

	// action
	gotErr := metricsDataCache.buildPVBatchParams(ctx, monitorType, params, batchParams)

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantRes, batchParams)
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

func TestStorageTypeMap_FusionStorageNas(t *testing.T) {
	// arrange & act & assert
	collectType, ok := storageTypeMap[constants.StorageFusionNas]
	assert.True(t, ok, "fusionstorage-nas should be in storageTypeMap")
	assert.Equal(t, constants.Namespace, collectType)
}

func TestPvPerformanceMap_Namespace(t *testing.T) {
	// arrange & act & assert
	indicators, ok := pvPerformanceMap[constants.Namespace]
	assert.True(t, ok, "namespace should be in pvPerformanceMap")
	assert.NotEmpty(t, indicators)

	// Verify it contains the NFS + DPC indicator IDs
	// NFS: 30001-30006,30011-30016,30076,30077 (14)
	// DPC: 30043-30046,30048,30049,30051,30052,31005 (9)
	// Total: 23 indicators
	allIndicators := ""
	for _, ind := range indicators {
		allIndicators += ind + ","
	}
	assert.Contains(t, allIndicators, "30001", "should contain NFS indicator")
	assert.Contains(t, allIndicators, "30043", "should contain DPC indicator")
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
