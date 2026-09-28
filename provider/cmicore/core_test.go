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
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"github.com/huawei/csm/v2/provider/constants"
)

func TestNewCore_NilConfig(t *testing.T) {
	// arrange
	var config *CoreConfig

	// action & assert
	assert.Panics(t, func() {
		NewCore(config)
	})
}

func TestNewCore_Success(t *testing.T) {
	// arrange
	config := &CoreConfig{
		BackendNamespace:     "huawei-csi",
		QueryStoragePageSize: 100,
		ClientMaxThreads:     20,
	}

	// action
	gotCore := NewCore(config)

	// assert
	assert.NotNil(t, gotCore)
	assert.NotNil(t, gotCore.clientCache)
	assert.NotNil(t, gotCore.informer)
	assert.NotNil(t, gotCore.labelService)
	assert.NotNil(t, gotCore.collectorFactory)
	assert.NotNil(t, gotCore.collectValidator)
}

func TestCore_StartSuccess(t *testing.T) {
	// arrange
	config := &CoreConfig{
		BackendNamespace:     "huawei-csi",
		QueryStoragePageSize: 100,
		ClientMaxThreads:     20,
	}
	core := NewCore(config)

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&BackendInformer{}, "Start", nil)
	patches.ApplyMethodReturn(&BackendInformer{}, "Stop")
	defer patches.Reset()

	// action
	gotErr := core.Start(context.Background())

	// assert
	assert.NoError(t, gotErr)

	// cleanup
	core.Stop()
}

func TestCore_StartIdempotent(t *testing.T) {
	// arrange
	config := &CoreConfig{
		BackendNamespace:     "huawei-csi",
		QueryStoragePageSize: 100,
		ClientMaxThreads:     20,
	}
	core := NewCore(config)

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&BackendInformer{}, "Start", nil)
	patches.ApplyMethodReturn(&BackendInformer{}, "Stop")
	defer patches.Reset()

	_ = core.Start(context.Background())

	// action - start again should not error
	gotErr := core.Start(context.Background())

	// assert
	assert.NoError(t, gotErr)

	// cleanup
	core.Stop()
}

func TestCore_StopIdempotent(t *testing.T) {
	// arrange
	config := &CoreConfig{
		BackendNamespace:     "huawei-csi",
		QueryStoragePageSize: 100,
		ClientMaxThreads:     20,
	}
	core := NewCore(config)

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(&BackendInformer{}, "Start", nil)
	patches.ApplyMethodReturn(&BackendInformer{}, "Stop")
	defer patches.Reset()

	_ = core.Start(context.Background())

	// action - Stop twice should not panic
	assert.NotPanics(t, func() {
		core.Stop()
		core.Stop()
	})
}

func TestCore_CreateLabel_NilLabelService(t *testing.T) {
	// arrange
	config := &CoreConfig{
		BackendNamespace:     "huawei-csi",
		QueryStoragePageSize: 100,
		ClientMaxThreads:     20,
	}
	core := NewCore(config)
	core.labelService = nil

	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "test-label",
		Kind:      "PersistentVolume",
	}

	// action
	gotErr := core.CreateLabel(context.Background(), req)

	// assert
	assert.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "not initialized")
}

func TestCore_DeleteLabel_NilLabelService(t *testing.T) {
	// arrange
	config := &CoreConfig{
		BackendNamespace:     "huawei-csi",
		QueryStoragePageSize: 100,
		ClientMaxThreads:     20,
	}
	core := NewCore(config)
	core.labelService = nil

	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "test-label",
		Kind:      "PersistentVolume",
	}

	// action
	gotErr := core.DeleteLabel(context.Background(), req)

	// assert
	assert.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "not initialized")
}

func TestCore_GetStorageType_Success(t *testing.T) {
	// arrange
	config := &CoreConfig{
		BackendNamespace:     "huawei-csi",
		QueryStoragePageSize: 100,
		ClientMaxThreads:     20,
	}
	core := NewCore(config)
	core.clientCache.RegisterClient("backend1", ClientInfo{
		StorageName: "backend1",
		StorageType: constants.OceanStorage,
		Client:      &mockRestClient{},
	})
	wantStorageType := constants.OceanStorage

	// action
	gotStorageType, gotErr := core.GetStorageType("backend1")

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, wantStorageType, gotStorageType)
}

func TestCore_GetStorageType_BackendNotFound(t *testing.T) {
	// arrange
	config := &CoreConfig{
		BackendNamespace:     "huawei-csi",
		QueryStoragePageSize: 100,
		ClientMaxThreads:     20,
	}
	core := NewCore(config)
	wantErr := fmt.Errorf("backend [nonexistent] not found in client cache")

	// action
	gotStorageType, gotErr := core.GetStorageType("nonexistent")

	// assert
	assert.Error(t, gotErr)
	assert.Equal(t, wantErr, gotErr)
	assert.Empty(t, gotStorageType)
}

func TestCore_Collect_ValidateFailed(t *testing.T) {
	// arrange
	config := &CoreConfig{
		BackendNamespace:     "huawei-csi",
		QueryStoragePageSize: 100,
		ClientMaxThreads:     20,
	}
	core := NewCore(config)
	req := &CollectRequest{
		BackendName: "",
		CollectType: constants.Lun,
		MetricsType: constants.Object,
	}

	// action
	gotResp, gotErr := core.Collect(context.Background(), req)

	// assert
	assert.Error(t, gotErr)
	assert.Nil(t, gotResp)
}

func TestCore_Collect_GetCollectorFailed(t *testing.T) {
	// arrange
	config := &CoreConfig{
		BackendNamespace:     "huawei-csi",
		QueryStoragePageSize: 100,
		ClientMaxThreads:     20,
	}
	core := NewCore(config)
	req := &CollectRequest{
		BackendName: "backend1",
		CollectType: constants.Lun,
		MetricsType: "unsupported",
	}

	// action
	gotResp, gotErr := core.Collect(context.Background(), req)

	// assert
	assert.Error(t, gotErr)
	assert.Nil(t, gotResp)
}

func TestCore_Collect_Success(t *testing.T) {
	// arrange
	config := &CoreConfig{
		BackendNamespace:     "huawei-csi",
		QueryStoragePageSize: 100,
		ClientMaxThreads:     20,
	}
	core := NewCore(config)

	// Pre-register a client in core's clientCache (same reference used by CollectorFactory)
	core.clientCache.RegisterClient("backend1", ClientInfo{
		StorageName: "backend1",
		StorageType: "test-storage",
		VolumeType:  constants.LunVolume,
		Client:      &mockRestClient{},
	})

	// Register a test object handler
	RegisterObjectHandlerDirect("test-storage", constants.Lun,
		func(ctx context.Context, client RestClient, req *CollectRequest) (*CollectResponse, error) {
			return &CollectResponse{
				BackendName: req.BackendName,
				CollectType: req.CollectType,
				MetricsType: req.MetricsType,
			}, nil
		})

	req := &CollectRequest{
		BackendName: "backend1",
		CollectType: constants.Lun,
		MetricsType: constants.Object,
	}

	// action
	gotResp, gotErr := core.Collect(context.Background(), req)

	// assert
	assert.NoError(t, gotErr)
	assert.NotNil(t, gotResp)
	assert.Equal(t, "backend1", gotResp.BackendName)
}
