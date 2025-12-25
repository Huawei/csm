/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

// Package metricscache Package metrics cache use to save query the data of the storage metrics once
package metricscache

import (
	"context"
	"errors"
	"slices"
	"strings"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xuanwuV1 "github.com/Huawei/eSDK_K8S_Plugin/v4/client/apis/xuanwu/v1"
	exporterConfig "github.com/huawei/csm/v2/config/exporter"
	"github.com/huawei/csm/v2/controller/utils/consts"
	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
	clientSet "github.com/huawei/csm/v2/server/prometheus-exporter/clientset"
	"github.com/huawei/csm/v2/utils/log"
)

// DefaultVstoreID the default vstore id when can not get vstoreID
const DefaultVstoreID = "0"

// DefaultVstoreName the default vstore name when can not get vstoreName
const DefaultVstoreName = "System_vStore"

// MetricsVstoreData save one batch data with special MetricsType,
// from prometheus request
type MetricsVstoreData struct {
	*BaseMetricsData
}

func init() {
	RegisterMetricsData("vstore", NewMetricsVstoreData)
}

// NewMetricsVstoreData creates a new MetricsVstoreData with special MetricsType
func NewMetricsVstoreData(backendName, metricsType string) (MetricsData, error) {
	return &MetricsVstoreData{BaseMetricsData: &BaseMetricsData{
		BackendName: backendName, MetricsType: metricsType}}, nil
}

// SetMetricsData set Vstore data MetricsDataResponse
func (metricsData *MetricsVstoreData) SetMetricsData(ctx context.Context,
	collectorName, monitorType string, metricsIndicators []string) error {
	log.AddContext(ctx).Infof("start to get vstore metrics data with collector name: %v, monitor type: %v, "+
		"indicators: %v", collectorName, monitorType, metricsIndicators)

	backendContent, err := getBackendContentFromApi(ctx)
	if err != nil {
		log.AddContext(ctx).Errorf("can not get backend content, err is [%v]", err)
		return err
	}
	batchCollectResponse := buildVstoreOutData(ctx, metricsData.BackendName, collectorName, backendContent)

	metricsData.MetricsDataResponse = batchCollectResponse
	log.AddContext(ctx).Infoln("get vstore metrics data success")
	return nil
}

// getBackendContentFromApi get Vstore and Storage data from sbcClient
func getBackendContentFromApi(ctx context.Context) (*xuanwuV1.StorageBackendContentList, error) {
	usedClientSet := clientSet.GetExporterClientSet()

	if usedClientSet.SbcClient == nil {
		return nil, errors.New("get sbc client failed")
	}

	backendContents, err := usedClientSet.SbcClient.XuanwuV1().StorageBackendContents().List(ctx, metaV1.ListOptions{})
	if err != nil {
		log.AddContext(ctx).Errorf("can not get sbct list, err is [%v]", err)
		return nil, err
	}

	backendClaims, err := usedClientSet.SbcClient.XuanwuV1().StorageBackendClaims(
		exporterConfig.GetStorageBackendNamespace()).List(ctx, metaV1.ListOptions{})
	if err != nil {
		log.AddContext(ctx).Errorf("can not get sbc list, err is [%v]", err)
		return nil, err
	}

	filterVstoreBackend(backendClaims, backendContents)

	return backendContents, nil
}

// filterVstoreBackend filter sbct list with supported storage type
func filterVstoreBackend(backendClaims *xuanwuV1.StorageBackendClaimList,
	backendContents *xuanwuV1.StorageBackendContentList) {
	supportStorageTypeContentList := make([]xuanwuV1.StorageBackendContent, 0)
	backendContentItemMap := make(map[string]xuanwuV1.StorageBackendContent)
	for _, sbct := range backendContents.Items {
		backendContentItemMap[sbct.Spec.BackendClaim] = sbct
	}
	for _, sbc := range backendClaims.Items {
		if slices.Contains(consts.SupportedType, sbc.Status.StorageType) {
			if sbct, ok := backendContentItemMap[sbc.Namespace+"/"+sbc.Name]; ok {
				supportStorageTypeContentList = append(supportStorageTypeContentList, sbct)
			}
		}
	}
	backendContents.Items = supportStorageTypeContentList
}

// buildVstoreOutData build vstore data from sbct list and return final matrix responses
func buildVstoreOutData(ctx context.Context, backendName, collectType string,
	backendContents *xuanwuV1.StorageBackendContentList) *storageGRPC.CollectResponse {
	outVstoreData := &storageGRPC.CollectResponse{
		BackendName: backendName,
		CollectType: collectType,
		Details:     []*storageGRPC.CollectDetail{}}
	for _, sbctInfo := range backendContents.Items {
		if sbctInfo.Status == nil {
			continue
		}

		if sbctInfo.Status.Pools == nil {
			continue
		}
		for _, pool := range sbctInfo.Status.Pools {
			fillVstorePoolDetail(pool, sbctInfo, outVstoreData)
		}
	}
	return outVstoreData
}

// fillVstorePoolDetail fill vstore pool detail data into outVstoreData
func fillVstorePoolDetail(pool xuanwuV1.Pool, sbctInfo xuanwuV1.StorageBackendContent,
	outVstoreData *storageGRPC.CollectResponse) {
	singleCollectDetail := &storageGRPC.CollectDetail{Data: make(map[string]string)}

	sbcConfigName := strings.Split(sbctInfo.Spec.ConfigmapMeta, "/")
	if len(sbcConfigName) != sbConfigMapLen {
		return
	}
	singleCollectDetail.Data["BackendName"] = sbcConfigName[1]

	if vstoreID, ok := sbctInfo.Status.Specification["VStoreID"]; ok {
		singleCollectDetail.Data["VStoreID"] = vstoreID
	} else {
		singleCollectDetail.Data["VStoreID"] = DefaultVstoreID
	}

	if vstoreName, ok := sbctInfo.Status.Specification["VStoreName"]; ok {
		singleCollectDetail.Data["VStoreName"] = vstoreName
	} else {
		singleCollectDetail.Data["VStoreName"] = DefaultVstoreName
	}

	singleCollectDetail.Data["PoolName"] = pool.Name

	singleCollectDetail.Data["TotalCapacity"] = pool.Capacities["TotalCapacity"]
	singleCollectDetail.Data["FreeCapacity"] = pool.Capacities["FreeCapacity"]
	singleCollectDetail.Data["UsedCapacity"] = pool.Capacities["UsedCapacity"]

	outVstoreData.Details = append(outVstoreData.Details, singleCollectDetail)
}
