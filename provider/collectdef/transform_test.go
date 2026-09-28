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
	"testing"
)

func TestReturnZero(t *testing.T) {
	got := ReturnZero(map[string]string{"any": "value"})
	want := "0.0"
	if got != want {
		t.Errorf("ReturnZero() = %q, want %q", got, want)
	}
}

func TestSectorsToGBOf(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "exact 1 GB",
			data: map[string]string{"CAPACITY": "2097152"},
			want: "1",
		},
		{
			name: "10 GB",
			data: map[string]string{"CAPACITY": "20971520"},
			want: "10",
		},
		{
			name: "fractional value",
			data: map[string]string{"CAPACITY": "3145728"},
			want: "1.5",
		},
		{
			name: "invalid number",
			data: map[string]string{"CAPACITY": "abc"},
			want: "",
		},
		{
			name: "missing key",
			data: map[string]string{"OTHER": "2097152"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := SectorsToGBOf("CAPACITY")
			got := fn(tt.data)
			if got != tt.want {
				t.Errorf("SectorsToGBOf(%q) = %q, want %q", "CAPACITY", got, tt.want)
			}
		})
	}
}

func TestVstoreBytesToGBOf(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "exact 1 GB",
			data: map[string]string{"TotalCapacity": "1073741824"},
			want: "1",
		},
		{
			name: "2 GB",
			data: map[string]string{"TotalCapacity": "2147483648"},
			want: "2",
		},
		{
			name: "invalid number",
			data: map[string]string{"TotalCapacity": "notanumber"},
			want: "",
		},
		{
			name: "missing key",
			data: map[string]string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := VstoreBytesToGBOf("TotalCapacity")
			got := fn(tt.data)
			if got != tt.want {
				t.Errorf("VstoreBytesToGBOf(%q) = %q, want %q", "TotalCapacity", got, tt.want)
			}
		})
	}
}

