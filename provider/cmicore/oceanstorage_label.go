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
	"fmt"

	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/provider/utils"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
	"github.com/huawei/csm/v2/utils/log"
)

// createLabelFunction is the function signature for creating labels
type createLabelFunction func(ctx context.Context, resourceId, resourceType string,
	client *centralizedstorage.CentralizedClient, req *LabelRequest) error

// deleteLabelFunction is the function signature for deleting labels
type deleteLabelFunction func(ctx context.Context, resourceId, resourceType string,
	client *centralizedstorage.CentralizedClient, req *LabelRequest) error

// createLabelFunctions maps resource kind to create label function
var createLabelFunctions = map[string]createLabelFunction{
	constants.PersistentVolumeKind: createPvLabel,
	constants.PodKind:              createPodLabel,
}

// deleteLabelFunctions maps resource kind to delete label function
var deleteLabelFunctions = map[string]deleteLabelFunction{
	constants.PersistentVolumeKind: deletePvLabel,
	constants.PodKind:              deletePodLabel,
}

// OceanStorageLabelService implements LabelService for OceanStorage backends
type OceanStorageLabelService struct {
	clientCache  *ClientCache
	discoverFunc func(context.Context, string) (ClientInfo, error)
}

// NewOceanStorageLabelService creates a new OceanStorageLabelService
func NewOceanStorageLabelService(
	clientCache *ClientCache,
	discoverFunc func(context.Context, string) (ClientInfo, error),
) *OceanStorageLabelService {
	return &OceanStorageLabelService{
		clientCache:  clientCache,
		discoverFunc: discoverFunc,
	}
}

// CreateLabel creates a label on the ocean storage backend
func (o *OceanStorageLabelService) CreateLabel(ctx context.Context, req *LabelRequest) error {
	param, err := o.prepareLabelRequest(ctx, req.VolumeId)
	if err != nil {
		log.AddContext(ctx).Errorf("create label failed, volumeId: %s, error: %v", req.VolumeId, err)
		return err
	}

	if param.resourceId == "" {
		log.AddContext(ctx).Errorln("not found resource id, perhaps the volume does not exist, " +
			"so returning failed")
		return errors.New("not found resource id")
	}

	if req.Namespace == "" {
		req.Namespace = constants.DefaultNameSpace
	}

	fun, ok := createLabelFunctions[req.Kind]
	if !ok {
		return fmt.Errorf("illegalArgumentError unsupported resource kind [%s]", req.Kind)
	}

	return fun(ctx, param.resourceId, param.resourceType, param.client, req)
}

// DeleteLabel deletes a label from the ocean storage backend
func (o *OceanStorageLabelService) DeleteLabel(ctx context.Context, req *LabelRequest) error {
	param, err := o.prepareLabelRequest(ctx, req.VolumeId)
	if err != nil {
		log.AddContext(ctx).Errorf("delete label failed, volumeId: %s, error: %v", req.VolumeId, err)
		return err
	}

	if param.resourceId == "" {
		// Return nil (idempotent success) for DeleteLabel when resource is not found,
		// since deleting a non-existent resource is a no-op and should not cause errors.
		log.AddContext(ctx).Infoln("not found resource id, perhaps the volume does not exist, " +
			"so returning success")
		return nil
	}

	if req.Namespace == "" {
		req.Namespace = constants.DefaultNameSpace
	}

	fun, ok := deleteLabelFunctions[req.Kind]
	if !ok {
		return fmt.Errorf("illegalArgumentError unsupported resource kind [%s]", req.Kind)
	}

	return fun(ctx, param.resourceId, param.resourceType, param.client, req)
}

// oceanStorageLabelParam holds the parameters for label operations
type oceanStorageLabelParam struct {
	resourceId   string
	resourceType string
	client       *centralizedstorage.CentralizedClient
}

