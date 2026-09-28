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
	"testing"

	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
)

func Test_option_AddFlags(t *testing.T) {
	o := &option{}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	o.AddFlags(fs)

	if f := fs.Lookup("leader-election-kube-api-qps"); f != nil {
		t.Error("AddFlags() should not register leader-election-kube-api-qps flag (now hardcoded)")
	}
	if f := fs.Lookup("leader-election-kube-api-burst"); f != nil {
		t.Error("AddFlags() should not register leader-election-kube-api-burst flag (now hardcoded)")
	}
}

func TestApplyLeaderElectionQPSBurst_NilConfig(t *testing.T) {
	ApplyLeaderElectionQPSBurst(nil)
}

func TestApplyLeaderElectionQPSBurst_SetsFixedQPSAndBurst(t *testing.T) {
	cfg := &rest.Config{}
	ApplyLeaderElectionQPSBurst(cfg)

	if cfg.QPS != float32(defaultLeaderElectionKubeAPIQPS) {
		t.Errorf("ApplyLeaderElectionQPSBurst() QPS = %v, want %v", cfg.QPS, defaultLeaderElectionKubeAPIQPS)
	}
	if cfg.Burst != defaultLeaderElectionKubeAPIBurst {
		t.Errorf("ApplyLeaderElectionQPSBurst() Burst = %v, want %v", cfg.Burst, defaultLeaderElectionKubeAPIBurst)
	}
}
