/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

// Package backend is a package that manager storage backend
package backend

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/storage/constant"
)

func TestClientInfoBuilder_WithVolumeType_Success(t *testing.T) {
	// arrange
	builder := &ClientInfoBuilder{
		ctx:        context.Background(),
		clientInfo: &ClientInfo{},
	}

	// act
	getRes := builder.WithVolumeType(constants.StorageNas)

	// assert
	if getRes.clientInfo.VolumeType != constants.NasVolume {
		t.Errorf("TestClientInfoBuilder_WithVolumeType_Success failed, want = %s, got = %s",
			constants.NasVolume, getRes.clientInfo.VolumeType)
	}
}

func TestClientInfoBuilder_WithVolumeType_UnsupportedErr(t *testing.T) {
	// arrange
	wantErr := errors.New("illegalArgumentError unsupported storage type")
	builder := &ClientInfoBuilder{
		ctx:        context.Background(),
		clientInfo: &ClientInfo{},
	}

	// act
	getRes := builder.WithVolumeType("noneType")

	// assert
	if !reflect.DeepEqual(wantErr, getRes.err) {
		t.Errorf("TestClientInfoBuilder_WithVolumeType_UnsupportedErr failed, wantErr = %v, gotErr = %v",
			wantErr, getRes.err)
	}
}

func TestClientInfoBuilder_WithVolumeType_ErrExisted(t *testing.T) {
	// arrange
	wantErr := errors.New("existed err")
	builder := &ClientInfoBuilder{
		ctx:        context.Background(),
		clientInfo: &ClientInfo{},
		err:        wantErr,
	}

	// act
	getRes := builder.WithVolumeType(constants.StorageNas)

	// assert
	if !reflect.DeepEqual(wantErr, getRes.err) {
		t.Errorf("TestClientInfoBuilder_WithVolumeType_ErrExisted failed, wantErr = %v, gotErr = %v",
			wantErr, getRes.err)
	}
}

func TestClientInfoBuilder_WithClient_ErrExisted(t *testing.T) {
	// arrange
	wantErr := errors.New("existed err")
	builder := &ClientInfoBuilder{
		ctx:        context.Background(),
		clientInfo: &ClientInfo{},
		err:        wantErr,
	}

	// act
	getRes := builder.WithClient(&constant.StorageBackendConfig{})

	// assert
	if !reflect.DeepEqual(wantErr, getRes.err) {
		t.Errorf("TestClientInfoBuilder_WithClient_ErrExisted failed, wantErr = %v, gotErr = %v",
			wantErr, getRes.err)
	}
}
