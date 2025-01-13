/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
)

func init() {
	RegisterCollector("storagepool", NewStoragePoolCollector)
}

const (
	storagePoolCapacityKey     = "USERTOTALCAPACITY"
	storagePoolUsedCapacityKey = "USERCONSUMEDCAPACITY"
)

var storagePoolPrometheusDescName = map[string]string{
	"object":      "storage_pool",
	"performance": "storagepool",
}

var storagePoolObjectMetricsLabelMap = map[string][]string{
	"total_capacity": {"name", "endpoint", "id", "object"},
	"free_capacity":  {"name", "endpoint", "id", "object"},
	"capacity_usage": {"name", "endpoint", "id", "object"},
	"used_capacity":  {"name", "endpoint", "id", "object"},
}
var storagePoolObjectMetricsHelpMap = map[string]string{
	"total_capacity": "Total capacity(GB) of storage pool",
	"free_capacity":  "Free capacity(GB) of storage pool",
	"capacity_usage": "Used capacity ratio(%) of storage pool",
	"used_capacity":  "Used capacity(GB) of storage pool",
}

var storagePoolObjectMetricsParseMap = map[string]parseRelation{
	"total_capacity": {"USERTOTALCAPACITY", parseStorageSectorsToGB},
	"free_capacity":  {"USERFREECAPACITY", parseStorageSectorsToGB},
	"capacity_usage": {"", parseStoragePoolCapacityUsage},
	"used_capacity":  {"USERCONSUMEDCAPACITY", parseStorageSectorsToGB},
}
var storagePoolObjectLabelParseMap = map[string]parseRelation{
	"endpoint": {"backendName", parseStorageData},
	"id":       {"ID", parseStorageData},
	"name":     {"NAME", parseStorageData},
	"object":   {"collectorName", parseStorageData},
}

// StoragePoolCollector implements the prometheus.Collector interface and build storage StoragePool info
type StoragePoolCollector struct {
	*BaseCollector
}

func parseStoragePoolCapacityUsage(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	capacity, err := strconv.ParseFloat(inData[storagePoolCapacityKey], bitSize)
	if err != nil || capacity == 0 {
		return ""
	}
	usedCapacity, err := strconv.ParseFloat(inData[storagePoolUsedCapacityKey], bitSize)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(usedCapacity/capacity*calculatePercentage, 'f', unlimitedPrecision, bitSize)
}

// Describe implements the prometheus.Collector interface.
// Use BuildDesc to build storage pool Desc then send to prometheus.
func (storagePoolCollector *StoragePoolCollector) Describe(ch chan<- *prometheus.Desc) {
	storagePoolCollector.BuildDesc()
	for _, i := range storagePoolCollector.metrics {
		ch <- i
	}
}

// BuildDesc use StoragePoolCollector.metricsDescMap create storage poo Collector prometheus.Desc
func (storagePoolCollector *StoragePoolCollector) BuildDesc() {
	if storagePoolCollector.metrics == nil {
		storagePoolCollector.metrics = make(map[string]*prometheus.Desc)
	}
	storagePoolDes, ok := storagePoolPrometheusDescName[storagePoolCollector.monitorType]
	if !ok {
		return
	}
	for metricsName, helpInfo := range storagePoolCollector.metricsHelpMap {
		storagePoolCollector.metrics[metricsName] =
			prometheus.NewDesc(
				prometheus.BuildFQName(
					MetricsNamespace, storagePoolDes, metricsName),
				helpInfo,
				storagePoolCollector.metricsLabelMap[metricsName],
				nil)
	}
}

func NewStoragePoolCollector(backendName, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	if monitorType == "object" {
		return &StoragePoolCollector{
			BaseCollector: (&BaseCollector{}).SetBackendName(backendName).
				SetMonitorType(monitorType).
				SetCollectorName("storagepool").
				SetMetricsHelpMap(storagePoolObjectMetricsHelpMap).
				SetMetricsLabelMap(storagePoolObjectMetricsLabelMap).
				SetLabelParseMap(storagePoolObjectLabelParseMap).
				SetMetricsParseMap(storagePoolObjectMetricsParseMap).
				SetMetricsDataCache(metricsDataCache).
				SetMetrics(make(map[string]*prometheus.Desc)),
		}, nil
	} else if monitorType == "performance" {
		performanceBaseCollector, err := NewPerformanceBaseCollector(
			backendName, monitorType, "storagepool", metricsIndicators, metricsDataCache)
		if err != nil {
			return nil, err
		}
		return &StoragePoolCollector{
			BaseCollector: performanceBaseCollector,
		}, nil
	}

	return nil, fmt.Errorf("can not create storage pool collector, " +
		"the monitor type not in object or performance")
}
