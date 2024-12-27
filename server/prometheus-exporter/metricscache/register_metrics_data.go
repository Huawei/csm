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

	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
)

// a MetricsData constructor
type metricsCacheInitFunc = func(backendName, metricsType string) (MetricsData, error)

// metricsFactories are routing table with MetricsData factory routing
// key is metrics name
// value is a MetricsData constructor
// e.g.
// |---------------|-------------------------|
// | metricsName   | collectorInitFunc       |
// |---------------|-------------------------|
// | array         | NewStorageMetricsCache  |
// |---------------|-------------------------|
var metricsFactories = make(map[string]metricsCacheInitFunc)

// RegisterMetricsData register a MetricsData constructor to factories
func RegisterMetricsData(collectorName string, factory metricsCacheInitFunc) {
	metricsFactories[collectorName] = factory
}

// MetricsData set metrics data, get data from storage and kubernetes
type MetricsData interface {
	GetMetricsDataResponse() *storageGRPC.CollectResponse
	SetMetricsData(ctx context.Context, collectorName, monitorType string, metricsIndicators []string) error
}

// BaseMetricsData save one batch data with special MetricsType
type BaseMetricsData struct {
	BackendName         string
	MetricsType         string
	MetricsDataResponse *storageGRPC.CollectResponse
}

// GetMetricsDataResponse implement MetricsData interface, get MetricsDataResponse
func (baseMetricsData *BaseMetricsData) GetMetricsDataResponse() *storageGRPC.CollectResponse {
	return baseMetricsData.MetricsDataResponse
}

// SetMetricsData implement MetricsData interface, set MetricsData from data source
func (baseMetricsData *BaseMetricsData) SetMetricsData(
	ctx context.Context, collectorName, monitorType string, metricsIndicators []string) error {
	return nil
}
