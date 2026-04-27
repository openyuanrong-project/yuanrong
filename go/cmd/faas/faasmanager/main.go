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

// Package main -
package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionmanager"
	"yuanrong.org/kernel/pkg/functionmanager/config"
	"yuanrong.org/kernel/pkg/functionmanager/healthcheck"
	"yuanrong.org/kernel/pkg/functionmanager/state"
)

var (
	// faasManager handles functions management for faas pattern
	faasManager *functionmanager.Manager
	stopCh      = make(chan struct{})
	errCh       = make(chan error)
)

// InitHandlerLibruntime is the init handler called by runtime
func InitHandlerLibruntime(args []api.Arg, libruntimeAPI api.LibruntimeAPI) ([]byte, error) {
	log.SetupLoggerLibruntime(libruntimeAPI.GetFormatLogger())
	log.GetLogger().Infof("trigger: faasmanager.InitHandler")
	if len(args) == 0 || libruntimeAPI == nil {
		return []byte(""), errors.New("init args empty")
	}
	if args[0].Type != api.Value {
		return []byte(""), errors.New("arg type error")
	}
	err := config.InitConfig(args[0].Data)
	if err != nil {
		return []byte(""), err
	}
	if err = config.InitEtcd(stopCh); err != nil {
		log.GetLogger().Errorf("failed to init etcd ,err:%s", err.Error())
		return []byte(""), err
	}
	state.InitState()
	stateByte, err := state.GetStateByte()
	if err == nil && len(stateByte) != 0 {
		return []byte(""), RecoverHandlerLibruntime(stateByte, libruntimeAPI)
	}
	if _, err = setupFaaSManagerLibruntime(libruntimeAPI); err != nil {
		return []byte(""), err
	}
	cfg := config.GetConfig()
	state.Update(&cfg)
	if faasManager != nil {
		go faasManager.WatchLeaseEvent()
	}
	if err = healthcheck.StartHealthCheck(errCh); err != nil {
		return []byte(""), err
	}
	return []byte(""), nil
}

// CallHandlerLibruntime is the call handler called by runtime
func CallHandlerLibruntime(args []api.Arg) ([]byte, error) {
	traceID := string(args[len(args)-1].Data)
	if faasManager == nil {
		return nil, fmt.Errorf("faas manager is not initialized, traceID: %s", traceID)
	}
	response := faasManager.ProcessSchedulerRequestLibruntime(args, traceID)
	if response == nil {
		return nil, fmt.Errorf("failed to process scheduler request, traceID: %s", traceID)
	}
	rspData, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	return rspData, nil
}

// CheckpointHandlerLibruntime is the checkpoint handler called by libruntime
func CheckpointHandlerLibruntime(checkpointID string) ([]byte, error) {
	return state.GetStateByte()
}

// RecoverHandlerLibruntime is the recover handler called by runtime
func RecoverHandlerLibruntime(stateData []byte, libruntimeAPI api.LibruntimeAPI) error {
	var err error
	log.SetupLoggerLibruntime(libruntimeAPI.GetFormatLogger())
	log.GetLogger().Infof("trigger: faasmanager.RecoverHandler")
	state.InitState()
	if err = state.SetState(stateData); err != nil {
		return fmt.Errorf("faaS manager recover error:%s", err.Error())
	}
	if _, err = setupFaaSManagerLibruntime(libruntimeAPI); err != nil {
		return fmt.Errorf("restart faaS controller libruntime error:%s", err.Error())
	}
	if faasManager != nil {
		faasManager.RecoverData()
	}
	cfg := config.GetConfig()
	state.Update(&cfg)
	if faasManager != nil {
		go faasManager.WatchLeaseEvent()
	}
	if err = healthcheck.StartHealthCheck(errCh); err != nil {
		return err
	}
	return nil
}

// ShutdownHandlerLibruntime is the shutdown handler called by libruntime
func ShutdownHandlerLibruntime(gracePeriodSecond uint64) error {
	log.GetLogger().Infof("trigger: faasmanager.ShutdownHandler")
	utils.SafeCloseChannel(stopCh)
	log.GetLogger().Sync()
	return nil
}

// SignalHandlerLibruntime is the signal handler called by libruntime
func SignalHandlerLibruntime(signal int, payload []byte) error {
	return nil
}

func setupFaaSManagerLibruntime(libruntimeAPI api.LibruntimeAPI) (interface{}, error) {
	var err error
	faasManager, err = functionmanager.NewFaaSManagerLibruntime(libruntimeAPI, stopCh)
	if err != nil {
		return "", err
	}
	return "", nil
}
