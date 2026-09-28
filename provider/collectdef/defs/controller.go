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

// ControllerDef declaratively defines controller's collection metadata.
var ControllerDef = &collectdef.ObjectDef{
	CollectType:        constants.Controller,
	SupportObject:      true,
	SupportPerformance: true,
	CollectMode:        collectdef.CollectModeList,
	UrlKeys: collectdef.UrlKeys{
		ListKey: "GetControllers",
	},
	Perf: &collectdef.PerfDef{
		TypeID: 207,
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
			Name: "cpu_usage", Help: "CPU utilization(%)",
			SourceKey: "CPUUSAGE",
		},
		{
			Name: "memory_usage", Help: "Memory utilization(%)",
			SourceKey: "MEMORYUSAGE",
		},
		{
			Name: "health_status", Help: "Health Status",
			SourceKey: "HEALTHSTATUS",
			Labels: []collectdef.LabelDef{
				{Name: "endpoint", SourceKey: "backendName"},
				{Name: "id", SourceKey: "ID"},
				{Name: "status", Transform: collectdef.HealthStatus},
				{Name: "name", SourceKey: "NAME"},
				{Name: "object", SourceKey: "collectorName"},
			},
		},
		{
			Name: "running_status", Help: "Running Status",
			SourceKey: "RUNNINGSTATUS",
			Labels: []collectdef.LabelDef{
				{Name: "endpoint", SourceKey: "backendName"},
				{Name: "id", SourceKey: "ID"},
				{Name: "status", Transform: collectdef.RunningStatus},
				{Name: "name", SourceKey: "NAME"},
				{Name: "object", SourceKey: "collectorName"},
			},
		},
	},
	ObjectLabels: []collectdef.LabelDef{
		{Name: "endpoint", SourceKey: "backendName"},
		{Name: "id", SourceKey: "ID"},
		{Name: "name", SourceKey: "NAME"},
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
	collectdef.MustRegister(constants.OceanStorage, ControllerDef)
}
