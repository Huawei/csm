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
	"net/http"

	"github.com/huawei/csm/v2/storage/client"
	"github.com/huawei/csm/v2/storage/constant"
	"github.com/huawei/csm/v2/storage/utils"
)

// CentralizedClient is used to use centralized storage related functions
type CentralizedClient struct {
	client.Client
}

// NewCentralizedClient is used to new centralized storage client
func NewCentralizedClient(ctx context.Context, config *constant.StorageBackendConfig) (*CentralizedClient, error) {
	centralizedClient := &CentralizedClient{
		Client: client.Client{
			Urls:                    config.Urls,
			User:                    config.User,
			SecretNamespace:         config.SecretNamespace,
			SecretName:              config.SecretName,
			StorageBackendNamespace: config.StorageBackendNamespace,
			StorageBackendName:      config.StorageBackendName,
			Semaphore:               utils.NewSemaphore(config.ClientMaxThreads),
		},
	}
	centralizedClient.SetAuthHeaders = centralizedClient.setAuthHeaders

	if err := centralizedClient.Client.InitHttpClient(ctx); err != nil {
		return nil, err
	}

	return centralizedClient, nil
}

// setAuthHeaders sets OceanStor-specific authentication headers.
func (c *CentralizedClient) setAuthHeaders(req *http.Request) {
	if c.Token != "" {
		req.Header.Set("iBaseToken", c.Token)
	}
}
