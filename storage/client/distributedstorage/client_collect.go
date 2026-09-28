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

package distributedstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	distributedstorage "github.com/huawei/csm/v2/storage/api/distributedstorage"
	"github.com/huawei/csm/v2/storage/utils"
	"github.com/huawei/csm/v2/utils/log"
)

const (
	// defaultPageSize is the default page size for paginated queries.
	defaultPageSize = 100
	// performanceTimeRangeSeconds is the time range in seconds for performance data queries.
	performanceTimeRangeSeconds = 900
	// systemTimeKey is the response field name for system UTC time in FusionStorage API.
	systemTimeKey = "CMO_SYS_UTC_TIME"
)

// GetSingleByUrlKey queries a single object by URL key.
func (c *DistributedClient) GetSingleByUrlKey(ctx context.Context,
	urlKey string) (map[string]interface{}, error) {
	apiPath, err := distributedstorage.GenerateUrl(urlKey, nil)
	if err != nil {
		return nil, fmt.Errorf("generate url failed for %s: %w", urlKey, err)
	}

	fullURL := buildFullURL(c.Curl, apiPath)
	resp, err := c.callWithRetry(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("get single by url key %s failed: %w", urlKey, err)
	}

	data, ok := utils.GetValue[map[string]interface{}](resp, "data")
	if !ok {
		return nil, fmt.Errorf("data field can not convert to map in response %v", resp)
	}

	return normalizeFieldNames(data), nil
}

// GetListByUrlKey queries a list of objects by URL key.
func (c *DistributedClient) GetListByUrlKey(ctx context.Context,
	urlKey string) ([]map[string]interface{}, error) {
	apiPath, err := distributedstorage.GenerateUrl(urlKey, nil)
	if err != nil {
		return nil, fmt.Errorf("generate url failed for %s: %w", urlKey, err)
	}

	fullURL := buildFullURL(c.Curl, apiPath)
	resp, err := c.callWithRetry(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("get list by url key %s failed: %w", urlKey, err)
	}

	dataList, ok := utils.GetValue[[]interface{}](resp, "data")
	if !ok {
		return nil, fmt.Errorf("data field can not convert to array in response %v", resp)
	}

	result := make([]map[string]interface{}, 0, len(dataList))
	for _, item := range dataList {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, normalizeFieldNames(m))
		}
	}
	return result, nil
}

// GetCountByUrlKey queries the count of objects by URL key.
func (c *DistributedClient) GetCountByUrlKey(ctx context.Context, urlKey string) (int, error) {
	apiPath, err := distributedstorage.GenerateUrl(urlKey, nil)
	if err != nil {
		return 0, fmt.Errorf("generate url failed for %s: %w", urlKey, err)
	}

	fullURL := buildFullURL(c.Curl, apiPath)
	resp, err := c.callWithRetry(ctx, "GET", fullURL, nil)
	if err != nil {
		return 0, fmt.Errorf("get count by url key %s failed: %w", urlKey, err)
	}

	data, ok := utils.GetValue[map[string]interface{}](resp, "data")
	if !ok {
		return 0, fmt.Errorf("data field can not convert to map in response %v", resp)
	}

	count, ok := utils.GetValue[float64](data, "count")
	if !ok {
		return 0, fmt.Errorf("count field can not convert to int in response data %v", data)
	}

	return int(count), nil
}

// GetPageByUrlKey queries a page of objects by URL key using JSON-encoded range parameter.
func (c *DistributedClient) GetPageByUrlKey(ctx context.Context,
	urlKey string, start, end int) ([]map[string]interface{}, error) {
	rangeParam := map[string]int{"offset": start, "limit": end - start}
	rangeJSON, err := json.Marshal(rangeParam)
	if err != nil {
		return nil, fmt.Errorf("marshal range parameter failed: %w", err)
	}

	args := map[string]interface{}{"range": url.QueryEscape(string(rangeJSON))}
	apiPath, err := distributedstorage.GenerateUrl(urlKey, args)
	if err != nil {
		return nil, fmt.Errorf("generate url failed for %s: %w", urlKey, err)
	}

	fullURL := buildFullURL(c.Curl, apiPath)
	resp, err := c.callWithRetry(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("get page by url key %s failed: %w", urlKey, err)
	}

	dataList, ok := utils.GetValue[[]interface{}](resp, "data")
	if !ok {
		return nil, fmt.Errorf("data field can not convert to array in response %v", resp)
	}

	result := make([]map[string]interface{}, 0, len(dataList))
	for _, item := range dataList {
		if m, ok := item.(map[string]interface{}); ok {
			result = append(result, normalizeFieldNames(m))
		}
	}
	return result, nil
}

