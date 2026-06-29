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

// Package client
package client

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
)

func Test_option_GetName(t *testing.T) {
	o := &option{}
	if got := o.GetName(); got != clientOptionName {
		t.Errorf("GetName() = %v, want %v", got, clientOptionName)
	}
}

func Test_option_ValidateConfig_NegativeClientBurst(t *testing.T) {
	o := &option{
		kubeConfig:   t.TempDir() + "/config",
		kubeAPIBurst: -1,
	}
	if err := o.ValidateConfig(); err == nil {
		t.Errorf("ValidateConfig() expected error for negative clientBurst")
	}
}

func Test_option_ValidateConfig_NegativeClientQPS(t *testing.T) {
	o := &option{
		kubeConfig: t.TempDir() + "/config",
		kubeAPIQPS: -1.0,
	}
	if err := o.ValidateConfig(); err == nil {
		t.Errorf("ValidateConfig() expected error for negative clientQPS")
	}
}

func Test_option_ValidateConfig_BurstLessThanQPS(t *testing.T) {
	o := &option{
		kubeConfig:   t.TempDir() + "/config",
		kubeAPIQPS:   10.0,
		kubeAPIBurst: 5,
	}
	if err := o.ValidateConfig(); err == nil {
		t.Errorf("ValidateConfig() expected error for burst < QPS")
	}
}

func Test_option_ValidateConfig_BurstLessThanQPS_Float(t *testing.T) {
	o := &option{
		kubeConfig:   t.TempDir() + "/config",
		kubeAPIQPS:   5.1,
		kubeAPIBurst: 5,
	}
	if err := o.ValidateConfig(); err == nil {
		t.Errorf("ValidateConfig() expected error for burst(5) < QPS(5.1)")
	}
}

func TestGetKubeConfig(t *testing.T) {
	orig := Option.kubeConfig
	Option.kubeConfig = "/test/path"
	defer func() { Option.kubeConfig = orig }()

	if got := GetKubeConfig(); got != "/test/path" {
		t.Errorf("GetKubeConfig() = %v, want /test/path", got)
	}
}

func TestGetKubeAPIQPS(t *testing.T) {
	orig := Option.kubeAPIQPS
	Option.kubeAPIQPS = 15.5
	defer func() { Option.kubeAPIQPS = orig }()

	if got := GetKubeAPIQPS(); got != 15.5 {
		t.Errorf("GetKubeAPIQPS() = %v, want 15.5", got)
	}
}

func TestGetKubeAPIBurst(t *testing.T) {
	orig := Option.kubeAPIBurst
	Option.kubeAPIBurst = 20
	defer func() { Option.kubeAPIBurst = orig }()

	if got := GetKubeAPIBurst(); got != 20 {
		t.Errorf("GetKubeAPIBurst() = %v, want 20", got)
	}
}

func TestApplyKubeAPIQPSBurst_NilConfig(t *testing.T) {
	ApplyKubeAPIQPSBurst(nil)
}

func TestApplyKubeAPIQPSBurst_SetsQPSAndBurst(t *testing.T) {
	origQPS := Option.kubeAPIQPS
	origBurst := Option.kubeAPIBurst
	Option.kubeAPIQPS = 20.0
	Option.kubeAPIBurst = 40
	defer func() {
		Option.kubeAPIQPS = origQPS
		Option.kubeAPIBurst = origBurst
	}()

	cfg := &rest.Config{}
	ApplyKubeAPIQPSBurst(cfg)

	if cfg.QPS != 20.0 {
		t.Errorf("ApplyKubeAPIQPSBurst() QPS = %v, want 20.0", cfg.QPS)
	}
	if cfg.Burst != 40 {
		t.Errorf("ApplyKubeAPIQPSBurst() Burst = %v, want 40", cfg.Burst)
	}
}

