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

// Package metricscache use to save query the data of the storage metrics once
package metricscache

import (
	"context"
	"errors"
	"strings"

	xuanwuV1 "github.com/Huawei/eSDK_K8S_Plugin/v4/client/apis/xuanwu/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	exporterConfig "github.com/huawei/csm/v2/config/exporter"
	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
	clientSet "github.com/huawei/csm/v2/server/prometheus-exporter/clientset"
	"github.com/huawei/csm/v2/utils/log"
)

// when use kubelet get pv -A -v6 the limit set is 500, so we set same as kubelet.
const getPVLimit = 500

// sbc configMap is sbcNameSpace/sbcName so when use strings.Split, it len is 2.
const sbConfigMapLen = 2

func getPVDataFromApi(ctx context.Context) []coreV1.PersistentVolume {
	usedClientSet := clientSet.GetExporterClientSet()
	var allVolumeItems []coreV1.PersistentVolume
	var continueKey string
	if usedClientSet.KubeClient == nil {
		return allVolumeItems
	}

	for {
		VolumeList, err := usedClientSet.KubeClient.CoreV1().PersistentVolumes().List(
			ctx, metaV1.ListOptions{Limit: getPVLimit, Continue: continueKey})
		if err != nil {
			log.AddContext(ctx).Errorln("can not get pv list")
			break
		}
		allVolumeItems = append(allVolumeItems, VolumeList.Items...)
		continueKey = VolumeList.ListMeta.Continue
		if continueKey == "" {
			break
		}
	}
	return allVolumeItems
}

func parseAllBackendInfo(allBackend *xuanwuV1.StorageBackendClaimList) map[string]map[string]string {
	var allSBCInfo = make(map[string]map[string]string)
	for _, sbcInfo := range allBackend.Items {
		if sbcInfo.Status == nil {
			continue
		}
		sbcStorageType := sbcInfo.Status.StorageType

		sbcConfigName := strings.Split(sbcInfo.Spec.ConfigMapMeta, "/")
		if len(sbcConfigName) != sbConfigMapLen {
			continue
		}
		sbcNameSpace := sbcConfigName[0]
		sbcName := sbcConfigName[1]

		allSBCInfo[sbcName] = map[string]string{"namespace": sbcNameSpace, "sbcStorageType": sbcStorageType}
	}
	return allSBCInfo
}

func getAllBackendFromApi(ctx context.Context) map[string]map[string]string {
	usedClientSet := clientSet.GetExporterClientSet()

	if usedClientSet.SbcClient == nil {
		return nil
	}

	allBackend, err := usedClientSet.SbcClient.XuanwuV1().StorageBackendClaims(
		exporterConfig.GetStorageBackendNamespace()).List(ctx, metaV1.ListOptions{})
	if err != nil {
		return nil
	}
	allBackendInfo := parseAllBackendInfo(allBackend)
	return allBackendInfo

}

func buildOutPVData(backendName, collectType string, allSBCInfo map[string]map[string]string,
	allPVData []coreV1.PersistentVolume) *storageGRPC.CollectResponse {
	outPVData := &storageGRPC.CollectResponse{
		BackendName: backendName,
		CollectType: collectType,
		Details:     []*storageGRPC.CollectDetail{}}
	for _, pvData := range allPVData {
		pvMapInfo := &parsePVMetrics{
			collectDetail: &storageGRPC.CollectDetail{Data: make(map[string]string)}}
		pvMapInfo.setCSIDriverNameMetrics(pvData).
			setVolumeHandleMetrics(pvData).
			setPVNameMetrics(pvData).
			setPVCNameMetrics(pvData)
		if pvMapInfo.parseError != nil {
			continue
		}
		pvBackendName, ok := pvMapInfo.collectDetail.Data["sbcName"]
		if !ok {
			continue
		}

		sbcInfo, ok := allSBCInfo[pvBackendName]
		if !ok {
			continue
		}
		pvMapInfo.collectDetail.Data["sbcStorageType"], ok = sbcInfo["sbcStorageType"]
		if !ok {
			continue
		}
		outPVData.Details = append(outPVData.Details, pvMapInfo.collectDetail)
	}
	return outPVData
}

// GetAndParsePVInfo gets and parses pv info with special collect type
func GetAndParsePVInfo(ctx context.Context, backendName, collectType string) (*storageGRPC.CollectResponse, error) {
	allPVData := getPVDataFromApi(ctx)
	if len(allPVData) == 0 {
		log.AddContext(ctx).Warningln("can not get pv data, pv is empty")
		return nil, errors.New("can not get pv data, pv is empty")
	}
	allSBCInfo := getAllBackendFromApi(ctx)
	if len(allSBCInfo) == 0 {
		log.AddContext(ctx).Warningln("can not get sbc data, sbc is empty")
		return nil, errors.New("can not get sbc data, sbc is empty")
	}

	outPVData := buildOutPVData(backendName, collectType, allSBCInfo, allPVData)
	return outPVData, nil
}
