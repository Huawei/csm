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

// Package collectdef provides declarative Prometheus Collector definitions driven by ObjectDef.
package collectdef

import (
	"context"
	"fmt"

	"github.com/huawei/csm/v2/provider/cmicore"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/provider/utils"
	"github.com/huawei/csm/v2/server/prometheus-exporter/collector"
	"github.com/huawei/csm/v2/server/prometheus-exporter/exporterhandler"
	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
	"github.com/huawei/csm/v2/storage/client/distributedstorage"
	"github.com/huawei/csm/v2/utils/log"
)

const DefaultIDKey = "ID"

// MustRegister registers all subsystems for a storage object type in one call.
// It wires up the object handler, performance handler, indicator mapping,
// Prometheus collector factory, metrics data factory, and HTTP whitelist entry.
// If CollectMode is CollectModeNone, ObjectHandler and PerformanceHandler
// generation is skipped.
// Returns an error if the ObjectDef configuration is invalid (e.g., missing
// required labels, unsupported CollectMode).
func MustRegister(storageType string, def *ObjectDef) {
	if def.CollectMode != CollectModeNone && def.SupportObject {
		handler, err := generateObjectHandler(def)
		if err != nil {
			log.Errorf("generate object handler for %s failed: %v", def.CollectType, err)
			return
		}
		cmicore.RegisterObjectHandlerDirect(storageType, def.CollectType, handler)
	}

	if def.CollectMode != CollectModeNone && def.SupportPerformance && def.Perf != nil {
		perfHandler, err := generatePerformanceHandler(def)
		if err != nil {
			log.Errorf("generate performance handler for %s failed: %v", def.CollectType, err)
			return
		}
		cmicore.RegisterPerformanceHandlerDirect(storageType, def.CollectType, perfHandler)
		cmicore.RegisterIndicatorMapping(def.CollectType, def.Perf.TypeID)

		// Register object type mapping for distributed storage performance queries
		if storageType == constants.FusionStorage {
			distributedstorage.RegisterObjectType(def.Perf.TypeID, def.UrlKeys.ListKey,
				def.UrlKeys.CountKey, def.UrlKeys.PageKey, DefaultIDKey)
		}
	}

	collector.RegisterCollector(def.CollectType, NewDefCollectorFactory(def))

	registerMetricsDataFactory(def)

	exporterhandler.AddMetricsObjectLegal(def.CollectType)
}

// generateObjectHandler creates an ObjectHandler from the ObjectDef's CollectMode and UrlKeys.
// It uses the RestClient interface for data retrieval, making it compatible with
// any storage client that implements RestClient.
// Returns an error if the CollectMode is not supported.
func generateObjectHandler(def *ObjectDef) (cmicore.ObjectHandler, error) {
	switch def.CollectMode {
	case CollectModeSingle:
		return func(ctx context.Context, client cmicore.RestClient, req *cmicore.CollectRequest) (*cmicore.CollectResponse, error) {
			data, err := client.GetSingleByUrlKey(ctx, def.UrlKeys.SingleKey)
			if err != nil {
				return nil, fmt.Errorf("GetSingleByUrlKey(%s) failed: %w", def.UrlKeys.SingleKey, err)
			}
			return cmicore.ConvertMapToResponse([]map[string]interface{}{data}, req), nil
		}, nil

	case CollectModeList:
		return func(ctx context.Context, client cmicore.RestClient, req *cmicore.CollectRequest) (*cmicore.CollectResponse, error) {
			data, err := client.GetListByUrlKey(ctx, def.UrlKeys.ListKey)
			if err != nil {
				return nil, fmt.Errorf("GetListByUrlKey(%s) failed: %w", def.UrlKeys.ListKey, err)
			}
			return cmicore.ConvertMapToResponse(data, req), nil
		}, nil

	case CollectModePaginated:
		return func(ctx context.Context, client cmicore.RestClient, req *cmicore.CollectRequest) (*cmicore.CollectResponse, error) {
			countFunc := func(ctx context.Context) (int, error) {
				return client.GetCountByUrlKey(ctx, def.UrlKeys.CountKey)
			}
			pageFunc := func(ctx context.Context, start, end int) ([]map[string]interface{}, error) {
				return client.GetPageByUrlKey(ctx, def.UrlKeys.PageKey, start, end)
			}
			data, err := cmicore.ConcurrentPaginate(ctx, countFunc, pageFunc)
			if err != nil {
				return nil, fmt.Errorf("ConcurrentPaginate for %s failed: %w", def.CollectType, err)
			}
			return cmicore.ConvertMapToResponse(data, req), nil
		}, nil

	default:
		return nil, fmt.Errorf("unsupported CollectMode %v for %s", def.CollectMode, def.CollectType)
	}
}

