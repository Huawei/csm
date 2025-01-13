/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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

func Test_parseLunCapacityUsage(t *testing.T) {
	// arrange
	mockInData := map[string]string{
		"CAPACITY":      "10",
		"ALLOCCAPACITY": "5",
	}

	// action
	got := parseLunCapacityUsage("", "", mockInData)
	want := "50"

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseStorageStatus() got = %v, want %v", got, want)
	}
}

func Test_buildObjectLunCollector(t *testing.T) {
	// arrange
	var wantCollector = &LunCollector{
		BaseCollector: &BaseCollector{
			backendName:      "fake_backend",
			monitorType:      "object",
			collectorName:    "lun",
			metricsHelpMap:   lunObjectMetricsHelpMap,
			metricsLabelMap:  lunObjectMetricsLabelMap,
			labelParseMap:    lunObjectLabelParseMap,
			metricsParseMap:  lunObjectMetricsParseMap,
			metricsDataCache: nil,
			metrics:          make(map[string]*prometheus.Desc),
		},
	}

	// action
	got, err := NewLunCollector("fake_backend", "object", []string{""},
		nil)

	// assert
	if err != nil {
		t.Errorf("NewLunCollector() error = %v", err)
		return
	}
	if !reflect.DeepEqual(got, wantCollector) {
		t.Errorf("NewLunCollector() got = %v, want %v", got, nil)
	}
}
