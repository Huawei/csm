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

// Package resource defines some support resource interface for topology
package resource

import (
	"fmt"

	coreV1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiXuanwuV1 "github.com/huawei/csm/v2/client/apis/xuanwu/v1"
	"github.com/huawei/csm/v2/controller/utils"
	"github.com/huawei/csm/v2/controller/utils/consts"
	"github.com/huawei/csm/v2/utils/log"
	"github.com/huawei/csm/v2/utils/resource"
)

const (
	podV1Kind       = "Pod"
	podV1ApiVersion = "v1"
)

// PodV1Tag represents a pod in v1 version
type PodV1Tag struct {
	name      string
	namespace string
}

// InitFromResourceInfo init PodV1Tag struct from resource info
func (p *PodV1Tag) InitFromResourceInfo(info apiXuanwuV1.ResourceInfo) {
	p.name = info.Name
	p.namespace = info.Namespace
}

// HasOwner checks if pod has OwnerReferences field
func (p *PodV1Tag) HasOwner() bool {
	pod := &coreV1.Pod{}
	err := utils.RetryFunc(func() (bool, error) {
		var err error
		pod, err = resource.Instance().GetPodByNameSpaceAndName(p.namespace, p.name, metaV1.GetOptions{})
		if err != nil && apiErrors.IsNotFound(err) {
			return true, err
		}
		if err != nil {
			return false, err
		}
		return true, nil
	}, consts.RetryTimes, consts.RetryDurationInit, consts.RetryDurationMax)

	if err != nil {
		log.Errorf("get Pod [%s/%s] failed: [%v]", p.namespace, p.name, err)
		return false
	}
	return len(pod.OwnerReferences) != 0
}

// GetOwner returns the list of owners inner tag of the pod
func (p *PodV1Tag) GetOwner() (InnerTag, error) {
	var info apiXuanwuV1.ResourceInfo
	err := utils.RetryFunc(func() (bool, error) {
		pod, err := resource.Instance().GetPodByNameSpaceAndName(p.namespace, p.name, metaV1.GetOptions{})
		if err != nil && apiErrors.IsNotFound(err) {
			return true, err
		}
		if err != nil {
			return false, err
		}

		info = p.getOwnerResourceInfo(pod, p.namespace)

		return true, nil
	}, consts.RetryTimes, consts.RetryDurationInit, consts.RetryDurationMax)

	if err != nil {
		log.Errorf("get Pod [%s/%s] failed: [%v]", p.namespace, p.name, err)
		return nil, err
	}

	innerTag, err := NewInnerTag(info.TypeMeta)
	if err != nil {
		log.Warningf("get pod [%s/%s] owner references failed: [%v]", p.namespace, p.name, err)
		return nil, nil
	}

	return innerTag, nil
}

// Exists checks if the pod exists in cluster
func (p *PodV1Tag) Exists() bool {
	err := utils.RetryFunc(func() (bool, error) {
		_, err := resource.Instance().GetPodByNameSpaceAndName(p.namespace, p.name, metaV1.GetOptions{})
		if err == nil {
			return true, nil
		}
		if apiErrors.IsNotFound(err) {
			return true, err
		}
		return false, err
	}, consts.RetryTimes, consts.RetryDurationInit, consts.RetryDurationMax)

	if err != nil {
		log.Errorf("get Pod [%s/%s] failed: [%v]", p.namespace, p.name, err)
		return false
	}
	return true
}

// Ready checks whether the pod is in Running or Succeeded status
func (p *PodV1Tag) Ready() bool {
	err := utils.RetryFunc(func() (bool, error) {
		pod, err := resource.Instance().GetPodByNameSpaceAndName(p.namespace, p.name, metaV1.GetOptions{})
		if err != nil && apiErrors.IsNotFound(err) {
			return true, err
		}
		if err != nil {
			return false, err
		}
		if pod.Status.Phase == coreV1.PodFailed {
			return true, err
		}
		if pod.Status.Phase != coreV1.PodRunning && pod.Status.Phase != coreV1.PodSucceeded {
			return false, fmt.Errorf("pv is not in [%s/%s] status, "+
				"cur status [%s]", coreV1.PodRunning, coreV1.PodSucceeded, pod.Status.Phase)
		}
		return true, nil
	}, consts.RetryTimes, consts.RetryDurationInit, consts.RetryDurationMax)

	if err != nil {
		log.Errorf("check Pod [%s/%s] ready failed: [%v]", p.namespace, p.name, err)
		return false
	}
	return true
}

// IsDeleting checks whether the pod has DeletionTimestamp
func (p *PodV1Tag) IsDeleting() bool {
	pod := &coreV1.Pod{}
	err := utils.RetryFunc(func() (bool, error) {
		var err error
		pod, err = resource.Instance().GetPodByNameSpaceAndName(p.namespace, p.name, metaV1.GetOptions{})
		if err != nil && apiErrors.IsNotFound(err) {
			return true, err
		}
		if err != nil {
			return false, err
		}
		return true, nil
	}, consts.RetryTimes, consts.RetryDurationInit, consts.RetryDurationMax)

	if err != nil {
		log.Errorf("check Pod [%s/%s] ready failed: [%v]", p.namespace, p.name, err)
		return false
	}
	return pod.DeletionTimestamp != nil
}

// ToResourceInfo converts a PodV1Tag to a ResourceInfo
func (p *PodV1Tag) ToResourceInfo() apiXuanwuV1.ResourceInfo {
	return apiXuanwuV1.ResourceInfo{
		TypeMeta: metaV1.TypeMeta{
			Kind:       podV1Kind,
			APIVersion: podV1ApiVersion,
		},
		Namespace: p.namespace,
		Name:      p.name,
	}
}

func (p *PodV1Tag) getOwnerResourceInfo(pod *coreV1.Pod, namespace string) apiXuanwuV1.ResourceInfo {
	var info apiXuanwuV1.ResourceInfo
	for _, ref := range pod.OwnerReferences {
		if !*ref.Controller {
			continue
		}
		info = apiXuanwuV1.ResourceInfo{
			TypeMeta: metaV1.TypeMeta{
				Kind:       ref.Kind,
				APIVersion: ref.APIVersion,
			},
			Namespace: namespace,
			Name:      ref.Name,
		}
		break
	}
	return info
}
