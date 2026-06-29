/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2026. All rights reserved.
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

package collect

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestFileSystemObject_SnapshotUseCapacity(t *testing.T) {
	const sectorsTOGb = 1024 * 1024 * 2

	// arrange
	testCases := []struct {
		name                  string
		jsonInput             string
		expectedSnapshotCap   string
		expectedSnapshotCapGB float64
	}{
		{
			name: "1 GB snapshot capacity",
			jsonInput: `{
				"ID": "1",
				"NAME": "test_fs",
				"CAPACITY": "10485760",
				"SNAPSHOTUSECAPACITY": "2097152"
			}`,
			expectedSnapshotCap:   "2097152",
			expectedSnapshotCapGB: 1.0,
		},
		{
			name: "500 MB snapshot capacity",
			jsonInput: `{
				"ID": "2",
				"NAME": "test_fs_2",
				"CAPACITY": "10485760",
				"SNAPSHOTUSECAPACITY": "1048576"
			}`,
			expectedSnapshotCap:   "1048576",
			expectedSnapshotCapGB: 0.5,
		},
		{
			name: "Zero snapshot capacity",
			jsonInput: `{
				"ID": "3",
				"NAME": "test_fs_3",
				"CAPACITY": "10485760",
				"SNAPSHOTUSECAPACITY": "0"
			}`,
			expectedSnapshotCap:   "0",
			expectedSnapshotCapGB: 0.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			var fsObj FileSystemObject
			err := json.Unmarshal([]byte(tc.jsonInput), &fsObj)
			if err != nil {
				t.Fatalf("Failed to unmarshal FileSystemObject: %v", err)
			}

			// assert
			if fsObj.SnapshotUsedCapacity != tc.expectedSnapshotCap {
				t.Errorf("SnapshotUsedCapacity = %s, want %s", fsObj.SnapshotUsedCapacity, tc.expectedSnapshotCap)
			}
			snapshotUsedCapacity, err := strconv.ParseFloat(fsObj.SnapshotUsedCapacity, 64)
			if err != nil {
				t.Errorf("Failed to convert snapshot capacity to float64: %v", err)
			}
			snapshotUsedCapacityGB := snapshotUsedCapacity / sectorsTOGb
			if snapshotUsedCapacityGB != tc.expectedSnapshotCapGB {
				t.Errorf("SnapshotUseCapacityGB = %f, want %f", snapshotUsedCapacityGB, tc.expectedSnapshotCapGB)
			}
		})
	}
}
