/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

// Package cmi defines config of cmi service
package cmi

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestNewProviderOption_Success(t *testing.T) {
	// action
	gotOption := NewProviderOption()

	// assert
	assert.NotNil(t, gotOption)
	assert.Equal(t, defaultQueryStoragePageSize, gotOption.queryStoragePageSize)
	assert.Equal(t, defaultClientMaxThreads, gotOption.clientMaxThreads)
}

func TestProviderOption_GetName_Success(t *testing.T) {
	// arrange
	option := NewProviderOption()

	// action
	gotName := option.GetName()

	// assert
	assert.Equal(t, defaultProviderOptionName, gotName)
}

func TestProviderOption_AddFlags_Success(t *testing.T) {
	// arrange
	option := NewProviderOption()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

	// action
	option.AddFlags(fs)

	// assert
	pageSize, _ := fs.GetInt("page-size")
	maxThreads, _ := fs.GetInt("client-max-threads")
	assert.Equal(t, defaultQueryStoragePageSize, pageSize)
	assert.Equal(t, defaultClientMaxThreads, maxThreads)
}

func TestProviderOption_AddFlags_CustomValues(t *testing.T) {
	// arrange
	option := NewProviderOption()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	option.AddFlags(fs)

	// action
	err := fs.Parse([]string{"--page-size=200", "--client-max-threads=50"})

	// assert
	assert.NoError(t, err)
	assert.Equal(t, 200, option.queryStoragePageSize)
	assert.Equal(t, 50, option.clientMaxThreads)
}

func TestProviderOption_ValidateConfig_Success(t *testing.T) {
	// arrange
	option := NewProviderOption()

	// action
	gotErr := option.ValidateConfig()

	// assert
	assert.NoError(t, gotErr)
}

func TestGetQueryStoragePageSize_Success(t *testing.T) {
	// arrange
	Option = NewProviderOption()

	// action
	gotPageSize := GetQueryStoragePageSize()

	// assert
	assert.Equal(t, defaultQueryStoragePageSize, gotPageSize)
}

func TestGetClientMaxThreads_Success(t *testing.T) {
	// arrange
	Option = NewProviderOption()

	// action
	gotMaxThreads := GetClientMaxThreads()

	// assert
	assert.Equal(t, defaultClientMaxThreads, gotMaxThreads)
}
