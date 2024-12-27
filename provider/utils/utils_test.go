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
