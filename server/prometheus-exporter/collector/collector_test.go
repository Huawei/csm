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

package collector

import (
	"context"
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
)

func TestRegisterCollector(t *testing.T) {
	// arrange
	var CollectorName = "fake_collector"
	var NewCollectorFunc = func(backendName, monitorType string, metricsIndicators []string,
		pMetricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
		return nil, nil
	}

	// mock
	RegisterCollector(CollectorName, NewCollectorFunc)

	// assert
	newCollectorFunc, ok := factories[CollectorName]
	if !ok {
		t.Errorf("RegisterCollector() want func, but got = %v", nil)
	}

	if reflect.ValueOf(NewCollectorFunc).Pointer() !=
		reflect.ValueOf(newCollectorFunc).Pointer() {
		t.Error("RegisterCollector get func error")
	}
}

func TestBaseCollector_BuildDesc(t *testing.T) {
	// arrange
	var mockMetricsLabelMap = map[string][]string{
		"fake_key1": {"fake_label1", "fake_label2"},
		"fake_key2": {"fake_label2", "fake_label3"}}
	var mockMetricsHelpMap = map[string]string{
		"fake_key1": "fake help info1",
		"fake_key2": "fake help info2",
	}
	var mockCollector = BaseCollector{
		backendName:     "fake_name",
		monitorType:     "fake_monitortype",
		collectorName:   "fake_name",
		metricsHelpMap:  mockMetricsHelpMap,
		metricsLabelMap: mockMetricsLabelMap,
		metrics:         make(map[string]*prometheus.Desc),
	}
	var mockMetrics = map[string]*prometheus.Desc{
		"fake_key1": prometheus.NewDesc(
			prometheus.BuildFQName(
				MetricsNamespace, mockCollector.collectorName, "fake_key1"),
			"fake help info1",
			mockCollector.metricsLabelMap["fake_key1"],
			nil),
		"fake_key2": prometheus.NewDesc(
			prometheus.BuildFQName(
				MetricsNamespace, mockCollector.collectorName, "fake_key2"),
			"fake help info2",
			mockCollector.metricsLabelMap["fake_key2"],
			nil),
	}

	// action
	mockCollector.BuildDesc()

	// assert
	if !reflect.DeepEqual(mockMetrics, mockCollector.metrics) {
		t.Errorf("BuildDesc() want = %v, but got = %v", mockMetrics, mockCollector)
	}
}

func TestNewCollectorSet_FactoriesNotHaveCollector(t *testing.T) {
	// arrange
	var (
		ctx              = context.TODO()
		CollectorName    = "fake_collector"
		NewCollectorFunc = func(backendName, monitorType string, metricsIndicators []string,
			pMetricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
			return nil, nil
		}
	)
	factories[CollectorName] = NewCollectorFunc
	var wantCollector *CollectorSet = nil

	// action
	got, err := NewCollectorSet(ctx, nil, "", "",
		nil)

	// assert
	if (err != nil) != true {
		t.Errorf("NewCollectorSet() error = %v, wantErr %v", err, true)
		return
	}
	if !reflect.DeepEqual(got, wantCollector) {
		t.Errorf("NewCollectorSet() got = %v, want %v", got, nil)
	}
}

func TestNewCollectorSet_GetCollectorSetSuccess(t *testing.T) {
	// arrange
	var (
		ctx              = context.TODO()
		CollectorName    = "fake_collector"
		NewCollectorFunc = func(backendName, monitorType string, metricsIndicators []string,
			pMetricsDataCache *metricsCache.MetricsDataCache) (prometheus.Collector, error) {
			return nil, nil
		}
	)
	factories[CollectorName] = NewCollectorFunc
	mockparams := map[string][]string{"fake_collector": {"fake_data"}}
	var wantCollector = &CollectorSet{
		collectors: []prometheus.Collector{nil},
	}

	// action
	got, err := NewCollectorSet(ctx, mockparams,
		"fake_backend_name", "fake_type",
		nil)

	// assert
	if (err != nil) != false {
		t.Errorf("NewCollectorSet() error = %v, wantErr %v", err, true)
		return
	}
	if !reflect.DeepEqual(got, wantCollector) {
		t.Errorf("NewCollectorSet() got = %v, want %v", got, nil)
	}
}

