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

package metricscache

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
)

func TestMetricsPVData_SetMetricsData(t *testing.T) {
	// arrange
	mockPVMetricsData := &MetricsPVData{BaseMetricsData: &BaseMetricsData{BackendName: "fake_backend_name"}}
	ctx := context.Background()

	// mock
	mock := gomonkey.NewPatches()
	mock.ApplyFunc(GetAndParsePVInfo, func(ctx context.Context,
		backendName, collectType string) (*storageGRPC.CollectResponse, error) {
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
