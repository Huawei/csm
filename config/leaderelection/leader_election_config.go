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

// Package leaderelection
package leaderelection

import (
	"time"

	"github.com/spf13/pflag"

	confConsts "github.com/huawei/csm/v2/config/consts"
)

const (
	leaderElectionOptionName   = "LeaderElectionOption"
	defaultLeaderLockNamespace = "default"
	defaultLeaderLeaseDuration = 8 * time.Second
	defaultLeaderRenewDeadline = 6 * time.Second
	defaultLeaderRetryPeriod   = 2 * time.Second
	enableLeaderElection       = false
)

// Option is a client option instance for manager init
var Option = &option{}

type option struct {
	leaderLeaseDuration  time.Duration
	leaderRenewDeadline  time.Duration
	leaderRetryPeriod    time.Duration
	leaderLockNamespace  string
	enableLeaderElection bool
}

// GetName return name string of client option
func (o *option) GetName() string {
	return leaderElectionOptionName
}

// AddFlags is to add flags for client configurations
func (o *option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.enableLeaderElection, confConsts.EnableLeaderElection, enableLeaderElection,
		"Start a leader election client and gain leadership for controller")
	fs.StringVar(&o.leaderLockNamespace, confConsts.LeaderLockNamespace, defaultLeaderLockNamespace,
		"Configure leader election lock namespace")
	fs.DurationVar(&o.leaderLeaseDuration, confConsts.LeaderLeaseDuration, defaultLeaderLeaseDuration,
		"Configure leader election lease duration")
	fs.DurationVar(&o.leaderRenewDeadline, confConsts.LeaderRenewDeadline, defaultLeaderRenewDeadline,
		"Configure leader election lease renew deadline")
	fs.DurationVar(&o.leaderRetryPeriod, confConsts.LeaderRetryPeriod, defaultLeaderRetryPeriod,
		"Configure leader election lease retry period")
}

// ValidateConfig is to validate input client configurations
func (o *option) ValidateConfig() error {
	return nil
}

// GetLeaderLeaseDuration returns the duration of leader lease
func GetLeaderLeaseDuration() time.Duration {
	return Option.leaderLeaseDuration
}

// GetLeaderRenewDeadline returns the deadline of renew a leader
func GetLeaderRenewDeadline() time.Duration {
	return Option.leaderRenewDeadline
}

// GetLeaderRetryPeriod returns leader retry period
func GetLeaderRetryPeriod() time.Duration {
	return Option.leaderRetryPeriod
}

// EnableLeaderElection returns leader election is on
func EnableLeaderElection() bool {
	return Option.enableLeaderElection
}

// GetLeaderLockNamespace returns the leader lock namespace
func GetLeaderLockNamespace() string {
	return Option.leaderLockNamespace
}
