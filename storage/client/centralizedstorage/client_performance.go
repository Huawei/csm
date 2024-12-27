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
	"fmt"
	"strings"

	"github.com/huawei/csm/v2/storage/api/centralizedstorage"
	"github.com/huawei/csm/v2/storage/httpcode"
	"github.com/huawei/csm/v2/utils/log"
)

// GetPerformance query storage performance
func (c *CentralizedClient) GetPerformance(ctx context.Context, objectType int,
	indicators []int) ([]map[string]interface{}, error) {
	var temp = make([]string, len(indicators))
	for k, v := range indicators {
		temp[k] = fmt.Sprintf("%d", v)
	}
	var indicatorsParam = "[" + strings.Join(temp, ",") + "]"

	data := map[string]interface{}{
		"objectType": objectType,
		"indicators": indicatorsParam,
	}

	url, err := centralizedstorage.GenerateUrl("PerformanceData", data)
	if err != nil {
		log.AddContext(ctx).Errorf("get system performance url error: %v", err)
		return nil, err
	}

	callFunc := func() ([]map[string]interface{}, *float64, error) {
		resp, err := c.get(ctx, url, nil)
		if err != nil {
			log.AddContext(ctx).Errorf("get performance error: %v", err)
			return nil, nil, err
		}

		return c.getResultListFromResponseList(ctx, resp)
	}
	return c.Client.RetryListCall(ctx, httpcode.RetryCodes, callFunc)
}

// GetPerformanceByPost query storage performance by post
func (c *CentralizedClient) GetPerformanceByPost(ctx context.Context, objectType int,
	indicators []int) ([]map[string]interface{}, error) {
	data := map[string]interface{}{
		"object_type": objectType,
		"indicators":  indicators,
	}

	url, err := centralizedstorage.GenerateUrl("PerformanceDataPost", nil)
	if err != nil {
		log.AddContext(ctx).Errorf("get system performance url error: %v", err)
		return nil, err
	}

	callFunc := func() ([]map[string]interface{}, *float64, error) {
		resp, err := c.post(ctx, url, data)
		if err != nil {
			log.AddContext(ctx).Errorf("get performance by post error: %v", err)
			return nil, nil, err
		}

		return c.getResultListFromResponseList(ctx, resp)
	}
	return c.Client.RetryListCall(ctx, httpcode.RetryCodes, callFunc)
}
