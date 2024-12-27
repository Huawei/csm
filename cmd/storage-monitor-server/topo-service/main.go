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

// Package main is the process entry
package main

import (
	"context"
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	k8sInformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/huawei/csm/v2/config"
	clientConfig "github.com/huawei/csm/v2/config/client"
	leaderElectionConfig "github.com/huawei/csm/v2/config/leaderelection"
	logConfig "github.com/huawei/csm/v2/config/log"
	controllerConfig "github.com/huawei/csm/v2/config/topology"
	leaderElection "github.com/huawei/csm/v2/controller/leaderelection"
	"github.com/huawei/csm/v2/controller/resourcetopology"
	"github.com/huawei/csm/v2/controller/utils"
	csmScheme "github.com/huawei/csm/v2/pkg/client/clientset/versioned/scheme"
	informers "github.com/huawei/csm/v2/pkg/client/informers/externalversions"
	"github.com/huawei/csm/v2/utils/log"
	"github.com/huawei/csm/v2/utils/version"
)

const (
	defaultNamespace     = "huawei-csm"
	containerName        = "topo-service"
	leaderLockObjectName = "resource-topology"
	namespaceEnv         = "NAMESPACE"
	versionCmName        = "huawei-csm-version"
)

var topoService = &cobra.Command{
	Use:  "controller",
	Long: `resource topology controller`,
}

func main() {
	manager := config.NewOptionManager(topoService.Flags(),
		logConfig.Option, controllerConfig.Option, leaderElectionConfig.Option, clientConfig.Option)
	manager.AddFlags()

	topoService.Run = func(cmd *cobra.Command, args []string) {
		err := manager.ValidateConfig()
		if err != nil {
			logrus.Errorf("validate config err: [%v]", err)
			return
		}

		err = log.InitLogging(logConfig.GetLogFile())
		if err != nil {
			logrus.Errorf("init log config err: [%v]", err)
			return
		}

		err = version.InitVersionConfigMapWithName(containerName,
			version.CsmTopoServiceVersion, namespaceEnv, defaultNamespace, versionCmName)
		if err != nil {
			log.Errorf("init version file error: [%v]", err)
			return
		}

		clientsSet, err := utils.NewClientsSet(clientConfig.GetKubeConfig(), controllerConfig.GetCmiAddress())
		if err != nil {
			log.Errorf("new client set error: [%v]", err)
			return
		}

		ctx := context.WithValue(context.Background(), "controller", "resourceTopologyController")

		signalChan := make(chan os.Signal, 1)
		defer close(signalChan)

		startController(ctx, clientsSet, signalChan)

		err = waitTopoServiceStop(ctx, signalChan)
		if err != nil {
			log.Errorf("wait topo service stop error: [%v]", err)
			return
		}
	}

	if err := topoService.Execute(); err != nil {
		log.Errorf("server meet err: [%v], exit", err)
		return
	}
}

func startController(ctx context.Context, clientSets *utils.ClientsSet, ch chan os.Signal) {
	if leaderElectionConfig.EnableLeaderElection() {
		go leaderElection.Run(ctx, clientSets, newLeaderElectionParams(), runController, ch)
	} else {
		log.AddContext(ctx).Infoln("start resourceTopology controller without leader election")
		go runController(ctx, clientSets, ch)
	}
}

func waitTopoServiceStop(ctx context.Context, signalChan chan os.Signal) error {
	// Stop the main when stop signals are received
	utils.WaitSignal(ctx, signalChan)
	return nil
}

func newLeaderElectionParams() *leaderElection.Params {
	return leaderElection.NewParams().
		SetDefaultNamespace(leaderElectionConfig.GetLeaderLockNamespace()).
		SetLockName(leaderLockObjectName).
		SetLeaderLeaseDuration(leaderElectionConfig.GetLeaderLeaseDuration()).
		SetLeaderRenewDeadline(leaderElectionConfig.GetLeaderRenewDeadline()).
		SetLeaderRetryPeriod(leaderElectionConfig.GetLeaderRetryPeriod())
}

func runController(ctx context.Context, clients *utils.ClientsSet, ch chan os.Signal) {
	factory := informers.NewSharedInformerFactory(clients.XuanwuClient, controllerConfig.GetResyncPeriod())
	k8sFactory := k8sInformers.NewSharedInformerFactory(clients.KubeClient, 0)
	// Add ResourceTopology types to the default Kubernetes so events can be logged for them
	if err := csmScheme.AddToScheme(scheme.Scheme); err != nil {
		log.AddContext(ctx).Errorf("add to scheme error: %v", err)
		ch <- syscall.SIGINT
		return
	}

	ctrl := resourcetopology.NewController(resourcetopology.ControllerRequest{
		KubeClient:       clients.KubeClient,
		XuanwuClient:     clients.XuanwuClient,
		TopologyInformer: factory.Xuanwu().V1().ResourceTopologies(),
		VolumeInformer:   k8sFactory.Core().V1().PersistentVolumes(),
		ClaimInformer:    k8sFactory.Core().V1().PersistentVolumeClaims(),
		PodInformer:      k8sFactory.Core().V1().Pods(),
		ReSyncPeriod:     controllerConfig.GetResyncPeriod(),
		EventRecorder:    clients.EventRecorder,
		CmiClient:        clients.CmiClient,
	})

	run := func(ctx context.Context) {
		// Run the controller process
		stopCh := make(chan struct{})
		factory.Start(stopCh)
		k8sFactory.Start(stopCh)
		go ctrl.Run(ctx, controllerConfig.GetControllerWorkers(), stopCh)

		// Stop the controller until get signal
		utils.WaitExitSignal(ctx, "controller")

		close(stopCh)
	}

	run(ctx)
}
