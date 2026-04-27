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
	"errors"
	"os"

	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/sts"
	"yuanrong.org/kernel/pkg/common/faas_common/sts/cert"
	commontls "yuanrong.org/kernel/pkg/common/faas_common/tls"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

// EtcdAuth etcd authentication interface
type EtcdAuth interface {
	GetEtcdConfig() (*clientv3.Config, error)
}

type noAuth struct {
}

type tlsAuth struct {
	caFile   string
	certFile string
	keyFile  string
	user     string
	password string
}

type pwdAuth struct {
	user     string
	password string
}

// GetEtcdAuthType etcd authentication type
func GetEtcdAuthType(etcdConfig EtcdConfig) EtcdAuth {
	if etcdConfig.SslEnable {
		if os.Getenv(sts.EnvSTSEnable) == "true" {
			return &tlsAuth{}
		}
		return &tlsAuth{
			caFile:   etcdConfig.CaFile,
			certFile: etcdConfig.CertFile,
			keyFile:  etcdConfig.KeyFile,
		}
	}
	if etcdConfig.User == "" || etcdConfig.Password == "" {
		return &noAuth{}
	}
	return &pwdAuth{
		user:     etcdConfig.User,
		password: etcdConfig.Password,
	}
}

func (n *noAuth) GetEtcdConfig() (*clientv3.Config, error) {
	return &clientv3.Config{}, nil
}

func (t *tlsAuth) GetEtcdConfig() (*clientv3.Config, error) {
	if os.Getenv(sts.EnvSTSEnable) == "true" {
		return BuildStsCfg()
	}
	pool, err := commontls.GetX509CACertPool(t.caFile)
	if err != nil {
		log.GetLogger().Errorf("failed to getX509CACertPool: %s", err.Error())
		return nil, err
	}

	var certs []tls.Certificate
	if certs, err = commontls.LoadServerTLSCertificate(t.certFile, t.keyFile, "", "LOCAL", false); err != nil {
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
	if len(t.user) != 0 && len(t.password) != 0 {
		pwd, err := localauth.Decrypt(t.password)
		if err != nil {
			log.GetLogger().Errorf("failed to decrypt etcd config with error %s", err)
			return nil, err
		}
		cfg.Username = t.user
		cfg.Password = string(pwd)
		utils.ClearStringMemory(t.password)
	}
	return cfg, nil
}

func (p *pwdAuth) GetEtcdConfig() (*clientv3.Config, error) {
	if len(p.user) == 0 || len(p.password) == 0 {
		return nil, errors.New("etcd user or password is empty")
	}
	pwd, err := localauth.Decrypt(p.password)
	if err != nil {
		log.GetLogger().Errorf("failed to decrypt etcd config with error %s", err)
		return nil, err
	}
	cfg := &clientv3.Config{
		Username: p.user,
		Password: string(pwd),
	}
	utils.ClearStringMemory(p.password)
	return cfg, nil
}

// BuildStsCfg - Construct tlsConfig from sts p12
func BuildStsCfg() (*clientv3.Config, error) {
	caCertsPool, tlsCert, err := cert.LoadCerts()
	if err != nil {
		log.GetLogger().Errorf("failed to get X509CACertPool and TLSCertificate: %s", err.Error())
		return nil, err
	}

	clientAuthMode := tls.NoClientCert
	tlsConfig := &clientv3.Config{
		TLS: &tls.Config{
			RootCAs:      caCertsPool,
			Certificates: []tls.Certificate{*tlsCert},
			ClientAuth:   clientAuthMode,
		},
	}
	return tlsConfig, nil
}
