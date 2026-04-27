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

// Package etcd3 -
package etcd3

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
	"k8s.io/api/core/v1"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

func TestGetSharedEtcdClient(t *testing.T) {
	etcdConfig123 := &EtcdConfig{
		Servers: []string{"1", "2", "3"},
	}
	convey.Convey("get client failed", t, func() {
		defer gomonkey.ApplyFunc(clientv3.New, func(cfg clientv3.Config) (*clientv3.Client, error) {
			return nil, errors.New("some error")
		}).Reset()
		_, err := GetSharedEtcdClient(etcdConfig123)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("get client success", t, func() {
		defer gomonkey.ApplyFunc(clientv3.New, func(cfg clientv3.Config) (*clientv3.Client, error) {
			return &clientv3.Client{}, nil
		}).Reset()
		_, err := GetSharedEtcdClient(etcdConfig123)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("load client", t, func() {
		_, err := GetSharedEtcdClient(etcdConfig123)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetValueFromEtcdWithRetry(t *testing.T) {
	funcKey := "123/testFunc/1"
	tenantID, funcName, funcVersion := utils.ParseFuncKey(funcKey)
	silentEtcdKey := fmt.Sprintf(constant.SilentFuncKey, tenantID, funcName, funcVersion)
	convey.Convey("Test GetValueFromEtcdWithRetry", t, func() {
		convey.Convey("etcd connection loss", func() {
			etcdClient := &EtcdClient{}
			_, err := GetValueFromEtcdWithRetry(silentEtcdKey, etcdClient)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("get values error", func() {
			etcdClient := &EtcdClient{
				etcdStatusAfterLostContact: true,
				Client:                     &clientv3.Client{},
			}
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdClient{}), "GetValues",
				func(_ *EtcdClient, ctxInfo EtcdCtxInfo, key string, opts ...clientv3.OpOption) ([]string, error) {
					return nil, errors.New("error")
				}).Reset()
			_, err := GetValueFromEtcdWithRetry(silentEtcdKey, etcdClient)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("value got from etcd is empty", func() {
			etcdClient := &EtcdClient{
				etcdStatusAfterLostContact: true,
				Client:                     &clientv3.Client{},
			}
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdClient{}), "GetValues",
				func(_ *EtcdClient, ctxInfo EtcdCtxInfo, key string, opts ...clientv3.OpOption) ([]string, error) {
					return []string{}, nil
				}).Reset()
			_, err := GetValueFromEtcdWithRetry(silentEtcdKey, etcdClient)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("fetch success", func() {
			etcdClient := &EtcdClient{
				etcdStatusAfterLostContact: true,
				Client:                     &clientv3.Client{},
			}
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdClient{}), "GetValues",
				func(_ *EtcdClient, ctxInfo EtcdCtxInfo, key string, opts ...clientv3.OpOption) ([]string, error) {
					return []string{"silent func"}, nil
				}).Reset()
			value, err := GetValueFromEtcdWithRetry(silentEtcdKey, etcdClient)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(value), convey.ShouldEqual, "silent func")
		})
	})
}

func TestGenerateETCDClientCertsVolumesAndMounts(t *testing.T) {
	t.Run("builder is nil", func(t *testing.T) {
		volumesData, volumesMountData, err := GenerateETCDClientCertsVolumesAndMounts("test-secret", nil)
		assert.Empty(t, volumesData)
		assert.Empty(t, volumesMountData)
		assert.EqualError(t, err, "etcd volume builder is nil")
	})

	t.Run("normal case", func(t *testing.T) {
		builder := utils.NewVolumeBuilder()

		secretName := "test-secret"
		volumesData, volumesMountData, err := GenerateETCDClientCertsVolumesAndMounts(secretName, builder)
		assert.NoError(t, err)

		var volumes []v1.Volume
		err = json.Unmarshal([]byte(volumesData), &volumes)
		assert.NoError(t, err)
		assert.Len(t, volumes, 1)
		assert.Equal(t, etcdClientCerts, volumes[0].Name)
		assert.Equal(t, secretName, volumes[0].VolumeSource.Secret.SecretName)

		var volumeMounts []v1.VolumeMount
		err = json.Unmarshal([]byte(volumesMountData), &volumeMounts)
		assert.NoError(t, err)
		assert.Len(t, volumeMounts, 1)
		assert.Equal(t, etcdClientCerts, volumeMounts[0].Name)
		assert.Equal(t, etcdCertsMountPath, volumeMounts[0].MountPath)
	})
}

func TestSetETCDTLSConfig(t *testing.T) {
	t.Run("etcdConfig", func(t *testing.T) {
		SetETCDTLSConfig(nil)

		etcdConfig := &EtcdConfig{}

		SetETCDTLSConfig(etcdConfig)

		assert.Equal(t, etcdCaFile, etcdConfig.CaFile, "CaFile should be set correctly")
		assert.Equal(t, etcdCertFile, etcdConfig.CertFile, "CertFile should be set correctly")
		assert.Equal(t, etcdKeyFile, etcdConfig.KeyFile, "KeyFile should be set correctly")
		assert.Equal(t, etcdPassphraseFile, etcdConfig.PassphraseFile, "PassphraseFile should be set correctly")
	})
}
