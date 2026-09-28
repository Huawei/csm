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
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/huawei/csm/v2/provider/cmicore"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/server/prometheus-exporter/collector"
	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
)

// ---------------------------------------------------------------------------
// TestNewDefCollectorFactory
// ---------------------------------------------------------------------------

func TestNewDefCollectorFactory(t *testing.T) {
	tests := []struct {
		name              string
		def               *ObjectDef
		monitorType       string
		metricsIndicators []string
		wantErr           bool
		wantMonitorType   string
	}{
		{
			name:              "object mode supported",
			def:               &ObjectDef{CollectType: "test", SupportObject: true},
			monitorType:       constants.Object,
			metricsIndicators: nil,
			wantErr:           false,
			wantMonitorType:   constants.Object,
		},
		{
			name:              "performance mode supported",
			def:               &ObjectDef{CollectType: "test", SupportPerformance: true, Perf: &PerfDef{TypeID: 1}},
			monitorType:       constants.Performance,
			metricsIndicators: []string{"1"},
			wantErr:           false,
			wantMonitorType:   constants.Performance,
		},
		{
			name:              "unsupported object mode",
			def:               &ObjectDef{CollectType: "test", SupportObject: false},
			monitorType:       constants.Object,
			metricsIndicators: nil,
			wantErr:           true,
			wantMonitorType:   "",
		},
		{
			name:              "unsupported performance mode",
			def:               &ObjectDef{CollectType: "test", SupportPerformance: false},
			monitorType:       constants.Performance,
			metricsIndicators: nil,
			wantErr:           true,
			wantMonitorType:   "",
		},
		{
			name:              "performance mode with nil Perf",
			def:               &ObjectDef{CollectType: "test", SupportPerformance: true, Perf: nil},
			monitorType:       constants.Performance,
			metricsIndicators: nil,
			wantErr:           true,
			wantMonitorType:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// act
			factory := NewDefCollectorFactory(tt.def)
			c, err := factory("backend1", tt.monitorType, tt.metricsIndicators, &metricsCache.MetricsDataCache{})

			// assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			dc, ok := c.(*DefCollector)
			assert.True(t, ok, "expected *DefCollector")
			assert.Equal(t, tt.wantMonitorType, dc.monitorType)
		})
	}
}

// ---------------------------------------------------------------------------
// TestBuildDesc
// ---------------------------------------------------------------------------

func TestBuildDescObjectMode(t *testing.T) {
	// arrange
	def := &ObjectDef{
		CollectType: "controller",
		ObjectMetrics: []MetricDef{
			{Name: "cpu_usage", Help: "CPU usage", SourceKey: "CPUUSAGE"},
			{Name: "memory_usage", Help: "Memory usage", SourceKey: "MEMORYUSAGE"},
		},
		ObjectLabels: []LabelDef{
			{Name: "endpoint", SourceKey: "backendName"},
			{Name: "id", SourceKey: "ID"},
		},
	}
	dc := &DefCollector{
		BaseCollector: (&collector.BaseCollector{}).
			SetMonitorType(constants.Object).
			SetCollectorName("controller").
			SetMetrics(make(map[string]*prometheus.Desc)),
		def:         def,
		monitorType: constants.Object,
	}

	// act
	dc.BuildDesc()

	// assert
	metrics := dc.GetMetrics()
	assert.Len(t, metrics, 2)
	desc, ok := metrics["cpu_usage"]
	assert.True(t, ok, "cpu_usage metric not found")
	assert.Contains(t, desc.String(), "huawei_storage_controller_cpu_usage")
}

func TestBuildDescObjectModeWithPrometheusName(t *testing.T) {
	// arrange — PrometheusName overrides CollectType in FQName
	def := &ObjectDef{
		CollectType: "storagepool",
		ObjectMetrics: []MetricDef{
			{Name: "total_capacity", Help: "Total capacity(GB) of storage pool", SourceKey: "USERTOTALCAPACITY"},
		},
		ObjectLabels: []LabelDef{
			{Name: "endpoint", SourceKey: "backendName"},
		},
	}
	dc := &DefCollector{
		BaseCollector: (&collector.BaseCollector{}).
			SetMonitorType(constants.Object).
			SetCollectorName("storagepool").
			SetMetrics(make(map[string]*prometheus.Desc)),
		def:         def,
		monitorType: constants.Object,
	}

	// act
	dc.BuildDesc()

	// assert
	desc, ok := dc.GetMetrics()["total_capacity"]
	assert.True(t, ok, "total_capacity metric not found")
	assert.Contains(t, desc.String(), "huawei_storage_storagepool_total_capacity")
}

