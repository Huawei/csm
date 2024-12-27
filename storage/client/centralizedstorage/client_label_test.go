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

package centralizedstorage

import (
	"context"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/csm/v2/storage/client"
	"github.com/huawei/csm/v2/storage/httpcode/label"
	"github.com/huawei/csm/v2/storage/utils"
)

func Test_CentralizedClient_CreateLabel(t *testing.T) {
	// arrange
	var cli *client.Client
	var centralizedCli = &CentralizedClient{
		Client: client.Client{Semaphore: utils.NewSemaphore(3)},
	}
	var data = map[string]interface{}{}

	// mock
	gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"Error": map[string]interface{}{
					"code": 0,
				},
				"Data": map[string]interface{}{},
			}, nil
		})

	// act
	_, err := centralizedCli.CreateLabel(context.Background(), "CreatePodLabel", data, label.PodLabelExist)

	// assert
	if err != nil {
		t.Errorf("Test_CentralizedClient_CreateLabel() error: %v", err)
	}
}

func Test_CentralizedClient_DeleteLabel(t *testing.T) {
	// arrange
	var cli *client.Client
	var centralizedCli = &CentralizedClient{
		Client: client.Client{Semaphore: utils.NewSemaphore(3)},
	}
	var data = map[string]interface{}{}

	// mock
	gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"Error": map[string]interface{}{
					"code": 0,
				},
				"Data": map[string]interface{}{},
			}, nil
		})

	// act
	_, err := centralizedCli.DeleteLabel(context.Background(), "DeletePodLabel", data, label.PodLabelNotExist)

	// assert
	if err != nil {
		t.Errorf("Test_CentralizedClient_DeletePodLabel() error: %v", err)
	}
}
