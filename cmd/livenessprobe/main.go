/*
Copyright (c) Huawei Technologies Co., Ltd. 2018-2026. All rights reserved.

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
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/huawei/csm/v2/config/consts"
	"github.com/huawei/csm/v2/utils/log"
	"github.com/huawei/csm/v2/utils/version"
)

const (
	containerName      = "liveness-probe"
	namespaceEnv       = "NAMESPACE"
	defaultNamespace   = "huawei-csm"
	defaultIpAddress   = "0.0.0.0"
	defaultHealthzPort = "9808"
	defaultLogFile     = "liveness-probe"
	versionCmName      = "huawei-csm-version"

	healthz = "/healthz"
	timeout = 10 * time.Second
)

// Command line flags
var (
	ipAddress   string
	healthzPort string
	logFile     string
	namespace   string

	livenessprobe = &cobra.Command{
		Use:  "livenessprobe",
		Long: `liveness probe for CSM services`,
	}
)

func main() {
	parseFlags()

	livenessprobe.Run = func(cmd *cobra.Command, args []string) {
		// Init the logging
		err := log.InitLogging(logFile)
		if err != nil {
			log.Errorf("init log error: [%v]", err)
			return
		}

		err = version.InitVersionConfigMapWithName(containerName,
			version.CsmLivenessProbeVersion, namespaceEnv, namespace, versionCmName)
		if err != nil {
			log.Errorf("init version file error: [%v]", err)
			return
		}

		mux := http.NewServeMux()
		mux.HandleFunc(healthz, probe)

		addr := fmt.Sprintf("%s:%s", ipAddress, healthzPort)
		log.Infof("serveMux listening at [%s]", addr)
		err = http.ListenAndServe(addr, mux)
		if err != nil {
			log.Errorf("failed to start http server with error: [%v]", err)
			return
		}
	}

	if err := livenessprobe.Execute(); err != nil {
		log.Errorf("start liveness probe server failed, error: [%v]", err)
		return
	}
}

// probe is a minimal health check that always returns success.
// In the in-process architecture, the process itself is the health indicator.
// Kubernetes handles process liveness natively via the container runtime.
func probe(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), timeout)
	defer cancel()
	log.AddContext(ctx).Infoln("liveness probe check passed - process is alive")
	w.WriteHeader(http.StatusOK)
}

func parseFlags() {
	livenessprobe.Flags().AddGoFlagSet(flag.CommandLine)
	livenessprobe.Flags().StringVar(&ipAddress, "ip-address", defaultIpAddress,
		"The listening ip address in the container.")
	livenessprobe.Flags().StringVar(&healthzPort, "healthz-port", defaultHealthzPort,
		"TCP ports for listening healthz requests.")
	livenessprobe.Flags().StringVar(&logFile, "log-file", defaultLogFile,
		"The log file name of the liveness probe")
	livenessprobe.Flags().StringVar(&namespace, consts.CSMNamespace, defaultNamespace,
		"Namespace of the csm")
}
