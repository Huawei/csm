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

// CollectMode defines how data is fetched from the storage REST API.
type CollectMode int

const (
	// CollectModeNone defines unstandard REST collection (e.g., vstore)
	CollectModeNone CollectMode = iota
	// CollectModeSingle defines single object query (e.g., array)
	CollectModeSingle
	// CollectModeList defines list query without pagination (e.g., controller)
	CollectModeList
	// CollectModePaginated defines count + concurrent page queries (e.g., lun)
	CollectModePaginated

	labelIdKey   = "id"
	labelNameKey = "name"
)

// TransformFunc transforms raw data fields into Prometheus metric or label values.
type TransformFunc func(data map[string]string) string

// UrlKeys holds the storageApiMap URL key names for each collection mode.
type UrlKeys struct {
	SingleKey string // CollectModeSingle (e.g., "GetSystemInfo")
	ListKey   string // CollectModeList (e.g., "GetControllers")
	CountKey  string // CollectModePaginated count (e.g., "GetLunCount")
	PageKey   string // CollectModePaginated page (e.g., "GetLuns")
}

// MetricDef defines a single Prometheus metric within an ObjectDef.
type MetricDef struct {
	SourceKey string        // Key in Detail.Data (e.g., "CPUUSAGE"). Empty if Transform computes value.
	Name      string        // Prometheus metric name suffix (e.g., "cpu_usage")
	Help      string        // Prometheus help text
	Transform TransformFunc // Optional; In Default, value is set by data[SourceKey]
	Labels    []LabelDef    // Per-metric labels. If nil, uses ObjectDef.ObjectLabels.
}

// LabelDef defines a single Prometheus label.
type LabelDef struct {
	SourceKey string        // Key in Detail.Data. Empty if Transform computes value.
	Name      string        // Prometheus label name
	Transform TransformFunc // Optional; In Default, value is set by data[SourceKey]
}

// PerfDef defines performance indicator metadata.
type PerfDef struct {
	TypeID     int            // Storage API object type ID (e.g., 11 for lun)
	Indicators map[int]string // indicator code → metric suffix (e.g., 22→"total_iops")
}

// ObjectDef declaratively defines all collection metadata for a storage object type.
type ObjectDef struct {
	CollectType        string
	PrometheusName     string // Override subsystem name; defaults to CollectType
	SupportObject      bool
	SupportPerformance bool
	CollectMode        CollectMode // CollectModeNone skips ObjectHandler generation
	UrlKeys            UrlKeys
	Perf               *PerfDef
	ObjectMetrics      []MetricDef
	ObjectLabels       []LabelDef // Default labels for object metrics (used when MetricDef.Labels is nil)
	PerfLabels         []LabelDef
	MetricsDataFactory MetricsDataFactoryFunc // Optional; defaults to StorageMetricsData
}

// MetricsDataFactoryFunc creates a MetricsData instance. Return type is interface{}
// to avoid importing the metricscache package (would create circular dep via cmicore).
// The returned value MUST implement metricscache.MetricsData; otherwise
// registerMetricsDataFactory returns an error at runtime.
type MetricsDataFactoryFunc func(backendName, metricsType string) (interface{}, error)

// PrometheusSubsystemName returns the Prometheus metric subsystem name.
func (d *ObjectDef) PrometheusSubsystemName() string {
	if d.PrometheusName != "" {
		return d.PrometheusName
	}
	return d.CollectType
}
