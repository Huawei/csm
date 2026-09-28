/*
 Copyright (c) Huawei Technologies Co., Ltd. 2023-2026. All rights reserved.

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
	"fmt"

	apiV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	coreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"

	sbcClient "github.com/Huawei/eSDK_K8S_Plugin/v4/pkg/client/clientset/versioned"
	"github.com/huawei/csm/v2/config/client"
	cmiConfig "github.com/huawei/csm/v2/config/cmi"
	leaderElectionConfig "github.com/huawei/csm/v2/config/leaderelection"
	controllerConfig "github.com/huawei/csm/v2/config/topology"
	xuanwuClient "github.com/huawei/csm/v2/pkg/client/clientset/versioned"
	"github.com/huawei/csm/v2/provider/cmicore"
	"github.com/huawei/csm/v2/utils/log"
)

// ClientsSet contains all clients needed by controller
type ClientsSet struct {
	Config             *rest.Config
	ElectionKubeClient kubernetes.Interface
	Core               *cmicore.Core
	KubeClient         kubernetes.Interface
	XuanwuClient       xuanwuClient.Interface
	SbcClient          sbcClient.Interface
	DynamicClient      dynamic.Interface
	EventBroadcaster   record.EventBroadcaster
	EventRecorder      record.EventRecorder
}

const (
	eventComponentName = "huawei-csm"
)

var (
	initFuncList = []func(*ClientsSet) error{
		initElectionKubeClient,
		initKubeClient,
		initXuanwuClient,
		initSbcClient,
		initDynamicClient,
		initEventBroadcaster,
		initEventRecorder,
		initCore,
	}
)

// NewClientsSet creates a new clients set with the given kube config
func NewClientsSet(config string) (*ClientsSet, error) {
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

	client.ApplyKubeAPIQPSBurst(kubeConfig)

	clientsSet := &ClientsSet{}
	clientsSet.Config = kubeConfig

	for i, initFunction := range initFuncList {
		err := initFunction(clientsSet)
		if err != nil {
			// If Core was already created (initCore succeeded), cleanup before returning
			if i > 0 && clientsSet.Core != nil {
				clientsSet.Core.Stop()
			}
			return nil, err
		}
	}

	return clientsSet, nil
}

func initElectionKubeClient(c *ClientsSet) error {
	log.Infoln("initial election kubernetes client")
	defer log.Infoln("initial election kubernetes client success")
	if c.ElectionKubeClient != nil {
		log.Warningln("ElectionKubeClient already exists")
		return nil
	}

	electionConfig := rest.CopyConfig(c.Config)
	leaderElectionConfig.ApplyLeaderElectionQPSBurst(electionConfig)

	kubeClient, err := kubernetes.NewForConfig(electionConfig)
	if err != nil {
		log.Errorf("init election kubernetes client error: [%v]", err)
		return err
	}

	c.ElectionKubeClient = kubeClient
	return nil
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

func initCore(c *ClientsSet) error {
	log.Infoln("initial cmicore")
	if c.Core != nil {
		return nil
	}

	coreConfig := &cmicore.CoreConfig{
		KubeClient:           c.KubeClient,
		SbcClient:            c.SbcClient,
		BackendNamespace:     controllerConfig.GetBackendNamespace(),
		QueryStoragePageSize: cmiConfig.GetQueryStoragePageSize(),
		ClientMaxThreads:     cmiConfig.GetClientMaxThreads(),
	}

	core := cmicore.NewCore(coreConfig)
	if err := core.Start(context.Background()); err != nil {
		return fmt.Errorf("start cmicore failed: [%w]", err)
	}

	c.Core = core
	log.Infoln("initial cmicore success")
	return nil
}

// DeleteClientsSet cleans up the ClientsSet resources
func DeleteClientsSet(c *ClientsSet) {
	if c == nil || c.Core == nil {
		return
	}
	c.Core.Stop()
}

func initSbcClient(c *ClientsSet) error {
	log.Infoln("initial sbc client")
	defer log.Infoln("initial sbc client success")
	if c.SbcClient != nil {
		return nil
	}

	client, err := sbcClient.NewForConfig(c.Config)
	if err != nil {
		log.Errorf("init sbc client error: [%v]", err)
		return err
	}

	c.SbcClient = client
	return nil
}
