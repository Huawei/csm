/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package centralizedstorage

import (
	"context"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/csm/v2/storage/client"
	"github.com/huawei/csm/v2/storage/utils"
)

func TestGetThenSuccess(t *testing.T) {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code": float64(0),
		},
		"data": []map[string]interface{}{},
	}

	var cli *client.Client
	httpGet := gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return response, nil
		})
	defer httpGet.Reset()

	centralizedCli := &CentralizedClient{
		Client: client.Client{
			Semaphore: utils.NewSemaphore(3),
		},
	}
	_, err := centralizedCli.get(ctx, "urlTest", nil)
	if err != nil {
		t.Errorf("get() error: %v", err)
	}
}

func TestGetResultFromResponseList_MoreThanOneData(t *testing.T) {
	response := &Response{
		Error: map[string]interface{}{
			"code": float64(0),
		},
		Data: []interface{}{
			map[string]interface{}{"key": "value1"},
			map[string]interface{}{"key": "value2"},
		},
	}

	centralizedCli := &CentralizedClient{}
	_, _, err := centralizedCli.getResultFromResponseList(ctx, response)
	if err == nil {
		t.Errorf("getResultFromResponseList() expected error for more than one data")
	}
}

func TestGetResultFromResponseList_CheckResponseCodeError(t *testing.T) {
	response := &Response{
		Error: map[string]interface{}{
			"code": float64(1),
		},
	}

	centralizedCli := &CentralizedClient{}
	_, _, err := centralizedCli.getResultFromResponseList(ctx, response)
	if err == nil {
		t.Errorf("getResultFromResponseList() expected error from checkResponseCode")
	}
}

func TestCheckResponseCode_CodeNotExist(t *testing.T) {
	response := &Response{
		Error: map[string]interface{}{},
	}

	centralizedCli := &CentralizedClient{}
	_, err := centralizedCli.checkResponseCode(ctx, response)
	if err == nil {
		t.Errorf("checkResponseCode() expected error for code not exist")
	}
}
