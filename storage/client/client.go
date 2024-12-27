/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2024. All rights reserved.

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

// Package client is related with storage common client and operation
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/huawei/csm/v2/storage/utils"
	"github.com/huawei/csm/v2/utils/log"
)

const (
	// 20000 is the approximate characters number of one page
	// If a log exceeds one page, it will be compressed
	charLimit = 20000

	sessionsSubStr = "/sessions"
)

// Client is used to extract storage common attribute
type Client struct {
	Curl     string
	Urls     []string
	User     string
	DeviceId string
	Token    string
	VStore   string
	Client   HttpClient

	SecretNamespace string
	SecretName      string

	StorageBackendNamespace string
	StorageBackendName      string

	ReLoginMutex sync.Mutex
	Semaphore    *utils.Semaphore
}

// HttpClient is used to define http interface
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Call is used to remote call storage interfaces
func (c *Client) Call(ctx context.Context, method string, url string,
	reqData map[string]interface{}) (map[string]interface{}, error) {
	if !strings.Contains(url, sessionsSubStr) {
		log.AddContext(ctx).Infof("call request %s %s, request: %v", method, url, reqData)
	}
	log.AddContext(ctx).Infof("call reloginLock: %v", c.ReLoginMutex)

	req, err := c.getRequest(ctx, method, url, reqData)
	if err != nil {
		log.AddContext(ctx).Errorf(
			"client http request error, method: %s, url: %s, error: %v", method, url, err)
		return nil, err
	}

	resp, err := c.getResponse(ctx, req)
	if err != nil {
		log.AddContext(ctx).Errorf("client http response error, method: %s, url: %s, error: %v",
			method, url, err)
		return nil, err
	}

	if !strings.Contains(url, sessionsSubStr) {
		responseStr := fmt.Sprintf("%v", resp)
		if len(responseStr) > charLimit {
			compressedStr, err := utils.CompressStr(responseStr)
			if err != nil {
				log.AddContext(ctx).Warningf("compress storage response fail, "+
					"the log will be printed without compression, err: %v", err)
			}
			log.AddContext(ctx).Infof("call response %s %s, response compressed by deflate algorithm: %s",
				method, url, compressedStr)
			log.AddContext(ctx).Debugf("call response %s %s, response: %s", method, url, responseStr)
		} else {
			log.AddContext(ctx).Infof("call response %s %s, response: %s", method, url, responseStr)
		}
	}
	return resp, nil
}

// RetryCall is used to retry remote call storage interfaces
func (c *Client) RetryCall(ctx context.Context, retryCodes []float64,
	call func() (map[string]interface{}, *float64, error)) (map[string]interface{}, error) {
	var err error
	var code *float64
	var respData map[string]interface{}

	retryFunc := func() bool {
		respData, code, err = call()
		// if code not exist, then do not retry
		if code == nil {
			return false
		}

		if err != nil {
			if !utils.IsFloat64InList(retryCodes, *code) {
				return false
			} else {
				log.AddContext(ctx).Infoln("storage client retry call...")
				return true
			}
		}

		return false
	}

	utils.RetryCallFunc(retryFunc)

	return respData, err
}

// RetryListCall is used to retry remote call storage interfaces
func (c *Client) RetryListCall(ctx context.Context, retryCodes []float64,
	call func() ([]map[string]interface{}, *float64, error)) ([]map[string]interface{}, error) {
	var err error
	var code *float64
	var respData []map[string]interface{}

	retryFunc := func() bool {
		respData, code, err = call()
		// if code not exist, then do not retry
		if code == nil {
			return false
		}

		if err != nil {
			if !utils.IsFloat64InList(retryCodes, *code) {
				return false
			} else {
				log.AddContext(ctx).Infoln("storage client retry list call...")
				return true
			}
		}

		return false
	}

	utils.RetryCallFunc(retryFunc)

	return respData, err
}

func (c *Client) getRequest(ctx context.Context, method string, url string,
	reqData map[string]interface{}) (*http.Request, error) {
	log.AddContext(ctx).Debugln("get request start...")
	defer log.AddContext(ctx).Debugln("get request end...")

	reqBody, err := c.getRequestBody(ctx, url, reqData)

	if err != nil {
		log.AddContext(ctx).Errorf("client http request body error, url: %s, error: %s", url, err.Error())
		return nil, err
	}

	req, err := c.newRequest(ctx, method, url, reqBody)
	if err != nil {
		log.AddContext(ctx).Errorf("client http new request error, url: %s, error: %s", url, err.Error())
		return nil, err
	}

	return req, nil
}

func (c *Client) getResponse(ctx context.Context, req *http.Request) (map[string]interface{}, error) {
	log.AddContext(ctx).Debugln("get response start...")
	defer log.AddContext(ctx).Debugln("get response end...")

	clientResp, err := c.Client.Do(req)
	if err != nil {
		log.AddContext(ctx).Errorf("client http response error, error: %v", err)
		return nil, err
	}
	defer clientResp.Body.Close()

	log.AddContext(ctx).Debugln("start read response body...")
	body, err := ioutil.ReadAll(clientResp.Body)
	if err != nil {
		log.AddContext(ctx).Errorf("client read response body error: %v", err)
		return nil, err
	}
	log.AddContext(ctx).Debugln("read response body success...")

	var resp map[string]interface{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.AddContext(ctx).Errorf("client response body convert response error, body: %s, error: %v", body, err)
		return nil, err
	}

	return resp, nil
}

func (c *Client) getRequestBody(ctx context.Context, url string, reqData map[string]interface{}) (io.Reader, error) {
	log.AddContext(ctx).Debugln("get request body start...")
	defer log.AddContext(ctx).Debugln("get request body end...")

	if reqData == nil {
		return nil, nil
	}

	reqBytes, err := json.Marshal(reqData)
	if err != nil {
		if strings.Contains(url, sessionsSubStr) {
			log.AddContext(ctx).Errorf("client http request body error: %v", err)
		} else {
			log.AddContext(ctx).Errorf("client http request body error, data: %v, error: %v", reqData, err)
		}

		return nil, err
	}

	return bytes.NewReader(reqBytes), nil
}

func (c *Client) newRequest(ctx context.Context, method string, reqUrl string,
	reqBody io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, reqUrl, reqBody)
	if err != nil {
		log.AddContext(ctx).Errorf("client http new request error: %s", err.Error())
		return req, err
	}

	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")

	if c.Token != "" {
		req.Header.Set("iBaseToken", c.Token)
	}

	return req, nil
}
