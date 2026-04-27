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

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestAiHostPathConfig(t *testing.T) {
	ah := aiHostPathConfig{
		WorkloadName: "testWorkload",
		Namespace:    "testNamespace",
	}
	vb := &volumeBuilder{
		volumes: make([]corev1.Volume, 0),
		mounts:  make(map[container][]corev1.VolumeMount),
	}
	ah.configVolume(vb)
	found := false
	for _, v := range vb.volumes {
		if v.Name == aiOpsVolume && v.VolumeSource.HostPath.Path == fmt.Sprintf(aiOpsVolumeHostPath, ah.Namespace,
			ah.WorkloadName) {
			found = true
		}
	}
	assert.Equal(t, true, found)
}

func TestBiHostPathConfig(t *testing.T) {
	bh := biHostPathConfig{
		WorkloadName: "testWorkload",
		Namespace:    "testNamespace",
	}
	vb := &volumeBuilder{
		volumes: make([]corev1.Volume, 0),
		mounts:  make(map[container][]corev1.VolumeMount),
	}
	bh.configVolume(vb)
	found := false
	for _, v := range vb.volumes {
		if v.Name == biLogsVolume && v.VolumeSource.HostPath.Path == fmt.Sprintf(biLogVolumeHostPath, bh.Namespace,
			bh.WorkloadName) {
			found = true
		}
	}
	assert.Equal(t, true, found)
}