// prepareLabelRequest gets client and resource object information
func (o *OceanStorageLabelService) prepareLabelRequest(ctx context.Context,
	volumeId string) (oceanStorageLabelParam, error) {
	backendName, volumeName := utils.SplitVolumeId(volumeId)

	clientInfo, err := o.clientCache.DiscoverClient(ctx, backendName, o.discoverFunc)
	if err != nil {
		log.AddContext(ctx).Errorf("get client failed, error: %v", err)
		return oceanStorageLabelParam{}, err
	}

	// Type assertion to *centralizedstorage.CentralizedClient is required here because
	// OceanStorage label operations need CentralizedClient-specific methods
	// (CreatePvLabel, DeletePvLabel, etc.) that are not part of the RestClient interface.
	// This is intentional: OceanStorageLabelService is an OceanStorage-specific implementation.
	client, ok := clientInfo.Client.(*centralizedstorage.CentralizedClient)
	if !ok {
		return oceanStorageLabelParam{}, errors.New("convert storage client failed")
	}

	resourceType := getResourceType(clientInfo.VolumeType)
	resourceId, err := getResourceId(ctx, volumeName, clientInfo.VolumeType, client)
	if err != nil {
		log.AddContext(ctx).Errorf("get resource id failed, error: %v", err)
		return oceanStorageLabelParam{}, err
	}

	return oceanStorageLabelParam{
		resourceId:   resourceId,
		resourceType: resourceType,
		client:       client,
	}, nil
}

// getResourceId gets the resource ID by volume name and type
func getResourceId(ctx context.Context, volumeName, volumeType string,
	client *centralizedstorage.CentralizedClient) (string, error) {
	if volumeType == constants.NasVolume {
		return client.GetFileSystemIdByName(ctx, volumeName)
	}
	return client.GetLunIdByName(ctx, volumeName)
}

// getResourceType returns the resource type for the volume type
func getResourceType(volumeType string) string {
	if volumeType == constants.NasVolume {
		return constants.ResourceTypeFilesystem
	}
	return constants.ResourceTypeLun
}

// createPvLabel creates a PV label
func createPvLabel(ctx context.Context, resourceId, resourceType string,
	client *centralizedstorage.CentralizedClient, req *LabelRequest) error {
	data := centralizedstorage.PvLabelRequest{
		ResourceId:   resourceId,
		ResourceType: resourceType,
		PvName:       req.LabelName,
		ClusterName:  req.ClusterName,
	}
	_, err := client.CreatePvLabel(ctx, data)
	if err != nil {
		log.AddContext(ctx).Errorf("create pv label failed, volumeId: %s, error: %v", req.VolumeId, err)
		return err
	}
	return nil
}

// createPodLabel creates a Pod label
func createPodLabel(ctx context.Context, resourceId, resourceType string,
	client *centralizedstorage.CentralizedClient, req *LabelRequest) error {
	data := centralizedstorage.PodLabelRequest{
		ResourceId:   resourceId,
		ResourceType: resourceType,
		PodName:      req.LabelName,
		NameSpace:    req.Namespace,
	}
	_, err := client.CreatePodLabel(ctx, data)
	if err != nil {
		log.AddContext(ctx).Errorf("create pod label failed, volumeId: %s, error: %v", req.VolumeId, err)
		return err
	}
	return nil
}

// deletePvLabel deletes a PV label
func deletePvLabel(ctx context.Context, resourceId, resourceType string,
	client *centralizedstorage.CentralizedClient, req *LabelRequest) error {
	_, err := client.DeletePvLabel(ctx, resourceId, resourceType)
	if err != nil {
		log.AddContext(ctx).Errorf("delete pv label failed, volumeId: %s, error: %v", req.VolumeId, err)
		return err
	}
	return nil
}

// deletePodLabel deletes a Pod label
func deletePodLabel(ctx context.Context, resourceId, resourceType string,
	client *centralizedstorage.CentralizedClient, req *LabelRequest) error {
	data := centralizedstorage.PodLabelRequest{
		ResourceId:   resourceId,
		ResourceType: resourceType,
		PodName:      req.LabelName,
		NameSpace:    req.Namespace,
	}
	_, err := client.DeletePodLabel(ctx, data)
	if err != nil {
		log.AddContext(ctx).Errorf("delete pod label failed, volumeId: %s, error: %v", req.VolumeId, err)
		return err
	}
	return nil
}
