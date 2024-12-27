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

// Package utils is related with storage client utils
package utils

import (
	"flag"
	"time"
)

const (
	maxRetryNumber = 5
	sleepTime      = 2 * time.Second
)

var storageClientMaxRetryTimes = flag.Int("storage-client-max-retry-times", maxRetryNumber, "maximum number of retries")
var storageClientRetryInterval = flag.Duration("storage-client-retry-interval", sleepTime, "retry interval")

// RetryCallFunc is used to retry call func
// the func return true will retry call func, return false will end call
func RetryCallFunc(retryFunc func() bool) {
	retryNumber := 0
	for {
		if shouldRetry := retryFunc(); !shouldRetry {
			break
		}

		if retryNumber++; retryNumber > *storageClientMaxRetryTimes {
			break
		}

		time.Sleep(*storageClientRetryInterval)
	}
}
