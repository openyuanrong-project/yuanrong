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
	"k8s.io/apimachinery/pkg/api/resource"

	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const (
	containerRaspInit container = "rasp-init"
	containerRasp     container = "rasp"
	raspDataPath                = "/opt/data"
)

func makeRaspContainer(funcSpec *types.FunctionSpecification) types.DelegateContainerSideCarConfig {
	raspEnv := []v1.EnvVar{
		{
			Name:  "RASP_SERVER_IP2",
			Value: funcSpec.ExtendedMetaData.RaspConfig.RaspServerIP,
		}, {
			Name:  "RASP_SERVER_PORT2",
			Value: funcSpec.ExtendedMetaData.RaspConfig.RaspServerPort,
		}, {
			Name:  "RASP_CONTAINER_DEPLOYTYPE",
			Value: "nonnuwaruntime",
		}, {
			Name: "RUNTIME_HOST_IP", ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1", FieldPath: "status.hostIP",
				}},
		}, {
			Name: podNameEnvNew, ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1", FieldPath: "metadata.name",
				}},
		},
		{Name: podIPEnv, ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				APIVersion: "v1", FieldPath: "status.podIP",
			}},
		},
	}

	for _, env := range funcSpec.ExtendedMetaData.RaspConfig.Envs {
		raspEnv = append(raspEnv, v1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		})
	}

	raspResource := v1.ResourceRequirements{}
	raspResource.Requests = map[v1.ResourceName]resource.Quantity{}
	raspResource.Requests["cpu"] = resource.MustParse(fmt.Sprintf("%dm", RaspDefaultCPU))
	raspResource.Requests["memory"] = resource.MustParse(fmt.Sprintf("%dMi", RaspDefaultMemory))
	raspResource.Limits = map[v1.ResourceName]resource.Quantity{}
	raspResource.Limits = map[v1.ResourceName]resource.Quantity{}
	raspResource.Limits["cpu"] = resource.MustParse(fmt.Sprintf("%dm", RaspDefaultCPU))
	raspResource.Limits["memory"] = resource.MustParse(fmt.Sprintf("%dMi", RaspDefaultMemory))

	raspMount := []v1.VolumeMount{
		{
			Name: DefaultDataVolumeName, MountPath: raspDataPath,
		},
		{
			Name: aiOpsVolume, MountPath: raspLogVolumeMountPath, SubPathExpr: raspLogVolumeMountSubPathExpr,
		},
	}

	ReadinessProbe := v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			Exec: &v1.ExecAction{Command: []string{"sh", "/opt/monitor/ready.sh"}},
		},
		InitialDelaySeconds: RaspDefaultInitialDelaySeconds,
		TimeoutSeconds:      RaspDefaultTimeoutSeconds,
		PeriodSeconds:       RaspDefaultPeriodSeconds,
		SuccessThreshold:    RaspDefaultSuccessThreshold,
		FailureThreshold:    RaspDefaultFailureThreshold,
	}
	LivenessProbe := v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			Exec: &v1.ExecAction{Command: []string{"sh", "/opt/monitor/health.sh"}},
		},
		InitialDelaySeconds: RaspDefaultInitialDelaySeconds,
		TimeoutSeconds:      RaspDefaultTimeoutSeconds,
		PeriodSeconds:       RaspDefaultPeriodSeconds,
		SuccessThreshold:    RaspDefaultSuccessThreshold,
		FailureThreshold:    RaspDefaultFailureThreshold,
	}
	return types.DelegateContainerSideCarConfig{
		Name:                 string(containerRasp),
		Image:                funcSpec.ExtendedMetaData.RaspConfig.RaspImage,
		Env:                  raspEnv,
		ResourceRequirements: raspResource,
		VolumeMounts:         raspMount,
		LivenessProbe:        LivenessProbe,
		ReadinessProbe:       ReadinessProbe,
	}
}

func makeRaspInitContainer(funcSpec *types.FunctionSpecification) types.DelegateInitContainerConfig {
	raspInitMount := []v1.VolumeMount{
		{
			Name: DefaultDataVolumeName, MountPath: raspDataPath,
		},
		{
			Name: aiOpsVolume, MountPath: raspLogVolumeMountPath, SubPathExpr: raspLogVolumeMountSubPathExpr,
		},
	}

	raspInitEnv := []v1.EnvVar{
		{
			Name: podNameEnvNew, ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				APIVersion: "v1", FieldPath: "metadata.name",
			}},
		},
		{Name: podIPEnv, ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				APIVersion: "v1", FieldPath: "status.podIP",
			}},
		},
	}

	raspInitResource := v1.ResourceRequirements{}
	raspInitResource.Requests = map[v1.ResourceName]resource.Quantity{}
	raspInitResource.Requests["cpu"] = resource.MustParse(fmt.Sprintf("%dm", RaspInitDefaultCPU))
	raspInitResource.Requests["memory"] = resource.MustParse(fmt.Sprintf("%dMi", RaspInitDefaultMemory))
	raspInitResource.Limits = map[v1.ResourceName]resource.Quantity{}
	raspInitResource.Limits = map[v1.ResourceName]resource.Quantity{}
	raspInitResource.Limits["cpu"] = resource.MustParse(fmt.Sprintf("%dm", RaspInitDefaultCPU))
	raspInitResource.Limits["memory"] = resource.MustParse(fmt.Sprintf("%dMi", RaspInitDefaultMemory))

	return types.DelegateInitContainerConfig{
		Name:                 string(containerRaspInit),
		Image:                funcSpec.ExtendedMetaData.RaspConfig.InitImage,
		Command:              []string{"sh"},
		Env:                  raspInitEnv,
		Args:                 []string{"-c", "/opt/huawei/secRASP/slaveagent_entrypoint.sh"},
		VolumeMounts:         raspInitMount,
		ResourceRequirements: raspInitResource,
	}
}
