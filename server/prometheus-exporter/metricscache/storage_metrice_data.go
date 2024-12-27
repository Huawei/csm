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

	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
	clientSet "github.com/huawei/csm/v2/server/prometheus-exporter/clientset"
	"github.com/huawei/csm/v2/utils/log"
)

// StorageMetricsData save one batch data with storage MetricsType,
// from prometheus request
type StorageMetricsData struct {
	*BaseMetricsData
}

func init() {
	RegisterMetricsData("array", NewStorageMetricsData)
	RegisterMetricsData("controller", NewStorageMetricsData)
	RegisterMetricsData("storagepool", NewStorageMetricsData)
	RegisterMetricsData("filesystem", NewStorageMetricsData)
	RegisterMetricsData("lun", NewStorageMetricsData)
}

// NewStorageMetricsData new a StorageMetricsData
func NewStorageMetricsData(backendName, metricsType string) (MetricsData, error) {
	return &StorageMetricsData{BaseMetricsData: &BaseMetricsData{
		BackendName: backendName, MetricsType: metricsType}}, nil
}

func (storageMetricsData *StorageMetricsData) buildTheStorageGRPCRequest(
	collectorName, monitorType string, metricsIndicators []string) *storageGRPC.CollectRequest {
	batchCollectRequest := &storageGRPC.CollectRequest{
		BackendName: storageMetricsData.BackendName,
		CollectType: collectorName,
		MetricsType: monitorType,
		Indicators:  []string{},
	}
	if monitorType != "performance" {
		return batchCollectRequest
	}

	if len(metricsIndicators) == 0 || metricsIndicators[0] == "" {
		return nil
	}

	batchCollectRequest.Indicators = strings.Split(metricsIndicators[0], ",")
	return batchCollectRequest
}

func (storageMetricsData *StorageMetricsData) getStorageData(ctx context.Context,
	batchCollectRequest *storageGRPC.CollectRequest,
	usedClientSet *clientSet.ClientsSet) (*storageGRPC.CollectResponse, error) {
	storageGRPCClient := usedClientSet.StorageGRPCClientSet.CollectorClient
	batchCollectResponse, err := storageGRPCClient.Collect(ctx, batchCollectRequest)
	if err != nil {
		log.AddContext(ctx).Warningf("can not get storage response the err is [%v]", err)
		return nil, errors.New("please continue collect next data")
	}
	if batchCollectResponse.Details == nil {
		log.AddContext(ctx).Warningln("can not get storage data")
		return nil, errors.New("please continue collect next data")
	}
	return batchCollectResponse, nil
}

// SetMetricsData set storage data MetricsDataResponse
func (storageMetricsData *StorageMetricsData) SetMetricsData(ctx context.Context,
	collectorName, monitorType string, metricsIndicators []string) error {
	log.AddContext(ctx).Infof("start to get storage metrics data with collector name: %v, monitor type: %v, "+
		"indicators: %v", collectorName, monitorType, metricsIndicators)
	usedClientSet := clientSet.GetExporterClientSet()
	if usedClientSet.InitError != nil {
		log.AddContext(ctx).Errorln("can not get Client Set when get data")
		return errors.New("can not get Client Set when get data")
	}

	// get storage data
	batchCollectRequest := storageMetricsData.buildTheStorageGRPCRequest(
		collectorName, monitorType, metricsIndicators)
	batchCollectResponse, err := storageMetricsData.getStorageData(ctx, batchCollectRequest, usedClientSet)
	if err != nil {
		return err
	}
	storageMetricsData.MetricsDataResponse = batchCollectResponse
	log.AddContext(ctx).Infoln("get storage metrics data success")
	return nil
}
