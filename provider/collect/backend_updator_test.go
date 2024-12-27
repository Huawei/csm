/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

// Package collect is a package that provides object and performance collect
package collect

import (
	"context"
	"reflect"
	"testing"

	csiV1 "github.com/Huawei/eSDK_K8S_Plugin/v4/client/apis/xuanwu/v1"
	"github.com/agiledragon/gomonkey/v2"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/huawei/csm/v2/provider/backend"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
)

func TestReleaseCache(t *testing.T) {
	//arrange
	backendName := "backend"
	var wantErr error
	client := &centralizedstorage.CentralizedClient{}
	clientInfo := backend.ClientInfo{Client: client}
	RegisterClient(backendName, clientInfo)

	//mock
	patches := gomonkey.NewPatches()
	patches.ApplyMethod(client, "Logout",
		func(_ *centralizedstorage.CentralizedClient, ctx context.Context) {})

	//act
	err := releaseCache(backendName)

	//assert
	if !reflect.DeepEqual(err, wantErr) {
		t.Errorf("releaseCache() error = %v, wantErr %v", err, wantErr)
	}

	//clean
	t.Cleanup(func() {
		patches.Reset()
	})

}

func TestUpdateBackendCache_SameSpec(t *testing.T) {
	//arrange
	oldObj := &csiV1.StorageBackendClaim{}
	newObj := &csiV1.StorageBackendClaim{}
	backendName := "backend"
	client := &centralizedstorage.CentralizedClient{}
	clientInfo := backend.ClientInfo{StorageName: "storage", Client: client}
	RegisterClient(backendName, clientInfo)

	//act
	updateBackendCache(oldObj, newObj)

	//assert
	if info, ok := clientCache[backendName]; !ok || !reflect.DeepEqual(info, clientInfo) {
		t.Errorf("TestUpdateBackendCache_SameSpec() failed")
	}

	//clean
	t.Cleanup(func() {
		RemoveClient(backendName)
	})
}

func TestUpdateBackendCache_DifferentSpec(t *testing.T) {
	//arrange
	backendName := "backend"
	oldObj := &csiV1.StorageBackendClaim{
		ObjectMeta: metaV1.ObjectMeta{Name: backendName},
		Spec:       csiV1.StorageBackendClaimSpec{SecretMeta: "meta1"},
	}
	newObj := &csiV1.StorageBackendClaim{
		ObjectMeta: metaV1.ObjectMeta{Name: backendName},
		Spec:       csiV1.StorageBackendClaimSpec{SecretMeta: "meta2"},
	}
	client := &centralizedstorage.CentralizedClient{}
	clientInfo := backend.ClientInfo{StorageName: "storage", Client: client}
	RegisterClient(backendName, clientInfo)

	//mock
	patches := gomonkey.NewPatches()
	patches.ApplyMethod(client, "Logout",
		func(_ *centralizedstorage.CentralizedClient, ctx context.Context) {})
	patches.ApplyFunc(GetClient, func(ctx context.Context, backendName string,
		discoverFunc func(context.Context, string) (backend.ClientInfo, error)) (backend.ClientInfo, error) {
		clientInfo := backend.ClientInfo{StorageName: "storage", StorageType: "type1", Client: client}
		RegisterClient(backendName, clientInfo)
		return backend.ClientInfo{}, nil
	})

	//act
	updateBackendCache(oldObj, newObj)

	//assert
	if info, ok := clientCache[backendName]; !ok || reflect.DeepEqual(info, clientInfo) {
		t.Errorf("TestUpdateBackendCache_DifferentSpec() failed")
	}

	//clean
	t.Cleanup(func() {
		RemoveClient(backendName)
		patches.Reset()
	})
}

func TestDeleteBackendCache(t *testing.T) {
	//arrange
	backendName := "backend"
	obj := &csiV1.StorageBackendClaim{
		ObjectMeta: metaV1.ObjectMeta{Name: backendName},
		Spec:       csiV1.StorageBackendClaimSpec{SecretMeta: "meta1"},
	}
	client := &centralizedstorage.CentralizedClient{}
	clientInfo := backend.ClientInfo{StorageName: "storage", Client: client}
	RegisterClient(backendName, clientInfo)

	//mock
	patches := gomonkey.NewPatches()
	patches.ApplyMethod(client, "Logout",
		func(_ *centralizedstorage.CentralizedClient, ctx context.Context) {})

	//act
	deleteBackendCache(obj)

	//assert
	if _, ok := clientCache[backendName]; ok {
		t.Errorf("TestDeleteBackendCache() failed")
	}

	//clean
	t.Cleanup(func() {
		patches.Reset()
	})
}
