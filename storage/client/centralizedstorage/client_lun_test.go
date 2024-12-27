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
	"testing"
)

var mockCountResponse = map[string]interface{}{
	"Error": map[string]interface{}{
		"code": 0,
	},
	"Data": map[string]string{
		"COUNT": "10",
	},
}

func TestCentralizedClient_countQuery(t *testing.T) {
	httpGet := MockHttpGet(mockCountResponse)
	defer httpGet.Reset()

	tests := []struct {
		name   string
		urlKey string
	}{
		{
			name:   "TestGetFilesystemCount",
			urlKey: "GetFilesystemCount",
		},
		{
			name:   "TestGetLunCount",
			urlKey: "GetLunCount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := centralizedCli.countQuery(context.Background(), tt.urlKey)
			if err != nil {
				t.Errorf("countQuery() error = %v,", err)
			}
		})
	}
}

func TestCentralizedClient_pageQuery(t *testing.T) {
	httpGet := MockHttpGet(mockGetresponse)
	defer httpGet.Reset()

	tests := []struct {
		name   string
		urlKey string
	}{
		{
			name:   "TestGetFilesystem",
			urlKey: "GetFilesystem",
		},
		{
			name:   "TestGetLuns",
			urlKey: "GetLuns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := centralizedCli.pageQuery(context.Background(), 0, 100, tt.urlKey)
			if err != nil {
				t.Errorf("pageQuery() error = %v,", err)
			}
		})
	}
}
