/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

package utils

import (
	"reflect"
	"testing"
)

type testStruct struct {
	Id   string `json:"Id" metrics:"Id"`
	Name string `json:"Name" metrics:"Name"`
}

func Test_MapStringToInt_Success(t *testing.T) {
	// arrange
	var want = []int{1, 2, 3}
	var stringSlice = []string{"1", "2", "3"}

	// action
	got := MapStringToInt(stringSlice)

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Test_MapStringToInt_Success() failed, want data = %v, but got = %v", want, got)
	}
}

func Test_MapToStruct_Success(t *testing.T) {
	// arrange
	input := map[string]string{
		"Id":   "test-id",
		"Name": "test-name",
	}

	// action
	_, err := MapToStruct[map[string]string, testStruct](input)

	// assert
	if err != nil {
		t.Errorf("Test_MapToStruct_Success() failed, error = %v", err)
	}
}

func Test_MapToStructSlice_Output_Is_Struct(t *testing.T) {
	// arrange
	input := map[string]interface{}{
		"Id":   "test-id",
		"Name": "test-name",
	}

	// action
	slice, err := MapToStructSlice[map[string]interface{}, testStruct](input)

	// assert
	if err != nil {
		t.Errorf("Test_MapToStructSlice_Output_Is_Struct() failed, error = %v", err)
	}

	if len(slice) != 1 {
		t.Errorf("Test_MapToStructSlice_Output_Is_Struct() failed, want len = %d, got = %d", 1, len(slice))
	}
}

func Test_MapToStructSlice_Output_Is_StructSlice(t *testing.T) {
	// arrange
	single := map[string]interface{}{
		"Id":   "test-id",
		"Name": "test-name",
	}
	input := []map[string]interface{}{single, single}

	// action
	slice, err := MapToStructSlice[[]map[string]interface{}, testStruct](input)

	// assert
	if err != nil {
		t.Errorf("Test_MapToStructSlice_Output_Is_StructSlice() failed, error = %v", err)
	}

	if len(slice) != 2 {
		t.Errorf("Test_MapToStructSlice_Output_Is_StructSlice() failed, want len = %d, got = %d", 1, len(slice))
	}
}

func Test_StructToMap_Success(t *testing.T) {
	// arrange
	object := testStruct{
		Id:   "1",
		Name: "NAME-1",
	}
	want := map[string]string{
		"Id":   "1",
		"Name": "NAME-1",
	}

	// action
	got := StructToMap(object)

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Test_StructToMap_Success() = %v, want %v", got, want)
	}
}

func Test_CompareVersions(t *testing.T) {
	// arrange
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{name: "compare 6.1.9 and 6.1.10", v1: "6.1.9", v2: "6.1.10", want: -1},
		{name: "compare 6.1.10 and 6.1.10", v1: "6.1.10", v2: "6.1.10", want: 0},
		{name: "compare 6.1.10 and 6.1.9", v1: "6.1.10", v2: "6.1.9", want: 1},
		{name: "compare 6.2.0 and 6.1.9999", v1: "6.2.0", v2: "6.1.9999", want: 1},
		{name: "compare 6.1.9999 and 6.2.0", v1: "6.1.9999", v2: "6.2.0", want: -1},
		{name: "compare 7.0.0 and 6.9.9", v1: "7.0.0", v2: "6.9.9", want: 1},
		{name: "compare 6.9.9 and 7.0.0", v1: "6.9.9", v2: "7.0.0", want: -1},
		{name: "compare 6.1.8 and 6.1.8.SPH001", v1: "6.1.8", v2: "6.1.8.SPH001", want: -1},
		{name: "compare 6.1.8.SPH001 and 6.1.8", v1: "6.1.8.SPH001", v2: "6.1.8", want: 1},
		{name: "compare 6.1.8.SPH002 and 6.1.8.SPH001", v1: "6.1.8.SPH002", v2: "6.1.8.SPH001", want: 1},
		{name: "compare 6.1.8.SPH001 and 6.1.8.SPH002", v1: "6.1.8.SPH001", v2: "6.1.8.SPH002", want: -1},
		{name: "compare 6.1.8.SPH001 and 6.1.8.SPH001", v1: "6.1.8.SPH001", v2: "6.1.8.SPH001", want: 0},
		{name: "compare 7.0 and 6.9.9", v1: "7.0", v2: "6.9.9", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// action
			got := CompareVersions(tt.v1, tt.v2)

			// assert
			if got != tt.want {
				t.Errorf("CompareVersions() got = %v, want %v", got, tt.want)
			}
		})
	}
}
