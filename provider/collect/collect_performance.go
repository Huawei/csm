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
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/provider/backend"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/provider/utils"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
	"github.com/huawei/csm/v2/utils/log"
)

func init() {
	RegisterPerformanceHandler(constants.OceanStorage, constants.Lun, GetLunNameMapping)
	RegisterPerformanceHandler(constants.OceanStorage, constants.Filesystem, GetFilesystemNameMapping)
	RegisterPerformanceHandler(constants.OceanStorage, constants.Controller, GetControllerNameMapping)
	RegisterPerformanceHandler(constants.OceanStorage, constants.StoragePool, GetStoragePoolNameMapping)
}

// PerformanceCollector performance data collector
type PerformanceCollector struct{}

// Collect performance data
func (p *PerformanceCollector) Collect(ctx context.Context, request *cmi.CollectRequest) (*cmi.CollectResponse, error) {
	clientInfo, err := GetClient(ctx, request.GetBackendName(), backend.GetClientByBackendName)
	if err != nil {
		log.AddContext(ctx).Errorf("objectCollector get Client failed, error: %v", err)
		return nil, err
	}

	client, ok := clientInfo.Client.(*centralizedstorage.CentralizedClient)
	if !ok {
		return nil, errors.New("convert Client to centralizedClient failed")
	}

	return CollectPerformance(ctx, client, request)
}

// CollectPerformance collect performance data
func CollectPerformance(ctx context.Context, client *centralizedstorage.CentralizedClient,
	request *cmi.CollectRequest) (*cmi.CollectResponse, error) {
	// get all performance data.
	performances, err := GetPerformanceData(ctx, client, request)
	if err != nil {
		log.AddContext(ctx).Errorf("collect performance data failed, error: %v", err)
		return nil, err
	}

	if len(performances) == 0 {
		return BuildResponse(request), nil
	}

	// get all objects id and name.
	nameMapping, err := GetMapping(ctx, request.GetCollectType(), client)
	if err != nil {
		log.AddContext(ctx).Errorf("collect object name mapping data failed, error: %v", err)
		return nil, err
	}

	// merge performance data and object name.
	return MergePerformance(performances, nameMapping, request), nil
}

// GetPerformanceData query performance data
func GetPerformanceData(ctx context.Context, client *centralizedstorage.CentralizedClient,
	request *cmi.CollectRequest) ([]PerformanceIndicators, error) {
	objectType, ok := IndicatorsMapping[request.CollectType]
	if !ok {
		return nil, errors.New("illegalArgumentErrorunsupported collect type")
	}

	storageInfo, err := client.GetSystemInfo(ctx)
	if err != nil {
		log.AddContext(ctx).Errorf("get storage system info failed, error: %v", err)
		return nil, err
	}

	var mapData []map[string]interface{}
	var postEnable bool
	indicators := utils.MapStringToInt(request.Indicators)
	version, ok := storageInfo["pointRelease"].(string)
	if !ok {
		// storage of V3 or V5 not has the pointRelease field
		postEnable = false
	} else if !strings.HasPrefix(version, constants.StorageV6PointReleasePrefix) {
		// only storage of V6 pointRelease is started with number,
		// storage with V7 or later version supports the Post request
		postEnable = true
	} else if version >= constants.MinVersionSupportPost {
		// 6.1.2 and later versions in V6 storage support the Post request
		postEnable = true
	}

	if postEnable {
		mapData, err = client.GetPerformanceByPost(ctx, objectType, indicators)
	} else {
		for i := 0; i < 5; i++ {
			mapData, err = client.GetPerformance(ctx, objectType, indicators)
			// For storage v6 earlier 6.1.2, if it can not return the performance data caused by concurrency,
			// both the mapData and err are nil. But in the same conditions for storage v3 or v5, the mapData
			// is nil while the err is not nil.
			if err != nil {
				break
			}
			if len(mapData) != 0 {
				break
			}
			time.Sleep(5 * time.Second)
		}
	}
	if err != nil {
		log.AddContext(ctx).Errorf("invoke the get performance method of storage client failed, error: %v", err)
		return nil, err
	}

	// For storage v6 earlier 6.1.2, the storage may return empty data even after 5 time retries.
	if len(mapData) == 0 {
		log.AddContext(ctx).Warningln("get empty data by the get performance method of storage client")
	}

	return utils.MapToStruct[[]map[string]interface{}, []PerformanceIndicators](mapData)
}

