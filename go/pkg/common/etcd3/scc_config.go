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

// Package etcd3 implements crud and watch operations based etcd clientv3
package etcd3

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

type clientTlsAuth struct {
	cerfile        string
	keyfile        string
	cafile         string
	passphrasefile string
}

func (e *EtcdConfig) getEtcdAuthType() etcdAuth {
	if e.AuthType == "TLS" {
		return &clientTlsAuth{
			cerfile:        e.CertFile,
			keyfile:        e.KeyFile,
			cafile:         e.CaFile,
			passphrasefile: e.PassphraseFile,
		}
	}
	if e.SslEnable {
		return &tlsAuth{
			cerfile: e.CertFile,
			keyfile: e.KeyFile,
			cafile:  e.CaFile,
		}
	}
	if e.User == "" || e.Passwd == "" {
		return &noAuth{}
	}
	return &pwdAuth{
		user:   e.User,
		passWd: []byte(e.Passwd),
	}
}

func (c *clientTlsAuth) getEtcdConfig() (*clientv3.Config, error) {
	var keyPwd []byte
	if _, err := os.Stat(c.passphrasefile); err == nil {
		keyPwd, err = ioutil.ReadFile(c.passphrasefile)
		if err != nil {
			log.GetLogger().Errorf("failed to read passphrasefile, %s", err.Error())
			return nil, err
		}
		if crypto.SCCInitialized() {
			pwd, err := crypto.SCCDecrypt(keyPwd)
			if err != nil {
				log.GetLogger().Errorf("failed to decrypt passphrasefile, %s", err.Error())
				return nil, err
			}
			keyPwd = []byte(pwd)
		}
	}

	certPEM, err := ioutil.ReadFile(c.cerfile)
	if err != nil {
		log.GetLogger().Errorf("failed to read cert file: %s", err.Error())
		return nil, err
	}

	caCertPEM, err := ioutil.ReadFile(c.cafile)
	if err != nil {
		log.GetLogger().Errorf("failed to read ca file: %s", err.Error())
		return nil, err
	}

	encryptedKeyPEM, err := ioutil.ReadFile(c.keyfile)
	if err != nil {
		log.GetLogger().Errorf("failed to read key file: %s", err.Error())
		return nil, err
	}

	keyBlock, _ := pem.Decode(encryptedKeyPEM)
	if keyBlock == nil {
		log.GetLogger().Errorf("failed to decode key PEM block")
		return nil, err
	}

	keyDER, err := crypto.DecryptPEMBlock(keyBlock, keyPwd)
	if err != nil {
		log.GetLogger().Errorf("failed to decrypt key: %s", err.Error())
		return nil, err
	}

	key, err := x509.ParsePKCS1PrivateKey(keyDER)
	if err != nil {
		log.GetLogger().Errorf("failed to parse private key: %s", err.Error())
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCertPEM) {
		log.GetLogger().Errorf("failed to append CA certificate")
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	clientCert, err := tls.X509KeyPair(certPEM, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	if err != nil {
		log.GetLogger().Errorf("failed to create client certificate: %s", err.Error())
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return &clientv3.Config{
		TLS: tlsConfig,
	}, nil
}

func (n *clientTlsAuth) renewToken(client *clientv3.Client, stop chan struct{}) {
	return
}
