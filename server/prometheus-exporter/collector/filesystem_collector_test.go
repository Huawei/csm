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
