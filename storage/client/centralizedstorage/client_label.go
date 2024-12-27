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

	"github.com/huawei/csm/v2/storage/api/centralizedstorage"
	"github.com/huawei/csm/v2/storage/httpcode"
	"github.com/huawei/csm/v2/storage/httpcode/label"
	"github.com/huawei/csm/v2/utils/log"
)

// PvLabelRequest create pv label request
type PvLabelRequest struct {
	ResourceId   string
	ResourceType string
	PvName       string
	ClusterName  string
}

// PodLabelRequest create and delete label request
type PodLabelRequest struct {
	ResourceId   string
	ResourceType string
	PodName      string
	NameSpace    string
}

// CreatePvLabel create pv label
func (c *CentralizedClient) CreatePvLabel(ctx context.Context, request PvLabelRequest) (map[string]interface{},
	error) {
	data := map[string]interface{}{
		"resourceId":   request.ResourceId,
		"resourceType": request.ResourceType,
		"pvName":       request.PvName,
		"clusterName":  request.ClusterName,
	}

	return c.CreateLabel(ctx, "CreatePvLabel", data, label.PvLabelExist)
}

// DeletePvLabel delete pv label
func (c *CentralizedClient) DeletePvLabel(ctx context.Context, requestId, resourceType string) (map[string]interface{},
	error) {
	data := map[string]interface{}{
		"resourceId":   requestId,
		"resourceType": resourceType,
	}

	return c.DeleteLabel(ctx, "DeletePvLabel", data, label.PvLabelNotExist)
}

// CreatePodLabel create pod label
func (c *CentralizedClient) CreatePodLabel(ctx context.Context, request PodLabelRequest) (map[string]interface{},
	error) {
	data := map[string]interface{}{
		"resourceId":   request.ResourceId,
		"resourceType": request.ResourceType,
		"podName":      request.PodName,
		"nameSpace":    request.NameSpace,
	}

	return c.CreateLabel(ctx, "CreatePodLabel", data, label.PodLabelExist)
}

// DeletePodLabel delete pod label
func (c *CentralizedClient) DeletePodLabel(ctx context.Context, request PodLabelRequest) (map[string]interface{},
	error) {
	data := map[string]interface{}{
		"resourceId":   request.ResourceId,
		"resourceType": request.ResourceType,
		"podName":      request.PodName,
		"nameSpace":    request.NameSpace,
	}

	return c.DeleteLabel(ctx, "DeletePodLabel", data, label.PodLabelNotExist)
}

// CreateLabel create label
func (c *CentralizedClient) CreateLabel(ctx context.Context, urlKey string, data map[string]interface{},
	permittedCode float64) (map[string]interface{}, error) {

	url, err := centralizedstorage.GenerateUrl(urlKey, data)
	if err != nil {
		log.AddContext(ctx).Errorf("create label get url failed, url: %s, error: %v", urlKey, err)
		return nil, err
	}

	callFunc := func() (map[string]interface{}, *float64, error) {
		resp, err := c.post(ctx, url, data)
		if err != nil {
			log.AddContext(ctx).Errorf("create label failed, url: %s ,error: %v", urlKey, err)
			return nil, nil, err
		}

		return getResponse(ctx, resp, url, permittedCode)
	}

	return c.Client.RetryCall(ctx, httpcode.RetryCodes, callFunc)
}

// DeleteLabel create label
func (c *CentralizedClient) DeleteLabel(ctx context.Context, urlKey string, data map[string]interface{},
	permittedCode float64) (map[string]interface{}, error) {

	url, err := centralizedstorage.GenerateUrl(urlKey, data)
	if err != nil {
		log.AddContext(ctx).Errorf("delete label get url failed, url: %s, error: %v", urlKey, err)
		return nil, err
	}

	callFunc := func() (map[string]interface{}, *float64, error) {
		resp, err := c.delete(ctx, url, data)
		if err != nil {
			log.AddContext(ctx).Errorf("delete label failed, url: %s error: %v", urlKey, err)
			return nil, nil, err
		}

		return getResponse(ctx, resp, url, permittedCode)
	}

	return c.Client.RetryCall(ctx, httpcode.RetryCodes, callFunc)
}

func getResponse(ctx context.Context, resp *Response, url string, permittedCode float64) (map[string]interface{},
	*float64, error) {

	respCode, err := getResponseCode(resp)
	if err != nil {
		log.AddContext(ctx).Errorf("get response code failed, url: %s, error: %v", url, err)
		return nil, respCode, err
	}

	if respCode != nil && *respCode == permittedCode {
		log.AddContext(ctx).Infoln("The specified resource object has been associated with the current label")
		*respCode = httpcode.SuccessCode
	}

	if respCode != nil && *respCode != httpcode.SuccessCode {
		msg := fmt.Sprintf("storage client response httpcode is not success code, "+
			"code: %v, description: %v", respCode, resp.Error["description"])
		log.AddContext(ctx).Errorf(msg)
		return nil, respCode, errors.New(msg)
	}

	responseData, err := getResponseData(resp)
	if err != nil {
		log.AddContext(ctx).Errorf("get response data failed, url: %s, error: %v", url, err)
		return nil, respCode, err
	}

	return responseData, respCode, nil
}

func getResponseCode(response *Response) (*float64, error) {
	respCode, exist := response.Error["code"].(float64)
	if !exist {
		msg := fmt.Sprintf("storage client response httpcode does not exist, response: %v", response)
		return nil, errors.New(msg)
	}
	return &respCode, nil
}

func getResponseData(response *Response) (map[string]interface{}, error) {
	if response.Data == nil {
		return map[string]interface{}{}, errors.New("response data is nil")
	}

	respData, exist := response.Data.(map[string]interface{})
	if !exist {
		msg := fmt.Sprintf("storage client response data can not convert to map[string]interface{},"+
			" response data: %v", response.Data)
		return map[string]interface{}{}, errors.New(msg)
	}
	return respData, nil
}
