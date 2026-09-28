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

// ArrayDef declaratively defines array's collection metadata.
var ArrayDef = &collectdef.ObjectDef{
	CollectType:        constants.Array,
	SupportObject:      true,
	SupportPerformance: false,
	CollectMode:        collectdef.CollectModeSingle,
	UrlKeys: collectdef.UrlKeys{
		SingleKey: "GetSystemInfo",
	},
	ObjectMetrics: []collectdef.MetricDef{
		{
			Name: "basic_info", Help: "Huawei Storage Array Basic Info",
			Transform: collectdef.ReturnZero,
			Labels: []collectdef.LabelDef{
				{Name: "endpoint", SourceKey: "backendName"},
				{Name: "sn", SourceKey: "ID"},
				{Name: "model", Transform: collectdef.ArrayModel},
				{Name: "version", Transform: collectdef.ArrayVersion},
				{Name: "object", SourceKey: "collectorName"},
			},
		},
		{
			Name: "health_status", Help: "Huawei Storage Array Health Status",
			SourceKey: "HEALTHSTATUS",
			Labels: []collectdef.LabelDef{
				{Name: "endpoint", SourceKey: "backendName"},
				{Name: "sn", SourceKey: "ID"},
				{Name: "status", Transform: collectdef.HealthStatus},
				{Name: "object", SourceKey: "collectorName"},
			},
		},
		{
			Name: "running_status", Help: "Huawei Storage Array Running Status",
			SourceKey: "RUNNINGSTATUS",
			Labels: []collectdef.LabelDef{
				{Name: "endpoint", SourceKey: "backendName"},
				{Name: "sn", SourceKey: "ID"},
				{Name: "status", Transform: collectdef.RunningStatus},
				{Name: "object", SourceKey: "collectorName"},
			},
		},
	},
}

func init() {
	collectdef.MustRegister(constants.OceanStorage, ArrayDef)
}
