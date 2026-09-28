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
	"sync"

	"github.com/huawei/csm/v2/config/cmi"
)

var (
	// IndicatorsMapping maps collect types to their API object type IDs
	IndicatorsMapping = map[string]int{}
)

// CountFunc is a function that returns the total count of items (e.g. total LUN count)
type CountFunc func(ctx context.Context) (int, error)

// QueryFunc is a function that queries all items at once
type QueryFunc func(ctx context.Context) ([]map[string]interface{}, error)

// PageFunc is a function that queries a page of items by range
type PageFunc func(ctx context.Context, start, end int) ([]map[string]interface{}, error)

// BuildResponse creates a CollectResponse with request metadata
func BuildResponse(req *CollectRequest) *CollectResponse {
	return &CollectResponse{
		BackendName: req.BackendName,
		CollectType: req.CollectType,
		MetricsType: req.MetricsType,
		Details:     []*CollectDetail{},
	}
}

// AddCollectDetailWithMap adds a map as a detail to the response
func AddCollectDetailWithMap(data map[string]string, response *CollectResponse) {
	detail := &CollectDetail{Data: data}
	response.Details = append(response.Details, detail)
}

// BuildFailedPageResult builds a failed paginated result
func BuildFailedPageResult(err error) PageResultTuple {
	return PageResultTuple{
		Data:  []map[string]interface{}{},
		Error: err,
	}
}

// BuildSuccessPageResult builds a successful paginated result
func BuildSuccessPageResult(data []map[string]interface{}) PageResultTuple {
	return PageResultTuple{Data: data}
}

// ConcurrentPaginate performs concurrent page queries using goroutines.
// Each page query runs in its own goroutine for parallel execution.
func ConcurrentPaginate(ctx context.Context, count CountFunc, query PageFunc) ([]map[string]interface{}, error) {
	total, err := count(ctx)
	if err != nil {
		return []map[string]interface{}{}, err
	}

	var wg sync.WaitGroup
	out := make(chan PageResultTuple)
	start, pageSize := 0, cmi.GetQueryStoragePageSize()

	// pageQuery is a closure capturing ctx, query, and out
	pageQuery := func(start, end int) {
		defer wg.Done()
		pageData, err := query(ctx, start, end)
		if err != nil {
			out <- BuildFailedPageResult(err)
			return
		}
		out <- BuildSuccessPageResult(pageData)
	}

	for total > 0 {
		end := start + pageSize
		wg.Add(1)
		go pageQuery(start, end)
		start = end
		total -= pageSize
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return ReadQueryResult(out)
}

// ReadQueryResult reads query results from a channel and aggregates them
func ReadQueryResult(input <-chan PageResultTuple) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	for tuple := range input {
		if tuple.Error != nil {
			return nil, tuple.Error
		}
		result = append(result, tuple.Data...)
	}
	return result, nil
}

// RegisterIndicatorMapping dynamically registers a collect type to API object type ID mapping.
// This is typically called during package initialization to populate IndicatorsMapping.
func RegisterIndicatorMapping(collectType string, typeID int) {
	IndicatorsMapping[collectType] = typeID
}

// ConvertMapToResponse converts raw map data to a CollectResponse without struct reflection.
// Used by generic handlers that fetch data via GetByUrlKey methods.
func ConvertMapToResponse(data []map[string]interface{}, req *CollectRequest) *CollectResponse {
	response := BuildResponse(req)
	for _, item := range data {
		detailMap := make(map[string]string)
		for k, v := range item {
			if v != nil {
				detailMap[k] = fmt.Sprintf("%v", v)
			}
		}
		AddCollectDetailWithMap(detailMap, response)
	}
	return response
}
