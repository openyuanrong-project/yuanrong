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
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/crypto"
)

func TestGetURLScheme(t *testing.T) {
	if "https" != GetURLScheme(true) {
		t.Error("GetURLScheme failed")
	}
	if "http" != GetURLScheme(false) {
		t.Error("GetURLScheme failed")
	}
}

func TestInitTLSConfig(t *testing.T) {
	p := gomonkey.ApplyFunc(ioutil.ReadFile, func(filename string) ([]byte, error) {
		return nil, nil
	})
	p.ApplyFunc(containPassPhase, func(keyContent []byte, passPhase string, decryptTool string, isHttps bool) (Content []byte, err error) {
		return nil, nil
	})
	p.ApplyFunc(tls.X509KeyPair, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {
		var cert tls.Certificate
		return cert, nil
	})
	defer p.Reset()
	os.Setenv("SSL_ROOT", "/home/sn/resource/https")
	var config InternalHTTPSConfig
	config.TLSProtocol = "TLSv1.2"
	config.TLSCiphers = "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, TLS_TEST"
	err := InitTLSConfig(config)
	assert.Equal(t, nil, err)
}

func TestGetClientTLSConfig(t *testing.T) {
	actual := GetClientTLSConfig()
	assert.Equal(t, tlsConfig, actual)
}

func TestContainPassPhase(t *testing.T) {
	convey.Convey("ContainPassPhase", t, func() {
		errCtrl := ""
		patches := []*gomonkey.Patches{
			gomonkey.ApplyFunc(crypto.DecryptPEMBlock, func(b *pem.Block, password []byte) ([]byte, error) {
				if errCtrl == "returnError" {
					return nil, errors.New("some error")
				}
				return nil, nil
			}),
		}
		defer func() {
			for idx := range patches {
				patches[idx].Reset()
			}
		}()
		convey.Convey("http error case 1", func() {
			keyContent := []byte{}
			passPhase := ""
			isHttps := false
			content, err := containPassPhase(keyContent, passPhase, "LOCAL", isHttps)
			convey.So(err, convey.ShouldBeNil)
			convey.So(content, convey.ShouldBeNil)
		})
		convey.Convey("https error case 1", func() {
			keyContent := []byte{}
			passPhase := ""
			isHttps := true
			content, err := containPassPhase(keyContent, passPhase, "LOCAL", isHttps)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "failed to decode key file")
			convey.So(content, convey.ShouldBeNil)
		})
		convey.Convey("Decrypt error", func() {
			keyContent := pem.EncodeToMemory(&pem.Block{
				Type:    "MESSAGE",
				Headers: map[string]string{"DEK-Info": "test"},
				Bytes:   []byte("test containPassPhase")})
			passPhase := "abc"
			isHttps := true
			errCtrl = "returnError"
			content, err := containPassPhase(keyContent, passPhase, "LOCAL", isHttps)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(content, convey.ShouldBeNil)
		})
		convey.Convey("Decrypt success", func() {
			keyContent := pem.EncodeToMemory(&pem.Block{
				Type:    "MESSAGE",
				Headers: map[string]string{"DEK-Info": "test"},
				Bytes:   []byte("test containPassPhase")})
			passPhase := "abc"
			isHttps := true
			errCtrl = ""
			content, err := containPassPhase(keyContent, passPhase, "LOCAL", isHttps)
			convey.So(err, convey.ShouldBeNil)
			convey.So(content, convey.ShouldNotBeNil)
		})
	})
}

func TestLoadCerts(t *testing.T) {
	convey.Convey("https Load Certs 1", t, func() {
		patch := gomonkey.ApplyFunc(os.Getenv, func(key string) string {
			return "aaa"
		})
		defer patch.Reset()
		cert := loadCerts("./test", "trust.cer")
		convey.So(cert, convey.ShouldNotBeNil)
	})
	convey.Convey("https Load Certs 2", t, func() {
		patch := gomonkey.ApplyFunc(os.Getenv, func(key string) string {
			return "aaa"
		})
		patch2 := gomonkey.ApplyFunc(filepath.Abs, func(path string) (string, error) {
			return "a", errors.New("bbb")
		})
		defer patch.Reset()
		defer patch2.Reset()
		cert := loadCerts("1", "trust.cer")
		convey.So(cert, convey.ShouldNotBeNil)
	})

}

