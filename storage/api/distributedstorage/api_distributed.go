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

// Package distributedstorage is related with distributed storage (FusionStorage) API
package distributedstorage

import (
	"github.com/huawei/csm/v2/storage/api"
)

var (
	storageApiMap = map[string]string{
		// session
		"Login":      "/api/v2/aa/sessions",
		"Logout":     "/api/v2/aa/sessions",
		"SystemTime": "/deviceManager/rest/{{.esn}}/system_utc_time",

		// namespace
		"GetNamespaceCount": "/api/v2/converged_service/namespaces_count",
		"GetNamespace":      "/api/v2/converged_service/namespaces?range={{.range}}",

		// performance
		"PerformanceData": "/api/v2/pms/performance_data",
	}

	storageApis = make(map[string]*api.StorageApi)
)

func init() {
	api.RegisterStorageApi(storageApiMap, storageApis)
}

// GenerateUrl is used to generate distributed storage request url
func GenerateUrl(name string, args map[string]interface{}) (string, error) {
	return api.GenerateUrl(storageApis, name, args)
}
