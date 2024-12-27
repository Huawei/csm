/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
)

func TestMergePVMetricsData_mergeKubePVAndStorageInfo_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	storageNameKey := "storageNameKey"
	pvNameKey := "pvNameKey"
	storageType := "storageType"
	metricsDataCache := &MetricsDataCache{}
	mergePVMetricsData := &MergePVMetricsData{}

	mockName := "name"
	mockId := "001"
	pvCacheData := []*storageGRPC.CollectDetail{{Data: map[string]string{
		pvNameKey: mockName,
		"field1":  "field1 context",
		"field2":  "field2 context",
	}}}

	wantRes := map[string]map[string]string{mockName + mockId: {
		pvNameKey:      mockName,
		"field1":       "field1 context",
		"field2":       "field2 context",
		storageNameKey: mockName,
		"ID":           mockId,
		"sameName":     mockName,
	}}

	// mock
	p := gomonkey.NewPatches()
	p.ApplyMethod(reflect.TypeOf(metricsDataCache), "GetMetricsData",
		func(_ *MetricsDataCache, metricsType string) *storageGRPC.CollectResponse {
			return &storageGRPC.CollectResponse{
				Details: []*storageGRPC.CollectDetail{{Data: map[string]string{
					storageNameKey: mockName,
					"ID":           mockId,
				}}},
			}
		})

	// action
	gotRes, gotErr := mergePVMetricsData.mergeKubePVAndStorageInfo(ctx, storageNameKey, pvNameKey,
		storageType, pvCacheData, metricsDataCache)

	// assert
	if !reflect.DeepEqual(gotRes, wantRes) {
		t.Errorf("TestMergePVMetricsData_mergeKubePVAndStorageInfo_Success failed, "+
			"gotRes [%v], wantRes [%v]", gotRes, wantRes)
	}
	if gotErr != nil {
		t.Errorf("TestMergePVMetricsData_mergeKubePVAndStorageInfo_Success failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, nil)
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestMergePVMetricsData_mergeKubePVAndStorageInfo_GetPvDataFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	storageNameKey := "storageNameKey"
	pvNameKey := "pvNameKey"
	storageType := "storageType"
	var pvCacheData []*storageGRPC.CollectDetail
	metricsDataCache := &MetricsDataCache{}
	mergePVMetricsData := &MergePVMetricsData{}

	getPvErr := fmt.Errorf("can not get the pv data when merge")
	wantErr := getPvErr

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
	pvCacheData := []*storageGRPC.CollectDetail{{Data: map[string]string{pvNameKey: "name"}}}
	metricsDataCache := &MetricsDataCache{}
	mergePVMetricsData := &MergePVMetricsData{}

	getStorageErr := fmt.Errorf("can not get the storage data when merge")
	wantErr := getStorageErr

	// mock
	p := gomonkey.NewPatches()
	p.ApplyMethod(reflect.TypeOf(metricsDataCache), "GetMetricsData",
		func(_ *MetricsDataCache, metricsType string) *storageGRPC.CollectResponse {
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
	ctx := context.TODO()
	indicator1 := "indicator1"
	indicator2 := "indicator2"
	mergePVMetricsData := &MergePVMetricsData{BaseMergeMetricsData: &BaseMergeMetricsData{
		monitorType:     "performance",
		mergeIndicators: []string{indicator1 + "," + indicator2},
	}}

	wantKey := "ObjectName"
	wantList := []string{indicator1, indicator2}

	// action
	gotKey, gotList, gotErr := mergePVMetricsData.getPVMergeParams(ctx)

	// assert
	if !reflect.DeepEqual(gotKey, wantKey) {
		t.Errorf("TestMergePVMetricsData_getPVMergeParams_PerformanceSuccess failed, "+
			"gotKey [%v], wantKey [%v]", gotKey, wantKey)
	}
	if !reflect.DeepEqual(gotList, wantList) {
		t.Errorf("TestMergePVMetricsData_getPVMergeParams_PerformanceSuccess failed, "+
			"gotList [%v], wantList [%v]", gotList, wantList)
	}
	if gotErr != nil {
		t.Errorf("TestMergePVMetricsData_getPVMergeParams_PerformanceSuccess failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, nil)
	}
}

func TestMergePVMetricsData_getPVMergeParams_ObjectSuccess(t *testing.T) {
	// arrange
	ctx := context.TODO()
	mergePVMetricsData := &MergePVMetricsData{BaseMergeMetricsData: &BaseMergeMetricsData{
		monitorType:     "object",
		mergeIndicators: nil,
	}}

	wantKey := "NAME"
	wantList := []string{"lun", "filesystem"}

	// action
	gotKey, gotList, gotErr := mergePVMetricsData.getPVMergeParams(ctx)

	// assert
	if !reflect.DeepEqual(gotKey, wantKey) {
		t.Errorf("TestMergePVMetricsData_getPVMergeParams_ObjectSuccess failed, "+
			"gotKey [%v], wantKey [%v]", gotKey, wantKey)
	}
	if !reflect.DeepEqual(gotList, wantList) {
		t.Errorf("TestMergePVMetricsData_getPVMergeParams_ObjectSuccess failed, "+
			"gotList [%v], wantList [%v]", gotList, wantList)
	}
	if gotErr != nil {
		t.Errorf("TestMergePVMetricsData_getPVMergeParams_ObjectSuccess failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, nil)
	}
}

func TestMergePVMetricsData_getPVMergeParams_EmptyIndicatorFailed(t *testing.T) {
	// arrange
	ctx := context.TODO()
	mergePVMetricsData := &MergePVMetricsData{BaseMergeMetricsData: &BaseMergeMetricsData{
		backendName:     "",
		monitorType:     "performance",
		metricsType:     "",
		mergeIndicators: nil,
	}}
	emptyErr := fmt.Errorf("when get pv merge params, " +
		"the monitorType is performance but mergeIndicators is empty")
	wantErr := emptyErr
	var wantKey string
	var wantIndicators []string

	// action
	gotKey, gotIndicators, gotErr := mergePVMetricsData.getPVMergeParams(ctx)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestMergePVMetricsData_getPVMergeParams_EmptyIndicatorFailed failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, wantErr)
	}
	if !reflect.DeepEqual(gotKey, wantKey) {
		t.Errorf("TestMergePVMetricsData_getPVMergeParams_EmptyIndicatorFailed failed, "+
			"gotKey [%v], wantKey [%v]", gotKey, wantKey)
	}
	if !reflect.DeepEqual(gotIndicators, wantIndicators) {
		t.Errorf("TestMergePVMetricsData_getPVMergeParams_EmptyIndicatorFailed failed, "+
			"gotIndicators [%v], wantIndicators [%v]", gotIndicators, wantIndicators)
	}
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
		func(ctx context.Context) (string, []string, error) {
			return "performance", []string{"indicator1"}, nil
		}).ApplyMethod(reflect.TypeOf(pvCacheData), "GetMetricsDataResponse",
		func(_ *BaseMetricsData) *storageGRPC.CollectResponse {
			return &storageGRPC.CollectResponse{
				Details: []*storageGRPC.CollectDetail{{Data: map[string]string{}}},
			}
		}).ApplyPrivateMethod(mergePVMetricsData, "mergeKubePVAndStorageInfo",
		func(ctx context.Context, storageNameKey, pvNameKey, storageType string,
			pvCacheData []*storageGRPC.CollectDetail, metricsDataCache *MetricsDataCache) (
			map[string]map[string]string, error) {
			return nil, nil
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert
	if gotErr != nil {
		t.Errorf("TestMergePVMetricsData_MergeData_Success failed, gotErr [%v], wantErr [%v]", gotErr, nil)
	}

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
	wantErr := getParamsErr

	// mock
	p := gomonkey.NewPatches()
	p.ApplyPrivateMethod(mergePVMetricsData, "getPVMergeParams",
		func(ctx context.Context) (string, []string, error) {
			return "", nil, getParamsErr
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestMergePVMetricsData_MergeData_GetParamsFailed failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, wantErr)
	}

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
	getCacheErr := fmt.Errorf("can not get pv cache data when MergePVAndStorageData")
	wantErr := getCacheErr

	// mock
	p := gomonkey.NewPatches()
	p.ApplyPrivateMethod(mergePVMetricsData, "getPVMergeParams",
		func(ctx context.Context) (string, []string, error) {
			return "performance", []string{"indicator1"}, nil
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestMergePVMetricsData_MergeData_GetCacheFailed failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, wantErr)
	}

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
	getMetricsErr := fmt.Errorf("can not get MetricsDataResponse data when MergePVAndStorageData")
	wantErr := getMetricsErr

	// mock
	p := gomonkey.NewPatches()
	p.ApplyPrivateMethod(mergePVMetricsData, "getPVMergeParams",
		func(ctx context.Context) (string, []string, error) {
			return "performance", []string{"indicator1"}, nil
		}).ApplyMethod(reflect.TypeOf(pvCacheData), "GetMetricsDataResponse",
		func(_ *BaseMetricsData) *storageGRPC.CollectResponse {
			return nil
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestMergePVMetricsData_MergeData_GetMetricsFailed failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, wantErr)
	}

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
	getMetricsDetailsErr := fmt.Errorf("can not get  MetricsDataResponse.Details when MergePVAndStorageData")
	wantErr := getMetricsDetailsErr

	// mock
	p := gomonkey.NewPatches()
	p.ApplyPrivateMethod(mergePVMetricsData, "getPVMergeParams",
		func(ctx context.Context) (string, []string, error) {
			return "performance", []string{"indicator1"}, nil
		}).ApplyMethod(reflect.TypeOf(pvCacheData), "GetMetricsDataResponse",
		func(_ *BaseMetricsData) *storageGRPC.CollectResponse {
			return &storageGRPC.CollectResponse{Details: nil}
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestMergePVMetricsData_MergeData_GetMetricsDetailsFailed failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, wantErr)
	}

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
	wantErr := mergeErr

	// mock
	p := gomonkey.NewPatches()
	p.ApplyPrivateMethod(mergePVMetricsData, "getPVMergeParams",
		func(ctx context.Context) (string, []string, error) {
			return "performance", []string{"indicator1"}, nil
		}).ApplyMethod(reflect.TypeOf(pvCacheData), "GetMetricsDataResponse",
		func(_ *BaseMetricsData) *storageGRPC.CollectResponse {
			return &storageGRPC.CollectResponse{
				Details: []*storageGRPC.CollectDetail{{Data: map[string]string{}}},
			}
		}).ApplyPrivateMethod(mergePVMetricsData, "mergeKubePVAndStorageInfo",
		func(ctx context.Context, storageNameKey, pvNameKey, storageType string,
			pvCacheData []*storageGRPC.CollectDetail, metricsDataCache *MetricsDataCache) (
			map[string]map[string]string, error) {
			return nil, mergeErr
		})

	// action
	gotErr := mergePVMetricsData.MergeData(ctx, metricsDataCache)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestMergePVMetricsData_MergeData_MergeFailed failed, "+
			"gotErr [%v], wantErr [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}
