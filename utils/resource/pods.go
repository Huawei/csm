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

// PodOps can get information of pod
type PodOps interface {
	GetPodListFilterByNamespace(namespace string, listOptions metaV1.ListOptions) (
		*coreV1.PodList, error)
	GetPodByNameSpaceAndName(namespace, name string, getOptions metaV1.GetOptions) (
		*coreV1.Pod, error)
}

// GetPodListFilterByNamespace gets pod list from k8s cluster by namespace and listOption
func (c *Client) GetPodListFilterByNamespace(namespace string, listOptions metaV1.ListOptions) (
	*coreV1.PodList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
}

// GetPodByNameSpaceAndName gets pod  from k8s cluster by namespace, name and listOption
func (c *Client) GetPodByNameSpaceAndName(namespace, name string, getOptions metaV1.GetOptions) (
	*coreV1.Pod, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	return c.kubernetes.CoreV1().Pods(namespace).Get(context.TODO(), name, getOptions)
}
