/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

// Package utils is related with storage client utils
package utils

import (
	"bytes"
	"compress/flate"
	"fmt"
)

// IsFloat64InList is used to check float64 element in list
func IsFloat64InList(list []float64, element float64) bool {
	for _, v := range list {
		if element == v {
			return true
		}
	}

	return false
}

// CleanBytes is used to clean bytes memory
func CleanBytes(bytes []byte) {
	for i := 0; i < len(bytes); i++ {
		bytes[i] = 0
	}
}

// CompressStr compress long string by deflate algorithm
// The result will be hexadecimal encoded
// The encoded result can be reverted to original str by DeCompressStr
func CompressStr(str string) (string, error) {
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, flate.BestCompression)
	defer w.Close()
	if err != nil {
		return str, err
	}
	_, err = w.Write([]byte(str))
	if err != nil {
		return str, err
	}
	err = w.Flush()
	if err != nil {
		return str, err
	}
	return fmt.Sprintf("%x", buf.Bytes()), nil
}
