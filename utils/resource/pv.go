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

// Package resource is used to obtain core resources in Kubernetes.
package resource

import (
	"context"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PVOps is related with pv
type PVOps interface {
	// GetPV get pv by given name
	GetPV(name string) (*coreV1.PersistentVolume, error)
	// ListPV get pv by options
	ListPV(options metaV1.ListOptions) (*coreV1.PersistentVolumeList, error)
}

// GetPV get pv by given name
func (c *Client) GetPV(name string) (*coreV1.PersistentVolume, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().PersistentVolumes().Get(context.TODO(), name, metaV1.GetOptions{})
}

// ListPV get pv by options
func (c *Client) ListPV(options metaV1.ListOptions) (*coreV1.PersistentVolumeList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().PersistentVolumes().List(context.TODO(), options)
}
