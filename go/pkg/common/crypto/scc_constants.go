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

// Package crypto for auth
package crypto

import (
	"sync"
)

var (
	sccInitialized bool = false
	m              sync.RWMutex
)

const (
	// Aes128Gcm  -
	Aes128Gcm = "AES128_GCM"
	// Aes256Gcm  -
	Aes256Gcm = "AES256_GCM"
	// Aes256Cbc  -
	Aes256Cbc = "AES256_CBC"
	// Sm4Cbc     -
	Sm4Cbc = "SM4_CBC"
	// Sm4Ctr     -
	Sm4Ctr = "SM4_CTR"
)

// SccConfig -
type SccConfig struct {
	Enable        bool   `json:"enable" valid:"optional"`
	Algorithm     string `json:"algorithm" valid:"optional"`
	SccConfigPath string `json:"sccConfigPath" valid:"optional"`
}
