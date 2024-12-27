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
	"errors"
	"fmt"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/constants"
)

var collectorMap = map[string]cmi.CollectorServer{
	constants.Object:      &ObjectCollector{},
	constants.Performance: &PerformanceCollector{},
}

// GetCollector get collector by metrics type
func GetCollector(metricsType string) (cmi.CollectorServer, error) {
	collector, ok := collectorMap[metricsType]
	if ok {
		return collector, nil
	}
	errMsg := fmt.Sprintf("not found collector, metricsType type is [%s] ", metricsType)
	return nil, errors.New(errMsg)
}
