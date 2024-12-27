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

// Package leaderelection offers leader election starter
package leaderelection

import "time"

const (
	defaultNamespaceEnv = "NAMESPACE"
)

// Params leader election parameters
type Params struct {
	leaderLeaseDuration time.Duration
	leaderRenewDeadline time.Duration
	leaderRetryPeriod   time.Duration
	lockName            string
	namespaceEnv        string
	defaultNamespace    string
}

// NewParams returns an init leader election parameters struct
func NewParams() *Params {
	return &Params{
		namespaceEnv: defaultNamespaceEnv,
	}
}

// SetLeaderLeaseDuration sets the leader lease duration
func (p *Params) SetLeaderLeaseDuration(leaderLeaseDuration time.Duration) *Params {
	p.leaderLeaseDuration = leaderLeaseDuration
	return p
}

// SetLeaderRenewDeadline sets the leader renew deadline
func (p *Params) SetLeaderRenewDeadline(leaderRenewDeadline time.Duration) *Params {
	p.leaderRenewDeadline = leaderRenewDeadline
	return p
}

// SetLeaderRetryPeriod sets the leader retry period
func (p *Params) SetLeaderRetryPeriod(leaderRetryPeriod time.Duration) *Params {
	p.leaderRetryPeriod = leaderRetryPeriod
	return p
}

// SetLockName sets the leader lock name
func (p *Params) SetLockName(lockName string) *Params {
	p.lockName = lockName
	return p
}

// SetDefaultNamespace sets default namespace
func (p *Params) SetDefaultNamespace(defaultNamespace string) *Params {
	p.defaultNamespace = defaultNamespace
	return p
}
