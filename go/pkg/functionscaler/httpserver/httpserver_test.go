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

package httpserver

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/localauth"
	commtls "yuanrong.org/kernel/pkg/common/faas_common/tls"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	"yuanrong.org/kernel/pkg/functionscaler"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

type mockListener struct{}

func (m *mockListener) Accept() (net.Conn, error) {
	return nil, fmt.Errorf("failed to accept")
}

func (m *mockListener) Close() error {
	return nil
}
func (m *mockListener) Addr() net.Addr {
	return nil
}

func TestStartHTTPServer(t *testing.T) {
	rawConfig := config.GlobalConfig
	defer func() {
		config.GlobalConfig = rawConfig
	}()
	convey.Convey("TestStartHTTPServer", t, func() {
		os.Setenv("POD_IP", "127.0.0.1")
		config.GlobalConfig = types.Configuration{
			HTTPSConfig: &commtls.InternalHTTPSConfig{
				HTTPSEnable: false},
			ModuleConfig: &types.ModuleConfig{ServicePort: "8889"},
		}
		errChan := make(chan error, 1)
		_, err := StartHTTPServer(errChan)
		convey.So(err, convey.ShouldBeNil)
		time.Sleep(1 * time.Second)
	})

	convey.Convey("TestStartHTTPSServer", t, func() {
		os.Setenv("POD_IP", "127.0.0.1")
		config.GlobalConfig = types.Configuration{
			HTTPSConfig: &commtls.InternalHTTPSConfig{
				HTTPSEnable: true},
			ModuleConfig: &types.ModuleConfig{ServicePort: "8899"},
		}
		defer gomonkey.ApplyFunc(commtls.InitTLSConfig, func(config commtls.InternalHTTPSConfig) error {
			return nil
		}).Reset()
		defer gomonkey.ApplyFunc(commtls.GetClientTLSConfig, func() *tls.Config {
			return &tls.Config{}
		}).Reset()
		errChan := make(chan error, 1)
		_, err := StartHTTPServer(errChan)
		convey.So(err, convey.ShouldBeNil)
		time.Sleep(1 * time.Second)
	})

}

