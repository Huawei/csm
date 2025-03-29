/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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
	"fmt"
	"strings"

	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/utils/log"
)

// MergePVMetricsData implement MergeMetricsData interface
type MergePVMetricsData struct {
	*BaseMergeMetricsData
}

func init() {
	RegisterMergeMetricsData("pv", NewMergePVMetricsData)
}

// NewMergePVMetricsData new a MergePVMetricsData
func NewMergePVMetricsData(backendName, monitorType, metricsType string,
	metricsIndicators []string) (MergeMetricsData, error) {
	return &MergePVMetricsData{BaseMergeMetricsData: &BaseMergeMetricsData{
		backendName: backendName, monitorType: monitorType, metricsType: metricsType,
		mergeIndicators: metricsIndicators}}, nil
}

func (mergePVMetricsData *MergePVMetricsData) mergeKubePVAndStorageInfo(ctx context.Context,
	storageNameKey, pvNameKey, volumeType string, pvCacheData []*storageGRPC.CollectDetail,
	metricsDataCache *MetricsDataCache) (map[string]map[string]string, error) {
	if len(pvCacheData) == 0 {
		return nil, errors.New("can not get the pv data when merge")
	}

	storageCacheData := metricsDataCache.GetMetricsData(volumeType)
	if storageCacheData == nil || len(storageCacheData.Details) == 0 {
		return nil, errors.New("can not get the storage data when merge")
	}

	var pvCacheDataMap = make(map[string]map[string]string, len(pvCacheData))
	for _, pvData := range pvCacheData {
		if pvData.Data[pvNameKey] == "" {
			continue
		}

		storageTypeName, storageTypeExit := pvData.Data["sbcStorageType"]
		if storageTypeExit && storageTypeMap[storageTypeName] != volumeType {
			continue
		}

		pvCacheDataMap[pvData.Data[pvNameKey]] = pvData.Data
	}

	var resultMerge = make(map[string]map[string]string)
	for _, mergeData := range storageCacheData.Details {
		if len(mergeData.Data) == 0 {
			continue
		}
		sameName := mergeData.Data[storageNameKey]
		if sameName == "" {
			continue
		}
		mergeData.Data["sameName"] = sameName
		resultMerge[sameName+mergeData.Data["ID"]] = mergeData.Data

		sameData, sameNameExist := pvCacheDataMap[sameName]
		if !sameNameExist {
			continue
		}

		for key, value := range sameData {
			resultMerge[sameName+mergeData.Data["ID"]][key] = value
		}
	}
	return resultMerge, nil
}

func (mergePVMetricsData *MergePVMetricsData) getPVMergeParams() (string, []string, error) {
	var metricsIndicatorsList []string
	var storageNameKey string
	if mergePVMetricsData.monitorType == "performance" && len(mergePVMetricsData.mergeIndicators) == 0 {
		errorStr := "when get pv merge params, the monitorType is performance but mergeIndicators is empty"
		return storageNameKey, metricsIndicatorsList, errors.New(errorStr)
	}

	if mergePVMetricsData.monitorType == "performance" {
		storageNameKey = "ObjectName"
		metricsIndicatorsList = strings.Split(mergePVMetricsData.mergeIndicators[0], ",")
	} else {
		storageNameKey = "NAME"
		metricsIndicatorsList = []string{"lun", "filesystem"}
	}
	if len(metricsIndicatorsList) == 0 {
		errorStr := "when get pv merge params, the metricsIndicatorsList is empty"
		return storageNameKey, metricsIndicatorsList, errors.New(errorStr)
	}

	return storageNameKey, metricsIndicatorsList, nil
}

// MergeData merge pv data and storage data
func (mergePVMetricsData *MergePVMetricsData) MergeData(ctx context.Context,
	metricsDataCache *MetricsDataCache) error {
	log.AddContext(ctx).Infoln("start to merge pv and storage data")
	storageNameKey, metricsIndicatorsList, err := mergePVMetricsData.getPVMergeParams()
	if err != nil {
		return fmt.Errorf("can not get pv merge params, err is [%w]", err)
	}

	pvCacheData, ok := metricsDataCache.CacheDataMap["pv"]
	if !ok {
		return errors.New("can not get pv cache data when MergePVAndStorageData")
	}

	pvMetricsDataResponse := pvCacheData.GetMetricsDataResponse()
	if pvMetricsDataResponse == nil {
		return errors.New("can not get MetricsDataResponse data when MergePVAndStorageData")
	}

	if len(pvMetricsDataResponse.Details) == 0 {
		return errors.New("can not get MetricsDataResponse.Details when MergePVAndStorageData")
	}

	pvTempData := pvMetricsDataResponse.Details
	pvMetricsDataResponse.Details = nil
	for _, volumeType := range metricsIndicatorsList {
		mergeMapData, err := mergePVMetricsData.mergeKubePVAndStorageInfo(
			ctx, storageNameKey, "storageName", volumeType, pvTempData, metricsDataCache)
		if err != nil {
			log.AddContext(ctx).Errorf("merge pv metricsData of [%s] failed, "+
				"err is [%v], try to merge next data", volumeType, err)
			continue
		}
		for _, value := range mergeMapData {
			_, ok := value["pvName"]
			if !ok {
				continue
			}
			pvMetricsDataResponse.Details = append(pvMetricsDataResponse.Details,
				&storageGRPC.CollectDetail{Data: value})
		}
	}
	log.AddContext(ctx).Infoln("merge pv and storage data success")
	return nil
}
