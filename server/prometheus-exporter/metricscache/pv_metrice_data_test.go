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
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/csm/v2/provider/cmicore"
)

func TestMetricsPVData_SetMetricsData(t *testing.T) {
	// arrange
	mockPVMetricsData := &MetricsPVData{BaseMetricsData: &BaseMetricsData{BackendName: "fake_backend_name"}}
	ctx := context.Background()

	// mock
	mock := gomonkey.NewPatches()
	mock.ApplyFunc(GetAndParsePVInfo, func(ctx context.Context,
		backendName, collectType string) (*cmicore.CollectResponse, error) {
		return nil, nil
	})

	// action
	got := mockPVMetricsData.SetMetricsData(ctx, "fake_name", "object", []string{})

	// assert
	if got != nil {
		t.Errorf("SetMetricsData() err got = %v, want nil", got)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestMetricsPVData_SetMetricsData_WithError(t *testing.T) {
	// arrange
	mockPVMetricsData := &MetricsPVData{BaseMetricsData: &BaseMetricsData{BackendName: "fake_backend_name"}}
	ctx := context.Background()
	wantErr := errors.New("get pv info failed")

	// mock
	mock := gomonkey.NewPatches()
	mock.ApplyFunc(GetAndParsePVInfo, func(ctx context.Context,
		backendName, collectType string) (*cmicore.CollectResponse, error) {
		return nil, wantErr
	})

	// action
	got := mockPVMetricsData.SetMetricsData(ctx, "fake_name", "object", []string{})

	// assert
	if got == nil {
		t.Error("SetMetricsData() err got = nil, want error")
	}
	if !strings.Contains(got.Error(), wantErr.Error()) {
		t.Errorf("SetMetricsData() err got = %v, want contains %v", got, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestMetricsPVData_NewMetricsPVData(t *testing.T) {
	// arrange
	backendName := "test_backend"
	metricsType := "pv"

	// action
	got, err := NewMetricsPVData(backendName, metricsType)

	// assert
	if err != nil {
		t.Errorf("NewMetricsPVData() error = %v, want nil", err)
	}
	if got == nil {
		t.Fatal("NewMetricsPVData() got = nil, want non-nil")
	}

	// Type assert to access concrete type fields
	pvData, ok := got.(*MetricsPVData)
	if !ok {
		t.Fatal("NewMetricsPVData() return type is not *MetricsPVData")
	}
	if pvData.BackendName != backendName {
		t.Errorf("NewMetricsPVData() BackendName = %v, want %v", pvData.BackendName, backendName)
	}
	if pvData.MetricsType != metricsType {
		t.Errorf("NewMetricsPVData() MetricsType = %v, want %v", pvData.MetricsType, metricsType)
	}
}
