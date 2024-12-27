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
	"github.com/huawei/csm/v2/storage/httpcode/filesystem"
	"github.com/huawei/csm/v2/utils/log"
)

// GetFileSystemByName get filesystem by name
func (c *CentralizedClient) GetFileSystemByName(ctx context.Context, name string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"fsName": name,
	}

	url, err := centralizedstorage.GenerateUrl("GetFileSystemByName", data)
	if err != nil {
		log.AddContext(ctx).Errorf("storage client get filesystem by name generate url error: %v", err)
		return nil, err
	}

	callFunc := func() (map[string]interface{}, *float64, error) {
		resp, err := c.get(ctx, url, nil)
		if err != nil {
			log.AddContext(ctx).Errorf("storage client get filesystem by name error: %v", err)
			return nil, nil, err
		}

		return c.getResultFromResponseList(ctx, resp)
	}

	return c.Client.RetryCall(ctx, filesystem.GetRetryCodes(), callFunc)
}

// GetFileSystemIdByName get filesystem id by name
func (c *CentralizedClient) GetFileSystemIdByName(ctx context.Context, name string) (string, error) {
	data, err := c.GetFileSystemByName(ctx, name)
	if err != nil {
		return "", err
	}
	id, ok := data["ID"].(string)
	if !ok {
		return "", nil
	}
	return id, nil
}

// GetFilesystem is used to get filesystems
func (c *CentralizedClient) GetFilesystem(ctx context.Context, start, end int) ([]map[string]interface{}, error) {
	return c.pageQuery(ctx, start, end, "GetFilesystem")
}

// GetFilesystemCount used to get filesystem count
func (c *CentralizedClient) GetFilesystemCount(ctx context.Context) (int, error) {
	return c.countQuery(ctx, "GetFilesystemCount")
}
