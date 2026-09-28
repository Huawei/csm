/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2026. All rights reserved.
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

package metricscache

import (
	"reflect"
	"testing"

	"github.com/huawei/csm/v2/provider/cmicore"
)

func TestStorageMetricsData_buildTheStorageGRPCRequest(t *testing.T) {
	// arrange
	wantRequest := &cmicore.CollectRequest{
		BackendName: "fake_backend_name",
		CollectType: "fake_collector_name",
		MetricsType: "object",
		Indicators:  []string{},
	}
	mockStorageData := StorageMetricsData{BaseMetricsData: &BaseMetricsData{BackendName: "fake_backend_name"}}

	// action
	got := mockStorageData.buildTheStorageGRPCRequest(
		"fake_collector_name", "object", []string{})

	// assert
	if !reflect.DeepEqual(got, wantRequest) {
		t.Errorf("buildTheStorageGRPCRequest() got = [%v], want [%v]", got, wantRequest)
	}
}

func TestStorageMetricsData_buildTheStorageGRPCRequest_WithIndicators(t *testing.T) {
	// arrange
	mockStorageData := StorageMetricsData{BaseMetricsData: &BaseMetricsData{BackendName: "fake_backend_name"}}
	indicators := []string{"21,22,370"}

	// action
	got := mockStorageData.buildTheStorageGRPCRequest("fake_collector_name", "performance", indicators)

	// assert
	if got == nil {
		t.Fatal("buildTheStorageGRPCRequest() got = nil, want non-nil")
	}
	if got.MetricsType != "performance" {
		t.Errorf("buildTheStorageGRPCRequest() MetricsType = %v, want performance", got.MetricsType)
	}
	if !reflect.DeepEqual(got.Indicators, []string{"21", "22", "370"}) {
		t.Errorf("buildTheStorageGRPCRequest() Indicators = %v, want [%v]", got.Indicators, []string{"21", "22", "370"})
	}
}

func TestStorageMetricsData_buildTheStorageGRPCRequest_EmptyIndicators(t *testing.T) {
	// arrange
	mockStorageData := StorageMetricsData{BaseMetricsData: &BaseMetricsData{BackendName: "fake_backend_name"}}
	indicators := []string{""}

	// action
	got := mockStorageData.buildTheStorageGRPCRequest("fake_collector_name", "performance", indicators)

	// assert
	if got != nil {
		t.Errorf("buildTheStorageGRPCRequest() with empty indicator got = %v, want nil", got)
	}
}

func TestStorageMetricsData_buildTheStorageGRPCRequest_NotPerformance(t *testing.T) {
	// arrange
	mockStorageData := StorageMetricsData{BaseMetricsData: &BaseMetricsData{BackendName: "fake_backend_name"}}
	indicators := []string{"some_indicator"}

	// action
	got := mockStorageData.buildTheStorageGRPCRequest("fake_collector_name", "object", indicators)

	// assert
	if got == nil {
		t.Fatal("buildTheStorageGRPCRequest() got = nil, want non-nil")
	}
	// For non-performance, indicators should be empty
	if len(got.Indicators) != 0 {
		t.Errorf("buildTheStorageGRPCRequest() Indicators = %v, want empty for object type", got.Indicators)
	}
}
