/*
 Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

// Package resourcetopology defines to reconcile action of resources topologies
package resourcetopology

import (
	"context"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiXuanwuV1 "github.com/huawei/csm/v2/client/apis/xuanwu/v1"
	"github.com/huawei/csm/v2/controller/utils/cmi"
	grpc "github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/utils/log"
)

// UpdateResourceTopologiesStatus updates resource topologies status
func (ctrl *Controller) UpdateResourceTopologiesStatus(ctx context.Context,
	resourceTopologyCopy *apiXuanwuV1.ResourceTopology) (*apiXuanwuV1.ResourceTopology, error) {
	resourceTopology, err := ctrl.xuanwuClient.
		XuanwuV1().
		ResourceTopologies().
		UpdateStatus(ctx, resourceTopologyCopy, metaV1.UpdateOptions{})
	return resourceTopology, err
}

// CmiCreateLabel create label by cmi grpc connection
func (ctrl *Controller) CmiCreateLabel(ctx context.Context, params *cmi.Params) error {
	request := &grpc.CreateLabelRequest{
		VolumeId:  params.VolumeId(),
		LabelName: params.LabelName(),
		Kind:      params.Kind(),
	}

	if params.ClusterName() != "" {
		request.ClusterName = params.ClusterName()
	}

	if params.Namespace() != "" {
		request.Namespace = params.Namespace()
	}
	_, err := ctrl.cmiClient.LabelClient.CreateLabel(ctx, request)
	if err != nil {
		log.AddContext(ctx).Errorf("create label [%v] on storage failed: [%v]", params, err)
		return err
	}

	return err
}

// CmiDeleteLabel delete label by cmi grpc connection
func (ctrl *Controller) CmiDeleteLabel(ctx context.Context, params *cmi.Params) error {
	request := &grpc.DeleteLabelRequest{
		VolumeId:  params.VolumeId(),
		LabelName: params.LabelName(),
		Kind:      params.Kind(),
	}

	if params.Namespace() != "" {
		request.Namespace = params.Namespace()
	}

	_, err := ctrl.cmiClient.LabelClient.DeleteLabel(ctx, request)
	if err != nil {
		log.AddContext(ctx).Errorf("delete label [%v] on storage failed: [%v]", params, err)
		return err
	}

	return err
}
