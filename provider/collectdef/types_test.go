/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

// Package collectdef provides declarative Prometheus Collector definitions driven by ObjectDef.
package collectdef

import "testing"

func TestObjectDefPrometheusSubsystemName(t *testing.T) {
	tests := []struct {
		name     string
		def      ObjectDef
		expected string
	}{
		{"uses CollectType when PrometheusName is empty",
			ObjectDef{CollectType: "controller"},
			"controller"},
		{"uses PrometheusName when set",
			ObjectDef{CollectType: "storagepool"},
			"storagepool"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.def.PrometheusSubsystemName(); got != tt.expected {
				t.Errorf("PrometheusSubsystemName() = %v, want %v", got, tt.expected)
			}
		})
	}
}
