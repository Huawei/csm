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
	"fmt"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/huawei/csm/v2/provider/utils"
	"github.com/huawei/csm/v2/storage/constant"
	"github.com/huawei/csm/v2/utils/log"
)

// discoverClient discovers a storage client for the given backend name.
// This replaces the original backend.GetClientByBackendName which used helper.GetClientSet() singleton.
// Instead, it uses the injected KubeClient and SbcClient from CoreConfig.
func (c *Core) discoverClient(ctx context.Context, backendName string) (ClientInfo, error) {
	log.AddContext(ctx).Infof("Start to discover client, name: %s", backendName)

	// Build storage backend config using injected clients (not singleton)
	config, err := c.buildStorageBackendConfig(ctx, backendName)
	if err != nil {
		log.AddContext(ctx).Errorf("build storage config failed, name: %s, error: %v", backendName, err)
		return ClientInfo{}, err
	}

	// Build client info using the config
	return c.buildClientInfo(ctx, config)
}

// buildStorageBackendConfig builds the storage backend config using injected KubeClient and SbcClient.
// This mirrors the logic in provider/backend/storage_config_builder.go but uses injected clients.
func (c *Core) buildStorageBackendConfig(ctx context.Context,
	backendName string) (*constant.StorageBackendConfig, error) {
	// Get StorageBackendClaim CR using injected SbcClient
	sbc, err := c.config.SbcClient.XuanwuV1().StorageBackendClaims(c.config.BackendNamespace).
		Get(ctx, backendName, metaV1.GetOptions{})
	if err != nil {
		log.AddContext(ctx).Errorf("Get StorageBackendClaims failed, error: %v", err)
		return nil, err
	}

	config := &constant.StorageBackendConfig{
		StorageBackendNamespace: sbc.Namespace,
		StorageBackendName:      sbc.Name,
		ClientMaxThreads:        c.config.ClientMaxThreads,
	}

	// Get Secret using injected KubeClient
	secret, err := c.getSecret(ctx, sbc.Status.SecretMeta)
	if err != nil {
		log.AddContext(ctx).Errorf("Get Secret failed, error: %v", err)
		return nil, err
	}

	if err := parseSecretInfo(secret, config); err != nil {
		log.AddContext(ctx).Errorf("parse Secret failed, error: %v", err)
		return nil, err
	}

	// Get ConfigMap using injected KubeClient
	configMap, err := c.getConfigMap(ctx, sbc.Status.ConfigmapMeta)
	if err != nil {
		log.AddContext(ctx).Errorf("get ConfigMap failed, error: %v", err)
		return nil, err
	}

	if err := parseConfigmapInfo(ctx, configMap, config); err != nil {
		log.AddContext(ctx).Errorf("parse ConfigMap failed, error: %v", err)
		return nil, err
	}

	return config, nil
}

// getSecret retrieves a Kubernetes Secret using the injected KubeClient
func (c *Core) getSecret(ctx context.Context, meta string) (*v1.Secret, error) {
	namespace, name, err := cache.SplitMetaNamespaceKey(meta)
	if err != nil {
		return nil, fmt.Errorf("split secret meta %s namespace failed, error: %w", meta, err)
	}

	secret, err := c.config.KubeClient.CoreV1().Secrets(namespace).Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get secret with name %s and namespace %s failed, error: %w",
			name, namespace, err)
	}
	return secret, nil
}

// getConfigMap retrieves a Kubernetes ConfigMap using the injected KubeClient
func (c *Core) getConfigMap(ctx context.Context, configmapMeta string) (*v1.ConfigMap, error) {
	namespace, name, err := cache.SplitMetaNamespaceKey(configmapMeta)
	if err != nil {
		return nil, fmt.Errorf("split configmap meta %s namespace failed, error: %w", configmapMeta, err)
	}

	configmap, err := c.config.KubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, name, metaV1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get configmap for [%s] failed, error: %w", configmapMeta, err)
	}
	return configmap, nil
}

// buildClientInfo builds ClientInfo from the config using ClientInfoBuilder
func (c *Core) buildClientInfo(ctx context.Context, config *constant.StorageBackendConfig) (ClientInfo, error) {
	return NewClientInfoBuilder(ctx).
		WithVolumeType(config.StorageType).
		WithClient(config).Build()
}

// parseSecretInfo parses secret data into the storage config
func parseSecretInfo(secret *v1.Secret, storageConfig *constant.StorageBackendConfig) error {
	if secret.Data == nil {
		return fmt.Errorf("the Data not exist in secret %s", secret.Name)
	}

	user, exist := secret.Data["user"]
	if !exist {
		return fmt.Errorf("the [user] field not exist in secret")
	}
	storageConfig.User = string(user)
	storageConfig.SecretNamespace = secret.Namespace
	storageConfig.SecretName = secret.Name

	return nil
}

// parseConfigmapInfo parses configmap data into the storage config
func parseConfigmapInfo(ctx context.Context, configmap *v1.ConfigMap, config *constant.StorageBackendConfig) error {
	configDataMap, err := utils.ConvertConfigmapToMap(configmap)
	if err != nil {
		return fmt.Errorf("convert configmap data to map failed: %w", err)
	}

	err = parseBackendType(configDataMap, config)
	if err != nil {
		return err
	}

	return parseBackendUrls(configDataMap, config)
}

// parseBackendType extracts storage type from configmap data
func parseBackendType(config map[string]interface{}, storageConfig *constant.StorageBackendConfig) error {
	storage, exist := config["storage"]
	if !exist {
		return fmt.Errorf("the storage field not exist in configmap Data %v", config)
	}
	storageConfig.StorageType = fmt.Sprintf("%s", storage)
	return nil
}

// parseBackendUrls extracts URLs from configmap data
func parseBackendUrls(config map[string]interface{}, storageConfig *constant.StorageBackendConfig) error {
	urlsValue, exists := config["urls"]
	if !exists {
		return fmt.Errorf("the urls field not exist in configmap Data %v", config)
	}

	configUrls, ok := urlsValue.([]interface{})
	if !ok {
		return fmt.Errorf("the urls field of config %v convert to []interface{} failed", config)
	}

	urls := make([]string, len(configUrls))
	for i, arg := range configUrls {
		urls[i], ok = arg.(string)
		if !ok {
			return fmt.Errorf("convert interface{} [%v] to string failed", arg)
		}
	}

	storageConfig.Urls = urls
	return nil
}
