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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/huawei/csm/v2/storage/httpcode"
	"github.com/huawei/csm/v2/utils/log"
)

// Response is used to receive storage response when client remote call storage interfaces
type Response struct {
	Error map[string]interface{} `json:"error"`
	Data  interface{}            `json:"data,omitempty"`
}

func (c *CentralizedClient) get(ctx context.Context, methodUrl string,
	reqData map[string]interface{}) (*Response, error) {
	return c.callCentralizedStorage(ctx, "GET", methodUrl, reqData)
}

func (c *CentralizedClient) post(ctx context.Context, methodUrl string,
	reqData map[string]interface{}) (*Response, error) {
	return c.callCentralizedStorage(ctx, "POST", methodUrl, reqData)
}

func (c *CentralizedClient) delete(ctx context.Context, methodUrl string,
	reqData map[string]interface{}) (*Response, error) {
	return c.callCentralizedStorage(ctx, "DELETE", methodUrl, reqData)
}

func (c *CentralizedClient) put(ctx context.Context, methodUrl string,
	reqData map[string]interface{}) (*Response, error) {
	return c.callCentralizedStorage(ctx, "PUT", methodUrl, reqData)
}

func (c *CentralizedClient) callCentralizedStorage(ctx context.Context, method string,
	methodUrl string, reqData map[string]interface{}) (*Response, error) {
	if methodUrl == "/sessions" || methodUrl == "/xx/sessions" {
		return c.baseCall(ctx, method, methodUrl, reqData)
	}

	response, err := c.baseCall(ctx, method, methodUrl, reqData)
	if err != nil {
		return c.reLoginCall(ctx, method, methodUrl, reqData)
	}

	code, _ := c.checkResponseCode(ctx, response)
	log.AddContext(ctx).Infof("call check response code: %v", code)
	if code != nil && *code == httpcode.NoAuthentication {
		log.AddContext(ctx).Infof("%v no authentication, need reLogin", code)
		return c.reLoginCall(ctx, method, methodUrl, reqData)
	}

	return response, err
}

func (c *CentralizedClient) baseCall(ctx context.Context, method string,
	methodUrl string, reqData map[string]interface{}) (*Response, error) {
	c.Semaphore.Acquire()
	defer c.Semaphore.Release()
	log.AddContext(ctx).Infof("%s call semaphore: %d", c.Curl, c.Semaphore.AvailablePermits())

	url := c.getRequestUrl(methodUrl)
	response, err := c.Call(ctx, method, url, reqData)
	if err != nil && strings.Contains(err.Error(), "x509") {
		if err = c.initHttpClient(ctx); err != nil {
			return nil, err
		}

		response, err = c.Call(ctx, method, url, reqData)
	}
	if err != nil {
		return nil, err
	}

	return c.convertToCallResponse(ctx, response)
}

func (c *CentralizedClient) reLoginCall(ctx context.Context, method string,
	methodUrl string, reqData map[string]interface{}) (*Response, error) {
	log.AddContext(ctx).Infof("storage client reLogin call start. method: %s, url: %s", method, methodUrl)
	defer log.AddContext(ctx).Infof("storage client reLogin call success. method: %s, url: %s", method, methodUrl)

	if err := c.ReLogin(ctx); err != nil {
		return nil, err
	}

	return c.baseCall(ctx, method, methodUrl, reqData)
}

func (c *CentralizedClient) getRequestUrl(methodUrl string) string {
	// If the API is api/v2, need to reconstruct the request URL. The differences are as follows:
	// default c.Curl is: 'https://${ip}:${port}/deviceManager/rest/${deviceId}/'
	// api/v2 real url is: 'https://${ip}:${port}/api/v2/remote_execute'
	if strings.HasPrefix(methodUrl, "/api/v2") {
		urlSplit := strings.Split(c.Curl, "/deviceManager/rest")
		if len(urlSplit) < 1 {
			return c.Curl + methodUrl
		}
		return urlSplit[0] + methodUrl
	}

	if c.DeviceId != "" && methodUrl != "/xx/sessions" {
		return c.Curl + "/" + c.DeviceId + methodUrl
	}

	return c.Curl + methodUrl
}

