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

// Package client is related with storage common client and operation
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"

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

// TestCloneFileSystemThenSuccess test Call() success
func TestCallThenSuccess(t *testing.T) {
	resp := map[string]interface{}{
		"code": "0",
	}
	js, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("Call() error: %v", err)
	}
	httpCli := &http.Client{}
	httpCLiDo := gomonkey.ApplyMethod(reflect.TypeOf(httpCli), "Do",
		func(_ *http.Client, req *http.Request) (*http.Response, error) {
			return &http.Response{
				Body: ioutil.NopCloser(bytes.NewReader(js)),
			}, nil
		})
	defer httpCLiDo.Reset()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Errorf("Call() error: %v", err)
	}
	cli := &Client{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar:     jar,
			Timeout: 60 * time.Second,
		},
	}
	_, err = cli.Call(ctx, "method", "url", map[string]interface{}{})
	if err != nil {
		t.Errorf("Call() error: %v", err)
	}
}

func TestClient_getRequest_JsonMarshalFailed(t *testing.T) {
	// arrange
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Errorf("Call() error: %v", err)
	}

	cli := &Client{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar:     jar,
			Timeout: 60 * time.Second,
		},
	}
	reqData := map[string]interface{}{"username": "test_user", "password": "123456"}
	wantErr := errors.New("mock err")

	// mock
	p := gomonkey.NewPatches()
	p.ApplyFunc(json.Marshal, func(v any) ([]byte, error) {
		return nil, wantErr
	})

	// action
	_, gotErr := cli.getRequest(ctx, "method", sessionsSubStr, reqData)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("TestClient_getRequest_JsonMarshalFailed failed, want err = %v, get err = %v", wantErr, gotErr)
	}

	// cleanup
	t.Cleanup(func() {
		p.Reset()
	})
}
