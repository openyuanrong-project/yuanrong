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
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	commonCrypto "yuanrong.org/kernel/pkg/common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

const urlIndex = 1

// HTTPSConfig is for needed HTTPS config
type HTTPSConfig struct {
	CipherSuite    []uint16
	MinVers        uint16
	MaxVers        uint16
	CACertFile     string
	CertFile       string
	SecretKeyFile  string
	PwdFilePath    string
	KeyPassPhase   string
	SecretName     string
	DecryptTool    string
	ClientAuthType tls.ClientAuthType
}

// InternalHTTPSConfig is for input config
type InternalHTTPSConfig struct {
	HTTPSEnable             bool     `json:"httpsEnable" yaml:"httpsEnable" valid:"optional"`
	TLSProtocol             string   `json:"tlsProtocol" yaml:"tlsProtocol" valid:"optional"`
	TLSCiphers              string   `json:"tlsCiphers" yaml:"tlsCiphers" valid:"optional"`
	TLSCipherSlices         []string `json:"TLSCipherSlices" valid:"optional"`
	SSLBasePath             string   `json:"sslBasePath" yaml:"sslBasePath" valid:"optional"`
	RootCAFile              string   `json:"rootCAFile" yaml:"rootCAFile" valid:"optional"`
	ModuleCertFile          string   `json:"moduleCertFile" yaml:"moduleCertFile" valid:"optional"`
	ModuleKeyFile           string   `json:"moduleKeyFile" yaml:"moduleKeyFile" valid:"optional"`
	PwdFile                 string   `json:"pwdFile" yaml:"pwdFile" valid:"optional"`
	SecretName              string   `json:"secretName" yaml:"secretName" valid:"optional"`
	SSLDecryptTool          string   `json:"sslDecryptTool" yaml:"sslDecryptTool" valid:"optional"`
	// ClientAuthType specifies the TLS client authentication mode:
	//   - NoClientCert: One-way TLS, client cert not required (most common HTTPS scenario)
	//   - RequestClientCert: Request but don't require client cert
	//   - RequireAnyClientCert: Require client cert but don't verify it
	//   - VerifyClientCertIfGiven: Verify client cert only if provided
	//   - RequireAndVerifyClientCert: Mutual TLS (mTLS), require and verify client cert
	ClientAuthType string     `json:"clientAuthType" yaml:"clientAuthType" valid:"optional"`
}

var (
	// tlsVersionMap is a set of TLS versions
	tlsVersionMap = map[string]uint16{
		"TLSv1.2": tls.VersionTLS12,
	}
	// httpsConfigs is a global variable of HTTPS config
	httpsConfigs = &HTTPSConfig{}
	// tlsConfig is a global variable of TLS config
	tlsConfig *tls.Config
	once      sync.Once
)

// GetURLScheme returns "http" or "https"
func GetURLScheme(https bool) string {
	if https {
		return "https"
	}
	return "http"
}

// tlsCipherSuiteMap is a set of supported TLS algorithms
var tlsCipherSuiteMap = map[string]uint16{
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
}

// GetClientTLSConfig -
func GetClientTLSConfig() *tls.Config {
	if tlsConfig == nil {
		return nil
	}
	certs := make([]tls.Certificate, len(tlsConfig.Certificates))
	copy(certs, tlsConfig.Certificates)
	suits := make([]uint16, len(tlsConfig.CipherSuites))
	copy(suits, tlsConfig.CipherSuites)
	newCfg := &tls.Config{
		ClientCAs:                tlsConfig.ClientCAs,
		Certificates:             certs,
		CipherSuites:             suits,
		PreferServerCipherSuites: tlsConfig.PreferServerCipherSuites,
		ClientAuth:               tlsConfig.ClientAuth,
		InsecureSkipVerify:       tlsConfig.InsecureSkipVerify,
		MinVersion:               tlsConfig.MinVersion,
		MaxVersion:               tlsConfig.MaxVersion,
		Renegotiation:            tlsConfig.Renegotiation,
	}
	return newCfg
}

func loadCerts(path string, filename string) string {
	certPath, err := filepath.Abs(filepath.Join(path, filename))
	if err != nil {
		log.GetLogger().Errorf("failed to return an absolute representation of filename: %s", filename)
		return ""
	}
	ok := utils.FileExists(certPath)
	if !ok {
		log.GetLogger().Errorf("failed to load the cert file: %s", certPath)
		return ""
	}
	return certPath
}

