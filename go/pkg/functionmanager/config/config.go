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
	"fmt"

	"github.com/asaskevich/govalidator/v11"

	"yuanrong.org/kernel/pkg/common/faas_common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/functionmanager/types"
)

var cfg types.ManagerConfig

// GetConfig return the current config
func GetConfig() types.ManagerConfig {
	return cfg
}

// InitConfig is used to initialize the config
func InitConfig(data []byte) error {
	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return fmt.Errorf("failed to parse the config data: %s", err.Error())
	}
	if cfg.RouterEtcd.UseSecret {
		etcd3.SetETCDTLSConfig(&cfg.RouterEtcd)
	}
	if cfg.MetaEtcd.UseSecret {
		etcd3.SetETCDTLSConfig(&cfg.MetaEtcd)
	}
	_, err = govalidator.ValidateStruct(cfg)
	if err != nil {
		return fmt.Errorf("invalid config: %s", err.Error())
	}
	if cfg.SccConfig.Enable && crypto.InitializeSCC(cfg.SccConfig) != nil {
		return fmt.Errorf("failed to initialize scc")
	}
	return nil
}

// InitEtcd - init router etcd and meta etcd
func InitEtcd(stopCh <-chan struct{}) error {
	if &cfg == nil {
		return fmt.Errorf("config is not initialized")
	}
	if err := etcd3.InitParam().
		WithRouteEtcdConfig(cfg.RouterEtcd).
		WithStopCh(stopCh).
		WithAlarmSwitch(cfg.AlarmConfig.EnableAlarm).
		InitClient(); err != nil {
		return fmt.Errorf("faaSManager failed to init route etcd: %s", err.Error())
	}
	if err := etcd3.InitParam().
		WithMetaEtcdConfig(cfg.MetaEtcd).
		WithStopCh(stopCh).
		WithAlarmSwitch(cfg.AlarmConfig.EnableAlarm).
		InitClient(); err != nil {
		return fmt.Errorf("faaSManager failed to init metadata etcd: %s", err.Error())
	}
	return nil
}
