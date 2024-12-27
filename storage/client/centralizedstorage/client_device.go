/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

// Package centralizedstorage is related with storage client
package centralizedstorage

import (
	"context"

	"github.com/huawei/csm/v2/storage/api/centralizedstorage"
	"github.com/huawei/csm/v2/utils/log"
)

// GetSystemInfo is used to get system info
func (c *CentralizedClient) GetSystemInfo(ctx context.Context) (map[string]interface{}, error) {
	url, err := centralizedstorage.GenerateUrl("GetSystemInfo", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.get(ctx, url, nil)
	if err != nil {
		log.AddContext(ctx).Errorf("storage client get system info error: %v", err)
		return nil, err
	}

	result, _, err := c.getResultFromResponse(ctx, resp)
	return result, err
}
