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
	volumeStsConfig  = "runtime-sts-config"
	volumeCertVolume = "runtime-certs-volume"
	baseCertFilePath = "/opt/certs"

	// CertMode -
	CertMode = 384 // 600:rw- --- ---
)

type stsSecret struct {
	param *types.FunctionSpecification
	req   types.PodRequest
}

func (v *stsSecret) configVolume(b *volumeBuilder) {
	serviceName := v.param.StsMetaData.ServiceName
	microServiceName := v.param.StsMetaData.MicroService
	certVolumeQuality := resource.MustParse("10Mi")
	securityFileMode := int32(CertMode)
	b.addVolume(v1.Volume{Name: volumeStsConfig, VolumeSource: v1.VolumeSource{
		Secret: &v1.SecretVolumeSource{
			SecretName:  v.req.FunSvcID,
			DefaultMode: &securityFileMode,
			Items: []v1.KeyToPath{{Key: "a", Path: "a"}, {Key: "b", Path: "b"},
				{Key: "c", Path: "c"}, {Key: "d", Path: "d"},
				{Key: microServiceName + ".ini", Path: microServiceName + ".ini"},
				{Key: microServiceName + ".sts.p12", Path: microServiceName + ".sts.p12"}}}}})

	b.addVolume(v1.Volume{Name: volumeCertVolume, VolumeSource: v1.VolumeSource{
		EmptyDir: &v1.EmptyDirVolumeSource{SizeLimit: &certVolumeQuality, Medium: v1.StorageMediumMemory}}})

	// a/b/c/d - four sub-files of the rootKey
	b.addVolumeMount(containerRuntimeManager, v1.VolumeMount{Name: volumeCertVolume,
		MountPath: fmt.Sprintf("%s/%s/%s/", baseCertFilePath, serviceName, microServiceName)})
	b.addVolumeMount(containerRuntimeManager, v1.VolumeMount{Name: volumeStsConfig,
		MountPath: fmt.Sprintf("%s/%s/%s/%s/apple/a", baseCertFilePath, serviceName, microServiceName,
			microServiceName),
		SubPath: "a",
	})
	b.addVolumeMount(containerRuntimeManager, v1.VolumeMount{Name: volumeStsConfig,
		MountPath: fmt.Sprintf("%s/%s/%s/%s/boy/b", baseCertFilePath, serviceName, microServiceName, microServiceName),
		SubPath:   "b",
	})
	b.addVolumeMount(containerRuntimeManager, v1.VolumeMount{Name: volumeStsConfig,
		MountPath: fmt.Sprintf("%s/%s/%s/%s/cat/c", baseCertFilePath, serviceName, microServiceName, microServiceName),
		SubPath:   "c",
	})
	b.addVolumeMount(containerRuntimeManager, v1.VolumeMount{Name: volumeStsConfig,
		MountPath: fmt.Sprintf("%s/%s/%s/%s/dog/d", baseCertFilePath, serviceName, microServiceName, microServiceName),
		SubPath:   "d",
	})
	b.addVolumeMount(containerRuntimeManager, v1.VolumeMount{Name: volumeStsConfig,
		MountPath: fmt.Sprintf("%s/%s/%s/%s.ini", baseCertFilePath, serviceName, microServiceName, microServiceName),
		SubPath:   microServiceName + ".ini",
	})
	b.addVolumeMount(containerRuntimeManager, v1.VolumeMount{Name: volumeStsConfig,
		MountPath: fmt.Sprintf("%s/%s/%s/%s.sts.p12", baseCertFilePath, serviceName, microServiceName,
			microServiceName),
		SubPath: microServiceName + ".sts.p12",
	})
}

func configEnv(b *envBuilder, env map[string]string) {
	for k, v := range env {
		b.addEnvVar(containerDelegate, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
}
