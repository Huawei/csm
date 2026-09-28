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

	"github.com/huawei/csm/v2/utils/log"
)

// Ensure ObjectCollector implements CollectorInterface
var _ CollectorInterface = (*ObjectCollector)(nil)

// ObjectCollector collects object data from storage backends
type ObjectCollector struct {
	clientCache  *ClientCache
	discoverFunc func(context.Context, string) (ClientInfo, error)
}

// NewObjectCollector creates a new ObjectCollector
func NewObjectCollector(
	clientCache *ClientCache,
	discoverFunc func(context.Context, string) (ClientInfo, error),
) *ObjectCollector {
	return &ObjectCollector{
		clientCache:  clientCache,
		discoverFunc: discoverFunc,
	}
}

// Collect finds a handler and invokes it for object data collection
func (o *ObjectCollector) Collect(ctx context.Context, req *CollectRequest) (*CollectResponse, error) {
	clientInfo, err := o.clientCache.DiscoverClient(ctx, req.BackendName, o.discoverFunc)
	if err != nil {
		log.AddContext(ctx).Errorf("objectCollector get client failed, error: [%v]", err)
		return nil, err
	}

	handler, err := GetObjectHandler(clientInfo.StorageType, req.CollectType)
	if err != nil {
		log.AddContext(ctx).Warningf("objectCollector get handler function failed, error: [%v]", err)
		return nil, err
	}

	return handler(ctx, clientInfo.Client, req)
}
