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

	"k8s.io/api/core/v1"
)

const (
	aiOpsVolume  = "aiops-logs"
	biLogsVolume = "bi-logs"
)

var (
	raspLogVolumeMountPath        = "/opt/logs"
	raspLogVolumeMountSubPathExpr = fmt.Sprintf("$(%s)_$(%s)/rasp", podNameEnvNew, podIPEnv)

	caasUserVolumeMountPath        = "/opt/logs/caas/user"
	caasUserVolumeMountSubPathExpr = fmt.Sprintf("$(%s)_$(%s)/caasUser", podNameEnvNew, podIPEnv)

	dataSystemVolumeMountPath        = "/opt/logs/caas/dataSystem"
	dataSystemVolumeMountSubPathExpr = fmt.Sprintf("$(%s)_$(%s)/dataSystem", podNameEnvNew, podIPEnv)

	customAIOpsMountPath        = "/opt/logs"
	customAIOpsMountSubPathExpr = fmt.Sprintf("$(%s)_$(%s)/custom", podNameEnvNew, podIPEnv)

	biLogVolumeMountPath        = "/opt/logs/caas/bi"
	biLogVolumeMountSubPathExpr = fmt.Sprintf("$(%s)_$(%s)/caasUser", podNameEnvNew, podIPEnv)

	aiOpsVolumeHostPath = "/mnt/daemonset/aiops/%s/%s"
	biLogVolumeHostPath = "/mnt/daemonset/bi/%s/%s"
)

type aiHostPathConfig struct {
	WorkloadName string
	Namespace    string
}

func (ai *aiHostPathConfig) configVolume(vb *volumeBuilder) {
	hostPathDirectoryOrCreate := v1.HostPathDirectoryOrCreate
	vb.addVolume(v1.Volume{
		Name: aiOpsVolume,
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: fmt.Sprintf(aiOpsVolumeHostPath, ai.Namespace, ai.WorkloadName),
				Type: &hostPathDirectoryOrCreate,
			},
		},
	})

	vb.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:        aiOpsVolume,
		MountPath:   caasUserVolumeMountPath,
		SubPathExpr: caasUserVolumeMountSubPathExpr,
	})
	vb.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:        aiOpsVolume,
		MountPath:   dataSystemVolumeMountPath,
		SubPathExpr: dataSystemVolumeMountSubPathExpr,
	})
	vb.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:        aiOpsVolume,
		MountPath:   customAIOpsMountPath,
		SubPathExpr: customAIOpsMountSubPathExpr,
	})
}

type biHostPathConfig struct {
	WorkloadName string
	Namespace    string
}

func (bc *biHostPathConfig) configVolume(vb *volumeBuilder) {
	hostPathDirectoryOrCreate := v1.HostPathDirectoryOrCreate
	vb.addVolume(v1.Volume{
		Name: biLogsVolume,
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: fmt.Sprintf(biLogVolumeHostPath, bc.Namespace, bc.WorkloadName),
				Type: &hostPathDirectoryOrCreate,
			},
		},
	})
	vb.addVolumeMount(containerDelegate, v1.VolumeMount{
		Name:        biLogsVolume,
		MountPath:   biLogVolumeMountPath,
		SubPathExpr: biLogVolumeMountSubPathExpr,
	})
}
