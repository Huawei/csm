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

// Package label is a package that provide operation storage label
package label

import (
	"context"
	"errors"
	"fmt"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
	"github.com/huawei/csm/v2/utils/log"
)

// createLabelFunctions create label functions
// including creating pod and pv
var createLabelFunctions = map[string]createLabelFunction{
	constants.PersistentVolumeKind: createPvLabel,
	constants.PodKind:              createPodLabel,
}

// deleteLabelFunctions delete label functions
// including deleting pod and pv
var deleteLabelFunctions = map[string]deleteLabelFunction{
	constants.PersistentVolumeKind: deletePvLabel,
	constants.PodKind:              deletePodLabel,
}

// createLabelFunction create label function format
type createLabelFunction func(ctx context.Context, resourceId, resourceType string,
	client *centralizedstorage.CentralizedClient, request *cmi.CreateLabelRequest) (*cmi.CreateLabelResponse, error)

// deleteLabelFunction delete label function format
type deleteLabelFunction func(ctx context.Context, resourceId, resourceType string,
	client *centralizedstorage.CentralizedClient, request *cmi.DeleteLabelRequest) (*cmi.DeleteLabelResponse, error)

// OceanStorageLabelService ocean storage label service
type OceanStorageLabelService struct{}

// CreateLabel create label in ocean storage
func (o *OceanStorageLabelService) CreateLabel(ctx context.Context,
	request *cmi.CreateLabelRequest) (*cmi.CreateLabelResponse, error) {

	param, err := PrepareLabelRequest(ctx, request.GetVolumeId())
	if err != nil {
		log.AddContext(ctx).Errorf("create label failed, volumeId: %s, error: %v", request.GetVolumeId(), err)
		return nil, err
	}

	if param.resourceId == "" {
		log.AddContext(ctx).Errorln("not found resource id, perhaps the volume does not exist, " +
			"so returning failed")
		return nil, errors.New("not found resource id")
	}

	if request.GetNamespace() == "" {
		request.Namespace = constants.DefaultNameSpace
	}

	fun, ok := createLabelFunctions[request.Kind]
	if !ok {
		return nil, errors.New(fmt.Sprintf("illegalArgumentError unsupported resource kind [%s]", request.Kind))
	}

	return fun(ctx, param.resourceId, param.resourceType, param.client, request)
}

// DeleteLabel delete label in ocean storage
func (o *OceanStorageLabelService) DeleteLabel(ctx context.Context,
	request *cmi.DeleteLabelRequest) (*cmi.DeleteLabelResponse, error) {

	param, err := PrepareLabelRequest(ctx, request.GetVolumeId())
	if err != nil {
		log.AddContext(ctx).Errorf("delete label failed, volumeId: %s, error: %v", request.GetVolumeId(), err)
		return nil, err
	}

	if param.resourceId == "" {
		log.AddContext(ctx).Infoln("not found resource id, perhaps the volume does not exist, " +
			"so returning success")
		return &cmi.DeleteLabelResponse{}, nil
	}

	if request.GetNamespace() == "" {
		request.Namespace = constants.DefaultNameSpace
	}

	fun, ok := deleteLabelFunctions[request.Kind]
	if !ok {
		return nil, errors.New(fmt.Sprintf("illegalArgumentError unsupported resource kind [%s]", request.Kind))
	}

	return fun(ctx, param.resourceId, param.resourceType, param.client, request)
}

// createPvLabel create pv label
func createPvLabel(ctx context.Context, resourceId, resourceType string, client *centralizedstorage.CentralizedClient,
	request *cmi.CreateLabelRequest) (*cmi.CreateLabelResponse, error) {

	var data = centralizedstorage.PvLabelRequest{
		ResourceId:   resourceId,
		ResourceType: resourceType,
		PvName:       request.GetLabelName(),
		ClusterName:  request.GetClusterName(),
	}
	_, err := client.CreatePvLabel(ctx, data)
	if err != nil {
		log.AddContext(ctx).Errorf("create pv label failed, volumeId: %s, error: %v", request.GetVolumeId(), err)
		return nil, err
	}
	return &cmi.CreateLabelResponse{}, nil
}

// createPodLabel create pod label
func createPodLabel(ctx context.Context, resourceId, resourceType string, client *centralizedstorage.CentralizedClient,
	request *cmi.CreateLabelRequest) (*cmi.CreateLabelResponse, error) {

	var data = centralizedstorage.PodLabelRequest{
		ResourceId:   resourceId,
		ResourceType: resourceType,
		PodName:      request.GetLabelName(),
		NameSpace:    request.GetNamespace(),
	}
	_, err := client.CreatePodLabel(ctx, data)
	if err != nil {
		log.AddContext(ctx).Errorf("create pod label failed, volumeId: %s, error: %v", request.VolumeId, err)
		return nil, err
	}
	return &cmi.CreateLabelResponse{}, nil
}

// deletePvLabel delete pod label
func deletePvLabel(ctx context.Context, resourceId, resourceType string, client *centralizedstorage.CentralizedClient,
	request *cmi.DeleteLabelRequest) (*cmi.DeleteLabelResponse, error) {

	_, err := client.DeletePvLabel(ctx, resourceId, resourceType)
	if err != nil {
		log.AddContext(ctx).Errorf("delete pv label failed, volumeId: %s, error: %v", request.VolumeId, err)
		return nil, err
	}
	return &cmi.DeleteLabelResponse{}, nil
}

// deletePodLabel delete pod label
func deletePodLabel(ctx context.Context, resourceId, resourceType string, client *centralizedstorage.CentralizedClient,
	request *cmi.DeleteLabelRequest) (*cmi.DeleteLabelResponse, error) {

	var data = centralizedstorage.PodLabelRequest{
		ResourceId:   resourceId,
		ResourceType: resourceType,
		PodName:      request.GetLabelName(),
		NameSpace:    request.GetNamespace(),
	}
	_, err := client.DeletePodLabel(ctx, data)
	if err != nil {
		log.AddContext(ctx).Errorf("delete pod label failed, volumeId: %s, error: %v", request.VolumeId, err)
		return nil, err
	}
	return &cmi.DeleteLabelResponse{}, nil
}
