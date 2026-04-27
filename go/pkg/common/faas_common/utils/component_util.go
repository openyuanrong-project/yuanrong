/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package utils -
package utils

import (
	"k8s.io/api/core/v1"
)

type container string

const (
	// ContainerRuntimeManager -
	ContainerRuntimeManager container = "runtime-manager"
)

// VolumeBuilder -
type VolumeBuilder struct {
	Volumes []v1.Volume
	Mounts  map[container][]v1.VolumeMount
}

// AddVolume -
func (vc *VolumeBuilder) AddVolume(volume v1.Volume) {
	vc.Volumes = append(vc.Volumes, volume)
}

// AddVolumeMount -
func (vc *VolumeBuilder) AddVolumeMount(name container, mount v1.VolumeMount) {
	vc.Mounts[name] = append(vc.Mounts[name], mount)
}

// NewVolumeBuilder -
func NewVolumeBuilder() *VolumeBuilder {
	return &VolumeBuilder{
		Mounts: make(map[container][]v1.VolumeMount),
	}
}
