/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2024-2026. All rights reserved.
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
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"github.com/huawei/csm/v2/provider/cmicore"
)

func TestMergePVMetricsData_mergeKubePVAndStorageInfo_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	storageNameKey := "storageNameKey"
	pvNameKey := "pvNameKey"
	volumeType := "lun" // Must match storageTypeMap["oceanstor-san"]

	mockName := "name"
	mockId := "001"
	pvCacheData := []*cmicore.CollectDetail{{Data: map[string]string{
		pvNameKey:        mockName,
		"field1":         "field1 context",
		"field2":         "field2 context",
		"sbcStorageType": "oceanstor-san", // Must map to "lun" in storageTypeMap
	}}}

	// Provide storage response data that will be returned by GetMetricsData
	storageResponse := &cmicore.CollectResponse{
		Details: []*cmicore.CollectDetail{{Data: map[string]string{
			storageNameKey: mockName,
			"ID":           mockId,
		}}},
	}
	storageMetricsData := &BaseMetricsData{
		MetricsDataResponse: storageResponse,
	}
	metricsDataCache := &MetricsDataCache{
		CacheDataMap: map[string]MetricsData{
			volumeType: storageMetricsData,
		},
	}
	mergePVMetricsData := &MergePVMetricsData{}

	wantRes := map[string]map[string]string{mockName + mockId: {
		pvNameKey:        mockName,
		"field1":         "field1 context",
		"field2":         "field2 context",
		"sbcStorageType": "oceanstor-san",
		storageNameKey:   mockName,
		"ID":             mockId,
		"sameName":       mockName,
	}}

	// action
	gotRes, gotErr := mergePVMetricsData.mergeKubePVAndStorageInfo(ctx, storageNameKey, pvNameKey,
		volumeType, pvCacheData, metricsDataCache)

	// assert
	if !reflect.DeepEqual(gotRes, wantRes) {
		t.Errorf("TestMergePVMetricsData_mergeKubePVAndStorageInfo_Success failed, "+
			"gotRes [%v], wantRes [%v]", gotRes, wantRes)
	}
	if gotErr != nil {
		t.Errorf("TestMergePVMetricsData_mergeKubePVAndStorageInfo_Success failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, nil)
	}
}

func TestMergePVMetricsData_mergeKubePVAndStorageInfo_GetPvDataFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	storageNameKey := "storageNameKey"
	pvNameKey := "pvNameKey"
	storageType := "storageType"
	var pvCacheData []*cmicore.CollectDetail
	metricsDataCache := &MetricsDataCache{}
	mergePVMetricsData := &MergePVMetricsData{}

	wantErr := fmt.Errorf("can not get the pv data when merge")

	// action
	gotRes, gotErr := mergePVMetricsData.mergeKubePVAndStorageInfo(ctx, storageNameKey, pvNameKey,
		storageType, pvCacheData, metricsDataCache)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestMergePVMetricsData_mergeKubePVAndStorageInfo_GetPvDataFailed failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, wantErr)
	}
	if gotRes != nil {
		t.Errorf("TestMergePVMetricsData_mergeKubePVAndStorageInfo_GetPvDataFailed failed, "+
			"gotRes [%v], wantErr [%v]", gotRes, nil)
	}
}

