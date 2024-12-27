/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

// Package backend is a package that manager storage backend
package backend

import (
	"context"

	"github.com/huawei/csm/v2/utils/log"
)

// GetClientByBackendName get Client by backend name
func GetClientByBackendName(ctx context.Context, backendName string) (ClientInfo, error) {
	log.AddContext(ctx).Infof("Start to get client, name: %s", backendName)
	config, err := NewStorageBackendConfigBuilder(ctx, backendName).
		WithSbcInfo().
		WithSecretInfo().
		WithConfigMapInfo().Build()
	if err != nil {
		log.AddContext(ctx).Errorf("build storage config failed, name: %s, error: %v", backendName, err)
		return ClientInfo{}, err
	}

	return NewClientInfoBuilder(ctx).
		WithVolumeType(config.StorageType).WithClient(config).Build()
}
