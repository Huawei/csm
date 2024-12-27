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
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	coreV1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiXuanwuV1 "github.com/huawei/csm/v2/client/apis/xuanwu/v1"
	"github.com/huawei/csm/v2/controller/utils"
	"github.com/huawei/csm/v2/controller/utils/cmi"
	"github.com/huawei/csm/v2/utils/log"
)

const (
	updateReason       = "Update"
	updateFailedReason = "UpdateFailed"
	syncedFailedReason = "Synced"

	failedUpdateResourceTopologyStatusPhaseMessage  = "Failed to update ResourceTopology status into"
	successUpdateResourceTopologyStatusPhaseMessage = "Success to update ResourceTopology status into"
	failedUpdateResourceTopologyTagsFieldMessage    = "Failed to update ResourceTopology tags field to"
	successUpdateResourceTopologyTagsFieldMessage   = "Success to update ResourceTopology tags field to"
	failedUpdateResourceTopologyFinalizersMessage   = "Failed to update ResourceTopology finalizers"

	resourceTopologyFinalizerBySelf = "resourcetopology.xuanwu.huawei.io/resourcetopology-protection"

	rtPrefix          = "rt-"
	updateRetryTimes  = 10
	updateRetryPeriod = 100 * time.Millisecond

	resourceRequeueInterval = 10 * time.Second
)

func (ctrl *Controller) updateResourceTopologyStatusPhase(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology,
	statusPhase apiXuanwuV1.ResourceTopologyStatusPhase) (*apiXuanwuV1.ResourceTopology, error) {
	log.AddContext(ctx).Infof("update resourceTopology [%s] into status [%s]", resourceTopology.Name, statusPhase)
	statusCopy := resourceTopology.Status.DeepCopy()
	statusCopy.Status = statusPhase

	resourceTopologyNew, err := ctrl.updateResourceTopologyStatusStruct(ctx, resourceTopology, *statusCopy)
	if err != nil {
		ctrl.eventRecorder.Event(resourceTopology, coreV1.EventTypeWarning, updateFailedReason,
			fmt.Sprintf("%s %s", failedUpdateResourceTopologyStatusPhaseMessage, statusPhase))
		return nil, err
	}

	*resourceTopology = *resourceTopologyNew
	ctrl.eventRecorder.Event(resourceTopology, coreV1.EventTypeNormal, updateReason,
		fmt.Sprintf("%s %s", successUpdateResourceTopologyStatusPhaseMessage, statusPhase))
	return resourceTopology, nil
}

