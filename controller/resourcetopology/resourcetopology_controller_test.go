/*
 Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

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

// Package resourcetopology defines to reconcile action of resources topologies
package resourcetopology

import (
	"reflect"
	"testing"

	"github.com/Huawei/eSDK_K8S_Plugin/v4/pkg/client/clientset/versioned/fake"
	sbcInformers "github.com/Huawei/eSDK_K8S_Plugin/v4/pkg/client/informers/externalversions"
	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	controllerConfig "github.com/huawei/csm/v2/config/topology"
)

func TestController_enqueuePersistentVolume_HasEmptyDTreeParentName(t *testing.T) {
	// arrange
	ctrl := &Controller{volumeQueue: workqueue.NewRateLimitingQueue(workqueue.DefaultItemBasedRateLimiter())}
	pv := &v1.PersistentVolume{
		Spec: v1.PersistentVolumeSpec{PersistentVolumeSource: v1.PersistentVolumeSource{
			CSI: &v1.CSIPersistentVolumeSource{
				Driver:           controllerConfig.GetCSIDriverName(),
				VolumeAttributes: map[string]string{"dTreeParentName": ""},
			},
		}},
	}
	count := 0

	// mock
	mock := gomonkey.ApplyMethodFunc(reflect.TypeOf(ctrl.volumeQueue), "Add", func(item interface{}) {
		count++
	})
	mock = mock.ApplyPrivateMethod(reflect.TypeOf(ctrl), "verifyPersistentVolumeValid",
		func(pv *v1.PersistentVolume) error {
			return nil
		})
	mock = mock.ApplyFunc(cache.DeletionHandlingMetaNamespaceKeyFunc,
		func(obj interface{}) (string, error) {
			return "fakeObjName", nil
		})

	// act
	ctrl.enqueuePersistentVolume(pv)

	// assert
	if count != 1 {
		t.Errorf("TestController_enqueuePersistentVolume_HasEmptyDTreeParentName failed, "+
			"want add 1 time, actually add %d", count)
	}

	// clean up
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestController_enqueuePersistentVolume_HasNotEmptyDTreeParentName(t *testing.T) {
	// arrange
	clientSet := fake.NewSimpleClientset()
	sbcFactory := sbcInformers.NewSharedInformerFactory(clientSet, 0)
	ctrl := &Controller{volumeQueue: workqueue.NewRateLimitingQueue(workqueue.DefaultItemBasedRateLimiter()),
		backendInformer: sbcFactory.Xuanwu().V1().StorageBackendClaims()}
	pv := &v1.PersistentVolume{
		Spec: v1.PersistentVolumeSpec{PersistentVolumeSource: v1.PersistentVolumeSource{
			CSI: &v1.CSIPersistentVolumeSource{
				Driver:           controllerConfig.GetCSIDriverName(),
				VolumeAttributes: map[string]string{"dTreeParentName": "dTree"},
			},
		}},
	}
	count := 0

	// mock
	mock := gomonkey.ApplyMethodFunc(reflect.TypeOf(ctrl.volumeQueue), "Add", func(item interface{}) {
		count++
	})

	// act
	ctrl.enqueuePersistentVolume(pv)

	// assert
	if count != 0 {
		t.Errorf("TestController_enqueuePersistentVolume_HasNotEmptyDTreeParentName failed, "+
			"want add 0 time, actually add %d", count)
	}

	// clean up
	t.Cleanup(func() {
		mock.Reset()
	})
}

func TestController_enqueuePersistentVolume_EmptyVolumeAttributes(t *testing.T) {
	// arrange
	clientSet := fake.NewSimpleClientset()
	sbcFactory := sbcInformers.NewSharedInformerFactory(clientSet, 0)
	ctrl := &Controller{volumeQueue: workqueue.NewRateLimitingQueue(workqueue.DefaultItemBasedRateLimiter()),
		backendInformer: sbcFactory.Xuanwu().V1().StorageBackendClaims()}
	pv := &v1.PersistentVolume{
		Spec: v1.PersistentVolumeSpec{PersistentVolumeSource: v1.PersistentVolumeSource{
			CSI: &v1.CSIPersistentVolumeSource{
				Driver: controllerConfig.GetCSIDriverName(),
			},
		}},
	}
	count := 0

	// mock
	mock := gomonkey.ApplyMethodFunc(reflect.TypeOf(ctrl.volumeQueue), "Add", func(item interface{}) {
		count++
	})

	// act
	ctrl.enqueuePersistentVolume(pv)

	// assert
	if count != 0 {
		t.Errorf("TestController_enqueuePersistentVolume_EmptyVolumeAttributes failed, "+
			"want add 0 time, actually add %d", count)
	}

	// clean up
	t.Cleanup(func() {
		mock.Reset()
	})
}
