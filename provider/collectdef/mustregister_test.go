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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"github.com/huawei/csm/v2/provider/cmicore"
	"github.com/huawei/csm/v2/server/prometheus-exporter/collector"
	"github.com/huawei/csm/v2/server/prometheus-exporter/exporterhandler"
	metricsCache "github.com/huawei/csm/v2/server/prometheus-exporter/metricscache"
	"github.com/huawei/csm/v2/utils/log"
)

// mockRestClient implements cmicore.RestClient for handler invocation tests.
type mockRestClient struct{}

func (m *mockRestClient) GetSingleByUrlKey(_ context.Context, _ string) (map[string]interface{}, error) {
	return map[string]interface{}{"ID": "1", "NAME": "test"}, nil
}
func (m *mockRestClient) GetListByUrlKey(_ context.Context, _ string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"ID": "1", "NAME": "obj1"},
		{"ID": "2", "NAME": "obj2"},
	}, nil
}
func (m *mockRestClient) GetCountByUrlKey(_ context.Context, _ string) (int, error) {
	return 2, nil
}
func (m *mockRestClient) GetPageByUrlKey(_ context.Context, _ string, _, _ int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"ID": "1", "NAME": "obj1"},
	}, nil
}
func (m *mockRestClient) QueryPerformanceData(_ context.Context, _ int, _ []string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"indicators":       []int{1},
			"indicator_values": []float64{100.0},
			"object_id":        "1",
		},
	}, nil
}
func (m *mockRestClient) Login(_ context.Context) error { return nil }
func (m *mockRestClient) Logout(_ context.Context)      {}

// applyMustRegisterMocks applies common gomonkey patches to isolate MustRegister
// from global state. Returns the patches and tracking flags for handler registrations.
func applyMustRegisterMocks(t *testing.T) (*gomonkey.Patches, *bool, *bool) {
	t.Helper()
	var gotObjRegistered, gotPerfRegistered bool

	patches := gomonkey.ApplyFunc(cmicore.RegisterObjectHandlerDirect,
		func(string, string, cmicore.ObjectHandler) { gotObjRegistered = true })
	patches.ApplyFunc(cmicore.RegisterPerformanceHandlerDirect,
		func(string, string, cmicore.PerformanceHandler) { gotPerfRegistered = true })
	patches.ApplyFuncReturn(cmicore.RegisterIndicatorMapping)
	patches.ApplyFuncReturn(collector.RegisterCollector)
	patches.ApplyFuncReturn(metricsCache.RegisterMetricsData)
	patches.ApplyFuncReturn(exporterhandler.AddMetricsObjectLegal)
	patches.ApplyFuncReturn(log.Errorf)

	return patches, &gotObjRegistered, &gotPerfRegistered
}

// applyMustRegisterMocksCapturing applies gomonkey patches and captures the registered
// handlers for later invocation. Returns patches, captured handlers, and tracking flags.
func applyMustRegisterMocksCapturing(t *testing.T) (*gomonkey.Patches,
	*cmicore.ObjectHandler, *cmicore.PerformanceHandler) {
	t.Helper()
	var capturedObjHandler cmicore.ObjectHandler
	var capturedPerfHandler cmicore.PerformanceHandler

	patches := gomonkey.ApplyFunc(cmicore.RegisterObjectHandlerDirect,
		func(_ string, _ string, handler cmicore.ObjectHandler) { capturedObjHandler = handler })
	patches.ApplyFunc(cmicore.RegisterPerformanceHandlerDirect,
		func(_ string, _ string, handler cmicore.PerformanceHandler) { capturedPerfHandler = handler })
	patches.ApplyFuncReturn(cmicore.RegisterIndicatorMapping)
	patches.ApplyFuncReturn(collector.RegisterCollector)
	patches.ApplyFuncReturn(metricsCache.RegisterMetricsData)
	patches.ApplyFuncReturn(exporterhandler.AddMetricsObjectLegal)
	patches.ApplyFuncReturn(log.Errorf)

	return patches, &capturedObjHandler, &capturedPerfHandler
}

