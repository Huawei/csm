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

// Package collect is a package that provides object and performance collect
package collect

import (
	"reflect"
	"testing"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
)

func Test_AddCollectDetail_Success(t *testing.T) {
	// arrange
	response := &cmi.CollectResponse{}
	data := map[string]string{
		"Id":   "test-id",
		"Name": "test-name",
	}
	detail := struct {
		Id   string `json:"Id" metrics:"Id"`
		Name string `json:"Name" metrics:"Name"`
	}{Id: "test-id", Name: "test-name"}

	// action
	AddCollectDetail(detail, response)

	// assert
	if len(response.GetDetails()) != 1 {
		t.Errorf("Test_AddCollectDetail_Success() failed, want deltails = 1, but got = %d",
			len(response.GetDetails()))
		return
	}

	got := response.GetDetails()[0].GetData()
	if !reflect.DeepEqual(got, data) {
		t.Errorf("Test_AddCollectDetail_Success() failed, want data = %v, but got = %v", data, got)
	}
}

func Test_AddCollectDetailWithMap_Success(t *testing.T) {
	// arrange
	response := &cmi.CollectResponse{}
	data := map[string]string{
		"Id":   "test-id",
		"Name": "test-name",
	}
	// action
	AddCollectDetailWithMap(data, response)

	// assert
	if len(response.GetDetails()) != 1 {
		t.Errorf("Test_AddCollectDetailWithMap_Success() failed, want deltails = 1, but got = %d",
			len(response.GetDetails()))
		return
	}

	got := response.GetDetails()[0].GetData()
	if !reflect.DeepEqual(got, data) {
		t.Errorf("Test_AddCollectDetailWithMap_Success() failed, want data = %v, but got = %v", data, got)
	}
}

func TestConvertToResponse(t *testing.T) {
	// arrange
	request := &cmi.CollectRequest{
		BackendName: "test-backend",
		CollectType: "test-collect",
		MetricsType: "test-metrics",
	}
	input := []map[string]interface{}{
		{
			"ID":   "1",
			"NAME": "TEST-1",
		},
		{
			"ID":   "2",
			"NAME": "TEST-2",
		},
	}

	// action
	_, err := ConvertToResponse[[]map[string]interface{}, LunObject](input, request)

	// assert
	if err != nil {
		t.Errorf("TestConvertToResponse() failed, error = %v", err)
	}
}
