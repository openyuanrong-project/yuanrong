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

// Package etcd3 client
package etcd3

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
)

type KvMock struct {
}

func (k *KvMock) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k *KvMock) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return nil, nil
}

func (k *KvMock) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k *KvMock) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k *KvMock) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (k *KvMock) Txn(ctx context.Context) clientv3.Txn {
	//TODO implement me
	panic("implement me")
}

func TestInitEtcdClientOK(t *testing.T) {
	stopCh := make(chan struct{})
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc(buildClient, func(config *EtcdConfig) (*EtcdClient, error) {
			return &EtcdClient{clientExitCh: make(chan struct{}), cond: sync.NewCond(&sync.Mutex{})}, nil
		}),
	}
	defer func() {
		close(stopCh)
		for _, patch := range patches {
			patch.Reset()
		}
	}()

	convey.Convey("new RouteClient", t, func() {
		err := InitParam().
			WithRouteEtcdConfig(EtcdConfig{}).
			WithStopCh(stopCh).InitClient()
		convey.So(err, convey.ShouldBeNil)
		convey.So(GetRouterEtcdClient(), convey.ShouldNotBeNil)
	})

	convey.Convey("new MetadataClient", t, func() {
		err := InitParam().
			WithMetaEtcdConfig(EtcdConfig{}).
			WithStopCh(stopCh).InitClient()
		convey.So(err, convey.ShouldBeNil)
		convey.So(GetMetaEtcdClient(), convey.ShouldNotBeNil)
	})

	convey.Convey("new CAEMetadataClient", t, func() {
		err := InitParam().
			WithCAEMetaEtcdConfig(EtcdConfig{}).
			WithStopCh(stopCh).InitClient()
		convey.So(err, convey.ShouldBeNil)
		convey.So(GetCAEMetaEtcdClient(), convey.ShouldNotBeNil)
	})

	convey.Convey("new DataSystemEtcdClient", t, func() {
		err := InitParam().
			WithDataSystemEtcdConfig(EtcdConfig{}).
			WithStopCh(stopCh).InitClient()
		convey.So(err, convey.ShouldBeNil)
		convey.So(GetDataSystemEtcdClient(), convey.ShouldNotBeNil)
	})
}

