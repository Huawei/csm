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
	"errors"
	"fmt"
	"sync"

	"github.com/huawei/csm/v2/utils/log"
)

// Core is the main facade that provides all cmicore functionality.
// It composes ClientCache, BackendInformer, LabelService, and CollectorFactory.
// Instead of using the global helper.GetClientSet() singleton, Core uses
// pre-initialized KubeClient and SbcClient from CoreConfig.
type Core struct {
	config           *CoreConfig
	clientCache      *ClientCache
	informer         *BackendInformer
	labelService     LabelService
	collectorFactory *CollectorFactory
	collectValidator *Validator[*CollectRequest]

	mu       sync.Mutex
	started  bool
	stopOnce sync.Once
}

// NewCore creates a new Core instance with the given configuration.
// All clients in the config must be pre-initialized by the caller.
func NewCore(config *CoreConfig) *Core {
	c := &Core{
		config: config,
	}

	// Create client cache (logout is handled by caller, not by cache)
	c.clientCache = NewClientCache()

	// Create discover function that uses the injected clients from CoreConfig
	discoverFunc := c.discoverClient

	// Create backend informer for watching StorageBackendClaim CRs
	c.informer = NewBackendInformer(
		config.SbcClient,
		config.BackendNamespace,
		c.clientCache,
		discoverFunc,
	)

	// Create label service (OceanStorage implementation)
	c.labelService = NewOceanStorageLabelService(c.clientCache, discoverFunc)

	// Create collector factory
	c.collectorFactory = NewCollectorFactory(c.clientCache, discoverFunc)

	// Create collect request validator (composable validation chain)
	c.collectValidator = NewValidator[*CollectRequest](
		c.validateBackendName,
		c.validateCollectType,
		c.validateMetricsType,
	)

	return c
}

// Start begins watching StorageBackendClaims and initializes the system.
// This should be called after NewCore, before any Collect/Label operations.
func (c *Core) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return nil // Already started, idempotent
	}

	if err := c.informer.Start(ctx); err != nil {
		c.informer.Stop()
		return fmt.Errorf("start backend informer failed: %w", err)
	}

	c.started = true
	log.AddContext(ctx).Infof("cmicore Core started successfully")
	return nil
}

// Stop halts all operations, logs out all cached clients, and cleans up resources.
// Safe to call multiple times (uses sync.Once).
// Logout is handled by Core (the business caller), not by the cache.
func (c *Core) Stop() {
	c.stopOnce.Do(func() {
		c.informer.Stop()

		// Logout all cached clients before clearing cache
		for _, backendName := range c.clientCache.GetAllBackendNames() {
			if clientInfo, ok := c.clientCache.GetClient(backendName); ok {
				clientInfo.Client.Logout(context.Background())
			}
		}
		c.clientCache.ClearAll()

		c.mu.Lock()
		c.started = false
		c.mu.Unlock()

		log.Infof("cmicore Core stopped")
	})
}

// CreateLabel creates a label on the storage backend.
// Delegates to the registered LabelService implementation.
func (c *Core) CreateLabel(ctx context.Context, req *LabelRequest) error {
	if c.labelService == nil {
		return errors.New("LabelService not initialized")
	}
	return c.labelService.CreateLabel(ctx, req)
}

// DeleteLabel deletes a label from the storage backend.
// Delegates to the registered LabelService implementation.
func (c *Core) DeleteLabel(ctx context.Context, req *LabelRequest) error {
	if c.labelService == nil {
		return errors.New("LabelService not initialized")
	}
	return c.labelService.DeleteLabel(ctx, req)
}

// GetStorageType returns the storage type (e.g. "oceanStorage", "fusionStorage") for the given backend.
// Returns an error if the backend is not found in the client cache.
func (c *Core) GetStorageType(backendName string) (string, error) {
	clientInfo, ok := c.clientCache.GetClient(backendName)
	if !ok {
		return "", fmt.Errorf("backend [%s] not found in client cache", backendName)
	}
	return clientInfo.StorageType, nil
}

// Collect collects metrics from the storage backend.
// Validates the request, dispatches to the appropriate collector based on MetricsType,
// and returns the collected response.
func (c *Core) Collect(ctx context.Context, req *CollectRequest) (*CollectResponse, error) {
	log.AddContext(ctx).Infof("Start to collect, request: %v", req)
	defer log.AddContext(ctx).Infof("Finish to collect, backend name %s", req.BackendName)

	// Validate request using composable validator
	if err := c.collectValidator.Validate(req); err != nil {
		return nil, err
	}

	// Get appropriate collector from factory
	collector, err := c.collectorFactory.GetCollector(req.MetricsType)
	if err != nil {
		log.AddContext(ctx).Errorf("Get collector failed, error: %v", err)
		return nil, err
	}
	log.AddContext(ctx).Infof("Get collector success, metricsType: %s", req.MetricsType)

	// Execute collection
	return collector.Collect(ctx, req)
}
