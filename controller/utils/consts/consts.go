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

// Package consts constants for topology services
package consts

import "time"

const (
	// RetryTimes default function retry times
	RetryTimes = 10
	// RetryDurationInit default function retry init duration
	RetryDurationInit = 500 * time.Millisecond
	// RetryDurationMax default function retry max duration
	RetryDurationMax = 10 * time.Second
)

const (
	// Pod is a tag type Pod
	Pod = "Pod"

	// PersistentVolume is a tag type PersistentVolume
	PersistentVolume = "PersistentVolume"

	// TopologyKind is topology resource kind
	TopologyKind = "ResourceTopology"

	// KubernetesV1 is kubernetes v1 api version
	KubernetesV1 = "v1"

	// XuanwuV1 is xuanwu v1 api version
	XuanwuV1 = "xuanwu.huawei.io/v1"

	// AnnDynamicallyProvisioned is annotation pointed to provider
	AnnDynamicallyProvisioned = "pv.kubernetes.io/provisioned-by"

	// VolumeHandleKeyLabel is label key used to search volume handle
	VolumeHandleKeyLabel = "volumehandlekey"
)
