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
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
)

func TestGetURLScheme(t *testing.T) {
	if "https" != GetURLScheme(true) {
		t.Error("GetURLScheme failed")
	}
	if "http" != GetURLScheme(false) {
		t.Error("GetURLScheme failed")
	}
}

func TestInitTLSConfig1(t *testing.T) {
	os.Setenv("SSL_ROOT", "/home/sn/resource/https")
	tlsCiphers := []byte("TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, TLS_TEST")
	err := InitTLSConfig("TLSv1.2", tlsCiphers)
	assert.NotEqual(t, nil, err)

	patch := gomonkey.ApplyFunc(loadHTTPSConfig, func(string, []byte) error {
		return nil
	})
	InitTLSConfig("TLSv1.2", tlsCiphers)
	patch.Reset()
}

func TestInitTLSConfig2(t *testing.T) {
	os.Setenv("SSL_ROOT", "/home/sn/resource/https")
	tlsCiphers := []byte("TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, TLS_TEST")
	patch := gomonkey.ApplyFunc(loadHTTPSConfig, func(string, []byte) error {
		return nil
	})
	InitTLSConfig("TLSv1.2", tlsCiphers)
	patch.Reset()
}

func Test_loadServerTLSCertificate(t *testing.T) {
	HTTPSConfigs.CertFile = "/home/snuser"
	_, err := loadServerTLSCertificate()
	assert.NotNil(t, err)

	convey.Convey("test loadServerTLSCertificate", t, func() {
		convey.Convey("LoadCertAndKeyBytes success", func() {
			patch1 := gomonkey.ApplyFunc(LoadCertAndKeyBytes, func(certFilePath, keyFilePath, passPhase string) (certPEMBlock,
				keyPEMBlock []byte, err error) {
				return nil, nil, nil
			})
			convey.Convey("X509KeyPair success", func() {
				patch2 := gomonkey.ApplyFunc(tls.X509KeyPair, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {
					return tls.Certificate{}, nil
				})
				_, err := loadServerTLSCertificate()
				convey.So(err, convey.ShouldBeNil)
				_, err = LoadServerTLSCertificate("", "")
				convey.So(err, convey.ShouldBeNil)
				defer patch2.Reset()
			})
			convey.Convey("X509KeyPair fail", func() {
				patch3 := gomonkey.ApplyFunc(tls.X509KeyPair, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {
					return tls.Certificate{}, errors.New("fail to load X509KeyPair")
				})
				_, err := loadServerTLSCertificate()
				convey.So(err, convey.ShouldNotBeNil)
				_, err = LoadServerTLSCertificate("", "")
				convey.So(err, convey.ShouldNotBeNil)
				defer patch3.Reset()
			})
			defer patch1.Reset()
		})
		convey.Convey("LoadCertAndKeyBytes fail", func() {
			patch4 := gomonkey.ApplyFunc(LoadCertAndKeyBytes, func(certFilePath, keyFilePath, passPhase string) (certPEMBlock,
				keyPEMBlock []byte, err error) {
				return nil, nil, errors.New("fail to LoadCertAndKeyBytes")
			})
			_, err := loadServerTLSCertificate()
			convey.So(err, convey.ShouldNotBeNil)
			_, err = LoadServerTLSCertificate("", "")
			convey.So(err, convey.ShouldNotBeNil)
			defer patch4.Reset()
		})
	})
}

func TestHTTPListenAndServeTLS(t *testing.T) {
	server := &http.Server{}
	err := HTTPListenAndServeTLS("127.0.0.1", server)
	assert.NotNil(t, err)
}

func Test_loadTLSConfig(t *testing.T) {
	HTTPSConfigs = &HTTPSConfig{}
	err := loadTLSConfig()
	assert.NotNil(t, err)

	convey.Convey("test loadTLSConfig", t, func() {
		patches := [...]*gomonkey.Patches{
			gomonkey.ApplyFunc(GetX509CACertPool, func(caCertFilePath string) (caCertPool *x509.CertPool, err error) {
				return x509.NewCertPool(), nil
			}),
		}
		defer func() {
			for idx := range patches {
				patches[idx].Reset()
			}
		}()
		err = loadTLSConfig()
		assert.NotNil(t, err)
	})
}