// GetMapping get object mapping
// result map key is object id.
//
//	result map value is object name.
func GetMapping(ctx context.Context, collectType string,
	client *centralizedstorage.CentralizedClient) (map[string]string, error) {
	handler, err := GetPerformanceHandler(constants.OceanStorage, collectType)
	if err != nil {
		log.AddContext(ctx).Errorf("get performance handler failed, error: %v", err)
		return nil, err
	}

	return handler(ctx, client)
}

// MergePerformance merge performance data
func MergePerformance(performances []PerformanceIndicators, nameMapping map[string]string,
	request *cmi.CollectRequest) *cmi.CollectResponse {

	response := BuildResponse(request)
	for _, performance := range performances {
		mapData := performance.ToMap()
		objectName, ok := nameMapping[performance.ObjectId]
		if !ok {
			continue
		}
		mapData[constants.ObjectName] = objectName
		mapData[constants.ObjectId] = performance.ObjectId
		AddCollectDetailWithMap(mapData, response)
	}

	return response
}

// ToMap Parse performance data and convert it into a map
func (p PerformanceIndicators) ToMap() map[string]string {
	if len(p.Indicators) == 0 || len(p.Indicators) != len(p.IndicatorValues) {
		return map[string]string{}
	}

	var dataMap = map[string]string{}
	for i, indicator := range p.Indicators {
		key := strconv.Itoa(indicator)
		dataMap[key] = strconv.FormatFloat(p.IndicatorValues[i], 'f', 4, 64)
	}
	return dataMap
}

// GetNameMapping A universal function for obtaining name mapping
func GetNameMapping(ctx context.Context, queryFunc QueryFunc) (map[string]string, error) {
	data, err := queryFunc(ctx)
	if err != nil {
		log.AddContext(ctx).Errorf("query storage to get name mapping failed, error: %v", err)
		return map[string]string{}, nil
	}
	return DoNameMapping(data), nil
}

// GetNameMappingWithPage A universal function for obtaining name mapping with page query
func GetNameMappingWithPage(ctx context.Context, countFunc CountFunc, pageFunc PageFunc) (map[string]string, error) {
	data, err := ConcurrentPaginate(ctx, countFunc, pageFunc)
	if err != nil {
		log.AddContext(ctx).Errorf("concurrent Paginate failed, error: %v", err)
		return nil, err
	}
	return DoNameMapping(data), nil
}

// DoNameMapping A universal function for parsing name mapping
func DoNameMapping(data []map[string]interface{}) map[string]string {
	var nameMapping = map[string]string{}
	for _, item := range data {
		id, ok := item["ID"].(string)
		if !ok {
			continue
		}
		name, ok := item["NAME"].(string)
		if !ok {
			continue
		}
		nameMapping[id] = name
	}
	return nameMapping
}

// GetLunNameMapping get lun name mapping
// Key is lun id.
// Value is lun name.
func GetLunNameMapping(ctx context.Context, client *centralizedstorage.CentralizedClient) (map[string]string, error) {
	return GetNameMappingWithPage(ctx, client.GetLunCount, client.GetLuns)
}

// GetFilesystemNameMapping get filesystem name mapping
// Key is filesystem id.
// Value is filesystem name.
func GetFilesystemNameMapping(ctx context.Context,
	client *centralizedstorage.CentralizedClient) (map[string]string, error) {
	return GetNameMappingWithPage(ctx, client.GetFilesystemCount, client.GetFilesystem)
}

// GetControllerNameMapping get controller name mapping
// Key is controller id.
// Value is controller name.
func GetControllerNameMapping(ctx context.Context,
	client *centralizedstorage.CentralizedClient) (map[string]string, error) {
	return GetNameMapping(ctx, client.GetControllers)
}

// GetStoragePoolNameMapping get storage pool name mapping
// Key is storage pool id.
// Value is storage pool name.
func GetStoragePoolNameMapping(ctx context.Context,
	client *centralizedstorage.CentralizedClient) (map[string]string, error) {
	return GetNameMapping(ctx, client.GetStoragePools)
}