func TestMustRegister_Success(t *testing.T) {
	// arrange
	def := &ObjectDef{
		CollectType:        "test_ut_success",
		SupportObject:      true,
		SupportPerformance: true,
		CollectMode:        CollectModeList,
		UrlKeys:            UrlKeys{ListKey: "GetTestList"},
		Perf: &PerfDef{
			TypeID:     888,
			Indicators: map[int]string{1: "test_iops"},
		},
		ObjectMetrics: []MetricDef{
			{Name: "test_metric", Help: "test", SourceKey: "TEST"},
		},
		ObjectLabels: []LabelDef{
			{Name: "id", SourceKey: "ID"},
			{Name: "name", SourceKey: "NAME"},
		},
		PerfLabels: []LabelDef{
			{Name: "id", SourceKey: "ID"},
		},
	}

	// mock
	patches, gotObjRegistered, gotPerfRegistered := applyMustRegisterMocks(t)
	defer patches.Reset()

	// action
	MustRegister("oceanStorage", def)

	// assert
	assert.Equal(t, true, *gotObjRegistered)
	assert.Equal(t, true, *gotPerfRegistered)
}

func TestMustRegister_CollectModeNoneSuccess(t *testing.T) {
	// arrange
	def := &ObjectDef{
		CollectType:   "test_ut_none",
		SupportObject: true,
		CollectMode:   CollectModeNone,
		ObjectMetrics: []MetricDef{
			{Name: "test_metric", Help: "test", SourceKey: "TEST"},
		},
		ObjectLabels: []LabelDef{
			{Name: "id", SourceKey: "ID"},
		},
	}

	// mock
	patches, gotObjRegistered, gotPerfRegistered := applyMustRegisterMocks(t)
	defer patches.Reset()

	// action
	MustRegister("oceanStorage", def)

	// assert — handlers should NOT be registered with CollectModeNone
	assert.Equal(t, false, *gotObjRegistered)
	assert.Equal(t, false, *gotPerfRegistered)
}

func TestMustRegister_SkipsPerfWhenNoPerfDefSuccess(t *testing.T) {
	// arrange
	def := &ObjectDef{
		CollectType:        "test_ut_no_perf",
		SupportObject:      true,
		SupportPerformance: true,
		CollectMode:        CollectModeList,
		UrlKeys:            UrlKeys{ListKey: "GetTestList"},
		ObjectMetrics: []MetricDef{
			{Name: "test_metric", Help: "test", SourceKey: "TEST"},
		},
		ObjectLabels: []LabelDef{
			{Name: "id", SourceKey: "ID"},
			{Name: "name", SourceKey: "NAME"},
		},
	}

	// mock
	patches, gotObjRegistered, gotPerfRegistered := applyMustRegisterMocks(t)
	defer patches.Reset()

	// action
	MustRegister("oceanStorage", def)

	// assert — object handler registered, performance handler skipped
	assert.Equal(t, true, *gotObjRegistered)
	assert.Equal(t, false, *gotPerfRegistered)
}

func TestMustRegister_UnsupportedCollectMode(t *testing.T) {
	// arrange
	def := &ObjectDef{
		CollectType:   "test_ut_bad_mode",
		SupportObject: true,
		CollectMode:   CollectMode(99),
		ObjectLabels: []LabelDef{
			{Name: "id", SourceKey: "ID"},
		},
	}

	// mock
	patches, gotObjRegistered, gotPerfRegistered := applyMustRegisterMocks(t)
	defer patches.Reset()

	// action
	MustRegister("oceanStorage", def)

	// assert — no handler registered due to unsupported CollectMode
	assert.Equal(t, false, *gotObjRegistered)
	assert.Equal(t, false, *gotPerfRegistered)
}

