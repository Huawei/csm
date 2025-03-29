/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2025. All rights reserved.

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

// Package centralizedstorage is related with centralized storage api
package centralizedstorage

import (
	"github.com/huawei/csm/v2/storage/api"
)

var (
	storageApiMap = map[string]string{
		// system
		"GetSystemInfo": "/system/",

		// filesystem
		"CreateFileSystem":    "/filesystem",
		"GetFileSystemByName": "/filesystem?filter=NAME::{{.fsName}}&range=[0-100]",
		"GetFileSystemById":   "/filesystem/{{.id}}",
		"GetFilesystem":       "/filesystem?range=[{{.start}}-{{.end}}]",
		"GetFilesystemCount":  "/filesystem/count",

		// performance
		"PerformanceData":     "/performance_data?object_type={{.objectType}}&indicators={{.indicators}}",
		"PerformanceDataPost": "/performance_data",

		// storage info
		"GetStoragePools": "/storagepool",
		"GetControllers":  "/controller",

		// lun
		"GetLuns":      "/lun?filter=SUBTYPE::0&range=[{{.start}}-{{.end}}]",
		"GetLunCount":  "/lun/count",
		"GetLunByName": "/lun?filter=NAME::{{.lunName}}&range=[0-100]",

		// label
		"CreatePvLabel":  "/container_pv",
		"DeletePvLabel":  "/container_pv",
		"CreatePodLabel": "/container_pod",
		"DeletePodLabel": "/container_pod",
	}

	storageApis = make(map[string]*api.StorageApi)
)

func init() {
	api.RegisterStorageApi(storageApiMap, storageApis)
}

// GenerateUrl is used to generate centralized storage request url
func GenerateUrl(name string, args map[string]interface{}) (string, error) {
	return api.GenerateUrl(storageApis, name, args)
}
