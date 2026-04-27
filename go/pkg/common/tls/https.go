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
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"yuanrong.org/kernel/pkg/common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/reader"
)

// HTTPSConfig is for needed HTTPS config
type HTTPSConfig struct {
	CipherSuite   []uint16
	MinVers       uint16
	MaxVers       uint16
	CACertFile    string
	CertFile      string
	SecretKeyFile string
	PwdFilePath   string
	KeyPassPhase  string
}

// InternalHTTPSConfig is for input config
type InternalHTTPSConfig struct {
	HTTPSEnable bool   `json:"httpsEnable" yaml:"httpsEnable" valid:"optional"`
	TLSProtocol string `json:"tlsProtocol" yaml:"tlsProtocol" valid:"optional"`
	TLSCiphers  string `json:"tlsCiphers" yaml:"tlsCiphers" valid:"optional"`
}

var (
	// HTTPSConfigs is a global variable of HTTPS config
	HTTPSConfigs = &HTTPSConfig{}
	// tlsConfig is a global variable of TLS config
	tlsConfig *tls.Config
	once      sync.Once

	// tlsVersionMap is a set of TLS versions
	tlsVersionMap = map[string]uint16{
		"TLSv1.2": tls.VersionTLS12,
	}
)

// tlsCipherSuiteMap is a set of supported TLS algorithms
var tlsCipherSuiteMap = map[string]uint16{
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
}

// GetURLScheme returns "http" or "https"
func GetURLScheme(https bool) string {
	if https {
		return "https"
	}
	return "http"
}

// HTTPListenAndServeTLS listens and serves by TLS in HTTP
func HTTPListenAndServeTLS(addr string, server *http.Server) error {
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}

	tlsListener := tls.NewListener(listener, tlsConfig)

	if err = server.Serve(tlsListener); err != nil {
		return err
	}
	return nil
}

// GetClientTLSConfig returns the config of TLS
func GetClientTLSConfig() *tls.Config {
	return tlsConfig
}

// GetHTTPTransport get http transport
func GetHTTPTransport() *http.Transport {
	tr, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil
	}
	tr.TLSClientConfig = GetClientTLSConfig()

	return tr
}

func loadCerts(path string) string {
	env := os.Getenv("SSL_ROOT")
	if len(env) == 0 {
		log.GetLogger().Errorf("failed to get SSL_ROOT")
		return ""
	}
	certPath, err := filepath.Abs(filepath.Join(env, path))
	if err != nil {
		log.GetLogger().Errorf("failed to return an absolute representation of path: %s", path)
		return ""
	}
	ok := utils.FileExists(certPath)
	if !ok {
		log.GetLogger().Errorf("failed to load the cert file: %s", certPath)
		return ""
	}
	return certPath
}

func loadTLSConfig() (err error) {
	clientAuthMode := tls.NoClientCert
	var pool *x509.CertPool

	pool, err = GetX509CACertPool(HTTPSConfigs.CACertFile)
	if err != nil {
		log.GetLogger().Errorf("failed to GetX509CACertPool: %s", err.Error())
		return err
	}

	var certs []tls.Certificate
	certs, err = loadServerTLSCertificate()
	if err != nil {
		log.GetLogger().Errorf("failed to loadServerTLSCertificate: %s", err.Error())
		return err
	}

	tlsConfig = &tls.Config{
		ClientCAs:                pool,
		Certificates:             certs,
		CipherSuites:             HTTPSConfigs.CipherSuite,
		PreferServerCipherSuites: true,
		ClientAuth:               clientAuthMode,
		InsecureSkipVerify:       true,
		MinVersion:               HTTPSConfigs.MinVers,
		MaxVersion:               HTTPSConfigs.MaxVers,
		Renegotiation:            tls.RenegotiateNever,
	}

	return nil
}

// loadHTTPSConfig loads the protocol and ciphers of TLS
func loadHTTPSConfig(tlsProtocols string, tlsCiphers []byte) error {
	HTTPSConfigs = &HTTPSConfig{
		MinVers:       tls.VersionTLS12,
		MaxVers:       tls.VersionTLS12,
		CipherSuite:   nil,
		CACertFile:    loadCerts("trust.cer"),
		CertFile:      loadCerts("server.cer"),
		SecretKeyFile: loadCerts("server_key.pem"),
		PwdFilePath:   loadCerts("cert_pwd"),
		KeyPassPhase:  "",
	}

	minVersion := parseSSLProtocol(tlsProtocols)
	if HTTPSConfigs.MinVers == 0 {
		return errors.New("invalid TLS protocol")
	}
	HTTPSConfigs.MinVers = minVersion
	cipherSuites := parseSSLCipherSuites(tlsCiphers)
	if len(cipherSuites) == 0 {
		return errors.New("invalid TLS ciphers")
	}
	HTTPSConfigs.CipherSuite = cipherSuites

	keyPassPhase, err := reader.ReadFileWithTimeout(HTTPSConfigs.PwdFilePath)
	if err != nil {
		log.GetLogger().Errorf("failed to read file cert_pwd: %s", err.Error())
		return err
	}
	HTTPSConfigs.KeyPassPhase = string(keyPassPhase)

	return nil
}

// InitTLSConfig inits config of HTTPS
func InitTLSConfig(tlsProtocols string, tlsCiphers []byte) (err error) {
	once.Do(func() {
		err = loadHTTPSConfig(tlsProtocols, tlsCiphers)
		if err != nil {
			err = errors.New("failed to load HTTPS config")
			return
		}
		err = loadTLSConfig()
		if err != nil {
			return
		}
	})
	return err
}

