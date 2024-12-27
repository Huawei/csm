/*
 Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

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
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiXuanwuV1 "github.com/huawei/csm/v2/client/apis/xuanwu/v1"
	"github.com/huawei/csm/v2/config/cmi"
	"github.com/huawei/csm/v2/controller/utils"
	"github.com/huawei/csm/v2/controller/utils/consts"
	"github.com/huawei/csm/v2/utils/log"
)

func (ctrl *Controller) syncPersistentVolume(ctx context.Context, pv *coreV1.PersistentVolume) error {
	log.AddContext(ctx).Infof("[pv-controller] start to sync pv [%s]", pv.Name)
	defer log.AddContext(ctx).Infof("[pv-controller] finished sync pv [%s]", pv.Name)

	if pv.Status.Phase == coreV1.VolumeFailed {
		log.AddContext(ctx).Debugf("[pv-controller] pv [%s] is in failed status, skip to next", pv.Name)
		return nil
	}

	rtName := getResourceTopologyName(pv.Name)
	labelSelector := &metaV1.LabelSelector{
		MatchLabels: map[string]string{consts.VolumeHandleKeyLabel: utils.EncryptMD5(pv.Spec.CSI.VolumeHandle)},
	}
	selector, err := metaV1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return err
	}

	rtList, err := ctrl.topologyInformer.Lister().List(selector)
	if err != nil {
		return err
	}

	if len(rtList) > 0 {
		if rtList[0].Name != rtName {
			log.AddContext(ctx).Warningf("[pv-controller] current resource [%s] has bound to another rt [%s]",
				pv.Spec.CSI.VolumeHandle, rtList[0].Name)
			return nil
		}

		if rtList[0].DeletionTimestamp != nil {
			log.Warningf("[pv-controller] rt [%s] is deleting, "+
				"wait the deletion finished to recreated", rtName)
			ctrl.volumeQueue.AddAfter(pv.Name, resourceRequeueInterval)
			return nil
		}

		return nil
	}

	return ctrl.createResourceTopology(ctx, pv, rtName)
}

func (ctrl *Controller) removeResourceTopology(ctx context.Context, pvName string) error {
	rtName := getResourceTopologyName(pvName)
	log.AddContext(ctx).Infof("[pv-controller] start to delete rt [%s]", rtName)
	defer log.AddContext(ctx).Infof("[pv-controller] finished delete rt [%s]", rtName)

	err := ctrl.xuanwuClient.XuanwuV1().ResourceTopologies().Delete(ctx, rtName, metaV1.DeleteOptions{})
	if err != nil && !apiErrors.IsNotFound(err) {
		return err
	}

	log.AddContext(ctx).Infof("[pv-controller] rt [%s] deleted by pv [%s] success", rtName, pvName)
	return nil
}

func (ctrl *Controller) createResourceTopology(ctx context.Context,
	pv *coreV1.PersistentVolume, rtName string) error {
	log.AddContext(ctx).Debugf("[pv-controller] start to create rt [%s]", rtName)
	defer log.AddContext(ctx).Debugf("[pv-controller] finished create rt [%s]", rtName)

	rtLabels := make(map[string]string)
	rtLabels[consts.VolumeHandleKeyLabel] = utils.EncryptMD5(pv.Spec.CSI.VolumeHandle)

	topologySpec := apiXuanwuV1.ResourceTopologySpec{
		Provisioner:  cmi.GetProviderName(),
		VolumeHandle: pv.Spec.CSI.VolumeHandle,
		Tags: []apiXuanwuV1.Tag{
			{
				ResourceInfo: apiXuanwuV1.ResourceInfo{
					TypeMeta: metaV1.TypeMeta{Kind: consts.PersistentVolume, APIVersion: consts.KubernetesV1},
					Name:     pv.Name,
				},
			},
		},
	}

	rt := &apiXuanwuV1.ResourceTopology{
		TypeMeta:   metaV1.TypeMeta{Kind: consts.TopologyKind, APIVersion: consts.XuanwuV1},
		ObjectMeta: metaV1.ObjectMeta{Name: rtName, Labels: rtLabels},
		Spec:       topologySpec,
	}

	_, err := ctrl.xuanwuClient.XuanwuV1().ResourceTopologies().Create(ctx, rt, metaV1.CreateOptions{})
	if err != nil {
		return err
	}

	log.AddContext(ctx).Infof("[pv-controller] rt [%s] created by pv [%s] success", rtName, pv.Name)
	return nil
}