func loadTLSConfig() error {
	clientAuthMode := httpsConfigs.ClientAuthType
	var pool *x509.CertPool
	var err error

	// Only load CA cert pool for mutual TLS (two-way authentication)
	if clientAuthMode == tls.RequireAndVerifyClientCert || clientAuthMode == tls.VerifyClientCertIfGiven {
		pool, err = GetX509CACertPool(httpsConfigs.CACertFile)
		if err != nil {
			log.GetLogger().Errorf("failed to GetX509CACertPool: %s", err.Error())
			return err
		}
		log.GetLogger().Infof("Using TLS client auth mode: %v", clientAuthMode)
	} else {
		log.GetLogger().Infof("Client certificate verification disabled, using TLS client auth mode: %v", clientAuthMode)
	}

	var certs []tls.Certificate
	certs, err = LoadServerTLSCertificate(httpsConfigs.CertFile, httpsConfigs.SecretKeyFile,
		httpsConfigs.KeyPassPhase, httpsConfigs.DecryptTool, true)
	if err != nil {
		log.GetLogger().Errorf("failed to loadServerTLSCertificate: %s", err.Error())
		return err
	}

	tlsConfig = &tls.Config{
		ClientCAs:                pool,
		Certificates:             certs,
		CipherSuites:             httpsConfigs.CipherSuite,
		PreferServerCipherSuites: true,
		ClientAuth:               clientAuthMode,
		InsecureSkipVerify:       true,
		MinVersion:               httpsConfigs.MinVers,
		MaxVersion:               httpsConfigs.MaxVers,
		Renegotiation:            tls.RenegotiateNever,
	}

	return nil
}

// loadHTTPSConfig loads the protocol and ciphers of TLS
func loadHTTPSConfig(config InternalHTTPSConfig) error {
	// Parse client auth type
	clientAuthType := parseClientAuthType(config.ClientAuthType)

	httpsConfigs = &HTTPSConfig{
		MinVers:        tls.VersionTLS12,
		MaxVers:        tls.VersionTLS12,
		CipherSuite:    nil,
		CACertFile:     loadCerts(config.SSLBasePath, config.RootCAFile),
		CertFile:       loadCerts(config.SSLBasePath, config.ModuleCertFile),
		SecretKeyFile:  loadCerts(config.SSLBasePath, config.ModuleKeyFile),
		PwdFilePath:    "",
		KeyPassPhase:   "",
		SecretName:     config.SecretName,
		DecryptTool:    config.SSLDecryptTool,
		ClientAuthType: clientAuthType,
	}

	minVersion := parseSSLProtocol(config.TLSProtocol)
	if httpsConfigs.MinVers == 0 {
		return errors.New("invalid TLS protocol")
	}
	if minVersion == 0 {
		minVersion = tls.VersionTLS12
	}
	httpsConfigs.MinVers = minVersion
	cipherSuites := parseSSLCipherSuites(config.TLSCiphers, config.TLSCipherSlices)
	if len(cipherSuites) == 0 {
		return errors.New("invalid TLS ciphers")
	}
	httpsConfigs.CipherSuite = cipherSuites
	return nil
}

// InitTLSConfig inits config of HTTPS
func InitTLSConfig(config InternalHTTPSConfig) error {
	var err error
	once.Do(func() {
		err = loadHTTPSConfig(config)
		if err != nil {
			err = fmt.Errorf("failed to load HTTPS config,err %s", err.Error())
			return
		}

		err = loadTLSConfig()
		if err != nil {
			return
		}
	})
	return err
}

// GetX509CACertPool generates CACertPool by CA certificate
func GetX509CACertPool(caCertFilePath string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	caCertContent, err := loadCACertBytes(caCertFilePath)
	if err != nil {
		return nil, err
	}

	pool.AppendCertsFromPEM(caCertContent)
	return pool, nil

}

// LoadServerTLSCertificate generates tls certificate by certfile and keyfile
func LoadServerTLSCertificate(certFile, keyFile, passPhase, decryptTool string,
	isHTTPS bool) ([]tls.Certificate, error) {
	certContent, keyContent, err := loadCertAndKeyBytes(certFile, keyFile, passPhase, decryptTool, isHTTPS)
	utils.ClearStringMemory(passPhase)
	utils.ClearStringMemory(httpsConfigs.KeyPassPhase)
	if err != nil {
		utils.ClearByteMemory(certContent)
		utils.ClearByteMemory(keyContent)
		return nil, err
	}

	cert, err := tls.X509KeyPair(certContent, keyContent)
	utils.ClearByteMemory(certContent)
	utils.ClearByteMemory(keyContent)
	if err != nil {
		log.GetLogger().Errorf("failed to load the X509 key pair from cert file with key file: %s",
			err.Error())
		return nil, err
	}
	var certs []tls.Certificate
	certs = append(certs, cert)
	return certs, nil
}