func TestMergePVMetricsData_mergeKubePVAndStorageInfo_GetStorageDataFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	storageNameKey := "storageNameKey"
	pvNameKey := "pvNameKey"
	storageType := "storageType"
	pvCacheData := []*cmicore.CollectDetail{{Data: map[string]string{pvNameKey: "name"}}}
	metricsDataCache := &MetricsDataCache{}
	mergePVMetricsData := &MergePVMetricsData{}

	wantErr := errors.New("can not get the storage data when merge")

	// mock
	p := gomonkey.NewPatches()
	p.ApplyMethod(reflect.TypeOf(metricsDataCache), "GetMetricsData",
		func(_ *MetricsDataCache, metricsType string) *cmicore.CollectResponse {
			return nil
		})

	// action
	gotRes, gotErr := mergePVMetricsData.mergeKubePVAndStorageInfo(ctx, storageNameKey, pvNameKey,
		storageType, pvCacheData, metricsDataCache)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestMergePVMetricsData_mergeKubePVAndStorageInfo_GetStorageDataFailed failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, wantErr)
	}
	if gotRes != nil {
		t.Errorf("TestMergePVMetricsData_mergeKubePVAndStorageInfo_GetPvDataFailed failed, "+
			"gotRes [%v], wantErr [%v]", gotRes, nil)
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestMergePVMetricsData_getPVMergeParams_PerformanceSuccess(t *testing.T) {
	// arrange
	indicator1 := "indicator1"
	indicator2 := "indicator2"
	mergePVMetricsData := &MergePVMetricsData{BaseMergeMetricsData: &BaseMergeMetricsData{
		monitorType:     "performance",
		mergeIndicators: []string{indicator1 + "," + indicator2},
	}}
	metricsDataCache := &MetricsDataCache{}

	wantKey := "ObjectName"
	wantList := []string{indicator1, indicator2}

	// action
	gotKey, gotList, gotErr := mergePVMetricsData.getPVMergeParams(metricsDataCache)

	// assert
	assert.Equal(t, wantKey, gotKey)
	assert.Equal(t, wantList, gotList)
	assert.NoError(t, gotErr)
}

func TestMergePVMetricsData_getPVMergeParams_ObjectSuccess(t *testing.T) {
	// arrange
	mergePVMetricsData := &MergePVMetricsData{BaseMergeMetricsData: &BaseMergeMetricsData{
		monitorType:     "object",
		mergeIndicators: nil,
	}}
	metricsDataCache := &MetricsDataCache{
		CacheDataMap: map[string]MetricsData{
			"pv":         &BaseMetricsData{},
			"lun":        &BaseMetricsData{},
			"filesystem": &BaseMetricsData{},
		},
	}

	wantKey := "NAME"
	// Object branch now derives from CacheDataMap keys, excluding "pv"
	wantList := []string{"lun", "filesystem"}

	// action
	gotKey, gotList, gotErr := mergePVMetricsData.getPVMergeParams(metricsDataCache)

	// assert
	assert.Equal(t, wantKey, gotKey)
	assert.NoError(t, gotErr)
	// Map iteration order is non-deterministic, so check containment instead of exact order
	assert.ElementsMatch(t, wantList, gotList)
}

func TestMergePVMetricsData_getPVMergeParams_ObjectSuccessFusionStorage(t *testing.T) {
	// arrange
	mergePVMetricsData := &MergePVMetricsData{BaseMergeMetricsData: &BaseMergeMetricsData{
		monitorType:     "object",
		mergeIndicators: nil,
	}}
	// FusionStorage backend only has "namespace" in CacheDataMap
	metricsDataCache := &MetricsDataCache{
		CacheDataMap: map[string]MetricsData{
			"pv":        &BaseMetricsData{},
			"namespace": &BaseMetricsData{},
		},
	}

	wantKey := "NAME"
	wantList := []string{"namespace"}

	// action
	gotKey, gotList, gotErr := mergePVMetricsData.getPVMergeParams(metricsDataCache)

	// assert
	assert.Equal(t, wantKey, gotKey)
	assert.Equal(t, wantList, gotList)
	assert.NoError(t, gotErr)
}

func TestMergePVMetricsData_getPVMergeParams_EmptyIndicatorFailed(t *testing.T) {
	// arrange
	mergePVMetricsData := &MergePVMetricsData{BaseMergeMetricsData: &BaseMergeMetricsData{
		backendName:     "",
		monitorType:     "performance",
		metricsType:     "",
		mergeIndicators: nil,
	}}
	metricsDataCache := &MetricsDataCache{}
	wantErr := fmt.Errorf("when get pv merge params, " +
		"the monitorType is performance but mergeIndicators is empty")
	var wantKey string
	var wantIndicators []string

	// action
	gotKey, gotIndicators, gotErr := mergePVMetricsData.getPVMergeParams(metricsDataCache)

	// assert
	assert.Equal(t, wantErr, gotErr)
	assert.Equal(t, wantKey, gotKey)
	assert.Equal(t, wantIndicators, gotIndicators)
}

func TestMergePVMetricsData_MergeData_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	pvCacheData := &BaseMetricsData{}
	metricsDataCache := &MetricsDataCache{CacheDataMap: map[string]MetricsData{
		"pv": pvCacheData,
	}}
	mergePVMetricsData := &MergePVMetricsData{}

	// mock
	p := gomonkey.NewPatches()
	p.ApplyPrivateMethod(mergePVMetricsData, "getPVMergeParams",
		func(_ *MergePVMetricsData, cache *MetricsDataCache) (string, []string, error) {
			return "performance", []string{"indicator1"}, nil
		}).ApplyMethod(reflect.TypeOf(pvCacheData), "GetMetricsDataResponse",
		func(_ *BaseMetricsData) *cmicore.CollectResponse {
			return &cmicore.CollectResponse{
				Details: []*cmicore.CollectDetail{{Data: map[string]string{}}},
			}
		}).ApplyPrivateMethod(mergePVMetricsData, "mergeKubePVAndStorageInfo",
		func(ctx context.Context, storageNameKey, pvNameKey, storageType string,
			pvCacheData []*cmicore.CollectDetail, metricsDataCache *MetricsDataCache) (
			map[string]map[string]string, error) {
			return nil, nil
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert
	assert.NoError(t, gotErr)

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestMergePVMetricsData_MergeData_GetParamsFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	pvCacheData := &BaseMetricsData{}
	metricsDataCache := &MetricsDataCache{CacheDataMap: map[string]MetricsData{
		"pv": pvCacheData,
	}}
	mergePVMetricsData := &MergePVMetricsData{}
	getParamsErr := fmt.Errorf("get merge params err")
	wantErr := fmt.Errorf("can not get pv merge params, err is [%w]", getParamsErr)

	// mock
	p := gomonkey.NewPatches()
	p.ApplyPrivateMethod(mergePVMetricsData, "getPVMergeParams",
		func(_ *MergePVMetricsData, cache *MetricsDataCache) (string, []string, error) {
			return "", nil, getParamsErr
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert
	assert.Equal(t, wantErr, gotErr)

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestMergePVMetricsData_MergeData_GetCacheFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	metricsDataCache := &MetricsDataCache{CacheDataMap: map[string]MetricsData{}}
	mergePVMetricsData := &MergePVMetricsData{}
	wantErr := fmt.Errorf("can not get pv cache data when MergePVAndStorageData")

	// mock
	p := gomonkey.NewPatches()
	p.ApplyPrivateMethod(mergePVMetricsData, "getPVMergeParams",
		func(_ *MergePVMetricsData, cache *MetricsDataCache) (string, []string, error) {
			return "performance", []string{"indicator1"}, nil
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert
	assert.Equal(t, wantErr, gotErr)

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestMergePVMetricsData_MergeData_GetMetricsFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	pvCacheData := &BaseMetricsData{}
	metricsDataCache := &MetricsDataCache{CacheDataMap: map[string]MetricsData{
		"pv": pvCacheData,
	}}
	mergePVMetricsData := &MergePVMetricsData{}
	wantErr := fmt.Errorf("can not get MetricsDataResponse data when MergePVAndStorageData")

	// mock
	p := gomonkey.NewPatches()
	p.ApplyPrivateMethod(mergePVMetricsData, "getPVMergeParams",
		func(_ *MergePVMetricsData, cache *MetricsDataCache) (string, []string, error) {
			return "performance", []string{"indicator1"}, nil
		}).ApplyMethod(reflect.TypeOf(pvCacheData), "GetMetricsDataResponse",
		func(_ *BaseMetricsData) *cmicore.CollectResponse {
			return nil
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert
	assert.Equal(t, wantErr, gotErr)

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestMergePVMetricsData_MergeData_GetMetricsDetailsFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	pvCacheData := &BaseMetricsData{}
	metricsDataCache := &MetricsDataCache{CacheDataMap: map[string]MetricsData{
		"pv": pvCacheData,
	}}
	mergePVMetricsData := &MergePVMetricsData{}
	wantErr := fmt.Errorf("can not get MetricsDataResponse.Details when MergePVAndStorageData")

	// mock
	p := gomonkey.NewPatches()
	p.ApplyPrivateMethod(mergePVMetricsData, "getPVMergeParams",
		func(_ *MergePVMetricsData, cache *MetricsDataCache) (string, []string, error) {
			return "performance", []string{"indicator1"}, nil
		}).ApplyMethod(reflect.TypeOf(pvCacheData), "GetMetricsDataResponse",
		func(_ *BaseMetricsData) *cmicore.CollectResponse {
			return &cmicore.CollectResponse{Details: nil}
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert
	assert.Equal(t, wantErr, gotErr)

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestMergePVMetricsData_MergeData_MergeFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	pvCacheData := &BaseMetricsData{}
	metricsDataCache := &MetricsDataCache{CacheDataMap: map[string]MetricsData{
		"pv": pvCacheData,
	}}
	mergePVMetricsData := &MergePVMetricsData{}
	mergeErr := fmt.Errorf("merge pv and storage info error")

	// mock
	p := gomonkey.NewPatches()
	p.ApplyPrivateMethod(mergePVMetricsData, "getPVMergeParams",
		func(_ *MergePVMetricsData, cache *MetricsDataCache) (string, []string, error) {
			return "performance", []string{"indicator1"}, nil
		}).ApplyMethod(reflect.TypeOf(pvCacheData), "GetMetricsDataResponse",
		func(_ *BaseMetricsData) *cmicore.CollectResponse {
			return &cmicore.CollectResponse{
				Details: []*cmicore.CollectDetail{{Data: map[string]string{}}},
			}
		}).ApplyPrivateMethod(mergePVMetricsData, "mergeKubePVAndStorageInfo",
		func(ctx context.Context, storageNameKey, pvNameKey, storageType string,
			pvCacheData []*cmicore.CollectDetail, metricsDataCache *MetricsDataCache) (
			map[string]map[string]string, error) {
			return nil, mergeErr
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert - mergeKubePVAndStorageInfo failure is logged as warning, not returned as error
	assert.NoError(t, gotErr)

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}
