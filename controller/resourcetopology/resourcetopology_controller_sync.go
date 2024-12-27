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
	"reflect"
	"sort"

	coreV1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiXuanwuV1 "github.com/huawei/csm/v2/client/apis/xuanwu/v1"
	controller "github.com/huawei/csm/v2/config/topology"
	innerTag "github.com/huawei/csm/v2/controller/resource"
	"github.com/huawei/csm/v2/controller/utils"
	"github.com/huawei/csm/v2/controller/utils/cmi"
	"github.com/huawei/csm/v2/controller/utils/consts"
	grpc "github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/utils/log"
)

type provisioner struct {
	provider   string
	capability map[string]bool
}

var (
	cmiProvisioner provisioner
)

func (ctrl *Controller) syncResourceTopology(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology) error {
	log.AddContext(ctx).Infof("start to sync resourceTopology [%s]", resourceTopology.Name)
	defer log.AddContext(ctx).Infof("finished sync resourceTopology [%s]", resourceTopology.Name)

	resourceTopologyNew, err := ctrl.addResourceTopologyFinalizers(ctx,
		resourceTopology, resourceTopologyFinalizerBySelf)
	if err != nil {
		ctrl.eventRecorder.Event(resourceTopology, coreV1.EventTypeWarning,
			syncedFailedReason, failedUpdateResourceTopologyFinalizersMessage)
		return err
	}

	addList, delList := getChangeList(resourceTopologyNew)
	if len(addList) != 0 || len(delList) != 0 {
		err = ctrl.provisionerCheck(ctx, resourceTopologyNew)
		if err != nil {
			return err
		}

		log.AddContext(ctx).Infof("new tags [%v], delete tags [%v]", addList, delList)
		resourceTopologyNew, err = ctrl.handlePendingStatus(ctx, resourceTopologyNew, delList, addList)
		if err != nil {
			return err
		}
		return nil
	}

	if resourceTopologyNew.Status.Status != apiXuanwuV1.ResourceTopologyStatusNormal {
		resourceTopologyNew, err = ctrl.updateResourceTopologyStatusPhase(ctx, resourceTopologyNew,
			apiXuanwuV1.ResourceTopologyStatusNormal)
		if err != nil {
			return err
		}
	}

	// check resources
	return ctrl.checkResourceTopology(ctx, resourceTopologyNew)
}

