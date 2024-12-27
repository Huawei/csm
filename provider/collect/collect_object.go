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

// Package collect is a package that provides object and performance collect
package collect

import (
	"context"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/backend"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
	"github.com/huawei/csm/v2/utils/log"
)

// This function will register all objectHandlers
// If a handler is not registered here, an error will be reported when calling DoCollect
// These objectHandlers will be saved in the Global variable register
func init() {
	RegisterObjectHandler(constants.OceanStorage, constants.Lun, CollectLun)
	RegisterObjectHandler(constants.OceanStorage, constants.Array, CollectArray)
	RegisterObjectHandler(constants.OceanStorage, constants.Controller, CollectController)
	RegisterObjectHandler(constants.OceanStorage, constants.Filesystem, CollectFilesystem)
	RegisterObjectHandler(constants.OceanStorage, constants.StoragePool, CollectStoragePool)
}

// ObjectCollector object data collector
type ObjectCollector struct{}

// Collect this purpose of this function is to find a handler and invoke it
func (o *ObjectCollector) Collect(ctx context.Context, request *cmi.CollectRequest) (*cmi.CollectResponse, error) {
	clientInfo, err := GetClient(ctx, request.GetBackendName(), backend.GetClientByBackendName)
	if err != nil {
		log.AddContext(ctx).Errorf("objectCollector get client failed, error: [%v]", err)
		return nil, err
	}

	handler, err := GetObjectHandler(clientInfo.StorageType, request.GetCollectType())
	if err != nil {
		log.AddContext(ctx).Errorf("objectCollector get handler function failed, error: [%v]", err)
		return nil, err
	}

	return handler(ctx, clientInfo.Client, request)
}

// CollectArray collect object data of array in storage
func CollectArray(ctx context.Context, client *centralizedstorage.CentralizedClient,
	request *cmi.CollectRequest) (*cmi.CollectResponse, error) {
	return DoCollect[map[string]interface{}, ArrayObject](ctx, request, client.GetSystemInfo)
}

// CollectController collect object data of array in storage
func CollectController(ctx context.Context, client *centralizedstorage.CentralizedClient,
	request *cmi.CollectRequest) (*cmi.CollectResponse, error) {
	return DoCollect[[]map[string]interface{}, ControllerObject](ctx, request, client.GetControllers)
}

// CollectStoragePool collect object data of storage pool in storage
func CollectStoragePool(ctx context.Context, client *centralizedstorage.CentralizedClient,
	request *cmi.CollectRequest) (*cmi.CollectResponse, error) {
	return DoCollect[[]map[string]interface{}, StoragePoolObject](ctx, request, client.GetStoragePools)
}

// CollectLun collect object data of lun in storage
func CollectLun(ctx context.Context, client *centralizedstorage.CentralizedClient,
	request *cmi.CollectRequest) (*cmi.CollectResponse, error) {
	return DoPageCollect[LunObject](ctx, request, client.GetLunCount, client.GetLuns)
}

// CollectFilesystem collect object data of filesystem in storage
func CollectFilesystem(ctx context.Context, client *centralizedstorage.CentralizedClient,
	request *cmi.CollectRequest) (*cmi.CollectResponse, error) {
	return DoPageCollect[FileSystemObject](ctx, request, client.GetFilesystemCount, client.GetFilesystem)
}

// DoCollect collect data in storage
func DoCollect[I, T any](ctx context.Context, request *cmi.CollectRequest,
	fn func(context.Context) (I, error)) (*cmi.CollectResponse, error) {
	data, err := fn(ctx)
	if err != nil {
		log.AddContext(ctx).Errorf("do collect failed, error: %v", err)
		return nil, err
	}
	return ConvertToResponse[I, T](data, request)
}

// DoPageCollect page collect data in storage
func DoPageCollect[T any](ctx context.Context, request *cmi.CollectRequest,
	countFunc CountFunc, pageFunc PageFunc) (*cmi.CollectResponse, error) {
	data, err := ConcurrentPaginate(ctx, countFunc, pageFunc)
	if err != nil {
		log.AddContext(ctx).Errorf("do page collect failed, error: %v", err)
		return nil, err
	}
	return ConvertToResponse[[]map[string]interface{}, T](data, request)
}
