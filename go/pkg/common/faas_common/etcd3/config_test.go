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
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	clientv3 "go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/sts/cert"
	commontls "yuanrong.org/kernel/pkg/common/faas_common/tls"
	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
)

func TestGetEtcdAuthType(t *testing.T) {
	convey.Convey("tlsAuth", t, func() {
		etcdConfig := EtcdConfig{
			SslEnable: true,
		}
		etcdAuth := GetEtcdAuthType(etcdConfig)
		convey.So(etcdAuth, convey.ShouldResemble, &tlsAuth{
			certFile: etcdConfig.CertFile,
			keyFile:  etcdConfig.KeyFile,
			caFile:   etcdConfig.CaFile,
		})
	})
	convey.Convey("tlsAuth", t, func() {
		etcdConfig := EtcdConfig{
			SslEnable: false,
			Password:  "",
		}
		etcdAuth := GetEtcdAuthType(etcdConfig)
		convey.So(etcdAuth, convey.ShouldResemble, &noAuth{})
	})
	convey.Convey("tlsAuth", t, func() {
		etcdConfig := EtcdConfig{
			SslEnable: false,
			User:      "u123",
			Password:  "p123",
		}
		etcdAuth := GetEtcdAuthType(etcdConfig)
		convey.So(etcdAuth, convey.ShouldResemble, &pwdAuth{
			user:     etcdConfig.User,
			password: etcdConfig.Password,
		})
	})
}

func TestGetEtcdConfig(t *testing.T) {
	defer gomonkey.ApplyFunc(localauth.Decrypt, func(src string) ([]byte, error) {
		return []byte(strings.Clone(src)), nil
	}).Reset()
	convey.Convey("noAuth", t, func() {
		noAuth := &noAuth{}
		cfg, err := noAuth.GetEtcdConfig()
		convey.So(cfg, convey.ShouldResemble, &clientv3.Config{})
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("tlsAuth", t, func() {
		defer gomonkey.ApplyFunc(tls.X509KeyPair, func(certPEMBlock []byte, keyPEMBlock []byte) (tls.Certificate, error) {
			return tls.Certificate{}, nil
		}).Reset()
		defer gomonkey.ApplyFunc(ioutil.ReadFile, func(filename string) ([]byte, error) {
			return []byte{}, nil
		}).Reset()
		defer gomonkey.ApplyFunc(commontls.LoadServerTLSCertificate, func(certFile, keyFile, passPhase, decryptTool string,
			isHTTPS bool) ([]tls.Certificate, error) {
			return nil, nil
		}).Reset()
		tlsAuth := &tlsAuth{
			user:     "root",
			password: string([]byte("123")),
		}
		cfg, err := tlsAuth.GetEtcdConfig()
		convey.So(err, convey.ShouldBeNil)
		convey.So(cfg, convey.ShouldNotBeNil)
	})
	convey.Convey("tlsAuth error", t, func() {
		defer gomonkey.ApplyFunc(localauth.Decrypt, func(src string) ([]byte, error) {
			return nil, errors.New("some error")
		}).Reset()
		tlsAuth := &tlsAuth{}
		_, err := tlsAuth.GetEtcdConfig()
		convey.So(err, convey.ShouldNotBeNil)
		tlsAuth.user, tlsAuth.password = "root", "123"
		_, err = tlsAuth.GetEtcdConfig()
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("pwdAuth", t, func() {
		pwdAuth := &pwdAuth{
			user:     "root",
			password: string([]byte("123")),
		}
		cfg, err := pwdAuth.GetEtcdConfig()
		convey.So(cfg.Password, convey.ShouldEqual, "123")
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("pwdAuth error", t, func() {
		defer gomonkey.ApplyFunc(localauth.Decrypt, func(src string) ([]byte, error) {
			return nil, errors.New("some error")
		}).Reset()
		pwdAuth := &pwdAuth{
			password: string([]byte("123")),
		}
		_, err := pwdAuth.GetEtcdConfig()
		convey.So(err, convey.ShouldNotBeNil)
		pwdAuth.user = "root"
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestBuildStsCfg(t *testing.T) {
	tests := []struct {
		name        string
		wantErr     bool
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1", false, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(cert.LoadCerts, func() (*x509.CertPool, *tls.Certificate,
					error) {
					return &x509.CertPool{}, &tls.Certificate{}, nil
				})})
			return patches
		}},
		{"case2 LoadCerts error", true, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(cert.LoadCerts, func() (*x509.CertPool, *tls.Certificate,
					error) {
					return &x509.CertPool{}, &tls.Certificate{}, errors.New("error")
				})})
			return patches
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			_, err := BuildStsCfg()
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildStsCfg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			patches.ResetAll()
		})
	}
}
