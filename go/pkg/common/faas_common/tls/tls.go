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

// Package tls provides tls utils
package tls

import (
	"crypto/x509"
	"errors"
	"io/ioutil"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// LoadRootCAs returns system cert pool with caFiles added
func LoadRootCAs(caFiles ...string) (*x509.CertPool, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	for _, file := range caFiles {
		cert, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		if !rootCAs.AppendCertsFromPEM(cert) {
			return nil, err
		}
	}
	return rootCAs, nil
}

// VerifyCert Used to verity the server certificate
func VerifyCert(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	certs := make([]*x509.Certificate, len(rawCerts))
	if len(certs) == 0 {
		log.GetLogger().Errorf("cert number is 0")
		return errors.New("cert number is 0")
	}
	opts := x509.VerifyOptions{
		Roots:         tlsConfig.ClientCAs,
		CurrentTime:   time.Now(),
		DNSName:       "",
		Intermediates: x509.NewCertPool(),
	}
	for i, asn1Data := range rawCerts {
		cert, err := x509.ParseCertificate(asn1Data)
		if err != nil {
			log.GetLogger().Errorf("failed to parse certificate from server: %s", err.Error())
			return err
		}
		certs[i] = cert
		if i == 0 {
			continue
		}
		opts.Intermediates.AddCert(cert)
	}
	_, err := certs[0].Verify(opts)
	return err
}
