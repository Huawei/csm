/*
Copyright 2018 The Kubernetes Authors.

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
	"net"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/utils/log"
	"github.com/huawei/csm/v2/utils/version"
)

const (
	containerName       = "liveness-probe"
	namespaceEnv        = "NAMESPACE"
	defaultNamespace    = "huawei-csm"
	defaultIpAddress    = "0.0.0.0"
	defaultHealthzPort  = "9808"
	defaultProbeTimeout = 10 * time.Second
	defaultCmiAddress   = "/cmi/cmi.sock"
	defaultLogFile      = "liveness-probe"
	versionCmName       = "huawei-csm-version"

	healthz = "/healthz"
)

// Command line flags
var (
	probeTimeout time.Duration
	cmiAddress   string
	ipAddress    string
	healthzPort  string
	logFile      string

	livenessprobe = &cobra.Command{
		Use:  "livenessprobe",
		Long: `liveness probe for CSM services`,
	}
)

type healthProbe struct {
	client *cmi.ClientSet
}

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
			version.CsmLivenessProbeVersion, namespaceEnv, defaultNamespace, versionCmName)
		if err != nil {
			log.Errorf("init version file error: [%v]", err)
			return
		}

		clientSet, err := cmi.GetClientSet(cmiAddress)
		if err != nil {
			log.Errorf("get cmi client set failed: [%v]", err)
			return
		}
		defer clientSet.Conn.Close()

		mux := http.NewServeMux()

		hp := healthProbe{client: clientSet}
		mux.HandleFunc(healthz, hp.probe)

		addr := net.JoinHostPort(ipAddress, healthzPort)
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

func (hp *healthProbe) probe(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), probeTimeout)
	defer cancel()
	log.AddContext(ctx).Infoln("start to probe cmi service")

	_, err := hp.client.IdentityClient.Probe(ctx, &cmi.ProbeRequest{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("probe cmi service failed: [%v]", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.AddContext(ctx).Infoln("probe cmi service succeeded")
}

func parseFlags() {
	livenessprobe.Flags().AddGoFlagSet(flag.CommandLine)
	livenessprobe.Flags().DurationVar(&probeTimeout, "probe-timeout", defaultProbeTimeout,
		"Probe timeout in seconds.")
	livenessprobe.Flags().StringVar(&cmiAddress, "cmi-address", defaultCmiAddress,
		"Address of the CMI driver socket.")
	livenessprobe.Flags().StringVar(&ipAddress, "ip-address", defaultIpAddress,
		"The listening ip address in the container.")
	livenessprobe.Flags().StringVar(&healthzPort, "healthz-port", defaultHealthzPort,
		"TCP ports for listening healthz requests.")
	livenessprobe.Flags().StringVar(&logFile, "log-file", defaultLogFile,
		"The log file name of the liveness probe")
}