// QueryPerformanceData queries performance data from FusionStorage.
// It internally: (1) gets system time, (2) looks up objectTypeMapping to get object query keys,
// (3) queries object count and pages to extract IDs, (4) builds v2 request body,
// (5) POSTs to performance_data, (6) transposes long-format to wide-format.
func (c *DistributedClient) QueryPerformanceData(ctx context.Context,
	objectType int, indicators []string) ([]map[string]interface{}, error) {
	// Get system time for time range
	endTime, err := c.getSystemTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("get system time failed: %w", err)
	}
	startTime := endTime - performanceTimeRangeSeconds

	// Look up objectTypeMapping
	typeInfo, ok := objectTypeMapping[objectType]
	if !ok {
		return nil, fmt.Errorf("objectType %d is not registered in objectTypeMapping", objectType)
	}

	// Get object IDs
	ids, err := c.getObjectIds(ctx, typeInfo)
	if err != nil {
		log.AddContext(ctx).Errorf("get object ids for objectType %d failed: %v", objectType, err)
		return nil, err
	}

	if len(ids) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Convert indicators to int
	intIndicators := make([]int, 0, len(indicators))
	for _, ind := range indicators {
		val, err := strconv.Atoi(ind)
		if err != nil {
			log.AddContext(ctx).Warningf("invalid indicator %q, skipping: %v", ind, err)
			continue
		}
		intIndicators = append(intIndicators, val)
	}
	if len(intIndicators) == 0 {
		return nil, fmt.Errorf("no valid indicators after conversion, input: %v", indicators)
	}

	// Build request body
	reqBody := map[string]interface{}{
		"begin_time": startTime,
		"end_time":   endTime,
		"objects": []map[string]interface{}{
			{
				"object_type": objectType,
				"indicators":  intIndicators,
				"ids":         ids,
			},
		},
	}

	// POST to performance_data
	apiPath, err := distributedstorage.GenerateUrl("PerformanceData", nil)
	if err != nil {
		return nil, fmt.Errorf("generate performance url failed: %w", err)
	}

	fullURL := buildFullURL(c.Curl, apiPath)
	resp, err := c.callWithRetry(ctx, "POST", fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("query performance data failed: %w", err)
	}

	dataList, ok := utils.GetValue[[]interface{}](resp, "data")
	if !ok {
		return nil, fmt.Errorf("data field can not convert to array in response %v", resp)
	}

	return c.groupByObjectId(dataList), nil
}

// getSystemTime queries the system UTC time via ESN-based URL.
func (c *DistributedClient) getSystemTime(ctx context.Context) (int64, error) {
	args := map[string]interface{}{"esn": c.ESN}
	apiPath, err := distributedstorage.GenerateUrl("SystemTime", args)
	if err != nil {
		return 0, fmt.Errorf("generate system time url failed: %w", err)
	}

	fullURL := buildFullURL(c.Curl, apiPath)
	resp, err := c.callWithRetry(ctx, "GET", fullURL, nil)
	if err != nil {
		return 0, fmt.Errorf("get system time failed: %w", err)
	}

	data, ok := utils.GetValue[map[string]interface{}](resp, "data")
	if !ok {
		return 0, fmt.Errorf("data field can not convert to map in response %v", resp)
	}

	timeVal, ok := data[systemTimeKey]
	if !ok {
		return 0, fmt.Errorf("get system time: CMO_SYS_UTC_TIME field not found")
	}

	switch v := timeVal.(type) {
	case float64:
		return int64(v), nil
	case string:
		t, err := strconv.ParseInt(v, decimalBase, int64BitSize)
		if err != nil {
			return 0, fmt.Errorf("get system time: CMO_SYS_UTC_TIME value %q is not a number", v)
		}
		return t, nil
	default:
		return 0, fmt.Errorf("get system time: CMO_SYS_UTC_TIME has unexpected type %T", timeVal)
	}
}

