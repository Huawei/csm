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
	RegisterCollector("controller", NewControllerCollector)
}

var controllerObjectMetricsLabelMap = map[string][]string{
	"cpu_usage":      {"endpoint", "id", "name", "object"},
	"memory_usage":   {"endpoint", "id", "name", "object"},
	"health_status":  {"endpoint", "id", "status", "name", "object"},
	"running_status": {"endpoint", "id", "status", "name", "object"},
}
var controllerObjectMetricsHelpMap = map[string]string{
	"cpu_usage":      "CPU utilization(%)",
	"memory_usage":   "Memory utilization(%)",
	"health_status":  "Health Status",
	"running_status": "Running Status",
}

var controllerObjectMetricsParseMap = map[string]parseRelation{
	"cpu_usage":      {"CPUUSAGE", parseStorageData},
	"memory_usage":   {"MEMORYUSAGE", parseStorageData},
	"health_status":  {"HEALTHSTATUS", parseStorageData},
	"running_status": {"RUNNINGSTATUS", parseStorageData},
}
var controllerObjectLabelParseMap = map[string]parseRelation{
	"endpoint": {"backendName", parseStorageData},
	"id":       {"ID", parseStorageData},
	"name":     {"NAME", parseStorageData},
	"status":   {"", parseStorageStatus},
	"object":   {"collectorName", parseStorageData},
}

// ControllerCollector implements the prometheus.Collector interface and build storage Controller info
type ControllerCollector struct {
	*BaseCollector
}

func NewControllerCollector(backendName, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	if monitorType == "object" {
		return &ControllerCollector{
			BaseCollector: (&BaseCollector{}).SetBackendName(backendName).
				SetMonitorType(monitorType).
				SetCollectorName("controller").
				SetMetricsHelpMap(controllerObjectMetricsHelpMap).
				SetMetricsLabelMap(controllerObjectMetricsLabelMap).
				SetLabelParseMap(controllerObjectLabelParseMap).
				SetMetricsParseMap(controllerObjectMetricsParseMap).
				SetMetricsDataCache(metricsDataCache).
				SetMetrics(make(map[string]*prometheus.Desc)),
		}, nil
	} else if monitorType == "performance" {
		performanceBaseCollector, err := NewPerformanceBaseCollector(
			backendName, monitorType, "controller", metricsIndicators, metricsDataCache)
		if err != nil {
			return nil, err
		}
		return &ControllerCollector{
			BaseCollector: performanceBaseCollector,
		}, nil
	}

	return nil, fmt.Errorf("can not create controller collector, " +
		"the monitor type not in object or performance")
}
