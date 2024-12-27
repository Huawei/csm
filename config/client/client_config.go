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

// Package client is used to init client configurations and flags
package client

import (
	"errors"
	"path/filepath"

	"github.com/spf13/pflag"

	"github.com/huawei/csm/v2/config/consts"
)

const (
	clientOptionName = "ClientOption"
)

// Option is a client option instance for manager init
var Option = &option{}

type option struct {
	kubeConfig string
}

// GetName return name string of client option
func (o *option) GetName() string {
	return clientOptionName
}

// AddFlags is to add flags for client configurations
func (o *option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeConfig, consts.KubeConfig, "", "The absolute path to the kubeConfig file.")
}

// ValidateConfig is to validate input client configurations
func (o *option) ValidateConfig() error {
	if o.kubeConfig == "" {
		return nil
	}
	if !filepath.IsAbs(o.kubeConfig) {
		return errors.New("kubeConfig file path is not absolute")
	}
	return nil
}

// GetKubeConfig returns the kube config file path
func GetKubeConfig() string {
	return Option.kubeConfig
}
