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
)

// a MergeMetricsData constructor
type mergeMetricsInitFunc = func(backendName, monitorType, metricsType string,
	metricsIndicators []string) (MergeMetricsData, error)

// mergeMetricsFactories are routing table with MergeMetricsData factory routing
// key is metrics name
// value is a MergeMetricsData constructor
// e.g.
// |---------------|-------------------------|
// | metricsName   | collectorInitFunc       |
// |---------------|-------------------------|
// | pv            | NewPVMergeMetricsCache  |
// |---------------|-------------------------|
var mergeMetricsFactories = make(map[string]mergeMetricsInitFunc)

// RegisterMergeMetricsData register a MergeMetricsData constructor to factories
func RegisterMergeMetricsData(collectorName string, factory mergeMetricsInitFunc) {
	mergeMetricsFactories[collectorName] = factory
}

// MergeMetricsData is the interface for need Merge Metrics like pv
type MergeMetricsData interface {
	MergeData(ctx context.Context, metricsDataCache *MetricsDataCache) error
}

// BaseMergeMetricsData is the base for need Merge Metrics
type BaseMergeMetricsData struct {
	backendName     string
	monitorType     string
	metricsType     string
	mergeIndicators []string
}