func TestApplyKubeAPIQPSBurst_ZeroQPSDoesNotOverride(t *testing.T) {
	origQPS := Option.kubeAPIQPS
	origBurst := Option.kubeAPIBurst
	Option.kubeAPIQPS = 0
	Option.kubeAPIBurst = 0
	defer func() {
		Option.kubeAPIQPS = origQPS
		Option.kubeAPIBurst = origBurst
	}()

	cfg := &rest.Config{QPS: 99.0, Burst: 99}
	ApplyKubeAPIQPSBurst(cfg)

	if cfg.QPS != 99.0 {
		t.Errorf("ApplyKubeAPIQPSBurst() QPS = %v, want 99.0 (should not override)", cfg.QPS)
	}
	if cfg.Burst != 99 {
		t.Errorf("ApplyKubeAPIQPSBurst() Burst = %v, want 99 (should not override)", cfg.Burst)
	}
}

func Test_option_AddFlags(t *testing.T) {
	o := &option{}
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	o.AddFlags(fs)

	if f := fs.Lookup("kube-config"); f == nil {
		t.Error("AddFlags() missing kube-config flag")
	}
	if f := fs.Lookup("kube-api-qps"); f == nil {
		t.Error("AddFlags() missing kube-api-qps flag")
	}
	if f := fs.Lookup("kube-api-burst"); f == nil {
		t.Error("AddFlags() missing kube-api-burst flag")
	}
}

func Test_option_ValidateConfig_Success(t *testing.T) {
	// arrange
	o := &option{
		kubeConfig:   t.TempDir() + "/config",
		kubeAPIQPS:   5.0,
		kubeAPIBurst: 10,
	}

	// act
	err := o.ValidateConfig()

	// assert
	if err != nil {
		t.Errorf("Test_option_ValidateConfig_Success failed: [%v]", err)
	}
}

func Test_option_ValidateConfig_BurstEqualToQPS(t *testing.T) {
	o := &option{
		kubeConfig:   t.TempDir() + "/config",
		kubeAPIQPS:   5.0,
		kubeAPIBurst: 5,
	}
	if err := o.ValidateConfig(); err == nil {
		t.Errorf("ValidateConfig() expected error for burst == QPS")
	}
}

func Test_option_ValidateConfig_ZeroQPSAndBurst(t *testing.T) {
	o := &option{
		kubeConfig:   t.TempDir() + "/config",
		kubeAPIQPS:   0,
		kubeAPIBurst: 0,
	}
	if err := o.ValidateConfig(); err != nil {
		t.Errorf("ValidateConfig() expected no error for QPS=0, Burst=0, got: %v", err)
	}
}

func Test_option_ValidateConfig_ZeroQPSPositiveBurst(t *testing.T) {
	o := &option{
		kubeConfig:   t.TempDir() + "/config",
		kubeAPIQPS:   0,
		kubeAPIBurst: 10,
	}
	if err := o.ValidateConfig(); err != nil {
		t.Errorf("ValidateConfig() expected no error for QPS=0, Burst=10, got: %v", err)
	}
}

func Test_option_ValidateConfig_BurstLessThanQPS_FloatBoundary(t *testing.T) {
	o := &option{
		kubeConfig:   t.TempDir() + "/config",
		kubeAPIQPS:   5.0,
		kubeAPIBurst: 4,
	}
	if err := o.ValidateConfig(); err == nil {
		t.Errorf("ValidateConfig() expected error for burst(4) < QPS(5.0)")
	}
}

func Test_option_ValidateConfig_Fail(t *testing.T) {
	// arrange
	o := &option{
		kubeConfig: "fakeConfig",
	}

	// act
	got := o.ValidateConfig()

	// assert
	if got == nil {
		t.Errorf("Test_option_ValidateConfig_Fail: want error, got nil")
	} else if !strings.Contains(got.Error(), "invalid kubeConfig path") {
		t.Errorf("Test_option_ValidateConfig_Fail: want error containing [invalid kubeConfig path], got [%v]", got)
	}
}
