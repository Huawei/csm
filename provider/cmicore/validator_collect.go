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

// Package cmicore provides a shared library for storage backend operations
package cmicore

import (
	"errors"
	"slices"

	"github.com/huawei/csm/v2/provider/constants"
)

// validCollectTypes defines the allowed collect type values
var validCollectTypes = []string{
	constants.Lun,
	constants.Array,
	constants.Controller,
	constants.Filesystem,
	constants.StoragePool,
	constants.Namespace,
}

// validMetricsTypes defines the allowed metrics type values
var validMetricsTypes = []string{
	constants.Object,
	constants.Performance,
}

func (c *Core) validateBackendName(req *CollectRequest) error {
	if req.BackendName == "" {
		return errors.New("illegalArgumentError backend name is blank")
	}
	return nil
}

func (c *Core) validateCollectType(req *CollectRequest) error {
	if req.CollectType == "" {
		return errors.New("illegalArgumentError collect type is blank")
	}
	if !slices.Contains(validCollectTypes, req.CollectType) {
		return errors.New("illegalArgumentError unsupported collect type")
	}
	return nil
}

func (c *Core) validateMetricsType(req *CollectRequest) error {
	if req.MetricsType == "" {
		return errors.New("illegalArgumentError metrics type is blank")
	}
	if !slices.Contains(validMetricsTypes, req.MetricsType) {
		return errors.New("illegalArgumentError unsupported metrics type")
	}
	return nil
}
