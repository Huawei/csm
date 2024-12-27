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

// Package filesystem is used to list filesystem related api response code
package filesystem

import "github.com/huawei/csm/v2/storage/httpcode"

const (
	operatorFail float64 = -1

	// FileSystemNotExist means file system not exist
	FileSystemNotExist float64 = 1073752065

	// FileSystemExist means file system exist
	FileSystemExist float64 = 1077948993

	// CloneFileSystemNotEmpty means clone file system not empty
	CloneFileSystemNotEmpty float64 = 1073844244
)

var retryCodes = []float64{operatorFail, httpcode.SystemBusy1, httpcode.SystemBusy2}

// GetRetryCodes is used to get api response code which can retry call api
func GetRetryCodes() []float64 {
	return retryCodes
}
