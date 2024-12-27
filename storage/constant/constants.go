/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

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

// Package constant is related with storage client constant
package constant

// StorageBackendConfig contains storage standard info
type StorageBackendConfig struct {
	StorageType string
	Urls        []string
	Pools       []string
	User        string

	SecretNamespace string
	SecretName      string

	StorageBackendNamespace string
	StorageBackendName      string

	ClientMaxThreads int
}

const (
	// CertificateKeyName refer to certificate config key name
	CertificateKeyName = "tls.crt"
)
