/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

// Package defs provides the defines of the monitor metrics need to be collected
package defs

import (
	"github.com/huawei/csm/v2/provider/collectdef"
	"github.com/huawei/csm/v2/provider/constants"
	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
)

// VstoreDef declaratively defines vstore's collection metadata.
// Vstore uses CollectModeNone because it reads from K8s CRs, not the storage REST API.
// A custom MetricsDataFactory wraps the existing VstoreMetricsData.
var VstoreDef = &collectdef.ObjectDef{
	CollectType:        "vstore",
	SupportObject:      true,
	SupportPerformance: false,
	CollectMode:        collectdef.CollectModeNone,
	ObjectMetrics: []collectdef.MetricDef{
		{
			Name: "total_capacity", Help: "vstore capacity(GB)",
			Transform: collectdef.VstoreBytesToGBOf("TotalCapacity"),
		},
		{
			Name: "free_capacity", Help: "vstore free capacity(GB)",
			Transform: collectdef.VstoreBytesToGBOf("FreeCapacity"),
		},
		{
			Name: "used_capacity", Help: "vstore used capacity(GB)",
			Transform: collectdef.VstoreBytesToGBOf("UsedCapacity"),
		},
		{
			Name: "capacity_usage", Help: "vstore capacity usage(%)",
			Transform: collectdef.VstoreCapacityUsage,
		},
	},
	ObjectLabels: []collectdef.LabelDef{
		{Name: "endpoint", SourceKey: "BackendName"},
		{Name: "id", SourceKey: "VStoreID"},
		{Name: "name", SourceKey: "VStoreName"},
		{Name: "object", SourceKey: "collectorName"},
		{Name: "pool", SourceKey: "PoolName"},
	},
	MetricsDataFactory: func(backendName, metricsType string) (interface{}, error) {
		return metricsCache.NewMetricsVstoreData(backendName, metricsType)
	},
}

func init() {
	collectdef.MustRegister(constants.OceanStorage, VstoreDef)
}