func (ctrl *Controller) updateResourceTopologyStatusTagsWithRetry(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology, tags []apiXuanwuV1.Tag) (*apiXuanwuV1.ResourceTopology, error) {
	log.AddContext(ctx).Infof("update resourceTopology [%s] tags field in status", resourceTopology.Name)
	rtCopy := resourceTopology.DeepCopy()
	var err error
	for attempt := 0; attempt < updateRetryTimes; attempt++ {
		rtCopy.Status.Tags = tags
		rtCopy, err = ctrl.xuanwuClient.XuanwuV1().ResourceTopologies().
			UpdateStatus(ctx, rtCopy, metaV1.UpdateOptions{})
		if err == nil {
			ctrl.eventRecorder.Event(rtCopy, coreV1.EventTypeNormal, updateReason,
				fmt.Sprintf("%s %v", successUpdateResourceTopologyTagsFieldMessage, tags))
			return rtCopy, nil
		}

		if !apiErrors.IsConflict(err) {
			ctrl.eventRecorder.Event(resourceTopology, coreV1.EventTypeWarning, updateFailedReason,
				fmt.Sprintf("%s %v", failedUpdateResourceTopologyTagsFieldMessage, tags))
			return nil, err
		}

		log.AddContext(ctx).Infof("conflict when trying to update resourceTopology [%s], need to try again",
			resourceTopology.Name)
		time.Sleep(updateRetryPeriod)
		rtCopy, err = ctrl.xuanwuClient.XuanwuV1().ResourceTopologies().
			Get(ctx, resourceTopology.Name, metaV1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("too many conflicts when trying to update resourceTopology [%s]",
		resourceTopology.Name)
}

func (ctrl *Controller) updateResourceTopologyStatusStruct(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology,
	status apiXuanwuV1.ResourceTopologyStatus) (*apiXuanwuV1.ResourceTopology, error) {
	resourceTopologyCopy := resourceTopology.DeepCopy()
	resourceTopologyCopy.Status = status

	resourceTopologyNew, err := ctrl.UpdateResourceTopologiesStatus(ctx, resourceTopologyCopy)
	if err != nil {
		errMsg := fmt.Sprintf("update resourceTopology [%s] status struct failed: [%v]", resourceTopology.Name, err)
		log.AddContext(ctx).Errorln(errMsg)
		return nil, errors.New(errMsg)
	}

	*resourceTopology = *resourceTopologyNew
	log.AddContext(ctx).Infof("update resourceTopology [%s] status struct succeed, the new status struct is [%v]",
		resourceTopology.Name, resourceTopology.Status)
	return resourceTopology, nil
}

func (ctrl *Controller) addResourceTopologyFinalizers(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology, target string) (*apiXuanwuV1.ResourceTopology, error) {
	finalizers := resourceTopology.Finalizers
	if utils.Contains(finalizers, target) {
		return resourceTopology, nil
	}

	resourceTopologyCopy := resourceTopology.DeepCopy()
	resourceTopologyCopy.Finalizers = append(finalizers, target)

	resourceTopologyNew, err := ctrl.xuanwuClient.XuanwuV1().ResourceTopologies().
		Update(ctx, resourceTopologyCopy, metaV1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("add resourceTopology [%s] finalizers failed, errors is [%v]",
			resourceTopology.Name, err)
	}

	log.AddContext(ctx).Infof("add resourceTopology [%s] finalizers succeed, the new finalizers is [%v]",
		resourceTopology.Name, resourceTopologyNew.Finalizers)

	return resourceTopologyNew, nil
}

func (ctrl *Controller) deleteResourceTopologyFinalizers(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology, target string) (*apiXuanwuV1.ResourceTopology, error) {
	finalizers := resourceTopology.Finalizers
	if !utils.Contains(finalizers, target) {
		return resourceTopology, nil
	}

	resourceTopologyCopy := resourceTopology.DeepCopy()
	resourceTopologyCopy.Finalizers = utils.DeleteElementFromSlice(finalizers, target)

	resourceTopologyNew, err := ctrl.xuanwuClient.XuanwuV1().ResourceTopologies().
		Update(ctx, resourceTopologyCopy, metaV1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("delete resourceTopology [%s] finalizers failed, errors is [%v]",
			resourceTopology.Name, err)
	}

	log.AddContext(ctx).Infof("delete resourceTopology [%s] finalizers succeed, the new finalizers is [%v]",
		resourceTopology.Name, resourceTopologyNew.Finalizers)

	return resourceTopologyNew, nil
}

func getCmiParams(resourceTopology *apiXuanwuV1.ResourceTopology, tag apiXuanwuV1.Tag) *cmi.Params {
	params := &cmi.Params{}
	return params.SetVolumeId(resourceTopology.Spec.VolumeHandle).
		SetKind(tag.Kind).
		SetNamespace(tag.Namespace).
		SetLabelName(tag.Name).
		SetClusterName(os.Getenv("CLUSTER_NAME"))
}

func checkResourceTopologyName(rtName string) bool {
	return strings.HasPrefix(rtName, rtPrefix)
}

func getResourceTopologyName(pvName string) string {
	return rtPrefix + pvName
}

func getPvNameByResourceTopologyName(rtName string) string {
	return strings.TrimPrefix(rtName, rtPrefix)
}
