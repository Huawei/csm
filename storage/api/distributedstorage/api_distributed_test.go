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

import "testing"

func TestGenerateUrl_GetNamespaceCount(t *testing.T) {
	// arrange & act
	url, err := GenerateUrl("GetNamespaceCount", nil)
	// assert
	if err != nil {
		t.Errorf("GenerateUrl returned error: %v", err)
	}
	if url != "/api/v2/converged_service/namespaces_count" {
		t.Errorf("GetNamespaceCount url = %q, want %q", url, "/api/v2/converged_service/namespaces_count")
	}
}

func TestGenerateUrl_GetNamespace(t *testing.T) {
	// arrange
	args := map[string]interface{}{"range": `{"offset":0,"limit":100}`}
	// act
	url, err := GenerateUrl("GetNamespace", args)
	// assert
	if err != nil {
		t.Errorf("GenerateUrl returned error: %v", err)
	}
	expected := `/api/v2/converged_service/namespaces?range={"offset":0,"limit":100}`
	if url != expected {
		t.Errorf("GetNamespace url = %q, want %q", url, expected)
	}
}

func TestGenerateUrl_SystemTime(t *testing.T) {
	// arrange
	args := map[string]interface{}{"esn": "ABC123"}
	// act
	url, err := GenerateUrl("SystemTime", args)
	// assert
	if err != nil {
		t.Errorf("GenerateUrl returned error: %v", err)
	}
	if url != "/deviceManager/rest/ABC123/system_utc_time" {
		t.Errorf("SystemTime url = %q, want %q", url, "/deviceManager/rest/ABC123/system_utc_time")
	}
}

func TestGenerateUrl_PerformanceData(t *testing.T) {
	// arrange & act
	url, err := GenerateUrl("PerformanceData", nil)
	// assert
	if err != nil {
		t.Errorf("GenerateUrl returned error: %v", err)
	}
	if url != "/api/v2/pms/performance_data" {
		t.Errorf("PerformanceData url = %q, want %q", url, "/api/v2/pms/performance_data")
	}
}

func TestGenerateUrl_Login(t *testing.T) {
	// arrange & act
	url, err := GenerateUrl("Login", nil)
	// assert
	if err != nil {
		t.Errorf("GenerateUrl returned error: %v", err)
	}
	if url != "/api/v2/aa/sessions" {
		t.Errorf("Login url = %q, want %q", url, "/api/v2/aa/sessions")
	}
}

func TestGenerateUrl_InvalidKey(t *testing.T) {
	// arrange & act
	_, err := GenerateUrl("InvalidKey", nil)
	// assert
	if err == nil {
		t.Error("GenerateUrl should return error for invalid key")
	}
}
