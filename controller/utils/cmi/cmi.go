/*
 Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

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

// Package cmi provides CreateLabel and DeleteLabel interface for cmi
package cmi

// Params used to call cmi grpc interface
type Params struct {
	volumeId    string
	labelName   string
	kind        string
	namespace   string
	clusterName string
}

// VolumeId get volume id
func (p *Params) VolumeId() string {
	return p.volumeId
}

// LabelName get label name
func (p *Params) LabelName() string {
	return p.labelName
}

// Kind get kind
func (p *Params) Kind() string {
	return p.kind
}

// Namespace get namespace
func (p *Params) Namespace() string {
	return p.namespace
}

// ClusterName get cluster name
func (p *Params) ClusterName() string {
	return p.clusterName
}

// SetVolumeId sets volumeId field
func (p *Params) SetVolumeId(volumeId string) *Params {
	p.volumeId = volumeId
	return p
}

// SetLabelName sets labelName field
func (p *Params) SetLabelName(labelName string) *Params {
	p.labelName = labelName
	return p
}

// SetKind sets kind field
func (p *Params) SetKind(kind string) *Params {
	p.kind = kind
	return p
}

// SetNamespace sets namespace field
func (p *Params) SetNamespace(namespace string) *Params {
	p.namespace = namespace
	return p
}

// SetClusterName sets clusterName field
func (p *Params) SetClusterName(clusterName string) *Params {
	p.clusterName = clusterName
	return p
}
