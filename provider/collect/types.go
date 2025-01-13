/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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

// PageResultTuple page query result
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

// ArrayObject array object information
type ArrayObject struct {
	Id                string `json:"ID" metrics:"ID"`
	ProductModeString string `json:"productModeString" metrics:"productModeString"`
	ProductMode       string `json:"PRODUCTMODE" metrics:"PRODUCTMODE"`
	ProductVersion    string `json:"PRODUCTVERSION" metrics:"PRODUCTVERSION"`
	HealthStatus      string `json:"HEALTHSTATUS" metrics:"HEALTHSTATUS"`
	RunningStatus     string `json:"RUNNINGSTATUS" metrics:"RUNNINGSTATUS"`
}

// LunObject lun object information
type LunObject struct {
	Id            string `json:"ID" metrics:"ID"`
	Name          string `json:"NAME" metrics:"NAME"`
	Capacity      string `json:"CAPACITY" metrics:"CAPACITY"`
	AllocCapacity string `json:"ALLOCCAPACITY" metrics:"ALLOCCAPACITY"`
}

// ControllerObject controller object information
type ControllerObject struct {
	Id            string `json:"ID" metrics:"ID"`
	Name          string `json:"NAME" metrics:"NAME"`
	CpuUsage      string `json:"CPUUSAGE" metrics:"CPUUSAGE"`
	MemoryUsage   string `json:"MEMORYUSAGE" metrics:"MEMORYUSAGE"`
	RunningStatus string `json:"RUNNINGSTATUS" metrics:"RUNNINGSTATUS"`
	HealthStatus  string `json:"HEALTHSTATUS" metrics:"HEALTHSTATUS"`
}

// StoragePoolObject storage pool object information
type StoragePoolObject struct {
	Id            string `json:"ID" metrics:"ID"`
	Name          string `json:"NAME" metrics:"NAME"`
	FreeCapacity  string `json:"USERFREECAPACITY" metrics:"USERFREECAPACITY"`
	UsedCapacity  string `json:"USERCONSUMEDCAPACITY" metrics:"USERCONSUMEDCAPACITY"`
	TotalCapacity string `json:"USERTOTALCAPACITY" metrics:"USERTOTALCAPACITY"`
	CapacityUsage string `json:"USERCONSUMEDCAPACITYPERCENTAGE" metrics:"USERCONSUMEDCAPACITYPERCENTAGE"`
}

// FileSystemObject filesystem object information
type FileSystemObject struct {
	Id                             string `json:"ID" metrics:"ID"`
	Name                           string `json:"NAME" metrics:"NAME"`
	Capacity                       string `json:"CAPACITY" metrics:"CAPACITY"`
	AllocCapacity                  string `json:"ALLOCCAPACITY" metrics:"ALLOCCAPACITY"`
	AllocatedPoolQuota             string `json:"allocatedPoolQuota" metrics:"allocatedPoolQuota"`
	AvailableAndAllocCapacityRatio string `json:"AVAILABLEANDALLOCCAPACITYRATIO" metrics:"AVAILABLEANDALLOCCAPACITYRATIO"`
}