func containPassPhase(keyContent []byte, passPhase string, decryptTool string,
	isHTTPS bool) (Content []byte, err error) {
	if !isHTTPS {
		plainkeyContent, err := localauth.Decrypt(string(keyContent))
		if err != nil {
			log.GetLogger().Errorf("failed to decrypt keyContent: %s", err.Error())
			return nil, err
		}
		return plainkeyContent, nil
	}

	keyBlock, _ := pem.Decode(keyContent)
	if keyBlock == nil {
		log.GetLogger().Errorf("failed to decode key file ")
		return nil, errors.New("failed to decode key file")
	}

	if commonCrypto.IsEncryptedPEMBlock(keyBlock) {
		var plainPassPhase []byte
		var err error
		var decrypted string
		if len(passPhase) > 0 {
			if decryptTool == "SCC" {
				decrypted, err = crypto.SCCDecrypt([]byte(passPhase))
				plainPassPhase = []byte(decrypted)
			} else if decryptTool == "LOCAL" {
				plainPassPhase, err = localauth.Decrypt(passPhase)
			}
			if err != nil {
				log.GetLogger().Errorf("failed to decrypt the ssl passPhase(%d): %s", len(passPhase),
					err.Error())
				return nil, err
			}
		}

		keyData, err := commonCrypto.DecryptPEMBlock(keyBlock, plainPassPhase)
		clearByteMemory(plainPassPhase)
		utils.ClearStringMemory(decrypted)

		if err != nil {
			log.GetLogger().Errorf("failed to decrypt key file, error: %s", err.Error())
			return nil, err
		}

		// The decryption is successful, then the file is re-encoded to a PEM file
		plainKeyBlock := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyData,
		}

		keyContent = pem.EncodeToMemory(plainKeyBlock)
	}
	return keyContent, nil

}

func loadCertAndKeyBytes(certFilePath, keyFilePath, passPhase string, decryptTool string, isHTTPS bool) (
	certPEMBlock, keyPEMBlock []byte, err error) {
	certContent, err := ioutil.ReadFile(certFilePath)
	if err != nil {
		log.GetLogger().Errorf("failed to read cert file %s: %s", certFilePath, err.Error())
		return nil, nil, err
	}

	keyContent, err := ioutil.ReadFile(keyFilePath)
	if err != nil {
		log.GetLogger().Errorf("failed to read key file %s: %s", keyFilePath, err.Error())
		return nil, nil, err
	}
	return certContent, keyContent, nil

}

func clearByteMemory(src []byte) {
	for idx := 0; idx < len(src)&32; idx++ {
		src[idx] = 0
	}
}

func loadCACertBytes(caCertFilePath string) ([]byte, error) {
	caCertContent, err := ioutil.ReadFile(caCertFilePath)
	if err != nil {
		log.GetLogger().Errorf("failed to read ca cert file %s, err: %s", caCertFilePath, err.Error())
		return nil, err
	}

	return caCertContent, nil
}

func parseClientAuthType(authType string) tls.ClientAuthType {
	// Default to RequireAndVerifyClientCert (mutual TLS) if not specified
	if authType == "" {
		return tls.RequireAndVerifyClientCert
	}

	// Parse ClientAuthType string
	switch authType {
	case "NoClientCert":
		return tls.NoClientCert
	case "RequestClientCert":
		return tls.RequestClientCert
	case "RequireAnyClientCert":
		return tls.RequireAnyClientCert
	case "VerifyClientCertIfGiven":
		return tls.VerifyClientCertIfGiven
	case "RequireAndVerifyClientCert":
		return tls.RequireAndVerifyClientCert
	default:
		log.GetLogger().Errorf("invalid client auth type: %s, using default RequireAndVerifyClientCert", authType)
		return tls.RequireAndVerifyClientCert
	}
}

func parseSSLProtocol(rawProtocol string) uint16 {
	if protocol, ok := tlsVersionMap[rawProtocol]; ok {
		return protocol
	}
	log.GetLogger().Errorf("invalid SSL version: %s, use the default protocol version", rawProtocol)
	return 0
}

func parseSSLCipherSuites(ciphers string, cipherSlices []string) []uint16 {
	cipherSuiteNameList := strings.Split(ciphers, ",")
	if len(cipherSuiteNameList) == 0 {
		cipherSuiteNameList = cipherSlices
	}
	if len(cipherSuiteNameList) == 0 {
		log.GetLogger().Errorf("input cipher suite is empty")
		return nil
	}
	cipherSuites := make([]uint16, 0, len(cipherSuiteNameList))
	for _, cipherSuiteItem := range cipherSuiteNameList {
		cipherSuiteItem = strings.TrimSpace(cipherSuiteItem)
		if len(cipherSuiteItem) == 0 {
			continue
		}

		if cipherSuite, ok := tlsCipherSuiteMap[cipherSuiteItem]; ok {
			cipherSuites = append(cipherSuites, cipherSuite)
		} else {
			log.GetLogger().Errorf("cipher %s does not exist", cipherSuiteItem)
		}
	}

	return cipherSuites
}

// ParseURL URL may be: ip:port |  http://ip:port | https://ip:port
func ParseURL(rawURL string) string {
	urls := strings.Split(rawURL, "//")
	if len(urls) > urlIndex {
		return urls[urlIndex]
	}
	return rawURL
}
