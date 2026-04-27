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

// Package instancepool -
package instancepool

import "k8s.io/api/core/v1"

type container string

const (
	containerDelegate          container = "delegate-container"
	containerRuntimeManager    container = "runtime-manager"
	containerFunctionAgent     container = "function-agent"
	containerFunctionAgentInit container = "function-agent-init"
)

type volumeBuilder struct {
	volumes []v1.Volume
	mounts  map[container][]v1.VolumeMount
}

func (vc *volumeBuilder) addVolume(volume v1.Volume) {
	vc.volumes = append(vc.volumes, volume)
}

func (vc *volumeBuilder) addVolumeMount(name container, mount v1.VolumeMount) {
	vc.mounts[name] = append(vc.mounts[name], mount)
}

func newVolumeBuilder() *volumeBuilder {
	return &volumeBuilder{
		mounts: make(map[container][]v1.VolumeMount),
	}
}

type envBuilder struct {
	envs map[container][]v1.EnvVar
}

func (b *envBuilder) addEnvVar(name container, env v1.EnvVar) {
	b.envs[name] = append(b.envs[name], env)
}

func newEnvBuilder() *envBuilder {
	return &envBuilder{
		envs: make(map[container][]v1.EnvVar),
	}
}
