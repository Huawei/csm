/*
 Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.

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

	coreV1 "k8s.io/api/core/v1"

	apiXuanwuV1 "github.com/huawei/csm/v2/client/apis/xuanwu/v1"
	"github.com/huawei/csm/v2/controller/utils/consts"
	"github.com/huawei/csm/v2/utils/log"
)

func (ctrl *Controller) deleteResourceTopology(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology) error {
	log.AddContext(ctx).Infof("start deleteResourceTopology [%s]", resourceTopology.Name)
	defer log.AddContext(ctx).Infof("end deleteResourceTopology [%s]", resourceTopology.Name)

	var err error

	// update the status of resourceTopology to Deleting
	if resourceTopology.Status.Status != apiXuanwuV1.ResourceTopologyStatusDeleting {
		resourceTopology, err = ctrl.updateResourceTopologyStatusPhase(ctx,
			resourceTopology, apiXuanwuV1.ResourceTopologyStatusDeleting)
		if err != nil {
			return err
		}
	}

	// delete resources in status on storage
	for _, tag := range resourceTopology.Status.Tags {
		err = ctrl.CmiDeleteLabel(ctx, getCmiParams(resourceTopology, tag))
		if err != nil {
			return err
		}
	}

	// reload resources in spec on cluster
	for _, tag := range resourceTopology.Spec.Tags {
		switch tag.Kind {
		case consts.Pod:
			ctrl.podQueue.Add(tag.Namespace + "/" + tag.Name)
			break
		case consts.PersistentVolume:
			ctrl.volumeQueue.Add(tag.Name)
			break
		default:
			log.AddContext(ctx).Errorf("unsupported tag type: [%s]", tag.Kind)
		}
	}

	// remove self finalizer
	_, err = ctrl.deleteResourceTopologyFinalizers(ctx, resourceTopology, resourceTopologyFinalizerBySelf)
	if err != nil {
		ctrl.eventRecorder.Event(resourceTopology, coreV1.EventTypeWarning,
			syncedFailedReason, failedUpdateResourceTopologyFinalizersMessage)
		return err
	}

	return nil
}
