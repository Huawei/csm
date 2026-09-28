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
)

func TestNewOceanStorageLabelService_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{}, nil
	}

	// action
	gotService := NewOceanStorageLabelService(cache, discoverFunc)

	// assert
	assert.NotNil(t, gotService)
}

func TestOceanStorageLabelService_CreateLabel_DiscoverClientFailed(t *testing.T) {
	// arrange
	cache := NewClientCache()
	wantErr := errors.New("discover client failed")
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{}, wantErr
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)
	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pv-name",
		Kind:      constants.PersistentVolumeKind,
	}

	// action
	gotErr := service.CreateLabel(context.Background(), req)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestOceanStorageLabelService_CreateLabel_ConvertClientFailed(t *testing.T) {
	// arrange
	cache := NewClientCache()
	// Return a ClientInfo with a RestClient that is NOT *centralizedstorage.CentralizedClient
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      &mockRestClient{},
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)
	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pv-name",
		Kind:      constants.PersistentVolumeKind,
	}

	// action
	gotErr := service.CreateLabel(context.Background(), req)

	// assert
	assert.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "convert storage client failed")
}

func TestOceanStorageLabelService_CreateLabel_ResourceIdNotFound(t *testing.T) {
	// arrange
	cache := NewClientCache()
	client := &centralizedstorage.CentralizedClient{}
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      client,
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(client, "GetLunIdByName", "", nil)
	defer patches.Reset()

	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pv-name",
		Kind:      constants.PersistentVolumeKind,
	}

	// action
	gotErr := service.CreateLabel(context.Background(), req)

	// assert
	assert.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "not found resource id")
}

func TestOceanStorageLabelService_CreateLabel_UnsupportedKind(t *testing.T) {
	// arrange
	cache := NewClientCache()
	client := &centralizedstorage.CentralizedClient{}
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      client,
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(client, "GetLunIdByName", "resource-id-123", nil)
	defer patches.Reset()

	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pv-name",
		Kind:      "UnsupportedKind",
	}

	// action
	gotErr := service.CreateLabel(context.Background(), req)

	// assert
	assert.Error(t, gotErr)
	assert.Contains(t, gotErr.Error(), "unsupported resource kind")
}

func TestOceanStorageLabelService_CreateLabel_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	client := &centralizedstorage.CentralizedClient{}
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      client,
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(client, "GetLunIdByName", "resource-id-123", nil)
	patches.ApplyMethodReturn(client, "CreatePvLabel", map[string]interface{}{}, nil)
	defer patches.Reset()

	req := &LabelRequest{
		VolumeId:    "backend/volume",
		LabelName:   "pv-name",
		Kind:        constants.PersistentVolumeKind,
		ClusterName: "cluster1",
	}

	// action
	gotErr := service.CreateLabel(context.Background(), req)

	// assert
	assert.NoError(t, gotErr)
}

func TestOceanStorageLabelService_DeleteLabel_ResourceIdNotFoundReturnsNil(t *testing.T) {
	// arrange
	cache := NewClientCache()
	client := &centralizedstorage.CentralizedClient{}
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      client,
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(client, "GetLunIdByName", "", nil)
	defer patches.Reset()

	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pv-name",
		Kind:      constants.PersistentVolumeKind,
	}

	// action - DeleteLabel should return nil (idempotent) when resourceId is empty
	gotErr := service.DeleteLabel(context.Background(), req)

	// assert
	assert.NoError(t, gotErr)
}

func TestOceanStorageLabelService_DeleteLabel_DiscoverClientFailed(t *testing.T) {
	// arrange
	cache := NewClientCache()
	wantErr := errors.New("discover client failed")
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{}, wantErr
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)
	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pv-name",
		Kind:      constants.PersistentVolumeKind,
	}

	// action
	gotErr := service.DeleteLabel(context.Background(), req)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestOceanStorageLabelService_DeleteLabel_Success(t *testing.T) {
	// arrange
	cache := NewClientCache()
	client := &centralizedstorage.CentralizedClient{}
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      client,
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(client, "GetLunIdByName", "resource-id-123", nil)
	patches.ApplyMethodReturn(client, "DeletePvLabel", map[string]interface{}{}, nil)
	defer patches.Reset()

	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pv-name",
		Kind:      constants.PersistentVolumeKind,
	}

	// action
	gotErr := service.DeleteLabel(context.Background(), req)

	// assert
	assert.NoError(t, gotErr)
}

