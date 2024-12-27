/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

// Package server is a package that implement grpc interface
package server

import (
	"context"

	cmiConfig "github.com/huawei/csm/v2/config/cmi"
	"github.com/huawei/csm/v2/grpc/lib/go/cmi"
	"github.com/huawei/csm/v2/utils/log"
)

// Identity this object implements the cmi.IdentityServer service.
type Identity struct{}

// Probe return running status.
func (i *Identity) Probe(ctx context.Context, request *cmi.ProbeRequest) (*cmi.ProbeResponse, error) {
	log.AddContext(ctx).Debugln("Start probe")
	return &cmi.ProbeResponse{}, nil
}

// GetProvisionerInfo get provider info
func (i *Identity) GetProvisionerInfo(ctx context.Context,
	request *cmi.GetProviderInfoRequest) (*cmi.GetProviderInfoResponse, error) {
	log.AddContext(ctx).Infoln("Start get provider information")

	return &cmi.GetProviderInfoResponse{
		Provider: cmiConfig.GetProviderName(),
	}, nil

}

// GetProviderCapabilities get provider capabilities
func (i *Identity) GetProviderCapabilities(ctx context.Context,
	request *cmi.GetProviderCapabilitiesRequest) (*cmi.GetProviderCapabilitiesResponse, error) {
	log.AddContext(ctx).Infoln("Start get provider Capabilities")

	return &cmi.GetProviderCapabilitiesResponse{
		Capabilities: []*cmi.ProviderCapability{
			{
				Type: cmi.ProviderCapability_ProviderCapability_Label_Service,
			},
			{
				Type: cmi.ProviderCapability_ProviderCapability_Collect_Service,
			},
		},
	}, nil
}
