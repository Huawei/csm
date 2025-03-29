/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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
	"fmt"
	"reflect"
	"testing"

	xuanwuV1 "github.com/Huawei/eSDK_K8S_Plugin/v4/client/apis/xuanwu/v1"
	"github.com/agiledragon/gomonkey/v2"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
)

func Test_buildOutPVData(t *testing.T) {
	// arrange
	ctx := context.TODO()
	mockOutPVData := &storageGRPC.CollectResponse{
		BackendName: "fake_backend",
		CollectType: "fake_type",
		Details:     []*storageGRPC.CollectDetail{}}
	mockAllPVData := []coreV1.PersistentVolume{{}}
	mockAllSBCInfo := map[string]map[string]string{}
	mockParse := &parsePVMetrics{}

	// mock
	mock := gomonkey.NewPatches()
	mock.ApplyPrivateMethod(mockParse, "setCSIDriverNameMetrics",
		func(volume coreV1.PersistentVolume) *parsePVMetrics {
			return mockParse
		}).ApplyPrivateMethod(mockParse, "setVolumeHandleMetrics",
		func(volume coreV1.PersistentVolume) *parsePVMetrics {
			return mockParse
		}).ApplyPrivateMethod(mockParse, "setPVNameMetrics",
		func(volume coreV1.PersistentVolume) *parsePVMetrics {
			return mockParse
		}).ApplyPrivateMethod(mockParse, "setPVCNameMetrics",
		func(volume coreV1.PersistentVolume) *parsePVMetrics {
			mockParse.parseError = errors.New("fake error")
			return mockParse
		})

	// action
	got := buildOutPVData(ctx, "fake_backend", "fake_type", mockAllSBCInfo, mockAllPVData)

	// assert
	if !reflect.DeepEqual(got, mockOutPVData) {
		t.Errorf("buildOutPVData() got = %v, want %v", got, mockOutPVData)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_parseAllBackendInfo_Success(t *testing.T) {
	// arrange
	namespace := "mockNamespace"
	backendName := "mockBackendName"
	storageType := "mockStorageType"
	mockBackendList := &xuanwuV1.StorageBackendClaimList{
		TypeMeta: metaV1.TypeMeta{},
		ListMeta: metaV1.ListMeta{},
		Items: []xuanwuV1.StorageBackendClaim{{
			TypeMeta:   metaV1.TypeMeta{},
			ObjectMeta: metaV1.ObjectMeta{},
			Spec: xuanwuV1.StorageBackendClaimSpec{
				ConfigMapMeta: namespace + "/" + backendName,
			},
			Status: &xuanwuV1.StorageBackendClaimStatus{
				StorageType: storageType,
			},
		}},
	}
	wantRes := map[string]map[string]string{backendName: {"namespace": namespace, "sbcStorageType": storageType}}

	// act
	gotRes := parseAllBackendInfo(mockBackendList)

	// assert
	if !reflect.DeepEqual(gotRes, wantRes) {
		t.Errorf("Test_parseAllBackendInfo_Success failed, gotRes [%v], wantRes [%v]", gotRes, wantRes)
	}
}

func TestGetAndParsePVInfo_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	backendName := "mockBackendName"
	collectType := "mockCollectType"
	outPVData := &storageGRPC.CollectResponse{
		BackendName: "fake_backend",
		CollectType: "fake_type",
		Details:     []*storageGRPC.CollectDetail{}}
	wantRes := outPVData

	// mock
	p := gomonkey.NewPatches()
	p.ApplyFunc(getPVDataFromApi, func(ctx context.Context) []coreV1.PersistentVolume {
		return []coreV1.PersistentVolume{{Spec: coreV1.PersistentVolumeSpec{}}}
	}).ApplyFunc(getAllBackendFromApi, func(ctx context.Context) map[string]map[string]string {
		return map[string]map[string]string{
			backendName: {"namespace": "mockNamespace", "sbcStorageType": "mockStorageType"},
		}
	}).ApplyFunc(buildOutPVData, func(ctx context.Context, backendName, collectType string,
		allSBCInfo map[string]map[string]string, allPVData []coreV1.PersistentVolume) *storageGRPC.CollectResponse {
		return outPVData
	})

	// action
	gotRes, gotErr := GetAndParsePVInfo(ctx, backendName, collectType)

	// assert
	if !reflect.DeepEqual(gotRes, wantRes) {
		t.Errorf("TestGetAndParsePVInfo_Success failed, gotRes [%v], wantRes [%v]", gotRes, wantRes)
	}
	if gotErr != nil {
		t.Errorf("TestGetAndParsePVInfo_Success failed, gotErr [%v], wantErr [%v]", gotErr, nil)
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestGetAndParsePVInfo_GetPvFail(t *testing.T) {
	// arrange
	ctx := context.TODO()
	backendName := "mockBackendName"
	collectType := "mockCollectType"
	getPvErr := fmt.Errorf("can not get pv data, pv is empty")
	wantErr := getPvErr

	// mock
	p := gomonkey.NewPatches()
	p.ApplyFunc(getPVDataFromApi, func(ctx context.Context) []coreV1.PersistentVolume {
		return nil
	})

	// action
	gotRes, gotErr := GetAndParsePVInfo(ctx, backendName, collectType)

	// assert
	if gotRes != nil {
		t.Errorf("TestGetAndParsePVInfo_GetPvFail failed, gotRes [%v], wantRes [%v]", gotRes, nil)
	}
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestGetAndParsePVInfo_GetPvFail failed, gotErr [%v], wantErr [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}

func TestGetAndParsePVInfo_GetBackendFail(t *testing.T) {
	// arrange
	ctx := context.TODO()
	backendName := "mockBackendName"
	collectType := "mockCollectType"
	getBackendErr := fmt.Errorf("can not get sbc data, sbc is empty")
	wantErr := getBackendErr

	// mock
	p := gomonkey.NewPatches()
	p.ApplyFunc(getPVDataFromApi, func(ctx context.Context) []coreV1.PersistentVolume {
		return []coreV1.PersistentVolume{{Spec: coreV1.PersistentVolumeSpec{}}}
	}).ApplyFunc(getAllBackendFromApi, func(ctx context.Context) map[string]map[string]string {
		return nil
	})

	// action
	gotRes, gotErr := GetAndParsePVInfo(ctx, backendName, collectType)

	// assert
	if gotRes != nil {
		t.Errorf("TestGetAndParsePVInfo_GetBackendFail failed, gotRes [%v], wantRes [%v]", gotRes, nil)
	}
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestGetAndParsePVInfo_GetBackendFail failed, gotErr [%v], wantErr [%v]", gotErr, wantErr)
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}
