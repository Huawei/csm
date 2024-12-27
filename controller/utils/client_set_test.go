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

// Package utils
package utils

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	fakeDynamicClient "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	xuanwuClient "github.com/huawei/csm/v2/pkg/client/clientset/versioned"
	fakeXuanwuClient "github.com/huawei/csm/v2/pkg/client/clientset/versioned/fake"
)

func TestNewClientsSet_EmptyConfigEmptyInitFuncList_Success(t *testing.T) {
	// arrange
	config := ""
	wantClientsSet := &ClientsSet{Config: &rest.Config{}}

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}).ApplyGlobalVar(&initFuncList, []func(*ClientsSet) error{})

	// act
	clients, err := NewClientsSet(config, "")

	// assert
	if err != nil {
		t.Errorf("TestNewClientsSet_EmptyConfig_Success failed: [%v]", err)
	}
	if !reflect.DeepEqual(clients, wantClientsSet) {
		t.Errorf("TestNewClientsSet_EmptyConfig_Success failed, want: [%v], get [%v]", wantClientsSet, clients)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestNewClientsSet_EmptyConfig_Success(t *testing.T) {
	// arrange
	config := ""
	simpleCsiXuanwuClient := fakeXuanwuClient.NewSimpleClientset()
	simpleKubeClient := fake.NewSimpleClientset()
	simpleDynamicClient := fakeDynamicClient.NewSimpleDynamicClient(scheme.Scheme)
	wantClientsSet := &ClientsSet{
		Config:        &rest.Config{},
		XuanwuClient:  simpleCsiXuanwuClient,
		KubeClient:    simpleKubeClient,
		DynamicClient: simpleDynamicClient,
	}

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}).ApplyGlobalVar(&initFuncList, []func(*ClientsSet) error{
		initXuanwuClient,
		initKubeClient,
		initDynamicClient,
	}).ApplyFunc(initXuanwuClient, func(c *ClientsSet) error {
		c.XuanwuClient = simpleCsiXuanwuClient
		return nil
	}).ApplyFunc(initKubeClient, func(c *ClientsSet) error {
		c.KubeClient = simpleKubeClient
		return nil
	}).ApplyFunc(initDynamicClient, func(c *ClientsSet) error {
		c.DynamicClient = simpleDynamicClient
		return nil
	})

	// act
	clients, err := NewClientsSet(config, "")

	// assert
	if err != nil {
		t.Errorf("TestNewClientsSet_EmptyConfig_Success failed: %v", err)
	}
	if !reflect.DeepEqual(clients, wantClientsSet) {
		t.Errorf("TestNewClientsSet_EmptyConfig_Success failed, want: %v,get %v", wantClientsSet, clients)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestNewClientsSet_EmptyConfigEmptyInitFuncList_Fail(t *testing.T) {
	// arrange
	config := ""
	wantErr := errors.New("getting kubeConfig [] err: [fake error]")

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
		return nil, errors.New("fake error")
	}).ApplyGlobalVar(&initFuncList, []func(*ClientsSet) error{})

	// act
	clients, err := NewClientsSet(config, "/cmi/cmi.sock")

	// assert
	if reflect.DeepEqual(err, wantErr) {
		t.Errorf("TestNewClientsSet_EmptyConfig_Fail failed, want: %v, got: %v", wantErr, err)
	}
	if clients != nil {
		t.Errorf("TestNewClientsSet_EmptyConfig_Fail failed want nil clients set, got: %v", clients)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestNewClientsSet_EmptyConfigInitFuncErr_Fail(t *testing.T) {
	// arrange
	config := ""
	wantErr := errors.New("init kubernetes client error: [fake error]")

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(rest.InClusterConfig, func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}).ApplyGlobalVar(&initFuncList, []func(*ClientsSet) error{
		initKubeClient,
	}).ApplyFunc(initKubeClient, func(c *ClientsSet) error {
		return errors.New("fake error")
	})

	// act
	clients, err := NewClientsSet(config, "/cmi/cmi.sock")

	// assert
	if reflect.DeepEqual(err, wantErr) {
		t.Errorf("TestNewClientsSet_EmptyConfig_Fail failed, want: %v, got: %v", wantErr, err)
	}
	if clients != nil {
		t.Errorf("TestNewClientsSet_EmptyConfig_Fail failed want nil clients set, got: %v", clients)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_initKubeClient_NilClient_Success(t *testing.T) {
	// arrange
	c := &ClientsSet{Config: &rest.Config{}}
	want := &ClientsSet{Config: &rest.Config{}, KubeClient: &kubernetes.Clientset{}}

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
		return &kubernetes.Clientset{}, nil
	})

	// act
	err := initKubeClient(c)

	// assert
	if err != nil {
		t.Errorf("Test_initKubeClient_NilClient_Success err: [%v]", err)
	}
	if !reflect.DeepEqual(c, want) {
		t.Errorf("Test_initKubeClient_NilClient_Success failed: want: [%v], got: [%v]", want, c)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_initKubeClient_NilClient_Fail(t *testing.T) {
	// arrange
	c := &ClientsSet{Config: &rest.Config{}}
	wantErr := errors.New("fake error")

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
		return nil, errors.New("fake error")
	})

	// act
	gotErr := initKubeClient(c)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("Test_initKubeClient_NilClient_Fail failed: wantErr: [%v], gotErr: [%v]", gotErr, wantErr)
	}
	if c.KubeClient != nil {
		t.Error("Test_initKubeClient_NilClient_Fail failed, kube client should be nil")
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_initKubeClient_WithClient_Success(t *testing.T) {
	// arrange
	kubeClient := &kubernetes.Clientset{}
	c := &ClientsSet{Config: &rest.Config{}, KubeClient: kubeClient}

	// act
	err := initKubeClient(c)

	// assert
	if err != nil {
		t.Errorf("Test_initKubeClient_WithClient_Success failed, err: [%v]", err)
	}
	if c.KubeClient != kubeClient {
		t.Error("Test_initKubeClient_WithClient_Success failed, kube client changed")
	}
}

func Test_initCsiClient_NilClient_Success(t *testing.T) {
	// arrange
	c := &ClientsSet{Config: &rest.Config{}}
	want := &ClientsSet{Config: &rest.Config{}, XuanwuClient: &xuanwuClient.Clientset{}}

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(xuanwuClient.NewForConfig, func(c *rest.Config) (*xuanwuClient.Clientset, error) {
		return &xuanwuClient.Clientset{}, nil
	})

	// act
	err := initXuanwuClient(c)

	// assert
	if err != nil {
		t.Errorf("Test_initCsiClient_NilClient_Success err: [%v]", err)
	}
	if !reflect.DeepEqual(c, want) {
		t.Errorf("Test_initCsiClient_NilClient_Success failed: want: [%v], got: [%v]", want, c)
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_initCsiClient_NilClient_Fail(t *testing.T) {
	// arrange
	c := &ClientsSet{Config: &rest.Config{}}
	wantErr := errors.New("fake error")

	// mock
	mock := gomonkey.NewPatches()

	// expect
	mock.ApplyFunc(xuanwuClient.NewForConfig, func(c *rest.Config) (*xuanwuClient.Clientset, error) {
		return nil, errors.New("fake error")
	})

	// act
	gotErr := initXuanwuClient(c)

	// assert
	if !reflect.DeepEqual(gotErr, wantErr) {
		t.Errorf("Test_initCsiClient_NilClient_Fail failed: wantErr: [%v], gotErr: [%v]", gotErr, wantErr)
	}
	if c.KubeClient != nil {
		t.Error("Test_initCsiClient_NilClient_Fail failed, kube client should be nil")
	}

	// cleanup
	t.Cleanup(func() {
		mock.Reset()
	})
}

func Test_initCsiClient_WithClient_Success(t *testing.T) {
	// arrange
	csiClient := &xuanwuClient.Clientset{}
	c := &ClientsSet{Config: &rest.Config{}, XuanwuClient: csiClient}

	// act
	err := initXuanwuClient(c)

	// assert
	if err != nil {
		t.Errorf("Test_initCsiClient_WithClient_Success failed, err: [%v]", err)
	}
	if c.XuanwuClient != csiClient {
		t.Error("Test_initCsiClient_WithClient_Success failed, csi client changed")
	}
}
