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
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/server/prometheus-exporter/collector"
	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
)

// SkipReportValue defines the value which should not be reported
const SkipReportValue = "skipReportValue"

// DefCollector is a generic Prometheus Collector driven by ObjectDef.
// It replaces the per-type collector structs and 4-map pattern with
// declarative metric and label definitions from ObjectDef.
type DefCollector struct {
	*collector.BaseCollector
	def               *ObjectDef
	monitorType       string
	metricsIndicators []string
}

// NewDefCollectorFactory returns a factory that creates DefCollector instances.
// The factory signature matches collector.collectorInitFunc so it can be
// registered via collector.RegisterCollector.
func NewDefCollectorFactory(def *ObjectDef) func(backendName, monitorType string,
	metricsIndicators []string, metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
	return func(backendName, monitorType string, metricsIndicators []string,
		metricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
		if monitorType == constants.Object && !def.SupportObject {
			return nil, fmt.Errorf("collectType %s does not support object monitoring", def.CollectType)
		}
		if monitorType == constants.Performance && !def.SupportPerformance {
			return nil, fmt.Errorf("collectType %s does not support performance monitoring", def.CollectType)
		}
		if monitorType == constants.Performance && def.Perf == nil {
			return nil, fmt.Errorf("collectType %s supports performance but Perf is nil", def.CollectType)
		}
		return &DefCollector{
			BaseCollector: (&collector.BaseCollector{}).
				SetBackendName(backendName).
				SetMonitorType(monitorType).
				SetCollectorName(def.CollectType).
				SetMetricsDataCache(metricsDataCache).
				SetMetrics(make(map[string]*prometheus.Desc)),
			def:               def,
			monitorType:       monitorType,
			metricsIndicators: metricsIndicators,
		}, nil
	}
}

// Describe implements prometheus.Collector.
func (dc *DefCollector) Describe(ch chan<- *prometheus.Desc) {
	dc.BuildDesc()
	for _, desc := range dc.GetMetrics() {
		ch <- desc
	}
}

// BuildDesc creates prometheus.Desc entries from ObjectDef.
func (dc *DefCollector) BuildDesc() {
	metrics := dc.GetMetrics()
	if metrics == nil {
		metrics = make(map[string]*prometheus.Desc)
		dc.SetMetrics(metrics)
	}
	subsystem := dc.def.PrometheusSubsystemName()

	switch dc.monitorType {
	case constants.Object:
		for _, md := range dc.def.ObjectMetrics {
			labels := md.Labels
			if labels == nil {
				labels = dc.def.ObjectLabels
			}
			labelNames := labelNamesFrom(labels)
			metrics[md.Name] = prometheus.NewDesc(
				prometheus.BuildFQName(collector.MetricsNamespace, subsystem, md.Name),
				md.Help,
				labelNames,
				nil,
			)
		}
	case constants.Performance:
		indicators := dc.filterIndicators(dc.def.Perf.Indicators, dc.metricsIndicators)
		for code, metricSuffix := range indicators {
			labelNames := labelNamesFrom(dc.def.PerfLabels)
			metrics[metricSuffix] = prometheus.NewDesc(
				prometheus.BuildFQName(collector.MetricsNamespace, subsystem, metricSuffix),
				fmt.Sprintf("Performance indicator %d", code),
				labelNames,
				nil,
			)
		}
	default:
		// Unrecognized monitorType; no metrics to register.
	}
}

// Collect implements prometheus.Collector.
func (dc *DefCollector) Collect(ch chan<- prometheus.Metric) {
	cacheData := dc.GetMetricsDataCache().GetMetricsData(dc.GetCollectorName())
	if cacheData == nil || len(cacheData.Details) == 0 {
		return
	}
	for _, detail := range cacheData.Details {
		data := detail.Data
		if len(data) == 0 {
			continue
		}
		data["backendName"] = cacheData.BackendName
		data["collectorName"] = cacheData.CollectType

		switch dc.monitorType {
		case constants.Object:
			dc.collectObjectMetrics(ch, data)
		case constants.Performance:
			dc.collectPerformanceMetrics(ch, data)
		default:
			// Unrecognized monitorType; nothing to collect.
		}
	}
}

