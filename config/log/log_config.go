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

// Package log is used to init log configurations and flags
package log

import (
	"flag"

	"github.com/spf13/pflag"

	"github.com/huawei/csm/v2/config/consts"
)

const (
	logOptionName      = "LogOption"
	defaultLogFileName = "topo-service"
)

// Option is a log option instance for manager init
var Option = &option{}

type option struct {
	logFile string
}

// GetName return name string of log option
func (o *option) GetName() string {
	return logOptionName
}

// AddFlags is to add flags for log configurations
func (o *option) AddFlags(fs *pflag.FlagSet) {
	fs.AddGoFlagSet(flag.CommandLine)
	fs.StringVar(&o.logFile, consts.LogFile, defaultLogFileName,
		"The log file name of the resource topology service.")
}

// ValidateConfig is to validate input log configurations
func (o *option) ValidateConfig() error {
	return nil
}

// GetLogFile returns the log file name
func GetLogFile() string {
	return Option.logFile
}
