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

// Package cmi provides grpc clients
package cmi

import (
	"fmt"
	"strings"

	"google.golang.org/grpc"

	"github.com/huawei/csm/v2/utils/log"
)

// ClientSet provides clients to access services
type ClientSet struct {
	LabelClient     LabelServiceClient
	CollectorClient CollectorClient
	IdentityClient  IdentityClient
	Conn            *grpc.ClientConn
}

// GetClientSet get client set
func GetClientSet(address string) (*ClientSet, error) {
	connect, err := buildGrpcConnect(address)
	if err != nil {
		return nil, err
	}
	return &ClientSet{
		LabelClient:     NewLabelServiceClient(connect),
		CollectorClient: NewCollectorClient(connect),
		IdentityClient:  NewIdentityClient(connect),
		Conn:            connect,
	}, nil
}

func buildGrpcConnect(address string) (*grpc.ClientConn, error) {
	log.Infof("Connecting to %s", address)

	unixPrefix := "unix://"
	if strings.HasPrefix(address, "/") {
		// It looks like filesystem path.
		address = unixPrefix + address
	}

	if !strings.HasPrefix(address, unixPrefix) {
		return nil, fmt.Errorf("invalid unix domain path [%s]", address)
	}

	dialOptions := []grpc.DialOption{
		grpc.WithInsecure()}

	return grpc.Dial(address, dialOptions...)
}
