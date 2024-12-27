/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package metricscache

import (
	"errors"
	"strings"

	coreV1 "k8s.io/api/core/v1"

	exporterConfig "github.com/huawei/csm/v2/config/exporter"
	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
)

// volumeHandle str is sbcName.storageName. So when use strings.Split, volumeHandleStrLen is 2.
const volumeHandleStrLen = 2

type parsePVMetrics struct {
	collectDetail *storageGRPC.CollectDetail
	parseError    error
}

func (pvMetrics *parsePVMetrics) setCSIDriverNameMetrics(volume coreV1.PersistentVolume) *parsePVMetrics {
	if pvMetrics.parseError != nil {
		return pvMetrics
	}

	if volume.Spec.CSI == nil {
		pvMetrics.parseError = errors.New("can not get CSIDriverName")
		return pvMetrics
	}

	if volume.Spec.CSI.Driver != exporterConfig.GetCSIDriverName() {
		pvMetrics.parseError = errors.New("unsupported driver")
		return pvMetrics
	}

	pvMetrics.collectDetail.Data["driverName"] = volume.Spec.CSI.Driver
	return pvMetrics
}

func (pvMetrics *parsePVMetrics) setVolumeHandleMetrics(volume coreV1.PersistentVolume) *parsePVMetrics {
	if pvMetrics.parseError != nil {
		return pvMetrics
	}

	if volume.Spec.CSI == nil {
		pvMetrics.parseError = errors.New("can not get volumeHandle")
		return pvMetrics
	}
	volumeHandle := volume.Spec.CSI.VolumeHandle
	sbcConfigName := strings.Split(volumeHandle, ".")
	if len(sbcConfigName) != volumeHandleStrLen {
		pvMetrics.parseError = errors.New("can not parse volumeHandle, split error")
		return pvMetrics
	}

	pvMetrics.collectDetail.Data["sbcName"] = sbcConfigName[0]
	pvMetrics.collectDetail.Data["storageName"] = sbcConfigName[1]
	return pvMetrics
}

func (pvMetrics *parsePVMetrics) setPVNameMetrics(volume coreV1.PersistentVolume) *parsePVMetrics {
	if pvMetrics.parseError != nil {
		return pvMetrics
	}

	pvMetrics.collectDetail.Data["pvName"] = volume.ObjectMeta.Name
	return pvMetrics
}

func (pvMetrics *parsePVMetrics) setPVCNameMetrics(volume coreV1.PersistentVolume) *parsePVMetrics {
	if pvMetrics.parseError != nil {
		return pvMetrics
	}

	if volume.Spec.ClaimRef == nil {
		pvMetrics.parseError = errors.New("can not get PVCName")
		return pvMetrics
	}

	if volume.Spec.ClaimRef.Kind == "PersistentVolumeClaim" {
		pvMetrics.collectDetail.Data["pvcName"] = volume.Spec.ClaimRef.Name
	}
	return pvMetrics
}
