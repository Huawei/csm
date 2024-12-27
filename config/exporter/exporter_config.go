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

// Package exporter is to init configuration and flags for resource prometheus exporter
package exporter

import (
	"github.com/spf13/pflag"

	confConsts "github.com/huawei/csm/v2/config/consts"
)

const (
	defaultIpAddress               = "0.0.0.0"
	defaultExporterPort            = "8887"
	controllerOptionName           = "PrometheusExporterOption"
	defaultStorageGRPCSock         = "/var/cmi/cmi.sock"
	defaultStorageBackendNamespace = "huawei-csi"
	defaultCSIDriverName           = "csi.huawei.com"
	defaultUseHttps                = true
)

var (
	// Option is a prometheus exporter option instance fot manager init
	Option = &option{}
)

type option struct {
	ipAddress               string
	exporterPort            string
	storageGRPCSock         string
	storageBackendNamespace string
	csiDriverName           string
	useHttps                bool
}

// GetName return the name string of the ControllerOption
func (o *option) GetName() string {
	return controllerOptionName
}

// AddFlags is to add flags for the resource topology controller configurations
func (o *option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ipAddress, confConsts.IpAddress, defaultIpAddress,
		"The listening ip address.")
	fs.StringVar(&o.exporterPort, confConsts.ExporterPort, defaultExporterPort,
		"The exporter port in the container.")
	fs.StringVar(&o.storageGRPCSock, confConsts.StorageGRPCSock, defaultStorageGRPCSock,
		"The storage grpc client sock file name.")
	fs.StringVar(&o.storageBackendNamespace, confConsts.StorageBackendNamespace, defaultStorageBackendNamespace,
		"The storage backend namespace name.")
	fs.StringVar(&o.csiDriverName, confConsts.CSIDriverName, defaultCSIDriverName,
		"The CSI driver name.")
	fs.BoolVar(&o.useHttps, confConsts.UseHttps, defaultUseHttps,
		"Use https or not.")
}

// ValidateConfig is to validate input resource topology controller configurations
func (o *option) ValidateConfig() error {
	return nil
}

// GetIpAddress returns the listening ip address
func GetIpAddress() string {
	return Option.ipAddress
}

// GetExporterPort returns the exporter port
func GetExporterPort() string {
	return Option.exporterPort
}

// GetStorageGRPCSock returns the storage GRPC sock
func GetStorageGRPCSock() string {
	return Option.storageGRPCSock
}

// GetStorageBackendNamespace returns the storage backend namespace
func GetStorageBackendNamespace() string {
	return Option.storageBackendNamespace
}

// GetCSIDriverName returns the storage backend namespace
func GetCSIDriverName() string {
	return Option.csiDriverName
}

// GetUseHttps returns use https or not
func GetUseHttps() bool {
	return Option.useHttps
}
