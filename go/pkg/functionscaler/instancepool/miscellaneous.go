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

	"k8s.io/api/core/v1"

	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
)

const (
	defaultCaasSysConfigSuffix          = "-sys-config"
	defaultCaasSysConfigVolumeName      = "caas-sys-config"
	defaultCaasSysConfigVolumeMountPath = "/opt/config/caas-sys-dynamic-config"
	defaultWorkerStsSuffix              = "-worker-sts"
	defaultWorkerStsVolumeName          = "sts-config"
	defaultAgentStsSuffix               = "-agent-sts"
	defaultAgentStsVolumeName           = "agent-sts-config"
	defaultWorkerStsVolumeMountPath     = "/opt/certs/HMSClientCloudAccelerateService/HMSCaaSYuanRongWorker/"
	defaultAgentStsVolumeMountPath      = "/opt/certs/WiseCloudElasticResourceService/ERSDataSystem/"
)

type otelShared struct{}

func (o *otelShared) configVolume(b *volumeBuilder, dir string) {
	b.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:      otelShardDirVolumeName,
		MountPath: dir,
	})
	b.addVolume(v1.Volume{
		Name:         otelShardDirVolumeName,
		VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
	})
}

type dataSystemSocket struct{}

func (u *dataSystemSocket) configVolume(b *volumeBuilder) {
	b.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:      "datasystem-socket",
		MountPath: "/home/uds",
	})
}

type serviceAccountToken struct{}

func (s *serviceAccountToken) configVolume(b *volumeBuilder) {
	b.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:      "oidc-token",
		MountPath: "/var/run/secrets/tokens",
		ReadOnly:  true,
	})
	expirationSeconds := int64(7200)
	defaultMode := int32(urnutils.DefaultMode)
	b.addVolume(v1.Volume{
		Name: "oidc-token",
		VolumeSource: v1.VolumeSource{
			Projected: &v1.ProjectedVolumeSource{
				Sources: []v1.VolumeProjection{{ServiceAccountToken: &v1.ServiceAccountTokenProjection{
					Audience:          "agentrunservice",
					ExpirationSeconds: &expirationSeconds,
					Path:              "oidc-token",
				}}},
				DefaultMode: &defaultMode,
			},
		},
	})
}

type cgroupMemory struct{}

func (f *cgroupMemory) configVolume(b *volumeBuilder) {
	b.addVolume(v1.Volume{
		Name: "cgroup-memory",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: "/sys/fs/cgroup/memory/kubepods/burstable",
			},
		},
	})

	mount := v1.VolumeMount{
		Name:        "cgroup-memory",
		MountPath:   "/runtime/memory",
		SubPathExpr: "pod$(POD_ID)",
	}
	b.addVolumeMount(containerRuntimeManager, mount)
}

type dockerSocket struct{}

func (v *dockerSocket) configVolume(b *volumeBuilder) {
	b.addVolume(v1.Volume{
		Name: "docker-socket",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: "/var/run/docker.sock",
			},
		},
	})
	b.addVolumeMount(containerRuntimeManager, v1.VolumeMount{
		Name:      "docker-socket",
		MountPath: "/var/run/docker.sock",
	})
}

type dockerRootDir struct{}

func (v *dockerRootDir) configVolume(b *volumeBuilder) {
	dockerRootPath := os.Getenv(config.DockerRootPathEnv)
	b.addVolume(v1.Volume{
		Name: "docker-rootdir",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: dockerRootPath,
			},
		},
	})

	b.addVolumeMount(containerRuntimeManager, v1.VolumeMount{
		Name:      "docker-rootdir",
		MountPath: "/var/lib/docker",
	})
}

type dataVolume struct{}

func (c *dataVolume) configVolume(b *volumeBuilder) {
	b.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:      DefaultDataVolumeName,
		MountPath: "/opt/data",
	})
}

type faasAgentSts struct {
	crName string
}

