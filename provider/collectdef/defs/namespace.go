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

package defs

import (
	"github.com/huawei/csm/v2/provider/collectdef"
	"github.com/huawei/csm/v2/provider/constants"
)

// NamespaceDef defines the namespace object and performance metrics for FusionStorage.
var NamespaceDef = &collectdef.ObjectDef{
	CollectType:        constants.Namespace,
	SupportObject:      true,
	SupportPerformance: true,
	CollectMode:        collectdef.CollectModePaginated,
	UrlKeys: collectdef.UrlKeys{
		CountKey: "GetNamespaceCount",
		PageKey:  "GetNamespace",
	},
	Perf: &collectdef.PerfDef{
		TypeID: 57356,
		Indicators: map[int]string{
			// NFS performance indicators
			30001: "nfs_read_bandwidth",
			30002: "nfs_write_bandwidth",
			30003: "nfs_total_bandwidth",
			30004: "nfs_read_ops",
			30005: "nfs_write_ops",
			30006: "nfs_total_ops",
			30011: "nfs_avg_latency",
			30012: "nfs_max_latency",
			30013: "nfs_write_avg_latency",
			30014: "nfs_write_max_latency",
			30015: "nfs_read_avg_latency",
			30016: "nfs_read_max_latency",
			30076: "nfs_read_io_avg_size",
			30077: "nfs_write_io_avg_size",
			// DPC performance indicators
			30043: "dpc_read_bandwidth",
			30044: "dpc_write_bandwidth",
			30045: "dpc_read_ops",
			30046: "dpc_write_ops",
			30048: "dpc_avg_read_latency",
			30049: "dpc_avg_write_latency",
			30051: "dpc_total_bandwidth",
			30052: "dpc_total_ops",
			31005: "dpc_avg_total_latency",
		},
	},
	ObjectMetrics: []collectdef.MetricDef{
		{
			Name: "space_hard_quota", Help: "Namespace Space Hard Quota(GB)",
			SourceKey: "SPACE_HARD_QUOTA", Transform: collectdef.KBToGBOf("SPACE_HARD_QUOTA"),
		},
		{
			Name: "space_soft_quota", Help: "Namespace Space Soft Quota(GB)",
			SourceKey: "SPACE_SOFT_QUOTA", Transform: collectdef.KBToGBOf("SPACE_SOFT_QUOTA"),
		},
		{
			Name: "space_used_capacity", Help: "Namespace Used Capacity(GB)",
			SourceKey: "SPACE_USED", Transform: collectdef.KBToGBOf("SPACE_USED"),
		},
		{
			Name: "space_used_rate", Help: "Namespace Space Usage Rate(%)",
			SourceKey: "SPACE_USED_RATE", Transform: collectdef.DirectValueOf("SPACE_USED_RATE"),
		},
		{
			Name: "file_hard_quota", Help: "Namespace File Hard Quota",
			SourceKey: "FILE_HARD_QUOTA", Transform: collectdef.DirectValueOf("FILE_HARD_QUOTA"),
		},
		{
			Name: "file_soft_quota", Help: "Namespace File Soft Quota",
			SourceKey: "FILE_SOFT_QUOTA", Transform: collectdef.DirectValueOf("FILE_SOFT_QUOTA"),
		},
		{
			Name: "file_used_counts", Help: "Namespace Used File Counts",
			SourceKey: "FILE_USED", Transform: collectdef.DirectValueOf("FILE_USED"),
		},
		{
			Name: "file_used_rate", Help: "Namespace File Usage Rate(%)",
			SourceKey: "FILE_USED_RATE", Transform: collectdef.DirectValueOf("FILE_USED_RATE"),
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
	collectdef.MustRegister(constants.FusionStorage, NamespaceDef)
}
