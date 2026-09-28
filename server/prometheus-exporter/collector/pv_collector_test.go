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

	"github.com/stretchr/testify/assert"
)

func Test_parsePVStorageIDGetName(t *testing.T) {
	// arrange
	mockInDataKey := ""
	mockMetricsName := ""
	mockInData := map[string]string{
		"NAME": "fake_name",
		"ID":   "fake_data",
	}

	// action
	got := parsePVStorageID(mockInDataKey, mockMetricsName, mockInData)

	// assert
	if !reflect.DeepEqual(got, "fake_data") {
		t.Errorf("parseStorageData() got = %v, want %v", got, "fake_data")
	}
}

func Test_parsePVStorageIDGetObjectName(t *testing.T) {
	// arrange
	mockInDataKey := ""
	mockMetricsName := ""
	mockInData := map[string]string{
		"ObjectName": "fake_name",
		"ObjectId":   "fake_data",
	}

	// action
	got := parsePVStorageID(mockInDataKey, mockMetricsName, mockInData)

	// assert
	if !reflect.DeepEqual(got, "fake_data") {
		t.Errorf("parseStorageData() got = %v, want %v", got, "fake_data")
	}
}

func Test_parsePVCapacityUsageSan(t *testing.T) {
	// arrange
	mockInDataKey := ""
	mockMetricsName := ""
	mockInData := map[string]string{
		"sbcStorageType": "oceanstor-san",
		"CAPACITY":       "100",
		"ALLOCCAPACITY":  "10",
	}

	// action
	got := parsePVCapacityUsage(mockInDataKey, mockMetricsName, mockInData)
	want := "10"

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseStorageData() got = %v, want %v", got, want)
	}
}

func TestPvTypePrometheusMetrics_Namespace(t *testing.T) {
	// arrange & act & assert
	metrics, ok := pvTypePrometheusMetrics["namespace"]
	assert.True(t, ok, "namespace should be in pvTypePrometheusMetrics")
	assert.Len(t, metrics, 23)
}

func TestParseNamespaceData_FusionStorageNas(t *testing.T) {
	// arrange
	data := map[string]string{
		"sbcStorageType": "fusionstorage-nas",
		"ID":             "35",
		"30001":          "1024",
	}
	// act
	result := parseNamespaceData("30001", "namespace_nfs_read_bandwidth", data)
	// assert
	assert.Equal(t, "1024", result)
}

func TestParseNamespaceData_SkipNonFusionStorage(t *testing.T) {
	// arrange
	data := map[string]string{
		"sbcStorageType": "oceanstor-nas",
		"ID":             "10",
		"30001":          "1024",
	}
	// act
	result := parseNamespaceData("30001", "namespace_nfs_read_bandwidth", data)
	// assert
	assert.Equal(t, skipReportValue, result)
}

func TestParseNamespaceData_EmptyData(t *testing.T) {
	// arrange
	data := map[string]string{}
	// act
	result := parseNamespaceData("30001", "namespace_nfs_read_bandwidth", data)
	// assert
	assert.Equal(t, "", result)
}

func TestParsePVStorageID_WithLowercaseId(t *testing.T) {
	// arrange — defensive fallback for lowercase id
	data := map[string]string{
		"id": "35",
	}
	// act
	result := parsePVStorageID("ID", "", data)
	// assert
	assert.Equal(t, "35", result)
}

func TestParsePVCapacityUsage_FusionStorageNas(t *testing.T) {
	// arrange
	data := map[string]string{
		"sbcStorageType":  "fusionstorage-nas",
		"SPACE_USED_RATE": "29",
	}
	// act
	result := parsePVCapacityUsage("", "", data)
	// assert
	assert.Equal(t, "29", result)
}
