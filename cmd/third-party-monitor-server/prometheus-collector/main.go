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

// Package main is the process entry
package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/huawei/csm/v2/config"
	clientConfig "github.com/huawei/csm/v2/config/client"
	"github.com/huawei/csm/v2/config/cmi"
	"github.com/huawei/csm/v2/config/common"
	exporterConfig "github.com/huawei/csm/v2/config/exporter"
	logConfig "github.com/huawei/csm/v2/config/log"
	_ "github.com/huawei/csm/v2/provider/collectdef/defs"
	clientSet "github.com/huawei/csm/v2/server/prometheus-exporter/clientset"
	exporterHandler "github.com/huawei/csm/v2/server/prometheus-exporter/exporterhandler"
	"github.com/huawei/csm/v2/utils/log"
	"github.com/huawei/csm/v2/utils/version"
)

const (
	containerName = "prometheus-collector"
	namespaceEnv  = "NAMESPACE"
	versionCmName = "huawei-csm-version"

	healthz = "/healthz"

	httpsCert = "/etc/secret-volume/tls.crt"
	httpsKey  = "/etc/secret-volume/tls.key"
)

var prometheusExporter = &cobra.Command{
	Use:  "exporter",
	Long: `prometheus exporter`,
}

func main() {
	manager := config.NewOptionManager(prometheusExporter.Flags(), logConfig.Option, clientConfig.Option,
		exporterConfig.Option, common.Option, cmi.Option)
	manager.AddFlags()

	prometheusExporter.Run = func(cmd *cobra.Command, args []string) {
		err := manager.ValidateConfig()
		if err != nil {
			// ValidateConfig error the log ValidateConfig will print
			logrus.Errorf("validate config err: [%v]", err)
			return
		}

		err = verifyStartInfo()
		if err != nil {
			// error log in verifyStartInfo
			logrus.Errorf("verify start info err: [%v]", err)
			return
		}

		err = initListener()
		if err != nil {
			// error log in initListener
			log.Errorf("init listener err: [%v]", err)
			return
		}
	}

	if err := prometheusExporter.Execute(); err != nil {
		log.Errorf("server meet err: [%v], exit", err)
		return
	}
}

func verifyStartInfo() error {
	err := log.InitLogging(logConfig.GetLogFile())
	if err != nil {
		logrus.Errorf("init log config err: [%v]", err)
		return err
	}

	err = version.InitVersionConfigMapWithName(containerName,
		version.CsmPrometheusCollectorVersion, namespaceEnv, common.GetNamespace(), versionCmName)
	if err != nil {
		log.Errorf("init version file error: [%v]", err)
		return err
	}
	return nil
}

func initListener() error {
	client := clientSet.InitExporterClientSet()
	if client.InitError != nil {
		clientSet.DeleteExporterClientSet()
		return fmt.Errorf("init exporter client set err: [%v]", client.InitError)
	}

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	shutdownChan := make(chan struct{})
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Infoln("received shutdown signal, cleaning up...")
		close(shutdownChan)
	}()

	http.HandleFunc("/", exporterHandler.MetricsHandler)
	http.HandleFunc(healthz, exporterHandler.HealthHandler)
	server := &http.Server{
		Addr: exporterConfig.GetIpAddress() + ":" + exporterConfig.GetExporterPort(),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	errCh := make(chan error, 1)
	go func() {
		if exporterConfig.GetUseHttps() {
			errCh <- server.ListenAndServeTLS(httpsCert, httpsKey)
		} else {
			errCh <- server.ListenAndServe()
		}
	}()

	select {
	case <-shutdownChan:
		clientSet.DeleteExporterClientSet()
		return nil
	case err := <-errCh:
		clientSet.DeleteExporterClientSet()
		return fmt.Errorf("start service error: %v", err)
	}
}