func Test_parseSSLProtocol(t *testing.T) {
	convey.Convey("Test_parseSSLProtocol", t, func() {
		convey.So(parseSSLProtocol("TLSv1.2"), convey.ShouldEqual, tls.VersionTLS12)
		convey.So(parseSSLProtocol("abc"), convey.ShouldEqual, 0)
	})
}

func Test_parseURL(t *testing.T) {
	url := ParseURL("http://test.com")
	assert.Equal(t, url, "test.com")
	url1 := ParseURL("test.com")
	assert.Equal(t, url1, "test.com")
}

func TestGetClientTLSConfig_Multi(t *testing.T) {
	old := tlsConfig

	tlsConfig = &tls.Config{}

	defer func() {
		tlsConfig = old
	}()

	a := GetClientTLSConfig()
	a.CipherSuites = append(a.CipherSuites, 10)

	b := GetClientTLSConfig()

	assert.NotEqual(t, a, b)
	assert.NotSame(t, a, b)
	assert.Equal(t, 1, len(a.CipherSuites))
	assert.Equal(t, 0, len(b.CipherSuites))
}

func TestLoadServerTLSCertificate(t *testing.T) {
	readFileCtrl := ""
	readFileCtrlCount := 0
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc(tls.X509KeyPair, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {
			return tls.Certificate{}, errors.New("some error")
		}),
		gomonkey.ApplyFunc(ioutil.ReadFile, func(filename string) ([]byte, error) {
			if readFileCtrl == "successOnce" {
				if readFileCtrlCount == 0 {
					readFileCtrlCount++
					return nil, nil
				}
				return nil, errors.New("some error")
			}
			readFileCtrlCount = 0
			return nil, nil
		}),
	}
	defer func() {
		for idx := range patches {
			patches[idx].Reset()
		}
	}()
	passLiteral := "testPassPhase"
	passByteArray := []byte(passLiteral)
	passPhase := string(passByteArray)
	readFileCtrl = "successOnce"
	certs, err := LoadServerTLSCertificate("testCertFile", "testKeyFile", passPhase, "LOCAL", true)
	assert.NotNil(t, err)
	assert.Empty(t, certs)
	certs, err = LoadServerTLSCertificate("testCertFile", "testKeyFile", passPhase, "LOCAL", true)
	assert.NotNil(t, err)
	assert.Empty(t, certs)
	readFileCtrl = ""
	certs, err = LoadServerTLSCertificate("testCertFile", "testKeyFile", passPhase, "LOCAL", true)
	assert.NotNil(t, err)
	assert.Empty(t, certs)
}

func Test_loadCertAndKeyBytes(t *testing.T) {
	patches := [...]*gomonkey.Patches{
		gomonkey.ApplyFunc(ioutil.ReadFile, func(filename string) ([]byte, error) {
			return []byte("abc"), nil
		}),
		gomonkey.ApplyFunc(pem.Decode, func(data []byte) (p *pem.Block, rest []byte) {
			return &pem.Block{}, []byte{}
		}),
		gomonkey.ApplyFunc(crypto.IsEncryptedPEMBlock, func(b *pem.Block) bool {
			return true
		}),
		gomonkey.ApplyFunc(crypto.DecryptPEMBlock, func(b *pem.Block, password []byte) ([]byte, error) {
			return []byte{}, nil
		}),
		gomonkey.ApplyFunc(pem.EncodeToMemory, func(b *pem.Block) []byte {
			return []byte("abc")
		}),
	}
	defer func() {
		for idx := range patches {
			patches[idx].Reset()
		}
	}()
	convey.Convey("loadCertAndKeyBytes", t, func() {
		bytes, keyPEMBlock, err := loadCertAndKeyBytes("path1", "path2", "", "", true)
		convey.So(err, convey.ShouldBeNil)
		convey.So(string(bytes), convey.ShouldEqual, "abc")
		convey.So(string(keyPEMBlock), convey.ShouldEqual, "abc")
	})
}
