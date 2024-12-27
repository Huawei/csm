/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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

// Package backend is a package that manager storage backend
package backend

import (
	"context"
	"fmt"

	xuanwuV1 "github.com/Huawei/eSDK_K8S_Plugin/v4/client/apis/xuanwu/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	cmiConfig "github.com/huawei/csm/v2/config/cmi"
	"github.com/huawei/csm/v2/provider/constants"
	"github.com/huawei/csm/v2/provider/grpc/helper"
	"github.com/huawei/csm/v2/provider/utils"
	"github.com/huawei/csm/v2/storage/constant"
	"github.com/huawei/csm/v2/utils/log"
)

// volumeTypes volume type mapping
// Map key is the storage field of the backend.
// Map key is the volume type.
var volumeTypes = map[string]string{
	constants.StorageNas: constants.NasVolume,
	constants.StorageSan: constants.LunVolume,
}

// StorageBackendConfigBuilder storage backend config builder
type StorageBackendConfigBuilder struct {
	ctx         context.Context
	err         error
	sbc         *xuanwuV1.StorageBackendClaim
	config      *constant.StorageBackendConfig
	backendName string
}

// NewStorageBackendConfigBuilder init an instance of StorageBackendConfigBuilder
func NewStorageBackendConfigBuilder(ctx context.Context, backendName string) *StorageBackendConfigBuilder {
	return &StorageBackendConfigBuilder{ctx: ctx, backendName: backendName, config: &constant.StorageBackendConfig{}}
}

// Build init an instance of StorageBackendConfig
func (b *StorageBackendConfigBuilder) Build() (*constant.StorageBackendConfig, error) {
	return b.config, b.err
}

// WithSbcInfo build with sbc info
func (b *StorageBackendConfigBuilder) WithSbcInfo() *StorageBackendConfigBuilder {
	if b.err != nil {
		return b
	}

	sbc, err := helper.GetClientSet().SbcClient.XuanwuV1().StorageBackendClaims(cmiConfig.GetNamespace()).
		Get(b.ctx, b.backendName, metaV1.GetOptions{})
	if err != nil {
		log.AddContext(b.ctx).Errorf("Get StorageBackendClaims failed, error: %v", err)
		b.err = err
		return b
	}

	b.sbc = sbc
	b.config.StorageBackendNamespace = sbc.Namespace
	b.config.StorageBackendName = sbc.Name
	b.config.ClientMaxThreads = cmiConfig.GetClientMaxThreads()
	return b
}

// WithSecretInfo build with secret info
func (b *StorageBackendConfigBuilder) WithSecretInfo() *StorageBackendConfigBuilder {
	if b.err != nil {
		return b
	}

	secret, err := getSecretInfo(b.ctx, b.sbc.Status.SecretMeta)
	if err != nil {
		log.AddContext(b.ctx).Errorf("Get Secret failed, error: %v", err)
		b.err = err
		return b
	}

	if err := parseSecretInfo(secret, b.config); err != nil {
		log.AddContext(b.ctx).Errorf("parse Secret failed, error: %v", err)
		b.err = err
		return b
	}
	return b
}

// WithConfigMapInfo build with config map info
func (b *StorageBackendConfigBuilder) WithConfigMapInfo() *StorageBackendConfigBuilder {
	if b.err != nil {
		return b
	}

	configMap, err := getConfigmapInfo(b.ctx, b.sbc.Status.ConfigmapMeta)
	if err != nil {
		log.AddContext(b.ctx).Errorf("get ConfigMap failed, error: %v", err)
		b.err = err
		return b
	}

	if err := parseConfigmapInfo(b.ctx, configMap, b.config); err != nil {
		log.AddContext(b.ctx).Errorf("parse ConfigMap failed, error: %v", err)
		b.err = err
		return b
	}

	return b
}

func getSecretInfo(ctx context.Context, meta string) (*v1.Secret, error) {
	namespace, name, err := cache.SplitMetaNamespaceKey(meta)
	if err != nil {
		return nil, fmt.Errorf("split secret meta %s namespace failed, error: %v", meta, err)
	}

	secret, err := helper.GetClientSet().KubeClient.CoreV1().Secrets(namespace).Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get secret with name %s and namespace %s failed, error: %v",
			name, namespace, err)
	}
	return secret, nil
}

func getConfigmapInfo(ctx context.Context, configmapMeta string) (*v1.ConfigMap, error) {
	namespace, name, err := cache.SplitMetaNamespaceKey(configmapMeta)
	if err != nil {
		return nil, fmt.Errorf("split configmap meta %s namespace failed, error: %v", configmapMeta, err)
	}

	configmap, err := helper.GetClientSet().KubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get configmap for [%s] failed, error: %v", configmapMeta, err)
	}
	return configmap, nil
}

func parseConfigmapInfo(ctx context.Context, configmap *v1.ConfigMap, config *constant.StorageBackendConfig) error {
	configDataMap, err := utils.ConvertConfigmapToMap(configmap)
	if err != nil {
		return fmt.Errorf("convert configmap data to map failed. err is [%v]", err)
	}

	err = parseBackendType(configDataMap, config)
	if err != nil {
		return err
	}

	return parseBackendUrls(configDataMap, config)
}

func parseSecretInfo(secret *v1.Secret, storageConfig *constant.StorageBackendConfig) error {
	if secret.Data == nil {
		return fmt.Errorf("the Data not exist in secret %s", secret.Name)
	}

	if err := parseBackendUser(secret.Data, storageConfig); err != nil {
		return err
	}

	storageConfig.SecretNamespace = secret.Namespace
	storageConfig.SecretName = secret.Name

	return nil
}

func parseBackendUser(config map[string][]byte,
	storageConfig *constant.StorageBackendConfig) error {
	user, exist := config["user"]
	if !exist {
		return fmt.Errorf("the [user] filed not exist in secret")
	}
	storageConfig.User = string(user)
	return nil
}

func parseBackendUrls(config map[string]interface{}, storageConfig *constant.StorageBackendConfig) error {
	configUrls, ok := config["urls"].([]interface{})
	if !ok {
		return fmt.Errorf("the urls filed of config %v convert to []interface{} failed, please check", config)
	}

	urls := make([]string, len(configUrls))
	for i, arg := range configUrls {
		urls[i], ok = arg.(string)
		if !ok {
			return fmt.Errorf("convert interface{} [%v] to string failed, "+
				"configUrls is %v, please check ", arg, configUrls)
		}
	}

	storageConfig.Urls = urls
	return nil
}

func parseBackendType(config map[string]interface{}, storageConfig *constant.StorageBackendConfig) error {
	storage, exist := config["storage"]
	if !exist {
		return fmt.Errorf("the storage filed not exist in configmap Data %v", config)
	}

	storageConfig.StorageType = fmt.Sprintf("%s", storage)
	return nil
}
