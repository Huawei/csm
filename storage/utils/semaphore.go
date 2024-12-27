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

// Semaphore stores semaphore info
type Semaphore struct {
	permits int
	channel chan int
}

// NewSemaphore is used to new semaphore
func NewSemaphore(permits int) *Semaphore {
	return &Semaphore{
		channel: make(chan int, permits),
		permits: permits,
	}
}

// Acquire is used to get semaphore
func (s *Semaphore) Acquire() {
	s.channel <- 0
}

// Release is used to remove semaphore
func (s *Semaphore) Release() {
	<-s.channel
}

// AvailablePermits get available permits
func (s *Semaphore) AvailablePermits() int {
	return s.permits - len(s.channel)
}