func buildMockBaseCollector() *BaseCollector {
	var mockMetricsLabelMap = map[string][]string{"fake_key1": {"fake_label1", "fake_label2", "fake_label3"}}
	var mockMetricsHelpMap = map[string]string{"fake_key1": "fake help info1"}
	var mockLabelParseMap = map[string]parseRelation{
		"fake_label1": {"backendName", parseStorageData},
		"fake_label2": {"collectorName", parseStorageData},
		"fake_label3": {"fake_data", parseStorageData},
	}
	var mockMetricsParseMap = map[string]parseRelation{
		"fake_key1": {"", parseStorageReturnZero},
	}
	var mockMetrics = map[string]*prometheus.Desc{
		"fake_key1": prometheus.NewDesc(
			prometheus.BuildFQName(
				MetricsNamespace, "fake_name", "fake_key1"),
			"fake help info1",
			mockMetricsLabelMap["fake_key1"],
			nil),
	}
	mockCollectDetail := storageGRPC.CollectDetail{
		Data: map[string]string{"fake_data": "test_data"},
	}
	mockCollectResponse := storageGRPC.CollectResponse{
		BackendName: "fake_backend_name",
		CollectType: "fake_collector_name",
		MetricsType: "fake_type",
		Details:     []*storageGRPC.CollectDetail{&mockCollectDetail},
	}
	mockMetricsData := metricsCache.BaseMetricsData{
		MetricsType:         "fake_collector_name",
		MetricsDataResponse: &mockCollectResponse,
	}
	mockMetricsDataCache := &metricsCache.MetricsDataCache{
		BackendName: "fake_name",
		CacheDataMap: map[string]metricsCache.MetricsData{
			"fake_collector_name": &mockMetricsData},
	}
	var mockCollector = BaseCollector{
		backendName:      "fake_backend_name",
		monitorType:      "fake_monitor_type",
		collectorName:    "fake_collector_name",
		metricsHelpMap:   mockMetricsHelpMap,
		metricsLabelMap:  mockMetricsLabelMap,
		labelParseMap:    mockLabelParseMap,
		metricsParseMap:  mockMetricsParseMap,
		metricsDataCache: mockMetricsDataCache,
		metrics:          mockMetrics,
	}
	return &mockCollector
}

func TestBaseCollector_Collect_ParseSuccess(t *testing.T) {
	// arrange
	mockCollector := buildMockBaseCollector()
	mockMetricChan := make(chan prometheus.Metric, 2)
	mockLabelValueSlice := []string{
		"fake_backend_name", "fake_collector_name", "test_data"}
	wantPrometheusMetric := prometheus.MustNewConstMetric(
		mockCollector.metrics["fake_key1"],
		prometheus.GaugeValue,
		0.0,
		mockLabelValueSlice...,
	)

	// action
	mockCollector.Collect(mockMetricChan)
	defer close(mockMetricChan)

	// assert
	got, ok := <-mockMetricChan
	if !ok {
		t.Errorf("Collect() got = %v", got)
	}
	if !reflect.DeepEqual(got, wantPrometheusMetric) {
		t.Errorf("Collect() got = %v, want %v", got, wantPrometheusMetric)
	}
}

func TestBaseCollector_SetBackendName(t *testing.T) {
	// arrange
	mockBackendName := "fake_backend_name"
	mockCollector := &BaseCollector{}

	// action
	mockCollector.SetBackendName(mockBackendName)

	// assert
	if !reflect.DeepEqual(mockCollector.backendName, mockBackendName) {
		t.Errorf("SetBackendName() got = %v, want %v", mockCollector.backendName, mockBackendName)
	}
}

func TestBaseCollector_SetMonitorType(t *testing.T) {
	// arrange
	mockMonitorType := "fake_monitor_type"
	mockCollector := &BaseCollector{}

	// action
	mockCollector.SetMonitorType(mockMonitorType)

	// assert
	if !reflect.DeepEqual(mockCollector.monitorType, mockMonitorType) {
		t.Errorf("SetBackendName() got = %v, want %v", mockCollector.monitorType, mockMonitorType)
	}
}

