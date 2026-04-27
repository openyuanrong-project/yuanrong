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
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	server    http.Server
	rootKEY   string
	rootPEM   string
	rootSRL   string
	serverKEY string
	serverPEM string
	serverCSR string
}

func (s *TestSuite) SetupSuite() {
	certificatePath, err := os.Getwd()
	if err != nil {
		s.T().Errorf("failed to get current working dictionary: %s", err.Error())
		return
	}

	certificatePath += "/../../../test/"
	s.rootKEY = certificatePath + "ca.key"
	s.rootPEM = certificatePath + "ca.crt"
	s.rootSRL = certificatePath + "ca.srl"
	s.serverKEY = certificatePath + "server.key"
	s.serverPEM = certificatePath + "server.crt"
	s.serverCSR = certificatePath + "server.csr"

	body := "Hello"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	})

	s.server = http.Server{
		Addr:    "127.0.0.1:6061",
		Handler: handler,
	}

	go func() {
		if err := s.server.ListenAndServeTLS(s.serverPEM, s.serverKEY); err != nil {
			s.T().Logf("failed to start server: %s", err.Error())
		}
	}()
}

func (s *TestSuite) TearDownSuite() {
	s.server.Shutdown(context.Background())

	os.Remove(s.serverKEY)
	os.Remove(s.serverPEM)
	os.Remove(s.serverCSR)
	os.Remove(s.rootKEY)
	os.Remove(s.rootPEM)
	os.Remove(s.rootSRL)
}

func TestOptionTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func TestWithFunc(t *testing.T) {
	WithCertsContent(nil, nil)
	WithCerts("", "")
	WithClientCAs()
	patch := gomonkey.ApplyFunc(LoadRootCAs, func(caFiles ...string) (*x509.CertPool, error) {
		return nil, errors.New("LoadRootCAs failed")
	})
	WithClientCAs("")
	patch.Reset()
}

func TestVerifyCert(t *testing.T) {
	var raw [][]byte
	tlsConfig = &tls.Config{}
	tlsConfig.ClientCAs = x509.NewCertPool()
	err := VerifyCert(raw, nil)
	assert.NotNil(t, err)

	raw = [][]byte{
		[]byte("0"),
		[]byte("1"),
	}
	err = VerifyCert(raw, nil)
	assert.NotNil(t, err)

	patch1 := gomonkey.ApplyFunc(x509.ParseCertificate, func([]byte) (*x509.Certificate, error) {
		return &x509.Certificate{}, nil
	})
	VerifyCert(raw, nil)
	patch1.Reset()
}

func TestApply(t *testing.T) {
	cli := clientCAOption{}
	cli.apply(&tls.Config{})
}
