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

func TestRegisterIndicatorMapping_Success(t *testing.T) {
	// arrange
	wantTypeID := 99
	defer delete(IndicatorsMapping, "test_type")

	// action
	RegisterIndicatorMapping("test_type", wantTypeID)

	// assert
	assert.Equal(t, wantTypeID, IndicatorsMapping["test_type"])
}

func TestBuildResponse_Success(t *testing.T) {
	// arrange
	req := &CollectRequest{BackendName: "b1", CollectType: "lun", MetricsType: "object"}

	// action
	gotResp := BuildResponse(req)

	// assert
	assert.Equal(t, "b1", gotResp.BackendName)
	assert.Equal(t, "lun", gotResp.CollectType)
	assert.Equal(t, "object", gotResp.MetricsType)
	assert.Empty(t, gotResp.Details)
}

func TestAddCollectDetailWithMap_Success(t *testing.T) {
	// arrange
	resp := BuildResponse(&CollectRequest{BackendName: "b1"})
	data := map[string]string{"ID": "1", "NAME": "lun1"}

	// action
	AddCollectDetailWithMap(data, resp)

	// assert
	assert.Len(t, resp.Details, 1)
	assert.Equal(t, "1", resp.Details[0].Data["ID"])
	assert.Equal(t, "lun1", resp.Details[0].Data["NAME"])
}

func TestBuildFailedPageResult_Success(t *testing.T) {
	// arrange
	wantErr := errors.New("query failed")

	// action
	gotTuple := BuildFailedPageResult(wantErr)

	// assert
	assert.ErrorIs(t, gotTuple.Error, wantErr)
	assert.Empty(t, gotTuple.Data)
}

func TestBuildSuccessPageResult_Success(t *testing.T) {
	// arrange
	data := []map[string]interface{}{
		{"ID": "1", "NAME": "lun1"},
		{"ID": "2", "NAME": "lun2"},
	}

	// action
	gotTuple := BuildSuccessPageResult(data)

	// assert
	assert.NoError(t, gotTuple.Error)
	assert.Len(t, gotTuple.Data, 2)
	assert.Equal(t, "1", gotTuple.Data[0]["ID"])
}

func TestConcurrentPaginate_SinglePage(t *testing.T) {
	// arrange - count=1 fits in one page regardless of pageSize
	countFunc := func(ctx context.Context) (int, error) {
		return 1, nil
	}
	pageFunc := func(ctx context.Context, start, end int) ([]map[string]interface{}, error) {
		return []map[string]interface{}{{"start": start, "end": end}}, nil
	}

	// action
	gotResult, gotErr := ConcurrentPaginate(context.Background(), countFunc, pageFunc)

	// assert
	assert.NoError(t, gotErr)
	assert.Len(t, gotResult, 1)
	assert.Equal(t, 0, gotResult[0]["start"])
}

func TestConcurrentPaginate_CountFailed(t *testing.T) {
	// arrange
	wantErr := errors.New("count failed")
	countFunc := func(ctx context.Context) (int, error) {
		return 0, wantErr
	}

	// action
	_, gotErr := ConcurrentPaginate(context.Background(), countFunc, nil)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestConcurrentPaginate_CountZero(t *testing.T) {
	// arrange
	countFunc := func(ctx context.Context) (int, error) {
		return 0, nil
	}
	pageFunc := func(ctx context.Context, start, end int) ([]map[string]interface{}, error) {
		return nil, nil
	}

	// action
	gotResult, gotErr := ConcurrentPaginate(context.Background(), countFunc, pageFunc)

	// assert
	assert.NoError(t, gotErr)
	assert.Empty(t, gotResult)
}

func TestConcurrentPaginate_QueryFailed(t *testing.T) {
	// arrange - count=1 creates one pageQuery that fails
	wantErr := errors.New("query failed")
	countFunc := func(ctx context.Context) (int, error) {
		return 1, nil
	}
	pageFunc := func(ctx context.Context, start, end int) ([]map[string]interface{}, error) {
		return nil, wantErr
	}

	// action
	_, gotErr := ConcurrentPaginate(context.Background(), countFunc, pageFunc)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestReadQueryResult_MultiplePages(t *testing.T) {
	// arrange - simulate multi-page aggregation via channel directly
	ch := make(chan PageResultTuple, 3)
	ch <- BuildSuccessPageResult([]map[string]interface{}{{"page": 1}})
	ch <- BuildSuccessPageResult([]map[string]interface{}{{"page": 2}})
	ch <- BuildSuccessPageResult([]map[string]interface{}{{"page": 3}})
	close(ch)

	// action
	gotResult, gotErr := ReadQueryResult(ch)

	// assert
	assert.NoError(t, gotErr)
	assert.Len(t, gotResult, 3)
}

func TestReadQueryResult_ErrorInStream(t *testing.T) {
	// arrange
	wantErr := errors.New("stream error")
	ch := make(chan PageResultTuple, 2)
	ch <- BuildFailedPageResult(wantErr)
	close(ch)

	// action
	_, gotErr := ReadQueryResult(ch)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestReadQueryResult_EmptyChannel(t *testing.T) {
	// arrange
	ch := make(chan PageResultTuple)
	close(ch)

	// action
	gotResult, gotErr := ReadQueryResult(ch)

	// assert
	assert.NoError(t, gotErr)
	assert.Nil(t, gotResult)
}

func TestConvertMapToResponse_Success(t *testing.T) {
	// arrange
	data := []map[string]interface{}{
		{"ID": "1", "NAME": "lun1", "CAPACITY": 1024},
		{"ID": "2", "NAME": "lun2", "CAPACITY": 2048},
	}
	req := &CollectRequest{BackendName: "b1", CollectType: "lun", MetricsType: "object"}

	// action
	gotResp := ConvertMapToResponse(data, req)

	// assert
	assert.Equal(t, "b1", gotResp.BackendName)
	assert.Len(t, gotResp.Details, 2)
	assert.Equal(t, "1", gotResp.Details[0].Data["ID"])
	assert.Equal(t, "1024", gotResp.Details[0].Data["CAPACITY"])
}

func TestConvertMapToResponse_NilValueSkipped(t *testing.T) {
	// arrange
	data := []map[string]interface{}{
		{"ID": "1", "NAME": nil},
	}
	req := &CollectRequest{BackendName: "b1", CollectType: "test", MetricsType: "object"}

	// action
	gotResp := ConvertMapToResponse(data, req)

	// assert
	assert.Len(t, gotResp.Details, 1)
	_, ok := gotResp.Details[0].Data["NAME"]
	assert.False(t, ok, "nil value should be skipped")
	assert.Equal(t, "1", gotResp.Details[0].Data["ID"])
}

func TestConvertMapToResponse_EmptyData(t *testing.T) {
	// arrange
	data := []map[string]interface{}{}
	req := &CollectRequest{BackendName: "b1", CollectType: "lun", MetricsType: "object"}

	// action
	gotResp := ConvertMapToResponse(data, req)

	// assert
	assert.Empty(t, gotResp.Details)
}
