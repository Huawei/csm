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

	"github.com/huawei/csm/v2/storage/api/centralizedstorage"
	"github.com/huawei/csm/v2/storage/httpcode"
	"github.com/huawei/csm/v2/utils/log"
)

// GetStoragePools is used to get storage pools
func (c *CentralizedClient) GetStoragePools(ctx context.Context) ([]map[string]interface{}, error) {
	return c.GetByUrl(ctx, "GetStoragePools")
}

// GetControllers is used to get storage controllers
func (c *CentralizedClient) GetControllers(ctx context.Context) ([]map[string]interface{}, error) {
	return c.GetByUrl(ctx, "GetControllers")
}

// GetByUrl is used to query storage information based on a specified URL, requiring no parameters when querying
func (c *CentralizedClient) GetByUrl(ctx context.Context, urlKey string) ([]map[string]interface{}, error) {
	data := map[string]interface{}{}

	url, err := centralizedstorage.GenerateUrl(urlKey, data)
	if err != nil {
		log.AddContext(ctx).Errorf("get url failed, url: %s, error: %v", urlKey, err)
		return nil, err
	}

	callFunc := func() ([]map[string]interface{}, *float64, error) {
		resp, err := c.get(ctx, url, nil)
		if err != nil {
			log.AddContext(ctx).Errorf("get by url failed, url: %s error: %v", urlKey, err)
			return nil, nil, err
		}

		return c.getResultListFromResponseList(ctx, resp)
	}

	return c.Client.RetryListCall(ctx, httpcode.RetryCodes, callFunc)
}
