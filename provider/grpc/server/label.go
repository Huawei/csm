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

// Package server is a package that implement grpc interface
package server

import (
	"context"
	"errors"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/provider/grpc/helper"
	"github.com/huawei/csm/v2/provider/label"
	"github.com/huawei/csm/v2/utils/log"
)

// createLabelValidator verify the parameters when creating labels, e.g. volumeId, clusterName...
var createLabelValidator = helper.NewValidator[label.Validator](validateVolumeId, validateLabelName, validateKind,
	validateClusterName)

// deleteLabelValidator verify the parameters when deleting labels, e.g. volumeId, labelName...
var deleteLabelValidator = helper.NewValidator[label.Validator](validateVolumeId, validateLabelName, validateKind)

// Label implement cmi.LabelServiceServer
type Label struct{}

// CreateLabel create label in storage
func (l *Label) CreateLabel(ctx context.Context, request *cmi.CreateLabelRequest) (*cmi.CreateLabelResponse, error) {
	log.AddContext(ctx).Infof("Start to create label, request: %v", request)

	labelRequest := label.ConvertCreateRequest(request)
	if err := createLabelValidator.Validate(labelRequest); err != nil {
		return nil, err
	}

	service := label.GetLabelService()

	return service.CreateLabel(ctx, request)
}

// DeleteLabel delete label in storage
func (l *Label) DeleteLabel(ctx context.Context, request *cmi.DeleteLabelRequest) (*cmi.DeleteLabelResponse, error) {
	log.AddContext(ctx).Infof("Start to delete label, request: %v", request)

	labelRequest := label.ConvertDeleteRequest(request)
	if err := deleteLabelValidator.Validate(labelRequest); err != nil {
		return nil, err
	}

	service := label.GetLabelService()
	return service.DeleteLabel(ctx, request)
}

// validateLabelName validate if the label name is blank
func validateVolumeId(request label.Validator) error {
	if request.VolumeId == "" {
		return errors.New("illegalArgumentError volume id is blank")
	}
	return nil
}

// validateLabelName validate if the label name is blank
func validateLabelName(request label.Validator) error {
	if request.LabelName == "" {
		return errors.New("illegalArgumentError label name is blank")
	}
	return nil
}

// validateLabelName validate if the label name is blank
func validateKind(request label.Validator) error {
	if request.Kind == "" {
		return errors.New("illegalArgumentError kind is blank")
	}

	if request.Kind != constants.PodKind && request.Kind != constants.PersistentVolumeKind {
		return errors.New("illegalArgumentError unsupported kind")
	}
	return nil
}

// validateLabelName validate if the label name is blank
func validateClusterName(request label.Validator) error {
	if request.Kind == constants.PersistentVolumeKind && request.ClusterName == "" {
		return errors.New("illegalArgumentError cluster name is blank")
	}
	return nil
}
