/*
 Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.

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

// Package topology is to init configuration and flags for resource topology controller
package topology

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/pflag"

	confConsts "github.com/huawei/csm/v2/config/consts"
)

const (
	defaultControllerWorkers = 4
	controllerOptionName     = "ControllerOption"
	defaultRtRetryBaseDelay  = 5 * time.Second
	defaultPvRetryBaseDelay  = 5 * time.Second
	defaultPodRetryBaseDelay = 5 * time.Second
	defaultRtRetryMaxDelay   = 5 * time.Minute
	defaultPvRetryMaxDelay   = 1 * time.Minute
	defaultPodRetryMaxDelay  = 1 * time.Minute
	defaultResyncPeriod      = 15 * time.Minute
	defaultCmiAddress        = "/cmi/cmi.sock"
	minSupportResourceNum    = 2
	defaultCSIDriverName     = "csi.huawei.com"
	defaultCSINamespace      = "huawei-csi"
	minResyncPeriod          = 5 * time.Minute
)

var (
	defaultSupportResources = []string{"Pod", "PersistentVolume"}
	// Option is a controller option instance fot manager init
	Option = &option{}
)

type option struct {
	supportResources  []string
	rtRetryBaseDelay  time.Duration
	pvRetryBaseDelay  time.Duration
	podRetryBaseDelay time.Duration
	rtRetryMaxDelay   time.Duration
	pvRetryMaxDelay   time.Duration
	podRetryMaxDelay  time.Duration
	resyncPeriod      time.Duration
	cmiAddress        string
	controllerWorkers int
	csiDriverName     string
	backendNamespace  string
}

// GetName return the name string of the ControllerOption
func (o *option) GetName() string {
	return controllerOptionName
}

// AddFlags is to add flags for the resource topology controller configurations
func (o *option) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&o.controllerWorkers, confConsts.ControllerWorkers, defaultControllerWorkers,
		"Number of worker for controller")
	fs.StringArrayVar(&o.supportResources, confConsts.SupportResources, defaultSupportResources,
		"Define which resources can be added to tags. Example: --supportedResources=Pod,PersistentVolume")
	fs.DurationVar(&o.rtRetryBaseDelay, confConsts.RtRetryBaseDelay, defaultRtRetryBaseDelay,
		"Base retry delay of failed resourceTopology creation or deletion. "+
			"It doubles with each failure, up to rt-retry-interval-max.")
	fs.DurationVar(&o.pvRetryBaseDelay, confConsts.PvRetryBaseDelay, defaultPvRetryBaseDelay,
		"Base retry delay of failed pv work task. "+
			"It doubles with each failure, up to pv-retry-interval-max.")
	fs.DurationVar(&o.podRetryBaseDelay, confConsts.PodRetryBaseDelay, defaultPodRetryBaseDelay,
		"Base retry delay of failed pod work task. "+
			"It doubles with each failure, up to pod-retry-interval-max.")
	fs.DurationVar(&o.rtRetryMaxDelay, confConsts.RtRetryMaxDelay, defaultRtRetryMaxDelay,
		"Maximum retry delay of failed resourceTopology creation or deletion.")
	fs.DurationVar(&o.pvRetryMaxDelay, confConsts.PvRetryMaxDelay, defaultPvRetryMaxDelay,
		"Maximum retry delay of failed pv work task.")
	fs.DurationVar(&o.podRetryMaxDelay, confConsts.PodRetryMaxDelay, defaultPodRetryMaxDelay,
		"Maximum retry delay of failed pod work task.")
	fs.DurationVar(&o.resyncPeriod, confConsts.ResyncPeriod, defaultResyncPeriod,
		"The reSync interval of the controller.")
	fs.StringVar(&o.cmiAddress, confConsts.CmiAddress, defaultCmiAddress,
		"The socket address of container monitoring interface.")
	fs.StringVar(&o.csiDriverName, confConsts.CSIDriverName, defaultCSIDriverName,
		"The CSI driver name.")
	fs.StringVar(&o.backendNamespace, "backend-namespace", defaultCSINamespace, "Namespace of backend.")
}

// ValidateConfig is to validate input resource topology controller configurations
func (o *option) ValidateConfig() error {
	workers := o.controllerWorkers
	if workers < 1 {
		return fmt.Errorf("invalid controller workers count [%d]", workers)
	}

	if len(o.supportResources) < minSupportResourceNum {
		return errors.New("supported resources should be at least 2")
	}

	if o.rtRetryMaxDelay < o.rtRetryBaseDelay {
		return fmt.Errorf("rt retry max delay [%s] is less than rt retry base delay [%s]",
			o.rtRetryMaxDelay, o.rtRetryBaseDelay)
	}

	if o.pvRetryMaxDelay < o.pvRetryBaseDelay {
		return fmt.Errorf("pv retry max delay [%s] is less than pv retry base delay [%s]",
			o.pvRetryMaxDelay, o.pvRetryBaseDelay)
	}

	if o.podRetryMaxDelay < o.podRetryBaseDelay {
		return fmt.Errorf("pod retry max delay [%s] is less than pod retry base delay [%s]",
			o.podRetryMaxDelay, o.podRetryBaseDelay)
	}

	if o.resyncPeriod <= minResyncPeriod {
		return fmt.Errorf("resync period [%s] is less than min resync period [%s]",
			o.resyncPeriod, minResyncPeriod)
	}

	return nil
}

// GetControllerWorkers returns the number of controller workers
func GetControllerWorkers() int {
	return Option.controllerWorkers
}

// GetSupportResources returns the supported resource name list
func GetSupportResources() []string {
	return Option.supportResources
}

// GetRtRetryBaseDelay returns the base retry delay of rt controller
func GetRtRetryBaseDelay() time.Duration {
	return Option.rtRetryBaseDelay
}

// GetRtRetryMaxDelay returns the max retry delay of rt controller
func GetRtRetryMaxDelay() time.Duration {
	return Option.rtRetryMaxDelay
}

// GetPvRetryBaseDelay returns the base retry delay of pv controller
func GetPvRetryBaseDelay() time.Duration {
	return Option.pvRetryBaseDelay
}

// GetPvRetryMaxDelay returns the max retry delay of pv controller
func GetPvRetryMaxDelay() time.Duration {
	return Option.pvRetryMaxDelay
}

// GetPodRetryBaseDelay returns the base retry delay of pod controller
func GetPodRetryBaseDelay() time.Duration {
	return Option.podRetryBaseDelay
}

// GetPodRetryMaxDelay returns the max retry delay of pod controller
func GetPodRetryMaxDelay() time.Duration {
	return Option.podRetryMaxDelay
}

// GetResyncPeriod returns the resync interval
func GetResyncPeriod() time.Duration {
	return Option.resyncPeriod
}

// GetCmiAddress returns container monitoring interface address
func GetCmiAddress() string {
	return Option.cmiAddress
}

// GetCSIDriverName returns the csi driver name
func GetCSIDriverName() string {
	return Option.csiDriverName
}

// GetBackendNamespace returns the storage backend namespace
func GetBackendNamespace() string {
	return Option.backendNamespace
}
