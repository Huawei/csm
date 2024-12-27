/*
Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

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

// Package version used to set and clean the service version
package version

var (
	buildVersion string
	buildArch    string
)

var (
	common = buildVersion
	// OSArch the architecture of service
	OSArch = buildArch
	// ContainerMonitorInterfaceVersion the version of service containerMonitorInterface
	ContainerMonitorInterfaceVersion = common
	// CsmLivenessProbeVersion the version of service csm-liveness-probe
	CsmLivenessProbeVersion = common
	// CsmTopoServiceVersion the version of service csm-topo-service
	CsmTopoServiceVersion = common
	// CsmPrometheusCollectorVersion the version of service csm-prometheus-collector
	CsmPrometheusCollectorVersion = common
)
