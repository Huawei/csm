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

// Package collector includes all huawei storage collectors to gather and export huawei storage metrics.
package collector

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
	"github.com/huawei/csm/v2/utils/log"
)

const MetricsNamespace = "huawei_storage"

// a collector constructor
type collectorInitFunc = func(backendName, monitorType string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error)

// factories are routing table with collector factory routing
// key is collector name
// value is a collector constructor
// e.g.
// |---------------|--------------------|
// | collectorName | collectorInitFunc  |
// |---------------|--------------------|
// | array         | NewArrayCollector  |
// |---------------|--------------------|
var factories = make(map[string]collectorInitFunc)

// RegisterCollector register a collector constructor to factories
func RegisterCollector(collectorName string,
	factory collectorInitFunc) {
	factories[collectorName] = factory
}

// BaseCollector implements the prometheus.Collector interface.
type BaseCollector struct {
	backendName      string
	monitorType      string
	collectorName    string
	metricsHelpMap   map[string]string
	metricsLabelMap  map[string][]string
	labelParseMap    map[string]parseRelation
	metricsParseMap  map[string]parseRelation
	metricsDataCache *metricsCache.MetricsDataCache
	metrics          map[string]*prometheus.Desc
}

// Describe implements the prometheus.Collector interface.
// Use BuildDesc to build prometheus.Desc then send to prometheus.
func (baseCollector *BaseCollector) Describe(ch chan<- *prometheus.Desc) {
	baseCollector.BuildDesc()
	for _, i := range baseCollector.metrics {
		ch <- i
	}
}

// NewPerformanceBaseCollector build a performance BaseCollector to other collector
func NewPerformanceBaseCollector(backendName, monitorType, collectorName string, metricsIndicators []string,
	metricsDataCache *metricsCache.MetricsDataCache) (*BaseCollector, error) {
	if len(metricsIndicators) == 0 || metricsIndicators[0] == "" {
		return nil, fmt.Errorf("can not create [%s] collector, "+
			"the metricsIndicators is empty or error", collectorName)
	}
	metricsData := strings.Split(metricsIndicators[0], ",")
	return (&BaseCollector{}).SetBackendName(backendName).
		SetMonitorType(monitorType).
		SetCollectorName(collectorName).
		SetMetricsHelpMap(pickPerformanceParsMap[string](metricsData, performanceMetricsHelpMap)).
		SetMetricsLabelMap(pickPerformanceParsMap[[]string](metricsData, performanceMetricsLabelMap)).
		SetLabelParseMap(performanceLabelParseMap).
		SetMetricsParseMap(pickPerformanceParsMap[parseRelation](metricsData, performanceMetricsParseMap)).
		SetMetricsDataCache(metricsDataCache).
		SetMetrics(make(map[string]*prometheus.Desc)), nil
}

func (baseCollector *BaseCollector) setPrometheusMetric(ctx context.Context, ch chan<- prometheus.Metric,
	metricsName string, detailData map[string]string) {
	metricsParseRelation, ok := baseCollector.metricsParseMap[metricsName]
	if !ok {
		log.AddContext(ctx).Warningln("can not get the metricsParseRelation")
		return
	}
	// parse metricsValue
	metricsValue := metricsParseRelation.parseFunc(
		metricsParseRelation.parseKey, metricsName, detailData)
	metricsValueFloat, err := strconv.ParseFloat(metricsValue, bitSize)
	if err != nil {
		log.AddContext(ctx).Debugf("can not get the metricsValueFloat the metricsName is [%v]", metricsName)
		return
	}
	// parse metricsLabel, from label key get label value
	labelKeys, ok := baseCollector.metricsLabelMap[metricsName]
	if !ok {
		log.AddContext(ctx).Warningln("can not get the labelKeys")
		return
	}
	labelValueSlice := parseLabelListToLabelValueSlice(
		labelKeys, baseCollector.labelParseMap, metricsName, detailData)
	if len(labelValueSlice) != len(labelKeys) {
		log.AddContext(ctx).Warningln("can not get the labelValueSlice")
		return
	}
	ch <- prometheus.MustNewConstMetric(
		baseCollector.metrics[metricsName],
		prometheus.GaugeValue,
		metricsValueFloat,
		labelValueSlice...,
	)
}

// Collect implements the prometheus.Collector interface.
// Parse the data cached in MetricsDataCache and generate the return required by Prometheus.
// Use metricsParseMap to obtain the metric data parseRelation to parse the metric.
// Use metricsLabelMap to obtain the metric label parseRelation to parse tag information.
func (baseCollector *BaseCollector) Collect(ch chan<- prometheus.Metric) {
	collectorCacheData := baseCollector.metricsDataCache.GetMetricsData(baseCollector.collectorName)
	ctx := context.Background()
	if collectorCacheData == nil || len(collectorCacheData.Details) == 0 {
		log.AddContext(ctx).Warningln("can not get the collectorCacheData")
		return
	}

	for _, storageCollectDetail := range collectorCacheData.Details {
		detailData := storageCollectDetail.Data
		if len(detailData) == 0 {
			log.AddContext(ctx).Warningln("can not get the detailData")
			continue
		}

		// Set backendName and collectorName to one cacheData, used by parseRelation.parseFunc
		detailData["backendName"] = collectorCacheData.BackendName
		detailData["collectorName"] = collectorCacheData.CollectType
		for metricsName := range baseCollector.metrics {
			baseCollector.setPrometheusMetric(ctx, ch, metricsName, detailData)
		}
	}
}