func (ctrl *Controller) provisionerCheck(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology) error {
	// check if using right provisioner name
	err := ctrl.checkProvisionerName(ctx, resourceTopology)
	if err != nil {
		return err
	}

	// check if provisioner supports labels capability
	err = ctrl.checkProvisionerCapability(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *Controller) checkProvisionerName(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology) error {
	if cmiProvisioner.provider == "" {
		info, err := ctrl.cmiClient.IdentityClient.GetProvisionerInfo(ctx, &grpc.GetProviderInfoRequest{})
		if err != nil {
			return fmt.Errorf("error getting provisioner info: [%v]", err)
		}

		cmiProvisioner.provider = info.Provider
	}

	if resourceTopology.Spec.Provisioner == cmiProvisioner.provider {
		return nil
	}

	return fmt.Errorf("provider not correct, in resourceTopology is [%s], from cmi got: [%s]",
		resourceTopology.Spec.Provisioner, cmiProvisioner.provider)
}

func (ctrl *Controller) checkProvisionerCapability(ctx context.Context) error {
	if cmiProvisioner.capability == nil || len(cmiProvisioner.capability) == 0 {
		cmiProvisioner.capability = make(map[string]bool)
		capabilities, err := ctrl.cmiClient.IdentityClient.GetProviderCapabilities(ctx,
			&grpc.GetProviderCapabilitiesRequest{})
		if err != nil {
			return errors.New("error getting provider capabilities")
		}

		for _, capability := range capabilities.GetCapabilities() {
			cmiProvisioner.capability[grpc.ProviderCapability_Type_name[int32(capability.Type)]] = true
		}
	}

	if cmiProvisioner.capability[grpc.ProviderCapability_Type_name[int32(
		grpc.ProviderCapability_ProviderCapability_Label_Service)]] {
		return nil
	}

	return errors.New("cmi unsupported label capability")
}

func (ctrl *Controller) handlePendingStatus(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology,
	delList []apiXuanwuV1.Tag, addList []apiXuanwuV1.Tag) (*apiXuanwuV1.ResourceTopology, error) {
	var err error
	resourceTopology, err = ctrl.updateResourceTopologyStatusPhase(ctx, resourceTopology,
		apiXuanwuV1.ResourceTopologyStatusPending)
	if err != nil {
		return nil, err
	}

	if len(delList) != 0 {
		resourceTopology, err = ctrl.handleDeleteTags(ctx, resourceTopology, delList)
		if err != nil {
			return nil, err
		}
	}
	if len(addList) != 0 {
		resourceTopology, err = ctrl.handleAddTags(ctx, resourceTopology, addList)
		if err != nil {
			return nil, err
		}
	}
	return resourceTopology, nil
}

func (ctrl *Controller) handleAddTags(ctx context.Context, resourceTopology *apiXuanwuV1.ResourceTopology,
	addList []apiXuanwuV1.Tag) (*apiXuanwuV1.ResourceTopology, error) {
	log.AddContext(ctx).Infof("start to add tags [%v] to resourceTopology [%s]", addList, resourceTopology.Name)
	defer log.AddContext(ctx).Infof("finished add tags [%v] to resourceTopology [%s]",
		addList, resourceTopology.Name)

	var err error
	for _, tag := range addList {
		log.AddContext(ctx).Infof("trying to add tag [%v]", tag)
		err = ctrl.CmiCreateLabel(ctx, getCmiParams(resourceTopology, tag))
		if err != nil {
			return nil, err
		}
	}

	statusTags := append(resourceTopology.Status.Tags, addList...)
	resourceTopology, err = ctrl.updateResourceTopologyStatusTagsWithRetry(ctx, resourceTopology, statusTags)
	if err != nil {
		return nil, err
	}

	return resourceTopology, nil
}

func (ctrl *Controller) rollBack(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology, tag apiXuanwuV1.Tag) {
	log.AddContext(ctx).Infof("rolling back resource topology tag [%v]", tag)
	rollBackErr := ctrl.CmiDeleteLabel(ctx, getCmiParams(resourceTopology, tag))
	if rollBackErr != nil {
		log.AddContext(ctx).Errorf("roll back label on storage err: [%v]", rollBackErr)
	}
}

func (ctrl *Controller) addTagToSlice(ctx context.Context, tags []apiXuanwuV1.Tag,
	tag apiXuanwuV1.Tag, params *cmi.Params) ([]apiXuanwuV1.Tag, error) {
	if contains := utils.Contains(controller.GetSupportResources(), tag.Kind); !contains {
		log.Infof("tag kind [%s] is unsupported or already added, skipping", tag.Kind)
		return tags, nil
	}

	inner, err := innerTag.NewInnerTag(tag.TypeMeta)
	if err != nil {
		log.AddContext(ctx).Errorf("new inner tag failed: [%v]", err)
		return tags, err
	}
	inner.InitFromResourceInfo(tag.ResourceInfo)

	// try to add owner of the resource
	if inner.HasOwner() {
		ownerInfo, tags, err := ctrl.addOwnerTag(ctx, tags, inner, params)
		if err != nil {
			return tags, err
		}
		if utils.Contains(controller.GetSupportResources(), ownerInfo.Kind) {
			tag.Owner = ownerInfo
		}
	}

	// check resource enable added
	if err := addResourceCheck(inner); err != nil {
		return tags, err
	}

	err = ctrl.CmiCreateLabel(ctx, params)
	if err != nil {
		return nil, err
	}

	tags = append(tags, tag)
	return tags, nil
}

func (ctrl *Controller) addOwnerTag(ctx context.Context, tags []apiXuanwuV1.Tag,
	inner innerTag.InnerTag, params *cmi.Params) (apiXuanwuV1.ResourceInfo, []apiXuanwuV1.Tag, error) {
	ownerInfo, err := getOwnerInfo(inner)
	if err != nil {
		return apiXuanwuV1.ResourceInfo{}, tags, err
	}
	// if the owner kind is not supported, kind in owner info will be empty
	if ownerInfo.Kind == "" {
		return apiXuanwuV1.ResourceInfo{}, tags, nil
	}

	tags, err = ctrl.addTagToSlice(ctx, tags, apiXuanwuV1.Tag{ResourceInfo: ownerInfo}, params)
	if err != nil {
		return apiXuanwuV1.ResourceInfo{}, tags, err
	}
	return ownerInfo, tags, nil
}

func (ctrl *Controller) handleDeleteTags(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology,
	delList []apiXuanwuV1.Tag) (*apiXuanwuV1.ResourceTopology, error) {
	log.AddContext(ctx).Infof("start to remove tags [%v] from resourceTopology [%s]",
		delList, resourceTopology.Name)
	defer log.AddContext(ctx).Infof("finished remove tags [%v] from resourceTopology [%s]",
		delList, resourceTopology.Name)

	statusTags := resourceTopology.Status.Tags

	var err error
	for _, tag := range delList {
		log.AddContext(ctx).Infof("trying to delete tag [%v]", tag)
		err = ctrl.CmiDeleteLabel(ctx, getCmiParams(resourceTopology, tag))
		if err != nil {
			return nil, err
		}
		statusTags = deleteTag(statusTags, tag)
	}

	resourceTopology, err = ctrl.updateResourceTopologyStatusTagsWithRetry(ctx, resourceTopology, statusTags)
	if err != nil {
		return nil, err
	}

	return resourceTopology, nil
}

func (ctrl *Controller) deleteTagFromSlice(ctx context.Context, tags []apiXuanwuV1.Tag,
	tag apiXuanwuV1.Tag, params *cmi.Params) ([]apiXuanwuV1.Tag, error) {
	inner, err := innerTag.NewInnerTag(tag.TypeMeta)
	if err != nil {
		log.AddContext(ctx).Errorf("new inner tag failed: [%v]", err)
		return tags, err
	}

	inner.InitFromResourceInfo(tag.ResourceInfo)
	if inner.HasOwner() {
		tags, err = ctrl.deleteOwnerTag(ctx, tags, inner, params)
		if err != nil {
			return tags, err
		}
	}

	if inner.Exists() && !inner.IsDeleting() {
		return tags, fmt.Errorf("resource of tag [%v] still exist but not in deleting status", inner)
	}

	err = utils.RetryFunc(func() (bool, error) {
		if inner.Exists() && inner.IsDeleting() {
			return false, fmt.Errorf("resource [%s] still deleting, check again later", tag)
		}
		return true, nil
	}, consts.RetryTimes, consts.RetryDurationInit, consts.RetryDurationMax)
	if err != nil {
		return tags, fmt.Errorf("delete tag [%v] failed: [%v]", tag, err)
	}

	err = ctrl.CmiDeleteLabel(ctx, params)
	if err != nil {
		return tags, err
	}

	tags = deleteTag(tags, tag)

	return tags, nil
}

func (ctrl *Controller) deleteOwnerTag(ctx context.Context, tags []apiXuanwuV1.Tag,
	inner innerTag.InnerTag, params *cmi.Params) ([]apiXuanwuV1.Tag, error) {
	ownerInfo, err := getOwnerInfo(inner)
	if err != nil {
		return tags, err
	}
	// if the owner kind is not supported, kind in owner info will be empty
	if ownerInfo.Kind == "" {
		return tags, nil
	}

	tags, err = ctrl.deleteTagFromSlice(ctx, tags, apiXuanwuV1.Tag{ResourceInfo: ownerInfo}, params)
	if err != nil {
		return tags, err
	}
	return tags, nil
}

func (ctrl *Controller) checkResourceTopology(ctx context.Context,
	resourceTopology *apiXuanwuV1.ResourceTopology) error {
	log.AddContext(ctx).Infof("start to check resources of resourceTopology [%s]", resourceTopology.Name)
	defer log.AddContext(ctx).Infof("finished check resources of resourceTopology [%s]", resourceTopology.Name)

	// check whether pv exists
	pvName := getPvNameByResourceTopologyName(resourceTopology.Name)
	_, err := ctrl.volumeInformer.Lister().Get(pvName)
	if apiErrors.IsNotFound(err) {
		log.Errorf("pv [%s] is not exists, start to delete resourceTopology [%s]", pvName, resourceTopology.Name)
		return ctrl.xuanwuClient.XuanwuV1().ResourceTopologies().
			Delete(ctx, resourceTopology.Name, metaV1.DeleteOptions{})
	}

	if err != nil {
		return err
	}

	// check whether pods exist
	tags := resourceTopology.Spec.Tags
	for _, tag := range resourceTopology.Status.Tags {
		if tag.Kind == consts.PersistentVolume {
			continue
		}

		exist, err := ctrl.isPodExisted(ctx, tag.ResourceInfo)
		if err != nil {
			return err
		}

		if !exist {
			tags = deleteTag(tags, tag)
		}
	}

	if reflect.DeepEqual(resourceTopology.Spec.Tags, tags) {
		return nil
	}

	// update resourceTopology spec tags
	resourceTopology.Spec.Tags = tags
	_, err = ctrl.xuanwuClient.XuanwuV1().ResourceTopologies().Update(ctx, resourceTopology, metaV1.UpdateOptions{})

	return err
}

func (ctrl *Controller) isPodExisted(ctx context.Context, resourceInfo apiXuanwuV1.ResourceInfo) (bool, error) {
	if resourceInfo.Kind != consts.Pod {
		return false, fmt.Errorf("unsupported resource tag type [%s]", resourceInfo.Kind)
	}

	_, err := ctrl.podInformer.Lister().Pods(resourceInfo.Namespace).Get(resourceInfo.Name)
	if apiErrors.IsNotFound(err) {
		log.AddContext(ctx).Errorf("pod [%s] is not existed, try to delete tag", resourceInfo.Name)
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func getOwnerInfo(inner innerTag.InnerTag) (apiXuanwuV1.ResourceInfo, error) {
	ownerInnerTag, err := inner.GetOwner()
	if err != nil {
		return apiXuanwuV1.ResourceInfo{}, err
	}
	if ownerInnerTag == nil {
		return apiXuanwuV1.ResourceInfo{}, nil
	}
	ownerInfo := ownerInnerTag.ToResourceInfo()
	return ownerInfo, nil
}

func deleteTag(tags []apiXuanwuV1.Tag, target apiXuanwuV1.Tag) []apiXuanwuV1.Tag {
	idx := 0
	for i, tag := range tags {
		if target.ResourceInfo == tag.ResourceInfo {
			idx = i
			break
		}
	}
	tags = append(tags[:idx], tags[idx+1:]...)
	return tags
}

func addResourceCheck(inner innerTag.InnerTag) error {
	if !inner.Exists() {
		return fmt.Errorf("resource of tag [%v] does not exist", inner)
	}

	if !inner.Ready() {
		return fmt.Errorf("resource of tag [%v] not ready", inner)
	}

	if inner.IsDeleting() {
		return fmt.Errorf("resource of tag [%v] is deleting", inner)
	}
	return nil
}

func getChangeList(topology *apiXuanwuV1.ResourceTopology) ([]apiXuanwuV1.Tag, []apiXuanwuV1.Tag) {
	spec := getPodAndPvTags(topology.Spec.Tags)
	status := getPodAndPvTags(topology.Status.Tags)
	return getAddTagsList(status, spec), getDeleteTagsList(spec, status)
}

func getDeleteTagsList(spec []apiXuanwuV1.Tag, status []apiXuanwuV1.Tag) []apiXuanwuV1.Tag {
	return getChangedTags(spec, status)
}

func getAddTagsList(status []apiXuanwuV1.Tag, spec []apiXuanwuV1.Tag) []apiXuanwuV1.Tag {
	add := getChangedTags(status, spec)

	// if need to add PersistentVolume label to storage, must add first
	sort.Slice(add, func(i, j int) bool {
		if add[i].Kind == consts.PersistentVolume {
			return true
		}
		return false
	})
	return add
}

func getChangedTags(origin []apiXuanwuV1.Tag, newList []apiXuanwuV1.Tag) []apiXuanwuV1.Tag {
	set := make(map[apiXuanwuV1.ResourceInfo]struct{})
	for _, tag := range origin {
		set[tag.ResourceInfo] = struct{}{}
	}

	var add []apiXuanwuV1.Tag
	for _, tag := range newList {
		if _, ok := set[tag.ResourceInfo]; !ok {
			add = append(add, tag)
		}
	}
	return add
}

func getPodAndPvTags(tags []apiXuanwuV1.Tag) []apiXuanwuV1.Tag {
	var result []apiXuanwuV1.Tag
	for _, tag := range tags {
		if tag.Kind == consts.Pod || tag.Kind == consts.PersistentVolume {
			result = append(result, apiXuanwuV1.Tag{ResourceInfo: tag.ResourceInfo})
		}
	}
	return result
}
