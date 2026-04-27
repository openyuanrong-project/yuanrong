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
	"crypto/x509"
	"encoding/pem"
	"yuanrong.org/kernel/pkg/common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/reader"
)

// NewTLSConfig returns tls.Config with given options
func NewTLSConfig(opts ...Option) *tls.Config {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			// for TLS1.2
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			// for TLS1.3
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
		PreferServerCipherSuites: true,
		Renegotiation:            tls.RenegotiateNever,
	}
	for _, opt := range opts {
		opt.apply(config)
	}
	return config
}

// Option is optional argument for tls.Config
type Option interface {
	apply(*tls.Config)
}

type rootCAOption struct {
	cas *x509.CertPool
}

func (r *rootCAOption) apply(config *tls.Config) {
	config.RootCAs = r.cas
}

// WithRootCAs returns Option that applies root CAs to tls.Config
func WithRootCAs(caFiles ...string) Option {
	rootCAs, err := LoadRootCAs(caFiles...)
	if err != nil {
		log.GetLogger().Warnf("failed to load root ca, err: %s", err.Error())
		rootCAs = nil
	}
	return &rootCAOption{
		cas: rootCAs,
	}
}

type certsOption struct {
	certs []tls.Certificate
}

func (c *certsOption) apply(config *tls.Config) {
	config.Certificates = c.certs
}

// WithCerts returns Option that applies cert file and key file to tls.Config
func WithCerts(certFile, keyFile string) Option {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.GetLogger().Warnf("load cert.pem and key.pem error: %s", err)
		cert = tls.Certificate{}
	}
	return &certsOption{
		certs: []tls.Certificate{cert},
	}
}

// WithCertsByEncryptedKey returns Option that applies cert file and encrypted key file to tls.Config
func WithCertsByEncryptedKey(certFile, keyFile, passPhase string) Option {
	cert := tls.Certificate{}
	certPEM, err := reader.ReadFileWithTimeout(certFile)
	if err != nil {
		return &certsOption{certs: []tls.Certificate{cert}}
	}
	keyPEMBlock, err := getKeyContent(keyFile)
	if err != nil {
		return &certsOption{certs: []tls.Certificate{cert}}
	}
	keyBlock, _ := pem.Decode(keyPEMBlock)
	if keyBlock == nil {
		log.GetLogger().Errorf("failed to decode key file ")
		return &certsOption{certs: []tls.Certificate{cert}}
	}

	if crypto.IsEncryptedPEMBlock(keyBlock) {
		var plainPassPhase []byte
		var err error
		var decrypted string
		if len(passPhase) > 0 {
			decrypted, err = crypto.Decrypt(keyPEMBlock, crypto.GetRootKey())
			plainPassPhase = []byte(decrypted)
			if err != nil {
				log.GetLogger().Errorf("failed to decrypt the ssl passPhase(%d): %s",
					len(passPhase), err.Error())
				return &certsOption{certs: []tls.Certificate{cert}}
			}
			keyData, err := crypto.DecryptPEMBlock(keyBlock, plainPassPhase)
			clearByteMemory(plainPassPhase)
			utils.ClearStringMemory(decrypted)

			if err != nil {
				return &certsOption{certs: []tls.Certificate{cert}}
			}
			// The decryption is successful, then the file is re-encoded to a PEM file
			plainKeyBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyData}
			keyPEMBlock = pem.EncodeToMemory(plainKeyBlock)
		}
	}

	if err != nil {
		return &certsOption{certs: []tls.Certificate{cert}}
	}

	cert, err = tls.X509KeyPair(certPEM, keyPEMBlock)
	if err != nil {
		cert = tls.Certificate{}
	}
	utils.ClearByteMemory(keyPEMBlock)
	return &certsOption{certs: []tls.Certificate{cert}}
}

func getKeyContent(keyFile string) ([]byte, error) {
	var err error
	keyPEMBlock, err := reader.ReadFileWithTimeout(keyFile)

	if err != nil {
		log.GetLogger().Errorf("fialed to SCCDecrypt, err: %s", err.Error())
		return nil, err
	}
	return keyPEMBlock, nil
}

// WithCertsContent returns Option that applies cert content and key content to tls.Config
func WithCertsContent(certContent, keyContent []byte) Option {
	cert, err := tls.X509KeyPair(certContent, keyContent)
	utils.ClearByteMemory(keyContent)
	if err != nil {
		log.GetLogger().Warnf("load cert.pem and key.pem error: %s", err)
		cert = tls.Certificate{}
	}
	return &certsOption{
		certs: []tls.Certificate{cert},
	}
}

type skipVerifyOption struct {
}

func (s *skipVerifyOption) apply(config *tls.Config) {
	config.InsecureSkipVerify = true
}

// WithSkipVerify returns Option that skips to verify certificates
func WithSkipVerify() Option {
	return &skipVerifyOption{}
}

type clientAuthOption struct {
	clientAuthType tls.ClientAuthType
}

func (a *clientAuthOption) apply(config *tls.Config) {
	config.ClientAuth = a.clientAuthType
}

// WithClientAuthType returns Option with client auth strategy
func WithClientAuthType(t tls.ClientAuthType) Option {
	return &clientAuthOption{
		clientAuthType: t,
	}
}

type clientCAOption struct {
	clientCAs *x509.CertPool
}

func (r *clientCAOption) apply(config *tls.Config) {
	config.ClientCAs = r.clientCAs
}

// WithClientCAs returns Option that applies client CAs to tls.Config
func WithClientCAs(caFiles ...string) Option {
	clientCAs, err := LoadRootCAs(caFiles...)
	if err != nil {
		log.GetLogger().Warnf("failed to load client ca, err: %s", err.Error())
		clientCAs = nil
	}
	return &clientCAOption{
		clientCAs: clientCAs,
	}
}

type serverNameOption struct {
	serverName string
}

func (sn *serverNameOption) apply(config *tls.Config) {
	config.ServerName = sn.serverName
}

// WithServerName returns Option that applies server name to tls.Config
func WithServerName(name string) Option {
	return &serverNameOption{
		serverName: name,
	}
}
