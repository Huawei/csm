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

// Package distributedstorage provides a client for FusionStorage (distributed storage) APIs.
package distributedstorage

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/huawei/csm/v2/storage/api/distributedstorage"
	"github.com/huawei/csm/v2/storage/client"
	"github.com/huawei/csm/v2/storage/constant"
	"github.com/huawei/csm/v2/storage/utils"
	"github.com/huawei/csm/v2/utils/log"
)

const (
	// v2LoginSuccessCode is the success code for v2 login API.
	v2LoginSuccessCode = 0
	decimalBase        = 10
	int64BitSize       = 64
	float64BitSize     = 64
)

// objectTypeInfo holds the query info needed to fetch object IDs for performance queries.
// For CollectModeList, ListKey is set; for CollectModePaginated, CountKey+PageKey are set.
type objectTypeInfo struct {
	ListKey  string // CollectModeList: list URL key (e.g., "GetControllers")
	CountKey string // CollectModePaginated: count URL key (e.g., "GetLunCount")
	PageKey  string // CollectModePaginated: page URL key (e.g., "GetLuns")
	IDKey    string // field name for object ID in response data
}

// objectTypeMapping is a global registry of object type ID to query info.
// It is populated by RegisterObjectType (called from MustRegister during init),
// and only read at runtime. This init-write + runtime-read pattern is safe in Go
// without additional synchronization. If runtime registration is needed in the future,
// switch to sync.Map or add sync.RWMutex protection.
var objectTypeMapping = map[int]objectTypeInfo{}

// RegisterObjectType registers the query info for a given object type ID.
// Set listKey for CollectModeList, or countKey+pageKey for CollectModePaginated.
func RegisterObjectType(typeID int, listKey, countKey, pageKey, idKey string) {
	objectTypeMapping[typeID] = objectTypeInfo{
		ListKey:  listKey,
		CountKey: countKey,
		PageKey:  pageKey,
		IDKey:    idKey,
	}
}

// DistributedClient implements RestClient for FusionStorage backends.
// It embeds client.Client to reuse HTTP request/response logic and URL management.
type DistributedClient struct {
	client.Client
	ESN string
}

// NewDistributedClient creates a new DistributedClient.
func NewDistributedClient(ctx context.Context, config *constant.StorageBackendConfig) (*DistributedClient, error) {
	dc := &DistributedClient{
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

	dc.SetAuthHeaders = dc.setAuthHeaders

	if err := dc.Client.InitHttpClient(ctx); err != nil {
		return nil, err
	}

	return dc, nil
}

// setAuthHeaders sets FusionStorage-specific authentication headers.
func (c *DistributedClient) setAuthHeaders(req *http.Request) {
	if c.Token != "" {
		req.Header.Set("X-Auth-Token", c.Token)
		req.Header.Set("Cookie", "X_Auth_Token="+c.Token)
	}
}

// Login authenticates using the v2 login API and caches Token, ESN.
// On success, c.Curl is set to the successful URL for subsequent requests.
func (c *DistributedClient) Login(ctx context.Context) error {
	passwordBytes, err := c.getSecret(ctx)
	if err != nil {
		return fmt.Errorf("get login secret failed: %w", err)
	}

	reqBody := map[string]interface{}{
		"user_name": c.User,
		"password":  string(passwordBytes),
	}

	// Zero password bytes immediately after converting to string for the request
	for i := range passwordBytes {
		passwordBytes[i] = 0
	}

	apiPath, err := distributedstorage.GenerateUrl("Login", nil)
	if err != nil {
		return fmt.Errorf("generate login url failed: %w", err)
	}

	for _, rawURL := range c.Urls {
		c.Curl = rawURL
		fullURL := buildFullURL(rawURL, apiPath)
		resp, err := c.Call(ctx, http.MethodPost, fullURL, reqBody)
		if err != nil {
			log.AddContext(ctx).Errorf("login request failed, url: %s, err: %v", fullURL, err)
			continue
		}

		_, err = checkResponseCode(resp)
		if err != nil {
			log.AddContext(ctx).Warningf("check login response code failed, url: %s, err: %v", fullURL, err)
			continue
		}

		data, ok := resp["data"].(map[string]interface{})
		if !ok {
			log.AddContext(ctx).Warningf("login response missing data field for url %s", fullURL)
			continue
		}

		c.Token, _ = utils.GetValue[string](data, "x_auth_token")

		c.ESN, _ = utils.GetValue[string](data, "system_esn")

		// Clear password from request body after successful login
		reqBody["password"] = ""

		log.AddContext(ctx).Infof("login success, backend: %s, url: %s", c.StorageBackendName, c.Curl)
		return nil
	}

	reqBody["password"] = ""
	c.Token = ""
	c.ESN = ""
	c.Curl = ""
	return fmt.Errorf("login failed for all urls")
}

// Logout deletes the session using the v2 logout API.
func (c *DistributedClient) Logout(ctx context.Context) {
	logoutURL, err := distributedstorage.GenerateUrl("Logout", nil)
	if err != nil {
		log.AddContext(ctx).Errorf("generate logout url failed: %v", err)
		return
	}

	fullURL := buildFullURL(c.Curl, logoutURL)
	if _, err := c.Call(ctx, http.MethodDelete, fullURL, nil); err != nil {
		log.AddContext(ctx).Warningf("logout request failed: %v", err)
	}

	log.AddContext(ctx).Infof("logout success, backend: %s, url: %s", c.StorageBackendName, c.Curl)
}

// normalizeFieldNames converts all keys to UPPER_CASE in the data map.
func normalizeFieldNames(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(data))
	for key, value := range data {
		upperKey := strings.ToUpper(key)
		result[upperKey] = value
	}
	return result
}

// normalizeFieldNamesSlice normalizes each map in a slice.
func normalizeFieldNamesSlice(data []map[string]interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, len(data))
	for i, item := range data {
		result[i] = normalizeFieldNames(item)
	}
	return result
}

// checkResponseCode check the result.code from a v2 API response.
func checkResponseCode(resp map[string]interface{}) (float64, error) {
	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("can not convert resp result to map, resp: %v", result)
	}

	code, ok := utils.GetValue[float64](result, "code")
	if !ok {
		return 0, fmt.Errorf("get code failed from response result, result: %v", result)
	}

	if code != v2LoginSuccessCode {
		return code, fmt.Errorf("response code is %v", code)
	}

	return code, nil
}

// getSecret retrieves the password bytes from the Kubernetes Secret with SBC CRD fallback.
func (c *DistributedClient) getSecret(ctx context.Context) ([]byte, error) {
	secret, err := c.Client.GetSecretWithFallback(ctx)
	if err != nil {
		return nil, err
	}

	password, exist := secret.Data["password"]
	if !exist {
		return nil, fmt.Errorf("password key not found in secret [%s/%s]", c.SecretNamespace, c.SecretName)
	}
	return password, nil
}
