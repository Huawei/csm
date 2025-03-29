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
	"fmt"
	"strings"

	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/utils/log"
)

var storageTypeMap = map[string]string{
	"oceanstor-san": "lun",
	"oceanstor-nas": "filesystem",
}
var pvPerformanceMap = map[string][]string{
	"lun":        {"21,22,370"},
	"filesystem": {"182,524,525"},
}

// MetricsDataCache save one batch data from prometheus request
type MetricsDataCache struct {
	BackendName  string
	CacheDataMap map[string]MetricsData
	MergeMetrics map[string]MergeMetricsData
}

// GetMetricsData get the CollectResponse from storage
func (metricsDataCache *MetricsDataCache) GetMetricsData(metricsType string) *storageGRPC.CollectResponse {
	if _, ok := metricsDataCache.CacheDataMap[metricsType]; !ok {
		return nil
	}

	return metricsDataCache.CacheDataMap[metricsType].GetMetricsDataResponse()
}

// SetBatchDataFromSource set batch data to CacheDataMap
func (metricsDataCache *MetricsDataCache) SetBatchDataFromSource(ctx context.Context,
	monitorType string, params map[string][]string) {
	log.AddContext(ctx).Infoln("start to fill batch data from source")

	for collectorName, metricsIndicators := range params {
		metricsData, ok := metricsDataCache.CacheDataMap[collectorName]
		if !ok {
			log.AddContext(ctx).Errorf("get [%s] metrics data failed with monitorType [%s]",
				collectorName, monitorType)
			continue
		}
		// if the collectorName already set we not set again
		metricsDataResponse := metricsData.GetMetricsDataResponse()
		if metricsDataResponse != nil {
			log.AddContext(ctx).Debugf("the metrics data of %s response is already got", collectorName)
			continue
		}

		err := metricsData.SetMetricsData(ctx, collectorName, monitorType, metricsIndicators)
		if err != nil {
			log.AddContext(ctx).Errorf("set metrics data for %s failed, err is [%v], try to collect next data",
				collectorName, err)
			continue
		}
	}
	log.AddContext(ctx).Infoln("fill batch data success")
}

// MergeBatchData Merge batch data of MergeMetrics is not empty.
// use the MergeData interface to get need Merge Metrics like pv Metrics
func (metricsDataCache *MetricsDataCache) MergeBatchData(ctx context.Context) {
	if len(metricsDataCache.MergeMetrics) == 0 {
		return
	}
	log.AddContext(ctx).Infoln("start to merge metrics data")
	for mergeMetricsName, mergeMetricsClass := range metricsDataCache.MergeMetrics {
		err := mergeMetricsClass.MergeData(ctx, metricsDataCache)
		if err != nil {
			log.AddContext(ctx).Errorf("can not merge [%s] metrics data, err is [%v], try to merge next",
				mergeMetricsName, err)
		}
	}
	log.AddContext(ctx).Infoln("merge metrics data success")
}

func (metricsDataCache *MetricsDataCache) buildPVBatchParams(ctx context.Context,
	monitorType string, params, batchParams map[string][]string) error {
	if batchParams == nil {
		batchParams = make(map[string][]string)
	}

	indicators, ok := params["pv"]
	if !ok {
		log.AddContext(ctx).Debugf("no need to build pv class with params [%v]", params)
		return nil
	}

	if monitorType == "performance" {
		if len(indicators) == 0 || indicators[0] == "" {
			return fmt.Errorf("pv indicators [%v] are invalid with performance metrics type", indicators)
		}

		indicatorsList := strings.Split(indicators[0], ",")
		for _, metrics := range indicatorsList {
			metricsIndicators, exits := pvPerformanceMap[metrics]
			if !exits {
				log.AddContext(ctx).Warningf("pv indiscator [%v] is invalid with performance metrics type", metrics)
				continue
			}
			batchParams[metrics] = metricsIndicators
		}
	} else {
		batchParams["lun"] = []string{""}
		batchParams["filesystem"] = []string{""}
	}
	return nil
}

func (metricsDataCache *MetricsDataCache) buildPVClass(ctx context.Context,
	monitorType string, params, batchParams map[string][]string) {
	log.AddContext(ctx).Infoln("start to build pv class")
	if batchParams == nil {
		batchParams = make(map[string][]string)
	}

	metricsIndicators, ok := params["pv"]
	if !ok {
		return
	}

	err := metricsDataCache.buildPVBatchParams(ctx, monitorType, params, batchParams)
	if err != nil {
		log.AddContext(ctx).Errorf("build pv batch params failed, err is [%v]", err)
		return
	}

	mergeFunc, exist := mergeMetricsFactories["pv"]
	if !exist {
		log.AddContext(ctx).Errorf("can not get pv merge func")
		return
	}

	mergeDataType, err := mergeFunc(metricsDataCache.BackendName, monitorType, "pv", metricsIndicators)
	if err != nil {
		log.AddContext(ctx).Errorf("can not get pv mergeDataType, err is [%v]", err)
		return
	}
	metricsDataCache.MergeMetrics["pv"] = mergeDataType
	log.AddContext(ctx).Infof("build pv class success with batch params: %v", batchParams)
	return
}

func (metricsDataCache *MetricsDataCache) buildStorageClass(ctx context.Context,
	monitorType string, params, batchParams map[string][]string) {
	log.AddContext(ctx).Infoln("start to build storage class")
	if batchParams == nil {
		batchParams = make(map[string][]string)
	}
	for collectorName, metricsIndicators := range params {
		_, exist := batchParams[collectorName]
		if exist {
			continue
		}
		batchParams[collectorName] = metricsIndicators
	}

	for collectorName := range batchParams {
		metricsFunc, exist := metricsFactories[collectorName]
		if !exist {
			log.AddContext(ctx).Errorf("New metrics data error, the factories not have %s", collectorName)
			continue
		}
		metricsData, err := metricsFunc(metricsDataCache.BackendName, collectorName)
		if err != nil {
			log.AddContext(ctx).Errorf("New metrics data for %s, the monitorType : %s, error: %v",
				collectorName, monitorType, err)
			continue
		}
		metricsDataCache.CacheDataMap[collectorName] = metricsData
	}
	log.AddContext(ctx).Infof("build storage class success with batch params: %v", batchParams)
}

// BuildBatchDataClass get one batch cache class. use they to get metrics data
func (metricsDataCache *MetricsDataCache) BuildBatchDataClass(ctx context.Context,
	monitorType string, params map[string][]string) (map[string][]string, error) {
	batchParams := make(map[string][]string)

	metricsDataCache.buildPVClass(ctx, monitorType, params, batchParams)
	metricsDataCache.buildStorageClass(ctx, monitorType, params, batchParams)

	return batchParams, nil
}