func (f *faasAgentSts) configVolume(b *volumeBuilder) {
	securityFileMode := int32(urnutils.OwnerReadWrite)
	b.addVolume(v1.Volume{
		Name: defaultAgentStsVolumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				DefaultMode: &securityFileMode,
				SecretName:  f.crName + defaultAgentStsSuffix,
				Items: []v1.KeyToPath{
					{
						Key:  "a",
						Path: "a",
					},
					{
						Key:  "b",
						Path: "b",
					},
					{
						Key:  "c",
						Path: "c",
					},
					{
						Key:  "d",
						Path: "d",
					},
					{
						Key:  "ERSDataSystem.ini",
						Path: "ERSDataSystem.ini",
					},
					{
						Key:  "ERSDataSystem.sts.p12",
						Path: "ERSDataSystem.sts.p12",
					},
				},
			},
		},
	})
	b.addVolumeMount(containerFunctionAgent, v1.VolumeMount{
		Name:      defaultAgentStsVolumeName,
		MountPath: defaultAgentStsVolumeMountPath + "ERSDataSystem/apple/a",
		SubPath:   "a",
	})
	b.addVolumeMount(containerFunctionAgent, v1.VolumeMount{
		Name:      defaultAgentStsVolumeName,
		MountPath: defaultAgentStsVolumeMountPath + "ERSDataSystem/boy/b",
		SubPath:   "b",
	})
	b.addVolumeMount(containerFunctionAgent, v1.VolumeMount{
		Name:      defaultAgentStsVolumeName,
		MountPath: defaultAgentStsVolumeMountPath + "ERSDataSystem/cat/c",
		SubPath:   "c",
	})
	b.addVolumeMount(containerFunctionAgent, v1.VolumeMount{
		Name:      defaultAgentStsVolumeName,
		MountPath: defaultAgentStsVolumeMountPath + "ERSDataSystem/dog/d",
		SubPath:   "d",
	})
	b.addVolumeMount(containerFunctionAgent, v1.VolumeMount{
		Name:      defaultAgentStsVolumeName,
		MountPath: defaultAgentStsVolumeMountPath + "ERSDataSystem.ini",
		SubPath:   "ERSDataSystem.ini",
	})
	b.addVolumeMount(containerFunctionAgent, v1.VolumeMount{
		Name:      defaultAgentStsVolumeName,
		MountPath: defaultAgentStsVolumeMountPath + "ERSDataSystem.sts.p12",
		SubPath:   "ERSDataSystem.sts.p12",
	})
}

type caasSysConfig struct {
	crName string
}

func (c *caasSysConfig) configVolume(vb *volumeBuilder) {
	securityFileMode := int32(urnutils.OwnerReadWrite)
	vb.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:      defaultCaasSysConfigVolumeName,
		MountPath: defaultCaasSysConfigVolumeMountPath,
	})
	vb.addVolume(v1.Volume{
		Name: defaultCaasSysConfigVolumeName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: c.crName + defaultCaasSysConfigSuffix,
				},
				DefaultMode: &securityFileMode,
			},
		},
	})
}

type functionDefaultConfig struct {
	configName string
	mount      v1.VolumeMount
}

func (f *functionDefaultConfig) configVolume(vb *volumeBuilder) {
	securityFileMode := int32(urnutils.DefaultMode)
	vb.addVolumeMount(containerDelegate, f.mount)
	vb.addVolume(v1.Volume{
		Name: f.mount.Name,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: f.configName,
				},
				DefaultMode: &securityFileMode,
			},
		},
	})
}

type npuDriver struct {
}

func (nh *npuDriver) configVolume(vb *volumeBuilder) {
	hostPathDirectoryOrCreate := v1.HostPathDirectoryOrCreate
	vb.addVolume(v1.Volume{
		Name: "npu-driver",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: "/var/queue_schedule",
				Type: &hostPathDirectoryOrCreate,
			},
		},
	})
	vb.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:      "npu-driver",
		MountPath: "/var/queue_schedule",
	})
}

type ascendConfig struct{}

func (ad *ascendConfig) configVolume(vb *volumeBuilder) {
	var (
		ascendDriverPathVolumeName = "ascend-driver-path"
		ascendConfigVolumeName     = "ascend-config"
	)
	hostPathFile := v1.HostPathFile
	vb.addVolume(v1.Volume{
		Name: "ascend-npu-smi",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: "/usr/local/sbin/npu-smi",
				Type: &hostPathFile,
			},
		},
	})
	vb.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:      ascendDriverPathVolumeName,
		MountPath: "/usr/local/Ascend/driver",
		ReadOnly:  true,
	})
	vb.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:      ascendConfigVolumeName,
		MountPath: "/opt/config/ascend_config",
	})
	vb.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:      "ascend-npu-smi",
		MountPath: "/usr/local/sbin/npu-smi",
	})
}
