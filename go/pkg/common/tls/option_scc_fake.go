//go:build !cryptoapi
// +build !cryptoapi

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

package tls

import (
	"crypto/tls"
	"encoding/pem"

	"yuanrong.org/kernel/pkg/common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/reader"
)

// WithCertsByEncryptedKeyScc returns Option that applies cert file and encrypted key file to tls.Config
func WithCertsByEncryptedKeyScc(certFile, keyFile, passPhase string) Option {
	cert := tls.Certificate{}
	fakeCertPEM, err := reader.ReadFileWithTimeout(certFile)

	if err != nil {
		return &certsOption{certs: []tls.Certificate{cert}}
	}
	fakeKeyPEMBlock, err := reader.ReadFileWithTimeout(keyFile)

	if err != nil {
		return &certsOption{certs: []tls.Certificate{cert}}
	}
	fakeKeyBlock, _ := pem.Decode(fakeKeyPEMBlock)
	if fakeKeyBlock == nil {
		log.GetLogger().Errorf("failed to decode key file ")
		return &certsOption{certs: []tls.Certificate{cert}}
	}

	if crypto.IsEncryptedPEMBlock(fakeKeyBlock) {
		var plainPassPhase []byte
		var err error
		var decrypted string
		if len(passPhase) > 0 {
			decrypted, err = crypto.SCCDecrypt([]byte(passPhase))
			plainPassPhase = []byte(decrypted)
			if err != nil {
				log.GetLogger().Errorf("failed to decrypt the ssl passPhase(%d): %s",
					len(passPhase), err.Error())
				return &certsOption{certs: []tls.Certificate{cert}}
			}
			keyData, err := crypto.DecryptPEMBlock(fakeKeyBlock, plainPassPhase)
			clearByteMemory(plainPassPhase)
			utils.ClearStringMemory(decrypted)

			if err != nil {
				return &certsOption{certs: []tls.Certificate{cert}}
			}
			// The decryption is successful, then the file is re-encoded to a PEM file
			plainKeyBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyData}
			fakeKeyPEMBlock = pem.EncodeToMemory(plainKeyBlock)
		}
	}

	if err != nil {
		return &certsOption{certs: []tls.Certificate{cert}}
	}

	cert, err = tls.X509KeyPair(fakeCertPEM, fakeKeyPEMBlock)
	if err != nil {
		cert = tls.Certificate{}
	}
	utils.ClearByteMemory(fakeKeyPEMBlock)
	return &certsOption{certs: []tls.Certificate{cert}}
}
