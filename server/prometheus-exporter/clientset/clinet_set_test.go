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

package clientset

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"google.golang.org/grpc"

	storageGRPC "github.com/huawei/csm/v2/grpc/lib/go/cmi"
)

func TestInitExporterClientSet(t *testing.T) {
	// arrange
	want := &ClientsSet{}

	// mock
	patches := gomonkey.
		ApplyFunc(storageGRPC.GetClientSet, func(address string) (*ClientsSet, error) {
			return nil, nil
		}).ApplyFunc(initKubeClientAndSbcClient, func() { return })
	defer patches.Reset()

	// action
	got := InitExporterClientSet("fake_data")

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetExporterClientSet() got = %v, want %v", got, want)
	}
}

func TestDeleteExporterClientSet(t *testing.T) {
	// array
	called := false

	// mock
	patches := gomonkey.
		ApplyGlobalVar(&exporterClientSet, &ClientsSet{
			StorageGRPCClientSet: &storageGRPC.ClientSet{Conn: &grpc.ClientConn{}}}).
		ApplyMethodFunc(exporterClientSet.StorageGRPCClientSet.Conn, "Close", func() error {
			called = true
			return nil
		})
	defer patches.Reset()

	// action
	DeleteExporterClientSet()

	// assert
	if called != true {
		t.Errorf("DeleteExporterClientSet() called = %v, want true", called)
	}
}
