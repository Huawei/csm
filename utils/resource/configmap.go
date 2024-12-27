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

// Package resource used to access the k8s core resource by API
package resource

import (
	"context"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigmapOps can get information of configmap
type ConfigmapOps interface {
	// CreateConfigmap creates the given configmap
	CreateConfigmap(*coreV1.ConfigMap) (*coreV1.ConfigMap, error)
	// GetConfigmap gets the configmap object given its name and namespace
	GetConfigmap(name, namespace string) (*coreV1.ConfigMap, error)
	// UpdateConfigmap update the configmap object given its name and namespace
	UpdateConfigmap(*coreV1.ConfigMap) (*coreV1.ConfigMap, error)
}

// CreateConfigmap creates the given configmap
func (c *Client) CreateConfigmap(configmap *coreV1.ConfigMap) (*coreV1.ConfigMap, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().ConfigMaps(configmap.Namespace).Create(
		context.TODO(), configmap, metaV1.CreateOptions{})
}

// GetConfigmap gets the configmap object given its name and namespace
func (c *Client) GetConfigmap(name, namespace string) (*coreV1.ConfigMap, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metaV1.GetOptions{})
}

// UpdateConfigmap update the given configmap
func (c *Client) UpdateConfigmap(configmap *coreV1.ConfigMap) (*coreV1.ConfigMap, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubernetes.CoreV1().ConfigMaps(configmap.Namespace).Update(
		context.TODO(), configmap, metaV1.UpdateOptions{})
}
