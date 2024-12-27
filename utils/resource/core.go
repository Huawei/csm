/*
Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

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

// Package resource is used to obtain core resources in Kubernetes.
package resource

import (
	"fmt"
	"os"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
)

var (
	instance Ops
	once     sync.Once
)

// Ops includes k8s options
type Ops interface {
	SecretOps
	PodOps
	ConfigmapOps
	PVOps
}

// Instance returns a singleton instance of the client.
func Instance() Ops {
	once.Do(func() {
		if instance == nil {
			instance = &Client{}
		}
	})
	return instance
}

// WatchFunc is a callback provided to the Watch functions
// which is invoked when the given object is changed.
type WatchFunc func(object runtime.Object) error

// Client is a wrapper for kubernetes core client.
type Client struct {
	config     *rest.Config
	kubernetes kubernetes.Interface
	// eventRecorders is a map of component to event recorders
	eventRecorders     map[string]record.EventRecorder
	eventRecordersLock sync.Mutex
	eventBroadcaster   record.EventBroadcaster
}

// initClient the k8s client if uninitialized
func (c *Client) initClient() error {
	if c.kubernetes != nil {
		return nil
	}

	return c.setClient()
}

// setClient instantiates a client.
func (c *Client) setClient() error {
	var err error

	if c.config != nil {
		err = c.loadClient()
	} else {
		kubeconfig := os.Getenv("KUBECONFIG")
		if len(kubeconfig) > 0 {
			err = c.loadClientFromKubeconfig(kubeconfig)
		} else {
			err = c.loadClientFromServiceAccount()
		}

	}
	return err
}

func (c *Client) loadClient() error {
	if c.config == nil {
		return fmt.Errorf("rest config is not provided")
	}

	var err error

	c.kubernetes, err = kubernetes.NewForConfig(c.config)
	if err != nil {
		return err
	}
	return nil
}

// loadClientFromServiceAccount loads a k8s client from a ServiceAccount specified in the pod running px
func (c *Client) loadClientFromServiceAccount() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	c.config = config
	return c.loadClient()
}

func (c *Client) loadClientFromKubeconfig(kubeconfig string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	c.config = config
	return c.loadClient()
}
