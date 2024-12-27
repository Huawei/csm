/*
 Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.

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
	"context"
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"
	"time"

	"github.com/pkg/errors"
	admissionV1 "k8s.io/api/admission/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/huawei/csm/v2/utils/log"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetTrueAdmissionResponse is used to get trueAdmissionResponse
func GetTrueAdmissionResponse() *admissionV1.AdmissionResponse {
	return &admissionV1.AdmissionResponse{
		Allowed: true,
	}
}

// GetFalseAdmissionResponse is used to get falseAdmissionResponse with err
func GetFalseAdmissionResponse(err error) *admissionV1.AdmissionResponse {
	return &admissionV1.AdmissionResponse{
		Allowed: false,
		Result: &metaV1.Status{
			Message: err.Error(),
		},
	}
}

// WaitExitSignal is used to wait exits signal, components e.g. webhook, controller
func WaitExitSignal(ctx context.Context, components string) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGILL, syscall.SIGKILL, syscall.SIGTERM)
	stopSignal := <-signalChan
	log.AddContext(ctx).Warningf("stop %s, stopSignal is [%v]", components, stopSignal)
	close(signalChan)
}

// WaitSignal stop the main when stop signals are received
func WaitSignal(ctx context.Context, signalChan chan os.Signal) {
	if signalChan == nil {
		log.AddContext(ctx).Errorln("the channel should not be nil")
		return
	}

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGILL, syscall.SIGKILL, syscall.SIGTERM)
	stopSignal := <-signalChan
	log.AddContext(ctx).Warningf("stop main, stopSignal is [%v]", stopSignal)
}

// RetryFunc retry function, depend on following params and return error
func RetryFunc(function func() (bool, error), retryTimes int,
	retryDurationInit, retryDurationMax time.Duration) error {
	duration := retryDurationInit
	name := runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
	stop := false
	var err error
	for i := 0; i < retryTimes; i++ {
		stop, err = function()
		if stop {
			return err
		}
		if err != nil {
			log.Errorf("retry function [%s] for [%d] times failed: [%v]", name, i+1, err)
		}

		time.Sleep(duration)
		jitter := time.Duration(rand.Int63n(int64(duration)))
		duration = (duration + jitter) * 2
		if duration > retryDurationMax {
			duration = duration / 4
		}
	}

	return errors.Wrap(err, "exceeded retry limit")
}

// Contains returns true if slice s contains item target
func Contains[T comparable](s []T, target T) bool {
	for _, item := range s {
		if item == target {
			return true
		}
	}
	return false
}

// DeleteElementFromSlice is used to delete element from slice
func DeleteElementFromSlice[T comparable](s []T, target T) []T {
	ret := make([]T, 0, len(s))
	for _, item := range s {
		if item != target {
			ret = append(ret, item)
		}
	}
	return ret
}

// GetNameSpaceFromEnv get the namespace from the env
func GetNameSpaceFromEnv(namespaceEnv, defaultNamespace string) string {
	ns := os.Getenv(namespaceEnv)
	if ns == "" {
		ns = defaultNamespace
	}

	return ns
}

// HasDifference check if there is difference between two slices
func HasDifference[T comparable](s1 []T, s2 []T) bool {
	if len(s1) != len(s2) {
		return true
	}

	hMap := make(map[T]struct{})
	for _, item := range s2 {
		hMap[item] = struct{}{}
	}

	for _, item := range s1 {
		if _, ok := hMap[item]; !ok {
			return true
		}
	}

	return false
}

func EncryptMD5(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
