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
	persistentVolumeV1Kind       = "PersistentVolume"
	persistentVolumeV1ApiVersion = "v1"
)

// PersistentVolumeV1Tag represents a persistent volume in v1 version
type PersistentVolumeV1Tag struct {
	name string
}

// InitFromResourceInfo init PersistentVolumeV1Tag struct from resource info
func (p *PersistentVolumeV1Tag) InitFromResourceInfo(info apiXuanwuV1.ResourceInfo) {
	p.name = info.Name
}

// HasOwner persistent volume don't need mark owner
func (p *PersistentVolumeV1Tag) HasOwner() bool {
	return false
}

// GetOwner persistent volume don't need mark owner
func (p *PersistentVolumeV1Tag) GetOwner() (InnerTag, error) {
	return nil, nil
}

// Exists check if persistent volume exist in cluster
func (p *PersistentVolumeV1Tag) Exists() bool {
	err := utils.RetryFunc(func() (bool, error) {
		_, err := resource.Instance().GetPV(p.name)
		if err == nil {
			return true, nil
		}
		if apiErrors.IsNotFound(err) {
			return true, err
		}
		return false, err
	}, consts.RetryTimes, consts.RetryDurationInit, consts.RetryDurationMax)

	if err != nil {
		log.Errorf("get PersistentVolume [%s] failed: [%v]", p.name, err)
		return false
	}
	return true
}

// Ready check if persistent volume is Bounded
func (p *PersistentVolumeV1Tag) Ready() bool {
	err := utils.RetryFunc(func() (bool, error) {
		pv, err := resource.Instance().GetPV(p.name)
		if err != nil && apiErrors.IsNotFound(err) {
			return true, err
		}
		if err != nil {
			return false, err
		}
		if pv.Status.Phase != coreV1.VolumeBound {
			return false, fmt.Errorf("pv [%s] is not in bound status", p.name)
		}
		return true, nil
	}, consts.RetryTimes, consts.RetryDurationInit, consts.RetryDurationMax)
	if err != nil {
		log.Errorf("check PersistentVolume [%s] ready failed: [%v]", p.name, err)
		return false
	}
	return true
}

// IsDeleting check if persistent volume has DeletionTimestamp field
func (p *PersistentVolumeV1Tag) IsDeleting() bool {
	pv := &coreV1.PersistentVolume{}
	err := utils.RetryFunc(func() (bool, error) {
		var err error
		pv, err = resource.Instance().GetPV(p.name)
		if err != nil && apiErrors.IsNotFound(err) {
			return true, err
		}

		if err != nil {
			return false, err
		}

		return true, nil
	}, consts.RetryTimes, consts.RetryDurationInit, consts.RetryDurationMax)
	if err != nil {
		log.Errorf("check PersistentVolume [%s] deleting stage failed: [%v]", p.name, err)
		return false
	}
	return pv.DeletionTimestamp != nil
}

// ToResourceInfo converts PersistentVolumeV1Tag to resource info
func (p *PersistentVolumeV1Tag) ToResourceInfo() apiXuanwuV1.ResourceInfo {
	return apiXuanwuV1.ResourceInfo{
		TypeMeta: metaV1.TypeMeta{
			Kind:       persistentVolumeV1Kind,
			APIVersion: persistentVolumeV1ApiVersion,
		},
		Name: p.name,
	}
}
