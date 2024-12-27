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

var performanceMetricsIndicatorsMap = map[string]string{
	"22":  "total_iops",
	"25":  "read_iops",
	"28":  "write_iops",
	"21":  "total_bandwidth",
	"23":  "read_bandwidth",
	"26":  "write_bandwidth",
	"370": "avg_io_response_time",
	"182": "ops",
	"524": "avg_read_ops_response_time",
	"525": "avg_write_ops_response_time",
}

var performanceMetricsLabelMap = map[string][]string{
	"total_iops":                  {"endpoint", "id", "object", "name"},
	"read_iops":                   {"endpoint", "id", "object", "name"},
	"write_iops":                  {"endpoint", "id", "object", "name"},
	"total_bandwidth":             {"endpoint", "id", "object", "name"},
	"read_bandwidth":              {"endpoint", "id", "object", "name"},
	"write_bandwidth":             {"endpoint", "id", "object", "name"},
	"avg_io_response_time":        {"endpoint", "id", "object", "name"},
	"ops":                         {"endpoint", "id", "object", "name"},
	"avg_read_ops_response_time":  {"endpoint", "id", "object", "name"},
	"avg_write_ops_response_time": {"endpoint", "id", "object", "name"},
}
var performanceMetricsHelpMap = map[string]string{
	"total_iops":                  "Total IOPS(IO/s)",
	"read_iops":                   "Read IOPS(IO/s)",
	"write_iops":                  "Write IOPS(IO/s)",
	"total_bandwidth":             "Total Bandwidth(MB/s)",
	"read_bandwidth":              "Read Bandwidth(MB/s)",
	"write_bandwidth":             "Write Bandwidth(MB/s)",
	"avg_io_response_time":        "Avg IO Response Time(us)",
	"ops":                         "OPS",
	"avg_read_ops_response_time":  "Avg Read OPS Response Time(us)",
	"avg_write_ops_response_time": "Avg Write OPS Response Time(us)",
}

var performanceMetricsParseMap = map[string]parseRelation{
	"total_iops":                  {"22", parseStorageData},
	"read_iops":                   {"25", parseStorageData},
	"write_iops":                  {"28", parseStorageData},
	"total_bandwidth":             {"21", parseStorageData},
	"read_bandwidth":              {"23", parseStorageData},
	"write_bandwidth":             {"26", parseStorageData},
	"avg_io_response_time":        {"370", parseStorageData},
	"ops":                         {"182", parseStorageData},
	"avg_read_ops_response_time":  {"524", parseStorageData},
	"avg_write_ops_response_time": {"525", parseStorageData},
}
var performanceLabelParseMap = map[string]parseRelation{
	"endpoint": {"backendName", parseStorageData},
	"id":       {"ObjectId", parseStorageData},
	"name":     {"ObjectName", parseStorageData},
	"object":   {"collectorName", parseStorageData},
}

func pickPerformanceParsMap[T any](needPerformanceMetrics []string, allPerformanceParseMap map[string]T) map[string]T {
	var performanceParseMap = make(map[string]T)
	if len(needPerformanceMetrics) == 0 {
		return performanceParseMap
	}
	for _, indicatorName := range needPerformanceMetrics {
		metricsName, ok := performanceMetricsIndicatorsMap[indicatorName]
		if !ok {
			continue
		}

		parseInfo, ok := allPerformanceParseMap[metricsName]
		if ok {
			performanceParseMap[metricsName] = parseInfo
		}
	}
	return performanceParseMap
}
