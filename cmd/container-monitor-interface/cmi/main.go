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

// Package main is the process entry
package main

import (
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/huawei/csm/v2/config"
	cmiConfig "github.com/huawei/csm/v2/config/cmi"
	"github.com/huawei/csm/v2/config/common"
	logConfig "github.com/huawei/csm/v2/config/log"
	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/collect"
	grpchelper "github.com/huawei/csm/v2/provider/grpc/helper"
	"github.com/huawei/csm/v2/provider/grpc/server"
	"github.com/huawei/csm/v2/provider/utils"
	"github.com/huawei/csm/v2/utils/log"
	"github.com/huawei/csm/v2/utils/version"
)

const (
	containerName = "cmi-controller"
	namespaceEnv  = "NAMESPACE"
	versionCmName = "huawei-csm-version"
)

var cmiService = &cobra.Command{
	Use:  "cmi",
	Long: `container monitor interface`,
}

func main() {
	manager := config.NewOptionManager(cmiService.Flags(), logConfig.Option, cmiConfig.Option, common.Option)
	manager.AddFlags()

	cmiService.Run = func(cmd *cobra.Command, args []string) {
		err := manager.ValidateConfig()
		if err != nil {
			log.Errorf("validate config failed, error: %v", err)
			return
		}

		err = log.InitLogging(logConfig.GetLogFile())
		if err != nil {
			logrus.Errorf("init log config err: %v", err)
			return
		}

		err = version.InitVersionConfigMapWithName(containerName,
			version.ContainerMonitorInterfaceVersion, namespaceEnv, common.GetNamespace(), versionCmName)
		if err != nil {
			log.Errorf("init version file error: [%v]", err)
			return
		}

		err = grpchelper.InitClientSet()
		if err != nil {
			log.Errorf("init client set failed, error: %v", err)
			return
		}

		stopCh := make(chan struct{})
		defer close(stopCh)
		startBackendWatcher(stopCh)
		err = StartGrpcServer(cmiConfig.GetCmiAddress())
		if err != nil {
			log.Errorf("start grpc server failed, error: %v", err)
			return
		}
	}

	if err := cmiService.Execute(); err != nil {
		log.Errorf("Start cmi server failed, error: %v", err)
		return
	}
}

// StartGrpcServer start grpc server
func StartGrpcServer(address string) error {
	log.Infoln("Starting cmi server")
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(log.EnsureGRPCContext),
	}
	grpcServer := grpc.NewServer(opts...)

	cmi.RegisterIdentityServer(grpcServer, &server.Identity{})
	cmi.RegisterLabelServiceServer(grpcServer, &server.Label{})
	cmi.RegisterCollectorServer(grpcServer, &server.Collector{})

	if err := utils.CleanupSocketFile(address); err != nil {
		return fmt.Errorf("cleanup unix socket failed, error: %v", err)
	}

	lis, err := net.Listen("unix", address)
	if err != nil {
		return fmt.Errorf("listen unix socket failed, error: %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Errorf("cmi server stopped serving, error: %v", err)
			signalChan <- syscall.SIGINT
		}
	}()

	defer func() {
		grpcServer.GracefulStop()
	}()

	// terminate grpc server gracefully before leaving main function
	<-signalChan

	return nil
}

func startBackendWatcher(stopCh chan struct{}) {
	go collect.RunBackendInformer(stopCh)
}
