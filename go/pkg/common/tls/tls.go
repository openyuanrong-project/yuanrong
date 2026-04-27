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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/reader"
)

// MutualTLSConfig indicates tls config
type MutualTLSConfig struct {
	TLSEnable      bool   `json:"tlsEnable" yaml:"tlsEnable" valid:"optional"`
	RootCAFile     string `json:"rootCAFile" yaml:"rootCAFile" valid:"optional"`
	ModuleCertFile string `json:"moduleCertFile" yaml:"moduleCertFile" valid:"optional"`
	ModuleKeyFile  string `json:"moduleKeyFile" yaml:"moduleKeyFile" valid:"optional"`
	ServerName     string `json:"serverName" yaml:"serverName" valid:"optional"`
	SecretName     string `json:"secretName" yaml:"secretName" valid:"optional"`
	PwdFile        string `json:"pwdFile" yaml:"pwdFile" valid:"optional"`
	DecryptTool    string `json:"sslDecryptTool" yaml:"sslDecryptTool" valid:"optional"`
}

// MutualSSLConfig indicates ssl config
type MutualSSLConfig struct {
	SSLEnable      bool   `json:"sslEnable" yaml:"sslEnable" valid:"optional"`
	RootCAFile     string `json:"rootCAFile" yaml:"rootCAFile" valid:"optional"`
	ModuleCertFile string `json:"moduleCertFile" yaml:"moduleCertFile" valid:"optional"`
	ModuleKeyFile  string `json:"moduleKeyFile" yaml:"moduleKeyFile" valid:"optional"`
	ServerName     string `json:"serverName" yaml:"serverName" valid:"optional"`
	PwdFile        string `json:"pwdFile" yaml:"pwdFile" valid:"optional"`
	DecryptTool    string `json:"sslDecryptTool" yaml:"sslDecryptTool" valid:"optional"`
}

// BuildClientTLSConfOpts is to build an option array for mostly used client tlsConf
func BuildClientTLSConfOpts(mutualConf MutualTLSConfig) []Option {
	var opts []Option
	passPhase, err := reader.ReadFileWithTimeout(mutualConf.PwdFile)
	if err != nil {
		log.GetLogger().Errorf("failed to read file PwdFile: %s", err.Error())
		return opts
	}
	opts = append(opts, WithRootCAs(mutualConf.RootCAFile),
		WithCertsByEncryptedKey(mutualConf.ModuleCertFile, mutualConf.ModuleKeyFile,
			string(passPhase)),
		WithServerName(mutualConf.ServerName))
	return opts
}

// BuildServerTLSConfOpts is to build an option array for mostly used server tlsConf
func BuildServerTLSConfOpts(mutualConf MutualTLSConfig) []Option {
	var opts []Option
	var passPhase []byte
	var err error
	if mutualConf.PwdFile != "" {
		passPhase, err = reader.ReadFileWithTimeout(mutualConf.PwdFile)
		if err != nil {
			log.GetLogger().Errorf("failed to read file PwdFile: %s", err.Error())
			return opts
		}
	}
	opts = append(opts, WithRootCAs(mutualConf.RootCAFile),
		WithCertsByEncryptedKey(mutualConf.ModuleCertFile, mutualConf.ModuleKeyFile,
			string(passPhase)),
		WithClientCAs(mutualConf.RootCAFile),
		WithClientAuthType(tls.RequireAndVerifyClientCert))
	return opts
}

// LoadRootCAs returns system cert pool with caFiles added
func LoadRootCAs(caFiles ...string) (*x509.CertPool, error) {
	rootCAs := x509.NewCertPool()
	for _, file := range caFiles {
		cert, err := reader.ReadFileWithTimeout(file)
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
