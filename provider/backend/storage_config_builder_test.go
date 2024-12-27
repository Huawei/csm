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
)

func TestStorageBackendConfigBuilder_WithSbcInfo_ErrExisted(t *testing.T) {
	// arrange
	wantErr := errors.New("existed err")
	builder := &StorageBackendConfigBuilder{
		ctx: context.Background(),
		err: wantErr,
	}

	// act
	getRes := builder.WithSbcInfo()

	// assert
	if !reflect.DeepEqual(wantErr, getRes.err) {
		t.Errorf("TestStorageBackendConfigBuilder_WithSbcInfo_ErrExisted failed, wantErr = %v, gotErr = %v",
			wantErr, getRes.err)
	}
}
