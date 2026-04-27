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

package constant

const (
	// DelegateVolumeMountKey is the key for DELEGATE_VOLUME_MOUNTS in CreateOption
	DelegateVolumeMountKey = "DELEGATE_VOLUME_MOUNTS"
	// DelegateInitVolumeMountKey is the key for DELEGATE_INIT_VOLUME_MOUNTS in CreateOption
	DelegateInitVolumeMountKey = "DELEGATE_INIT_VOLUME_MOUNTS"
	// DelegateAgentVolumeMountKey is the key for DELEGATE_AGENT_VOLUME_MOUNTS i
	DelegateAgentVolumeMountKey = "DELEGATE_AGENT_VOLUME_MOUNTS"
	// DelegateVolumesKey is the key for DELEGATE_VOLUMES in CreateOption
	DelegateVolumesKey = "DELEGATE_VOLUMES"
	// DelegateHostAliases is the key for DELEGATE_HOST_ALIASES in CreateOption
	DelegateHostAliases = "DELEGATE_HOST_ALIASES"
	// DelegateDownloadKey is the key for DelegateDownload in CreateOption
	DelegateDownloadKey = "DELEGATE_DOWNLOAD"
	// DelegateBootstrapKey is the key for DelegateStart in CreateOption
	DelegateBootstrapKey = "DELEGATE_BOOTSTRAP"
	// DelegateLayerDownloadKey is the key for DelegateLayerDownload in CreateOption
	DelegateLayerDownloadKey = "DELEGATE_LAYER_DOWNLOAD"
	// DelegateMountKey is the key for DELEGATE_MOUNT in CreateOption
	DelegateMountKey = "DELEGATE_MOUNT"
	// DelegateEncryptKey is the key for DELEGATE_ENCRYPT in CreateOption
	DelegateEncryptKey = "DELEGATE_ENCRYPT"
	// DelegateContainerKey is the key for DELEGATE_CONTAINER in CreateOption
	DelegateContainerKey = "DELEGATE_CONTAINER"
	// DelegateContainerSideCars is the key for DELEGATE_SIDECARS in CreateOption
	DelegateContainerSideCars = "DELEGATE_SIDECARS"
	// DelegateInitContainers is the key for DELEGATE_INIT_CONTAINERS in CreateOption
	DelegateInitContainers = "DELEGATE_INIT_CONTAINERS"
	// DelegatePodAnnotations is used to transfer pod annotations to the kernel during instance creation
	DelegatePodAnnotations = "DELEGATE_POD_ANNOTATIONS"
	// DelegatePodLabels is used to transfer pod labels to the kernel during instance creation
	DelegatePodLabels = "DELEGATE_POD_LABELS"
	// DelegatePodInitLabels -
	DelegatePodInitLabels = "DELEGATE_POD_INIT_LABELS"
	// DelegatePodSeccompProfile is key for DELEGATE_POD_SECCOMP_PROFILE in CreateOption
	DelegatePodSeccompProfile = "DELEGATE_POD_SECCOMP_PROFILE"
	// DelegateInitVolumeMounts is key for DELEGATE_INIT_VOLUME_MOUNTS in CreateOption
	DelegateInitVolumeMounts = "DELEGATE_INIT_VOLUME_MOUNTS"
	// DelegateNuwaRuntimeInfo is key for DELEGATE_NUWA_RUNTIME_INFO in CreateOption
	DelegateNuwaRuntimeInfo = "DELEGATE_NUWA_RUNTIME_INFO"
	// DelegateInitEnv is key for DelegateInitEnv in CreateOption
	DelegateInitEnv = "DELEGATE_INIT_ENV"
	// EnvDelegateEncrypt -
	EnvDelegateEncrypt = "DELEGATE_ENCRYPT"
	// DelegateTolerations is the key for DELEGATE_TOLERATIONS in CreateOption
	DelegateTolerations = "DELEGATE_TOLERATIONS"

	// DelegateRuntimeManagerTag the key of runtime-manager's image tag
	DelegateRuntimeManagerTag = "DELEGATE_RUNTIME_MANAGER"
	// DelegateNodeAffinity is the key for DELEGATE_NODE_AFFINITY in CreateOption
	DelegateNodeAffinity = "DELEGATE_NODE_AFFINITY"

	// DelegateNodeAffinityPolicy -
	DelegateNodeAffinityPolicy = "DELEGATE_NODE_AFFINITY_POLICY"
	// DelegateAffinity -
	DelegateAffinity = "DELEGATE_AFFINITY"
	// DelegateNodeAffinityPolicyCoverage -
	DelegateNodeAffinityPolicyCoverage = "coverage"
	// DelegateNodeAffinityPolicyAggregation -
	DelegateNodeAffinityPolicyAggregation = "aggregation"

	// DelegateNetworkConfig -
	DelegateNetworkConfig = "DELEGATE_NETWORK_CONFIG"

	// InstanceLifeCycle -
	InstanceLifeCycle = "lifecycle"
	// InstanceLifeCycleDetached -
	InstanceLifeCycleDetached = "detached"

	// DelegateDirectoryInfo is the path that will be monitored its disk usage
	DelegateDirectoryInfo = "DELEGATE_DIRECTORY_INFO"
	// DelegateDirectoryQuota is the quota of the path
	DelegateDirectoryQuota = "DELEGATE_DIRECTORY_QUOTA"
	// PostStartExec -
	PostStartExec = "POST_START_EXEC"
	// DelegateEnvVar -
	DelegateEnvVar = "DELEGATE_ENV_VAR"
	// BusinessTypeTypeNote - is used to decribe the instance business type: "Serve", "FaaS", "Actor"
	BusinessTypeTypeNote = "BUSINESS_TYPE_NOTE"
	// FaasInvokeTimeout is function exec timeout
	FaasInvokeTimeout = "INVOKE_TIMEOUT"
	// ImagePullSecrets image pull secrets
	ImagePullSecrets = "IMAGE_PULL_SECRETS"
)