// BuildDesc use BaseCollector.metricsDescMap create different Collector prometheus.Desc
func (baseCollector *BaseCollector) BuildDesc() {
	if baseCollector.metrics == nil {
		baseCollector.metrics = make(map[string]*prometheus.Desc)
	}
	for metricsName, helpInfo := range baseCollector.metricsHelpMap {
		baseCollector.metrics[metricsName] =
			prometheus.NewDesc(
				prometheus.BuildFQName(
					MetricsNamespace, baseCollector.collectorName, metricsName),
				helpInfo,
				baseCollector.metricsLabelMap[metricsName],
				nil)
	}
}

// SetBackendName set backendName
func (baseCollector *BaseCollector) SetBackendName(backendName string) *BaseCollector {
	baseCollector.backendName = backendName
	return baseCollector
}

// SetMonitorType set monitorType
func (baseCollector *BaseCollector) SetMonitorType(monitorType string) *BaseCollector {
	baseCollector.monitorType = monitorType
	return baseCollector
}

// SetCollectorName set collectorName
func (baseCollector *BaseCollector) SetCollectorName(collectorName string) *BaseCollector {
	baseCollector.collectorName = collectorName
	return baseCollector
}

// SetMetricsHelpMap set metricsHelpMap
func (baseCollector *BaseCollector) SetMetricsHelpMap(metricsHelpMap map[string]string) *BaseCollector {
	baseCollector.metricsHelpMap = metricsHelpMap
	return baseCollector
}

// SetMetricsLabelMap set metricsLabelMap
func (baseCollector *BaseCollector) SetMetricsLabelMap(metricsLabelMap map[string][]string) *BaseCollector {
	baseCollector.metricsLabelMap = metricsLabelMap
	return baseCollector
}

// SetLabelParseMap set labelParseMap
func (baseCollector *BaseCollector) SetLabelParseMap(labelParseMap map[string]parseRelation) *BaseCollector {
	baseCollector.labelParseMap = labelParseMap
	return baseCollector
}

// SetMetricsParseMap set metricsParseMap
func (baseCollector *BaseCollector) SetMetricsParseMap(metricsParseMap map[string]parseRelation) *BaseCollector {
	baseCollector.metricsParseMap = metricsParseMap
	return baseCollector
}

// SetMetricsDataCache set metricsDataCache
func (baseCollector *BaseCollector) SetMetricsDataCache(
	metricsDataCache *metricsCache.MetricsDataCache) *BaseCollector {
	baseCollector.metricsDataCache = metricsDataCache
	return baseCollector
}

// SetMetrics set metrics
func (baseCollector *BaseCollector) SetMetrics(metrics map[string]*prometheus.Desc) *BaseCollector {
	baseCollector.metrics = metrics
	return baseCollector
}

// CollectorSet implements the prometheus.Collector interface.
// Save Multi BaseCollector
type CollectorSet struct {
	collectors []prometheus.Collector
}

// NewCollectorSet create all objects that need to be collected in this batch
func NewCollectorSet(ctx context.Context, params map[string][]string, backendName, monitorType string,
	metricsDataCache *metricsCache.MetricsDataCache) (*CollectorSet, error) {
	var collectors []prometheus.Collector

	for collectorName, metricsIndicators := range params {
		collectorFunc, ok := factories[collectorName]
		if !ok {
			log.AddContext(ctx).Errorf("New collector error, the factories not have %s", collectorName)
			continue
		}
		collector, err := collectorFunc(backendName, monitorType, metricsIndicators, metricsDataCache)
		if err != nil {
			log.AddContext(ctx).Errorf("New collector for %s, the monitorType : %s, error: %v",
				collectorName, monitorType, err)
			continue
		}
		collectors = append(collectors, collector)
	}

	if len(collectors) == 0 {
		return nil, fmt.Errorf("can not get the collector")
	}

	return &CollectorSet{
		collectors: collectors,
	}, nil
}

func (collectorSet *CollectorSet) Describe(ch chan<- *prometheus.Desc) {
	for _, collector := range collectorSet.collectors {
		collector.Describe(ch)
	}
}

func (collectorSet *CollectorSet) Collect(ch chan<- prometheus.Metric) {
	for _, collector := range collectorSet.collectors {
		collector.Collect(ch)
	}
}
