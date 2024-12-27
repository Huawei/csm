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

// Package collect is a package that provides object and performance collect
package collect

import (
	"context"
	"fmt"
	"reflect"

	csiV1 "github.com/Huawei/eSDK_K8S_Plugin/v4/client/apis/xuanwu/v1"
	csiInformers "github.com/Huawei/eSDK_K8S_Plugin/v4/pkg/client/informers/externalversions"
	"k8s.io/client-go/tools/cache"

	"github.com/huawei/csm/v2/provider/backend"
	"github.com/huawei/csm/v2/provider/grpc/helper"
	"github.com/huawei/csm/v2/storage/client/centralizedstorage"
	"github.com/huawei/csm/v2/utils/log"
)

// RunBackendInformer run backend informer
func RunBackendInformer(stopCh chan struct{}) {
	factory := csiInformers.NewSharedInformerFactory(helper.GetClientSet().SbcClient, 0)
	factory.Xuanwu().V1().StorageBackendClaims().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) { updateBackendCache(oldObj, newObj) },
			DeleteFunc: func(obj interface{}) { deleteBackendCache(obj) },
		},
	)

	factory.Start(stopCh)
}

func updateBackendCache(oldObj, newObj interface{}) {
	if unknown, ok := newObj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		newObj = unknown.Obj
	}

	if unknown, ok := oldObj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		oldObj = unknown.Obj
	}

	// Check whether obj is a storageBackendClaim CR.
	oldStorageBackendClaim, ok := oldObj.(*csiV1.StorageBackendClaim)
	if !ok {
		log.Errorf("failed to convert old obj to storageBackendClaim, oldObj is [%v]", oldObj)
		return
	}

	newStorageBackendClaim, ok := newObj.(*csiV1.StorageBackendClaim)
	if !ok {
		log.Errorf("failed to convert new obj to storageBackendClaim, newObj is [%v]", newObj)
		return
	}

	if reflect.DeepEqual(newStorageBackendClaim.Spec, oldStorageBackendClaim.Spec) {
		log.Debugf("the spec struct of storageBackendClaim [%s] are not changed, "+
			"do not update backend cache", oldStorageBackendClaim.Name)
		return
	}

	err := releaseCache(newStorageBackendClaim.Name)
	if err != nil {
		log.Errorln(err)
		return
	}

	_, err = GetClient(context.Background(), newStorageBackendClaim.Name, backend.GetClientByBackendName)
	if err != nil {
		log.Errorf("get Client failed, err is [%v]", err)
		return
	}
}

func deleteBackendCache(obj interface{}) {
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}

	storageBackendClaim, ok := obj.(*csiV1.StorageBackendClaim)
	if !ok {
		log.Errorf("failed to convert obj to storageBackendClaim, obj is [%v]", obj)
		return
	}

	err := releaseCache(storageBackendClaim.Name)
	if err != nil {
		log.Errorln(err)
	}
}

func releaseCache(backendName string) error {
	log.Infof("start release backend [%s]", backendName)
	client, ok := clientCache[backendName]
	if !ok {
		log.Infof("backend [%s] client does not exist", backendName)
		return nil
	}

	centralizedClient, ok := client.Client.(*centralizedstorage.CentralizedClient)
	if !ok {
		return fmt.Errorf("backend [%s] client convert to centralizedClient failed", backendName)
	}

	centralizedClient.Logout(context.Background())
	RemoveClient(backendName)

	return nil
}
