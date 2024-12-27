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

// Package collector includes all huawei storage collectors to gather and export huawei storage metrics.
// In this file we write shared parsing methods
package collector

import "strconv"

const (
	healthStatusToPrometheus  = "health_status"
	healthStatusFromStorage   = "HEALTHSTATUS"
	runningStatusToPrometheus = "running_status"
	runningStatusFromStorage  = "RUNNINGSTATUS"
	sectorsTOGb               = 1024 * 1024 * 2
	capacityKey               = "CAPACITY"
	allocCapacityKey          = "ALLOCCAPACITY"
	calculatePercentage       = 100
	bitSize                   = 64
	precisionOfTwo            = 2
	precisionOfFour           = 4
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

func parseStorageReturnZero(inDataKey, metricsName string, inData map[string]string) string {
	return "0.0"
}

func parseStorageStatus(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	if metricsName == healthStatusToPrometheus {
		return StorageHealthStatus[inData[healthStatusFromStorage]]
	}
	if metricsName == runningStatusToPrometheus {
		return StorageRunningStatus[inData[runningStatusFromStorage]]
	}
	return ""
}

func parseStorageSectorsToGB(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	sectorsData, err := strconv.ParseFloat(inData[inDataKey], bitSize)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(sectorsData/sectorsTOGb, 'f', precisionOfFour, bitSize)
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

func parseCapacityUsage(inDataKey, metricsName string, inData map[string]string) string {
	if len(inData) == 0 {
		return ""
	}
	capacity, err := strconv.ParseFloat(inData[capacityKey], bitSize)
	if err != nil || capacity == 0 {
		return ""
	}
	allocCapacity, err := strconv.ParseFloat(inData[allocCapacityKey], bitSize)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(allocCapacity/capacity*calculatePercentage, 'f', precisionOfTwo, bitSize)
}
