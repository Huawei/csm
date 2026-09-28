/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

package clientset

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/client-go/rest"

	cmicore "github.com/huawei/csm/v2/provider/cmicore"
)

func TestInitExporterClientSet(t *testing.T) {
	// create mock core
	mockCore := &cmicore.Core{}

	// mock - mock the cmicore NewCore to avoid actual initialization
	patches := gomonkey.
		ApplyFunc(initKubeClientAndSbcClient, func() { return }).
		ApplyFunc(cmicore.NewCore, func(config *cmicore.CoreConfig) *cmicore.Core {
			return mockCore
		}).
		ApplyMethodFunc(mockCore, "Start", func(ctx context.Context) error {
			return nil
		})
	defer patches.Reset()

	// action
	got := InitExporterClientSet()

	// assert - just verify it doesn't panic and returns non-nil
	if got == nil {
		t.Error("InitExporterClientSet() got = nil, want non-nil")
	}
	if got.Core != mockCore {
		t.Errorf("InitExporterClientSet() got.Core = %v, want %v", got.Core, mockCore)
	}
}

func TestDeleteExporterClientSet(t *testing.T) {
	// arrange
	called := false

	// create mock core
	mockCore := &cmicore.Core{}

	// mock
	patches := gomonkey.
		ApplyGlobalVar(&exporterClientSet, &ClientsSet{
			Core: mockCore,
		}).
		ApplyMethodFunc(mockCore, "Stop", func() {
			called = true
		})
	defer patches.Reset()

	// action
	DeleteExporterClientSet()

	// assert
	if called != true {
		t.Errorf("DeleteExporterClientSet() called = %v, want true", called)
	}
}

func TestGetExporterClientSet_Nil(t *testing.T) {
	orig := exporterClientSet
	exporterClientSet = nil
	defer func() { exporterClientSet = orig }()

	if got := GetExporterClientSet(); got != nil {
		t.Errorf("GetExporterClientSet() = %v, want nil", got)
	}
}

func TestDeleteExporterClientSet_NilClientSet(t *testing.T) {
	orig := exporterClientSet
	exporterClientSet = nil
	defer func() { exporterClientSet = orig }()

	DeleteExporterClientSet()
}

func TestDeleteExporterClientSet_NilCore(t *testing.T) {
	orig := exporterClientSet
	exporterClientSet = &ClientsSet{}
	defer func() { exporterClientSet = orig }()

	DeleteExporterClientSet()
}

func TestDeleteExporterClientSet_NilGRPCClientSet(t *testing.T) {
	orig := exporterClientSet
	exporterClientSet = &ClientsSet{
		Core: nil,
	}
	defer func() { exporterClientSet = orig }()

	DeleteExporterClientSet()
}

func TestInitKubeClientAndSbcClient_NilExporterClientSet(t *testing.T) {
	orig := exporterClientSet
	exporterClientSet = nil
	defer func() { exporterClientSet = orig }()

	initKubeClientAndSbcClient()
}

func TestInitKubeClientAndSbcClient_InClusterConfigError(t *testing.T) {
	orig := exporterClientSet
	exporterClientSet = &ClientsSet{}
	defer func() { exporterClientSet = orig }()

	mock := gomonkey.NewPatches()
	defer mock.Reset()

	mock.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
		return nil, errors.New("in cluster config error")
	})

	initKubeClientAndSbcClient()

	if exporterClientSet.InitError == nil {
		t.Error("expected InitError to be set")
	}
}

func TestInitExporterClientSet_AlreadyInitialized(t *testing.T) {
	origOnce := once
	once = sync.Once{}
	defer func() { once = origOnce }()

	origExporterClientSet := exporterClientSet
	exporterClientSet = &ClientsSet{}
	defer func() { exporterClientSet = origExporterClientSet }()

	cs := InitExporterClientSet()
	if cs != exporterClientSet {
		t.Error("should return existing clientSet")
	}
}

func TestInitExporterClientSet_ConcurrentInit(t *testing.T) {
	// This test verifies that concurrent initialization doesn't cause race conditions
	origOnce := once
	once = sync.Once{}
	defer func() { once = origOnce }()

	origExporterClientSet := exporterClientSet
	defer func() { exporterClientSet = origExporterClientSet }()

	// Reset the global for this test
	exporterClientSet = nil

	var wg sync.WaitGroup
	results := make([]*ClientsSet, 10)

	// act - concurrent initialization
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = InitExporterClientSet()
		}(i)
	}
	wg.Wait()

	// assert - all should get the same client set
	for i := 1; i < 10; i++ {
		if results[i] != results[0] {
			t.Errorf("concurrent init returned different clientSet at index %d", i)
		}
	}
}
