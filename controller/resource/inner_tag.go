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

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiXuanwuV1 "github.com/huawei/csm/v2/client/apis/xuanwu/v1"
)

var factoryMap map[metaV1.TypeMeta]func() InnerTag

func init() {
	factoryMap = make(map[metaV1.TypeMeta]func() InnerTag)
	factoryMap[metaV1.TypeMeta{Kind: podV1Kind, APIVersion: podV1ApiVersion}] = func() InnerTag { return &PodV1Tag{} }
	factoryMap[metaV1.TypeMeta{Kind: persistentVolumeV1Kind, APIVersion: persistentVolumeV1ApiVersion}] =
		func() InnerTag { return &PersistentVolumeV1Tag{} }
}

// InnerTag defines some support resource interface for topology tags change
type InnerTag interface {
	// InitFromResourceInfo init InnerTag by resource info
	InitFromResourceInfo(apiXuanwuV1.ResourceInfo)
	// HasOwner check if resource has owner
	HasOwner() bool
	// GetOwner returns owner resource inner tags of resource
	GetOwner() (InnerTag, error)
	// Exists checks if resource exists
	Exists() bool
	// Ready checks if resource is ready
	Ready() bool
	// IsDeleting check if resource is deleting
	IsDeleting() bool
	// ToResourceInfo convert inner tag to resource info
	ToResourceInfo() apiXuanwuV1.ResourceInfo
}

// NewInnerTag returns an empty InnerTag interface implementation object
func NewInnerTag(meta metaV1.TypeMeta) (InnerTag, error) {
	factory, ok := factoryMap[meta]
	if !ok {
		return nil, fmt.Errorf("unsupported tag type [%v]", meta)
	}
	return factory(), nil
}
