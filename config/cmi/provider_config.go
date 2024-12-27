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

// Package cmi defines config of cmi service
package cmi

import (
	"github.com/spf13/pflag"
)

const (
	defaultQueryPageSize      = 100
	defaultClientMaxThreads   = 20
	defaultProviderName       = "cmi.huawei.com"
	defaultProviderOptionName = "providerOptionName"
	defaultCmiAddress         = "/cmi/cmi.sock"
	defaultNamespace          = "huawei-csi"
)

// Option contains provider option args
var Option = NewProviderOption()

type providerOption struct {
	queryStoragePageSize int
	clientMaxThreads     int
	providerName         string
	cmiAddress           string
	backendNamespace     string
}

// GetName return option name
func (p *providerOption) GetName() string {
	return defaultProviderOptionName
}

// AddFlags add flags
func (p *providerOption) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&p.providerName, "cmi-name", defaultProviderName, "Name of provider")
	fs.StringVar(&p.cmiAddress, "cmi-address", defaultCmiAddress, "Path to cmi socket")
	fs.IntVar(&p.queryStoragePageSize, "page-size", defaultQueryPageSize, "Max size of query storage")
	fs.StringVar(&p.backendNamespace, "backend-namespace", defaultNamespace, "Namespace of backend")
	fs.IntVar(&p.clientMaxThreads, "client-max-threads", defaultClientMaxThreads, "Max client threads")
}

// ValidateConfig validate config
func (p *providerOption) ValidateConfig() error {
	return nil
}

// NewProviderOption init an instance of ProviderOption
func NewProviderOption() *providerOption {
	return &providerOption{
		queryStoragePageSize: defaultQueryPageSize,
		providerName:         defaultProviderName,
		cmiAddress:           defaultCmiAddress,
	}
}

// GetProviderName get provider name
func GetProviderName() string {
	return Option.providerName
}

// GetCmiAddress get cmi address
func GetCmiAddress() string {
	return Option.cmiAddress
}

// GetNamespace get namespace
func GetNamespace() string {
	return Option.backendNamespace
}

// GetQueryStoragePageSize get query storage page size
func GetQueryStoragePageSize() int {
	return Option.queryStoragePageSize
}

// GetClientMaxThreads get client max threads
func GetClientMaxThreads() int {
	return Option.clientMaxThreads
}
