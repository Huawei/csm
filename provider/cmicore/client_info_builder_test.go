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

	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
	"github.com/huawei/csm/v2/storage/client/distributedstorage"
	"github.com/huawei/csm/v2/storage/constant"
)

func TestNewClientInfoBuilder_Success(t *testing.T) {
	// arrange
	ctx := context.Background()

	// action
	gotBuilder := NewClientInfoBuilder(ctx)

	// assert
	assert.NotNil(t, gotBuilder)
	assert.NotNil(t, gotBuilder.clientInfo)
	assert.Nil(t, gotBuilder.err)
}

func TestClientInfoBuilder_Build_Success(t *testing.T) {
	// arrange
	builder := NewClientInfoBuilder(context.Background())

	// action
	gotInfo, gotErr := builder.Build()

	// assert
	assert.NoError(t, gotErr)
	assert.Equal(t, ClientInfo{}, gotInfo)
}

func TestClientInfoBuilder_WithVolumeType_Success(t *testing.T) {
	// arrange
	builder := NewClientInfoBuilder(context.Background())

	// action
	gotBuilder := builder.WithVolumeType(constants.StorageNas)

	// assert
	assert.Equal(t, constants.NasVolume, gotBuilder.clientInfo.VolumeType)
	assert.Nil(t, gotBuilder.err)
}

func TestClientInfoBuilder_WithVolumeType_UnsupportedStorageType(t *testing.T) {
	// arrange
	builder := NewClientInfoBuilder(context.Background())

	// action
	gotBuilder := builder.WithVolumeType("unsupported")

	// assert
	assert.Error(t, gotBuilder.err)
	assert.Contains(t, gotBuilder.err.Error(), "unsupported storage type")
	assert.Empty(t, gotBuilder.clientInfo.VolumeType)
}

func TestClientInfoBuilder_WithVolumeType_ErrExisted(t *testing.T) {
	// arrange
	wantErr := errors.New("existed err")
	builder := NewClientInfoBuilder(context.Background())
	builder.err = wantErr

	// action
	gotBuilder := builder.WithVolumeType(constants.StorageNas)

	// assert
	assert.Equal(t, wantErr, gotBuilder.err)
	assert.Empty(t, gotBuilder.clientInfo.VolumeType)
}

func TestClientInfoBuilder_WithClient_ErrExisted(t *testing.T) {
	// arrange
	wantErr := errors.New("existed err")
	builder := NewClientInfoBuilder(context.Background())
	builder.err = wantErr

	// action
	gotBuilder := builder.WithClient(&constant.StorageBackendConfig{})

	// assert
	assert.Equal(t, wantErr, gotBuilder.err)
}

