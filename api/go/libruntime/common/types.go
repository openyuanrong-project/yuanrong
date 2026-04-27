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

// Package common for types
package common

// AuthContext -
type AuthContext struct {
	ServerAuthEnable   bool   `json:"serverAuthEnable" valid:"optional"`
	RootCertData       []byte `json:"rootCertData" valid:"optional"`
	ModuleCertData     []byte `json:"moduleCertData" valid:"optional"`
	ModuleKeyData      []byte `json:"moduleKeyData" valid:"optional"`
	Token              string `json:"token" valid:"optional"`
	Salt               string `json:"salt" valid:"optional"`
	EnableServerMode   bool   `json:"enableServerMode" valid:"optional"`
	ServerNameOverride string `json:"serverNameOverride" valid:"optional"`
}

// TLSConfOptions defines the struct of TLS config options, which is used to initialize grpc transport credentials.
type TLSConfOptions struct {
	TLSEnable  bool
	RootCAData []byte
	CertData   []byte
	KeyData    []byte
}
