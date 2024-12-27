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

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/backend"
	"github.com/huawei/csm/v2/provider/collect"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/provider/utils"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
	"github.com/huawei/csm/v2/utils/log"
)

// Validator label validator contains all fields to be verified
type Validator struct {
	VolumeId    string
	LabelName   string
	Kind        string
	Namespace   string
	ClusterName string
}

// OceanStorageLabelRequest operation ocean storage label
type OceanStorageLabelRequest struct {
	resourceId   string
	resourceType string
	client       *centralizedstorage.CentralizedClient
}

// ConvertCreateRequest convert CreateLabelRequest to LabelValidator
func ConvertCreateRequest(request *cmi.CreateLabelRequest) Validator {
	return Validator{
		VolumeId:    request.GetVolumeId(),
		LabelName:   request.GetLabelName(),
		Kind:        request.GetKind(),
		Namespace:   request.GetNamespace(),
		ClusterName: request.GetClusterName(),
	}
}

// ConvertDeleteRequest convert DeleteLabelRequest to LabelValidator
func ConvertDeleteRequest(request *cmi.DeleteLabelRequest) Validator {
	return Validator{
		VolumeId:  request.GetVolumeId(),
		LabelName: request.GetLabelName(),
		Kind:      request.GetKind(),
		Namespace: request.GetNamespace(),
	}
}

// PrepareLabelRequest get client and resource object information
func PrepareLabelRequest(ctx context.Context, volumeId string) (OceanStorageLabelRequest, error) {
	backendName, volumeName := utils.SplitVolumeId(volumeId)
	clientInfo, err := collect.GetClient(ctx, backendName, backend.GetClientByBackendName)
	if err != nil {
		log.AddContext(ctx).Errorf("delete label get client failed, error: %v", err)
		return OceanStorageLabelRequest{}, err
	}

	client, ok := clientInfo.Client.(*centralizedstorage.CentralizedClient)
	if !ok {
		return OceanStorageLabelRequest{}, errors.New("convert storage client failed")
	}

	resourceType := getResourceType(clientInfo.VolumeType)
	resourceId, err := getResourceId(ctx, volumeName, clientInfo.VolumeType, client)
	if err != nil {
		log.AddContext(ctx).Errorf("delete label get resource id failed, error: %v", err)
		return OceanStorageLabelRequest{}, err
	}

	return OceanStorageLabelRequest{resourceId: resourceId, resourceType: resourceType, client: client}, nil
}

func getResourceId(ctx context.Context, volumeName, volumeType string,
	client *centralizedstorage.CentralizedClient) (string, error) {

	if volumeType == constants.NasVolume {
		return client.GetFileSystemIdByName(ctx, volumeName)
	}

	return client.GetLunIdByName(ctx, volumeName)
}

func getResourceType(volumeType string) string {
	if volumeType == constants.NasVolume {
		return constants.ResourceTypeFilesystem
	}
	return constants.ResourceTypeLun
}
