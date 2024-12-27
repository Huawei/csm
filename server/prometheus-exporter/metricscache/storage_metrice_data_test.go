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
	"reflect"
	"testing"

	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
)

func TestStorageMetricsData_buildTheStorageGRPCRequest(t *testing.T) {
	// arrange
	wantRequest := &storageGRPC.CollectRequest{
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