func TestBuildDescPerformanceModeWithFilter(t *testing.T) {
	// arrange
	def := &ObjectDef{
		CollectType: "controller",
		Perf: &PerfDef{
			TypeID: 12,
			Indicators: map[int]string{
				22: "response_time",
				23: "iops",
				24: "throughput",
			},
		},
		PerfLabels: []LabelDef{
			{Name: "endpoint", SourceKey: "backendName"},
		},
	}
	dc := &DefCollector{
		BaseCollector: (&collector.BaseCollector{}).
			SetMonitorType(constants.Performance).
			SetCollectorName("controller").
			SetMetrics(make(map[string]*prometheus.Desc)),
		def:               def,
		monitorType:       constants.Performance,
		metricsIndicators: []string{"22"},
	}

	// act
	dc.BuildDesc()

	// assert — only indicator 22 should be in metrics
	metrics := dc.GetMetrics()
	assert.Len(t, metrics, 1)
	assert.Contains(t, metrics, "response_time")
	assert.NotContains(t, metrics, "iops")
}

// ---------------------------------------------------------------------------
// TestCollect
// ---------------------------------------------------------------------------

func TestCollectObjectMetrics(t *testing.T) {
	def := &ObjectDef{
		CollectType: "controller",
		ObjectMetrics: []MetricDef{
			{Name: "cpu_usage", Help: "CPU usage", SourceKey: "CPUUSAGE"},
		},
		ObjectLabels: []LabelDef{
			{Name: "endpoint", SourceKey: "backendName"},
			{Name: "id", SourceKey: "ID"},
		},
	}

	tests := []struct {
		name          string
		detailData    map[string]string // nil means no cache entry for this collectType
		expectedCount int
	}{
		{
			name:          "normal value",
			detailData:    map[string]string{"CPUUSAGE": "85.5", "ID": "12345"},
			expectedCount: 1,
		},
		{
			name:          "invalid numeric value",
			detailData:    map[string]string{"CPUUSAGE": "not-a-number", "ID": "12345"},
			expectedCount: 0,
		},
		{
			name:          "skipReportValue",
			detailData:    map[string]string{"CPUUSAGE": SkipReportValue, "ID": "12345"},
			expectedCount: 0,
		},
		{
			name:          "empty value",
			detailData:    map[string]string{"CPUUSAGE": "", "ID": "12345"},
			expectedCount: 0,
		},
		{
			name:          "no cache data",
			detailData:    nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange
			mockCache := buildMockCache("controller", "test-backend", tt.detailData)
			dc := &DefCollector{
				BaseCollector: (&collector.BaseCollector{}).
					SetMonitorType(constants.Object).
					SetCollectorName("controller").
					SetMetricsDataCache(mockCache).
					SetMetrics(make(map[string]*prometheus.Desc)),
				def:         def,
				monitorType: constants.Object,
			}
			dc.BuildDesc()

			// act
			ch := make(chan prometheus.Metric, 10)
			dc.Collect(ch)
			close(ch)

			// assert
			count := 0
			for range ch {
				count++
			}
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestCollectObjectMetrics_PerMetricLabels(t *testing.T) {
	// arrange — MetricDef.Labels override ObjectDef.ObjectLabels
	def := &ObjectDef{
		CollectType: "controller",
		ObjectMetrics: []MetricDef{
			{
				Name:      "cpu_usage",
				Help:      "CPU usage",
				SourceKey: "CPUUSAGE",
				Labels: []LabelDef{
					{Name: "endpoint", SourceKey: "backendName"},
					{Name: "id", SourceKey: "ID"},
				},
			},
			{
				Name:      "status",
				Help:      "Status",
				SourceKey: "STATUS",
				// No Labels — uses ObjectLabels
			},
		},
		ObjectLabels: []LabelDef{
			{Name: "endpoint", SourceKey: "backendName"},
			{Name: "name", SourceKey: "NAME"},
		},
	}
	mockCache := buildMockCache("controller", "test-backend", map[string]string{
		"CPUUSAGE": "50.0",
		"STATUS":   "1",
		"ID":       "12345",
		"NAME":     "ctrl-0",
	})
	dc := &DefCollector{
		BaseCollector: (&collector.BaseCollector{}).
			SetMonitorType(constants.Object).
			SetCollectorName("controller").
			SetMetricsDataCache(mockCache).
			SetMetrics(make(map[string]*prometheus.Desc)),
		def:         def,
		monitorType: constants.Object,
	}
	dc.BuildDesc()

	// act
	ch := make(chan prometheus.Metric, 10)
	dc.Collect(ch)
	close(ch)

	// assert — both metrics should be emitted
	count := 0
	for range ch {
		count++
	}
	assert.Equal(t, 2, count)
}

func TestCollectObjectMetrics_WithTransformFunc(t *testing.T) {
	// arrange — capacity_usage computed via TransformFunc
	capacityUsageTransform := func(data map[string]string) string {
		total, err1 := strconv.ParseFloat(data["TOTAL"], constants.DefaultBitSize)
		used, err2 := strconv.ParseFloat(data["USED"], constants.DefaultBitSize)
		if err1 != nil || err2 != nil || total == 0 {
			return ""
		}
		return fmt.Sprintf("%.2f", used/total*100)
	}

	def := &ObjectDef{
		CollectType: "storagepool",
		ObjectMetrics: []MetricDef{
			{Name: "capacity_usage", Help: "Capacity usage %", Transform: capacityUsageTransform},
		},
		ObjectLabels: []LabelDef{
			{Name: "endpoint", SourceKey: "backendName"},
			{Name: "id", SourceKey: "ID"},
		},
	}
	mockCache := buildMockCache("storagepool", "test-backend", map[string]string{
		"TOTAL": "1000",
		"USED":  "750",
		"ID":    "sp-01",
	})
	dc := &DefCollector{
		BaseCollector: (&collector.BaseCollector{}).
			SetMonitorType(constants.Object).
			SetCollectorName("storagepool").
			SetMetricsDataCache(mockCache).
			SetMetrics(make(map[string]*prometheus.Desc)),
		def:         def,
		monitorType: constants.Object,
	}
	dc.BuildDesc()

	// act
	ch := make(chan prometheus.Metric, 10)
	dc.Collect(ch)
	close(ch)

	// assert
	metric, ok := <-ch
	assert.True(t, ok, "expected metric collected")
	assert.NotNil(t, metric)
}

func TestCollectPerformanceMetrics(t *testing.T) {
	// arrange
	def := &ObjectDef{
		CollectType: "controller",
		Perf: &PerfDef{
			TypeID: 12,
			Indicators: map[int]string{
				22: "response_time",
			},
		},
		PerfLabels: []LabelDef{
			{Name: "endpoint", SourceKey: "backendName"},
			{Name: "id", SourceKey: "ID"},
		},
	}
	mockCache := buildMockCache("controller", "test-backend", map[string]string{
		"22": "3.14",
		"ID": "12345",
	})
	dc := &DefCollector{
		BaseCollector: (&collector.BaseCollector{}).
			SetMonitorType(constants.Performance).
			SetCollectorName("controller").
			SetMetricsDataCache(mockCache).
			SetMetrics(make(map[string]*prometheus.Desc)),
		def:               def,
		monitorType:       constants.Performance,
		metricsIndicators: []string{"22"},
	}
	dc.BuildDesc()

	// act
	ch := make(chan prometheus.Metric, 10)
	dc.Collect(ch)
	close(ch)

	// assert
	metric, ok := <-ch
	assert.True(t, ok, "expected metric collected")
	assert.NotNil(t, metric)
}

// ---------------------------------------------------------------------------
// TestFilterIndicators (unexported, multiple input formats)
// ---------------------------------------------------------------------------

func TestFilterIndicators(t *testing.T) {
	all := map[int]string{
		22: "response_time",
		23: "iops",
		24: "throughput",
	}

	tests := []struct {
		name      string
		requested []string
		expected  map[int]string
	}{
		{
			name:      "all when nil",
			requested: nil,
			expected:  all,
		},
		{
			name:      "all when empty string",
			requested: []string{""},
			expected:  all,
		},
		{
			name:      "subset",
			requested: []string{"22", "24"},
			expected: map[int]string{
				22: "response_time",
				24: "throughput",
			},
		},
		{
			name:      "comma separated",
			requested: []string{"22,23"},
			expected: map[int]string{
				22: "response_time",
				23: "iops",
			},
		},
		{
			name:      "no match",
			requested: []string{"99"},
			expected:  map[int]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := &DefCollector{}
			result := dc.filterIndicators(all, tt.requested)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildMockCache constructs a MetricsDataCache with a single CollectDetail.
// If detailData is nil, the cache will have no entry for the given collectType.
func buildMockCache(collectType, backendName string, detailData map[string]string) *metricsCache.MetricsDataCache {
	if detailData == nil {
		return &metricsCache.MetricsDataCache{
			BackendName:  backendName,
			CacheDataMap: map[string]metricsCache.MetricsData{},
		}
	}
	mockCollectDetail := &cmicore.CollectDetail{Data: detailData}
	mockCollectResponse := &cmicore.CollectResponse{
		BackendName: backendName,
		CollectType: collectType,
		Details:     []*cmicore.CollectDetail{mockCollectDetail},
	}
	mockMetricsData := &metricsCache.BaseMetricsData{
		MetricsType:         collectType,
		MetricsDataResponse: mockCollectResponse,
	}
	return &metricsCache.MetricsDataCache{
		BackendName: backendName,
		CacheDataMap: map[string]metricsCache.MetricsData{
			collectType: mockMetricsData,
		},
	}
}
