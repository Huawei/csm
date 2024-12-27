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

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"

	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"

	"github.com/huawei/csm/v2/controller/utils"
	"github.com/huawei/csm/v2/utils/log"
)

type leaderElectionRunner struct {
	params *Params

	leaderElector        *leaderelection.LeaderElector
	hostname             string
	resLockConfig        resourcelock.ResourceLockConfig
	resLock              resourcelock.Interface
	leaderElectionConfig leaderelection.LeaderElectionConfig

	err error
}

// Run will run the func with leader election
func Run(ctx context.Context, clientSet *utils.ClientsSet, params *Params,
	runFunc func(context.Context, *utils.ClientsSet, chan os.Signal), ch chan os.Signal) {
	runner := &leaderElectionRunner{params: params}
	runner.channelCheck(ctx, ch).
		setHostname(ctx).
		setResLockConfig(clientSet.EventRecorder).
		setResLock(ctx, clientSet).
		setLeaderElectionConfig(ctx, runFunc, clientSet, ch).
		setLeaderElector(ctx).
		run(ctx, ch)
}

func (runner *leaderElectionRunner) channelCheck(ctx context.Context, ch chan os.Signal) *leaderElectionRunner {
	if ch == nil {
		errMsg := "the channel should not be nil"
		log.AddContext(ctx).Errorln(errMsg)
		runner.err = errors.New(errMsg)
	}
	return runner
}

func (runner *leaderElectionRunner) setHostname(ctx context.Context) *leaderElectionRunner {
	if runner.err != nil {
		return runner
	}

	hostname, err := os.Hostname()
	if err != nil {
		errMsg := fmt.Sprintf("error getting hostname: [%v]", err)
		log.AddContext(ctx).Errorln(errMsg)
		runner.err = errors.New(errMsg)
		return runner
	}

	runner.hostname = hostname
	return runner
}

func (runner *leaderElectionRunner) setResLockConfig(recorder record.EventRecorder) *leaderElectionRunner {
	if runner.err != nil {
		return runner
	}

	runner.resLockConfig = resourcelock.ResourceLockConfig{
		Identity:      runner.hostname,
		EventRecorder: recorder,
	}

	return runner
}

func (runner *leaderElectionRunner) setResLock(ctx context.Context, clientSet *utils.ClientsSet) *leaderElectionRunner {
	if runner.err != nil {
		return runner
	}

	lock, err := resourcelock.New(
		resourcelock.LeasesResourceLock,
		utils.GetNameSpaceFromEnv(runner.params.namespaceEnv, runner.params.defaultNamespace),
		runner.params.lockName,
		clientSet.KubeClient.CoreV1(),
		clientSet.KubeClient.CoordinationV1(),
		runner.resLockConfig)
	if err != nil {
		errMsg := fmt.Sprintf("error creating resource lock: [%v]", err)
		log.AddContext(ctx).Errorln(errMsg)
		runner.err = errors.New(errMsg)
		return runner
	}

	runner.resLock = lock
	return runner
}

func (runner *leaderElectionRunner) setLeaderElectionConfig(ctx context.Context,
	runFunc func(context.Context, *utils.ClientsSet, chan os.Signal),
	clientSet *utils.ClientsSet, ch chan os.Signal) *leaderElectionRunner {
	if runner.err != nil {
		return runner
	}

	runner.leaderElectionConfig = leaderelection.LeaderElectionConfig{
		Lock:          runner.resLock,
		LeaseDuration: runner.params.leaderLeaseDuration,
		RenewDeadline: runner.params.leaderRenewDeadline,
		RetryPeriod:   runner.params.leaderRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				runFunc(ctx, clientSet, ch)
			},
			OnStoppedLeading: func() {
				log.AddContext(ctx).Errorln("controller manager lost master")
				ch <- syscall.SIGINT
			},
			OnNewLeader: func(identity string) {
				log.AddContext(ctx).Infof("new leader elected, current leader [%s]", identity)
			},
		},
	}
	return runner
}

func (runner *leaderElectionRunner) setLeaderElector(ctx context.Context) *leaderElectionRunner {
	if runner.err != nil {
		return runner
	}

	leaderElector, err := leaderelection.NewLeaderElector(runner.leaderElectionConfig)
	if err != nil {
		errMsg := fmt.Sprintf("error creating leader elector: [%v]", err)
		log.AddContext(ctx).Errorln(errMsg)
		runner.err = errors.New(errMsg)
		return runner
	}

	runner.leaderElector = leaderElector
	return runner
}

func (runner *leaderElectionRunner) run(ctx context.Context, ch chan os.Signal) {
	if runner.err != nil {
		ch <- syscall.SIGINT
		return
	}

	runner.leaderElector.Run(ctx)
}
