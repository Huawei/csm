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

// Package cmicore provides a shared library for storage backend operations
package cmicore

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	csiV1 "github.com/Huawei/eSDK_K8S_Plugin/v4/client/apis/xuanwu/v1"
	sbcFake "github.com/Huawei/eSDK_K8S_Plugin/v4/pkg/client/clientset/versioned/fake"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
)

// buildTestSBC creates a StorageBackendClaim for testing
func buildTestSBC(backendName, namespace string) *csiV1.StorageBackendClaim {
	return &csiV1.StorageBackendClaim{
		ObjectMeta: metaV1.ObjectMeta{Name: backendName, Namespace: namespace},
		Spec: csiV1.StorageBackendClaimSpec{
			SecretMeta:    namespace + "/" + backendName + "-secret",
			ConfigMapMeta: namespace + "/" + backendName + "-configmap",
		},
		Status: &csiV1.StorageBackendClaimStatus{
			SecretMeta:    namespace + "/" + backendName + "-secret",
			ConfigmapMeta: namespace + "/" + backendName + "-configmap",
		},
	}
}

// buildTestSecret creates a Kubernetes Secret for testing
func buildTestSecret(name, namespace string) *coreV1.Secret {
	return &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{Name: name, Namespace: namespace},
		Data: map[string][]byte{
			"user": []byte("admin"),
		},
	}
}

// buildTestConfigMap creates a Kubernetes ConfigMap for testing
func buildTestConfigMap(name, namespace string) *coreV1.ConfigMap {
	csiJSON := `{"backends":{"storage":"oceanstor-nas","urls":["https://10.0.0.1:8088"]}}`
	return &coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{Name: name, Namespace: namespace},
		Data: map[string]string{
			"csi.json": csiJSON,
		},
	}
}

func TestCore_DiscoverClient_SbcNotFound(t *testing.T) {
	// arrange
	namespace := "huawei-csi"
	kubeClient := fake.NewSimpleClientset()
	sbcClient := sbcFake.NewSimpleClientset() // empty - no SBC registered
	config := &CoreConfig{
		KubeClient:       kubeClient,
		SbcClient:        sbcClient,
		BackendNamespace: namespace,
	}
	core := NewCore(config)

	// action
	gotInfo, gotErr := core.discoverClient(context.Background(), "backend1")

	// assert
	assert.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "storagebackendclaims")
	assert.Equal(t, ClientInfo{}, gotInfo)
}

func TestCore_DiscoverClient_NewClientFailed(t *testing.T) {
	// arrange
	namespace := "huawei-csi"
	backendName := "backend1"
	wantErr := errors.New("new client failed")

	kubeClient := fake.NewSimpleClientset(
		buildTestSecret(backendName+"-secret", namespace),
		buildTestConfigMap(backendName+"-configmap", namespace),
	)
	sbcClient := sbcFake.NewSimpleClientset(buildTestSBC(backendName, namespace))
	config := &CoreConfig{
		KubeClient:       kubeClient,
		SbcClient:        sbcClient,
		BackendNamespace: namespace,
	}
	core := NewCore(config)

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(centralizedstorage.NewCentralizedClient, nil, wantErr)
	defer patches.Reset()

	// action
	gotInfo, gotErr := core.discoverClient(context.Background(), backendName)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
	assert.Equal(t, ClientInfo{VolumeType: constants.NasVolume}, gotInfo)
}

func TestCore_DiscoverClient_LoginFailed(t *testing.T) {
	// arrange
	namespace := "huawei-csi"
	backendName := "backend1"
	wantErr := errors.New("login failed")
	client := &centralizedstorage.CentralizedClient{}

	kubeClient := fake.NewSimpleClientset(
		buildTestSecret(backendName+"-secret", namespace),
		buildTestConfigMap(backendName+"-configmap", namespace),
	)
	sbcClient := sbcFake.NewSimpleClientset(buildTestSBC(backendName, namespace))
	config := &CoreConfig{
		KubeClient:       kubeClient,
		SbcClient:        sbcClient,
		BackendNamespace: namespace,
	}
	core := NewCore(config)

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(centralizedstorage.NewCentralizedClient, client, nil)
	patches.ApplyMethodReturn(client, "Login", wantErr)
	patches.ApplyMethodReturn(client, "Logout")
	defer patches.Reset()

	// action
	gotInfo, gotErr := core.discoverClient(context.Background(), backendName)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
	assert.Equal(t, ClientInfo{VolumeType: constants.NasVolume}, gotInfo)
}

func TestCore_DiscoverClient_Success(t *testing.T) {
	// arrange
	namespace := "huawei-csi"
	backendName := "backend1"
	client := &centralizedstorage.CentralizedClient{}

	kubeClient := fake.NewSimpleClientset(
		buildTestSecret(backendName+"-secret", namespace),
		buildTestConfigMap(backendName+"-configmap", namespace),
	)
	sbcClient := sbcFake.NewSimpleClientset(buildTestSBC(backendName, namespace))
	config := &CoreConfig{
		KubeClient:       kubeClient,
		SbcClient:        sbcClient,
		BackendNamespace: namespace,
	}
	core := NewCore(config)

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(centralizedstorage.NewCentralizedClient, client, nil)
	patches.ApplyMethodReturn(client, "Login", nil)
	defer patches.Reset()

	// action
	gotInfo, gotErr := core.discoverClient(context.Background(), backendName)

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, backendName, gotInfo.StorageName)
	assert.Equal(t, constants.OceanStorage, gotInfo.StorageType)
	assert.Equal(t, constants.NasVolume, gotInfo.VolumeType)
	assert.Equal(t, client, gotInfo.Client)
}
