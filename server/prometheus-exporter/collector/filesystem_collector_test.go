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
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func Test_buildObjectFilesystemCollector(t *testing.T) {
	// arrange
	var wantCollector = &FilesystemCollector{
		BaseCollector: &BaseCollector{
			backendName:      "fake_backend",
			monitorType:      "object",
			collectorName:    "filesystem",
			metricsHelpMap:   filesystemObjectMetricsHelpMap,
			metricsLabelMap:  filesystemObjectMetricsLabelMap,
			labelParseMap:    filesystemObjectLabelParseMap,
			metricsParseMap:  filesystemObjectMetricsParseMap,
			metricsDataCache: nil,
			metrics:          make(map[string]*prometheus.Desc),
		},
	}

	// action
	got, err := NewFilesystemCollector("fake_backend", "object", []string{""},
		nil)

	// assert
	if err != nil {
		t.Errorf("NewFilesystemCollector() error = %v", err)
		return
	}
	if !reflect.DeepEqual(got, wantCollector) {
		t.Errorf("NewFilesystemCollector() got = %v, want %v", got, nil)
	}
}

func Test_ParseStorageSectorsToGB_SnapshotCapacity(t *testing.T) {
	// arrange
	testCases := []struct {
		name      string
		sectors   string
		expectGB  string
		expectErr bool
	}{
		{"1 GB", "2097152", "1", false},
		{"500 MB", "1048576", "0.5", false},
		{"10 GB", "20971520", "10", false},
		{"Zero", "0", "0", false},
		{"Invalid", "not_a_number", "", true},
		{"Empty", "", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputData := map[string]string{
				"SNAPSHOTUSECAPACITY": tc.sectors,
			}
			// action
			result := parseStorageSectorsToGB("SNAPSHOTUSECAPACITY", "snapshot_capacity", inputData)

			// assert
			if tc.expectErr && result == "" {
				return
			}
			if result != tc.expectGB {
				t.Errorf("parseStorageSectorsToGB() sectors=%s, got=%s, want=%s",
					tc.sectors, result, tc.expectGB)
			}
		})
	}
}

func Test_FilesystemSnapshotCapacityDefined(t *testing.T) {
	// verify that snapshot_capacity metric is defined in filesystem collector
	collector, err := NewFilesystemCollector("test_backend", "object", []string{""}, nil)
	if err != nil {
		t.Fatalf("NewFilesystemCollector() error = %v", err)
	}

	fsCollector, ok := collector.(*FilesystemCollector)
	if !ok {
		t.Fatalf("Expected FilesystemCollector, got %T", collector)
	}

	// check metricsParseMap
	if _, exists := fsCollector.metricsParseMap["snapshot_used_capacity"]; !exists {
		t.Error("Expected snapshot_used_capacity metric to be defined in metricsParseMap")
	}

	// check metricsHelpMap
	if _, exists := fsCollector.metricsHelpMap["snapshot_used_capacity"]; !exists {
		t.Error("Expected snapshot_used_capacity help to be defined in metricsHelpMap")
	}

	// check metricsLabelMap
	if _, exists := fsCollector.metricsLabelMap["snapshot_used_capacity"]; !exists {
		t.Error("Expected snapshot_used_capacity labels to be defined in metricsLabelMap")
	}
}

func Test_FilesystemObjectSnapshotCapacityField(t *testing.T) {
	// arrange
	testCases := []struct {
		name    string
		sectors int64
		wantGB  float64
	}{
		{"1 GB in sectors", 2097152, 1.0},
		{"500 MB in sectors", 1048576, 0.5},
		{"10 GB in sectors", 20971520, 10.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputData := map[string]string{
				"SNAPSHOTUSECAPACITY": fmt.Sprintf("%d", tc.sectors),
			}
			// action
			result := parseStorageSectorsToGB(filesystemSnapshotCapacityKey, "snapshot_capacity", inputData)
			if result == "" {
				t.Errorf("parseStorageSectorsToGB() returned empty string for %d sectors", tc.sectors)
				return
			}

			gotGB, err := strconv.ParseFloat(result, 64)
			if err != nil {
				t.Errorf("parseStorageSectorsToGB() result is not a valid float: %v", err)
				return
			}

			// assert
			if gotGB != tc.wantGB {
				t.Errorf("parseStorageSectorsToGB() = %f, want %f", gotGB, tc.wantGB)
			}
		})
	}
}

func Test_ParseFilesystemCapacityUsage(t *testing.T) {
	// arrange
	tests := []struct {
		name        string
		inDataKey   string
		metricsName string
		inData      map[string]string
		expected    string
	}{
		{
			name:        "empty map returns empty",
			inDataKey:   "capacity_usage",
			metricsName: "capacity_usage",
			inData:      map[string]string{},
			expected:    "",
		},
		{
			name:        "capacity parse error returns empty",
			inDataKey:   "capacity_usage",
			metricsName: "capacity_usage",
			inData: map[string]string{
				"CAPACITY":                "invalid",
				"allocatedPoolQuota":      "100",
				"SNAPSHOTRESERVECAPACITY": "10",
			},
			expected: "",
		},
		{
			name:        "capacity zero returns empty",
			inDataKey:   "capacity_usage",
			metricsName: "capacity_usage",
			inData: map[string]string{
				"CAPACITY":                "0",
				"allocatedPoolQuota":      "100",
				"SNAPSHOTRESERVECAPACITY": "10",
			},
			expected: "",
		},
		{
			name:        "usedCapacity parse error returns empty",
			inDataKey:   "capacity_usage",
			metricsName: "capacity_usage",
			inData: map[string]string{
				"CAPACITY":                "1000",
				"allocatedPoolQuota":      "invalid",
				"SNAPSHOTRESERVECAPACITY": "10",
			},
			expected: "",
		},
		{
			name:        "snapshotReserveCapacity parse error returns empty",
			inDataKey:   "capacity_usage",
			metricsName: "capacity_usage",
			inData: map[string]string{
				"CAPACITY":                "1000",
				"allocatedPoolQuota":      "100",
				"SNAPSHOTRESERVECAPACITY": "invalid",
			},
			expected: "",
		},
		{
			name:        "normal case returns correct percentage",
			inDataKey:   "capacity_usage",
			metricsName: "capacity_usage",
			inData: map[string]string{
				"CAPACITY":                "1000",
				"allocatedPoolQuota":      "500",
				"SNAPSHOTRESERVECAPACITY": "0",
			},
			expected: "50",
		},
		{
			name:        "with snapshot reserve capacity",
			inDataKey:   "capacity_usage",
			metricsName: "capacity_usage",
			inData: map[string]string{
				"CAPACITY":                "1000",
				"allocatedPoolQuota":      "450",
				"SNAPSHOTRESERVECAPACITY": "100",
			},
			expected: "50",
		},
		{
			name:        "100 percent usage",
			inDataKey:   "capacity_usage",
			metricsName: "capacity_usage",
			inData: map[string]string{
				"CAPACITY":                "1000",
				"allocatedPoolQuota":      "900",
				"SNAPSHOTRESERVECAPACITY": "100",
			},
			expected: "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// action
			result := parseFilesystemCapacityUsage(tt.inDataKey, tt.metricsName, tt.inData)
			// assert
			if result != tt.expected {
				t.Errorf("parseFilesystemCapacityUsage() = %v, want %v", result, tt.expected)
			}
		})
	}
}