// getObjectIds retrieves object IDs for performance queries.
// For CollectModeList, it fetches the full list and extracts IDs.
// For CollectModePaginated, it uses count+page queries.
func (c *DistributedClient) getObjectIds(ctx context.Context, info objectTypeInfo) ([]int, error) {
	if info.ListKey != "" {
		return c.getObjectIdsFromList(ctx, info)
	}
	return c.getObjectIdsFromPages(ctx, info)
}

// getObjectIdsFromList fetches all objects via list query and extracts IDs.
func (c *DistributedClient) getObjectIdsFromList(ctx context.Context, info objectTypeInfo) ([]int, error) {
	data, err := c.GetListByUrlKey(ctx, info.ListKey)
	if err != nil {
		return nil, fmt.Errorf("get object list failed: %w", err)
	}

	var ids []int
	for _, item := range data {
		id, ok := utils.GetValue[float64](item, info.IDKey)
		if !ok {
			continue
		}
		ids = append(ids, int(id))
	}
	return ids, nil
}

// getObjectIdsFromPages fetches object IDs using count and page queries.
func (c *DistributedClient) getObjectIdsFromPages(ctx context.Context, info objectTypeInfo) ([]int, error) {
	count, err := c.GetCountByUrlKey(ctx, info.CountKey)
	if err != nil {
		return nil, fmt.Errorf("get object count failed: %w", err)
	}

	if count == 0 {
		return []int{}, nil
	}

	var allIDs []int
	var hasFailedPage bool
	for start := 0; start < count; start += defaultPageSize {
		end := start + defaultPageSize
		if end > count {
			end = count
		}

		page, err := c.GetPageByUrlKey(ctx, info.PageKey, start, end)
		if err != nil {
			log.AddContext(ctx).Errorf("get object page failed, start=%d: %v", start, err)
			hasFailedPage = true
			continue
		}

		for _, item := range page {
			id, ok := utils.GetValue[float64](item, info.IDKey)
			if !ok {
				continue
			}
			allIDs = append(allIDs, int(id))
		}
	}

	if hasFailedPage {
		log.AddContext(ctx).Warningf("get object ids completed with failed pages, "+
			"expected=%d, got=%d", count, len(allIDs))
	}

	return allIDs, nil
}

// groupByObjectId converts long-format performance data to a struct-friendly format.
// API returns one row per object x indicator: {id, indicator, indicator_values, timestamp, ...}
// This function groups by object ID and produces maps compatible with PerformanceIndicators:
//
//	{"object_id": "179", "indicators": [30001, 30002], "indicator_values": [0, 54784]}
//
// The latest value is taken from indicator_values[len-1] (timestamps are in ascending order).
func (c *DistributedClient) groupByObjectId(data []interface{}) []map[string]interface{} {
	type objEntry struct {
		indicators      []int
		indicatorValues []float64
	}

	grouped := make(map[string]*objEntry)
	var idOrder []string

	for _, item := range data {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		objID, _ := utils.GetValue[string](row, "id")
		if objID == "" {
			continue
		}

		indicatorStr, _ := utils.GetValue[string](row, "indicator")
		indicatorInt, err := strconv.Atoi(indicatorStr)
		if err != nil {
			continue
		}

		indicatorValues, ok := utils.GetValue[[]interface{}](row, "indicator_values")
		if !ok || len(indicatorValues) == 0 {
			continue
		}

		// Take the latest value (indicator_values[len-1], timestamps are ascending)
		// FusionStorage returns indicator_values as string array
		latestValStr, ok := indicatorValues[len(indicatorValues)-1].(string)
		if !ok {
			continue
		}
		latestVal, err := strconv.ParseFloat(latestValStr, float64BitSize)
		if err != nil {
			continue
		}

		if _, exists := grouped[objID]; !exists {
			grouped[objID] = &objEntry{}
			idOrder = append(idOrder, objID)
		}
		grouped[objID].indicators = append(grouped[objID].indicators, indicatorInt)
		grouped[objID].indicatorValues = append(grouped[objID].indicatorValues, latestVal)
	}

	// Build result compatible with PerformanceIndicators JSON tags
	result := make([]map[string]interface{}, 0, len(grouped))
	for _, id := range idOrder {
		entry := grouped[id]
		result = append(result, map[string]interface{}{
			"object_id":        id,
			"indicators":       entry.indicators,
			"indicator_values": entry.indicatorValues,
		})
	}

	return result
}
