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

package centralizedstorage

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"github.com/huawei/csm/v2/provider/constants"
	centralizedstorage "github.com/huawei/csm/v2/storage/api/centralizedstorage"
	"github.com/huawei/csm/v2/storage/client"
)

func TestCentralizedClient_GetSingleByUrlKey_Success(t *testing.T) {
	// arrange
	wantData := map[string]interface{}{"ID": "1", "NAME": "test"}
	mockResponse := map[string]interface{}{
		"Error": map[string]interface{}{"code": float64(0)},
		"Data":  wantData,
	}

	patches := gomonkey.NewPatches()
	var cli *client.Client
	patches.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return mockResponse, nil
		})
	defer patches.Reset()

	// action
	gotData, gotErr := centralizedCli.GetSingleByUrlKey(context.Background(), "GetSystemInfo")

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantData, gotData)
}

func TestCentralizedClient_GetSingleByUrlKey_GenerateUrlFailed(t *testing.T) {
	// arrange
	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(centralizedstorage.GenerateUrl, "", errors.New("invalid url key"))
	defer patches.Reset()

	// action
	gotData, gotErr := centralizedCli.GetSingleByUrlKey(context.Background(), "InvalidKey")

	// assert
	assert.Error(t, gotErr)
	assert.Nil(t, gotData)
}

func TestCentralizedClient_GetListByUrlKey_Success(t *testing.T) {
	// arrange
	wantData := []map[string]interface{}{{"ID": "1"}, {"ID": "2"}}
	mockResponse := map[string]interface{}{
		"Error": map[string]interface{}{"code": float64(0)},
		"Data":  []interface{}{map[string]interface{}{"ID": "1"}, map[string]interface{}{"ID": "2"}},
	}

	patches := gomonkey.NewPatches()
	var cli *client.Client
	patches.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return mockResponse, nil
		})
	defer patches.Reset()

	// action
	gotData, gotErr := centralizedCli.GetListByUrlKey(context.Background(), "GetStoragePools")

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantData, gotData)
}

func TestCentralizedClient_GetCountByUrlKey_Success(t *testing.T) {
	// arrange
	wantCount := 10
	httpGet := MockHttpGet(mockCountResponse)
	defer httpGet.Reset()

	// action
	gotCount, gotErr := centralizedCli.GetCountByUrlKey(context.Background(), "GetFilesystemCount")

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantCount, gotCount)
}

func TestCentralizedClient_GetPageByUrlKey_Success(t *testing.T) {
	// arrange
	httpGet := MockHttpGet(mockGetresponse)
	defer httpGet.Reset()

	// action
	gotData, gotErr := centralizedCli.GetPageByUrlKey(context.Background(), "GetFilesystem", 0, 100)

	// assert
	assert.NoError(t, gotErr)
	assert.Nil(t, gotData)
}

func TestCentralizedClient_QueryPerformanceData_PostModeSuccess(t *testing.T) {
	// arrange
	mockStorageInfo := map[string]interface{}{"pointRelease": "6.1.2"}
	wantData := []map[string]interface{}{{"object_id": "1"}}

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&CentralizedClient{}, "GetSystemInfo", mockStorageInfo, nil)
	patches.ApplyMethodReturn(&CentralizedClient{}, "GetPerformanceByPost", wantData, nil)
	defer patches.Reset()

	// action
	gotData, gotErr := centralizedCli.QueryPerformanceData(context.Background(), 11, []string{"1"})

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantData, gotData)
}

func TestCentralizedClient_QueryPerformanceData_GetSystemInfoFailed(t *testing.T) {
	// arrange
	wantErr := errors.New("get system info failed")

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&CentralizedClient{}, "GetSystemInfo", nil, wantErr)
	defer patches.Reset()

	// action
	gotData, gotErr := centralizedCli.QueryPerformanceData(context.Background(), 11, []string{"1"})

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
	assert.Nil(t, gotData)
}

func TestCentralizedClient_QueryPerformanceData_GetPerformanceFailed(t *testing.T) {
	// arrange
	mockStorageInfo := map[string]interface{}{} // no pointRelease -> V3/V5 mode
	wantErr := errors.New("get performance failed")

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&CentralizedClient{}, "GetSystemInfo", mockStorageInfo, nil)
	patches.ApplyMethodReturn(&CentralizedClient{}, "GetPerformance", nil, wantErr)
	defer patches.Reset()

	// action
	gotData, gotErr := centralizedCli.QueryPerformanceData(context.Background(), 11, []string{"1"})

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
	assert.Nil(t, gotData)
}

func TestCentralizedClient_QueryPerformanceData_GetPerformanceByPostFailed(t *testing.T) {
	// arrange
	mockStorageInfo := map[string]interface{}{"pointRelease": constants.MinVersionSupportPost}
	wantErr := errors.New("post performance failed")

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&CentralizedClient{}, "GetSystemInfo", mockStorageInfo, nil)
	patches.ApplyMethodReturn(&CentralizedClient{}, "GetPerformanceByPost", nil, wantErr)
	defer patches.Reset()

	// action
	gotData, gotErr := centralizedCli.QueryPerformanceData(context.Background(), 11, []string{"1"})

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
	assert.Nil(t, gotData)
}

func TestCentralizedClient_QueryPerformanceData_GetModeRetrySuccess(t *testing.T) {
	// arrange
	mockStorageInfo := map[string]interface{}{} // no pointRelease -> V3/V5 mode
	callCount := 0
	wantData := []map[string]interface{}{{"object_id": "1"}}

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&CentralizedClient{}, "GetSystemInfo", mockStorageInfo, nil)
	patches.ApplyMethod(reflect.TypeOf(&CentralizedClient{}), "GetPerformance",
		func(_ *CentralizedClient, _ context.Context, _ int, _ []int) ([]map[string]interface{}, error) {
			callCount++
			if callCount < 2 {
				return nil, nil
			}
			return wantData, nil
		})
	patches.ApplyFuncReturn(time.Sleep)
	defer patches.Reset()

	// action
	gotData, gotErr := centralizedCli.QueryPerformanceData(context.Background(), 11, []string{"1"})

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantData, gotData)
}

func TestCentralizedClient_QueryPerformanceData_AllRetriesReturnEmpty(t *testing.T) {
	// arrange
	mockStorageInfo := map[string]interface{}{} // no pointRelease -> V3/V5 mode

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&CentralizedClient{}, "GetSystemInfo", mockStorageInfo, nil)
	patches.ApplyMethodReturn(&CentralizedClient{}, "GetPerformance", []map[string]interface{}{}, nil)
	patches.ApplyFuncReturn(time.Sleep)
	defer patches.Reset()

	// action
	gotData, gotErr := centralizedCli.QueryPerformanceData(context.Background(), 11, []string{"1"})

	// assert
	assert.NoError(t, gotErr)
	assert.Empty(t, gotData)
}
