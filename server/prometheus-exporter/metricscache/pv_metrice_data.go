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

	"github.com/huawei/csm/v2/utils/log"
)

// MetricsPVData save one batch data with special MetricsType,
// from prometheus request
type MetricsPVData struct {
	*BaseMetricsData
}

func init() {
	RegisterMetricsData("pv", NewMetricsPVData)
}

// NewMetricsPVData creates a new MetricsPVData with special MetricsType
func NewMetricsPVData(backendName, metricsType string) (MetricsData, error) {
	return &MetricsPVData{BaseMetricsData: &BaseMetricsData{
		BackendName: backendName, MetricsType: metricsType}}, nil
}

// SetMetricsData set pv data MetricsDataResponse
func (metricsData *MetricsPVData) SetMetricsData(ctx context.Context,
	collectorName, monitorType string, metricsIndicators []string) error {
	log.AddContext(ctx).Infof("start to get pv metrics data with collector name: %v, monitor type: %v, "+
		"indicators: %v", collectorName, monitorType, metricsIndicators)

	// get PV data
	batchCollectResponse, err := GetAndParsePVInfo(ctx, metricsData.BackendName, metricsData.MetricsType)
	if err != nil {
		return fmt.Errorf("get pv info failed, err is [%w]", err)
	}

	metricsData.MetricsDataResponse = batchCollectResponse
	log.AddContext(ctx).Infoln("get pv metrics data success")
	return nil
}
