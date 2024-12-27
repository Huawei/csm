/*
 Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

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

// Package client
package client

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
)

func Test_option_ValidateConfig_Success(t *testing.T) {
	// arrange
	o := &option{
		kubeConfig: "fakeConfig",
	}

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(filepath.IsAbs, func(path string) bool {
		return true
	})

	// act
	err := o.ValidateConfig()

	// assert
	if err != nil {
		t.Errorf("Test_option_ValidateConfig_Success failed: [%v]", err)
	}

	// clean
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_option_ValidateConfig_Fail(t *testing.T) {
	// arrange
	o := &option{
		kubeConfig: "fakeConfig",
	}
	want := errors.New("kubeConfig file path is not absolute")

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(filepath.IsAbs, func(path string) bool {
		return false
	})

	// act
	got := o.ValidateConfig()

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Test_option_ValidateConfig_Fail: want [%v], got [%v]", want, got)
	}

	// clean
	t.Cleanup(func() {
		mock.Reset()
	})
}
