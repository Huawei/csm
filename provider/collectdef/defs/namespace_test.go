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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/huawei/csm/v2/provider/collectdef"
	"github.com/huawei/csm/v2/provider/constants"
)

func TestNamespaceDef_Structure(t *testing.T) {
	// arrange & act & assert
	assert.Equal(t, constants.Namespace, NamespaceDef.CollectType)
	assert.Equal(t, collectdef.CollectModePaginated, NamespaceDef.CollectMode)
	assert.True(t, NamespaceDef.SupportObject)
	assert.True(t, NamespaceDef.SupportPerformance)
	assert.NotNil(t, NamespaceDef.Perf)
	assert.Equal(t, 57356, NamespaceDef.Perf.TypeID)
}

func TestNamespaceDef_ObjectMetrics(t *testing.T) {
	// arrange
	expectedNames := []string{
		"space_hard_quota", "space_soft_quota", "space_used_capacity", "space_used_rate",
		"file_hard_quota", "file_soft_quota", "file_used_counts", "file_used_rate",
	}

	// act & assert
	assert.Len(t, NamespaceDef.ObjectMetrics, 8)
	for i, expected := range expectedNames {
		assert.Equal(t, expected, NamespaceDef.ObjectMetrics[i].Name, "ObjectMetrics[%d].Name", i)
	}
}

func TestNamespaceDef_PerfIndicators(t *testing.T) {
	// arrange & act & assert
	nfsCount := 0
	dpcCount := 0
	for _, name := range NamespaceDef.Perf.Indicators {
		if len(name) >= 4 && name[:4] == "nfs_" {
			nfsCount++
		}
		if len(name) >= 4 && name[:4] == "dpc_" {
			dpcCount++
		}
	}
	assert.Equal(t, 14, nfsCount, "NFS indicator count")
	assert.Equal(t, 9, dpcCount, "DPC indicator count")
	assert.Len(t, NamespaceDef.Perf.Indicators, 23, "Total indicator count")
}

func TestNamespaceDef_UrlKeys(t *testing.T) {
	// arrange & act & assert
	assert.Equal(t, "GetNamespaceCount", NamespaceDef.UrlKeys.CountKey)
	assert.Equal(t, "GetNamespace", NamespaceDef.UrlKeys.PageKey)
}

func TestNamespaceDef_ObjectLabels(t *testing.T) {
	// arrange
	expectedLabels := map[string]string{
		"endpoint": "backendName",
		"id":       "ID",
		"name":     "NAME",
		"object":   "collectorName",
	}

	// act & assert
	for _, label := range NamespaceDef.ObjectLabels {
		assert.Equal(t, expectedLabels[label.Name], label.SourceKey, "Label %q SourceKey", label.Name)
	}
}

func TestNamespaceDef_PerfLabels(t *testing.T) {
	// arrange
	expectedLabels := map[string]string{
		"endpoint": "backendName",
		"id":       "ObjectId",
		"object":   "collectorName",
		"name":     "ObjectName",
	}

	// act & assert
	for _, label := range NamespaceDef.PerfLabels {
		assert.Equal(t, expectedLabels[label.Name], label.SourceKey, "PerfLabel %q SourceKey", label.Name)
	}
}

func TestNamespaceDef_Transforms(t *testing.T) {
	// arrange
	data := map[string]string{
		"SPACE_HARD_QUOTA": "10485760",
		"SPACE_SOFT_QUOTA": "10485760",
		"SPACE_USED":       "3072000",
		"SPACE_USED_RATE":  "29",
		"FILE_HARD_QUOTA":  "1000000",
		"FILE_SOFT_QUOTA":  "900000",
		"FILE_USED":        "500000",
		"FILE_USED_RATE":   "50",
	}

	// act & assert
	for _, m := range NamespaceDef.ObjectMetrics {
		assert.NotNil(t, m.Transform, "ObjectMetric %q has nil Transform", m.Name)
		result := m.Transform(data)
		assert.NotEmpty(t, result, "ObjectMetric %q Transform returned empty for valid data", m.Name)
	}
}
