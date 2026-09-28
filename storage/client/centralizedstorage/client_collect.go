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

package centralizedstorage

import (
	"context"
	"time"

	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/provider/utils"
	"github.com/huawei/csm/v2/storage/api/centralizedstorage"
	"github.com/huawei/csm/v2/storage/httpcode"
	"github.com/huawei/csm/v2/utils/log"
)

const (
	performanceRetryCount    = 5
	performanceRetryInterval = 5 * time.Second
)

// GetSingleByUrlKey queries a single object by URL key.
// Unlike GetListByUrlKey, this method uses getResultFromResponse which expects
// response.Data to be a map[string]interface{} (single object) rather than []interface{} (list).
func (c *CentralizedClient) GetSingleByUrlKey(ctx context.Context, urlKey string) (map[string]interface{}, error) {
	data := map[string]interface{}{}

	url, err := centralizedstorage.GenerateUrl(urlKey, data)
	if err != nil {
		log.AddContext(ctx).Errorf("get url failed, url: %s, error: %v", urlKey, err)
		return nil, err
	}

	callFunc := func() (map[string]interface{}, *float64, error) {
		resp, err := c.get(ctx, url, nil)
		if err != nil {
			log.AddContext(ctx).Errorf("get by url failed, url: %s error: %v", urlKey, err)
			return nil, nil, err
		}

		return c.getResultFromResponse(ctx, resp)
	}

	result, err := c.Client.RetryCall(ctx, httpcode.RetryCodes, callFunc)
	return result, err
}

// GetListByUrlKey queries a list of objects by URL key.
func (c *CentralizedClient) GetListByUrlKey(ctx context.Context, urlKey string) ([]map[string]interface{}, error) {
	return c.GetByUrl(ctx, urlKey)
}

// GetCountByUrlKey queries the count of objects by URL key.
func (c *CentralizedClient) GetCountByUrlKey(ctx context.Context, urlKey string) (int, error) {
	return c.countQuery(ctx, urlKey)
}

// GetPageByUrlKey queries a page of objects by URL key.
func (c *CentralizedClient) GetPageByUrlKey(ctx context.Context,
	urlKey string, start, end int) ([]map[string]interface{}, error) {
	return c.pageQuery(ctx, start, end, urlKey)
}

// QueryPerformanceData queries performance data from OceanStorage.
// It checks the storage version to decide between POST and GET request modes,
// and handles retries for older V6 firmware that may return empty data.
func (c *CentralizedClient) QueryPerformanceData(ctx context.Context,
	objectType int, indicators []string) ([]map[string]interface{}, error) {
	storageInfo, err := c.GetSystemInfo(ctx)
	if err != nil {
		log.AddContext(ctx).Errorf("get storage system info failed, error: %v", err)
		return nil, err
	}

	intIndicators := utils.MapStringToInt(indicators)

	var mapData []map[string]interface{}
	var postEnable bool
	version, ok := storageInfo["pointRelease"].(string)
	if !ok {
		// Storage of V3 or V5 does not have the pointRelease field
		postEnable = false
	}

	if utils.CompareVersions(version, constants.MinVersionSupportPost) != -1 {
		// 6.1.2 and later versions support the Post request
		postEnable = true
	}

	if postEnable {
		mapData, err = c.GetPerformanceByPost(ctx, objectType, intIndicators)
	} else {
		for i := 0; i < performanceRetryCount; i++ {
			mapData, err = c.GetPerformance(ctx, objectType, intIndicators)
			// For storage v6 earlier 6.1.2, if it can not return the performance data caused by concurrency,
			// both the mapData and err are nil. But in the same conditions for storage v3 or v5, the mapData
			// is nil while the err is not nil.
			if err != nil {
				break
			}
			if len(mapData) != 0 {
				break
			}
			time.Sleep(performanceRetryInterval)
		}
	}
	if err != nil {
		log.AddContext(ctx).Errorf("invoke the get performance method of storage client failed, error: %v", err)
		return nil, err
	}

	// For storage v6 earlier 6.1.2, the storage may return empty data even after 5 time retries.
	if len(mapData) == 0 {
		log.AddContext(ctx).Warningln("get empty data by the get performance method of storage client")
	}

	return mapData, nil
}