func TestMustRegister_MissingPerfLabels(t *testing.T) {
	// arrange — ObjectLabels has ID but missing NAME,
	// causing generatePerformanceHandler to return error via findIdNameKeys
	def := &ObjectDef{
		CollectType:        "test_ut_missing_name",
		SupportObject:      true,
		SupportPerformance: true,
		CollectMode:        CollectModeList,
		UrlKeys:            UrlKeys{ListKey: "GetTestList"},
		Perf: &PerfDef{
			TypeID:     777,
			Indicators: map[int]string{1: "test_iops"},
		},
		ObjectMetrics: []MetricDef{
			{Name: "test_metric", Help: "test", SourceKey: "TEST"},
		},
		ObjectLabels: []LabelDef{
			{Name: "id", SourceKey: "ID"},
		},
	}

	// mock
	patches, gotObjRegistered, gotPerfRegistered := applyMustRegisterMocks(t)
	defer patches.Reset()

	// action
	MustRegister("oceanStorage", def)

	// assert — object handler registered before perf error; perf handler not registered
	assert.Equal(t, true, *gotObjRegistered)
	assert.Equal(t, false, *gotPerfRegistered)
}

func TestMustRegister_MissingIdLabel(t *testing.T) {
	// arrange — ObjectLabels has NAME but no ID,
	// causing findIdNameKeys to return "no ID label" error
	def := &ObjectDef{
		CollectType:        "test_ut_missing_id",
		SupportObject:      true,
		SupportPerformance: true,
		CollectMode:        CollectModeList,
		UrlKeys:            UrlKeys{ListKey: "GetTestList"},
		Perf: &PerfDef{
			TypeID:     666,
			Indicators: map[int]string{1: "test_iops"},
		},
		ObjectMetrics: []MetricDef{
			{Name: "test_metric", Help: "test", SourceKey: "TEST"},
		},
		ObjectLabels: []LabelDef{
			{Name: "name", SourceKey: "NAME"},
		},
	}

	// mock
	patches, gotObjRegistered, gotPerfRegistered := applyMustRegisterMocks(t)
	defer patches.Reset()

	// action
	MustRegister("oceanStorage", def)

	// assert — object handler registered; perf handler not registered due to missing ID
	assert.Equal(t, true, *gotObjRegistered)
	assert.Equal(t, false, *gotPerfRegistered)
}

func TestMustRegister_CollectModeSingleSuccess(t *testing.T) {
	// arrange
	def := &ObjectDef{
		CollectType:   "test_ut_single",
		SupportObject: true,
		CollectMode:   CollectModeSingle,
		UrlKeys:       UrlKeys{SingleKey: "GetTestSingle"},
		ObjectMetrics: []MetricDef{
			{Name: "test_metric", Help: "test", SourceKey: "TEST"},
		},
		ObjectLabels: []LabelDef{
			{Name: "id", SourceKey: "ID"},
		},
	}

	// mock — capture handler for invocation
	patches, capturedObjHandler, _ := applyMustRegisterMocksCapturing(t)
	defer patches.Reset()

	// action
	MustRegister("oceanStorage", def)

	// assert — object handler registered
	assert.NotNil(t, *capturedObjHandler)

	// invoke the captured handler to exercise CollectModeSingle closure
	gotResp, gotErr := (*capturedObjHandler)(context.Background(), &mockRestClient{}, &cmicore.CollectRequest{})
	assert.NoError(t, gotErr)
	assert.NotNil(t, gotResp)
}

func TestMustRegister_CollectModePaginatedSuccess(t *testing.T) {
	// arrange
	def := &ObjectDef{
		CollectType:   "test_ut_paginated",
		SupportObject: true,
		CollectMode:   CollectModePaginated,
		UrlKeys:       UrlKeys{CountKey: "GetTestCount", PageKey: "GetTestPage"},
		ObjectMetrics: []MetricDef{
			{Name: "test_metric", Help: "test", SourceKey: "TEST"},
		},
		ObjectLabels: []LabelDef{
			{Name: "id", SourceKey: "ID"},
		},
	}

	// mock — capture handler for invocation
	patches, capturedObjHandler, _ := applyMustRegisterMocksCapturing(t)
	defer patches.Reset()

	// action
	MustRegister("oceanStorage", def)

	// assert — object handler registered
	assert.NotNil(t, *capturedObjHandler)

	// invoke the captured handler to exercise CollectModePaginated closure
	gotResp, gotErr := (*capturedObjHandler)(context.Background(), &mockRestClient{}, &cmicore.CollectRequest{})
	assert.NoError(t, gotErr)
	assert.NotNil(t, gotResp)
}

