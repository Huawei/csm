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
	"sync"

	"github.com/huawei/csm/v2/utils/log"
)

// ClientCache manages cached storage backend clients with thread-safe operations.
// Key is backend name (StorageBackendClaim name), value is ClientInfo.
// Cache only handles storage and retrieval; business logic (e.g., logout) is the caller's responsibility.
type ClientCache struct {
	mu      sync.RWMutex
	clients map[string]ClientInfo
}

// NewClientCache creates a new ClientCache instance
func NewClientCache() *ClientCache {
	return &ClientCache{
		clients: make(map[string]ClientInfo),
	}
}

// RegisterClient adds or updates a client in the cache
func (cc *ClientCache) RegisterClient(backendName string, info ClientInfo) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.clients[backendName] = info
}

// RemoveClient removes a client from the cache
func (cc *ClientCache) RemoveClient(backendName string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	delete(cc.clients, backendName)
}

// GetClient retrieves a client from cache by backend name
// Returns the client and true if found, empty ClientInfo and false otherwise
func (cc *ClientCache) GetClient(backendName string) (ClientInfo, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	client, ok := cc.clients[backendName]
	return client, ok
}

// DiscoverClient discovers a client using the provided function and caches the result.
// Uses double-check locking to avoid duplicate discovery when multiple goroutines
// request the same backend concurrently.
func (cc *ClientCache) DiscoverClient(ctx context.Context, backendName string,
	discoverFunc func(context.Context, string) (ClientInfo, error)) (ClientInfo, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Double-check: another goroutine may have already discovered
	if client, ok := cc.clients[backendName]; ok {
		return client, nil
	}

	client, err := discoverFunc(ctx, backendName)
	if err != nil {
		log.AddContext(ctx).Errorf("discover client failed, backend name: [%s], error: [%v]", backendName, err)
		return ClientInfo{}, err
	}

	cc.clients[backendName] = client
	return client, nil
}

// ClearAll removes all clients from the cache.
// Business logic such as logout should be handled by the caller before clearing.
func (cc *ClientCache) ClearAll() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.clients = make(map[string]ClientInfo)
}

// GetAllBackendNames returns all backend names currently in cache
func (cc *ClientCache) GetAllBackendNames() []string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	names := make([]string, 0, len(cc.clients))
	for name := range cc.clients {
		names = append(names, name)
	}
	return names
}

// Size returns the number of cached clients
func (cc *ClientCache) Size() int {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return len(cc.clients)
}