func (c *CentralizedClient) convertToCallResponse(ctx context.Context,
	response map[string]interface{}) (*Response, error) {
	jsData, err := json.Marshal(response)
	if err != nil {
		msg := fmt.Sprintf("storage client call response to json error: %v", err)
		log.AddContext(ctx).Errorln(msg)
		return nil, errors.New(msg)
	}

	var resp Response
	err = json.Unmarshal(jsData, &resp)
	if err != nil {
		msg := fmt.Sprintf("storage client call json to response error: %v", err)
		log.AddContext(ctx).Errorln(msg)
		return nil, errors.New(msg)
	}

	return &resp, nil
}

func (c *CentralizedClient) getResultFromResponseList(ctx context.Context,
	response *Response) (map[string]interface{}, *float64, error) {
	respCode, err := c.checkResponseCode(ctx, response)
	if err != nil {
		return nil, respCode, err
	}

	if response.Data == nil {
		log.AddContext(ctx).Infoln("find response data is nil")
		return nil, respCode, nil
	}

	respData, exist := response.Data.([]interface{})
	if !exist {
		msg := fmt.Sprintf(
			"storage client response data can not convert to []interface{}, response data: %v", response.Data)
		log.AddContext(ctx).Errorln(msg)
		return nil, respCode, errors.New(msg)
	}

	if len(respData) == 0 {
		log.AddContext(ctx).Infoln("storage client find response data list is empty")
		return nil, respCode, nil
	}

	if len(respData) > 1 {
		msg := fmt.Sprintf("storage client find more than one data in response data list: %v", respData)
		log.AddContext(ctx).Errorf(msg)
		return nil, respCode, errors.New(msg)
	}

	data, exist := respData[0].(map[string]interface{})
	if !exist {
		msg := fmt.Sprintf(
			"storage client response data can not convert to map[string]interface{}, response data: %v", respData[0])
		log.AddContext(ctx).Errorln(msg)
		return nil, respCode, errors.New(msg)
	}

	return data, respCode, nil
}

func (c *CentralizedClient) getResultListFromResponseList(ctx context.Context,
	response *Response) ([]map[string]interface{}, *float64, error) {
	respCode, err := c.checkResponseCode(ctx, response)
	if err != nil {
		return nil, respCode, err
	}

	if response.Data == nil {
		log.AddContext(ctx).Infoln("find response data is nil")
		return nil, respCode, nil
	}

	respData, exist := response.Data.([]interface{})
	if !exist {
		msg := fmt.Sprintf("response data list can not convert to []interface{}, data: %v", response.Data)
		log.AddContext(ctx).Errorln(msg)
		return nil, respCode, errors.New(msg)
	}

	if len(respData) == 0 {
		log.AddContext(ctx).Infoln("find response data list is empty")
		return nil, respCode, nil
	}

	var resultList []map[string]interface{}
	for _, data := range respData {
		result, exist := data.(map[string]interface{})
		if !exist {
			msg := fmt.Sprintf("response data can not convert to map[string]interface{}, data: %v", data)
			log.AddContext(ctx).Errorln(msg)
			return nil, respCode, errors.New(msg)
		}

		resultList = append(resultList, result)
	}

	return resultList, respCode, nil
}

func (c *CentralizedClient) getResultFromResponse(ctx context.Context,
	response *Response) (map[string]interface{}, *float64, error) {
	respCode, err := c.checkResponseCode(ctx, response)
	if err != nil {
		return nil, respCode, err
	}

	if response.Data == nil {
		log.AddContext(ctx).Infoln("find response data is nil")
		return nil, respCode, nil
	}

	respData, exist := response.Data.(map[string]interface{})
	if !exist {
		msg := fmt.Sprintf(
			"storage client response data can not convert to map[string]interface{}, response data: %v", response.Data)
		log.AddContext(ctx).Errorln(msg)
		return nil, respCode, errors.New(msg)
	}

	return respData, respCode, nil
}

func (c *CentralizedClient) checkResponseCode(ctx context.Context, response *Response) (*float64, error) {
	respCode, exist := response.Error["code"].(float64)
	if !exist {
		msg := fmt.Sprintf("storage client response httpcode does not exist, response: %v", response)
		log.AddContext(ctx).Errorf(msg)
		return nil, errors.New(msg)
	}

	if respCode != httpcode.SuccessCode {
		msg := fmt.Sprintf("storage client response httpcode is not success code, "+
			"code: %v, description: %v", respCode, response.Error["description"])
		log.AddContext(ctx).Errorf(msg)
		return &respCode, errors.New(msg)
	}

	return &respCode, nil
}
