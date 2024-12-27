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
	"context"
	"sync"

	cmiConfig "github.com/huawei/csm/v2/config/cmi"
	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/provider/utils"
)

var (
	// IndicatorsMapping storage indicators mapping
	IndicatorsMapping = map[string]int{
		constants.Filesystem:  40,
		constants.Lun:         11,
		constants.Controller:  207,
		constants.StoragePool: 216,
	}
)

// CountFunc count function, e.g. query total filesystem number in storage
type CountFunc func(ctx context.Context) (int, error)

// QueryFunc query function, e.g. query filesystem information
type QueryFunc func(context.Context) ([]map[string]interface{}, error)

// PageFunc page query function, e.g. page query filesystem information
type PageFunc func(context.Context, int, int) ([]map[string]interface{}, error)

// BuildResponse build a collect response
func BuildResponse(request *cmi.CollectRequest) *cmi.CollectResponse {
	return &cmi.CollectResponse{
		BackendName: request.GetBackendName(),
		CollectType: request.GetCollectType(),
		MetricsType: request.GetMetricsType(),
		Details:     []*cmi.CollectDetail{},
	}
}

// AddCollectDetail add detail to the response
// input type is struct
func AddCollectDetail[T any](t T, response *cmi.CollectResponse) {
	AddCollectDetailWithMap(utils.StructToMap(t), response)
}

// AddCollectDetailWithMap add detail to the response
// input type is a map
func AddCollectDetailWithMap(data map[string]string, response *cmi.CollectResponse) {
	detail := &cmi.CollectDetail{Data: data}
	response.Details = append(response.Details, detail)
}

// ConvertToResponse convert input to response
func ConvertToResponse[I, T any](input I, request *cmi.CollectRequest) (*cmi.CollectResponse, error) {
	targets, err := utils.MapToStructSlice[I, T](input)
	if err != nil {
		return nil, err
	}

	response := BuildResponse(request)
	for _, target := range targets {
		AddCollectDetail(target, response)
	}

	return response, nil
}

// BuildFailedPageResult build a failed paginated result
func BuildFailedPageResult(err error) PageResultTuple {
	return PageResultTuple{
		Data:  []map[string]interface{}{},
		Error: err,
	}
}

// BuildSuccessPageResult build a successful paginated result
func BuildSuccessPageResult(data []map[string]interface{}) PageResultTuple {
	return PageResultTuple{Data: data}
}

// ConcurrentPaginate a universal concurrent paging query function
// Each page will use a goroutine to query
func ConcurrentPaginate(ctx context.Context, count CountFunc, query PageFunc) ([]map[string]interface{}, error) {
	total, err := count(ctx)
	if err != nil {
		return []map[string]interface{}{}, err
	}

	var wg sync.WaitGroup
	var out = make(chan PageResultTuple)
	var start, pageSize = 0, cmiConfig.GetQueryStoragePageSize()
	for total > 0 {
		end := start + pageSize
		wg.Add(1)
		go pageQuery(ctx, start, end, &wg, query, out)
		start = end
		total -= pageSize
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return ReadQueryResult(out)
}

// pageQuery page query storage data
func pageQuery(ctx context.Context, start, end int, wg *sync.WaitGroup, query PageFunc, ch chan<- PageResultTuple) {
	defer wg.Done()
	var result PageResultTuple
	pageData, err := query(ctx, start, end)
	if err != nil {
		result = BuildFailedPageResult(err)
	}
	result = BuildSuccessPageResult(pageData)
	ch <- result
}

// ReadQueryResult read query result form channel
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
