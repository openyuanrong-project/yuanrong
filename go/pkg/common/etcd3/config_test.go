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

package etcd3

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey"
	. "github.com/smartystreets/goconvey/convey"

	"yuanrong.org/kernel/pkg/common/crypto"
	commontls "yuanrong.org/kernel/pkg/common/tls"
)

func TestGetETCDCertificatePath(t *testing.T) {
	Convey("Test ssl enable", t, func() {
		config := EtcdConfig{
			SslEnable: true,
			CaFile:    "",
			CertFile:  "",
			KeyFile:   "",
		}
		tlsConfig := commontls.MutualTLSConfig{
			RootCAFile:     "xxx.ca",
			ModuleCertFile: "xxx.cert",
			ModuleKeyFile:  "xxx.key",
		}
		etcdConfig := GetETCDCertificatePath(config, tlsConfig)
		So(etcdConfig.CaFile, ShouldEqual, tlsConfig.RootCAFile)
		So(etcdConfig.CertFile, ShouldEqual, tlsConfig.ModuleCertFile)
		So(etcdConfig.KeyFile, ShouldEqual, tlsConfig.ModuleKeyFile)
	})
	Convey("Test ssl disable", t, func() {
		config := EtcdConfig{
			SslEnable: false,
			CaFile:    "",
			CertFile:  "",
			KeyFile:   "",
		}
		tlsConfig := commontls.MutualTLSConfig{
			RootCAFile:     "xxx.ca",
			ModuleCertFile: "xxx.cert",
			ModuleKeyFile:  "xxx.key",
		}
		etcdConfig := GetETCDCertificatePath(config, tlsConfig)
		So(etcdConfig.CaFile, ShouldNotEqual, tlsConfig.RootCAFile)
		So(etcdConfig.CertFile, ShouldNotEqual, tlsConfig.ModuleCertFile)
		So(etcdConfig.KeyFile, ShouldNotEqual, tlsConfig.ModuleKeyFile)
	})
}

func TestGetEtcdConfigTlsAuth(t *testing.T) {
	tlsAuth := &tlsAuth{}
	Convey("GetX509CACertPool success", t, func() {
		patch := gomonkey.ApplyFunc(commontls.GetX509CACertPool, func(string) (*x509.CertPool, error) {
			return nil, nil
		})
		defer patch.Reset()
		Convey("LoadServerTLSCertificate success", func() {
			patch := gomonkey.ApplyFunc(commontls.LoadServerTLSCertificate, func(string, string) ([]tls.Certificate, error) {
				return nil, nil
			})
			defer patch.Reset()
			_, err := tlsAuth.getEtcdConfig()
			So(err, ShouldBeNil)
		})
		Convey("LoadServerTLSCertificate fail", func() {
			patch := gomonkey.ApplyFunc(commontls.LoadServerTLSCertificate, func(string, string) ([]tls.Certificate, error) {
				return nil, errors.New("LoadServerTLSCertificate fail")
			})
			defer patch.Reset()
			_, err := tlsAuth.getEtcdConfig()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestRenewTokenTlsAuth(t *testing.T) {
	tlsAuth := &tlsAuth{}
	cli := &EtcdClient{}
	stopCh := make(chan struct{})

	tlsAuth.renewToken(cli.Client, stopCh)
}

func TestRenewTokenPwdAuth(t *testing.T) {
	pwdAuth := &pwdAuth{}
	cli := &EtcdClient{}

	pwdAuth.renewToken(cli.Client, nil)
}

func TestGetEtcdConfigPwdAuth(t *testing.T) {
	pwdAuth := &pwdAuth{}
	Convey("Decrypt success", t, func() {
		patch := gomonkey.ApplyFunc(crypto.Decrypt, func([]byte, []byte) (string, error) {
			return "", nil
		})
		defer patch.Reset()
		_, err := pwdAuth.getEtcdConfig()
		So(err, ShouldBeNil)
	})
	Convey("Decrypt fail", t, func() {
		patch := gomonkey.ApplyFunc(crypto.Decrypt, func([]byte, []byte) (string, error) {
			return "", errors.New("decrypt fail")
		})
		defer patch.Reset()
		_, err := pwdAuth.getEtcdConfig()
		So(err, ShouldNotBeNil)
	})
}
