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
	"fmt"
	"strings"

	"github.com/huawei/csm/v2/utils/log"
)

const (
	// httpUnauthorizedCode is the response code indicating HTTP unauthorized.
	httpUnauthorizedCode float64 = -401
	// offLineCode is the response code indicating the storage system is offline.
	offLineCode float64 = 1077949069
	// noAuthenticatedCode is the response code indicating the session is not authenticated.
	noAuthenticatedCode float64 = 10000003
)

// callWithRetry executes an authenticated request with ReLogin support.
// If the response indicates unauthorized (v2UnauthorizedCode), it re-logs in and retries.
// If an x509 error occurs, it reinitializes the HTTP client and retries.
func (c *DistributedClient) callWithRetry(ctx context.Context, method, fullURL string,
	reqData map[string]interface{}) (map[string]interface{}, error) {

	resp, err := c.Call(ctx, method, fullURL, reqData)
	if err != nil {
		if strings.Contains(err.Error(), "x509") {
			if initErr := c.Client.InitHttpClient(ctx); initErr != nil {
				return nil, initErr
			}
			resp, err = c.Call(ctx, method, fullURL, reqData)
		}
		if err != nil {
			return nil, err
		}
	}

	code, err := checkResponseCode(resp)
	if c.needRetry(code) {
		log.AddContext(ctx).Infoln("token expired, attempting re-login")
		if reLoginErr := c.reLogin(ctx); reLoginErr != nil {
			return nil, fmt.Errorf("re-login failed: %w", reLoginErr)
		}
		return c.Call(ctx, method, fullURL, reqData)
	}

	if err != nil {
		return nil, fmt.Errorf("call %s %s check response code failed, err: %v", method, fullURL, err)
	}

	return resp, nil
}

func (c *DistributedClient) needRetry(code float64) bool {
	return code == httpUnauthorizedCode || code == offLineCode || code == noAuthenticatedCode
}

// reLogin performs a mutex-protected re-login.
// It follows the double-check pattern: if another goroutine has already re-logged in
// (token changed), this goroutine skips the re-login.
func (c *DistributedClient) reLogin(ctx context.Context) error {
	oldToken := c.Token

	c.ReLoginMutex.Lock()
	defer c.ReLoginMutex.Unlock()

	// Double-check: another goroutine may have already re-logged in
	if c.Token != "" && oldToken != c.Token {
		log.AddContext(ctx).Infoln("another goroutine already re-logged in, skipping")
		return nil
	}

	c.Logout(ctx)
	if err := c.Login(ctx); err != nil {
		log.AddContext(ctx).Errorf("re-login failed: %v", err)
		return fmt.Errorf("re-login failed: %w", err)
	}

	log.AddContext(ctx).Infoln("re-login successful")
	return nil
}

// buildFullURL concatenates a base URL with an API path.
func buildFullURL(baseURL, apiPath string) string {
	return strings.TrimRight(baseURL, "/") + apiPath
}
