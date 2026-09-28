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
	mockInData2 := map[string]string{
		"fake_key": "209715201",
	}

	// action
	got := parseStorageSectorsToGB(mockInDataKey, "", mockInData)
	got2 := parseStorageSectorsToGB(mockInDataKey, "", mockInData2)

	// assert
	if !reflect.DeepEqual(got, "100") {
		t.Errorf("parseStorageSectorsToGB() got = %v, want %v", got, "100")
	}
	if !reflect.DeepEqual(got2, "100.00000047683716") {
		t.Errorf("parseStorageSectorsToGB() got2 = %v, want %v", got2, "100.00000047683716")
	}
}

func Test_parseFilesystemCapacityUsage(t *testing.T) {
	tests := []struct {
		name   string
		inData map[string]string
		want   string
	}{
		{
			name:   "empty inData",
			inData: nil,
			want:   "",
		},
		{
			name: "capacity is zero",
			inData: map[string]string{
				"CAPACITY":                "0",
				"allocatedPoolQuota":      "50",
				"SNAPSHOTRESERVECAPACITY": "10",
			},
			want: "",
		},
		{
			name: "capacity parse error",
			inData: map[string]string{
				"CAPACITY":                "not_a_number",
				"allocatedPoolQuota":      "50",
				"SNAPSHOTRESERVECAPACITY": "10",
			},
			want: "",
		},
		{
			name: "usedCapacity parse error",
			inData: map[string]string{
				"CAPACITY":                "200",
				"allocatedPoolQuota":      "not_a_number",
				"SNAPSHOTRESERVECAPACITY": "10",
			},
			want: "",
		},
		{
			name: "snapshotReserveCapacity parse error",
			inData: map[string]string{
				"CAPACITY":                "200",
				"allocatedPoolQuota":      "50",
				"SNAPSHOTRESERVECAPACITY": "not_a_number",
			},
			want: "",
		},
		{
			name: "snapshotReserveCapacity equals capacity - no divide by zero",
			inData: map[string]string{
				"CAPACITY":                "100",
				"allocatedPoolQuota":      "50",
				"SNAPSHOTRESERVECAPACITY": "100",
			},
			want: "",
		},
		{
			name: "snapshotReserveCapacity greater than capacity",
			inData: map[string]string{
				"CAPACITY":                "100",
				"allocatedPoolQuota":      "50",
				"SNAPSHOTRESERVECAPACITY": "200",
			},
			want: "",
		},
		{
			name: "normal usage calculation",
			inData: map[string]string{
				"CAPACITY":                "200",
				"allocatedPoolQuota":      "50",
				"SNAPSHOTRESERVECAPACITY": "0",
			},
			want: "25",
		},
		{
			name: "usage with snapshot reserve deducted",
			inData: map[string]string{
				"CAPACITY":                "100",
				"allocatedPoolQuota":      "60",
				"SNAPSHOTRESERVECAPACITY": "20",
			},
			want: "75",
		},
		{
			name: "zero used capacity",
			inData: map[string]string{
				"CAPACITY":                "100",
				"allocatedPoolQuota":      "0",
				"SNAPSHOTRESERVECAPACITY": "0",
			},
			want: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFilesystemCapacityUsage("", "", tt.inData)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFilesystemCapacityUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}
