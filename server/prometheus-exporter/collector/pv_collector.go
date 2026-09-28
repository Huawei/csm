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
	"capacity":       {"CAPACITY", parsePVCapacity},
	"capacity_usage": {"", parsePVCapacityUsage},
}

var pvTypePrometheusMetrics = map[string][]string{
	"lun":        {"lun_total_bandwidth", "lun_pv_lun_total_iops", "lun_avg_io_response_time"},
	"filesystem": {"filesystem_ops", "filesystem_avg_read_ops_response_time", "filesystem_avg_write_ops_response_time"},
	"namespace": {
		"namespace_nfs_read_bandwidth", "namespace_nfs_write_bandwidth",
		"namespace_nfs_total_bandwidth", "namespace_nfs_read_ops",
		"namespace_nfs_write_ops", "namespace_nfs_total_ops",
		"namespace_nfs_avg_latency", "namespace_nfs_max_latency",
		"namespace_nfs_write_avg_latency", "namespace_nfs_write_max_latency",
		"namespace_nfs_read_avg_latency", "namespace_nfs_read_max_latency",
		"namespace_nfs_read_io_avg_size", "namespace_nfs_write_io_avg_size",
		"namespace_dpc_read_bandwidth", "namespace_dpc_write_bandwidth",
		"namespace_dpc_read_ops", "namespace_dpc_write_ops",
		"namespace_dpc_avg_read_latency", "namespace_dpc_avg_write_latency",
		"namespace_dpc_total_bandwidth", "namespace_dpc_total_ops",
		"namespace_dpc_avg_total_latency",
	},
}

var pvPrometheusMetricsLabelMap = map[string][]string{
	"lun_total_bandwidth":                    pvLabelSlice,
	"lun_pv_lun_total_iops":                  pvLabelSlice,
	"lun_avg_io_response_time":               pvLabelSlice,
	"filesystem_ops":                         pvLabelSlice,
	"filesystem_avg_read_ops_response_time":  pvLabelSlice,
	"filesystem_avg_write_ops_response_time": pvLabelSlice,
	"namespace_nfs_read_bandwidth":           pvLabelSlice,
	"namespace_nfs_write_bandwidth":          pvLabelSlice,
	"namespace_nfs_total_bandwidth":          pvLabelSlice,
	"namespace_nfs_read_ops":                 pvLabelSlice,
	"namespace_nfs_write_ops":                pvLabelSlice,
	"namespace_nfs_total_ops":                pvLabelSlice,
	"namespace_nfs_avg_latency":              pvLabelSlice,
	"namespace_nfs_max_latency":              pvLabelSlice,
	"namespace_nfs_write_avg_latency":        pvLabelSlice,
	"namespace_nfs_write_max_latency":        pvLabelSlice,
	"namespace_nfs_read_avg_latency":         pvLabelSlice,
	"namespace_nfs_read_max_latency":         pvLabelSlice,
	"namespace_nfs_read_io_avg_size":         pvLabelSlice,
	"namespace_nfs_write_io_avg_size":        pvLabelSlice,
	"namespace_dpc_read_bandwidth":           pvLabelSlice,
	"namespace_dpc_write_bandwidth":          pvLabelSlice,
	"namespace_dpc_read_ops":                 pvLabelSlice,
	"namespace_dpc_write_ops":                pvLabelSlice,
	"namespace_dpc_avg_read_latency":         pvLabelSlice,
	"namespace_dpc_avg_write_latency":        pvLabelSlice,
	"namespace_dpc_total_bandwidth":          pvLabelSlice,
	"namespace_dpc_total_ops":                pvLabelSlice,
	"namespace_dpc_avg_total_latency":        pvLabelSlice,
}

var pvPrometheusMetricsHelpMap = map[string]string{
	"lun_total_bandwidth":                    "Total Bandwidth(MB/s)",
	"lun_pv_lun_total_iops":                  "Total IOPS(IO/s)",
	"lun_avg_io_response_time":               "Avg IO Response Time(us)",
	"filesystem_ops":                         "OPS",
	"filesystem_avg_read_ops_response_time":  "Avg Read OPS Response Time(us)",
	"filesystem_avg_write_ops_response_time": "Avg Write OPS Response Time(us)",
	"namespace_nfs_read_bandwidth":           "Namespace NFS Read Bandwidth(KB/s)",
	"namespace_nfs_write_bandwidth":          "Namespace NFS Write Bandwidth(KB/s)",
	"namespace_nfs_total_bandwidth":          "Namespace NFS Total Bandwidth(KB/s)",
	"namespace_nfs_read_ops":                 "Namespace NFS Read OPS",
	"namespace_nfs_write_ops":                "Namespace NFS Write OPS",
	"namespace_nfs_total_ops":                "Namespace NFS Total OPS",
	"namespace_nfs_avg_latency":              "Namespace NFS Avg Latency(us)",
	"namespace_nfs_max_latency":              "Namespace NFS Max Latency(us)",
	"namespace_nfs_write_avg_latency":        "Namespace NFS Write Avg Latency(us)",
	"namespace_nfs_write_max_latency":        "Namespace NFS Write Max Latency(us)",
	"namespace_nfs_read_avg_latency":         "Namespace NFS Read Avg Latency(us)",
	"namespace_nfs_read_max_latency":         "Namespace NFS Read Max Latency(us)",
	"namespace_nfs_read_io_avg_size":         "Namespace NFS Read IO Avg Size(KB)",
	"namespace_nfs_write_io_avg_size":        "Namespace NFS Write IO Avg Size(KB)",
	"namespace_dpc_read_bandwidth":           "Namespace DPC Read Bandwidth(MB/s)",
	"namespace_dpc_write_bandwidth":          "Namespace DPC Write Bandwidth(MB/s)",
	"namespace_dpc_read_ops":                 "Namespace DPC Read OPS",
	"namespace_dpc_write_ops":                "Namespace DPC Write OPS",
	"namespace_dpc_avg_read_latency":         "Namespace DPC Avg Read Latency(ms)",
	"namespace_dpc_avg_write_latency":        "Namespace DPC Avg Write Latency(ms)",
	"namespace_dpc_total_bandwidth":          "Namespace DPC Total Bandwidth(MB/s)",
	"namespace_dpc_total_ops":                "Namespace DPC Total OPS",
	"namespace_dpc_avg_total_latency":        "Namespace DPC Avg Total Latency(ms)",
}

