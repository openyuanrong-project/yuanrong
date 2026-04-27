/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

package instancepool

import (
	"fmt"
	"net/url"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const (
	containerOtelInit      container = "otel-init"
	otelShardDirVolumeName string    = "otel-shard-dir"
	otelWhiteList          string    = "OTEL_WHITELIST"
	otelEndPointEnvKey     string    = "OTEL_EXPORTER_OTLP_ENDPOINT"
)

func makeOtelInitContainer(funcSpec *types.FunctionSpecification) types.DelegateInitContainerConfig {
	initContainer := funcSpec.ExtendedMetaData.UserOtelConfig.InitContainer
	otelInitMount := []v1.VolumeMount{
		{
			Name:      otelShardDirVolumeName,
			MountPath: initContainer.ShardDir,
		},
	}

	initResource := v1.ResourceRequirements{
		Requests: map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", initContainer.ResourceRequest.Cpu)),
			v1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", initContainer.ResourceRequest.Memory)),
		},
		Limits: map[v1.ResourceName]resource.Quantity{
			v1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%dm", initContainer.ResourceLimit.Cpu)),
			v1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", initContainer.ResourceLimit.Memory)),
		},
	}

	return types.DelegateInitContainerConfig{
		Name:                 string(containerOtelInit),
		Image:                initContainer.Image,
		Command:              initContainer.Command,
		VolumeMounts:         otelInitMount,
		ResourceRequirements: initResource,
	}
}

func addOtelSystemEnv(envs []v1.EnvVar) []v1.EnvVar {
	envs = append(envs, v1.EnvVar{Name: "OTEL_NODE_IP", ValueFrom: &v1.EnvVarSource{
		FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.hostIP"},
	}})
	envs = append(envs, v1.EnvVar{Name: "OTEL_POD_IP", ValueFrom: &v1.EnvVarSource{
		FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"},
	}})
	envs = append(envs, v1.EnvVar{Name: "OTEL_RESOURCE_ATTRIBUTES_POD_NAME", ValueFrom: &v1.EnvVarSource{
		FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.name"},
	}})
	envs = append(envs, v1.EnvVar{Name: "OTEL_RESOURCE_ATTRIBUTES_NODE_NAME", ValueFrom: &v1.EnvVarSource{
		FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "spec.nodeName"},
	}})
	return envs
}

func getFunctionAgentInitOtelEnv(funcSpec *types.FunctionSpecification) ([]v1.EnvVar, error) {
	otelEndPoint, exists := funcSpec.ExtendedMetaData.UserOtelConfig.OtelEnv[otelEndPointEnvKey]
	if !exists {
		return nil, nil
	}
	parsedURL, err := url.ParseRequestURI(otelEndPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OTEL_EXPORTER_OTLP_ENDPOINT: %s", err.Error())
	}
	domain := parsedURL.Hostname()
	port := parsedURL.Port()
	// collector.namespace.svc.cluster.local,None,4318/TCP,;
	return []v1.EnvVar{
		{
			Name:  otelWhiteList,
			Value: fmt.Sprintf("%s,None,%s/TCP,;", domain, port),
		},
	}, nil
}
