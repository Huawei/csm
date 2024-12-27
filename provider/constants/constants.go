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

// Package constants is a package that provide global variable
package constants

const (
	// OceanStorage is a storage type oceanStorage.
	OceanStorage = "oceanStorage"

	// Object is a metrics type object.
	Object = "object"

	// Performance is a metric type performance.
	Performance = "performance"

	// Array is a collect type array.
	Array = "array"

	// Controller is a collect type controller.
	Controller = "controller"

	// StoragePool is a collect type storagePool.
	StoragePool = "storagepool"

	// Lun is a collect type lun.
	Lun = "lun"

	// Filesystem is a collect type filesystem.
	Filesystem = "filesystem"

	// NasVolume is a volume type nas
	NasVolume = "nas"

	// LunVolume is a volume type lun
	LunVolume = "lun"

	// ResourceTypeLun is a resource type means lun
	ResourceTypeLun = "11"

	// ResourceTypeFilesystem is a resource type means filesystem
	ResourceTypeFilesystem = "40"

	// PersistentVolumeKind is a resource kind PersistentVolume
	PersistentVolumeKind = "PersistentVolume"

	// PodKind is a resource kind Pod
	PodKind = "Pod"

	// DefaultNameSpace default namespace
	DefaultNameSpace = "default"

	// StorageNas is a storage volume type oceanstor-nas
	StorageNas = "oceanstor-nas"

	// StorageSan is a storage volume type oceanstor-san
	StorageSan = "oceanstor-san"

	// ObjectId is the field objectId
	ObjectId = "ObjectId"

	// ObjectName is the field ObjectName
	ObjectName = "ObjectName"

	// MinVersionSupportPost post request to get performance data is supported since version 6.1.2
	MinVersionSupportPost = "6.1.2"

	// StorageV6PointReleasePrefix defines the number of storage version which supported point version
	StorageV6PointReleasePrefix = "6"
)
