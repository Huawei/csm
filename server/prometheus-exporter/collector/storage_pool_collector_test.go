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

	"github.com/prometheus/client_golang/prometheus"
)

func TestStoragePoolCollector_BuildDesc(t *testing.T) {
	// arrange
	var mockMetricsLabelMap = map[string][]string{
		"fake_key1": {"fake_label1", "fake_label2"},
		"fake_key2": {"fake_label2", "fake_label3"},
	}
	var mockMetricsHelpMap = map[string]string{
		"fake_key1": "fake help info1",
		"fake_key2": "fake help info2",
	}
	var mockCollector = StoragePoolCollector{
		BaseCollector: &BaseCollector{
			backendName:     "fake_name",
			monitorType:     "object",
			collectorName:   "storagepool",
			metricsHelpMap:  mockMetricsHelpMap,
			metricsLabelMap: mockMetricsLabelMap,
			metrics:         make(map[string]*prometheus.Desc)},
	}
	var mockMetrics = map[string]*prometheus.Desc{
		"fake_key1": prometheus.NewDesc(
			prometheus.BuildFQName(
				MetricsNamespace, "storage_pool", "fake_key1"),
			"fake help info1",
			mockCollector.metricsLabelMap["fake_key1"],
			nil),
		"fake_key2": prometheus.NewDesc(
			prometheus.BuildFQName(
				MetricsNamespace, "storage_pool", "fake_key2"),
			"fake help info2",
			mockCollector.metricsLabelMap["fake_key2"],
			nil),
	}

	// action
	mockCollector.BuildDesc()

	// assert
	if !reflect.DeepEqual(mockMetrics, mockCollector.metrics) {
		t.Errorf("BuildDesc() want = %v, but got = %v", mockMetrics, mockCollector)
	}
}

func TestNewStoragePoolCollector_GetObjectCollector(t *testing.T) {
	// arrange
	var wantCollector = &StoragePoolCollector{
		BaseCollector: &BaseCollector{
			backendName:      "fake_backend",
			monitorType:      "object",
			collectorName:    "storagepool",
			metricsHelpMap:   storagePoolObjectMetricsHelpMap,
			metricsLabelMap:  storagePoolObjectMetricsLabelMap,
			labelParseMap:    storagePoolObjectLabelParseMap,
			metricsParseMap:  storagePoolObjectMetricsParseMap,
			metricsDataCache: nil,
			metrics:          make(map[string]*prometheus.Desc),
		},
	}

	// action
	got, err := NewStoragePoolCollector("fake_backend", "object", []string{""},
		nil)

	// assert
	if err != nil {
		t.Errorf("NewStoragePoolCollector() error = %v", err)
		return
	}
	if !reflect.DeepEqual(got, wantCollector) {
		t.Errorf("NewStoragePoolCollector() got = %v, want %v", got, nil)
	}
}

func TestNewStoragePoolCollector_GetPerformanceCollector(t *testing.T) {
	// arrange
	var mockMetricsIndicators []string

	// action
	_, err := NewStoragePoolCollector("fake_backend", "performance", mockMetricsIndicators,
		nil)

	// assert
	if err == nil {
		t.Errorf("NewStoragePoolCollector() error = %v, wantErr %v", err, true)
		return
	}
}
