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

// Package cmicore provides a shared library for storage backend operations
package cmicore

import (
	"context"
	"reflect"
	"sync"

	"k8s.io/client-go/tools/cache"

	csiV1 "github.com/Huawei/eSDK_K8S_Plugin/v4/client/apis/xuanwu/v1"
	sbcClient "github.com/Huawei/eSDK_K8S_Plugin/v4/pkg/client/clientset/versioned"
	csiInformers "github.com/Huawei/eSDK_K8S_Plugin/v4/pkg/client/informers/externalversions"
	"github.com/huawei/csm/v2/utils/log"
)

// BackendInformer watches StorageBackendClaim CRs and maintains the client cache.
// When a SBC is updated, the old client is removed and re-discovered.
// When a SBC is deleted, the client is removed from cache.
// Logout is the caller's responsibility; the cache only handles storage and removal.
type BackendInformer struct {
	sbcClient    sbcClient.Interface
	namespace    string
	clientCache  *ClientCache
	discoverFunc func(context.Context, string) (ClientInfo, error)

	factory  csiInformers.SharedInformerFactory
	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewBackendInformer creates a new BackendInformer
func NewBackendInformer(
	sbcClient sbcClient.Interface,
	namespace string,
	clientCache *ClientCache,
	discoverFunc func(context.Context, string) (ClientInfo, error),
) *BackendInformer {
	return &BackendInformer{
		sbcClient:    sbcClient,
		namespace:    namespace,
		clientCache:  clientCache,
		discoverFunc: discoverFunc,
	}
}

// Start begins watching StorageBackendClaim CRs
func (bi *BackendInformer) Start(ctx context.Context) error {
	bi.stopCh = make(chan struct{})

	factory := csiInformers.NewSharedInformerFactory(bi.sbcClient, 0)
	bi.factory = factory

	factory.Xuanwu().V1().StorageBackendClaims().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				bi.onUpdate(ctx, oldObj, newObj)
			},
			DeleteFunc: func(obj interface{}) {
				bi.onDelete(ctx, obj)
			},
		},
	)

	factory.Start(bi.stopCh)
	log.AddContext(ctx).Infof("backend informer started, watching SBC in namespace [%s]", bi.namespace)
	return nil
}

// Stop halts the informer and releases the stop channel
func (bi *BackendInformer) Stop() {
	bi.stopOnce.Do(func() {
		close(bi.stopCh)
	})
}

// onUpdate handles StorageBackendClaim update events
func (bi *BackendInformer) onUpdate(ctx context.Context, oldObj, newObj interface{}) {
	if unknown, ok := newObj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		newObj = unknown.Obj
	}

	if unknown, ok := oldObj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		oldObj = unknown.Obj
	}

	// Check whether obj is a StorageBackendClaim CR
	oldStorageBackendClaim, ok := oldObj.(*csiV1.StorageBackendClaim)
	if !ok {
		log.AddContext(ctx).Errorf("failed to convert old obj to storageBackendClaim, oldObj is [%v]", oldObj)
		return
	}

	newStorageBackendClaim, ok := newObj.(*csiV1.StorageBackendClaim)
	if !ok {
		log.AddContext(ctx).Errorf("failed to convert new obj to storageBackendClaim, newObj is [%v]", newObj)
		return
	}

	if reflect.DeepEqual(newStorageBackendClaim.Spec, oldStorageBackendClaim.Spec) {
		log.AddContext(ctx).Debugf("the spec struct of storageBackendClaim [%s] are not changed, "+
			"do not update backend cache", oldStorageBackendClaim.Name)
		return
	}

	// Logout and remove old client from cache
	if clientInfo, ok := bi.clientCache.GetClient(newStorageBackendClaim.Name); ok {
		clientInfo.Client.Logout(ctx)
	}
	bi.clientCache.RemoveClient(newStorageBackendClaim.Name)

	// Re-discover client with new config
	if _, err := bi.clientCache.DiscoverClient(ctx, newStorageBackendClaim.Name, bi.discoverFunc); err != nil {
		log.AddContext(ctx).Errorf("get Client failed, err is [%v]", err)
		return
	}

	log.AddContext(ctx).Infof("backend [%s] client re-initialized after StorageBackendClaim update",
		newStorageBackendClaim.Name)
}

// onDelete handles StorageBackendClaim delete events
func (bi *BackendInformer) onDelete(ctx context.Context, obj interface{}) {
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}

	storageBackendClaim, ok := obj.(*csiV1.StorageBackendClaim)
	if !ok {
		log.AddContext(ctx).Errorf("failed to convert obj to storageBackendClaim, obj is [%v]", obj)
		return
	}

	// Logout and remove client from cache
	if clientInfo, ok := bi.clientCache.GetClient(storageBackendClaim.Name); ok {
		clientInfo.Client.Logout(ctx)
	}
	bi.clientCache.RemoveClient(storageBackendClaim.Name)

	log.AddContext(ctx).Infof("backend [%s] client removed after StorageBackendClaim deletion",
		storageBackendClaim.Name)
}