func TestOceanStorageLabelService_CreateLabel_PvLabelFailed(t *testing.T) {
	// arrange
	cache := NewClientCache()
	client := &centralizedstorage.CentralizedClient{}
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      client,
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)

	wantErr := errors.New("create pv label error")
	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(client, "GetLunIdByName", "resource-id-123", nil)
	patches.ApplyMethodReturn(client, "CreatePvLabel", nil, wantErr)
	defer patches.Reset()

	req := &LabelRequest{
		VolumeId:    "backend/volume",
		LabelName:   "pv-name",
		Kind:        constants.PersistentVolumeKind,
		ClusterName: "cluster1",
	}

	// action
	gotErr := service.CreateLabel(context.Background(), req)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestOceanStorageLabelService_CreateLabel_PodLabelSuccess(t *testing.T) {
	// arrange
	cache := NewClientCache()
	client := &centralizedstorage.CentralizedClient{}
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      client,
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(client, "GetLunIdByName", "resource-id-123", nil)
	patches.ApplyMethodReturn(client, "CreatePodLabel", map[string]interface{}{}, nil)
	defer patches.Reset()

	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pod-name",
		Kind:      constants.PodKind,
		Namespace: "test-ns",
	}

	// action
	gotErr := service.CreateLabel(context.Background(), req)

	// assert
	assert.NoError(t, gotErr)
}

func TestOceanStorageLabelService_CreateLabel_PodLabelFailed(t *testing.T) {
	// arrange
	cache := NewClientCache()
	client := &centralizedstorage.CentralizedClient{}
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      client,
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)

	wantErr := errors.New("create pod label error")
	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(client, "GetLunIdByName", "resource-id-123", nil)
	patches.ApplyMethodReturn(client, "CreatePodLabel", nil, wantErr)
	defer patches.Reset()

	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pod-name",
		Kind:      constants.PodKind,
		Namespace: "test-ns",
	}

	// action
	gotErr := service.CreateLabel(context.Background(), req)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestOceanStorageLabelService_DeleteLabel_PvLabelFailed(t *testing.T) {
	// arrange
	cache := NewClientCache()
	client := &centralizedstorage.CentralizedClient{}
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      client,
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)

	wantErr := errors.New("delete pv label error")
	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(client, "GetLunIdByName", "resource-id-123", nil)
	patches.ApplyMethodReturn(client, "DeletePvLabel", nil, wantErr)
	defer patches.Reset()

	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pv-name",
		Kind:      constants.PersistentVolumeKind,
	}

	// action
	gotErr := service.DeleteLabel(context.Background(), req)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestOceanStorageLabelService_DeleteLabel_PodLabelSuccess(t *testing.T) {
	// arrange
	cache := NewClientCache()
	client := &centralizedstorage.CentralizedClient{}
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      client,
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)

	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(client, "GetLunIdByName", "resource-id-123", nil)
	patches.ApplyMethodReturn(client, "DeletePodLabel", map[string]interface{}{}, nil)
	defer patches.Reset()

	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pod-name",
		Kind:      constants.PodKind,
		Namespace: "test-ns",
	}

	// action
	gotErr := service.DeleteLabel(context.Background(), req)

	// assert
	assert.NoError(t, gotErr)
}

func TestOceanStorageLabelService_DeleteLabel_PodLabelFailed(t *testing.T) {
	// arrange
	cache := NewClientCache()
	client := &centralizedstorage.CentralizedClient{}
	discoverFunc := func(_ context.Context, _ string) (ClientInfo, error) {
		return ClientInfo{
			StorageName: "backend",
			StorageType: "oceanStorage",
			VolumeType:  "lun",
			Client:      client,
		}, nil
	}
	service := NewOceanStorageLabelService(cache, discoverFunc)

	wantErr := errors.New("delete pod label error")
	patches := gomonkey.NewPatches()
	patches.ApplyMethodReturn(client, "GetLunIdByName", "resource-id-123", nil)
	patches.ApplyMethodReturn(client, "DeletePodLabel", nil, wantErr)
	defer patches.Reset()

	req := &LabelRequest{
		VolumeId:  "backend/volume",
		LabelName: "pod-name",
		Kind:      constants.PodKind,
		Namespace: "test-ns",
	}

	// action
	gotErr := service.DeleteLabel(context.Background(), req)

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}
