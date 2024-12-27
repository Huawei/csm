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

package collector

import (
	"reflect"
	"testing"
)

func Test_parseStorageData_GetDataSuccess(t *testing.T) {
	// arrange
	mockInDataKey := "fake_key"
	mockMetricsName := "fake_metrics"
	mockInData := map[string]string{
		"fake_key": "fake_data",
	}

	// action
	got := parseStorageData(mockInDataKey, mockMetricsName, mockInData)

	// assert
	if !reflect.DeepEqual(got, "fake_data") {
		t.Errorf("parseStorageData() got = %v, want %v", got, "fake_data")
	}
}

func Test_parseStorageData_GetDataEmpty(t *testing.T) {
	// arrange
	mockInDataKey := "fake_key1"
	mockMetricsName := "fake_metrics"
	mockInData := map[string]string{
		"fake_key": "fake_data",
	}

	// action
	got := parseStorageData(mockInDataKey, mockMetricsName, mockInData)

	// assert
	if !reflect.DeepEqual(got, "") {
		t.Errorf("parseStorageData() got = %v, want %v", got, "fake_data")
	}
}

func Test_parseStorageStatus_GetHealthStatus(t *testing.T) {
	// arrange
	mockInDataKey := ""
	mockMetricsName := "health_status"
	mockInData := map[string]string{
		"HEALTHSTATUS": "1",
	}

	// action
	got := parseStorageStatus(mockInDataKey, mockMetricsName, mockInData)

	// assert
	if !reflect.DeepEqual(got, "Normal") {
		t.Errorf("parseStorageStatus() got = %v, want %v", got, "fake_data")
	}
}

func Test_parseStorageStatus_GetRunningStatus(t *testing.T) {
	// arrange
	mockInDataKey := ""
	mockMetricsName := "running_status"
	mockInData := map[string]string{
		"RUNNINGSTATUS": "1",
	}

	// action
	got := parseStorageStatus(mockInDataKey, mockMetricsName, mockInData)

	// assert
	if !reflect.DeepEqual(got, "Normal") {
		t.Errorf("parseStorageStatus() got = %v, want %v", got, "fake_data")
	}
}

func Test_parseLabelListToLabelValueSlice_GetLabelValueSuccess(t *testing.T) {
	// arrange
	mockLabelKeys := []string{"fake_label_key1", "fake_label_key2"}
	mockLabelParseRelation := map[string]parseRelation{
		"fake_label_key1": {"fake_key1", parseStorageData},
		"fake_label_key2": {"fake_key2", parseStorageData},
	}
	mockInData := map[string]string{
		"fake_key1": "fake_data1",
		"fake_key2": "fake_data2",
	}
	wantlabelValueSlice := []string{"fake_data1", "fake_data2"}

	// action
	got := parseLabelListToLabelValueSlice(mockLabelKeys, mockLabelParseRelation, "", mockInData)

	// assert
	if !reflect.DeepEqual(got, wantlabelValueSlice) {
		t.Errorf("parseLabelListToLabelValueSlice() got = %v, want %v",
			got, wantlabelValueSlice)
	}
}

func Test_parseStorageSectorsToGB(t *testing.T) {
	// arrange
	mockInDataKey := "fake_key"
	mockInData := map[string]string{
		"fake_key": "209715200",
	}

	// action
	got := parseStorageSectorsToGB(mockInDataKey, "", mockInData)

	// assert
	if !reflect.DeepEqual(got, "100.0000") {
		t.Errorf("parseStorageStatus() got = %v, want %v", got, "100.0000")
	}
}
