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

// Package types -
package types

// ServiceAccountJwt service account config
type ServiceAccountJwt struct {
	NuwaRuntimeAddr      string `json:"nuwaRuntimeAddr,omitempty"`
	NuwaGatewayAddr      string `json:"nuwaGatewayAddr,omitempty"`
	OauthTokenUrl        string `json:"oauthTokenUrl"`
	ServiceAccountKeyStr string `json:"serviceAccountKey"`
	*ServiceAccount      `json:"-"`
	TlsConfig            *TLSConfig `json:"tlsConfig"`
}

// TLSConfig tls config
type TLSConfig struct {
	HttpsInsecureSkipVerify bool     `json:"httpsInsecureSkipVerify"`
	TlsCipherSuitesStr      []string `json:"tlsCipherSuites"`
	TlsCipherSuites         []uint16 `json:"-"`
}

// ServiceAccount service account config
type ServiceAccount struct {
	PrivateKey string `json:"privateKey"`
	ClientId   int64  `json:"clientId"`
	KeyId      string `json:"keyId"`
	PublicKey  string `json:"publicKey"`
	UserId     int64  `json:"userId"`
	Version    int32  `json:"version"`
}

// NuwaRuntimeInfo contains ers workload info for function
type NuwaRuntimeInfo struct {
	WisecloudRuntimeId     string `json:"wisecloudRuntimeId"`
	WisecloudSite          string `json:"wisecloudSite"`
	WisecloudTenantId      string `json:"wisecloudTenantId"`
	WisecloudApplicationId string `json:"wisecloudApplicationId"`
	WisecloudServiceId     string `json:"wisecloudServiceId"`
	WisecloudEnvironmentId string `json:"wisecloudEnvironmentId"`
	EnvLabel               string `json:"envLabel"`
}

// NuwaColdCreateInstanceReq request to nuwa
type NuwaColdCreateInstanceReq struct {
	RuntimeId   string `json:"runtimeId"`
	RuntimeType string `json:"type"`     // function/microservice
	PoolType    string `json:"poolType"` // java1.8/nodejs/python3
	EnvLabel    string `json:"envLabel"`
	Memory      int64  `json:"memory"`
	CPU         int64  `json:"cpu"`
}

// NuwaDestroyInstanceReq request to nuwa
type NuwaDestroyInstanceReq struct {
	RuntimeId    string `json:"runtimeId"`
	RuntimeType  string `json:"type"`
	InstanceId   string `json:"instanceId"` // podNamespace:podName
	WorkLoadName string `json:"workLoadName"`
}
