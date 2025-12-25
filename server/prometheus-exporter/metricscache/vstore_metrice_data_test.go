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
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xuanwuV1 "github.com/Huawei/eSDK_K8S_Plugin/v4/client/apis/xuanwu/v1"
	"github.com/huawei/csm/v2/controller/utils/consts"
)

func TestMetricsVstoreData_SetMetricsData(t *testing.T) {
	// arrange
	backendNamespace := "fake_namespace"
	backendName := "fake_backend_name"
	mockMetricsVstoreData, _ := NewMetricsVstoreData("fake_backend_name", "fakeType")
	ctx := context.Background()

	// mock
	mock := gomonkey.NewPatches()
	mock.ApplyFunc(getBackendContentFromApi, func(ctx context.Context) (*xuanwuV1.StorageBackendContentList, error) {
		backendClaims := &xuanwuV1.StorageBackendClaimList{
			TypeMeta: metaV1.TypeMeta{},
			ListMeta: metaV1.ListMeta{},
			Items: []xuanwuV1.StorageBackendClaim{
				{
					TypeMeta: metaV1.TypeMeta{},
					ObjectMeta: metaV1.ObjectMeta{
						Name:      backendName,
						Namespace: backendNamespace,
					},
					Spec: xuanwuV1.StorageBackendClaimSpec{},
					Status: &xuanwuV1.StorageBackendClaimStatus{
						StorageType: consts.StorageNas,
					},
				},
			},
		}
		backendContents := &xuanwuV1.StorageBackendContentList{
			Items: []xuanwuV1.StorageBackendContent{
				{
					TypeMeta:   metaV1.TypeMeta{},
					ObjectMeta: metaV1.ObjectMeta{},
					Spec: xuanwuV1.StorageBackendContentSpec{
						BackendClaim: backendNamespace + "/" + backendName,
					},
					Status: &xuanwuV1.StorageBackendContentStatus{
						Pools: []xuanwuV1.Pool{
							{
								Name: "fakePoolName",
								Capacities: map[string]string{
									"TotalCapacity": "100",
									"UsedCapacity":  "50",
									"FreeCapacity":  "50",
								},
							},
						},
						Specification: map[string]string{"VStoreID": "1", "VStoreName": "fake_vstore_name"},
					},
				},
			},
		}
		filterVstoreBackend(backendClaims, backendContents)
		return backendContents, nil
	})
	// action
	got := mockMetricsVstoreData.SetMetricsData(ctx, "fake_name", "object", []string{})

	// assert
	if got != nil {
		t.Errorf("SetMetricsData() err got = %v, want nil", got)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}
