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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/huawei/csm/v2/provider/collectdef"
)

func TestArrayDef_Success(t *testing.T) {
	// arrange
	wantCollectMode := collectdef.CollectModeSingle

	// action
	gotCollectMode := ArrayDef.CollectMode

	// assert
	assert.NotEmpty(t, ArrayDef.CollectType)
	assert.Equal(t, wantCollectMode, gotCollectMode)
	assert.True(t, ArrayDef.SupportObject)
	assert.False(t, ArrayDef.SupportPerformance)
	assert.NotNil(t, ArrayDef.ObjectMetrics)
}

func TestControllerDef_Success(t *testing.T) {
	// arrange
	wantCollectMode := collectdef.CollectModeList

	// action
	gotCollectMode := ControllerDef.CollectMode

	// assert
	assert.NotEmpty(t, ControllerDef.CollectType)
	assert.Equal(t, wantCollectMode, gotCollectMode)
	assert.NotNil(t, ControllerDef.Perf)
	assert.NotZero(t, ControllerDef.Perf.TypeID)
}

func TestStoragePoolDef_Success(t *testing.T) {
	// arrange
	wantCollectMode := collectdef.CollectModeList

	// action
	gotCollectMode := StoragePoolDef.CollectMode

	// assert
	assert.NotEmpty(t, StoragePoolDef.CollectType)
	assert.Equal(t, wantCollectMode, gotCollectMode)
}

func TestLunDef_Success(t *testing.T) {
	// arrange
	wantCollectMode := collectdef.CollectModePaginated

	// action
	gotCollectMode := LunDef.CollectMode

	// assert
	assert.NotEmpty(t, LunDef.CollectType)
	assert.Equal(t, wantCollectMode, gotCollectMode)
	assert.NotEmpty(t, LunDef.UrlKeys.CountKey)
	assert.NotEmpty(t, LunDef.UrlKeys.PageKey)
}

func TestFilesystemDef_Success(t *testing.T) {
	// arrange
	wantCollectMode := collectdef.CollectModePaginated

	// action
	gotCollectMode := FilesystemDef.CollectMode

	// assert
	assert.NotEmpty(t, FilesystemDef.CollectType)
	assert.Equal(t, wantCollectMode, gotCollectMode)
	assert.NotEmpty(t, FilesystemDef.UrlKeys.CountKey)
	assert.NotEmpty(t, FilesystemDef.UrlKeys.PageKey)
}

func TestVstoreDef_Success(t *testing.T) {
	// arrange
	wantCollectMode := collectdef.CollectModeNone

	// action
	gotCollectMode := VstoreDef.CollectMode

	// assert
	assert.NotEmpty(t, VstoreDef.CollectType)
	assert.Equal(t, wantCollectMode, gotCollectMode)
	assert.NotNil(t, VstoreDef.MetricsDataFactory)
}

func TestAllDefsHaveObjectLabels_Success(t *testing.T) {
	// arrange — defs that use per-metric Labels instead of ObjectLabels
	perMetricLabelDefs := map[string]bool{"array": true}
	defs := []*collectdef.ObjectDef{
		ArrayDef, ControllerDef, StoragePoolDef, LunDef, FilesystemDef, NamespaceDef, VstoreDef,
	}

	// action & assert
	for _, d := range defs {
		if d.SupportObject && !perMetricLabelDefs[d.CollectType] {
			assert.NotEmpty(t, d.ObjectLabels, "%s: SupportObject=true but ObjectLabels is empty", d.CollectType)
		}
	}
}

func TestAllPerfDefsHavePerfLabels_Success(t *testing.T) {
	// arrange
	defs := []*collectdef.ObjectDef{
		ArrayDef, ControllerDef, StoragePoolDef, LunDef, FilesystemDef, NamespaceDef,
	}

	// action & assert
	for _, d := range defs {
		if d.SupportPerformance {
			assert.NotEmpty(t, d.PerfLabels, "%s: SupportPerformance=true but PerfLabels is empty", d.CollectType)
		}
	}
}
