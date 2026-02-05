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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
)

func TestUnixSocketMethod(t *testing.T) {
	us := dataSystemSocket{}
	vb := volumeBuilder{
		volumes: make([]corev1.Volume, 0),
		mounts:  make(map[container][]corev1.VolumeMount),
	}
	us.configVolume(&vb)
	found := false
	for _, vm := range vb.mounts[containerDelegate] {
		if vm.Name == "datasystem-socket" && vm.MountPath == "/home/uds" {
			found = true
		}
	}
	assert.Equal(t, true, found)
}

func TestOtelSharedVolume(t *testing.T) {
	osd := otelShared{}
	vb := volumeBuilder{
		volumes: make([]corev1.Volume, 0),
		mounts:  make(map[container][]corev1.VolumeMount),
	}
	osd.configVolume(&vb, "otel")
	found := false
	for _, vm := range vb.mounts[containerDelegate] {
		if vm.Name == otelShardDirVolumeName && vm.MountPath == "otel" {
			found = true
		}
	}
	expectVolumeSource := corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}
	var actualVolumeSource corev1.VolumeSource
	for _, volume := range vb.volumes {
		if volume.Name == otelShardDirVolumeName {
			actualVolumeSource = volume.VolumeSource
		}
	}
	assert.Equal(t, true, found)
	assert.Equal(t, expectVolumeSource, actualVolumeSource)
}

func TestServiceAccountTokenVolume(t *testing.T) {
	sa := serviceAccountToken{}
	vb := volumeBuilder{
		volumes: make([]corev1.Volume, 0),
		mounts:  make(map[container][]corev1.VolumeMount),
	}
	sa.configVolume(&vb)
	found := false
	for _, vm := range vb.mounts[containerDelegate] {
		if vm.Name == "oidc-token" && vm.MountPath == "/var/run/secrets/tokens" {
			found = true
		}
	}
	expirationSeconds := int64(7200)
	defaultMode := int32(urnutils.DefaultMode)
	expectVolumeSource := corev1.VolumeSource{
		Projected: &corev1.ProjectedVolumeSource{
			Sources: []corev1.VolumeProjection{{ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
				Audience:          "agentrunservice",
				ExpirationSeconds: &expirationSeconds,
				Path:              "oidc-token",
			}}},
			DefaultMode: &defaultMode,
		},
	}
	var actualVolumeSource corev1.VolumeSource
	for _, volume := range vb.volumes {
		if volume.Name == "oidc-token" {
			actualVolumeSource = volume.VolumeSource
		}
	}
	assert.Equal(t, true, found)
	assert.Equal(t, expectVolumeSource, actualVolumeSource)
}

func TestCgroupMemoryMethod(t *testing.T) {
	cm := cgroupMemory{}
	vb := volumeBuilder{
		volumes: make([]corev1.Volume, 0),
		mounts:  make(map[container][]corev1.VolumeMount),
	}
	cm.configVolume(&vb)
	found := false
	for _, v := range vb.volumes {
		if v.Name == "cgroup-memory" && v.VolumeSource.HostPath.Path == "/sys/fs/cgroup/memory/kubepods/burstable" {
			found = true
		}
	}
	assert.Equal(t, true, found)
	found = false
	for _, vm := range vb.mounts[containerRuntimeManager] {
		if vm.Name == "cgroup-memory" && vm.MountPath == "/runtime/memory" {
			found = true
		}
	}
	assert.Equal(t, true, found)
}

func TestDockerSocketMethod(t *testing.T) {
	ds := dockerSocket{}
	vb := volumeBuilder{
		volumes: make([]corev1.Volume, 0),
		mounts:  make(map[container][]corev1.VolumeMount),
	}
	ds.configVolume(&vb)
	found := false
	for _, v := range vb.volumes {
		if v.Name == "docker-socket" && v.VolumeSource.HostPath.Path == "/var/run/docker.sock" {
			found = true
		}
	}
	assert.Equal(t, true, found)
	found = false
	for _, vm := range vb.mounts[containerRuntimeManager] {
		if vm.Name == "docker-socket" && vm.MountPath == "/var/run/docker.sock" {
			found = true
		}
	}
	assert.Equal(t, true, found)
}

func TestDockerRootDirMethod(t *testing.T) {
	dr := dockerRootDir{}
	vb := volumeBuilder{
		volumes: make([]corev1.Volume, 0),
		mounts:  make(map[container][]corev1.VolumeMount),
	}
	dockerRootPath := "/test/docker/root/path"
	os.Setenv(config.DockerRootPathEnv, dockerRootPath)
	dr.configVolume(&vb)
	found := false
	for _, v := range vb.volumes {
		if v.Name == "docker-rootdir" && v.VolumeSource.HostPath.Path == dockerRootPath {
			found = true
		}
	}
	assert.Equal(t, true, found)
	found = false
	for _, vm := range vb.mounts[containerRuntimeManager] {
		if vm.Name == "docker-rootdir" && vm.MountPath == "/var/lib/docker" {
			found = true
		}
	}
	assert.Equal(t, true, found)
}

func TestAscendConfig(t *testing.T) {
	sc := ascendConfig{}
	vb := volumeBuilder{
		volumes: make([]corev1.Volume, 0),
		mounts:  make(map[container][]corev1.VolumeMount),
	}
	sc.configVolume(&vb)
	found := false
	for _, v := range vb.volumes {
		if v.Name == "ascend-npu-smi" && v.VolumeSource.HostPath.Path == "/usr/local/sbin/npu-smi" {
			found = true
		}
	}
	assert.Equal(t, true, found)
}

func TestConfigVolume2(t *testing.T) {
	vb := volumeBuilder{
		volumes: make([]corev1.Volume, 0),
		mounts:  make(map[container][]corev1.VolumeMount),
	}
	nd := npuDriver{}
	nd.configVolume(&vb)
	cs := caasSysConfig{}
	cs.configVolume(&vb)
	dv := dataVolume{}
	dv.configVolume(&vb)

	assert.Equal(t, 2, len(vb.volumes))
}
