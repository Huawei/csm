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

// Package consts contains all the keys of configuration
package consts

const (
	// LogFile key name of log file config
	LogFile = "log-file"
	// LoggingModule key name of log mod config
	LoggingModule = "logging-module"
	// LogLevel key name of log level config
	LogLevel = "logging-level"
	// LogFileDir key name of log file dir config
	LogFileDir = "log-file-dir"
)

const (
	// SupportResources key name of support resources config
	SupportResources = "support-resources"
	// ControllerWorkers key name of controller works count config
	ControllerWorkers = "controller-workers"
	// RtRetryBaseDelay key name of base rt controller retry delay config
	RtRetryBaseDelay = "rt-retry-base-delay"
	// PvRetryBaseDelay key name of base pv controller retry delay config
	PvRetryBaseDelay = "pv-retry-base-delay"
	// PodRetryBaseDelay key name of base pod controller retry delay config
	PodRetryBaseDelay = "pod-retry-base-delay"
	// RtRetryMaxDelay key name of rt controller max retry delay config
	RtRetryMaxDelay = "rt-retry-max-delay"
	// PvRetryMaxDelay key name of pv controller max retry delay config
	PvRetryMaxDelay = "pv-retry-max-delay"
	// PodRetryMaxDelay key name of pod controller max retry delay config
	PodRetryMaxDelay = "pod-retry-max-delay"
	// ResyncPeriod key name of reSync interval of the controller
	ResyncPeriod = "resync-period"
	// CmiAddress key name of cmi endpoint address
	CmiAddress = "cmi-address"
)

const (
	// KubeConfig key name of kube config file path config
	KubeConfig = "kube-config"
)

const (
	// IpAddress the listening ip address of the prometheus
	IpAddress = "ip-address"
	// ExporterPort prometheus exporter key name of kube config file path config
	ExporterPort = "exporter-port"
	// StorageGRPCSock the path of the grpc sock file
	StorageGRPCSock = "cmi-address"
	// StorageBackendNamespace the namespace of the sbc
	StorageBackendNamespace = "storage-backend-namespace"
	// CSIDriverName the name of the csi driver
	CSIDriverName = "csi-driver-name"
	// UseHttps the option to use https or not
	UseHttps = "use-https"
)

const (
	// EnableLeaderElection key name of controller leader election switch config
	EnableLeaderElection = "enable-leader-election"
	// LeaderLockNamespace key name of controller leader election lock ns config
	LeaderLockNamespace = "leader-lock-namespace"
	// LeaderLeaseDuration key name of controller leader election lease duration config
	LeaderLeaseDuration = "leader-lease-duration"
	// LeaderRenewDeadline key name of controller leader election renew deadline config
	LeaderRenewDeadline = "leader-renew-deadline"
	// LeaderRetryPeriod key name of controller leader election retry period config
	LeaderRetryPeriod = "leader-retry-period"
)
