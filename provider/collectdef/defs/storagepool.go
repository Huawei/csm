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
)

// StoragePoolDef declaratively defines storagepool's collection metadata.
// PrometheusName "storage_pool" unifies the subsystem name for both object and performance modes.
var StoragePoolDef = &collectdef.ObjectDef{
	CollectType:        constants.StoragePool,
	SupportObject:      true,
	SupportPerformance: true,
	CollectMode:        collectdef.CollectModeList,
	UrlKeys: collectdef.UrlKeys{
		ListKey: "GetStoragePools",
	},
	Perf: &collectdef.PerfDef{
		TypeID: 216,
		Indicators: map[int]string{
			22:  "total_iops",
			25:  "read_iops",
			28:  "write_iops",
			21:  "total_bandwidth",
			23:  "read_bandwidth",
			26:  "write_bandwidth",
			370: "avg_io_response_time",
			182: "ops",
			524: "avg_read_ops_response_time",
			525: "avg_write_ops_response_time",
		},
	},
	ObjectMetrics: []collectdef.MetricDef{
		{
			Name: "total_capacity", Help: "Total capacity(GB) of storage pool",
			Transform: collectdef.SectorsToGBOf("USERTOTALCAPACITY"),
		},
		{
			Name: "free_capacity", Help: "Free capacity(GB) of storage pool",
			Transform: collectdef.SectorsToGBOf("USERFREECAPACITY"),
		},
		{
			Name: "capacity_usage", Help: "Used capacity ratio(%) of storage pool",
			Transform: collectdef.CapacityUsageOf("USERCONSUMEDCAPACITY", "USERTOTALCAPACITY"),
		},
		{
			Name: "used_capacity", Help: "Used capacity(GB) of storage pool",
			Transform: collectdef.SectorsToGBOf("USERCONSUMEDCAPACITY"),
		},
	},
	ObjectLabels: []collectdef.LabelDef{
		{Name: "name", SourceKey: "NAME"},
		{Name: "endpoint", SourceKey: "backendName"},
		{Name: "id", SourceKey: "ID"},
		{Name: "object", SourceKey: "collectorName"},
	},
	PerfLabels: []collectdef.LabelDef{
		{Name: "endpoint", SourceKey: "backendName"},
		{Name: "id", SourceKey: "ObjectId"},
		{Name: "object", SourceKey: "collectorName"},
		{Name: "name", SourceKey: "ObjectName"},
	},
}

func init() {
	collectdef.MustRegister(constants.OceanStorage, StoragePoolDef)
}
