/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/csm/v2/storage/client"
	"github.com/huawei/csm/v2/storage/httpcode"
)

func TestCentralizedClient_GetPerformance(t *testing.T) {
	httpGet := MockHttpGet(mockGetresponse)
	defer httpGet.Reset()

	_, err := centralizedCli.GetPerformance(context.Background(), 40, []int{})
	if err != nil {
		t.Errorf("GetPerformance() error = %v,", err)
	}
}

func TestCentralizedClient_GetPerformanceByPost_RetrySuccess(t *testing.T) {
	// arrange
	mockRetryResponse := map[string]interface{}{
		"Error": map[string]interface{}{
			"code": httpcode.RetryCodes[0],
		},
		"Data": []map[string]interface{}{},
	}

	mockSuccessResponse := map[string]interface{}{
		"Error": map[string]interface{}{
			"code": 0,
		},
		"Data": []map[string]interface{}{},
	}

	// mock
	retryTimes := 0
	var cli *client.Client
	p := gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			if retryTimes < 2 {
				retryTimes++
				return mockRetryResponse, nil
			}
			return mockSuccessResponse, nil
		})

	// action
	_, err := centralizedCli.GetPerformanceByPost(context.Background(), 40, []int{})

	// assert
	if retryTimes != 2 || err != nil {
		t.Errorf("TestCentralizedClient_GetPerformanceByPost_RetrySuccess failed, "+
			"want retries = 2, actul retries = %d, want err = nil, got err = %v,", retryTimes, err)
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}