func TestBaseCollector_SetCollectorName(t *testing.T) {
	// arrange
	mockCollectorName := "fake_collector_name"
	mockCollector := &BaseCollector{}

	// action
	mockCollector.SetCollectorName(mockCollectorName)

	// assert
	if !reflect.DeepEqual(mockCollector.collectorName, mockCollectorName) {
		t.Errorf("SetBackendName() got = %v, want %v", mockCollector.collectorName, mockCollectorName)
	}
}

func TestBaseCollector_SetMetricsHelpMap(t *testing.T) {
	// arrange
	metricsHelpMap := map[string]string{
		"fake_key": "fake_help_info",
	}
	mockCollector := &BaseCollector{}
	want := map[string][]string{
		"fake_key": {"fake_help_info"},
	}
	equalMaps := func(a, b map[string][]string) bool {
		if len(a) != len(b) {
			return false
		}
		for k, v := range a {
			if vb, ok := b[k]; !ok || !reflect.DeepEqual(vb, v) {
				return false
			}
		}
		return true
	}

	// action
	mockCollector.SetMetricsHelpMap(metricsHelpMap)

	// assert
	if equalMaps(mockCollector.metricsLabelMap, want) {
		t.Errorf("SetBackendName() got = %v, want %v", mockCollector.metricsHelpMap, want)
	}
}

func TestBaseCollector_SetMetricsLabelMap(t *testing.T) {
	// arrange
	metricsLabelMap := map[string][]string{
		"fake_key": {"fake_label1", "fake_label2"},
	}
	mockCollector := &BaseCollector{}

	// action
	mockCollector.SetMetricsLabelMap(metricsLabelMap)

	// assert
	if !reflect.DeepEqual(mockCollector.metricsLabelMap, metricsLabelMap) {
		t.Errorf("SetBackendName() got = %v, want %v", mockCollector.metricsLabelMap, metricsLabelMap)
	}
}

func TestBaseCollector_SetLabelParseMap(t *testing.T) {
	// arrange
	labelParseMap := map[string]parseRelation{
		"fake_label": {"fake_key", parseStorageData},
	}
	mockCollector := &BaseCollector{}

	// action
	mockCollector.SetLabelParseMap(labelParseMap)

	// assert
	if !reflect.DeepEqual(mockCollector.labelParseMap, labelParseMap) {
		t.Errorf("SetBackendName() got = %v, want %v", mockCollector.labelParseMap, labelParseMap)
	}
}

func TestBaseCollector_SetMetricsParseMap(t *testing.T) {
	// arrange
	metricsParseMap := map[string]parseRelation{
		"fake_key": {"", parseStorageReturnZero},
	}
	mockCollector := &BaseCollector{}

	// action
	mockCollector.SetMetricsParseMap(metricsParseMap)

	// assert
	if !reflect.DeepEqual(mockCollector.metricsParseMap, metricsParseMap) {
		t.Errorf("SetBackendName() got = %v, want %v", mockCollector.metricsParseMap, metricsParseMap)
	}
}

func TestBaseCollector_SetMetricsDataCache(t *testing.T) {
	// arrange
	metricsDataCache := &metricsCache.MetricsDataCache{
		BackendName:  "fake_backend_name",
		CacheDataMap: map[string]metricsCache.MetricsData{}}
	mockCollector := &BaseCollector{}

	// action
	mockCollector.SetMetricsDataCache(metricsDataCache)

	// assert
	if !reflect.DeepEqual(mockCollector.metricsDataCache, metricsDataCache) {
		t.Errorf("SetBackendName() got = %v, want %v", mockCollector.metricsDataCache, metricsDataCache)
	}
}

func TestBaseCollector_SetMetrics(t *testing.T) {
	// arrange
	metrics := make(map[string]*prometheus.Desc)
	mockCollector := &BaseCollector{}

	// action
	mockCollector.SetMetrics(metrics)

	// assert
	if !reflect.DeepEqual(mockCollector.metrics, metrics) {
		t.Errorf("SetBackendName() got = %v, want %v", mockCollector.metrics, metrics)
	}
}

func TestNewPerformanceBaseCollector(t *testing.T) {
	// arrange
	var mockMetricsIndicators []string

	// action
	_, err := NewPerformanceBaseCollector("fake_backend", "performance",
		"fake_collector", mockMetricsIndicators, nil)

	// assert
	if err == nil {
		t.Errorf("NewPerformanceBaseCollector() error = %v, wantErr %v", err, true)
		return
	}
}
