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

// Package sts provide methods for obtaining sensitive information
package sts

const (
	gcmStandardNonceSize = 12
	// CertPath Get certificate interface url
	CertPath = "/mgmt-inner/cms-mgmt/v1/deploy/%s/%s/cert-file"
	// DateKeyPath Get dataKey interface url
	DateKeyPath = "/mgmt-inner/cms-mgmt/v1/data-key"
	// SensitiveConfigPath Get sensitive config interface url
	SensitiveConfigPath = "/mgmt-inner/sts-mgmt/v2/sensitiveconfig"
	// Channel Deployment channel, fill in the service and microservice name corresponding to the BaaS service
	Channel = "HMSClientCloudAccelerateService_HMSCaaSYuanRongWorkerManager"
	// DefaultKeyVersion Default Key Version
	DefaultKeyVersion = "1"
	// MgmtServiceName - Mgmt ServiceName
	MgmtServiceName = "SecurityMgmtService"
	// MgmtMicroServiceName - Mgmt MicroServiceName
	MgmtMicroServiceName = "SecurityMgmtMicroService"

	baseCertFilePath = "/opt/certs"

	maxConfigIDPerRequest = 50

	privateKeyByteLen = 32
	// ECDHKeyLen - ECDH key len
	ECDHKeyLen = 16
)

// CertResponse - certificate interface resp
type CertResponse struct {
	CertFileData   string   `json:"certFileData"`
	PrivateKeyData string   `json:"privateKeyData"`
	Format         string   `json:"format"`
	Password       Password `json:"password"`
}

// Password -
type Password struct {
	ProtectKey string `json:"protectKey"`
	RootKey    string `json:"rootkey"`
	WorkKey    string `json:"workkey"`
	CipherPwd  string `json:"cipherPwd"`
}

// RootKey - in sts cert
type RootKey struct {
	Apple string `json:"apple"`
	Boy   string `json:"boy"`
	Cat   string `json:"cat"`
	Dog   string `json:"dog"`
}

// DataKeyReq - DataKey interface req
type DataKeyReq struct {
	Algo      string `json:"algo"`
	AppID     string `json:"appId"`
	PublicKey string `json:"publicKey"`
}

// DataKeyResponse - DataKey interface resp
type DataKeyResponse struct {
	DataKey   string `json:"dataKey,omitempty"`
	Version   int64  `json:"version,omitempty"`
	PublicKey string `json:"publicKey,omitempty"`
}

// EcdhKeyPair - ECDH key
type EcdhKeyPair struct {
	PublicKey  []byte `json:"publicKey,omitempty"`
	PrivateKey []byte `json:"privateKey,omitempty"`
}

type configIDsReq struct {
	ConfigIds []string `json:"configIds"`
}

// SensitiveConfigResponse - sensitive config interface resp
type SensitiveConfigResponse struct {
	Status             string              `json:"status"`
	Message            string              `json:"message"`
	ConfigItems        []ConfigItem        `json:"configItems"`
	MissingConfigItems []MissingConfigItem `json:"missingConfigItems"`
}

// ConfigItem - list of sensitive configurations that can successfully obtain ciphertext
type ConfigItem struct {
	ConfigID    string `json:"configId"`
	ConfigValue string `json:"configValue"`
}

// MissingConfigItem - List of sensitive configuration coordinates that cannot be queried
type MissingConfigItem struct {
	ConfigID string `json:"configId"`
}

// Info - sts cert info
type Info struct {
	RootKey *RootKey
	Cert    *CertResponse
	Ini     string
}
