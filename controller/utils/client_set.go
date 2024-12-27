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
	"fmt"

	apiV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	coreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"

	cmiGrpc "github.com/huawei/csm/v2/grpc/lib/go/cmi"
	xuanwuClient "github.com/huawei/csm/v2/pkg/client/clientset/versioned"
	"github.com/huawei/csm/v2/utils/log"
)

// ClientsSet contains all clients needed by controller
type ClientsSet struct {
	Config           *rest.Config
	CmiClient        *cmiGrpc.ClientSet
	KubeClient       kubernetes.Interface
	XuanwuClient     xuanwuClient.Interface
	DynamicClient    dynamic.Interface
	EventBroadcaster record.EventBroadcaster
	EventRecorder    record.EventRecorder
	CmiAddress       string
}

const (
	eventComponentName = "huawei-csm"
)

var (
	initFuncList = []func(*ClientsSet) error{
		initKubeClient,
		initXuanwuClient,
		initDynamicClient,
		initEventBroadcaster,
		initEventRecorder,
		initCmiClient,
	}
)

// NewClientsSet creates a new clients set with the given kube config
func NewClientsSet(config string, cmiAddress string) (*ClientsSet, error) {
	var kubeConfig *rest.Config
	var err error
	if config != "" {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", config)
	} else {
		kubeConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Errorf("getting kubeConfig [%s] err: [%v]", config, err)
		return nil, err
	}

	clientsSet := &ClientsSet{}
	clientsSet.Config = kubeConfig
	clientsSet.CmiAddress = cmiAddress

	for _, initFunction := range initFuncList {
		err := initFunction(clientsSet)
		if err != nil {
			return nil, err
		}
	}

	return clientsSet, nil
}

func initKubeClient(c *ClientsSet) error {
	log.Infoln("initial kubernetes client")
	defer log.Infoln("initial kubernetes client success")
	if c.KubeClient != nil {
		return nil
	}

	kubeClient, err := kubernetes.NewForConfig(c.Config)
	if err != nil {
		log.Errorf("init kubernetes client error: [%v]", err)
		return err
	}

	c.KubeClient = kubeClient
	return nil
}

func initXuanwuClient(c *ClientsSet) error {
	log.Infoln("initial xuanwu client")
	defer log.Infoln("initial xuanwu client success")
	if c.XuanwuClient != nil {
		return nil
	}

	client, err := xuanwuClient.NewForConfig(c.Config)
	if err != nil {
		log.Errorf("init xuanwu client error: [%v]", err)
		return err
	}

	c.XuanwuClient = client
	return nil
}

func initDynamicClient(c *ClientsSet) error {
	log.Infoln("initial dynamic client")
	if c.DynamicClient != nil {
		return nil
	}

	client, err := dynamic.NewForConfig(c.Config)
	if err != nil {
		log.Errorf("init dynamic client error: [%v]", err)
		return err
	}

	c.DynamicClient = client
	log.Infoln("initial dynamic client success")
	return nil
}

func initEventBroadcaster(c *ClientsSet) error {
	log.Infoln("initial event broadcaster")
	if c.EventBroadcaster != nil {
		return nil
	}

	if c.KubeClient == nil {
		client, err := kubernetes.NewForConfig(c.Config)
		if err != nil {
			log.Errorf("init kubernetes client error: [%v]", err)
			return err
		}
		c.KubeClient = client
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&coreV1.EventSinkImpl{Interface: c.KubeClient.CoreV1().Events("")})

	c.EventBroadcaster = eventBroadcaster
	log.Infoln("initial event broadcaster success")
	return nil
}

func initEventRecorder(c *ClientsSet) error {
	log.Infoln("initial event recorder")
	if c.EventRecorder != nil {
		return nil
	}

	if c.KubeClient == nil {
		client, err := kubernetes.NewForConfig(c.Config)
		if err != nil {
			log.Errorf("init xuanwu client error: [%v]", err)
			return err
		}
		c.KubeClient = client
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(
		&coreV1.EventSinkImpl{Interface: c.KubeClient.CoreV1().Events(apiV1.NamespaceAll)})
	c.EventRecorder = eventBroadcaster.NewRecorder(
		scheme.Scheme, apiV1.EventSource{Component: fmt.Sprintf(eventComponentName)})
	log.Infoln("initial event recorder success")
	return nil
}

func initCmiClient(c *ClientsSet) error {
	log.Infoln("initial cmi client")
	if c.CmiClient != nil {
		return nil
	}

	cmiClientSet, err := cmiGrpc.GetClientSet(c.CmiAddress)
	if err != nil {
		return fmt.Errorf("error getting client set of cmi: [%v]", err)
	}
	cmiClientSet.Conn.Connect()
	c.CmiClient = cmiClientSet

	log.Infoln("initial cmi client success")
	return nil
}
