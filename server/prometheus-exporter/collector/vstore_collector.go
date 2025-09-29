/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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
	"errors"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
)

func init() {
	RegisterCollector("vstore", NewVstoreCollector)
}

const (
	vstoreTotalCapacityKey = "TotalCapacity"
	vstoreUsedCapacityKey  = "UsedCapacity"
)

var vstoreLabelSlice = []string{"name", "endpoint", "id", "object", "pool"}

var vstorePrometheusDescName = map[string]string{
	"object": "vstore",
}

var vstoreObjectMetricsLabelMap = map[string][]string{
	"total_capacity": vstoreLabelSlice,
	"free_capacity":  vstoreLabelSlice,
	"used_capacity":  vstoreLabelSlice,
	"capacity_usage": vstoreLabelSlice,
}
var vstoreObjectMetricsHelpMap = map[string]string{
	"total_capacity": "vstore capacity(GB)",
	"free_capacity":  "vstore free capacity(GB)",
	"used_capacity":  "vstore used capacity(GB)",
	"capacity_usage": "vstore capacity usage(%)",
}
var vstoreObjectMetricsParseMap = map[string]parseRelation{
	"total_capacity": {"TotalCapacity", parseVstoreCapacityToGB},
	"free_capacity":  {"FreeCapacity", parseVstoreCapacityToGB},
	"used_capacity":  {"UsedCapacity", parseVstoreCapacityToGB},
	"capacity_usage": {"", parseVstoreCapacityUsage},
}
var vstoreObjectLabelParseMap = map[string]parseRelation{
	"endpoint": {"BackendName", parseStorageData},
	"id":       {"VStoreID", parseStorageData},
	"name":     {"VStoreName", parseStorageData},
	"object":   {"collectorName", parseStorageData},
	"pool":     {"PoolName", parseStorageData},
}

// VstoreCollector implements the prometheus.Collector interface and build storage Vstore info
type VstoreCollector struct {
	*BaseCollector
}

func parseVstoreCapacityUsage(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}

	capacity, err := strconv.ParseFloat(inData[vstoreTotalCapacityKey], bitSize)
	if err != nil || capacity == 0 {
		return ""
	}

	usedCapacity, err := strconv.ParseFloat(inData[vstoreUsedCapacityKey], bitSize)
	if err != nil {
		return ""
	}

	return strconv.FormatFloat(usedCapacity/capacity*calculatePercentage, 'f', unlimitedPrecision, bitSize)
}

// Describe implements the prometheus.Collector interface.
// Use BuildDesc to build vstore Desc then send to prometheus.
func (VstoreCollector *VstoreCollector) Describe(ch chan<- *prometheus.Desc) {
	VstoreCollector.BuildDesc()
	for _, i := range VstoreCollector.metrics {
		ch <- i
	}
}

// BuildDesc use VstoreCollector.metricsDescMap create vstore Collector prometheus.Desc
func (VstoreCollector *VstoreCollector) BuildDesc() {
	if VstoreCollector.metrics == nil {
		VstoreCollector.metrics = make(map[string]*prometheus.Desc)
	}
	vstoreDes, ok := vstorePrometheusDescName[VstoreCollector.monitorType]
	if !ok {
		return
	}
	for metricsName, helpInfo := range VstoreCollector.metricsHelpMap {
		VstoreCollector.metrics[metricsName] =
			prometheus.NewDesc(
				prometheus.BuildFQName(
					MetricsNamespace, vstoreDes, metricsName),
				helpInfo,
				VstoreCollector.metricsLabelMap[metricsName],
				nil)
	}
}

// NewVstoreCollector create and init a VstoreCollector.
func NewVstoreCollector(backendName, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	if monitorType == "object" {
		return &VstoreCollector{
			BaseCollector: (&BaseCollector{}).SetBackendName(backendName).
				SetMonitorType(monitorType).
				SetCollectorName("vstore").
				SetMetricsHelpMap(vstoreObjectMetricsHelpMap).
				SetMetricsLabelMap(vstoreObjectMetricsLabelMap).
				SetLabelParseMap(vstoreObjectLabelParseMap).
				SetMetricsParseMap(vstoreObjectMetricsParseMap).
				SetMetricsDataCache(metricsDataCache).
				SetMetrics(make(map[string]*prometheus.Desc)),
		}, nil
	}

	return nil, errors.New("can not create vstore collector, " +
		"the monitor type not in object")
}