func Test_LoadCertAndKeyBytes(t *testing.T) {
	convey.Convey("test LoadCertAndKeyBytes", t, func() {
		convey.Convey("ReadFile fail", func() {
			_, _, err := LoadCertAndKeyBytes("certPath", "keyPath", "pass")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("ReadFile success", func() {
			patches := [...]*gomonkey.Patches{
				gomonkey.ApplyFunc(ioutil.ReadFile, func(filename string) ([]byte, error) {
					return nil, nil
				}),
				gomonkey.ApplyFunc(crypto.GetRootKey, func() []byte {
					return nil
				}),
			}
			defer func() {
				for idx := range patches {
					patches[idx].Reset()
				}
			}()

			convey.Convey("DecryptByte fail", func() {
				patch := gomonkey.ApplyFunc(crypto.DecryptByte, func(cipherText []byte, secret []byte) ([]byte, error) {
					return []byte{}, errors.New("DecryptByte fail")
				})
				defer patch.Reset()

				_, _, err := LoadCertAndKeyBytes("certPath", "keyPath", "pass")
				convey.So(err, convey.ShouldNotBeNil)
			})
			convey.Convey("DecryptByte success", func() {
				patch := gomonkey.ApplyFunc(crypto.DecryptByte, func(cipherText []byte, secret []byte) ([]byte, error) {
					return []byte{}, nil
				})
				defer patch.Reset()

				convey.Convey("Decode fail", func() {
					patch := gomonkey.ApplyFunc(pem.Decode, func(data []byte) (p *pem.Block, rest []byte) {
						return nil, nil
					})
					defer patch.Reset()

					_, _, err := LoadCertAndKeyBytes("certPath", "keyPath", "pass")
					convey.So(err, convey.ShouldNotBeNil)
				})
				convey.Convey("Decode success", func() {
					patch := gomonkey.ApplyFunc(pem.Decode, func(data []byte) (p *pem.Block, rest []byte) {
						return &pem.Block{}, nil
					})
					defer patch.Reset()

					convey.Convey("crypto.IsEncryptedPEMBlock fail", func() {
						patch := gomonkey.ApplyFunc(crypto.IsEncryptedPEMBlock, func(b *pem.Block) bool {
							return false
						})
						defer patch.Reset()

						_, _, err := LoadCertAndKeyBytes("certPath", "keyPath", "pass")
						convey.So(err, convey.ShouldBeNil)
					})
					convey.Convey("crypto.IsEncryptedPEMBlock success", func() {
						patch := gomonkey.ApplyFunc(crypto.IsEncryptedPEMBlock, func(b *pem.Block) bool {
							return true
						})
						defer patch.Reset()

						convey.Convey("localauth.Decrypt fail", func() {
							patch := gomonkey.ApplyFunc(localauth.Decrypt, func(src string) ([]byte, error) {
								return nil, errors.New("localauth.Decrypt fail")
							})
							defer patch.Reset()

							_, _, err := LoadCertAndKeyBytes("certPath", "keyPath", "pass")
							convey.So(err, convey.ShouldNotBeNil)
						})
						convey.Convey("localauth.Decrypt success", func() {
							patch := gomonkey.ApplyFunc(localauth.Decrypt, func(src string) ([]byte, error) {
								return []byte{}, nil
							})
							defer patch.Reset()

							convey.Convey("crypto.DecryptPEMBlock fail", func() {
								patch := gomonkey.ApplyFunc(crypto.DecryptPEMBlock,
									func(b *pem.Block, password []byte) ([]byte, error) {
										return nil, errors.New("crypto.DecryptPEMBlock fail")
									})
								defer patch.Reset()

								_, _, err := LoadCertAndKeyBytes("certPath", "keyPath", "pass")
								convey.So(err, convey.ShouldNotBeNil)
							})
						})
					})
				})
			})
		})
	})
}

func Test_loadCerts(t *testing.T) {
	convey.Convey("env length 0", t, func() {
		patch := gomonkey.ApplyFunc(os.Getenv, func(key string) string {
			return ""
		})
		defer patch.Reset()
		result := loadCerts("")
		convey.So(result, convey.ShouldEqual, "")
	})
}

func Test_GetClientTLSConfig(t *testing.T) {
	convey.Convey("test GetClientTLSConfig", t, func() {
		convey.So(GetClientTLSConfig(), convey.ShouldBeNil)
	})
}

func Test_GetHTTPTransport(t *testing.T) {
	convey.Convey("test GetHTTPTransport", t, func() {
		convey.So(GetHTTPTransport(), convey.ShouldNotBeNil)
	})
}

func Test_BuildClientTLSConfOpts(t *testing.T) {
	BuildClientTLSConfOpts(MutualTLSConfig{})
}

func Test_BuildServerTLSConfOpts(t *testing.T) {
	BuildServerTLSConfOpts(MutualTLSConfig{})
}

func Test_ClearByteMemory(t *testing.T) {
	convey.Convey("test clearByteMemory", t, func() {
		s := make([]byte, 33)
		s = append(s, 'A')
		clearByteMemory(s)
		convey.So(s[0], convey.ShouldEqual, 0)
	})
}

func Test_parseSSLProtocol(t *testing.T) {
	parseSSLProtocol("")
}

func Test_HTTPListenAndServeTLS(t *testing.T) {
	patch := gomonkey.ApplyFunc(net.Listen, func(string, string) (net.Listener, error) {
		return nil, nil
	})
	defer patch.Reset()

	patch1 := gomonkey.ApplyFunc((*http.Server).Serve, func(*http.Server, net.Listener) error {
		return nil
	})
	HTTPListenAndServeTLS("", &http.Server{})
	patch1.Reset()

	patch2 := gomonkey.ApplyFunc((*http.Server).Serve, func(*http.Server, net.Listener) error {
		return errors.New("test")
	})
	HTTPListenAndServeTLS("", &http.Server{})
	patch2.Reset()
}

func Test_GetX509CACertPool(t *testing.T) {
	patch := gomonkey.ApplyFunc(LoadCACertBytes, func(string) ([]byte, error) {
		return []byte{'a'}, nil
	})
	GetX509CACertPool("")
	patch.Reset()
}
