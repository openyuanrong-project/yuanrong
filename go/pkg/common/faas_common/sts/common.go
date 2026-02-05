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

// Package sts -
package sts

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/core/v1"

	"yuanrong.org/kernel/pkg/common/faas_common/tls"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

// SecretConfig -
type SecretConfig struct{}

const (
	// FaasfrontendName -
	FaasfrontendName = "faasfrontend"
	// FaaSSchedulerName -
	FaaSSchedulerName      = "faasscheduler"
	mountPath              = "/opt/certs/HMSClientCloudAccelerateService/HMSCaaSYuanRongWorker/"
	faasSchedulerMountPath = "/opt/certs/HMSClientCloudAccelerateService/HMSCaaSYuanRongWorkerManager/"
	// HTTPSMountPath mount https certs
	HTTPSMountPath = "/home/sn/resource/https"
	// LocalSecretMountPath mount local secrets
	LocalSecretMountPath = "/home/sn/resource/cipher"
)

var readOnlyVolumeMode int32 = 0440

// ConfigVolume -
func (u *SecretConfig) ConfigVolume(b *utils.VolumeBuilder) {
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-config",
		MountPath: mountPath + "HMSCaaSYuanRongWorker/apple/a",
		SubPath:   "a",
	})
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-config",
		MountPath: mountPath + "HMSCaaSYuanRongWorker/boy/b",
		SubPath:   "b",
	})
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-config",
		MountPath: mountPath + "HMSCaaSYuanRongWorker/cat/c",
		SubPath:   "c",
	})
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-config",
		MountPath: mountPath + "HMSCaaSYuanRongWorker/dog/d",
		SubPath:   "d",
	})
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-config",
		MountPath: mountPath + "HMSCaaSYuanRongWorker.ini",
		SubPath:   "HMSCaaSYuanRongWorker.ini",
	})
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-config",
		MountPath: mountPath + "HMSCaaSYuanRongWorker.sts.p12",
		SubPath:   "HMSCaaSYuanRongWorker.sts.p12",
	})
}

// ConfigFaasSchedulerVolume -
func (u *SecretConfig) ConfigFaasSchedulerVolume(b *utils.VolumeBuilder) {
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-workermanager-config",
		MountPath: faasSchedulerMountPath + "HMSCaaSYuanRongWorkerManager/apple/a",
		SubPath:   "a",
	})
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-workermanager-config",
		MountPath: faasSchedulerMountPath + "HMSCaaSYuanRongWorkerManager/boy/b",
		SubPath:   "b",
	})
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-workermanager-config",
		MountPath: faasSchedulerMountPath + "HMSCaaSYuanRongWorkerManager/cat/c",
		SubPath:   "c",
	})
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-workermanager-config",
		MountPath: faasSchedulerMountPath + "HMSCaaSYuanRongWorkerManager/dog/d",
		SubPath:   "d",
	})
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-workermanager-config",
		MountPath: faasSchedulerMountPath + "HMSCaaSYuanRongWorkerManager.ini",
		SubPath:   "HMSCaaSYuanRongWorkerManager.ini",
	})
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "sts-workermanager-config",
		MountPath: faasSchedulerMountPath + "HMSCaaSYuanRongWorkerManager.sts.p12",
		SubPath:   "HMSCaaSYuanRongWorkerManager.sts.p12",
	})
}

// ConfigHTTPSAndLocalSecretVolume -
func (u *SecretConfig) ConfigHTTPSAndLocalSecretVolume(b *utils.VolumeBuilder, httpsConfig tls.InternalHTTPSConfig) {
	b.AddVolume(buildVolumeOfSecretSource("https", httpsConfig.SecretName))
	b.AddVolumeMount(utils.ContainerRuntimeManager, v1.VolumeMount{
		Name:      "https",
		MountPath: httpsConfig.SSLBasePath,
	})
}

func buildVolumeOfSecretSource(name string, secretName string) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				DefaultMode: &readOnlyVolumeMode,
				SecretName:  secretName,
			},
		},
	}
}

// GenerateSecretVolumeMounts -
func GenerateSecretVolumeMounts(systemFunctionName string, builder *utils.VolumeBuilder) ([]byte, error) {
	if builder == nil {
		return nil, fmt.Errorf("sts volume builder is nil")
	}
	sc := &SecretConfig{}
	if systemFunctionName == FaaSSchedulerName {
		sc.ConfigFaasSchedulerVolume(builder)
	} else {
		sc.ConfigVolume(builder)
	}
	bytesData, err := json.Marshal(builder.Mounts[utils.ContainerRuntimeManager])
	if err != nil {
		return nil, err
	}
	return bytesData, nil
}

// CustomKeyProvider -
type CustomKeyProvider struct {
	key      []byte
	tenantID string
}

// NewCustomKeyProvider -
func NewCustomKeyProvider(tenantID string, key []byte) *CustomKeyProvider {
	return &CustomKeyProvider{tenantID: tenantID, key: key}
}

// GenerateHTTPSAndLocalSecretVolumeMounts -
func GenerateHTTPSAndLocalSecretVolumeMounts(
	httpsConfig tls.InternalHTTPSConfig, builder *utils.VolumeBuilder) (string, string, error) {
	if builder == nil {
		return "", "", fmt.Errorf("https volume builder is nil")
	}
	sc := &SecretConfig{}
	sc.ConfigHTTPSAndLocalSecretVolume(builder, httpsConfig)

	volumesData, err := json.Marshal(builder.Volumes)
	if err != nil {
		return "", "", err
	}
	volumesMountData, err := json.Marshal(builder.Mounts[utils.ContainerRuntimeManager])
	if err != nil {
		return "", "", err
	}
	return string(volumesData), string(volumesMountData), nil
}