func TestStartServer_InvalidPodIP(t *testing.T) {
	originalPodIP := os.Getenv("POD_IP")
	defer os.Setenv("POD_IP", originalPodIP)

	os.Setenv("POD_IP", "invalid_ip")

	httpServer := &fasthttp.Server{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	err := startServer(httpServer)

	assert.NotNil(t, err)
	assert.Equal(t, "failed to get pod ip", err.Error())
}

func TestStartServer_FastHTTPListenAndServeTLS_Error(t *testing.T) {
	os.Setenv("POD_IP", "127.0.0.1")
	defer os.Unsetenv("POD_IP")

	httpServer := &fasthttp.Server{}
	config.GlobalConfig = types.Configuration{
		HTTPSConfig: &commtls.InternalHTTPSConfig{
			HTTPSEnable: true},
		ModuleConfig: &types.ModuleConfig{ServicePort: "8080"},
	}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(net.Listen, func(network, address string) (net.Listener, error) {
		return nil, fmt.Errorf("err")
	})

	err := startServer(httpServer)

	assert.NotNil(t, err)
}

func TestStartServer_ListenAndServe_Error(t *testing.T) {
	os.Setenv("POD_IP", "127.0.0.1")
	defer os.Unsetenv("POD_IP")

	httpServer := &fasthttp.Server{}

	config.GlobalConfig = types.Configuration{
		HTTPSConfig: &commtls.InternalHTTPSConfig{
			HTTPSEnable: false},
		ModuleConfig: &types.ModuleConfig{ServicePort: "8080"},
	}

	mockError := errors.New("mocked ListenAndServe error")

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyMethod(reflect.TypeOf(httpServer), "ListenAndServe", func(_ *fasthttp.Server, addr string) error {
		return mockError
	})

	err := startServer(httpServer)

	assert.NotNil(t, err)
	assert.Equal(t, mockError, err)
}

func TestRout(t *testing.T) {
	rawConfig := config.GlobalConfig
	defer func() {
		config.GlobalConfig = rawConfig
	}()
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc(etcd3.GetRouterEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
		gomonkey.ApplyFunc(etcd3.GetMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
		gomonkey.ApplyFunc(etcd3.GetCAEMetaEtcdClient, func() *etcd3.EtcdClient { return &etcd3.EtcdClient{} }),
	}
	defer func() {
		for _, patch := range patches {
			time.Sleep(100 * time.Millisecond)
			patch.Reset()
		}
	}()
	convey.Convey("TestRout", t, func() {
		convey.Convey("auth failed", func() {
			config.GlobalConfig = types.Configuration{
				AuthenticationEnable: true,
			}
			defer gomonkey.ApplyFunc(localauth.AuthCheckLocally, func(ak string, sk string,
				requestSign string, timestamp string, duration int) error {
				return errors.New("auth failed")
			}).Reset()
			ctx := &fasthttp.RequestCtx{}
			route(ctx)
			convey.So(ctx.Response.StatusCode(), convey.ShouldEqual, http.StatusUnauthorized)
		})
		convey.Convey("path error", func() {
			config.GlobalConfig = types.Configuration{}
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.URI().SetPath("/acquire")
			route(ctx)
			convey.So(ctx.Response.StatusCode(), convey.ShouldEqual, http.StatusInternalServerError)
		})
		convey.Convey("invoke unmarshal body error ", func() {
			config.GlobalConfig = types.Configuration{}
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.URI().SetPath(invokePath)
			ctx.Request.SetBody([]byte("aaa"))
			route(ctx)
			convey.So(ctx.Response.StatusCode(), convey.ShouldEqual, http.StatusInternalServerError)
		})
		convey.Convey("invoke scheduler is nil ", func() {
			config.GlobalConfig = types.Configuration{}
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.URI().SetPath(invokePath)
			args := []api.Arg{{Type: 1, Data: []byte("aaa")}}
			body, _ := json.Marshal(args)
			ctx.Request.SetBody(body)
			route(ctx)
			convey.So(ctx.Response.StatusCode(), convey.ShouldEqual, http.StatusInternalServerError)
		})
		convey.Convey("invoke ProcessInstanceRequestLibruntime success ", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&functionscaler.FaaSScheduler{}),
				"ProcessInstanceRequestLibruntime", func(_ *functionscaler.FaaSScheduler,
					args []api.Arg, traceID string) ([]byte, error) {
					return json.Marshal(&commonTypes.InstanceResponse{})
				}).Reset()
			defer gomonkey.ApplyFunc((*etcd3.EtcdWatcher).StartList, func(ew *etcd3.EtcdWatcher) {
				ew.ResultChan <- &etcd3.Event{
					Type:      etcd3.SYNCED,
					Key:       "",
					Value:     nil,
					PrevValue: nil,
					Rev:       0,
					ETCDType:  "",
				}
			}).Reset()
			defer gomonkey.ApplyFunc((*registry.FaasSchedulerRegistry).WaitForETCDList, func() {}).Reset()
			config.GlobalConfig = types.Configuration{}
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.URI().SetPath(invokePath)
			args := []api.Arg{{Type: 1, Data: []byte("aaa")}}
			stopCh := make(chan struct{})
			registry.InitRegistry(stopCh)
			functionscaler.InitGlobalScheduler(stopCh)
			body, _ := json.Marshal(args)
			ctx.Request.SetBody(body)
			route(ctx)
			convey.So(ctx.Response.StatusCode(), convey.ShouldEqual, http.StatusOK)
		})
	})
}

func TestFastHTTPListenAndServeTLS(t *testing.T) {
	server := fasthttp.Server{
		TLSConfig: &tls.Config{},
	}
	convey.Convey("FastHTTPListenAndServeTLS failed", t, func() {
		patch := gomonkey.ApplyFunc(net.Listen, func(network, address string) (net.Listener, error) {
			return &mockListener{}, errors.New("listen fail")
		})
		defer patch.Reset()
		err := fastHTTPListenAndServeTLS("123", &server)
		convey.So(err.Error(), convey.ShouldEqual, "listen fail")
	})
	convey.Convey("FastHTTPListenAndServeTLS server is nil", t, func() {
		patch := gomonkey.ApplyFunc(net.Listen, func(network, address string) (net.Listener, error) {
			return &mockListener{}, nil
		})
		defer patch.Reset()
		err := fastHTTPListenAndServeTLS("123", nil)
		convey.So(err.Error(), convey.ShouldEqual, "server or tls config is nil")
	})
}

func Test_auth(t *testing.T) {
	convey.Convey("Test auth", t, func() {
		ctx := &fasthttp.RequestCtx{}
		originalAuthenticationEnable := config.GlobalConfig.AuthenticationEnable
		defer func() {
			config.GlobalConfig.AuthenticationEnable = originalAuthenticationEnable
		}()
		convey.Convey("when authenticationEnable is false", func() {
			config.GlobalConfig.AuthenticationEnable = false
			err := auth(ctx)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("when sign is start with HmacSha256", func() {
			ctx.Request.Header.Set(constant.HeaderAuthorization, "HmacSha256 xxx")
			config.GlobalConfig.AuthenticationEnable = true
			err := auth(ctx)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when sign is not start with HmacSha256", func() {
			ctx.Request.Header.Set(constant.HeaderAuthorization, "xxx")
			ctx.Request.Header.Set(constant.HeaderAuthTimestamp, "xxx")
			config.GlobalConfig.AuthenticationEnable = true
			err := auth(ctx)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
