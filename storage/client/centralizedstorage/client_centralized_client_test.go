/*
 Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.

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

package centralizedstorage

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"github.com/huawei/csm/v2/storage/client"
	"github.com/huawei/csm/v2/storage/constant"
)

var testCtx = context.Background()

func TestCentralizedClient_NewCentralizedClient_Success(t *testing.T) {
	// arrange
	config := &constant.StorageBackendConfig{
		Urls:                    []string{"https://127.0.0.1:8080"},
		User:                    "admin",
		SecretNamespace:         "default",
		SecretName:              "test-secret",
		StorageBackendNamespace: "huawei-csi",
		StorageBackendName:      "test-backend",
		ClientMaxThreads:        10,
	}

	// mock InitHttpClient to succeed - use ApplyMethod with function
	patches := gomonkey.ApplyMethod(reflect.TypeOf(&client.Client{}), "InitHttpClient",
		func(_ *client.Client, ctx context.Context) error {
			return nil
		})
	defer patches.Reset()

	// action
	gotClient, gotErr := NewCentralizedClient(testCtx, config)

	// assert
	assert.Nil(t, gotErr)
	assert.NotNil(t, gotClient)
	assert.Equal(t, "admin", gotClient.User)
	assert.Equal(t, 10, gotClient.Semaphore.AvailablePermits())
}

func TestCentralizedClient_NewCentralizedClient_InitHttpClientFailed(t *testing.T) {
	// arrange
	config := &constant.StorageBackendConfig{
		Urls:                    []string{"https://127.0.0.1:8080"},
		User:                    "admin",
		SecretNamespace:         "default",
		SecretName:              "test-secret",
		StorageBackendNamespace: "huawei-csi",
		StorageBackendName:      "test-backend",
		ClientMaxThreads:        10,
	}
	wantErr := errors.New("init http client failed")

	// mock
	patches := gomonkey.ApplyMethod(reflect.TypeOf(&client.Client{}), "InitHttpClient",
		func(_ *client.Client, ctx context.Context) error {
			return wantErr
		})
	defer patches.Reset()

	// action
	gotClient, gotErr := NewCentralizedClient(testCtx, config)

	// assert
	assert.Nil(t, gotClient)
	assert.Equal(t, wantErr, gotErr)
}
