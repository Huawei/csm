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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPerformanceCollector_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{}, nil
	}

	// action
	gotCollector := NewPerformanceCollector(cache, discoverFunc)

	// assert
	assert.NotNil(t, gotCollector)
}

func TestPerformanceCollector_Collect_DiscoverClientFailed(t *testing.T) {
	// arrange
	wantErr := errors.New("discover client failed")
	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{}, wantErr
	}
	collector := NewPerformanceCollector(cache, discoverFunc)
	req := &CollectRequest{BackendName: "b1", CollectType: "lun", MetricsType: "performance"}

	// action
	_, gotErr := collector.Collect(context.Background(), req)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestPerformanceCollector_Collect_HandlerNotFound(t *testing.T) {
	// arrange
	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{StorageType: "nonexistent-perf-storage", Client: &mockRestClient{}}, nil
	}
	collector := NewPerformanceCollector(cache, discoverFunc)
	req := &CollectRequest{BackendName: "b1", CollectType: "lun", MetricsType: "performance"}
	wantErr := true

	// action
	_, gotErr := collector.Collect(context.Background(), req)

	// assert
	assert.Equal(t, wantErr, gotErr != nil)
	assert.Contains(t, gotErr.Error(), "nonexistent-perf-storage")
}

func TestPerformanceCollector_Collect_Success(t *testing.T) {
	// arrange
	wantBackendName := "perf-collect-test"
	handler := func(ctx context.Context, client RestClient, req *CollectRequest) (*CollectResponse, error) {
		return &CollectResponse{BackendName: wantBackendName}, nil
	}
	RegisterPerformanceHandlerDirect("perf-test-storage", "perf-test-collect", handler)

	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{StorageType: "perf-test-storage", Client: &mockRestClient{}}, nil
	}
	collector := NewPerformanceCollector(cache, discoverFunc)
	req := &CollectRequest{BackendName: "b1", CollectType: "perf-test-collect", MetricsType: "performance"}

	// action
	gotResp, gotErr := collector.Collect(context.Background(), req)

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantBackendName, gotResp.BackendName)
}

func TestPerformanceCollector_Collect_HandlerReturnsError(t *testing.T) {
	// arrange
	wantErr := errors.New("perf handler error")
	handler := func(ctx context.Context, client RestClient, req *CollectRequest) (*CollectResponse, error) {
		return nil, wantErr
	}
	RegisterPerformanceHandlerDirect("perf-err-storage", "perf-err-collect", handler)

	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{StorageType: "perf-err-storage", Client: &mockRestClient{}}, nil
	}
	collector := NewPerformanceCollector(cache, discoverFunc)
	req := &CollectRequest{BackendName: "b1", CollectType: "perf-err-collect", MetricsType: "performance"}

	// action
	_, gotErr := collector.Collect(context.Background(), req)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestMergePerformance_Success(t *testing.T) {
	// arrange
	performances := []PerformanceIndicators{
		{Indicators: []int{1, 2}, IndicatorValues: []float64{1.5, 2.5}, ObjectId: "obj1"},
		{Indicators: []int{1, 2}, IndicatorValues: []float64{3.5, 4.5}, ObjectId: "obj2"},
	}
	nameMapping := map[string]string{"obj1": "name1", "obj2": "name2"}
	req := &CollectRequest{BackendName: "b1", CollectType: "lun", MetricsType: "performance"}

	// action
	gotResp := MergePerformance(performances, nameMapping, req)

	// assert
	assert.Len(t, gotResp.Details, 2)
	assert.Equal(t, "name1", gotResp.Details[0].Data["ObjectName"])
	assert.Equal(t, "obj1", gotResp.Details[0].Data["ObjectId"])
}

func TestMergePerformance_ObjectIdNotInNameMapping(t *testing.T) {
	// arrange
	performances := []PerformanceIndicators{
		{Indicators: []int{1}, IndicatorValues: []float64{1.0}, ObjectId: "obj1"},
		{Indicators: []int{1}, IndicatorValues: []float64{2.0}, ObjectId: "obj2"},
	}
	nameMapping := map[string]string{"obj2": "name2"}
	req := &CollectRequest{BackendName: "b1", CollectType: "lun", MetricsType: "performance"}

	// action
	gotResp := MergePerformance(performances, nameMapping, req)

	// assert
	assert.Len(t, gotResp.Details, 1)
	assert.Equal(t, "obj2", gotResp.Details[0].Data["ObjectId"])
}

func TestPerformanceIndicators_ToMap_Success(t *testing.T) {
	// arrange
	pi := PerformanceIndicators{
		Indicators:      []int{1, 2},
		IndicatorValues: []float64{1.5, 2.5},
	}

	// action
	gotMap := pi.ToMap()

	// assert
	assert.Len(t, gotMap, 2)
	assert.Equal(t, "1.5", gotMap["1"])
	assert.Equal(t, "2.5", gotMap["2"])
}

func TestPerformanceIndicators_ToMap_EmptyIndicators(t *testing.T) {
	// arrange
	pi := PerformanceIndicators{
		Indicators:      []int{},
		IndicatorValues: []float64{},
	}

	// action
	gotMap := pi.ToMap()

	// assert
	assert.Empty(t, gotMap)
}

func TestPerformanceIndicators_ToMap_LengthMismatch(t *testing.T) {
	// arrange
	pi := PerformanceIndicators{
		Indicators:      []int{1, 2},
		IndicatorValues: []float64{1.5},
	}

	// action
	gotMap := pi.ToMap()

	// assert
	assert.Empty(t, gotMap)
}
