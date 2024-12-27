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

// Package utils is a package that provides utilities for controllers
package utils

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/pkg/errors"
	admissionV1 "k8s.io/api/admission/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetTrueAdmissionResponse(t *testing.T) {
	// arrange
	want := &admissionV1.AdmissionResponse{
		Allowed: true,
	}

	// act
	got := GetTrueAdmissionResponse()

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetTrueAdmissionResponse failed: want [%v], got: [%v]", got, want)
	}
}

func TestGetFalseAdmissionResponse(t *testing.T) {
	// arrange
	err := errors.New("fake error")
	want := &admissionV1.AdmissionResponse{
		Allowed: false,
		Result: &metaV1.Status{
			Message: errors.New("fake error").Error(),
		},
	}

	// act
	got := GetFalseAdmissionResponse(err)

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetFalseAdmissionResponse failed: want [%v], got: [%v]", got, want)
	}
}

func TestRetryFunc_Success(t *testing.T) {
	// arrange
	retryTimes := 2
	retryDurationInit := 1 * time.Second
	retryDurationMax := 10 * time.Second
	fakeFunc := func() (bool, error) {
		return true, nil
	}

	// act
	err := RetryFunc(fakeFunc, retryTimes, retryDurationInit, retryDurationMax)

	// assert
	if err != nil {
		t.Errorf("TestRetryFunc_Success failed: [%v]", err)
	}
}

func TestRetryFunc_Fail(t *testing.T) {
	// arrange
	retryTimes := 2
	retryDurationInit := 1 * time.Second
	retryDurationMax := 10 * time.Second
	fakeFunc := func() (bool, error) {
		return false, errors.New("fake error")
	}
	wantErr := errors.New("exceeded retry limit: fake error")

	// act
	got := RetryFunc(fakeFunc, retryTimes, retryDurationInit, retryDurationMax)

	// assert
	if got.Error() != wantErr.Error() {
		t.Errorf("TestRetryFunc_Fail failed: wantErr: [%v], got: [%v]", wantErr, got)
	}
}

func TestContains_True(t *testing.T) {
	// arrange
	s := []string{"a", "b", "c", "d"}
	target := "c"

	// act
	ok := Contains(s, target)

	// assert
	if !ok {
		t.Errorf("TestContains_True: got: [%t], want: [%t]", ok, true)
	}
}

func TestContains_False(t *testing.T) {
	// arrange
	s := []string{"a", "b", "c", "d"}
	target := "e"

	// act
	ok := Contains(s, target)

	// assert
	if ok {
		t.Errorf("TestContains_True: got: [%t], want: [%t]", ok, false)
	}
}

func TestGetNameSpaceFromEnv_DefaultNs(t *testing.T) {
	// arrange
	want := "fakeNamespace"

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(os.Getenv, func(s string) string {
		return "fakeNamespace"
	})

	// act
	got := GetNameSpaceFromEnv("", "")

	// assert
	if got != want {
		t.Errorf("TestGetNameSpaceFromEnv_DefaultNs: want: [%v], got: [%v]", want, got)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestGetNameSpaceFromEnv_SpecNs(t *testing.T) {
	// arrange
	want := "fakeNamespace"

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(os.Getenv, func(s string) string {
		return ""
	})

	// act
	got := GetNameSpaceFromEnv("", "fakeNamespace")

	// assert
	if got != want {
		t.Errorf("TestGetNameSpaceFromEnv_DefaultNs: want: [%v], got: [%v]", want, got)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestHasDifference_False(t *testing.T) {
	// arrange
	s1 := []string{"a", "b", "c", "d"}
	s2 := []string{"a", "b", "c", "d"}

	// act
	got := HasDifference(s1, s2)

	// assert
	if got {
		t.Errorf("TestHasDifference_False: got: [%t], want: [%t]", got, false)
	}
}

func TestHasDifference_DiffLength(t *testing.T) {
	// arrange
	s1 := []int{1, 2, 3, 4, 5}
	s2 := []int{1, 2, 3}

	// act
	got := HasDifference(s1, s2)

	// assert
	if !got {
		t.Errorf("TestHasDifference_DiffLength: got: [%t], want: [%t]", got, true)
	}
}

func TestHasDifference_DiffEle(t *testing.T) {
	// arrange
	s1 := []int{1, 2, 3, 4, 5}
	s2 := []int{1, 2, 3, 1, 2}

	// act
	got := HasDifference(s1, s2)

	// assert
	if !got {
		t.Errorf("TestHasDifference_DiffEle: got: [%t], want: [%t]", got, true)
	}
}
