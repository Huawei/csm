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

package collector

import (
	"reflect"
	"testing"
)

func Test_pickPerformanceParsMap_MapString(t *testing.T) {
	// arrange
	mockMetricsData := []string{"22", "25"}
	want := map[string]string{
		"total_iops": "Total IOPS(IO/s)",
		"read_iops":  "Read IOPS(IO/s)",
	}

	// action
	got := pickPerformanceParsMap[string](mockMetricsData, performanceMetricsHelpMap)

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("pickPerformanceParsMap() got = %v, want %v", got, want)
	}
}

func Test_pickPerformanceParsMap_MapParseRelation(t *testing.T) {
	// arrange
	mockMetricsData := []string{"22", "25"}
	want := map[string]parseRelation{
		"total_iops": {"22", parseStorageData},
		"read_iops":  {"25", parseStorageData},
	}

	// action
	got := pickPerformanceParsMap[parseRelation](mockMetricsData, performanceMetricsParseMap)

	// assert
	for name, wantData := range want {
		gotData, ok := got[name]
		if !ok {
			t.Error("pickPerformanceParsMap() can not got data")
			return
		}
		if gotData.parseKey != wantData.parseKey {
			t.Error("pickPerformanceParsMap() got key not same")
			return
		}
		if reflect.ValueOf(gotData.parseFunc).Pointer() !=
			reflect.ValueOf(wantData.parseFunc).Pointer() {
			t.Error("pickPerformanceParsMap() got func not same")
			return
		}
	}
}
