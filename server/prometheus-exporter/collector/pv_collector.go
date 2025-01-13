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
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
)

var pvBuildMap = map[string]collectorInitFunc{
	"object":      buildObjectPVCollector,
	"performance": buildPerformancePVCollector,
}

var pvLabelSlice = []string{"backend", "pv_name", "pvc_name", "object",
	"storage_volume_type", "storage_volume_id", "storage_volume_name"}

var pvObjectMetricsLabelMap = map[string][]string{
	"capacity":       pvLabelSlice,
	"capacity_usage": pvLabelSlice,
}

var pvObjectMetricsHelpMap = map[string]string{
	"capacity":       "Huawei Storage k8s PV Capacity(GB)",
	"capacity_usage": "Huawei Storage k8s PV Capacity Usage(%)",
}

var pvObjectMetricsParseMap = map[string]parseRelation{
	"capacity":       {"CAPACITY", parseStorageSectorsToGB},
	"capacity_usage": {"", parsePVCapacityUsage},
}

var pvTypePrometheusMetrics = map[string][]string{
	"lun":        {"lun_total_bandwidth", "lun_pv_lun_total_iops", "lun_avg_io_response_time"},
	"filesystem": {"filesystem_ops", "filesystem_avg_read_ops_response_time", "filesystem_avg_write_ops_response_time"},
}

var pvPrometheusMetricsLabelMap = map[string][]string{
	"lun_total_bandwidth":                    pvLabelSlice,
	"lun_pv_lun_total_iops":                  pvLabelSlice,
	"lun_avg_io_response_time":               pvLabelSlice,
	"filesystem_ops":                         pvLabelSlice,
	"filesystem_avg_read_ops_response_time":  pvLabelSlice,
	"filesystem_avg_write_ops_response_time": pvLabelSlice,
}

var pvPrometheusMetricsHelpMap = map[string]string{
	"lun_total_bandwidth":                    "Total Bandwidth(MB/s)",
	"lun_pv_lun_total_iops":                  "Total IOPS(IO/s)",
	"lun_avg_io_response_time":               "Avg IO Response Time(us)",
	"filesystem_ops":                         "OPS",
	"filesystem_avg_read_ops_response_time":  "Avg Read OPS Response Time(us)",
	"filesystem_avg_write_ops_response_time": "Avg Write OPS Response Time(us)",
}

var pvPrometheusMetricsParseMap = map[string]parseRelation{
	"lun_total_bandwidth":                    {"21", parseStorageData},
	"lun_pv_lun_total_iops":                  {"22", parseStorageData},
	"lun_avg_io_response_time":               {"370", parseStorageData},
	"filesystem_ops":                         {"182", parseStorageData},
	"filesystem_avg_read_ops_response_time":  {"524", parseStorageData},
	"filesystem_avg_write_ops_response_time": {"525", parseStorageData},
}

var pvLabelParseMap = map[string]parseRelation{
	"backend":             {"sbcName", parseStorageData},
	"pv_name":             {"pvName", parseStorageData},
	"pvc_name":            {"pvcName", parseStorageData},
	"storage_volume_type": {"sbcStorageType", parseStorageData},
	"storage_volume_id":   {"ID", parsePVStorageID},
	"storage_volume_name": {"sameName", parseStorageData},
	"object":              {"collectorName", parseStorageData},
}

func init() {
	RegisterCollector("pv", NewPVCollector)
}

type PVCollector struct {
	*BaseCollector
}

func parsePVStorageID(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	_, ok := inData["NAME"]
	if ok {
		return inData["ID"]
	}
	_, ok = inData["ObjectName"]
	if ok {
		return inData["ObjectId"]
	}
	return ""
}

func parsePVCapacityUsage(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	pvType, ok := inData["sbcStorageType"]
	if !ok {
		return ""
	}
	var pvCapacityUsage string
	if pvType == "oceanstor-san" {
		pvCapacityUsage = parseLunCapacityUsage(inDataKey, metricsName, inData)
	}
	if pvType == "oceanstor-nas" {
		pvCapacityUsage = parseFilesystemCapacityUsage(inDataKey, metricsName, inData)
	}
	return pvCapacityUsage
}

func buildObjectPVCollector(backendName string, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	return &PVCollector{
		BaseCollector: (&BaseCollector{}).SetBackendName(backendName).
			SetMonitorType(monitorType).
			SetCollectorName("pv").
			SetMetricsHelpMap(pvObjectMetricsHelpMap).
			SetMetricsLabelMap(pvObjectMetricsLabelMap).
			SetLabelParseMap(pvLabelParseMap).
			SetMetricsParseMap(pvObjectMetricsParseMap).
			SetMetricsDataCache(metricsDataCache).
			SetMetrics(make(map[string]*prometheus.Desc)),
	}, nil
}

func buildPerformancePVCollector(backendName string, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	if len(metricsIndicators) == 0 || metricsIndicators[0] == "" {
		return nil, fmt.Errorf("can not create [%s] collector, "+
			"the metricsIndicators is empty or error", "pv")
	}
	metricsData := strings.Split(metricsIndicators[0], ",")
	return &PVCollector{
		BaseCollector: (&BaseCollector{}).SetBackendName(backendName).
			SetMonitorType(monitorType).
			SetCollectorName("pv").
			SetMetricsHelpMap(pickPVPerformanceParsMap[string](metricsData, pvPrometheusMetricsHelpMap)).
			SetMetricsLabelMap(pickPVPerformanceParsMap[[]string](metricsData, pvPrometheusMetricsLabelMap)).
			SetLabelParseMap(pvLabelParseMap).
			SetMetricsParseMap(pickPVPerformanceParsMap[parseRelation](metricsData, pvPrometheusMetricsParseMap)).
			SetMetricsDataCache(metricsDataCache).
			SetMetrics(make(map[string]*prometheus.Desc)),
	}, nil
}

func pickPVPerformanceParsMap[T any](needPerformanceMetrics []string,
	allPerformanceParseMap map[string]T) map[string]T {
	var performanceParseMap = make(map[string]T)
	if len(needPerformanceMetrics) == 0 {
		return performanceParseMap
	}
	var allNeedMetricsSlice []string
	for _, pvType := range needPerformanceMetrics {
		needMetricsSlice, ok := pvTypePrometheusMetrics[pvType]
		if !ok {
			continue
		}
		allNeedMetricsSlice = append(allNeedMetricsSlice, needMetricsSlice...)
	}
	if len(allNeedMetricsSlice) == 0 {
		return performanceParseMap
	}

	for _, metricsName := range allNeedMetricsSlice {
		parseInfo, ok := allPerformanceParseMap[metricsName]
		if ok {
			performanceParseMap[metricsName] = parseInfo
		}
	}
	return performanceParseMap
}

func NewPVCollector(backendName, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	buildFunc, ok := pvBuildMap[monitorType]
	if !ok {
		return nil, fmt.Errorf("can not create filesystem collector, " +
			"the monitor type not in object or performance")
	}
	return buildFunc(backendName, monitorType, metricsIndicators, metricsDataCache)
}