func TestInitEtcdClientFail(t *testing.T) {
	var stopCh chan struct{}
	routerEtcdClient = nil
	metaEtcdClient = nil
	caeMetaEtcdClient = nil
	convey.Convey("new RouteClient", t, func() {
		err := InitParam().
			WithRouteEtcdConfig(EtcdConfig{}).
			WithStopCh(stopCh).InitClient()
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("new MetadataClient", t, func() {
		err := InitParam().
			WithMetaEtcdConfig(EtcdConfig{}).
			WithStopCh(stopCh).InitClient()
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("new RouteClient", t, func() {
		stopCh = make(chan struct{})
		err := InitParam().
			WithRouteEtcdConfig(EtcdConfig{}).
			WithStopCh(stopCh).InitClient()
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("new  CAE MetadataClient", t, func() {
		err := InitParam().
			WithCAEMetaEtcdConfig(EtcdConfig{}).
			WithStopCh(stopCh).InitClient()
		convey.So(err, convey.ShouldNotBeNil)
	})
	close(stopCh)
}

func TestInitEtcdClientKeepAliveOK(t *testing.T) {
	stopCh := make(chan struct{})
	kv := &KvMock{}
	client := &clientv3.Client{KV: kv}
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc(clientv3.New, func(cfg clientv3.Config) (*clientv3.Client, error) {
			return client, nil
		}),
		gomonkey.ApplyMethod(reflect.TypeOf(kv), "Get", func(_ *KvMock, ctx context.Context, key string,
			opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
			return nil, nil
		}),
		gomonkey.ApplyGlobalVar(&keepConnAliveTTL, time.Duration(100)*time.Millisecond),
	}
	defer func() {
		close(stopCh)
		for _, patch := range patches {
			patch.Reset()
		}
	}()

	convey.Convey("etcd client alive", t, func() {
		err := InitParam().
			WithRouteEtcdConfig(EtcdConfig{}).
			WithStopCh(stopCh).InitClient()
		time.Sleep(200 * time.Millisecond)
		convey.So(err, convey.ShouldBeNil)
		convey.So(GetRouterEtcdClient().GetEtcdStatusNow(), convey.ShouldEqual, true)
	})
}

func TestInitEtcdClientKeepAliveReconnect(t *testing.T) {
	routerEtcdClient = nil
	metaEtcdClient = nil
	stopCh := make(chan struct{})
	kv := &KvMock{}
	client := &clientv3.Client{KV: kv}
	patches := []*gomonkey.Patches{
		gomonkey.ApplyFunc(clientv3.New, func(cfg clientv3.Config) (*clientv3.Client, error) {
			return client, nil
		}),
		gomonkey.ApplyMethod(reflect.TypeOf(kv), "Get",
			func(_ *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return nil, errors.New("lost connection")
			}),
		gomonkey.ApplyMethod(reflect.TypeOf(client), "Close", func(_ *clientv3.Client) error {
			return nil
		}),
		gomonkey.ApplyGlobalVar(&keepConnAliveTTL, time.Duration(100)*time.Millisecond),
	}
	defer func() {
		close(stopCh)
		for _, patch := range patches {
			patch.Reset()
		}
	}()

	convey.Convey("lost etcd client and reconnect", t, func() {
		err := InitParam().
			WithRouteEtcdConfig(EtcdConfig{}).
			WithStopCh(stopCh).InitClient()
		time.Sleep(200 * time.Millisecond)
		convey.So(err, convey.ShouldBeNil)
		convey.So(GetRouterEtcdClient().GetEtcdStatusNow(), convey.ShouldEqual, false)

		patches = append(patches, gomonkey.ApplyMethod(reflect.TypeOf(kv), "Get",
			func(_ *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return nil, nil
			}))
		time.Sleep(200 * time.Millisecond)
		convey.So(GetRouterEtcdClient().GetEtcdStatusNow(), convey.ShouldEqual, true)
		convey.So(GetRouterEtcdClient().GetEtcdStatusLostContact(), convey.ShouldEqual, true)
	})
}

func TestInitMetaEtcdClient(t *testing.T) {
	convey.Convey("InitMetaEtcdClient", t, func() {
		convey.Convey("failed to init", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdInitParam{}), "InitClient",
				func(e *EtcdInitParam) error {
					return errors.New("failed to init")
				}).Reset()
			stop := make(chan struct{})
			err := InitMetaEtcdClient(EtcdConfig{}, alarm.Config{}, stop)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("failed to heat beat", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdInitParam{}), "InitClient",
				func(e *EtcdInitParam) error {
					metaEtcdClient = &EtcdClient{}
					return nil
				}).Reset()
			defer gomonkey.ApplyMethod(reflect.TypeOf(metaEtcdClient), "EtcdHeatBeat", func(e *EtcdClient) error {
				return errors.New("failed to heart beat")
			}).Reset()
			stop := make(chan struct{})
			err := InitMetaEtcdClient(EtcdConfig{}, alarm.Config{}, stop)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestInitCAEMetaEtcdClient(t *testing.T) {
	convey.Convey("InitCAEMetaEtcdClient", t, func() {
		convey.Convey("failed to init", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdInitParam{}), "InitClient",
				func(e *EtcdInitParam) error {
					return errors.New("failed to init")
				}).Reset()
			stop := make(chan struct{})
			err := InitMetaEtcdClient(EtcdConfig{}, alarm.Config{}, stop)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("failed to heat beat", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdInitParam{}), "InitClient",
				func(e *EtcdInitParam) error {
					caeMetaEtcdClient = &EtcdClient{}
					return nil
				}).Reset()
			defer gomonkey.ApplyMethod(reflect.TypeOf(caeMetaEtcdClient), "EtcdHeatBeat", func(e *EtcdClient) error {
				return errors.New("failed to heart beat")
			}).Reset()
			stop := make(chan struct{})
			err := InitCAEMetaEtcdClient(EtcdConfig{}, alarm.Config{}, stop)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestInitDataSystemEtcdClient(t *testing.T) {
	convey.Convey("InitDataSystemEtcdClient", t, func() {
		convey.Convey("failed to init", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdInitParam{}), "InitClient",
				func(e *EtcdInitParam) error {
					return errors.New("failed to init")
				}).Reset()
			stop := make(chan struct{})
			err := InitDataSystemEtcdClient(EtcdConfig{}, alarm.Config{}, stop)
			convey.So(err.Error(), convey.ShouldContainSubstring, "failed to init")
		})
		convey.Convey("failed to heat beat", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdInitParam{}), "InitClient",
				func(e *EtcdInitParam) error {
					dataSystemEtcdClient = &EtcdClient{}
					return nil
				}).Reset()
			defer gomonkey.ApplyMethod(reflect.TypeOf(dataSystemEtcdClient), "EtcdHeatBeat", func(e *EtcdClient) error {
				return errors.New("failed to heart beat")
			}).Reset()
			stop := make(chan struct{})
			err := InitDataSystemEtcdClient(EtcdConfig{}, alarm.Config{}, stop)
			convey.So(err.Error(), convey.ShouldContainSubstring, "failed to heart beat")
		})
	})
}

func TestInitRouterEtcdClient(t *testing.T) {
	convey.Convey("InitMetaEtcdClient", t, func() {
		convey.Convey("failed to init", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdInitParam{}), "InitClient",
				func(e *EtcdInitParam) error {
					return errors.New("failed to init")
				}).Reset()
			stop := make(chan struct{})
			err := InitRouterEtcdClient(EtcdConfig{}, alarm.Config{}, stop)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("failed to heat beat", func() {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdInitParam{}), "InitClient",
				func(e *EtcdInitParam) error {
					routerEtcdClient = &EtcdClient{}
					return nil
				}).Reset()
			defer gomonkey.ApplyMethod(reflect.TypeOf(routerEtcdClient), "EtcdHeatBeat", func(e *EtcdClient) error {
				return errors.New("failed to heart beat")
			}).Reset()
			stop := make(chan struct{})
			err := InitRouterEtcdClient(EtcdConfig{}, alarm.Config{}, stop)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func Test_reportOrClearAlarm(t *testing.T) {
	convey.Convey("reportOrClearAlarm", t, func() {
		convey.Convey("no test assertion", func() {
			e := EtcdClient{isAlarmEnable: true, etcdType: Router}
			e.reportOrClearAlarm(alarm.GenerateAlarmLog, "告警", "INFO")
		})
	})
}

func Test_AZPrefixProcess(t *testing.T) {
	convey.Convey("test AZPrefix", t, func() {
		convey.Convey("AttachAZPrefix", func() {
			e := EtcdClient{config: &EtcdConfig{
				AZPrefix: "az1",
			}}
			key := e.AttachAZPrefix("/sn/instance/xxx")
			convey.So(key, convey.ShouldEqual, "/az1/sn/instance/xxx")
			e.config.AZPrefix = ""
			key = e.AttachAZPrefix("/sn/instance/xxx")
			convey.So(key, convey.ShouldEqual, "/sn/instance/xxx")
		})
		convey.Convey("DetachAZPrefix", func() {
			e := EtcdClient{config: &EtcdConfig{
				AZPrefix: "az1",
			}}
			key := e.DetachAZPrefix("/az1/sn/instance/xxx")
			convey.So(key, convey.ShouldEqual, "/sn/instance/xxx")
			e.config.AZPrefix = ""
			key = e.DetachAZPrefix("/sn/instance/xxx")
			convey.So(key, convey.ShouldEqual, "/sn/instance/xxx")
		})
	})
}
