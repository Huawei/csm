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

// Package cmicore provides a shared library for storage backend operations
package cmicore

import (
	"context"

	"k8s.io/client-go/kubernetes"

	sbcClient "github.com/Huawei/eSDK_K8S_Plugin/v4/pkg/client/clientset/versioned"
)

// CoreConfig holds the configuration for creating a new Core instance.
// All fields are required to be pre-initialized by the caller.
type CoreConfig struct {
	// KubeClient is the Kubernetes client for accessing K8s API
	// Used for reading Secrets and ConfigMaps
	KubeClient kubernetes.Interface

	// SbcClient is the StorageBackendClaim client for watching SBC CRs
	SbcClient sbcClient.Interface

	// BackendNamespace is the namespace where StorageBackendClaim CRs exist
	// Default: "huawei-csi"
	BackendNamespace string

	// QueryStoragePageSize is the maximum page size for storage queries
	// Used in concurrent pagination for LUN/Filesystem
	// Default: 100
	QueryStoragePageSize int

	// ClientMaxThreads is the maximum concurrent client threads
	// Default: 20
	ClientMaxThreads int
}

// ObjectClient provides the methods for collecting object metrics from storage
type ObjectClient interface {
	// GetSingleByUrlKey queries a single object by URL key
	GetSingleByUrlKey(ctx context.Context, urlKey string) (map[string]interface{}, error)
	// GetListByUrlKey queries a list of objects by URL key
	GetListByUrlKey(ctx context.Context, urlKey string) ([]map[string]interface{}, error)
	// GetCountByUrlKey queries the count of objects by URL key
	GetCountByUrlKey(ctx context.Context, urlKey string) (int, error)
	// GetPageByUrlKey queries a page of objects by URL key with pagination
	GetPageByUrlKey(ctx context.Context, urlKey string, start, end int) ([]map[string]interface{}, error)
}

// PerformanceClient provides the methods for collecting performance metrics from storage
type PerformanceClient interface {
	// QueryPerformanceData queries performance indicators for an object type
	QueryPerformanceData(ctx context.Context, objectType int, indicators []string) ([]map[string]interface{}, error)
}

// RestClient defines the common interface for storage REST clients.
// Both centralized and distributed storage clients must implement
// session management (Login/Logout) and data query methods.
type RestClient interface {
	ObjectClient
	PerformanceClient
	// Login authenticates with the storage backend
	Login(ctx context.Context) error
	// Logout terminates the session with the storage backend
	Logout(ctx context.Context)
}

// ClientInfo storage client info
type ClientInfo struct {
	// storage name
	StorageName string
	// storage type, e.g. oceanStorage
	StorageType string
	// volume type, e.g. nas or lun
	VolumeType string
	// storage Client
	Client RestClient
}

// LabelRequest parameters for label operations (Go-native type, no protobuf)
type LabelRequest struct {
	// VolumeId is the Kubernetes PV volumeHandle
	// REQUIRED - format: backendName/volumeName
	VolumeId string

	// LabelName is the label name (PV name or Pod name)
	// REQUIRED
	LabelName string

	// Kind is the resource kind: "PersistentVolume" or "Pod"
	// REQUIRED
	Kind string

	// Namespace is the Pod namespace
	// OPTIONAL - default: "default"
	Namespace string

	// ClusterName is the cluster identifier for PV labels
	// OPTIONAL - only used for PersistentVolume kind
	ClusterName string

	// Parameters contains extension parameters
	// OPTIONAL
	Parameters map[string]string
}

// CollectRequest parameters for metrics collection (Go-native type, no protobuf)
type CollectRequest struct {
	// BackendName is the StorageBackendClaim name
	// REQUIRED
	BackendName string

	// CollectType is the type of data to collect
	// Values: "lun", "array", "controller", "filesystem", "storagepool"
	// REQUIRED
	CollectType string

	// MetricsType is the type of metrics
	// Values: "object" or "performance"
	// REQUIRED
	MetricsType string

	// Indicators are the indicator names for performance collection
	// OPTIONAL - only used for performance metrics
	Indicators []string

	// Parameters contains extension parameters
	// OPTIONAL
	Parameters map[string]string
}

// CollectResponse returned by metrics collection (Go-native type, no protobuf)
type CollectResponse struct {
	// BackendName is the StorageBackendClaim name
	BackendName string

	// CollectType is the type of data collected
	CollectType string

	// MetricsType is "object" or "performance"
	MetricsType string

	// Details are the collected data items
	Details []*CollectDetail
}

// CollectDetail represents a single collected data item
type CollectDetail struct {
	// Data contains key-value pairs of collected data
	Data map[string]string
}

// PageResultTuple page query result for concurrent pagination
type PageResultTuple struct {
	Error error
	Data  []map[string]interface{}
}

// PerformanceIndicators performance information
type PerformanceIndicators struct {
	Indicators      []int     `json:"indicators"`
	IndicatorValues []float64 `json:"indicator_values"`
	ObjectId        string    `json:"object_id"`
}

// LabelValidator contains all fields to be verified for label operations
type LabelValidator struct {
	VolumeId    string
	LabelName   string
	Kind        string
	Namespace   string
	ClusterName string
}
