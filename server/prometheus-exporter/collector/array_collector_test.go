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

func Test_parseArrayModel_GetProductModeString(t *testing.T) {
	// arrange
	mockInDataKey := "fake_key"
	mockMetricsName := "fake_metrics"
	mockInData := map[string]string{
		"productModeString": "fake_product_string",
		"PRODUCTMODE":       "fake_product_mode",
	}

	// action
	got := parseArrayModel(mockInDataKey, mockMetricsName, mockInData)

	// assert
	if !reflect.DeepEqual(got, "fake_product_string") {
		t.Errorf("parseStorageData() got = %v, want %v", got, "fake_data")
	}
}

func Test_parseArrayModel_GetProductMode(t *testing.T) {
	// arrange
	mockInDataKey := "fake_key"
	mockMetricsName := "fake_metrics"
	mockInData := map[string]string{
		"PRODUCTMODE": "61",
	}

	// action
	got := parseArrayModel(mockInDataKey, mockMetricsName, mockInData)

	// assert
	if !reflect.DeepEqual(got, "6800 V3") {
		t.Errorf("parseStorageData() got = %v, want %v", got, "fake_data")
	}
}

func TestNewArrayCollector(t *testing.T) {
	// arrange
	var wantCollector = &ArrayCollector{
		BaseCollector: &BaseCollector{
			backendName:      "fake_backend",
			monitorType:      "object",
			collectorName:    "array",
			metricsHelpMap:   arrayObjectMetricsHelpMap,
			metricsLabelMap:  arrayObjectMetricsLabelMap,
			labelParseMap:    arrayObjectLabelParseMap,
			metricsParseMap:  arrayObjectMetricsParseMap,
			metricsDataCache: nil,
			metrics:          make(map[string]*prometheus.Desc),
		},
	}

	// action
	got, err := NewArrayCollector("fake_backend", "object", []string{""},
		nil)

	// assert
	if (err != nil) != false {
		t.Errorf("NewArrayCollector() error = %v, wantErr %v", err, true)
		return
	}
	if !reflect.DeepEqual(got, wantCollector) {
		t.Errorf("NewArrayCollector() got = %v, want %v", got, nil)
	}
}
