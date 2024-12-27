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
	"github.com/huawei/csm/v2/storage/utils"
)

var centralizedCli = &CentralizedClient{
	Client: client.Client{
		Semaphore: utils.NewSemaphore(3),
	},
}

var mockGetresponse = map[string]interface{}{
	"Error": map[string]interface{}{
		"code": 0,
	},
	"Data": []map[string]interface{}{},
}

func MockHttpGet(response map[string]interface{}) *gomonkey.Patches {
	var cli *client.Client
	return gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return response, nil
		})
}

func TestCentralizedClient_GetByUrl(t *testing.T) {
	httpGet := MockHttpGet(mockGetresponse)
	defer httpGet.Reset()

	tests := []struct {
		name   string
		urlKey string
	}{
		{
			name:   "TestGetStoragePools",
			urlKey: "GetStoragePools",
		},
		{
			name:   "TestGetControllers",
			urlKey: "GetControllers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := centralizedCli.GetByUrl(context.Background(), tt.urlKey)
			if err != nil {
				t.Errorf("GetByUrl() error = %v,", err)
			}
		})
	}
}