// generatePerformanceHandler creates a PerformanceHandler from the ObjectDef.
// The handler performs the full performance collection pipeline:
// get raw performance data → get object name mapping → merge → return CollectResponse.
// Returns an error if the ObjectDef is missing required ID or NAME labels.
func generatePerformanceHandler(def *ObjectDef) (cmicore.PerformanceHandler, error) {
	idKey, nameKey, err := findIdNameKeys(def)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, client cmicore.RestClient, req *cmicore.CollectRequest) (*cmicore.CollectResponse, error) {
		mapData, err := client.QueryPerformanceData(ctx, def.Perf.TypeID, req.Indicators)
		if err != nil {
			return nil, fmt.Errorf("query performance data for %s failed: %w", def.CollectType, err)
		}

		performances, err := utils.MapToStruct[[]map[string]interface{}, []cmicore.PerformanceIndicators](mapData)
		if err != nil {
			return nil, fmt.Errorf("convert performance data for %s failed: %w", def.CollectType, err)
		}

		if len(performances) == 0 {
			return cmicore.BuildResponse(req), nil
		}

		nameMapping, err := getObjectNameMapping(ctx, client, def, idKey, nameKey)
		if err != nil {
			return nil, fmt.Errorf("get name mapping for %s failed: %w", def.CollectType, err)
		}

		return cmicore.MergePerformance(performances, nameMapping, req), nil
	}, nil
}

// getObjectNameMapping queries object data and builds an id-to-name mapping.
func getObjectNameMapping(ctx context.Context, client cmicore.RestClient,
	def *ObjectDef, idKey, nameKey string) (map[string]string, error) {

	switch def.CollectMode {
	case CollectModeList:
		data, err := client.GetListByUrlKey(ctx, def.UrlKeys.ListKey)
		if err != nil {
			return nil, err
		}
		return buildNameMapping(data, idKey, nameKey), nil

	case CollectModePaginated:
		countFunc := func(ctx context.Context) (int, error) {
			return client.GetCountByUrlKey(ctx, def.UrlKeys.CountKey)
		}
		pageFunc := func(ctx context.Context, start, end int) ([]map[string]interface{}, error) {
			return client.GetPageByUrlKey(ctx, def.UrlKeys.PageKey, start, end)
		}
		data, err := cmicore.ConcurrentPaginate(ctx, countFunc, pageFunc)
		if err != nil {
			return nil, err
		}
		return buildNameMapping(data, idKey, nameKey), nil

	default:
		return nil, fmt.Errorf("unsupported CollectMode %v for performance name mapping of %s", def.CollectMode, def.CollectType)
	}
}

// findIdNameKeys extracts the ID and NAME source keys from ObjectLabels.
// Returns an error if either key is not found.
func findIdNameKeys(def *ObjectDef) (idKey, nameKey string, err error) {
	for _, ld := range def.ObjectLabels {
		if ld.Name == labelIdKey {
			idKey = ld.SourceKey
		}
		if ld.Name == labelNameKey {
			nameKey = ld.SourceKey
		}
	}
	if idKey == "" {
		return "", "", fmt.Errorf("ObjectDef %s: no ID label found in ObjectLabels", def.CollectType)
	}
	if nameKey == "" {
		return "", "", fmt.Errorf("ObjectDef %s: no NAME label found in ObjectLabels", def.CollectType)
	}
	return
}

// buildNameMapping creates a map[id]name from raw data using the given keys.
func buildNameMapping(data []map[string]interface{}, idKey, nameKey string) map[string]string {
	result := make(map[string]string, len(data))
	for _, item := range data {
		id := fmt.Sprintf("%v", item[idKey])
		name := fmt.Sprintf("%v", item[nameKey])
		result[id] = name
	}
	return result
}

// registerMetricsDataFactory registers a MetricsData factory for the collect type.
// If ObjectDef.MetricsDataFactory is set, it is used; otherwise the default
// StorageMetricsData factory is registered.
func registerMetricsDataFactory(def *ObjectDef) {
	if def.MetricsDataFactory != nil {
		customFactory := def.MetricsDataFactory
		metricsCache.RegisterMetricsData(def.CollectType, func(backendName, metricsType string) (metricsCache.MetricsData, error) {
			md, err := customFactory(backendName, metricsType)
			if err != nil {
				return nil, err
			}
			data, ok := md.(metricsCache.MetricsData)
			if !ok {
				return nil, fmt.Errorf("MetricsDataFactory for %s did not return MetricsData", def.CollectType)
			}
			return data, nil
		})
		return
	}
	// Default: StorageMetricsData factory
	metricsCache.RegisterMetricsData(def.CollectType, metricsCache.NewStorageMetricsData)
}
