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

package exporterhandler

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
)

func Test_parseRequestPath(t *testing.T) {
	// arrange
	ctx := context.TODO()
	mockRequest := http.Request{URL: &url.URL{Path: "/object/backend_name"}}
	var mockResponse http.ResponseWriter

	// mock
	patches := gomonkey.
		ApplyFunc(checkMetricsObject,
			func(ctx context.Context, params map[string][]string, monitorType string) error {
				return nil
			})
	defer patches.Reset()

	// action
	monitorBackendName, monitorType, _ := parseRequestPath(ctx, mockResponse, &mockRequest)

	// assert
	if monitorBackendName != "backend_name" || monitorType != "object" {
		t.Errorf("parseRequestPath() error want backend_name and object")
	}
}

func Test_checkMetricsObject_Success(t *testing.T) {
	// arrange
	ctx := context.TODO()
	mockParams := map[string][]string{"array": {""}}

	// action
	err := checkMetricsObject(ctx, mockParams, "object")

	// assert
	if err != nil {
		t.Errorf("checkMetricsObject() error the err is [%v]", err)
	}
}

func Test_checkMetricsObject_PerformanceIndicatorsError(t *testing.T) {
	// arrange
	ctx := context.TODO()
	mockParams := map[string][]string{"array": {""}}

	// action
	err := checkMetricsObject(ctx, mockParams, "performance")

	// assert
	if err.Error() != "the metricsIndicators is error" {
		t.Errorf("checkMetricsObject() error the err is [%v]", err)
	}
}
