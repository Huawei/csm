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

// Package version used to set and clean the service version
package version

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	coreV1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/huawei/csm/v2/utils/log"
	"github.com/huawei/csm/v2/utils/resource"
)

var mutex sync.Mutex

// InitVersionConfigMapWithName used for init the version configmap of the service with a configmap name
func InitVersionConfigMapWithName(containerName string, version string, namespaceEnv string,
	defaultNamespace string, cmName string) error {
	log.Infof("Init version is %s, osArch is %s", version, OSArch)

	namespace := os.Getenv(namespaceEnv)
	if namespace == "" {
		namespace = defaultNamespace
	}

	mutex.Lock()
	defer mutex.Unlock()

	cm, err := resource.Instance().GetConfigmap(cmName, namespace)
	if apiErrors.IsNotFound(err) {
		err = createConfigMap(containerName, version, namespace, cmName)
		if err != nil {
			return err
		}
	} else if err != nil {
		errMsg := fmt.Sprintf("get configMap err: %s", err)
		return errors.New(errMsg)
	}

	for true {
		cm, err = resource.Instance().GetConfigmap(cmName, namespace)
		if err != nil {
			errMsg := fmt.Sprintf("get configMap err: %s", err)
			return errors.New(errMsg)
		}

		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		cm.Data[containerName] = version
		cm, err = resource.Instance().UpdateConfigmap(cm)
		if err != nil && apiErrors.IsConflict(err) {
			time.Sleep(time.Second)
			continue
		} else if err != nil {
			errMsg := fmt.Sprintf("update configMap err: %s", err)
			return errors.New(errMsg)
		}
		break
	}
	return nil
}

func createConfigMap(containerName, version, namespace, cmName string) error {
	cm := &coreV1.ConfigMap{}
	cm.Name = cmName
	cm.Namespace = namespace
	cm.Data = make(map[string]string)
	cm.Data[containerName] = version
	_, err := resource.Instance().CreateConfigmap(cm)
	if err != nil && !apiErrors.IsAlreadyExists(err) {
		errMsg := fmt.Sprintf("create configMap err: %s", err)
		return errors.New(errMsg)
	}
	return nil
}
