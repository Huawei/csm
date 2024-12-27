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

// Package httpcode is related with http call response code
package httpcode

const (
	// SuccessCode means call api success
	SuccessCode float64 = 0
	// SystemBusy1 means call system busy
	SystemBusy1 float64 = 1077949006
	// SystemBusy2 means call system busy
	SystemBusy2 float64 = 1077948995
	// NoAuthentication means no authentication
	NoAuthentication float64 = -401
)

// RetryCodes means these code need to retry
var RetryCodes = []float64{SystemBusy1, SystemBusy2}
