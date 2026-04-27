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

package wisecloudtool

import (
	"crypto/tls"
	"errors"
	"net"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/valyala/fasthttp"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/serviceaccount"
	"yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
)

func TestNewColdStarter(t *testing.T) {
	saJwt := &types.ServiceAccountJwt{
		NuwaRuntimeAddr: "http://test-addr",
		NuwaGatewayAddr: "http://gateway-addr",
		TlsConfig: &types.TLSConfig{
			HttpsInsecureSkipVerify: true,
			TlsCipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
	}

	po := NewColdStarter(saJwt, log.GetLogger())

	if po.nuwaConsoleAddr != saJwt.NuwaRuntimeAddr {
		t.Errorf("expected nuwaConsoleAddr %s, got %s", saJwt.NuwaRuntimeAddr, po.nuwaConsoleAddr)
	}
	if po.Client == nil {
		t.Error("expected non-nil client")
	}
}

func TestColdStart_Success(t *testing.T) {
	po := NewColdStarter(&types.ServiceAccountJwt{
		ServiceAccount: &types.ServiceAccount{},
		TlsConfig:      &types.TLSConfig{},
	}, log.GetLogger())
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(serviceaccount.GenerateJwtSignedHeaders, func(*fasthttp.Request, []byte, types.NuwaRuntimeInfo, *types.ServiceAccountJwt) error {
		return nil
	})
	patches.ApplyMethodFunc(&fasthttp.Client{}, "Do", func(*fasthttp.Request, *fasthttp.Response) error {
		return nil
	})

	err := po.ColdStart("funcKey", resspeckey.ResSpecKey{}, &types.NuwaRuntimeInfo{})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestDelPod_Success(t *testing.T) {
	po := NewColdStarter(&types.ServiceAccountJwt{
		ServiceAccount: &types.ServiceAccount{},
		TlsConfig:      &types.TLSConfig{},
	}, log.GetLogger())
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(serviceaccount.GenerateJwtSignedHeaders, func(*fasthttp.Request, []byte, types.NuwaRuntimeInfo, *types.ServiceAccountJwt) error {
		return nil
	})
	patches.ApplyMethodFunc(&fasthttp.Client{}, "Do", func(*fasthttp.Request, *fasthttp.Response) error {
		return nil
	})
	runtimeInfo := &types.NuwaRuntimeInfo{
		WisecloudRuntimeId: "test-runtime",
	}
	err := po.DelPod(runtimeInfo, "deploy1", "pod1")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestDelPod_Error(t *testing.T) {
	po := NewColdStarter(&types.ServiceAccountJwt{
		ServiceAccount: &types.ServiceAccount{},
		TlsConfig:      &types.TLSConfig{},
	}, log.GetLogger())
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(net.Listen, func(string, string) (net.Listener, error) {
		return nil, errors.New("test error")
	})

	err := po.DelPod(&types.NuwaRuntimeInfo{}, "deploy1", "pod1")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
