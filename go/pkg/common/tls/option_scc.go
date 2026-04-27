//go:build cryptoapi
// +build cryptoapi

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

const (
	rsaPrivateKey = "RSA PRIVATE KEY"
)

// WithCertsByEncryptedKeyScc returns Option that applies cert file and encrypted key file to tls.Config
func WithCertsByEncryptedKeyScc(certFile, keyFile, passPhaseStr string) Option {
	cert := tls.Certificate{}
	certPEM, err := reader.ReadFileWithTimeout(certFile)

	if err != nil {
		log.GetLogger().Errorf("failed to read file with timeout, err: %s", err.Error())
		return &certsOption{certs: []tls.Certificate{cert}}
	}
	keyPEMBlock, err := reader.ReadFileWithTimeout(keyFile)

	if err != nil {
		return &certsOption{certs: []tls.Certificate{cert}}
	}
	keyBlock, _ := pem.Decode(keyPEMBlock)
	if keyBlock == nil {
		log.GetLogger().Errorf("failed to decode key file.")
		return &certsOption{certs: []tls.Certificate{cert}}
	}

	if crypto.IsEncryptedPEMBlock(keyBlock) {
		var plainPassPhase []byte
		var err error
		var decrypted string
		if len(passPhaseStr) > 0 {
			decrypted, err = crypto.SCCDecrypt([]byte(passPhaseStr))
			plainPassPhase = []byte(decrypted)
			if err != nil {
				log.GetLogger().Errorf("failed to decrypt the ssl passPhase(%d), err: %s",
					len(passPhaseStr), err.Error())
				return &certsOption{certs: []tls.Certificate{cert}}
			}
			keyData, err := crypto.DecryptPEMBlock(keyBlock, plainPassPhase)
			clearByteMemory(plainPassPhase)
			utils.ClearStringMemory(decrypted)

			if err != nil {
				log.GetLogger().Errorf("failed to decrypt PEM Block, err: %s", err.Error())
				return &certsOption{certs: []tls.Certificate{cert}}
			}
			// The decryption is successful, then the file is re-encoded to a PEM file
			plainKeyBlock := &pem.Block{Type: rsaPrivateKey, Bytes: keyData}
			keyPEMBlock = pem.EncodeToMemory(plainKeyBlock)
		}
	}

	if err != nil {
		return &certsOption{certs: []tls.Certificate{cert}}
	}

	cert, err = tls.X509KeyPair(certPEM, keyPEMBlock)
	if err != nil {
		log.GetLogger().Errorf("failed to build X509KeyPair, err: %s", err.Error())
		cert = tls.Certificate{}
	}
	utils.ClearByteMemory(keyPEMBlock)
	return &certsOption{certs: []tls.Certificate{cert}}
}
