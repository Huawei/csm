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
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	apiXuanwuV1 "github.com/huawei/csm/v2/client/apis/xuanwu/v1"
	fakeXuanwuClient "github.com/huawei/csm/v2/pkg/client/clientset/versioned/fake"
)

const defaultBufferSize = 2048

func TestResourceTopologyController_syncResourceTopology_SuccessOnNoChange(t *testing.T) {
	// arrange
	fakeClient := fakeXuanwuClient.NewSimpleClientset()
	ctrl := &Controller{xuanwuClient: fakeClient, eventRecorder: record.NewFakeRecorder(defaultBufferSize)}
	ctx := context.TODO()
	rt := &apiXuanwuV1.ResourceTopology{ObjectMeta: metaV1.ObjectMeta{Name: "fakeResourcesTopology"},
		Spec: apiXuanwuV1.ResourceTopologySpec{Provisioner: "fakeProvisioner", VolumeHandle: "fakeVolumeHandle",
			Tags: []apiXuanwuV1.Tag{{ResourceInfo: apiXuanwuV1.ResourceInfo{TypeMeta: metaV1.TypeMeta{
				Kind: "PersistentVolume", APIVersion: "v1"}, Name: "fakePersistentVolume"}}}},
		Status: apiXuanwuV1.ResourceTopologyStatus{Tags: []apiXuanwuV1.Tag{{ResourceInfo: apiXuanwuV1.ResourceInfo{
			TypeMeta: metaV1.TypeMeta{Kind: "PersistentVolume", APIVersion: "v1"}, Name: "fakePersistentVolume"}}}}}

	// mock
	fakeClient.XuanwuV1().ResourceTopologies().Create(ctx, rt, metaV1.CreateOptions{})
	mock := gomonkey.NewPatches()
	mock.ApplyPrivateMethod(ctrl, "checkResourceTopology",
		func(ctx context.Context, resourceTopology *apiXuanwuV1.ResourceTopology) error {
			return nil
		})

	// act
	err := ctrl.syncResourceTopology(ctx, rt)

	// assert
	if err != nil {
		t.Errorf("TestResourceTopologyController_syncResourceTopology_SuccessOnNoChange failed: [%v]", err)
	}

	// cleanup
	t.Cleanup(func() {
		fakeClient.XuanwuV1().ResourceTopologies().Delete(ctx, rt.Name, metaV1.DeleteOptions{})
		mock.Reset()
	})
}

func TestResourceTopologyController_syncResourceTopology_FailOnRtNotExist(t *testing.T) {
	// arrange
	fakeClient := fakeXuanwuClient.NewSimpleClientset()
	ctrl := &Controller{xuanwuClient: fakeClient, eventRecorder: record.NewFakeRecorder(defaultBufferSize)}
	ctx := context.TODO()
	rt := &apiXuanwuV1.ResourceTopology{ObjectMeta: metaV1.ObjectMeta{Name: "fakeResourcesTopology"},
		Spec: apiXuanwuV1.ResourceTopologySpec{Provisioner: "fakeProvisioner", VolumeHandle: "fakeVolumeHandle",
			Tags: []apiXuanwuV1.Tag{{ResourceInfo: apiXuanwuV1.ResourceInfo{TypeMeta: metaV1.TypeMeta{
				Kind: "PersistentVolume", APIVersion: "v1"}, Name: "fakePersistentVolume"}}}},
		Status: apiXuanwuV1.ResourceTopologyStatus{Tags: []apiXuanwuV1.Tag{{ResourceInfo: apiXuanwuV1.ResourceInfo{
			TypeMeta: metaV1.TypeMeta{Kind: "PersistentVolume", APIVersion: "v1"}, Name: "fakePersistentVolume"}}}}}
	wantErr := errors.New("add resourceTopology [fakeResourcesTopology] finalizers failed, " +
		"errors is [resourcetopologies.xuanwu.huawei.io \"fakeResourcesTopology\" not found]")

	// act
	err := ctrl.syncResourceTopology(ctx, rt)

	// assert
	if !reflect.DeepEqual(err, wantErr) {
		t.Errorf("TestResourceTopologyController_syncResourceTopology_FailOnRtNotExist "+
			"failed: want :[%v], got: [%v]", wantErr, err)
	}
}

func TestResourceTopologyController_syncResourceTopology_StatusToNormal(t *testing.T) {
	// arrange
	fakeClient := fakeXuanwuClient.NewSimpleClientset()
	ctrl := &Controller{xuanwuClient: fakeClient, eventRecorder: record.NewFakeRecorder(defaultBufferSize)}
	ctx := context.TODO()
	rt := &apiXuanwuV1.ResourceTopology{ObjectMeta: metaV1.ObjectMeta{Name: "fakeResourcesTopology"},
		Spec: apiXuanwuV1.ResourceTopologySpec{Provisioner: "fakeProvisioner", VolumeHandle: "fakeVolumeHandle",
			Tags: []apiXuanwuV1.Tag{{ResourceInfo: apiXuanwuV1.ResourceInfo{TypeMeta: metaV1.TypeMeta{
				Kind: "PersistentVolume", APIVersion: "v1"}, Name: "fakePersistentVolume"}}}},
		Status: apiXuanwuV1.ResourceTopologyStatus{Status: apiXuanwuV1.ResourceTopologyStatusPending,
			Tags: []apiXuanwuV1.Tag{{ResourceInfo: apiXuanwuV1.ResourceInfo{TypeMeta: metaV1.TypeMeta{
				Kind: "PersistentVolume", APIVersion: "v1"}, Name: "fakePersistentVolume"}}}}}

	// mock
	rt, _ = fakeClient.XuanwuV1().ResourceTopologies().Create(ctx, rt, metaV1.CreateOptions{})
	mock := gomonkey.NewPatches()
	mock.ApplyPrivateMethod(ctrl, "checkResourceTopology",
		func(ctx context.Context, resourceTopology *apiXuanwuV1.ResourceTopology) error {
			return nil
		})

	// act
	err := ctrl.syncResourceTopology(ctx, rt)

	// assert
	if err != nil {
		t.Errorf("TestResourceTopologyController_syncResourceTopology_StatusToNormal failed: [%v]", err)
	}
	rt, _ = fakeClient.XuanwuV1().ResourceTopologies().Get(ctx, rt.Name, metaV1.GetOptions{})
	if rt.Status.Status != apiXuanwuV1.ResourceTopologyStatusNormal {
		t.Errorf("TestResourceTopologyController_syncResourceTopology_StatusToNormal failed: "+
			"want status: [%s], got status: [%s]", apiXuanwuV1.ResourceTopologyStatusNormal, rt.Status.Status)
	}

	// cleanup
	t.Cleanup(func() {
		fakeClient.XuanwuV1().ResourceTopologies().Delete(ctx, rt.Name, metaV1.DeleteOptions{})
		mock.Reset()
	})
}
