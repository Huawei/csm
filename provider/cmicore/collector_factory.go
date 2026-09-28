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
	"context"
	"fmt"

	"github.com/huawei/csm/v2/provider/constants"
)

// CollectorInterface is the interface for collectors
type CollectorInterface interface {
	Collect(ctx context.Context, req *CollectRequest) (*CollectResponse, error)
}

// CollectorFactory creates collectors by metrics type
type CollectorFactory struct {
	clientCache  *ClientCache
	discoverFunc func(context.Context, string) (ClientInfo, error)
}

// NewCollectorFactory creates a new CollectorFactory
func NewCollectorFactory(
	clientCache *ClientCache,
	discoverFunc func(context.Context, string) (ClientInfo, error),
) *CollectorFactory {
	return &CollectorFactory{
		clientCache:  clientCache,
		discoverFunc: discoverFunc,
	}
}

// GetCollector returns the appropriate collector for the metrics type
func (f *CollectorFactory) GetCollector(metricsType string) (CollectorInterface, error) {
	switch metricsType {
	case constants.Object:
		return NewObjectCollector(f.clientCache, f.discoverFunc), nil
	case constants.Performance:
		return NewPerformanceCollector(f.clientCache, f.discoverFunc), nil
	default:
		return nil, fmt.Errorf("not found collector, metrics type is [%s]", metricsType)
	}
}
