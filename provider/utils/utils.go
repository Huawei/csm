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

// Package utils is a package that provide util functions
package utils

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	coreV1 "k8s.io/api/core/v1"
)

const (
	// BackendNameMaxLength is the max length of backend name
	BackendNameMaxLength = 63

	// BackendNameUidMaxLength is the max length of backend name uid
	BackendNameUidMaxLength = 5

	dns1123LabelFmt     = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
	dns1123SubdomainFmt = dns1123LabelFmt + "(" + dns1123LabelFmt + ")*"
)

var dns1123SubdomainRegexp = regexp.MustCompile("^" + dns1123SubdomainFmt + "$")

// MapToStructSlice map a struct to slice
func MapToStructSlice[I, O any](input I) ([]O, error) {
	var targets []O
	valueType := reflect.ValueOf(input)
	if valueType.Kind() == reflect.Slice {
		return MapToStruct[I, []O](input)
	}

	target, err := MapToStruct[I, O](input)
	if err != nil {
		return []O{}, err
	}
	targets = append(targets, target)
	return targets, nil
}

// MapToStruct map to struct
func MapToStruct[I, O any](input I) (O, error) {
	var o O
	marshal, err := json.Marshal(input)
	if err != nil {
		return o, err
	}
	err = json.Unmarshal(marshal, &o)
	if err != nil {
		return o, err
	}
	return o, nil
}

// MapStringToInt map a string slice to an int slice
func MapStringToInt(sources []string) []int {
	var result []int
	for _, source := range sources {
		intVal, err := strconv.Atoi(source)
		if err != nil {
			continue
		}
		result = append(result, intVal)
	}
	return result
}

// StructToMap convert struct to map.
// input t is a struct with metrics tag
// return map key is metrics tag.
// return map value is filed value.
func StructToMap[T any](t T) map[string]string {
	mapping := map[string]string{}
	filedType := reflect.TypeOf(t)
	filedValue := reflect.ValueOf(t)
	for i := 0; i < filedType.NumField(); i++ {
		value, ok := filedValue.Field(i).Interface().(string)
		if !ok || value == "" {
			continue
		}
		mapping[filedType.Field(i).Tag.Get("metrics")] = value
	}
	return mapping
}

// CleanupSocketFile clean socket file
func CleanupSocketFile(filePath string) error {
	fileExists, err := DoesSocketExist(filePath)
	if err != nil {
		return err
	}

	if fileExists {
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to remove stale file=%s with error: %+v", filePath, err)
		}
	}
	return nil
}

// DoesSocketExist determine if the socket file exists
func DoesSocketExist(socketPath string) (bool, error) {
	if _, err := os.Lstat(socketPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to lstat the socket %s with error: %+v", socketPath, err)
	}
	return true, nil
}

// SplitVolumeId splits the volumeId to backend name and pv name
func SplitVolumeId(volumeId string) (string, string) {
	splits := strings.SplitN(volumeId, ".", 2)
	var backendName, pvName string
	if len(splits) == 2 {
		backendName, pvName = splits[0], splits[1]
	} else {
		backendName, pvName = splits[0], ""
	}
	return GetBackendName(backendName), pvName
}

// GetBackendName format the name of backend
func GetBackendName(name string) string {
	if IsDNSFormat(name) {
		return name
	}
	return BuildBackendName(name)
}

// IsDNSFormat Determine if the DNS format is met
func IsDNSFormat(source string) bool {
	if len(source) > BackendNameMaxLength {
		return false
	}
	return dns1123SubdomainRegexp.MatchString(source)
}

// BuildBackendName build backend name
func BuildBackendName(name string) string {
	nameLen := BackendNameMaxLength - BackendNameUidMaxLength - 1
	if len(name) > nameLen {
		name = name[:nameLen]
	}
	hashCode := GenerateHashCode(name, BackendNameUidMaxLength)
	mappingName := BackendNameMapping(name)
	return fmt.Sprintf("%s-%s", mappingName, hashCode)
}

// GenerateHashCode generate hash code
func GenerateHashCode(txt string, max int) string {
	hashInstance := sha256.New()
	hashInstance.Write([]byte(txt))
	sum := hashInstance.Sum(nil)
	result := fmt.Sprintf("%x", sum)
	if len(result) < max {
		return result
	}
	return result[:max]
}

// BackendNameMapping mapping backend name
func BackendNameMapping(name string) string {
	removeUnderline := strings.ReplaceAll(name, "_", "-")
	removePoint := strings.ReplaceAll(removeUnderline, ".", "-")
	return strings.ToLower(removePoint)
}

// CSIConfig holds the CSI config of backend resources
type CSIConfig struct {
	Backends map[string]interface{} `json:"backends"`
}

// ConvertConfigmapToMap formats configmap data to map struct
func ConvertConfigmapToMap(configmap *coreV1.ConfigMap) (map[string]interface{}, error) {
	if configmap.Data == nil {
		return nil, fmt.Errorf("configmap: [%s] the configmap.Data is nil", configmap.Name)
	}

	var csiConfig CSIConfig
	err := json.Unmarshal([]byte(configmap.Data["csi.json"]), &csiConfig)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal configmap.Data[\"csi.json\"] failed. err is [%v]", err)
	}

	return csiConfig.Backends, nil
}
