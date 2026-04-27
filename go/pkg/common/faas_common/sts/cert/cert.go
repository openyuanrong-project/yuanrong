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

// Package cert parsing certificate
package cert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"golang.org/x/crypto/pkcs12"
	
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

// LoadCerts - parsing certificate
func LoadCerts() (*x509.CertPool, *tls.Certificate, error) {
	var keyStorePath string
	var err error
	if err != nil {
		return nil, nil, err
	}
	caCertsPool := x509.NewCertPool()
	var bytes []byte
	if err != nil {
		return nil, nil, err
	}
	fileContent, err := os.ReadFile(keyStorePath)
	if err != nil {
		return nil, nil, err
	}
	pemBlocks, err := pkcs12.ToPEM(fileContent, string(bytes))
	utils.ClearByteMemory(fileContent)
	if err != nil {
		return nil, nil, err
	}

	caBytes, certByte, keyByte, err := parseSTSCerts(pemBlocks)
	if err != nil {
		return nil, nil, err

	}
	for _, caByte := range caBytes {
		caCertsPool.AppendCertsFromPEM(caByte)
	}
	tlsCert, err := tls.X509KeyPair(certByte, keyByte)
	utils.ClearByteMemory(certByte)
	utils.ClearByteMemory(keyByte)
	if err != nil {
		return nil, nil, err

	}
	return caCertsPool, &tlsCert, nil
}

func parseSTSCerts(pemBlocks []*pem.Block) ([][]byte, []byte, []byte, error) {
	var certByte, keyByte []byte
	var err error
	var caBytes [][]byte
	for _, pemBlock := range pemBlocks {
		pemEncoded := pem.EncodeToMemory(pemBlock)
		if pemBlock.Type == "PRIVATE KEY" {
			keyByte = pemEncoded
		} else {
			var cert *x509.Certificate
			if cert, err = x509.ParseCertificate(pemBlock.Bytes); err != nil {
				return nil, nil, nil, err
			}
			if cert == nil {
				return nil, nil, nil, fmt.Errorf("parse certificate err: cert is empty")
			}
			if cert.IsCA {
				pemBlock.Headers = map[string]string{}
				caBytes = append(caBytes, pem.EncodeToMemory(pemBlock))
			} else {
				certByte = append(certByte, pemEncoded...)
			}
		}
	}
	if len(caBytes) == 0 {
		return caBytes, certByte, keyByte, fmt.Errorf("ca certs not exists")
	}
	if len(certByte) == 0 {
		return caBytes, certByte, keyByte, fmt.Errorf("certs not exists")
	}
	if len(keyByte) == 0 {
		return caBytes, certByte, keyByte, fmt.Errorf("private key not exists")
	}
	return caBytes, certByte, keyByte, nil
}
