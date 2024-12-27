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

// Package config contains all configuration and flags parts for different services
package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// Option is to help load configuration
type Option interface {
	// GetName returns the name of the option
	GetName() string
	// AddFlags adds the flags to the option
	AddFlags(*pflag.FlagSet)
	// ValidateConfig validates the input configuration
	ValidateConfig() error
}

// Manager helps to manage the configuration of server
type Manager struct {
	options []Option
	flagSet *pflag.FlagSet
}

// NewOptionManager creates a new option manager of specified options using the specified flag set
func NewOptionManager(fs *pflag.FlagSet, options ...Option) *Manager {
	return &Manager{
		flagSet: fs,
		options: options,
	}
}

// AddFlags adds the topology service config needed flags to set
func (m *Manager) AddFlags() {
	for _, o := range m.options {
		logrus.Infof("loading config [%s]", o.GetName())
		o.AddFlags(m.flagSet)
	}
}

// ValidateConfig validate input config
func (m *Manager) ValidateConfig() error {
	for _, o := range m.options {
		logrus.Infof("validating config [%s]", o.GetName())
		err := o.ValidateConfig()
		if err != nil {
			return fmt.Errorf("validate config [%s] failed: [%v]", o.GetName(), err)
		}
	}

	return nil
}
