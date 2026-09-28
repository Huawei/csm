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
	"strconv"

	"github.com/huawei/csm/v2/provider/constants"
)

const (
	sectorsToGBDivisor   = 1024 * 1024 * 2
	bytesToGBDivisor     = 1024 * 1024 * 1024
	kbToGBDivisor        = 1024 * 1024
	mbToGBDivisor        = 1024
	percentageMultiplier = 100
)

// ReturnZero always returns "0.0". Used for info metrics like basic_info.
func ReturnZero(data map[string]string) string {
	return "0.0"
}

// SectorsToGBOf returns a TransformFunc that reads data[key] and converts sectors to GB.
func SectorsToGBOf(key string) TransformFunc {
	return func(data map[string]string) string {
		val, err := strconv.ParseFloat(data[key], constants.DefaultBitSize)
		if err != nil {
			return ""
		}
		return strconv.FormatFloat(val/sectorsToGBDivisor, 'f', constants.UnlimitedPrecision, constants.DefaultBitSize)
	}
}

// VstoreBytesToGBOf returns a TransformFunc that reads data[key] and converts bytes to GB.
func VstoreBytesToGBOf(key string) TransformFunc {
	return func(data map[string]string) string {
		val, err := strconv.ParseFloat(data[key], constants.DefaultBitSize)
		if err != nil {
			return ""
		}
		return strconv.FormatFloat(val/bytesToGBDivisor, 'f', constants.UnlimitedPrecision, constants.DefaultBitSize)
	}
}

// HealthStatus maps HEALTHSTATUS code to display text.
func HealthStatus(data map[string]string) string {
	return constants.StorageHealthStatus[data["HEALTHSTATUS"]]
}

// RunningStatus maps RUNNINGSTATUS code to display text.
func RunningStatus(data map[string]string) string {
	return constants.StorageRunningStatus[data["RUNNINGSTATUS"]]
}

// CapacityUsageOf returns a TransformFunc: (usedField / totalField) * 100.
func CapacityUsageOf(usedField, totalField string) TransformFunc {
	return func(data map[string]string) string {
		total, err := strconv.ParseFloat(data[totalField], constants.DefaultBitSize)
		if err != nil || total == 0 {
			return ""
		}
		used, err := strconv.ParseFloat(data[usedField], constants.DefaultBitSize)
		if err != nil {
			return ""
		}
		return strconv.FormatFloat(used/total*percentageMultiplier, 'f',
			constants.UnlimitedPrecision, constants.DefaultBitSize)
	}
}

// ArrayModel returns model from productModeString or PRODUCTMODE.
func ArrayModel(data map[string]string) string {
	if v := data["productModeString"]; v != "" {
		return v
	}
	if v := data["PRODUCTMODE"]; v != "" {
		return constants.StorageProductMode[v]
	}
	return ""
}

// ArrayVersion returns version from SoftwareVersion or PRODUCTVERSION.
func ArrayVersion(data map[string]string) string {
	if v := data["SoftwareVersion"]; v != "" {
		return v
	}
	if v := data["PRODUCTVERSION"]; v != "" {
		return v
	}
	return ""
}

// LunCapacityUsage computes ALLOCCAPACITY / CAPACITY * 100.
func LunCapacityUsage(data map[string]string) string {
	return CapacityUsageOf("ALLOCCAPACITY", "CAPACITY")(data)
}

// StoragePoolCapacityUsage reads the pre-computed USERCONSUMEDCAPACITYPERCENTAGE.
func StoragePoolCapacityUsage(data map[string]string) string {
	return data["USERCONSUMEDCAPACITYPERCENTAGE"]
}

// FilesystemCapacityUsage computes allocatedPoolQuota / (capacity - snapshotReserve) * 100.
func FilesystemCapacityUsage(data map[string]string) string {
	capacity, err := strconv.ParseFloat(data["CAPACITY"], constants.DefaultBitSize)
	if err != nil || capacity == 0 {
		return ""
	}

	snapshotReserve, err := strconv.ParseFloat(data["SNAPSHOTRESERVECAPACITY"], constants.DefaultBitSize)
	if err != nil {
		return ""
	}

	used, err := strconv.ParseFloat(data["allocatedPoolQuota"], constants.DefaultBitSize)
	if err != nil {
		return ""
	}

	denom := capacity - snapshotReserve
	if denom == 0 {
		return ""
	}

	return strconv.FormatFloat(used/denom*percentageMultiplier, 'f',
		constants.UnlimitedPrecision, constants.DefaultBitSize)
}

// VstoreCapacityUsage computes UsedCapacity / TotalCapacity * 100.
func VstoreCapacityUsage(data map[string]string) string {
	return CapacityUsageOf("UsedCapacity", "TotalCapacity")(data)
}

// KBToGBOf converts kilobytes to gigabytes by dividing by 1024*1024.
func KBToGBOf(key string) TransformFunc {
	return func(data map[string]string) string {
		val, err := strconv.ParseFloat(data[key], constants.DefaultBitSize)
		if err != nil {
			return ""
		}
		return strconv.FormatFloat(val/kbToGBDivisor, 'f', constants.UnlimitedPrecision, constants.DefaultBitSize)
	}
}

// MBToGBOf converts megabytes to gigabytes by dividing by 1024.
func MBToGBOf(key string) TransformFunc {
	return func(data map[string]string) string {
		val, err := strconv.ParseFloat(data[key], constants.DefaultBitSize)
		if err != nil {
			return ""
		}
		return strconv.FormatFloat(val/mbToGBDivisor, 'f', constants.UnlimitedPrecision, constants.DefaultBitSize)
	}
}

// DirectValueOf returns the value as-is without conversion.
func DirectValueOf(key string) TransformFunc {
	return func(data map[string]string) string {
		return data[key]
	}
}
