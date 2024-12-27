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

// Package server is a package that implement grpc interface
package server

import (
	"context"
	"errors"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/collect"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/provider/grpc/helper"
	"github.com/huawei/csm/v2/utils/log"
)

var collectValidator = helper.NewValidator[*cmi.CollectRequest](validateBackendName, validateCollectType,
	validateMetricsType)

// Collector This object implements the cmi.CollectorServer service.
type Collector struct{}

// Collect This method is the entry point for collecting data.
// The purpose is to find an adapter and call its collect method.
func (c *Collector) Collect(ctx context.Context, request *cmi.CollectRequest) (*cmi.CollectResponse, error) {
	log.AddContext(ctx).Infof("Start to collect, request: %v", request)
	defer log.AddContext(ctx).Infof("Finish to collect, backend name %s", request.BackendName)

	if err := collectValidator.Validate(request); err != nil {
		return nil, err
	}

	collector, err := collect.GetCollector(request.GetMetricsType())
	if err != nil {
		log.AddContext(ctx).Errorf("Get collector failed, error: %v", err)
		return nil, err
	}
	log.AddContext(ctx).Infof("Get collector success, collector: %v", collector)

	return collector.Collect(ctx, request)
}

// validateBackendName validate if the backend name is blank
func validateBackendName(request *cmi.CollectRequest) error {
	if request.GetBackendName() == "" {
		return errors.New("illegalArgumentError backend name is blank")
	}
	return nil
}

// validateCollectType validate if the collect type is blank
func validateCollectType(request *cmi.CollectRequest) error {
	if request.GetCollectType() == "" {
		return errors.New("illegalArgumentError collect type is blank")
	}
	return nil
}

// validateMetricsType validate if the metrics type is blank
func validateMetricsType(request *cmi.CollectRequest) error {
	if request.GetMetricsType() == "" {
		return errors.New("illegalArgumentError metrics type is blank")
	}
	if request.GetMetricsType() != constants.Object && request.GetMetricsType() != constants.Performance {
		return errors.New("illegalArgumentError unsupported metrics type")
	}
	return nil
}
