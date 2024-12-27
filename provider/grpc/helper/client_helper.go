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

// Package helper is a package that helper function
package helper

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	sbcXuanwuClient "github.com/Huawei/eSDK_K8S_Plugin/v4/pkg/client/clientset/versioned"
	"github.com/huawei/csm/v2/config/client"
	"github.com/huawei/csm/v2/utils/log"
)

var clientSet = &ClientSet{}

// ClientSet client set
// contains kubeClient and SbcClient
type ClientSet struct {
	KubeClient *kubernetes.Clientset
	SbcClient  *sbcXuanwuClient.Clientset
}

// InitClientSet init client set
func InitClientSet() error {
	var kubeConfig *rest.Config
	var err error
	if client.GetKubeConfig() != "" {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", client.GetKubeConfig())
	} else {
		kubeConfig, err = rest.InClusterConfig()
	}

	if err != nil {
		log.Errorf("getting kubeConfig [%s] err: [%v]", client.GetKubeConfig(), err)
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Errorf("init kube client failed, err: [%v]", err)
		return err
	}

	sbcClient, err := sbcXuanwuClient.NewForConfig(kubeConfig)
	if err != nil {
		log.Errorf("init sbc client failed, err: [%v]", err)
		return err
	}

	clientSet = &ClientSet{KubeClient: kubeClient, SbcClient: sbcClient}
	return nil
}

// GetClientSet get client set
func GetClientSet() *ClientSet {
	return clientSet
}
