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
	"context"
	"crypto/tls"
	"crypto/x509"

	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/apimachinery/pkg/util/wait"

	"yuanrong.org/kernel/pkg/common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commontls "yuanrong.org/kernel/pkg/common/tls"
)

// EtcdConfig the info to get function instance
type EtcdConfig struct {
	Servers        []string `json:"servers" yaml:"servers"  valid:"required"`
	User           string   `json:"user" yaml:"user"  valid:"optional"`
	Passwd         string   `json:"password" yaml:"password"  valid:"optional"`
	AuthType       string   `json:"authType" yaml:"authType" valid:"optional"`
	SslEnable      bool     `json:"sslEnable,omitempty" yaml:"sslEnable,omitempty" valid:"optional"`
	LimitRate      int      `json:"limitRate,omitempty" yaml:"limitRate,omitempty" valid:"optional"`
	LimitBurst     int      `json:"limitBurst,omitempty" yaml:"limitBurst,omitempty" valid:"optional"`
	LimitTimeout   int      `json:"limitTimeout,omitempty" yaml:"limitTimeout,omitempty" valid:"optional"`
	LeaseTTL       int64    `json:"leaseTTL,omitempty" yaml:"leaseTTL,omitempty" valid:"optional"`
	RenewTTL       int64    `json:"renewTTL,omitempty" yaml:"renewTTL,omitempty" valid:"optional"`
	CaFile         string   `json:"cafile,omitempty" yaml:"cafile,omitempty" valid:"optional"`
	CertFile       string   `json:"certfile,omitempty" yaml:"certfile,omitempty" valid:"optional"`
	KeyFile        string   `json:"keyfile,omitempty" yaml:"keyfile,omitempty" valid:"optional"`
	PassphraseFile string   `json:"passphraseFile,omitempty" yaml:"passphraseFile,omitempty" valid:"optional"`
	AZPrefix       string   `json:"azPrefix,omitempty" yaml:"azPrefix,omitempty" valid:"optional"`
	// DisableSync will not run the sync method to avoid endpoints being replaced by the domain name, default is FALSE
	DisableSync bool
}

// GetETCDCertificatePath get the certificate path from tlsConfig.
func GetETCDCertificatePath(config EtcdConfig, tlsConfig commontls.MutualTLSConfig) EtcdConfig {
	if !config.SslEnable {
		return config
	}
	config.CaFile = tlsConfig.RootCAFile
	config.CertFile = tlsConfig.ModuleCertFile
	config.KeyFile = tlsConfig.ModuleKeyFile
	return config
}

type etcdAuth interface {
	getEtcdConfig() (*clientv3.Config, error)
	renewToken(client *clientv3.Client, stop chan struct{})
}

type noAuth struct {
}

type tlsAuth struct {
	cerfile string
	keyfile string
	cafile  string
}

type pwdAuth struct {
	user   string
	passWd []byte
}

// this support no tls so we can leave scc dependencies behind
func (e *EtcdConfig) getEtcdAuthTypeNoTLS() etcdAuth {
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

func (n *noAuth) getEtcdConfig() (*clientv3.Config, error) {
	return &clientv3.Config{}, nil
}

func (n *noAuth) renewToken(client *clientv3.Client, stop chan struct{}) {
	return
}

func (t *tlsAuth) getEtcdConfig() (*clientv3.Config, error) {
	var pool *x509.CertPool
	pool, err := commontls.GetX509CACertPool(t.cafile)
	if err != nil {
		log.GetLogger().Errorf("failed to getX509CACertPool: %s", err.Error())
		return nil, err
	}

	var certs []tls.Certificate
	certs, err = commontls.LoadServerTLSCertificate(t.cerfile, t.keyfile)
	if err != nil {
		log.GetLogger().Errorf("failed to loadServerTLSCertificate: %s", err.Error())
		return nil, err
	}

	clientAuthMode := tls.NoClientCert
	cfg := &clientv3.Config{
		TLS: &tls.Config{
			RootCAs:      pool,
			Certificates: certs,
			ClientAuth:   clientAuthMode,
		},
	}
	return cfg, nil
}

func (t *tlsAuth) renewToken(client *clientv3.Client, stop chan struct{}) {
	return
}

func (p *pwdAuth) getEtcdConfig() (*clientv3.Config, error) {
	passWd, err := crypto.Decrypt(p.passWd, crypto.GetRootKey())
	if err != nil {
		log.GetLogger().Errorf("failed to decrypt, error:%s", err.Error())
		return nil, err
	}
	cfg := &clientv3.Config{
		Username: p.user,
		Password: passWd,
	}
	return cfg, nil
}

// renewToken can keep client token not expired, because
// etcd server will renew simple token TTL if token is not expired.
func (p *pwdAuth) renewToken(client *clientv3.Client, stop chan struct{}) {
	if stop == nil {
		log.GetLogger().Errorf("stop chan is nil")
		return
	}
	wait.Until(func() {
		ctx, cancel := context.WithTimeout(context.Background(), DurationContextTimeout)
		_, err := client.Get(ctx, renewKey)
		cancel()
		if err != nil {
			log.GetLogger().Warnf("renew token error: %s", err.Error())
		}
	}, tokenRenewTTL, stop)
	log.GetLogger().Infof("stopped to renew token")
}
