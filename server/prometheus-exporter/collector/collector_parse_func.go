/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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
// In this file we write shared parsing methods
package collector

import "strconv"

const (
	sectorsTOGb                = 1024 * 1024 * 2
	byteToGb                   = 1024 * 1024 * 1024
	mbToGBDivisor              = 1024
	kbToGBDivisor              = 1024 * 1024
	calculatePercentage        = 100
	bitSize                    = 64
	unlimitedPrecision         = -1
	precisionOfSmallest        = -1
	skipReportValue            = "skipReportValue"
	storageTypeSan             = "oceanstor-san"
	storageTypeNas             = "oceanstor-nas"
	storageTypeFusionNas       = "fusionstorage-nas"
	storageTypeKey             = "sbcStorageType"
	filesystemCapacityKey      = "CAPACITY"
	filesystemUsedCapacityKey  = "allocatedPoolQuota"
	snapshotReserveCapacityKey = "SNAPSHOTRESERVECAPACITY"
)

type metricsParseFunc func(inDataKey, metricsName string, inData map[string]string) string

type parseRelation struct {
	parseKey  string
	parseFunc metricsParseFunc
}

func parseStorageData(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	return inData[inDataKey]
}

func parseLabelListToLabelValueSlice(labelKeys []string,
	labelParseRelation map[string]parseRelation, metricsName string, inData map[string]string) []string {
	var labelValueSlice []string
	for _, labelName := range labelKeys {
		parseRelationData, exist := labelParseRelation[labelName]
		if !exist {
			labelValueSlice = append(labelValueSlice, "")
			continue
		}
		labelValue := parseRelationData.parseFunc(
			parseRelationData.parseKey, metricsName, inData)
		labelValueSlice = append(labelValueSlice, labelValue)
	}
	return labelValueSlice
}

func parseStorageSectorsToGB(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	sectorsData, err := strconv.ParseFloat(inData[inDataKey], bitSize)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(sectorsData/sectorsTOGb, 'f', precisionOfSmallest, bitSize)
}

func parseLunCapacityUsage(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	capacity, err := strconv.ParseFloat(inData["CAPACITY"], bitSize)
	if err != nil || capacity == 0 {
		return ""
	}
	allocCapacity, err := strconv.ParseFloat(inData["ALLOCCAPACITY"], bitSize)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(allocCapacity/capacity*calculatePercentage, 'f', unlimitedPrecision, bitSize)
}

func parseFilesystemCapacityUsage(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	capacity, err := strconv.ParseFloat(inData[filesystemCapacityKey], bitSize)
	if err != nil || capacity == 0 {
		return ""
	}
	usedCapacity, err := strconv.ParseFloat(inData[filesystemUsedCapacityKey], bitSize)
	if err != nil {
		return ""
	}
	snapshotReserveCapacity, err := strconv.ParseFloat(inData[snapshotReserveCapacityKey], bitSize)
	if err != nil {
		return ""
	}

	if snapshotReserveCapacity >= capacity {
		return ""
	}

	return strconv.FormatFloat(usedCapacity/(capacity-snapshotReserveCapacity)*calculatePercentage, 'f',
		unlimitedPrecision, bitSize)
}

// parsePVCapacity dispatches capacity parsing by storage type.
// LUN uses sector-based CAPACITY, filesystem uses CAPACITY in GB,
// namespace uses SPACE_HARD_QUOTA in MB (converted to GB by /1024).
func parsePVCapacity(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	pvType, ok := inData[storageTypeKey]
	if !ok {
		return ""
	}
	switch pvType {
	case storageTypeSan:
		return parseStorageSectorsToGB(inDataKey, metricsName, inData)
	case storageTypeNas:
		return parseStorageSectorsToGB(inDataKey, metricsName, inData)
	case storageTypeFusionNas:
		// Namespace SPACE_HARD_QUOTA is in KB, convert to GB
		val, err := strconv.ParseFloat(inData["SPACE_HARD_QUOTA"], bitSize)
		if err != nil {
			return ""
		}
		return strconv.FormatFloat(val/kbToGBDivisor, 'f', unlimitedPrecision, bitSize)
	default:
		return parseStorageSectorsToGB(inDataKey, metricsName, inData)
	}
}