// GetX509CACertPool get ca cert pool
func GetX509CACertPool(caCertFilePath string) (caCertPool *x509.CertPool, err error) {
	pool := x509.NewCertPool()
	caCertContent, err := LoadCACertBytes(caCertFilePath)
	if err != nil {
		return nil, err
	}

	pool.AppendCertsFromPEM(caCertContent)
	return pool, nil
}

func loadServerTLSCertificate() (tlsCert []tls.Certificate, err error) {
	certContent, keyContent, err := LoadCertAndKeyBytes(HTTPSConfigs.CertFile, HTTPSConfigs.SecretKeyFile,
		HTTPSConfigs.KeyPassPhase)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		log.GetLogger().Errorf("failed to load the X509 key pair from cert file %s with key file %s: %s",
			HTTPSConfigs.CertFile, HTTPSConfigs.SecretKeyFile, err.Error())
		return nil, err
	}

	var certs []tls.Certificate
	certs = append(certs, cert)

	return certs, nil
}

// LoadServerTLSCertificate generates tls certificate by certfile and keyfile
func LoadServerTLSCertificate(cerfile, keyfile string) (tlsCert []tls.Certificate, err error) {
	certContent, keyContent, err := LoadCertAndKeyBytes(cerfile, keyfile, "")

	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		log.GetLogger().Errorf("failed to load the X509 key pair from cert file %s with key file %s: %s",
			cerfile, keyfile, err.Error())
		return nil, err
	}
	var certs []tls.Certificate
	certs = append(certs, cert)
	return certs, nil
}

// LoadCertAndKeyBytes load cert and key bytes
func LoadCertAndKeyBytes(certFilePath, keyFilePath, passPhase string) (certPEMBlock, keyPEMBlock []byte, err error) {
	certContent, err := reader.ReadFileWithTimeout(certFilePath)
	if err != nil {
		log.GetLogger().Errorf("failed to read cert file %s", err.Error())
		return nil, nil, err
	}

	keyContent, err := reader.ReadFileWithTimeout(keyFilePath)
	if err != nil {
		log.GetLogger().Errorf("failed to read key file %s", err.Error())
		return nil, nil, err
	}

	keyContent, err = crypto.DecryptByte(keyContent, crypto.GetRootKey())
	if err != nil {
		log.GetLogger().Errorf("failed to decrypt key content, err: %s", err.Error())
		return nil, nil, err
	}
	keyBlock, _ := pem.Decode(keyContent)
	if keyBlock == nil {
		log.GetLogger().Errorf("failed to decode key file")
		return nil, nil, errors.New("failed to decode key file")
	}

	if crypto.IsEncryptedPEMBlock(keyBlock) {
		var plainPassPhase []byte
		if len(passPhase) > 0 {
			plainPassPhase, err = localauth.Decrypt(passPhase)
			if err != nil {
				log.GetLogger().Errorf("failed to decrypt the ssl passPhase(%d): %s", len(passPhase),
					err.Error())
				return nil, nil, err
			}
		}

		keyData, err := crypto.DecryptPEMBlock(keyBlock, plainPassPhase)
		clearByteMemory(plainPassPhase)
		if err != nil {
			log.GetLogger().Errorf("failed to decrypt key file, error: %s", err.Error())
			return nil, nil, err
		}

		// The decryption is successful, then the file is re-encoded to a PEM file
		plainKeyBlock := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyData,
		}

		keyContent = pem.EncodeToMemory(plainKeyBlock)
	}

	return certContent, keyContent, nil
}

func clearByteMemory(src []byte) {
	for idx := 0; idx < len(src)&32; idx++ {
		src[idx] = 0
	}
}

// LoadCACertBytes Load CA Cert Content
func LoadCACertBytes(caCertFilePath string) ([]byte, error) {
	caCertContent, err := reader.ReadFileWithTimeout(caCertFilePath)
	if err != nil {
		log.GetLogger().Errorf("failed to read ca cert file %s: %s", caCertFilePath, err.Error())
		return nil, err
	}

	return caCertContent, nil
}

func parseSSLProtocol(rawProtocol string) uint16 {
	if protocol, ok := tlsVersionMap[rawProtocol]; ok {
		return protocol
	}
	log.GetLogger().Errorf("invalid SSL version %s, use the default protocol version", rawProtocol)
	return 0
}

func parseSSLCipherSuites(ciphers []byte) []uint16 {
	cipherSuiteNameList := strings.Split(string(ciphers), ",")
	if len(cipherSuiteNameList) == 0 {
		log.GetLogger().Errorf("no input cipher suite")
		return nil
	}
	cipherSuiteList := make([]uint16, 0, len(cipherSuiteNameList))
	for _, cipherSuiteItem := range cipherSuiteNameList {
		cipherSuiteItem = strings.TrimSpace(cipherSuiteItem)
		if len(cipherSuiteItem) == 0 {
			continue
		}

		if cipherSuite, ok := tlsCipherSuiteMap[cipherSuiteItem]; ok {
			cipherSuiteList = append(cipherSuiteList, cipherSuite)
		} else {
			log.GetLogger().Errorf("cipher %s does not exist", cipherSuiteItem)
		}
	}

	return cipherSuiteList
}