func TestMustRegister_PerformanceHandlerInvokedSuccess(t *testing.T) {
	// arrange
	def := &ObjectDef{
		CollectType:        "test_ut_perf_invoke",
		SupportObject:      true,
		SupportPerformance: true,
		CollectMode:        CollectModeList,
		UrlKeys:            UrlKeys{ListKey: "GetTestList"},
		Perf: &PerfDef{
			TypeID:     999,
			Indicators: map[int]string{1: "test_iops"},
		},
		ObjectMetrics: []MetricDef{
			{Name: "test_metric", Help: "test", SourceKey: "TEST"},
		},
		ObjectLabels: []LabelDef{
			{Name: "id", SourceKey: "ID"},
			{Name: "name", SourceKey: "NAME"},
		},
		PerfLabels: []LabelDef{
			{Name: "id", SourceKey: "ID"},
		},
	}

	// mock — capture handlers for invocation
	patches, _, capturedPerfHandler := applyMustRegisterMocksCapturing(t)
	defer patches.Reset()

	// action
	MustRegister("oceanStorage", def)

	// assert — performance handler registered
	assert.NotNil(t, *capturedPerfHandler)

	// invoke the captured performance handler to exercise closure body
	gotResp, gotErr := (*capturedPerfHandler)(context.Background(), &mockRestClient{}, &cmicore.CollectRequest{
		Indicators: []string{"test_iops"},
	})
	assert.NoError(t, gotErr)
	assert.NotNil(t, gotResp)
}

func TestMustRegister_CustomMetricsDataFactorySuccess(t *testing.T) {
	// arrange
	def := &ObjectDef{
		CollectType:        "test_ut_custom_factory",
		SupportObject:      true,
		SupportPerformance: false,
		CollectMode:        CollectModeList,
		UrlKeys:            UrlKeys{ListKey: "GetTestList"},
		ObjectMetrics: []MetricDef{
			{Name: "test_metric", Help: "test", SourceKey: "TEST"},
		},
		ObjectLabels: []LabelDef{
			{Name: "id", SourceKey: "ID"},
		},
		MetricsDataFactory: func(backendName, metricsType string) (interface{}, error) {
			return &mockMetricsData{}, nil
		},
	}
	var capturedFactory func(string, string) (metricsCache.MetricsData, error)

	// mock — capture the factory registered with RegisterMetricsData
	patches := gomonkey.ApplyFunc(cmicore.RegisterObjectHandlerDirect,
		func(string, string, cmicore.ObjectHandler) {})
	patches.ApplyFunc(cmicore.RegisterPerformanceHandlerDirect,
		func(string, string, cmicore.PerformanceHandler) {})
	patches.ApplyFuncReturn(cmicore.RegisterIndicatorMapping)
	patches.ApplyFuncReturn(collector.RegisterCollector)
	patches.ApplyFunc(exporterhandler.AddMetricsObjectLegal,
		func(string) {})
	patches.ApplyFunc(log.Errorf,
		func(string, ...interface{}) {})
	patches.ApplyFunc(metricsCache.RegisterMetricsData,
		func(_ string, factory func(string, string) (metricsCache.MetricsData, error)) {
			capturedFactory = factory
		})
	defer patches.Reset()

	// action
	MustRegister("oceanStorage", def)

	// assert — object handler registered and custom factory captured
	assert.NotNil(t, capturedFactory)

	// invoke the captured factory to exercise custom MetricsDataFactory path
	gotMd, gotErr := capturedFactory("backend1", "object")
	assert.NoError(t, gotErr)
	assert.NotNil(t, gotMd)
}

// mockMetricsData is a minimal MetricsData implementation for custom factory tests.
type mockMetricsData struct{}

func (m *mockMetricsData) GetMetricsDataResponse() *cmicore.CollectResponse {
	return &cmicore.CollectResponse{}
}
func (m *mockMetricsData) SetMetricsData(_ context.Context, _, _ string, _ []string) error {
	return nil
}
