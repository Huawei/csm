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

// Package cmicore provides a shared library for storage backend operations
package cmicore

import (
	"context"
	"fmt"

	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
	"github.com/huawei/csm/v2/storage/client/distributedstorage"
	"github.com/huawei/csm/v2/storage/constant"
	"github.com/huawei/csm/v2/utils/log"
)

// volumeTypes maps storage field of the backend to volume type
var volumeTypes = map[string]string{
	constants.StorageNas:       constants.NasVolume,
	constants.StorageSan:       constants.LunVolume,
	constants.StorageFusionNas: constants.NasVolume,
}

// ClientInfoBuilder client builder
type ClientInfoBuilder struct {
	ctx        context.Context
	err        error
	clientInfo *ClientInfo
}

// NewClientInfoBuilder init an instance of ClientInfoBuilder
func NewClientInfoBuilder(ctx context.Context) *ClientInfoBuilder {
	return &ClientInfoBuilder{ctx: ctx, clientInfo: &ClientInfo{}}
}

// Build init an instance of ClientInfo
func (b *ClientInfoBuilder) Build() (ClientInfo, error) {
	return *b.clientInfo, b.err
}

// WithVolumeType build with volume type
func (b *ClientInfoBuilder) WithVolumeType(storageType string) *ClientInfoBuilder {
	if b.err != nil {
		return b
	}

	volumeType, ok := volumeTypes[storageType]
	if ok {
		b.clientInfo.VolumeType = volumeType
	} else {
		b.err = fmt.Errorf("illegalArgumentError unsupported storage type %s", storageType)
	}

	return b
}

// WithClient build with client, routes to centralized or distributed client based on storage type.
func (b *ClientInfoBuilder) WithClient(config *constant.StorageBackendConfig) *ClientInfoBuilder {
	if b.err != nil {
		return b
	}

	switch config.StorageType {
	case constants.StorageFusionNas:
		client, err := distributedstorage.NewDistributedClient(b.ctx, config)
		if err != nil {
			log.AddContext(b.ctx).Errorf("init distributed client failed, backendName: %s, error: %v",
				config.StorageBackendName, err)
			b.err = err
			return b
		}

		if err := client.Login(b.ctx); err != nil {
			log.AddContext(b.ctx).Errorf("login distributed storage failed, backendName: %s, error: %v",
				config.StorageBackendName, err)
			b.err = err
			return b
		}

		b.clientInfo.StorageName = config.StorageBackendName
		b.clientInfo.StorageType = constants.FusionStorage
		b.clientInfo.Client = client
	case constants.StorageNas, constants.StorageSan:
		client, err := centralizedstorage.NewCentralizedClient(b.ctx, config)
		if err != nil {
			log.AddContext(b.ctx).Errorf("init centralized client failed, backendName: %s, error: %v",
				config.StorageBackendName, err)
			b.err = err
			return b
		}

		if err := client.Login(b.ctx); err != nil {
			log.AddContext(b.ctx).Errorf("login storage failed, backendName: %s, error: %v",
				config.StorageBackendName, err)
			b.err = err
			return b
		}

		b.clientInfo.StorageName = config.StorageBackendName
		b.clientInfo.StorageType = constants.OceanStorage
		b.clientInfo.Client = client
	default:
		b.err = fmt.Errorf("illegalArgumentError unsupported storage type %s", config.StorageType)
	}

	return b
}