// collectObjectMetrics emits Prometheus metrics for object monitoring mode.
func (dc *DefCollector) collectObjectMetrics(ch chan<- prometheus.Metric, data map[string]string) {
	metrics := dc.GetMetrics()
	for _, md := range dc.def.ObjectMetrics {
		valueStr := dc.resolveMetricValue(md, data)
		if valueStr == SkipReportValue || valueStr == "" {
			continue
		}
		valueFloat, err := strconv.ParseFloat(valueStr, constants.DefaultBitSize)
		// Invalid numeric values are silently skipped: Prometheus collectors
		// should never expose malformed metrics; logging would be too noisy
		// at per-scrape granularity.
		if err != nil {
			continue
		}
		labels := md.Labels
		if labels == nil {
			labels = dc.def.ObjectLabels
		}
		labelValues := resolveLabelValues(labels, data)
		desc := metrics[md.Name]
		if desc == nil {
			continue
		}
		metric, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, valueFloat, labelValues...)
		if err != nil {
			continue
		}
		ch <- metric
	}
}

// collectPerformanceMetrics emits Prometheus metrics for performance monitoring mode.
func (dc *DefCollector) collectPerformanceMetrics(ch chan<- prometheus.Metric, data map[string]string) {
	metrics := dc.GetMetrics()
	indicators := dc.filterIndicators(dc.def.Perf.Indicators, dc.metricsIndicators)
	for code, metricSuffix := range indicators {
		valueStr, ok := data[fmt.Sprintf("%d", code)]
		if !ok || valueStr == "" {
			continue
		}
		valueFloat, err := strconv.ParseFloat(valueStr, constants.DefaultBitSize)
		// Invalid numeric values are silently skipped: Prometheus collectors
		// should never expose malformed metrics; logging would be too noisy
		// at per-scrape granularity.
		if err != nil {
			continue
		}
		labelValues := resolveLabelValues(dc.def.PerfLabels, data)
		desc := metrics[metricSuffix]
		if desc == nil {
			continue
		}
		metric, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, valueFloat, labelValues...)
		if err != nil {
			continue
		}
		ch <- metric
	}
}

// resolveMetricValue returns the metric value string from data using the MetricDef.
// If Transform is set, it is used; otherwise the SourceKey is looked up directly.
func (dc *DefCollector) resolveMetricValue(md MetricDef, data map[string]string) string {
	if md.Transform != nil {
		return md.Transform(data)
	}
	return data[md.SourceKey]
}

// resolveLabelValues returns the label values for the given label definitions,
// resolving each from data using SourceKey or Transform.
func resolveLabelValues(labels []LabelDef, data map[string]string) []string {
	values := make([]string, len(labels))
	for i, ld := range labels {
		if ld.Transform != nil {
			values[i] = ld.Transform(data)
		} else {
			values[i] = data[ld.SourceKey]
		}
	}
	return values
}

// labelNamesFrom extracts the Prometheus label names from a slice of LabelDef.
func labelNamesFrom(labels []LabelDef) []string {
	names := make([]string, len(labels))
	for i, ld := range labels {
		names[i] = ld.Name
	}
	return names
}

// filterIndicators returns the subset of indicators matching the requested list.
// If no indicators are requested, all indicators are returned.
func (dc *DefCollector) filterIndicators(all map[int]string, requested []string) map[int]string {
	if len(requested) == 0 || (len(requested) == 1 && requested[0] == "") {
		return all
	}
	result := make(map[int]string)
	requestedSet := make(map[string]struct{})
	for _, r := range requested {
		for _, code := range strings.Split(r, ",") {
			requestedSet[strings.TrimSpace(code)] = struct{}{}
		}
	}
	for code, suffix := range all {
		if _, ok := requestedSet[fmt.Sprintf("%d", code)]; ok {
			result[code] = suffix
		}
	}
	return result
}
