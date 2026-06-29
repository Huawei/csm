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
	"fmt"
	"math"
	"path/filepath"

	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"

	"github.com/huawei/csm/v2/config/consts"
)

const (
	clientOptionName    = "ClientOption"
	defaultKubeAPIQPS   = 5
	defaultKubeAPIBurst = 10
)

// Option is a client option instance for manager init
var Option = &option{}

type option struct {
	kubeConfig   string
	kubeAPIQPS   float64
	kubeAPIBurst int
}

// GetName return name string of client option
func (o *option) GetName() string {
	return clientOptionName
}

// AddFlags is to add flags for client configurations
func (o *option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeConfig, consts.KubeConfig, "", "The absolute path to the kubeConfig file.")
	fs.Float64Var(&o.kubeAPIQPS, consts.KubeAPIQPS, defaultKubeAPIQPS,
		"Indicates the maximum QPS in kubernetes client.")
	fs.IntVar(&o.kubeAPIBurst, consts.KubeAPIBurst, defaultKubeAPIBurst,
		"Indicates the maximum burst for throttle in kubernetes client.")
}

// ValidateConfig is to validate input client configurations
func (o *option) ValidateConfig() error {
	if o.kubeConfig != "" && !filepath.IsAbs(o.kubeConfig) {
		return fmt.Errorf("invalid kubeConfig path: %s (must be absolute path)", o.kubeConfig)
	}
	if o.kubeAPIBurst < 0 {
		return fmt.Errorf("invalid kube-api-burst value: %d (must be >= 0)", o.kubeAPIBurst)
	}
	if o.kubeAPIQPS < 0 {
		return fmt.Errorf("invalid kube-api-qps value: %f (must be >= 0)", o.kubeAPIQPS)
	}
	if o.kubeAPIQPS > 0 && float64(o.kubeAPIBurst) <= o.kubeAPIQPS {
		return fmt.Errorf("invalid kube-api-burst value: %d (must be > kube-api-qps: %.2f)", o.kubeAPIBurst,
			o.kubeAPIQPS)
	}
	return nil
}

// GetKubeConfig returns the kube config file path
func GetKubeConfig() string {
	return Option.kubeConfig
}

// GetKubeAPIQPS returns the k8s client QPS value
func GetKubeAPIQPS() float64 {
	return Option.kubeAPIQPS
}

// GetKubeAPIBurst returns the k8s client burst value
func GetKubeAPIBurst() int {
	return Option.kubeAPIBurst
}

// ApplyKubeAPIQPSBurst applies QPS and Burst settings to the given rest.Config
func ApplyKubeAPIQPSBurst(config *rest.Config) {
	if config == nil {
		return
	}
	if Option.kubeAPIQPS > 0.0 {
		if Option.kubeAPIQPS > math.MaxFloat32 {
			config.QPS = math.MaxFloat32
		} else {
			config.QPS = float32(Option.kubeAPIQPS)
		}
	}
	if Option.kubeAPIBurst > 0 {
		config.Burst = Option.kubeAPIBurst
	}
}
