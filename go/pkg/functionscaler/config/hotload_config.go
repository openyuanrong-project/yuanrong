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

// Package config -
package config

import (
	"encoding/json"
	"io/ioutil"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/monitor"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

const (
	// ConfigFilePath defines config file path of frontend
	configFilePath = "/home/sn/config/config.json"
)

var (
	configWatcher         monitor.FileWatcher
	configChangedCallback ChangedCallback
)

// ChangedCallback config change callback func
type ChangedCallback func()

// WatchConfig -
func WatchConfig(configPath string, stopCh <-chan struct{}, callback ChangedCallback) error {

	watcher, err := monitor.CreateFileWatcher(stopCh)
	if err != nil {
		return err
	}
	configWatcher = watcher
	configChangedCallback = callback
	configWatcher.RegisterCallback(configPath, hotLoadConfig)
	return nil
}

func hotLoadConfig(filename string, opType monitor.OpType) {
	log.GetLogger().Infof("file %s hot load start", filename)
	config, err := loadConfig(filename)
	if err != nil {
		log.GetLogger().Errorf("hotLoadConfig failed file: %s, opType: %d, err: %s",
			filename, opType, err.Error())
		return
	}
	hotLoadMetaEtcdConfig(config)
	hotLoadRouterEtcdConfig(config)
	hotLoadAutoScaleConfig(config)
	if configChangedCallback != nil {
		configChangedCallback()
	}
}

func loadConfig(configPath string) (*types.Configuration, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.GetLogger().Errorf("read file error, file path is %s", configPath)
		return nil, err
	}
	config := &types.Configuration{}
	err = json.Unmarshal(data, config)
	if err != nil {
		log.GetLogger().Errorf("failed to parse the config data: %s", err)
		return nil, err
	}
	err = loadFunctionConfig(config)
	if err != nil {
		return nil, err
	}
	return config, err
}

func hotLoadAutoScaleConfig(newAllConfig *types.Configuration) {
	if newAllConfig.AutoScaleConfig.SLAQuota > 0 && newAllConfig.AutoScaleConfig.BurstScaleNum > 0 &&
		newAllConfig.AutoScaleConfig.ScaleDownTime > 0 {
		GlobalConfig.AutoScaleConfig = newAllConfig.AutoScaleConfig
		autoScaleConfig := GlobalConfig.AutoScaleConfig
		log.GetLogger().Infof("autoScaleConfig  update, SLAQuota: %d,BurstScaleNum: %d,ScaleDownTime: %d ",
			autoScaleConfig.SLAQuota, autoScaleConfig.BurstScaleNum, autoScaleConfig.ScaleDownTime)
	}
	return
}

func hotLoadMetaEtcdConfig(newAllConfig *types.Configuration) {
	if newAllConfig.MetaETCDConfig.Servers != nil && len(newAllConfig.MetaETCDConfig.Servers) > 0 {
		newConfig := newAllConfig.MetaETCDConfig
		oldConfig := GlobalConfig.MetaETCDConfig
		oldConfig.Servers = newConfig.Servers
		log.GetLogger().Infof("etcd serverList update, new: %v", newConfig.Servers)
	}
	return
}

func hotLoadRouterEtcdConfig(newAllConfig *types.Configuration) {
	if newAllConfig.RouterETCDConfig.Servers != nil && len(newAllConfig.RouterETCDConfig.Servers) > 0 {
		newConfig := newAllConfig.RouterETCDConfig
		oldConfig := GlobalConfig.RouterETCDConfig
		oldConfig.Servers = newConfig.Servers
		log.GetLogger().Infof("etcd serverList update, new: %v", newConfig.Servers)
	}
	return
}