func TestHealthStatus(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "Normal status",
			data: map[string]string{"HEALTHSTATUS": "1"},
			want: "Normal",
		},
		{
			name: "Fault status",
			data: map[string]string{"HEALTHSTATUS": "2"},
			want: "Fault",
		},
		{
			name: "unknown code returns empty",
			data: map[string]string{"HEALTHSTATUS": "999"},
			want: "",
		},
		{
			name: "missing key returns empty",
			data: map[string]string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HealthStatus(tt.data)
			if got != tt.want {
				t.Errorf("HealthStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunningStatus(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "Normal status",
			data: map[string]string{"RUNNINGSTATUS": "1"},
			want: "Normal",
		},
		{
			name: "Online status",
			data: map[string]string{"RUNNINGSTATUS": "27"},
			want: "Online",
		},
		{
			name: "unknown code returns empty",
			data: map[string]string{"RUNNINGSTATUS": "999"},
			want: "",
		},
		{
			name: "missing key returns empty",
			data: map[string]string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RunningStatus(tt.data)
			if got != tt.want {
				t.Errorf("RunningStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCapacityUsageOf(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		usedKey  string
		totalKey string
		want     string
	}{
		{
			name:     "50 percent usage",
			data:     map[string]string{"USED": "50", "TOTAL": "100"},
			usedKey:  "USED",
			totalKey: "TOTAL",
			want:     "50",
		},
		{
			name:     "75 percent usage",
			data:     map[string]string{"USED": "75", "TOTAL": "100"},
			usedKey:  "USED",
			totalKey: "TOTAL",
			want:     "75",
		},
		{
			name:     "zero total returns empty",
			data:     map[string]string{"USED": "50", "TOTAL": "0"},
			usedKey:  "USED",
			totalKey: "TOTAL",
			want:     "",
		},
		{
			name:     "invalid total returns empty",
			data:     map[string]string{"USED": "50", "TOTAL": "abc"},
			usedKey:  "USED",
			totalKey: "TOTAL",
			want:     "",
		},
		{
			name:     "invalid used returns empty",
			data:     map[string]string{"USED": "abc", "TOTAL": "100"},
			usedKey:  "USED",
			totalKey: "TOTAL",
			want:     "",
		},
		{
			name:     "missing total returns empty",
			data:     map[string]string{"USED": "50"},
			usedKey:  "USED",
			totalKey: "TOTAL",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := CapacityUsageOf(tt.usedKey, tt.totalKey)
			got := fn(tt.data)
			if got != tt.want {
				t.Errorf("CapacityUsageOf(%q,%q) = %q, want %q", tt.usedKey, tt.totalKey, got, tt.want)
			}
		})
	}
}

func TestArrayModel(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "prefers productModeString",
			data: map[string]string{"productModeString": "MyModel", "PRODUCTMODE": "811"},
			want: "MyModel",
		},
		{
			name: "falls back to PRODUCTMODE map lookup",
			data: map[string]string{"PRODUCTMODE": "811"},
			want: "OceanStor Dorado 5000 V6",
		},
		{
			name: "unknown PRODUCTMODE returns empty from map",
			data: map[string]string{"PRODUCTMODE": "999"},
			want: "",
		},
		{
			name: "both missing returns empty",
			data: map[string]string{},
			want: "",
		},
		{
			name: "productModeString empty falls back",
			data: map[string]string{"productModeString": "", "PRODUCTMODE": "61"},
			want: "6800 V3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ArrayModel(tt.data)
			if got != tt.want {
				t.Errorf("ArrayModel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArrayVersion(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "prefers SoftwareVersion",
			data: map[string]string{"SoftwareVersion": "V6.1.0", "PRODUCTVERSION": "V5.0.0"},
			want: "V6.1.0",
		},
		{
			name: "falls back to PRODUCTVERSION",
			data: map[string]string{"PRODUCTVERSION": "V5.0.0"},
			want: "V5.0.0",
		},
		{
			name: "both missing returns empty",
			data: map[string]string{},
			want: "",
		},
		{
			name: "SoftwareVersion empty falls back",
			data: map[string]string{"SoftwareVersion": "", "PRODUCTVERSION": "V6.1.2"},
			want: "V6.1.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ArrayVersion(tt.data)
			if got != tt.want {
				t.Errorf("ArrayVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLunCapacityUsage(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "50 percent",
			data: map[string]string{"ALLOCCAPACITY": "50", "CAPACITY": "100"},
			want: "50",
		},
		{
			name: "zero capacity returns empty",
			data: map[string]string{"ALLOCCAPACITY": "50", "CAPACITY": "0"},
			want: "",
		},
		{
			name: "missing fields returns empty",
			data: map[string]string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LunCapacityUsage(tt.data)
			if got != tt.want {
				t.Errorf("LunCapacityUsage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStoragePoolCapacityUsage(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "reads pre-computed percentage",
			data: map[string]string{"USERCONSUMEDCAPACITYPERCENTAGE": "42.5"},
			want: "42.5",
		},
		{
			name: "missing key returns empty",
			data: map[string]string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StoragePoolCapacityUsage(tt.data)
			if got != tt.want {
				t.Errorf("StoragePoolCapacityUsage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFilesystemCapacityUsage(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "basic calculation with no snapshot reserve",
			data: map[string]string{"CAPACITY": "100", "allocatedPoolQuota": "50", "SNAPSHOTRESERVECAPACITY": "0"},
			want: "50",
		},
		{
			name: "calculation with snapshot reserve",
			data: map[string]string{"CAPACITY": "100", "allocatedPoolQuota": "60", "SNAPSHOTRESERVECAPACITY": "20"},
			want: "75",
		},
		{
			name: "zero capacity returns empty",
			data: map[string]string{"CAPACITY": "0", "allocatedPoolQuota": "50", "SNAPSHOTRESERVECAPACITY": "0"},
			want: "",
		},
		{
			name: "capacity equals snapshot reserve returns empty",
			data: map[string]string{"CAPACITY": "100", "allocatedPoolQuota": "50", "SNAPSHOTRESERVECAPACITY": "100"},
			want: "",
		},
		{
			name: "invalid capacity returns empty",
			data: map[string]string{"CAPACITY": "abc", "allocatedPoolQuota": "50", "SNAPSHOTRESERVECAPACITY": "0"},
			want: "",
		},
		{
			name: "invalid allocatedPoolQuota returns empty",
			data: map[string]string{"CAPACITY": "100", "allocatedPoolQuota": "abc", "SNAPSHOTRESERVECAPACITY": "0"},
			want: "",
		},
		{
			name: "missing snapshot reserve",
			data: map[string]string{"CAPACITY": "100", "allocatedPoolQuota": "25"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilesystemCapacityUsage(tt.data)
			if got != tt.want {
				t.Errorf("FilesystemCapacityUsage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKBToGBOf(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "3072000 KB to GB",
			data: map[string]string{"SPACE_USED": "3072000"},
			want: "2.9296875",
		},
		{
			name: "exact 1 GB",
			data: map[string]string{"SPACE_USED": "1048576"},
			want: "1",
		},
		{
			name: "invalid number",
			data: map[string]string{"SPACE_USED": "abc"},
			want: "",
		},
		{
			name: "missing key",
			data: map[string]string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := KBToGBOf("SPACE_USED")
			got := fn(tt.data)
			if got != tt.want {
				t.Errorf("KBToGBOf(%q) = %q, want %q", "SPACE_USED", got, tt.want)
			}
		})
	}
}

func TestMBToGBOf(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "10485760 MB to GB",
			data: map[string]string{"SPACE_HARD_QUOTA": "10485760"},
			want: "10240",
		},
		{
			name: "exact 1 GB",
			data: map[string]string{"SPACE_HARD_QUOTA": "1024"},
			want: "1",
		},
		{
			name: "invalid number",
			data: map[string]string{"SPACE_HARD_QUOTA": "abc"},
			want: "",
		},
		{
			name: "missing key",
			data: map[string]string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := MBToGBOf("SPACE_HARD_QUOTA")
			got := fn(tt.data)
			if got != tt.want {
				t.Errorf("MBToGBOf(%q) = %q, want %q", "SPACE_HARD_QUOTA", got, tt.want)
			}
		})
	}
}

func TestDirectValueOf(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "returns value as-is",
			data: map[string]string{"SPACE_USED_RATE": "29"},
			want: "29",
		},
		{
			name: "missing key returns empty",
			data: map[string]string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := DirectValueOf("SPACE_USED_RATE")
			got := fn(tt.data)
			if got != tt.want {
				t.Errorf("DirectValueOf(%q) = %q, want %q", "SPACE_USED_RATE", got, tt.want)
			}
		})
	}
}

func TestVstoreCapacityUsage(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		want string
	}{
		{
			name: "50 percent",
			data: map[string]string{"UsedCapacity": "50", "TotalCapacity": "100"},
			want: "50",
		},
		{
			name: "zero total returns empty",
			data: map[string]string{"UsedCapacity": "50", "TotalCapacity": "0"},
			want: "",
		},
		{
			name: "missing fields returns empty",
			data: map[string]string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VstoreCapacityUsage(tt.data)
			if got != tt.want {
				t.Errorf("VstoreCapacityUsage() = %q, want %q", got, tt.want)
			}
		})
	}
}
