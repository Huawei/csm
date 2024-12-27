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

// Package collector includes all huawei storage collectors to gather and export huawei storage metrics.
package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
)

func init() {
	RegisterCollector("filesystem", NewFilesystemCollector)
}

var filesystemBuildMap = map[string]collectorInitFunc{
	"object":      buildObjectFilesystemCollector,
	"performance": buildPerformanceFilesystemCollector,
}

var filesystemObjectMetricsLabelMap = map[string][]string{
	"capacity":       {"endpoint", "id", "name", "object"},
	"capacity_usage": {"endpoint", "id", "name", "object"},
}
var filesystemObjectMetricsHelpMap = map[string]string{
	"capacity":       "filesystem capacity(GB)",
	"capacity_usage": "filesystem capacity usage(%)",
}

var filesystemObjectMetricsParseMap = map[string]parseRelation{
	"capacity":       {"CAPACITY", parseStorageSectorsToGB},
	"capacity_usage": {"", parseCapacityUsage},
}
var filesystemObjectLabelParseMap = map[string]parseRelation{
	"endpoint": {"backendName", parseStorageData},
	"id":       {"ID", parseStorageData},
	"name":     {"NAME", parseStorageData},
	"object":   {"collectorName", parseStorageData},
}

type FilesystemCollector struct {
	*BaseCollector
}

func buildObjectFilesystemCollector(backendName string, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	return &FilesystemCollector{
		BaseCollector: (&BaseCollector{}).SetBackendName(backendName).
			SetMonitorType(monitorType).
			SetCollectorName("filesystem").
			SetMetricsHelpMap(filesystemObjectMetricsHelpMap).
			SetMetricsLabelMap(filesystemObjectMetricsLabelMap).
			SetLabelParseMap(filesystemObjectLabelParseMap).
			SetMetricsParseMap(filesystemObjectMetricsParseMap).
			SetMetricsDataCache(metricsDataCache).
			SetMetrics(make(map[string]*prometheus.Desc)),
	}, nil
}

func buildPerformanceFilesystemCollector(backendName string, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	performanceBaseCollector, err := NewPerformanceBaseCollector(
		backendName, monitorType, "filesystem", metricsIndicators, metricsDataCache)
	if err != nil {
		return nil, err
	}
	return &FilesystemCollector{
		BaseCollector: performanceBaseCollector,
	}, nil
}

func NewFilesystemCollector(backendName string, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	buildFunc, ok := filesystemBuildMap[monitorType]
	if !ok {
		return nil, fmt.Errorf("can not create filesystem collector, " +
			"the monitor type not in object or performance")
	}
	return buildFunc(backendName, monitorType, metricsIndicators, metricsDataCache)
}
