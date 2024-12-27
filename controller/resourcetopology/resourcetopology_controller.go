/*
 Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.

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
	"context"
	"fmt"
	"time"

	coreV1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	informersCoreV1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	apiXuanwuV1 "github.com/huawei/csm/v2/client/apis/xuanwu/v1"
	controllerConfig "github.com/huawei/csm/v2/config/topology"
	cmiGrpc "github.com/huawei/csm/v2/grpc/lib/go/cmi"
	xuanwuClient "github.com/huawei/csm/v2/pkg/client/clientset/versioned"
	xuanwuClientInformers "github.com/huawei/csm/v2/pkg/client/informers/externalversions/xuanwu/v1"
	"github.com/huawei/csm/v2/utils/log"
)

// Controller defines the resourceTopology controller parameters
type Controller struct {
	cmiClient     *cmiGrpc.ClientSet
	kubeClient    kubernetes.Interface
	xuanwuClient  xuanwuClient.Interface
	eventRecorder record.EventRecorder
	reSyncPeriod  time.Duration

	topologyQueue    workqueue.RateLimitingInterface
	topologyInformer xuanwuClientInformers.ResourceTopologyInformer

	volumeQueue    workqueue.RateLimitingInterface
	volumeInformer informersCoreV1.PersistentVolumeInformer

	claimInformer informersCoreV1.PersistentVolumeClaimInformer

	podQueue    workqueue.RateLimitingInterface
	podInformer informersCoreV1.PodInformer
	podStore    cache.Store
}

// ControllerRequest is a request for new controller
type ControllerRequest struct {
	CmiClient        *cmiGrpc.ClientSet
	KubeClient       kubernetes.Interface
	XuanwuClient     xuanwuClient.Interface
	TopologyInformer xuanwuClientInformers.ResourceTopologyInformer
	VolumeInformer   informersCoreV1.PersistentVolumeInformer
	ClaimInformer    informersCoreV1.PersistentVolumeClaimInformer
	PodInformer      informersCoreV1.PodInformer
	ReSyncPeriod     time.Duration
	EventRecorder    record.EventRecorder
}

// NewController return a new ResourceTopologyController
func NewController(request ControllerRequest) *Controller {
	rtRateLimiter := workqueue.NewItemExponentialFailureRateLimiter(controllerConfig.GetRtRetryBaseDelay(),
		controllerConfig.GetRtRetryMaxDelay())
	pvRateLimiter := workqueue.NewItemExponentialFailureRateLimiter(controllerConfig.GetPvRetryBaseDelay(),
		controllerConfig.GetPvRetryMaxDelay())
	podRateLimiter := workqueue.NewItemExponentialFailureRateLimiter(controllerConfig.GetPodRetryBaseDelay(),
		controllerConfig.GetPodRetryMaxDelay())
	resourceTopologyQueueConfig := workqueue.RateLimitingQueueConfig{Name: "resourceTopology"}
	volumeQueueConfig := workqueue.RateLimitingQueueConfig{Name: "persistentVolume"}
	podQueueConfig := workqueue.RateLimitingQueueConfig{Name: "pod"}

	ctrl := &Controller{
		kubeClient:       request.KubeClient,
		xuanwuClient:     request.XuanwuClient,
		eventRecorder:    request.EventRecorder,
		reSyncPeriod:     request.ReSyncPeriod,
		cmiClient:        request.CmiClient,
		topologyQueue:    workqueue.NewRateLimitingQueueWithConfig(rtRateLimiter, resourceTopologyQueueConfig),
		topologyInformer: request.TopologyInformer,
		volumeQueue:      workqueue.NewRateLimitingQueueWithConfig(pvRateLimiter, volumeQueueConfig),
		volumeInformer:   request.VolumeInformer,
		claimInformer:    request.ClaimInformer,
		podQueue:         workqueue.NewRateLimitingQueueWithConfig(podRateLimiter, podQueueConfig),
		podInformer:      request.PodInformer,
		podStore:         cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc),
	}

	ctrl.addEventFunc()
	return ctrl
}

func (ctrl *Controller) addEventFunc() {
	ctrl.topologyInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { ctrl.enqueueResourceTopology(obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueueResourceTopology(newObj) },
			DeleteFunc: func(obj interface{}) { ctrl.enqueueResourceTopology(obj) },
		},
	)

	ctrl.volumeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { ctrl.enqueuePersistentVolume(obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueuePersistentVolume(newObj) },
			DeleteFunc: func(obj interface{}) { ctrl.enqueuePersistentVolume(obj) },
		},
	)

	ctrl.claimInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {},
			DeleteFunc: func(obj interface{}) {},
		},
	)

	ctrl.podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { ctrl.enqueuePod(obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { ctrl.enqueuePod(newObj) },
			DeleteFunc: func(obj interface{}) { ctrl.enqueuePod(obj) },
		},
	)
}

func (ctrl *Controller) enqueueResourceTopology(obj interface{}) {
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}

	if resourceTopology, ok := obj.(*apiXuanwuV1.ResourceTopology); ok {
		if !checkResourceTopologyName(resourceTopology.Name) {
			log.Debugf("unsupported prefix of resourceTopology [%s], skip to next", resourceTopology.Name)
			return
		}

		objName, err := cache.DeletionHandlingMetaNamespaceKeyFunc(resourceTopology)
		if err != nil {
			log.Errorf("fail to get key from object [%v] err: [%v]", resourceTopology, err)
			return
		}

		log.Infof("enqueued resourceTopology [%v] for sync", objName)
		ctrl.topologyQueue.Add(objName)
	}
}

func (ctrl *Controller) enqueuePersistentVolume(obj interface{}) {
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}

	if pv, ok := obj.(*coreV1.PersistentVolume); ok {
		if pv.Spec.CSI == nil {
			log.Debugf("pv [%s] is not a csi pv, skip to next", pv.Name)
			return
		}

		if pv.Spec.CSI.Driver != controllerConfig.GetCSIDriverName() {
			log.Debugf("pv [%s] driver [%s] is not supported, skip to next", pv.Name, pv.Spec.CSI.Driver)
			return
		}

		objName, err := cache.DeletionHandlingMetaNamespaceKeyFunc(pv)
		if err != nil {
			log.Errorf("fail to get key from object [%v] err: [%v]", pv, err)
			return
		}

		log.Infof("enqueued pv [%v] for sync", objName)
		ctrl.volumeQueue.Add(objName)
	}
}

func (ctrl *Controller) enqueuePod(obj interface{}) {
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}

	if pod, ok := obj.(*coreV1.Pod); ok {
		objName, err := cache.DeletionHandlingMetaNamespaceKeyFunc(pod)
		if err != nil {
			log.Errorf("fail to get key from object [%v] err: [%v]", pod, err)
			return
		}

		log.Infof("enqueued pod [%v] for sync", objName)
		ctrl.podQueue.Add(objName)
	}
}

// Run defines the resourceTopology controller process
func (ctrl *Controller) Run(ctx context.Context, workers int, stopCh <-chan struct{}) {
	defer ctrl.topologyQueue.ShutDown()
	defer ctrl.podQueue.ShutDown()
	defer ctrl.volumeQueue.ShutDown()
	log.AddContext(ctx).Infoln("starting topology controller")
	defer log.AddContext(ctx).Infoln("shutting down topology controller")

	if !cache.WaitForCacheSync(stopCh,
		ctrl.topologyInformer.Informer().HasSynced,
		ctrl.volumeInformer.Informer().HasSynced,
		ctrl.claimInformer.Informer().HasSynced,
		ctrl.podInformer.Informer().HasSynced) {
		log.AddContext(ctx).Errorln("cannot sync caches")
		return
	}

	log.AddContext(ctx).Infoln("starting workers")
	for i := 0; i < workers; i++ {
		go wait.Until(func() { ctrl.runResourceTopologyWorker(ctx) }, time.Second, stopCh)
		go wait.Until(func() { ctrl.runPersistentVolumeWorker(ctx) }, time.Second, stopCh)
		go wait.Until(func() { ctrl.runPodWorker(ctx) }, time.Second, stopCh)
	}
	log.AddContext(ctx).Infoln("started workers")
	defer log.AddContext(ctx).Infoln("shutting down workers")
	if stopCh != nil {
		sign := <-stopCh
		log.AddContext(ctx).Infof("resourceTopology controller exited, reason: [%v]", sign)
	}
}

func (ctrl *Controller) runResourceTopologyWorker(ctx context.Context) {
	for {
		if processNext := ctrl.processNextResourceTopologyWorkItem(ctx); !processNext {
			break
		}
	}
}

func (ctrl *Controller) runPersistentVolumeWorker(ctx context.Context) {
	for {
		if processNext := ctrl.processNextPersistentVolumeWorkItem(ctx); !processNext {
			break
		}
	}
}

func (ctrl *Controller) runPodWorker(ctx context.Context) {
	for {
		if processNext := ctrl.processNextPodWorkItem(ctx); !processNext {
			break
		}
	}
}

func (ctrl *Controller) processNextResourceTopologyWorkItem(ctx context.Context) bool {
	obj, shutdown := ctrl.topologyQueue.Get()
	if shutdown {
		log.AddContext(ctx).Infof("processNextResourceTopologyWorkItem obj: [%v], shutdown: [%v]", obj, shutdown)
		return false
	}

	err := ctrl.handle(ctx, obj, ctrl.topologyQueue, ctrl.handleResourceTopologyWork)
	if err != nil {
		log.AddContext(ctx).Errorln(err)
		return false
	}
	return true
}

func (ctrl *Controller) processNextPersistentVolumeWorkItem(ctx context.Context) bool {
	obj, shutdown := ctrl.volumeQueue.Get()
	if shutdown {
		log.AddContext(ctx).Infof("processNextPersistentVolumeWorkItem obj: [%v], shutdown: [%v]", obj, shutdown)
		return false
	}

	err := ctrl.handle(ctx, obj, ctrl.volumeQueue, ctrl.handlePersistentVolumeWork)
	if err != nil {
		log.AddContext(ctx).Errorln(err)
		return false
	}
	return true
}

func (ctrl *Controller) processNextPodWorkItem(ctx context.Context) bool {
	obj, shutdown := ctrl.podQueue.Get()
	if shutdown {
		log.AddContext(ctx).Infof("processNextPodWorkItem obj: [%v], shutdown: [%v]", obj, shutdown)
		return false
	}

	err := ctrl.handle(ctx, obj, ctrl.podQueue, ctrl.handlePodWork)
	if err != nil {
		log.AddContext(ctx).Errorln(err)
		return false
	}
	return true
}

func (ctrl *Controller) handle(ctx context.Context, obj interface{},
	queue workqueue.RateLimitingInterface, function func(ctx context.Context, key string) error) error {
	defer queue.Done(obj)
	var key string
	var ok bool

	if key, ok = obj.(string); !ok {
		queue.Forget(obj)
		log.AddContext(ctx).Errorf("expected string in workqueue but got [%#v]", obj)
		return nil
	}

	log.AddContext(ctx).Infof("start handle object [%s]", key)

	ctx, err := log.SetRequestInfo(ctx)
	if err != nil {
		queue.AddRateLimited(key)
		return fmt.Errorf("get requestIdCtx failed, error is [%v], requeuing key [%s]", err, key)
	}

	if err = function(ctx, key); err != nil {
		queue.AddRateLimited(key)
		return fmt.Errorf("handle key [%s] failed: [%s], requeuing key [%s]", key, err.Error(), key)
	}

	queue.Forget(obj)
	log.AddContext(ctx).Infof("syncHandle object [%s] successfully", key)
	return nil
}

func (ctrl *Controller) handleResourceTopologyWork(ctx context.Context, key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.AddContext(ctx).Errorf("invalid resource key: [%s]", key)
		return nil
	}
	resourceTopology, err := ctrl.topologyInformer.Lister().Get(name)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			log.AddContext(ctx).Infof("resourceTopology [%s] is no longer exists, end this work", name)
			return nil
		}
		log.AddContext(ctx).Errorf("get resourceTopology [%s] from the indexer cache failed", name)
		return err
	}

	// delete directly do nothing
	if resourceTopology.ObjectMeta.DeletionTimestamp != nil {
		return ctrl.deleteResourceTopology(ctx, resourceTopology)
	}

	return ctrl.syncResourceTopology(ctx, resourceTopology)
}

func (ctrl *Controller) handlePersistentVolumeWork(ctx context.Context, key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.AddContext(ctx).Errorf("invalid resource key: [%s]", key)
		return nil
	}
	pv, err := ctrl.volumeInformer.Lister().Get(name)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			log.AddContext(ctx).Infof("pv [%s] is no longer exists", name)
			return ctrl.removeResourceTopology(ctx, name)
		}
		log.AddContext(ctx).Errorf("get pv [%s] from the indexer cache failed", name)
		return err
	}

	if pv.ObjectMeta.DeletionTimestamp != nil {
		retryErr := fmt.Errorf("pv [%s] is deleting, wait the deletion finished to remove rt", name)
		return retryErr
	}

	return ctrl.syncPersistentVolume(ctx, pv)
}

func (ctrl *Controller) handlePodWork(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.AddContext(ctx).Errorf("invalid resource key: [%s]", key)
		return nil
	}
	pod, err := ctrl.podInformer.Lister().Pods(namespace).Get(name)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			log.AddContext(ctx).Infof("pod [%s/%s] is no longer exists", namespace, name)
			return ctrl.removePodTag(ctx, key)
		}
		log.AddContext(ctx).Errorf("get pod [%s/%s] from the indexer cache failed", namespace, name)
		return err
	}

	if pod.ObjectMeta.DeletionTimestamp != nil {
		retryErr := fmt.Errorf("pod [%s/%s] is deleting, "+
			"wait the deletion finished to remove rt label", namespace, name)
		return retryErr
	}

	return ctrl.syncPod(ctx, pod)
}
