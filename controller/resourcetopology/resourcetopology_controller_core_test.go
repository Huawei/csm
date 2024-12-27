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

// Package resourcetopology
package resourcetopology

import (
	"context"
	"reflect"
	"testing"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiXuanwuV1 "github.com/huawei/csm/v2/client/apis/xuanwu/v1"
	fakeXuanwuClient "github.com/huawei/csm/v2/pkg/client/clientset/versioned/fake"
)

func TestResourceTopologyController_UpdateResourceTopologiesStatus_Success(t *testing.T) {
	// arrange
	ctrl := &Controller{}
	ctx := context.TODO()
	rt := &apiXuanwuV1.ResourceTopology{}
	want := &apiXuanwuV1.ResourceTopology{Status: apiXuanwuV1.ResourceTopologyStatus{
		Status: apiXuanwuV1.ResourceTopologyStatusNormal,
	}}

	// mock
	ctrl.xuanwuClient = fakeXuanwuClient.NewSimpleClientset()

	// expect
	ctrl.xuanwuClient.XuanwuV1().ResourceTopologies().Create(ctx, rt, metaV1.CreateOptions{})

	// act
	got, err := ctrl.UpdateResourceTopologiesStatus(ctx, want)

	// assert
	if err != nil {
		t.Errorf("TestResourceTopologyController_UpdateResourceTopologiesStatus_Success failed: [%v]", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestResourceTopologyController_UpdateResourceTopologiesStatus_Success failed: "+
			"want: [%v],got: [%v]", want, got)
	}
}

func TestResourceTopologyController_UpdateResourceTopologiesStatus_Failed(t *testing.T) {
	// arrange
	ctrl := &Controller{}
	ctx := context.TODO()
	fakeRt := &apiXuanwuV1.ResourceTopology{Status: apiXuanwuV1.ResourceTopologyStatus{
		Status: apiXuanwuV1.ResourceTopologyStatusNormal,
	}}
	want := "resourcetopologies.xuanwu.huawei.io \"\" not found"

	// mock
	ctrl.xuanwuClient = fakeXuanwuClient.NewSimpleClientset()

	// act
	_, got := ctrl.UpdateResourceTopologiesStatus(ctx, fakeRt)

	// assert
	if got == nil || got.Error() != want {
		t.Errorf("TestResourceTopologyController_UpdateResourceTopologiesStatus_Success failed: "+
			"wantErr: [%v], gotErr: [%v]", want, got)
	}
}
