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

// Package exporterhandler provide all handler use by prometheus exporter
package exporterhandler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/huawei/csm/v2/server/prometheus-exporter/collector"
	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
	"github.com/huawei/csm/v2/utils/log"
)

// url path is ip/monitorType/monitorBackendName. So when use strings.Split, pathLen is 3.
const pathLen = 3

var (
	// Supported monitoring types
	monitorTypeLegal = map[string]struct{}{
		"object":      {},
		"performance": {},
	}
	// Supported monitoring types
	metricsObjectLegal = map[string]struct{}{
		"array":       {},
		"controller":  {},
		"storagepool": {},
		"lun":         {},
		"filesystem":  {},
		"pv":          {},
	}
)

func checkMetricsObject(ctx context.Context, params map[string][]string, monitorType string) error {
	if monitorType == "" {
		return fmt.Errorf("the monitorType is empty")
	}

	for collectorName, metricsIndicators := range params {
		if _, err := metricsObjectLegal[collectorName]; !err {
			log.AddContext(ctx).Errorln("the collectorName is error")
			return fmt.Errorf("the collectorName is error")
		}

		if monitorType == "performance" && (len(metricsIndicators) == 0 || metricsIndicators[0] == "") {
			log.AddContext(ctx).Errorf("can not get the [%s] performance indicators", collectorName)
			return fmt.Errorf("the metricsIndicators is error")
		}
	}

	return nil
}

func parseRequestPath(ctx context.Context, w http.ResponseWriter, r *http.Request) (string, string, error) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) != pathLen {
		http.Error(w, "URL is error.", http.StatusBadRequest)
		return "", "", errors.New("URL is error")
	}

	monitorType := path[1]
	if _, err := monitorTypeLegal[monitorType]; !err {
		http.Error(w, "MonitorType is error.", http.StatusBadRequest)
		return "", "", errors.New("MonitorType is error")
	}

	monitorBackendName := path[2]
	params := r.URL.Query()
	checkError := checkMetricsObject(ctx, params, monitorType)
	if checkError != nil {
		http.Error(w, "MetricsObjectType is error.", http.StatusBadRequest)
		return "", "", errors.New("MetricsObjectType is error")
	}
	return monitorBackendName, monitorType, nil
}

func getBatchData(ctx context.Context, monitorBackendName, monitorType string,
	params map[string][]string) *metricsCache.MetricsDataCache {
	batchMetricsDataCache := metricsCache.MetricsDataCache{
		BackendName:  monitorBackendName,
		CacheDataMap: make(map[string]metricsCache.MetricsData),
		MergeMetrics: make(map[string]metricsCache.MergeMetricsData)}
	log.AddContext(ctx).Infof("start to get batch monitor data, backend: %v, type: %v, params: %v",
		monitorBackendName, monitorType, params)
	batchParams, err := batchMetricsDataCache.BuildBatchDataClass(ctx, monitorType, params)
	if err != nil {
		return nil
	}
	batchMetricsDataCache.SetBatchDataFromSource(ctx, monitorType, batchParams)
	batchMetricsDataCache.MergeBatchData(ctx)
	log.AddContext(ctx).Infoln("get batch monitor data finished, start to collect data for prometheus")
	return &batchMetricsDataCache
}

// MetricsHandler get the parse request get batch data and build data to prometheus
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, err := log.SetRequestInfo(context.Background())
	if err != nil {
		return
	}
	monitorBackendName, monitorType, err := parseRequestPath(ctx, w, r)
	if err != nil {
		return
	}

	params := r.URL.Query()

	batchMetricsDataCache := getBatchData(ctx, monitorBackendName, monitorType, params)
	if batchMetricsDataCache == nil {
		return
	}

	allCollectors, err := collector.NewCollectorSet(
		ctx, params, monitorBackendName, monitorType, batchMetricsDataCache)
	if err != nil {
		http.Error(w, "get allCollectors is error.", http.StatusBadRequest)
		log.AddContext(ctx).Errorf("get allCollectors failed, the error is [%v]", err)
		return
	}
	registry := prometheus.NewRegistry()
	registry.MustRegister(allCollectors)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
