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
	"strconv"

	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/utils/log"
)

// Ensure PerformanceCollector implements CollectorInterface
var _ CollectorInterface = (*PerformanceCollector)(nil)

// PerformanceCollector collects performance data from storage backends
type PerformanceCollector struct {
	clientCache  *ClientCache
	discoverFunc func(context.Context, string) (ClientInfo, error)
}

// NewPerformanceCollector creates a new PerformanceCollector
func NewPerformanceCollector(
	clientCache *ClientCache,
	discoverFunc func(context.Context, string) (ClientInfo, error),
) *PerformanceCollector {
	return &PerformanceCollector{
		clientCache:  clientCache,
		discoverFunc: discoverFunc,
	}
}

// Collect finds a handler and invokes it for performance data collection
func (p *PerformanceCollector) Collect(ctx context.Context, req *CollectRequest) (*CollectResponse, error) {
	clientInfo, err := p.clientCache.DiscoverClient(ctx, req.BackendName, p.discoverFunc)
	if err != nil {
		log.AddContext(ctx).Errorf("performanceCollector get client failed, error: %v", err)
		return nil, err
	}

	handler, err := GetPerformanceHandler(clientInfo.StorageType, req.CollectType)
	if err != nil {
		log.AddContext(ctx).Warningf("performanceCollector get handler function failed, error: %v", err)
		return nil, err
	}

	return handler(ctx, clientInfo.Client, req)
}

// MergePerformance merges performance data with object name mapping.
// This is a helper for performance handlers to compose the final response.
func MergePerformance(performances []PerformanceIndicators,
	nameMapping map[string]string, req *CollectRequest) *CollectResponse {

	response := BuildResponse(req)
	for _, performance := range performances {
		mapData := performance.ToMap()
		objectName, ok := nameMapping[performance.ObjectId]
		if !ok {
			continue
		}
		mapData[constants.ObjectName] = objectName
		mapData[constants.ObjectId] = performance.ObjectId
		AddCollectDetailWithMap(mapData, response)
	}

	return response
}

// ToMap converts PerformanceIndicators to a map[string]string
func (p PerformanceIndicators) ToMap() map[string]string {
	if len(p.Indicators) == 0 || len(p.Indicators) != len(p.IndicatorValues) {
		return map[string]string{}
	}

	dataMap := map[string]string{}
	for i, indicator := range p.Indicators {
		key := strconv.Itoa(indicator)
		dataMap[key] = strconv.FormatFloat(p.IndicatorValues[i],
			'f', constants.UnlimitedPrecision, constants.DefaultBitSize)
	}
	return dataMap
}
