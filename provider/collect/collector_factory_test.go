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

// Package collect is a package that provides object and performance collect
package collect

import (
	"reflect"
	"testing"

	"github.com/huawei/csm/v2/provider/constants"
)

func TestGetCollector_with_object_collector(t *testing.T) {
	// arrange
	var testMetricsType = constants.Object

	// action
	collector, err := GetCollector(testMetricsType)

	// assert
	if err != nil {
		t.Errorf("TestGetCollector_with_object_collector() error = %v", err)
		return
	}

	if !reflect.DeepEqual(collector, &ObjectCollector{}) {
		t.Errorf("GetCollector() got = %v, want %v", collector, &ObjectCollector{})
	}
}

func TestGetCollector_with_performance_collector(t *testing.T) {
	// arrange
	var testMetricsType = constants.Performance

	// action
	collector, err := GetCollector(testMetricsType)

	// assert
	if err != nil {
		t.Errorf("TestGetCollector_with_performance_collector() error = %v", err)
		return
	}

	if !reflect.DeepEqual(collector, &PerformanceCollector{}) {
		t.Errorf("TestGetCollector_with_performance_collector() got = %v, want %v",
			collector, &PerformanceCollector{})
	}
}

func TestGetCollector_with_collector_not_exist(t *testing.T) {
	// arrange
	var testMetricsType = "not-exist-type"

	// action
	_, err := GetCollector(testMetricsType)

	// assert
	if err == nil {
		t.Errorf("TestGetCollector_with_collector_not_exist() want an error but error is nil")
		return
	}
}
