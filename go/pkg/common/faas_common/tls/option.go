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

// Package tls -
package tls

import (
	"crypto/tls"
	"crypto/x509"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

const (
	// DefaultCAFile is the default file for tls client
	DefaultCAFile = "/home/sn/resource/ca/ca.pem"
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

type skipVerifyOption struct {
}

func (s *skipVerifyOption) apply(config *tls.Config) {
	config.InsecureSkipVerify = true
}

// WithSkipVerify returns Option that skips to verify certificates
func WithSkipVerify() Option {
	return &skipVerifyOption{}
}
