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
	"fmt"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiXuanwuV1 "github.com/huawei/csm/v2/client/apis/xuanwu/v1"
	"github.com/huawei/csm/v2/controller/utils/consts"
	"github.com/huawei/csm/v2/utils/log"
)

func (ctrl *Controller) syncPod(ctx context.Context, pod *coreV1.Pod) error {
	log.AddContext(ctx).Infof("[pod-controller] start to sync pod [%s/%s] relate tags",
		pod.Namespace, pod.Name)
	defer log.AddContext(ctx).Infof("[pod-controller] finished sync pod [%s/%s] relate tags",
		pod.Namespace, pod.Name)

	err := ctrl.podStore.Add(pod)
	if err != nil {
		return nil
	}

	pvcNameList, retryErr := filterAvailablePvcInPod(pod)

	err = ctrl.syncPodTagByPvcList(ctx, pvcNameList, pod.Name, pod.Namespace)
	if err != nil {
		return err
	}

	if retryErr != nil {
		log.AddContext(ctx).Warningln(retryErr.Error())
		ctrl.podQueue.AddAfter(getPodKey(pod.Namespace, pod.Name), resourceRequeueInterval)
		return nil
	}

	return nil
}

func (ctrl *Controller) syncPodTagByPvcList(ctx context.Context,
	pvcNameList []string, podName, namespace string) error {
	unAvailableRtNameList := make([]string, 0, len(pvcNameList))
	for _, pvcName := range pvcNameList {
		pv, err := ctrl.getPvByPvcName(pvcName, namespace)
		if errors.IsNotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		rtName := getResourceTopologyName(pv.Name)
		rt, err := ctrl.topologyInformer.Lister().Get(rtName)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}

		if errors.IsNotFound(err) || rt.DeletionTimestamp != nil {
			unAvailableRtNameList = append(unAvailableRtNameList, rtName)
			continue
		}

		if isTagExist(rt.Spec.Tags, podName, namespace) {
			continue
		}

		rt.Spec.Tags = addPodTag(rt.Spec.Tags, podName, namespace)
		_, err = ctrl.xuanwuClient.XuanwuV1().ResourceTopologies().Update(ctx, rt, metaV1.UpdateOptions{})
		if err != nil {
			return err
		}

		log.AddContext(ctx).Infof("[pod-controller] add tag [%s/%s] to rt [%s] success",
			namespace, podName, rt.Name)
	}

	if len(unAvailableRtNameList) > 0 {
		log.AddContext(ctx).Warningf("[pod-controller] resourceTopologies %v are not available, "+
			"need to retry later", unAvailableRtNameList)
		ctrl.podQueue.AddAfter(getPodKey(namespace, podName), resourceRequeueInterval)
		return nil
	}

	return nil
}

func (ctrl *Controller) removePodTag(ctx context.Context, key string) error {
	log.AddContext(ctx).Infof("[pod-controller] start to delete pod [%s] relate tags", key)
	defer log.AddContext(ctx).Infof("[pod-controller] finished delete pod [%s] relate tags", key)
	podObj, isExist, err := ctrl.podStore.GetByKey(key)
	if err != nil {
		return err
	}

	if !isExist {
		log.AddContext(ctx).Errorf("[pod-controller] can not find the pod [%s] in cache", key)
		return nil
	}

	pod, ok := podObj.(*coreV1.Pod)
	if !ok {
		log.AddContext(ctx).Errorf("[pod-controller] invalid struct of the pod [%s] in cache", key)
		return err
	}

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}

		rt, err := ctrl.getRtByPvcName(volume.PersistentVolumeClaim.ClaimName, pod.Namespace)
		if errors.IsNotFound(err) {
			log.AddContext(ctx).Debugf("[pod-controller] can not find the resource: [%v], skip to next", err)
			continue
		}

		if err != nil {
			return err
		}

		if rt.DeletionTimestamp != nil {
			log.AddContext(ctx).Debugf("[pod-controller] rt [%s] is deleting, no need to delete tag", rt.Name)
			continue
		}

		if !isTagExist(rt.Spec.Tags, pod.Name, pod.Namespace) {
			continue
		}

		log.AddContext(ctx).Debugf("[pod-controller] try to delete rt [%s] tag [%s]", rt.Name, key)
		rt.Spec.Tags = deletePodTag(rt.Spec.Tags, pod.Name, pod.Namespace)
		_, err = ctrl.xuanwuClient.XuanwuV1().ResourceTopologies().Update(ctx, rt, metaV1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return ctrl.podStore.Delete(pod)
}

