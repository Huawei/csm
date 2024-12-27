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

package centralizedstorage

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/huawei/csm/v2/storage/api/centralizedstorage"
	"github.com/huawei/csm/v2/storage/httpcode"
	"github.com/huawei/csm/v2/storage/httpcode/filesystem"
	"github.com/huawei/csm/v2/utils/log"
)

const (
	decimalBase  = 10
	int64BitSize = 64
)

// GetLuns is used to get luns information
func (c *CentralizedClient) GetLuns(ctx context.Context, start, end int) ([]map[string]interface{}, error) {
	return c.pageQuery(ctx, start, end, "GetLuns")
}

// GetLunCount used to get lun count
func (c *CentralizedClient) GetLunCount(ctx context.Context) (int, error) {
	return c.countQuery(ctx, "GetLunCount")
}

// GetLunByName used to get lun by name
func (c *CentralizedClient) GetLunByName(ctx context.Context, name string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"lunName": name,
	}

	url, err := centralizedstorage.GenerateUrl("GetLunByName", data)
	if err != nil {
		log.AddContext(ctx).Errorf("storage client get lun by name generate url error: %v", err)
		return nil, err
	}

	callFunc := func() (map[string]interface{}, *float64, error) {
		resp, err := c.get(ctx, url, nil)
		if err != nil {
			log.AddContext(ctx).Errorf("storage client get lun by name error: %v", err)
			return nil, nil, err
		}

		return c.getResultFromResponseList(ctx, resp)
	}

	return c.Client.RetryCall(ctx, filesystem.GetRetryCodes(), callFunc)
}

// GetLunIdByName used to get lun id by name
func (c *CentralizedClient) GetLunIdByName(ctx context.Context, name string) (string, error) {
	data, err := c.GetLunByName(ctx, name)
	if err != nil {
		return "", err
	}
	id, ok := data["ID"].(string)
	if !ok {
		return "", nil
	}
	return id, nil
}

// countQuery used to query count information
func (c *CentralizedClient) countQuery(ctx context.Context, urlKey string) (int, error) {
	data := map[string]interface{}{}
	url, err := centralizedstorage.GenerateUrl(urlKey, data)
	if err != nil {
		return 0, err
	}

	callFunc := func() (map[string]interface{}, *float64, error) {
		resp, err := c.get(ctx, url, nil)
		if err != nil {
			log.AddContext(ctx).Errorf("count query failed, url: %s error: %v", urlKey, err)
			return nil, nil, err
		}

		return c.getResultFromResponse(ctx, resp)
	}

	result, err := c.Client.RetryCall(ctx, httpcode.RetryCodes, callFunc)
	if err != nil {
		return 0, err
	}
	count, ok := result["COUNT"].(string)
	if !ok {
		return 0, errors.New(fmt.Sprintf("%s not found count, return result is %v", urlKey, result))
	}

	parseInt, err := strconv.ParseInt(count, decimalBase, int64BitSize)
	if err != nil {
		return 0, err
	}

	return int(parseInt), nil
}

// pageQuery is used to page query
func (c *CentralizedClient) pageQuery(ctx context.Context, start, end int,
	urlKey string) ([]map[string]interface{}, error) {

	data := map[string]interface{}{
		"start": start,
		"end":   end,
	}

	url, err := centralizedstorage.GenerateUrl(urlKey, data)
	if err != nil {
		return nil, err
	}

	callFunc := func() ([]map[string]interface{}, *float64, error) {
		resp, err := c.get(ctx, url, nil)
		if err != nil {
			log.AddContext(ctx).Errorf("page query failed, url: %s error: %v", urlKey, err)
			return nil, nil, err
		}

		return c.getResultListFromResponseList(ctx, resp)
	}

	return c.Client.RetryListCall(ctx, httpcode.RetryCodes, callFunc)
}
