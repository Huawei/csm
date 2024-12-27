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
	RegisterCollector("lun", NewLunCollector)
}

var lunBuildMap = map[string]collectorInitFunc{
	"object":      buildObjectLunCollector,
	"performance": buildPerformanceLunCollector,
}

var lunObjectMetricsLabelMap = map[string][]string{
	"capacity":       {"endpoint", "id", "name", "object"},
	"capacity_usage": {"endpoint", "id", "name", "object"},
}
var lunObjectMetricsHelpMap = map[string]string{
	"capacity":       "LUN capacity(GB)",
	"capacity_usage": "LUN capacity usage(%)",
}

var lunObjectMetricsParseMap = map[string]parseRelation{
	"capacity":       {"CAPACITY", parseStorageSectorsToGB},
	"capacity_usage": {"", parseCapacityUsage},
}
var lunObjectLabelParseMap = map[string]parseRelation{
	"endpoint": {"backendName", parseStorageData},
	"id":       {"ID", parseStorageData},
	"name":     {"NAME", parseStorageData},
	"object":   {"collectorName", parseStorageData},
}

type LunCollector struct {
	*BaseCollector
}

func buildObjectLunCollector(backendName string, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	return &LunCollector{
		BaseCollector: (&BaseCollector{}).SetBackendName(backendName).
			SetMonitorType(monitorType).
			SetCollectorName("lun").
			SetMetricsHelpMap(lunObjectMetricsHelpMap).
			SetMetricsLabelMap(lunObjectMetricsLabelMap).
			SetLabelParseMap(lunObjectLabelParseMap).
			SetMetricsParseMap(lunObjectMetricsParseMap).
			SetMetricsDataCache(metricsDataCache).
			SetMetrics(make(map[string]*prometheus.Desc)),
	}, nil
}

func buildPerformanceLunCollector(backendName string, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	performanceBaseCollector, err := NewPerformanceBaseCollector(
		backendName, monitorType, "lun", metricsIndicators, metricsDataCache)
	if err != nil {
		return nil, err
	}
	return &LunCollector{
		BaseCollector: performanceBaseCollector,
	}, nil
}

func NewLunCollector(backendName, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	buildFunc, ok := lunBuildMap[monitorType]
	if !ok {
		return nil, fmt.Errorf("can not create filesystem collector, " +
			"the monitor type not in object or performance")
	}
	return buildFunc(backendName, monitorType, metricsIndicators, metricsDataCache)
}
