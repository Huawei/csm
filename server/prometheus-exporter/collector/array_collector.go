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

// Package collector includes all huawei storage collectors to gather and export huawei storage metrics.
package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
)

func init() {
	RegisterCollector("array", NewArrayCollector)
}

const (
	productModeString = "productModeString"
	productMode       = "PRODUCTMODE"
	productVersion    = "PRODUCTVERSION"
	softwareVersion   = "SoftwareVersion"
)

var arrayObjectMetricsLabelMap = map[string][]string{
	"basic_info":     {"endpoint", "sn", "model", "version", "object"},
	"health_status":  {"endpoint", "sn", "status", "object"},
	"running_status": {"endpoint", "sn", "status", "object"},
}
var arrayObjectMetricsHelpMap = map[string]string{
	"basic_info":     "Huawei Storage Array Basic Info",
	"health_status":  "Huawei Storage Array Health Status",
	"running_status": "Huawei Storage Array Running Status",
}

var arrayObjectMetricsParseMap = map[string]parseRelation{
	"basic_info":     {"", parseStorageReturnZero},
	"health_status":  {"HEALTHSTATUS", parseStorageData},
	"running_status": {"RUNNINGSTATUS", parseStorageData},
}
var arrayObjectLabelParseMap = map[string]parseRelation{
	"endpoint": {"backendName", parseStorageData},
	"sn":       {"ID", parseStorageData},
	"model":    {"", parseArrayModel},
	"version":  {"", parseVersion},
	"status":   {"", parseStorageStatus},
	"object":   {"collectorName", parseStorageData},
}

func parseArrayModel(inDataKey, metricsName string, inData map[string]string) string {
	var modelName string
	if inData[productModeString] != "" {
		modelName = inData[productModeString]
	} else if inData[productMode] != "" {
		modelName = StorageProductMode[inData[productMode]]
	}
	return modelName
}

func parseVersion(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}

	var version string
	if inData[softwareVersion] != "" {
		// v7 storage production version field
		version = inData[softwareVersion]
	} else if inData[productVersion] != "" {
		// v6 or earlier storage production version field
		version = inData[productVersion]
	}

	return version
}

// ArrayCollector implements the prometheus.Collector interface and build storage array info
type ArrayCollector struct {
	*BaseCollector
}

func NewArrayCollector(backendName string, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	if monitorType == "object" {
		return &ArrayCollector{
			BaseCollector: (&BaseCollector{}).SetBackendName(backendName).
				SetMonitorType(monitorType).
				SetCollectorName("array").
				SetMetricsHelpMap(arrayObjectMetricsHelpMap).
				SetMetricsLabelMap(arrayObjectMetricsLabelMap).
				SetLabelParseMap(arrayObjectLabelParseMap).
				SetMetricsParseMap(arrayObjectMetricsParseMap).
				SetMetricsDataCache(metricsDataCache).
				SetMetrics(make(map[string]*prometheus.Desc)),
		}, nil
	}

	return nil, fmt.Errorf("can not create array collector, the monitor type is not object")
}
