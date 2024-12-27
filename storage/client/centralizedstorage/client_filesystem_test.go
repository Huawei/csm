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
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"github.com/huawei/csm/v2/storage/client"
	"github.com/huawei/csm/v2/storage/utils"
	"github.com/huawei/csm/v2/utils/log"
)

const (
	logName string = "storage_client_test"
	logDir  string = "/var/log/xuanwu"
)

// get ctx
var ctx = context.Background()

// TestMain used for setup and teardown
func TestMain(m *testing.M) {
	// init log
	if err := log.InitLogging(logName); err != nil {
		_ = fmt.Errorf("init logging: %s failed. error: %v", logName, err)
		return
	}

	m.Run()

	// remove all log file
	logFile := path.Join(logDir, logName)
	if err := os.RemoveAll(logFile); err != nil {
		log.Errorf("Remove file: %s failed. error: %s", logFile, err)
	}
}

// TestGetFileSystemByNameThenSuccess test GetFileSystemByName() success
func TestGetFileSystemByNameThenSuccess(t *testing.T) {
	response := map[string]interface{}{
		"Error": map[string]interface{}{
			"code": 0,
		},
		"Data": []map[string]interface{}{{
			"test1": 1,
			"test2": 2,
		}},
	}

	var cli *client.Client
	httpGet := gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return response, nil
		})
	defer httpGet.Reset()

	centralizedCli := &CentralizedClient{
		Client: client.Client{
			Semaphore: utils.NewSemaphore(3),
		},
	}
	_, err := centralizedCli.GetFileSystemByName(ctx, "nameTest")
	if err != nil {
		t.Errorf("GetFileSystemByName() error: %v", err)
	}
}

// TestGetFileSystemByNameWhenResponseErrorThenFailed test GetFileSystemByName() failed
func TestGetFileSystemByNameWhenResponseErrorThenFailed(t *testing.T) {
	response := map[string]interface{}{
		"Error": map[string]interface{}{
			"code": -1,
		},
		"Data": []map[string]interface{}{},
	}

	var cli *client.Client
	httpGet := gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return response, nil
		})
	defer httpGet.Reset()

	centralizedCli := &CentralizedClient{
		Client: client.Client{
			Semaphore: utils.NewSemaphore(3),
		},
	}
	_, err := centralizedCli.GetFileSystemByName(ctx, "nameTest")

	expectMsg := fmt.Sprintf("storage client response httpcode is not success code, "+
		"code: %v, description: %v", -1, nil)
	actualMsg := fmt.Sprintf("%v", err)

	if actualMsg != expectMsg {
		t.Errorf("GetFileSystemByName() error: %v", err)
	}
}

// TestGetFileSystemByNameWhenResponseCodeNotExistThenFailed test GetFileSystemByName() failed
func TestGetFileSystemByNameWhenResponseCodeNotExistThenFailed(t *testing.T) {
	response := map[string]interface{}{
		"Error": map[string]interface{}{},
		"Data":  []map[string]interface{}{},
	}

	var cli *client.Client
	httpGet := gomonkey.ApplyMethod(reflect.TypeOf(cli), "Call",
		func(_ *client.Client, ctx context.Context, method string,
			url string, reqData map[string]interface{}) (map[string]interface{}, error) {
			return response, nil
		})
	defer httpGet.Reset()

	centralizedCli := &CentralizedClient{
		Client: client.Client{
			Semaphore: utils.NewSemaphore(3),
		},
	}
	_, err := centralizedCli.GetFileSystemByName(ctx, "nameTest")

	expectMsg := fmt.Sprintf("storage client response httpcode does not exist, response: %v", &Response{
		Error: map[string]interface{}{},
		Data:  []interface{}{},
	})
	actualMsg := fmt.Sprintf("%v", err)

	if actualMsg != expectMsg {
		t.Errorf("GetFileSystemByName() error: %v", err)
	}
}
