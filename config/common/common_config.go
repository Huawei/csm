/*
 Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

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

// Package common is used to init common configurations and flags
package common

import (
	"github.com/spf13/pflag"

	"github.com/huawei/csm/v2/config/consts"
)

const (
	defaultNamespace = "huawei-csm"
	commonOptionName = "CommonOption"
)

var (
	// Option is a prometheus exporter option instance fot manager init
	Option = &option{}
)

type option struct {
	namcespace string
}

// GetName return name string of common option
func (o *option) GetName() string {
	return commonOptionName
}

// AddFlags is to add flags for client configurations
func (o *option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.namcespace, consts.CSMNamespace, defaultNamespace,
		"The namespace of csm")
}

// GetNamespace returns the storage backend namespace
func GetNamespace() string {
	return Option.namcespace
}

// ValidateConfig is to validate input common configurations
func (o *option) ValidateConfig() error {
	return nil
}
