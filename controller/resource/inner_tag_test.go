/*
 Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

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

// Package resource
package resource

import (
	"fmt"
	"reflect"
	"testing"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewInnerTag_PvV1(t *testing.T) {
	// arrange
	meta := metaV1.TypeMeta{Kind: persistentVolumeV1Kind, APIVersion: persistentVolumeV1ApiVersion}
	want := &PersistentVolumeV1Tag{}

	// act
	got, err := NewInnerTag(meta)

	// assert
	if err != nil {
		t.Errorf("TestNewInnerTag_PvV1 failed: [%v]", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestNewInnerTag_PvV1 failed: want [%v], got [%v]", want, got)
	}
}

func TestNewInnerTag_PodV1(t *testing.T) {
	// arrange
	meta := metaV1.TypeMeta{Kind: podV1Kind, APIVersion: podV1ApiVersion}
	want := &PodV1Tag{}

	// act
	got, err := NewInnerTag(meta)

	// assert
	if err != nil {
		t.Errorf("TestNewInnerTag_PodV1 failed: [%v]", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestNewInnerTag_PodV1 failed: want [%v], got [%v]", want, got)
	}
}

func TestNewInnerTag_UnSupportedType(t *testing.T) {
	// arrange
	meta := metaV1.TypeMeta{Kind: "fakeKind", APIVersion: "fakeAPIVersion"}
	want := fmt.Errorf("unsupported tag type [%v]", meta)

	// act
	_, got := NewInnerTag(meta)

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestNewInnerTag_UnSupportedType failed: want [%v], got [%v]", want, got)
	}
}
