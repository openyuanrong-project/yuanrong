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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3"
	"k8s.io/api/core/v1"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

const (
	// etcdDialTimeout is the timeout for establishing a connection.
	etcdDialTimeout = 20 * time.Second

	// etcdKeepaliveTime is the time after which client pings the server to see if
	etcdKeepaliveTime = 30 * time.Second

	// etcdKeepaliveTimeout is the time that the client waits for a response for the
	etcdKeepaliveTimeout = 10 * time.Second

	etcdClientCerts = "etcd-client-certs"

	etcdCertsMountPath = "/home/snuser/resource/etcd"

	etcdCaFile = "/home/snuser/resource/etcd/ca.crt"

	etcdCertFile = "/home/snuser/resource/etcd/client.crt"

	etcdKeyFile = "/home/snuser/resource/etcd/client.key"

	etcdPassphraseFile = "/home/snuser/resource/etcd/passphrase"
)

const (
	retrySleepTime = 100 * time.Millisecond
	maxRetryTime   = 3
)

var (
	etcdClientMap sync.Map
)

// GetEtcdConfigKey generates key for etcd config
func GetEtcdConfigKey(etcdConfig *EtcdConfig) string {
	sort.Strings(etcdConfig.Servers)
	return strings.Join(etcdConfig.Servers, "#")
}

func createETCDClient(config *EtcdConfig) (*clientv3.Client, error) {
	cfg, err := GetEtcdAuthType(*config).GetEtcdConfig()
	if err != nil {
		log.GetLogger().Errorf("failed to create shared etcd client error %s", err.Error())
		return nil, err
	}
	cfg.DialTimeout = etcdDialTimeout
	cfg.DialKeepAliveTime = etcdKeepaliveTime
	cfg.DialKeepAliveTimeout = etcdKeepaliveTimeout
	cfg.Endpoints = config.Servers
	etcdClient, err := clientv3.New(*cfg)
	if err != nil {
		log.GetLogger().Errorf("failed to create shared etcd client error %s", err.Error())
		return nil, err
	}
	return etcdClient, nil
}

// GetSharedEtcdClient returns a shared etcd client
func GetSharedEtcdClient(etcdConfig *EtcdConfig) (*clientv3.Client, error) {
	etcdConfigKey := GetEtcdConfigKey(etcdConfig)
	obj, exist := etcdClientMap.Load(etcdConfigKey)
	var err error
	if !exist {
		if obj, err = createETCDClient(etcdConfig); err != nil {
			return nil, err
		}
	}
	etcdClient, ok := obj.(*clientv3.Client)
	if !ok {
		return nil, errors.New("etcd client type error")
	}
	etcdClientMap.Store(etcdConfigKey, etcdClient)
	return etcdClient, nil
}

// GetValueFromEtcdWithRetry query value from etcd and retry only in case of timeout
func GetValueFromEtcdWithRetry(key string, etcdClient *EtcdClient) ([]byte, error) {
	if etcdClient.GetEtcdStatusLostContact() == false || etcdClient.Client == nil {
		return nil, errors.New("etcd connection loss")
	}
	var (
		values []string
		err    error
	)
	for i := 1; i <= maxRetryTime; i++ {
		defaultEtcdCtx := CreateEtcdCtxInfoWithTimeout(context.Background(), DurationContextTimeout)
		values, err = etcdClient.GetValues(defaultEtcdCtx, key)
		if err == nil {
			break
		}
		if err != context.DeadlineExceeded {
			return nil, err
		}
		log.GetLogger().Errorf("get value from etcd with key %s timeout, try time %d", key, i)
		time.Sleep(retrySleepTime)
	}

	if len(values) == 0 {
		log.GetLogger().Errorf("failed to get value from etcd, key: %s", key)
		return nil, fmt.Errorf("the value got from etcd is empty")
	}

	return []byte(values[0]), err
}

// GenerateETCDClientCertsVolumesAndMounts -
func GenerateETCDClientCertsVolumesAndMounts(secretName string, builder *utils.VolumeBuilder) (string, string, error) {
	if builder == nil {
		return "", "", fmt.Errorf("etcd volume builder is nil")
	}
	builder.AddVolume(v1.Volume{Name: etcdClientCerts,
		VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: secretName}}})
	builder.AddVolumeMount(utils.ContainerRuntimeManager,
		v1.VolumeMount{Name: etcdClientCerts, MountPath: etcdCertsMountPath})
	volumesData, err := json.Marshal(builder.Volumes)
	if err != nil {
		return "", "", err
	}
	volumesMountData, err := json.Marshal(builder.Mounts[utils.ContainerRuntimeManager])
	if err != nil {
		return "", "", err
	}
	return string(volumesData), string(volumesMountData), nil
}

// SetETCDTLSConfig -
func SetETCDTLSConfig(etcdConfig *EtcdConfig) {
	if etcdConfig == nil {
		return
	}
	etcdConfig.CaFile = etcdCaFile
	etcdConfig.CertFile = etcdCertFile
	etcdConfig.KeyFile = etcdKeyFile
	etcdConfig.PassphraseFile = etcdPassphraseFile
}
