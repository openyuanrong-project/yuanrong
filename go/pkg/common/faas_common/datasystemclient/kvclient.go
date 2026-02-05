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

// Package datasystemclient is data system kv client.
package datasystemclient

import (
	"fmt"
	"runtime"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// Option options for kv client
type Option struct {
	TenantID  string
	NodeIP    string
	Cluster   string
	WriteMode api.WriteModeEnum
	TTLSecond uint32
}

// KVPutWithRetry put kv to ds with retry
func KVPutWithRetry(key string, value []byte, option *Option, traceID string) error {
	log.GetLogger().Debugf("datasystem kv put %s, %s, traceID: %s", key, string(value), traceID)
	config := &Config{
		invalidIP:       []string{},
		useLastUsedNode: true,

		TenantID:     option.TenantID,
		NodeIP:       option.NodeIP,
		Cluster:      option.Cluster,
		KeyPrefix:    key,
		NoNeedGenKey: true,
	}
	setParam := api.SetParam{
		WriteMode: option.WriteMode,
		TTLSecond: option.TTLSecond,
	}
	for {
		retry, err := kvPut(value, setParam, config, traceID)
		if err == nil {
			return nil
		}
		if retry {
			log.GetLogger().Debugf("put with key will retry, failed ip: %s,traceID: %s, err: %s",
				config.NodeIP, traceID, err.Error())
			config.invalidIP = append(config.invalidIP, config.NodeIP)
			config.NodeIP = ""
			continue
		}
		log.GetLogger().Errorf("get with key failed err: %s,NodeIP: %s,traceID: %s", err.Error(),
			config.NodeIP, traceID)
		return err
	}
}

// KVGetWithRetry get kv to ds with retry
func KVGetWithRetry(key string, option *Option, traceID string) ([]byte, error) {
	log.GetLogger().Debugf("datasystem kv get %s, traceID: %s", key, traceID)
	config := &Config{
		invalidIP:       []string{},
		useLastUsedNode: true,

		TenantID: option.TenantID,
		NodeIP:   option.NodeIP,
		Cluster:  option.Cluster,
	}
	for {
		resp, retry, err := kvGet(key, config, traceID)
		if err == nil {
			log.GetLogger().Debugf("datasystem kv get %s, %s, traceID: %s", key, string(resp), traceID)
			return resp, nil
		}
		if retry {
			log.GetLogger().Debugf("get with key will retry, failed ip: %s,traceID: %s, err: %s",
				config.NodeIP, traceID, err.Error())
			config.invalidIP = append(config.invalidIP, config.NodeIP)
			config.NodeIP = ""
			continue
		}
		log.GetLogger().Errorf("get with key failed err: %s,NodeIP: %s,traceID: %s", err.Error(),
			config.NodeIP, traceID)
		return nil, err
	}
}

// KVDelWithRetry del kv to ds with retry
func KVDelWithRetry(key string, option *Option, traceID string) error {
	log.GetLogger().Debugf("datasystem kv del %s, traceID: %s", key, traceID)
	config := &Config{
		invalidIP:       []string{},
		useLastUsedNode: true,

		TenantID: option.TenantID,
		NodeIP:   option.NodeIP,
		Cluster:  option.Cluster,
	}
	for {
		retry, err := kvDel(key, config, traceID)
		if err == nil {
			return nil
		}
		if retry {
			log.GetLogger().Debugf("del with key will retry, failed ip: %s,traceID: %s, err: %s",
				config.NodeIP, traceID, err.Error())
			config.invalidIP = append(config.invalidIP, config.NodeIP)
			config.NodeIP = ""
			continue
		}
		log.GetLogger().Errorf("get with key failed err: %s,NodeIP: %s,traceID: %s", err.Error(),
			config.NodeIP, traceID)
		return err
	}
}

func kvPut(value []byte, param api.SetParam, config *Config, traceID string) (bool, error) {
	dsClient, retry, err := getClient(config, traceID)
	if err != nil {
		return retry, err
	}
	if dsClient.kvClient == nil {
		return false, fmt.Errorf("dsclient is nil")
	}
	key, _, err := getDataSystemKey(config, dsClient, traceID)
	if err != nil {
		return false, err
	}
	runtime.LockOSThread()
	dsClient.kvClient.SetTraceID(traceID)
	if err = localClientLibruntime.SetTenantID(config.TenantID); err != nil {
		runtime.UnlockOSThread()
		return false, err
	}
	errInfo := dsClient.kvClient.KVSet(key, value, param)
	runtime.UnlockOSThread()
	if errInfo.IsError() {
		if shouldRetry(errInfo.Code) {
			return true, errInfo.Err
		}
		return false, errInfo.Err
	}
	return false, nil
}

func kvGet(key string, config *Config, traceID string) ([]byte, bool, error) {
	dsClient, retry, err := getClient(config, traceID)
	if err != nil {
		return nil, retry, err
	}
	if dsClient.kvClient == nil {
		return nil, false, fmt.Errorf("dsclient is nil")
	}
	runtime.LockOSThread()
	dsClient.kvClient.SetTraceID(traceID)
	if err = localClientLibruntime.SetTenantID(config.TenantID); err != nil {
		runtime.UnlockOSThread()
		return nil, false, err
	}
	resp, errInfo := dsClient.kvClient.KVGet(key)
	runtime.UnlockOSThread()
	retry, err = checkStatus(errInfo, config, traceID)
	if err != nil {
		return nil, retry, err
	}
	return resp, false, nil
}

func kvDel(key string, config *Config, traceID string) (bool, error) {
	dsClient, retry, err := getClient(config, traceID)
	if err != nil {
		return retry, err
	}
	if dsClient.kvClient == nil {
		return false, fmt.Errorf("dsclient is nil")
	}
	runtime.LockOSThread()
	dsClient.kvClient.SetTraceID(traceID)
	if err = localClientLibruntime.SetTenantID(config.TenantID); err != nil {
		runtime.UnlockOSThread()
		return false, err
	}
	errInfo := dsClient.kvClient.KVDel(key)
	runtime.UnlockOSThread()
	if errInfo.IsError() {
		if shouldRetry(errInfo.Code) {
			return true, errInfo.Err
		}
		return false, errInfo.Err
	}
	return false, nil
}