func filterAvailablePvcInPod(pod *coreV1.Pod) ([]string, error) {
	// running container set
	runningContainerNameSet := make(map[string]bool)
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Running != nil {
			runningContainerNameSet[containerStatus.Name] = true
		}
	}

	// pv-pvc map
	volumeMap := make(map[string]string)
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			volumeMap[volume.Name] = volume.PersistentVolumeClaim.ClaimName
		}
	}

	// container-pvcList map
	containerMap := make(map[string][]string)
	for _, container := range pod.Spec.Containers {
		containerMap[container.Name] = make([]string, 0)
		for _, volume := range container.VolumeMounts {
			containerMap[container.Name] = append(containerMap[container.Name], volumeMap[volume.Name])
		}

		for _, volume := range container.VolumeDevices {
			containerMap[container.Name] = append(containerMap[container.Name], volumeMap[volume.Name])
		}
	}

	pvcNameList := make([]string, 0)
	needRetryContainerList := make([]string, 0)
	for containerName, pvcList := range containerMap {
		if runningContainerNameSet[containerName] {
			pvcNameList = append(pvcNameList, pvcList...)
		} else if len(pvcList) > 0 {
			needRetryContainerList = append(needRetryContainerList, containerName)
		}
	}

	if len(needRetryContainerList) > 0 {
		return pvcNameList, fmt.Errorf("[pod-controller] containers %v in pod [%s] are not running, "+
			"need to retry later", needRetryContainerList, pod.Name)
	}
	return pvcNameList, nil
}

func (ctrl *Controller) getPvByPvcName(pvcName, namespace string) (*coreV1.PersistentVolume, error) {
	pvc, err := ctrl.claimInformer.Lister().PersistentVolumeClaims(namespace).Get(pvcName)
	if err != nil {
		return nil, err
	}

	return ctrl.volumeInformer.Lister().Get(pvc.Spec.VolumeName)
}

func (ctrl *Controller) getRtByPvcName(pvcName, namespace string) (*apiXuanwuV1.ResourceTopology, error) {
	pvc, err := ctrl.claimInformer.Lister().PersistentVolumeClaims(namespace).Get(pvcName)
	if err != nil {
		return nil, err
	}

	pv, err := ctrl.volumeInformer.Lister().Get(pvc.Spec.VolumeName)
	if err != nil {
		return nil, err
	}

	rtName := getResourceTopologyName(pv.Name)
	return ctrl.topologyInformer.Lister().Get(rtName)
}

func isTagExist(tags []apiXuanwuV1.Tag, podName, namespace string) bool {
	for _, tag := range tags {
		if tag.Name == podName && tag.Namespace == namespace {
			return true
		}
	}

	return false
}

func addPodTag(tags []apiXuanwuV1.Tag, podName, namespace string) []apiXuanwuV1.Tag {
	for _, tag := range tags {
		if tag.Name == podName && tag.Namespace == namespace {
			return tags
		}
	}

	addTag := apiXuanwuV1.Tag{
		ResourceInfo: apiXuanwuV1.ResourceInfo{
			TypeMeta:  metaV1.TypeMeta{Kind: consts.Pod, APIVersion: consts.KubernetesV1},
			Namespace: namespace,
			Name:      podName,
		},
	}

	tags = append(tags, addTag)
	return tags
}

func deletePodTag(tags []apiXuanwuV1.Tag, podName, namespace string) []apiXuanwuV1.Tag {
	for index, tag := range tags {
		if tag.Name == podName && tag.Namespace == namespace {
			return append(tags[:index], tags[index+1:]...)
		}
	}

	return tags
}

func getPodKey(namespace, name string) string {
	return namespace + "/" + name
}
