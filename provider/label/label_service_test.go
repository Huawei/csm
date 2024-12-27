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
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
)

func Test_CreatePvLabel_Success(t *testing.T) {
	// arrange
	resourceId, resourceType := "", ""
	client := &centralizedstorage.CentralizedClient{}
	request := &cmi.CreateLabelRequest{}

	// mock
	methodFunc := gomonkey.ApplyMethodFunc(client, "CreatePvLabel", func(context.Context,
		centralizedstorage.PvLabelRequest) (map[string]interface{}, error) {
		return map[string]interface{}{}, nil
	})
	defer methodFunc.Reset()

	// action
	_, err := createPvLabel(context.Background(), resourceId, resourceType, client, request)

	// assert
	if err != nil {
		t.Errorf("createPvLabel() error = %v", err)
	}
}

func Test_CreatePodLabel_Success(t *testing.T) {
	// arrange
	resourceId, resourceType := "", ""
	client := &centralizedstorage.CentralizedClient{}
	request := &cmi.CreateLabelRequest{}

	// mock
	methodFunc := gomonkey.ApplyMethodFunc(client, "CreatePodLabel", func(context.Context,
		centralizedstorage.PodLabelRequest) (map[string]interface{}, error) {
		return map[string]interface{}{}, nil
	})
	defer methodFunc.Reset()

	// action
	_, err := createPodLabel(context.Background(), resourceId, resourceType, client, request)

	// assert
	if err != nil {
		t.Errorf("createPodLabel() error = %v", err)
	}
}

func Test_DeletePvLabel_Success(t *testing.T) {
	// arrange
	resourceId, resourceType := "", ""
	client := &centralizedstorage.CentralizedClient{}
	request := &cmi.DeleteLabelRequest{}

	// mock
	methodFunc := gomonkey.ApplyMethodFunc(client, "DeletePvLabel", func(context.Context,
		string, string) (map[string]interface{}, error) {
		return map[string]interface{}{}, nil
	})
	defer methodFunc.Reset()

	// action
	_, err := deletePvLabel(context.Background(), resourceId, resourceType, client, request)

	// assert
	if err != nil {
		t.Errorf("deletePvLabel() error = %v", err)
	}
}

func Test_DeletePodLabel_Success(t *testing.T) {
	// arrange
	resourceId, resourceType := "", ""
	client := &centralizedstorage.CentralizedClient{}
	request := &cmi.DeleteLabelRequest{}

	// mock
	methodFunc := gomonkey.ApplyMethodFunc(client, "DeletePodLabel", func(context.Context,
		centralizedstorage.PodLabelRequest) (map[string]interface{}, error) {
		return map[string]interface{}{}, nil
	})
	defer methodFunc.Reset()

	// action
	_, err := deletePodLabel(context.Background(), resourceId, resourceType, client, request)

	// assert
	if err != nil {
		t.Errorf("deletePodLabel() error = %v", err)
	}
}

func Test_OceanStorageLabelService_CreateLabel_Success(t *testing.T) {
	// arrange
	request := &cmi.CreateLabelRequest{Kind: constants.PodKind}
	service := &OceanStorageLabelService{}

	// mock
	methodFunc := gomonkey.
		ApplyFunc(PrepareLabelRequest, func(context.Context, string) (OceanStorageLabelRequest,
			error) {
			return OceanStorageLabelRequest{resourceId: "fakeResourceId"}, nil
		}).
		ApplyFunc(createPodLabel, func(context.Context, string, string, *centralizedstorage.CentralizedClient,
			*cmi.CreateLabelRequest) (*cmi.CreateLabelResponse, error) {
			return &cmi.CreateLabelResponse{}, nil
		})
	defer methodFunc.Reset()

	// action
	_, err := service.CreateLabel(context.Background(), request)

	// assert
	if err != nil {
		t.Errorf("CreateLabel() error = %v", err)
	}
}

func Test_OceanStorageLabelService_DeleteLabel_Success(t *testing.T) {
	// arrange
	request := &cmi.DeleteLabelRequest{Kind: constants.PodKind}
	service := &OceanStorageLabelService{}

	// mock
	methodFunc := gomonkey.
		ApplyFunc(PrepareLabelRequest, func(context.Context, string) (OceanStorageLabelRequest,
			error) {
			return OceanStorageLabelRequest{}, nil
		}).
		ApplyFunc(deletePodLabel, func(context.Context, string, string, *centralizedstorage.CentralizedClient,
			*cmi.DeleteLabelRequest) (*cmi.DeleteLabelResponse, error) {
			return &cmi.DeleteLabelResponse{}, nil
		})
	defer methodFunc.Reset()

	// action
	_, err := service.DeleteLabel(context.Background(), request)

	// assert
	if err != nil {
		t.Errorf("CreateLabel() error = %v", err)
	}
}