var pvPrometheusMetricsParseMap = map[string]parseRelation{
	"lun_total_bandwidth":                    {"21", parsePVLunData},
	"lun_pv_lun_total_iops":                  {"22", parsePVLunData},
	"lun_avg_io_response_time":               {"370", parsePVLunData},
	"filesystem_ops":                         {"182", parsePVFSData},
	"filesystem_avg_read_ops_response_time":  {"524", parsePVFSData},
	"filesystem_avg_write_ops_response_time": {"525", parsePVFSData},
	"namespace_nfs_read_bandwidth":           {"30001", parseNamespaceData},
	"namespace_nfs_write_bandwidth":          {"30002", parseNamespaceData},
	"namespace_nfs_total_bandwidth":          {"30003", parseNamespaceData},
	"namespace_nfs_read_ops":                 {"30004", parseNamespaceData},
	"namespace_nfs_write_ops":                {"30005", parseNamespaceData},
	"namespace_nfs_total_ops":                {"30006", parseNamespaceData},
	"namespace_nfs_avg_latency":              {"30011", parseNamespaceData},
	"namespace_nfs_max_latency":              {"30012", parseNamespaceData},
	"namespace_nfs_write_avg_latency":        {"30013", parseNamespaceData},
	"namespace_nfs_write_max_latency":        {"30014", parseNamespaceData},
	"namespace_nfs_read_avg_latency":         {"30015", parseNamespaceData},
	"namespace_nfs_read_max_latency":         {"30016", parseNamespaceData},
	"namespace_nfs_read_io_avg_size":         {"30076", parseNamespaceData},
	"namespace_nfs_write_io_avg_size":        {"30077", parseNamespaceData},
	"namespace_dpc_read_bandwidth":           {"30043", parseNamespaceData},
	"namespace_dpc_write_bandwidth":          {"30044", parseNamespaceData},
	"namespace_dpc_read_ops":                 {"30045", parseNamespaceData},
	"namespace_dpc_write_ops":                {"30046", parseNamespaceData},
	"namespace_dpc_avg_read_latency":         {"30048", parseNamespaceData},
	"namespace_dpc_avg_write_latency":        {"30049", parseNamespaceData},
	"namespace_dpc_total_bandwidth":          {"30051", parseNamespaceData},
	"namespace_dpc_total_ops":                {"30052", parseNamespaceData},
	"namespace_dpc_avg_total_latency":        {"31005", parseNamespaceData},
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

func parsePVFSData(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}

	pvType, ok := inData[storageTypeKey]
	if !ok {
		return ""
	}

	if pvType != storageTypeNas {
		return skipReportValue
	}

	return inData[inDataKey]
}

func parsePVLunData(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}

	pvType, ok := inData[storageTypeKey]
	if !ok {
		return ""
	}

	if pvType != storageTypeSan {
		return skipReportValue
	}

	return inData[inDataKey]
}

func parseNamespaceData(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}

	pvType, ok := inData[storageTypeKey]
	if !ok {
		return ""
	}

	if pvType != storageTypeFusionNas {
		return skipReportValue
	}

	return inData[inDataKey]
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
	// Defensive fallback for lowercase id (DistributedClient normalizes to uppercase,
	// but this handles any edge case where normalization may not apply)
	if id, ok := inData["id"]; ok {
		return id
	}
	return ""
}

func parsePVCapacityUsage(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	pvType, ok := inData[storageTypeKey]
	if !ok {
		return ""
	}
	var pvCapacityUsage string
	if pvType == storageTypeSan {
		pvCapacityUsage = parseLunCapacityUsage(inDataKey, metricsName, inData)
	}
	if pvType == storageTypeNas {
		pvCapacityUsage = parseFilesystemCapacityUsage(inDataKey, metricsName, inData)
	}
	if pvType == storageTypeFusionNas {
		pvCapacityUsage = inData["SPACE_USED_RATE"]
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
