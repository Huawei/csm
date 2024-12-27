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

// Package clientset provide all client use by prometheus exporter
package clientset

import (
	"sync"

	sbcXuanwuClient "github.com/Huawei/eSDK_K8S_Plugin/v4/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientConfig "github.com/huawei/csm/v2/config/client"
	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/utils/log"
)

var once sync.Once

// ClientsSet contains all clients needed by prometheus exporter
type ClientsSet struct {
	// KubeClient get the Kubernetes client to get Kubernetes resource
	KubeClient *kubernetes.Clientset
	// SbcClient get backend client
	SbcClient *sbcXuanwuClient.Clientset
	// StorageGRPCClientSet From grpc get storage data client
	StorageGRPCClientSet *storageGRPC.ClientSet
	// InitError when init error this will set the reason
	InitError error
}

// exporterClientSet all client needed by prometheus exporter
var exporterClientSet *ClientsSet

// GetExporterClientSet get the exporterClientSet
func GetExporterClientSet() *ClientsSet {
	return exporterClientSet
}

func initKubeClientAndSbcClient() {
	if exporterClientSet == nil {
		return
	}
	var kubeConfig *rest.Config
	var err error

	if clientConfig.GetKubeConfig() != "" {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", clientConfig.GetKubeConfig())
	} else {
		kubeConfig, err = rest.InClusterConfig()
	}

	if err != nil {
		log.Errorf("getting kubeConfig [%s] err: [%v]", clientConfig.GetKubeConfig(), err)
		exporterClientSet.InitError = err
		return
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Errorf("init kube client failed, err: [%v]", err)
		exporterClientSet.InitError = err
		return
	}

	sbcClient, err := sbcXuanwuClient.NewForConfig(kubeConfig)
	if err != nil {
		log.Errorf("init sbc client failed, err: [%v]", err)
		exporterClientSet.InitError = err
		return
	}
	exporterClientSet.KubeClient = kubeClient
	exporterClientSet.SbcClient = sbcClient
	return
}

// InitExporterClientSet return exporterClientSet. if it not init we will do it
func InitExporterClientSet(grpcSock string) *ClientsSet {
	if exporterClientSet == nil {
		log.Infoln("start to initExporterClientSet")
		once.Do(func() {
			exporterClientSet = &ClientsSet{}
			grpcClientSet, err := storageGRPC.GetClientSet(grpcSock)
			if err != nil {
				log.Errorln("can not get Client")
				exporterClientSet.InitError = err
				return
			}
			exporterClientSet.StorageGRPCClientSet = grpcClientSet
			initKubeClientAndSbcClient()
		})
	} else {
		log.Debugln("initExporterClientSet is already call")
	}
	return exporterClientSet
}

// DeleteExporterClientSet Release the exporterClientSet
func DeleteExporterClientSet() {
	if exporterClientSet == nil {
		return
	}
	if exporterClientSet.StorageGRPCClientSet == nil {
		return
	}
	if exporterClientSet.StorageGRPCClientSet.Conn == nil {
		return
	}
	err := exporterClientSet.StorageGRPCClientSet.Conn.Close()
	if err != nil {
		log.Errorln("can not delete storage grpc Client")
		return
	}
}