func TestClientInfoBuilder_WithClient_NewClientFailed(t *testing.T) {
	// arrange
	wantErr := errors.New("new client failed")
	builder := NewClientInfoBuilder(context.Background())
	config := &constant.StorageBackendConfig{
		StorageBackendName: "backend1",
		StorageType:        constants.StorageNas,
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(centralizedstorage.NewCentralizedClient, nil, wantErr)
	defer patches.Reset()

	// action
	gotBuilder := builder.WithClient(config)

	// assert
	assert.ErrorIs(t, gotBuilder.err, wantErr)
}

func TestClientInfoBuilder_WithClient_LoginFailed(t *testing.T) {
	// arrange
	wantErr := errors.New("login failed")
	builder := NewClientInfoBuilder(context.Background())
	client := &centralizedstorage.CentralizedClient{}
	config := &constant.StorageBackendConfig{
		StorageBackendName: "backend1",
		StorageType:        constants.StorageNas,
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(centralizedstorage.NewCentralizedClient, client, nil)
	patches.ApplyMethodReturn(client, "Login", wantErr)
	patches.ApplyMethodReturn(client, "Logout")
	defer patches.Reset()

	// action
	gotBuilder := builder.WithClient(config)

	// assert
	assert.ErrorIs(t, gotBuilder.err, wantErr)
}

func TestClientInfoBuilder_WithClient_Success(t *testing.T) {
	// arrange
	builder := NewClientInfoBuilder(context.Background())
	client := &centralizedstorage.CentralizedClient{}
	config := &constant.StorageBackendConfig{
		StorageBackendName: "backend1",
		StorageType:        constants.StorageNas,
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(centralizedstorage.NewCentralizedClient, client, nil)
	patches.ApplyMethodReturn(client, "Login", nil)
	defer patches.Reset()

	// action
	gotBuilder := builder.WithClient(config)

	// assert
	assert.Nil(t, gotBuilder.err)
	assert.Equal(t, "backend1", gotBuilder.clientInfo.StorageName)
	assert.Equal(t, constants.OceanStorage, gotBuilder.clientInfo.StorageType)
	assert.Equal(t, client, gotBuilder.clientInfo.Client)
}

func TestWithVolumeType_FusionStorageNas(t *testing.T) {
	// arrange
	builder := NewClientInfoBuilder(context.Background())

	// act
	builder.WithVolumeType(constants.StorageFusionNas)

	// assert
	info, err := builder.Build()
	assert.NoError(t, err)
	assert.Equal(t, constants.NasVolume, info.VolumeType)
}

func TestVolumeTypesMap_FusionStorageNas(t *testing.T) {
	// arrange & act & assert
	volumeType, ok := volumeTypes[constants.StorageFusionNas]
	assert.True(t, ok, "fusionstorage-nas should be in volumeTypes map")
	assert.Equal(t, constants.NasVolume, volumeType)
}

func TestClientInfoBuilder_WithClient_FusionStorageNas_NewDistributedClientFailed(t *testing.T) {
	// arrange
	wantErr := errors.New("new distributed client failed")
	builder := NewClientInfoBuilder(context.Background())
	config := &constant.StorageBackendConfig{
		StorageBackendName: "fusion-backend1",
		StorageType:        constants.StorageFusionNas,
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(distributedstorage.NewDistributedClient, nil, wantErr)
	defer patches.Reset()

	// action
	gotBuilder := builder.WithClient(config)

	// assert
	assert.ErrorIs(t, gotBuilder.err, wantErr)
}

func TestClientInfoBuilder_WithClient_FusionStorageNas_LoginFailed(t *testing.T) {
	// arrange
	wantErr := errors.New("login failed")
	builder := NewClientInfoBuilder(context.Background())
	client := &distributedstorage.DistributedClient{}
	config := &constant.StorageBackendConfig{
		StorageBackendName: "fusion-backend1",
		StorageType:        constants.StorageFusionNas,
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(distributedstorage.NewDistributedClient, client, nil)
	patches.ApplyMethodReturn(client, "Login", wantErr)
	patches.ApplyMethodReturn(client, "Logout")
	defer patches.Reset()

	// action
	gotBuilder := builder.WithClient(config)

	// assert
	assert.ErrorIs(t, gotBuilder.err, wantErr)
}

func TestClientInfoBuilder_WithClient_FusionStorageNas_Success(t *testing.T) {
	// arrange
	builder := NewClientInfoBuilder(context.Background())
	client := &distributedstorage.DistributedClient{}
	config := &constant.StorageBackendConfig{
		StorageBackendName: "fusion-backend1",
		StorageType:        constants.StorageFusionNas,
	}

	patches := gomonkey.NewPatches()
	patches.ApplyFuncReturn(distributedstorage.NewDistributedClient, client, nil)
	patches.ApplyMethodReturn(client, "Login", nil)
	defer patches.Reset()

	// action
	gotBuilder := builder.WithClient(config)

	// assert
	assert.Nil(t, gotBuilder.err)
	assert.Equal(t, "fusion-backend1", gotBuilder.clientInfo.StorageName)
	assert.Equal(t, constants.FusionStorage, gotBuilder.clientInfo.StorageType)
	assert.Equal(t, client, gotBuilder.clientInfo.Client)
}
