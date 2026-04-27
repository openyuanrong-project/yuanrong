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
	"errors"
	"fmt"
	"sync"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/faascontroller"
	"yuanrong.org/kernel/pkg/system_function_controller/state"
)

var (
	// faasController handles instance management for faasscheduler
	faasController *faascontroller.FaaSController
	stopCh         = make(chan struct{})
	shutdownOnce   sync.Once

	signalHandlerMap      = map[int32]func([]byte) error{}
	frontendUpdateSignal  = int32(65)
	schedulerUpdateSignal = int32(66)
	managerUpdateSignal   = int32(68)
)

// InitHandlerLibruntime is the init handler called by runtime based on multi libruntime
func InitHandlerLibruntime(args []api.Arg, libruntimeAPI api.LibruntimeAPI) ([]byte, error) {
	log.SetupLoggerLibruntime(libruntimeAPI.GetFormatLogger())
	var err error
	defer func() {
		if err != nil {
			fmt.Printf("panic, module: faascontroller, err: %s\n", err.Error())
			log.GetLogger().Errorf("panic, module: faascontroller, err: %s", err.Error())
		}
		log.GetLogger().Sync()
	}()
	log.GetLogger().Infof("trigger: faascontroller.InitHandler")
	if err = checkArgsLibruntime(args); err != nil {
		return []byte(""), err
	}
	if err = config.InitConfig(args[0].Data); err != nil {
		log.GetLogger().Errorf("failed to init config, err:%s", err.Error())
		return []byte(""), err
	}
	if err = config.InitEtcd(stopCh); err != nil {
		log.GetLogger().Errorf("failed to init etcd ,err:%s", err.Error())
		return []byte(""), err
	}
	state.InitState(config.GetFaaSControllerConfig().SchedulerExclusivity)
	if err = setupFaaSControllerLibruntime(libruntimeAPI); err != nil {
		return []byte(""), err
	}
	log.GetLogger().Infof("exit: faascontroller.InitHandler")
	return []byte(""), nil
}

func checkArgsLibruntime(args []api.Arg) error {
	if len(args) == 0 {
		log.GetLogger().Errorf("init args empty")
		return errors.New("init args empty")
	}
	if args[0].Type != api.Value {
		log.GetLogger().Errorf("arg type error")
		return errors.New("arg type error")
	}
	return nil
}

func setupFaaSControllerLibruntime(libruntimeAPI api.LibruntimeAPI) error {
	var err error
	// create Faas controller instance
	faasController, err = faascontroller.NewFaaSControllerLibruntime(libruntimeAPI, stopCh)
	return PrepareFaasController(err)
}

// PrepareFaasController -
func PrepareFaasController(err error) error {
	if err != nil {
		log.GetLogger().Errorf("failed to create faas controller instance")
		return err
	}
	signalHandlerMap[frontendUpdateSignal] = faasController.FrontendSignalHandler
	signalHandlerMap[schedulerUpdateSignal] = faasController.SchedulerSignalHandler
	signalHandlerMap[managerUpdateSignal] = faasController.ManagerSignalHandler
	return nil
}

// CallHandlerLibruntime is the call handler called by runtime
func CallHandlerLibruntime(args []api.Arg) ([]byte, error) {
	return nil, nil
}

// CheckpointHandlerLibruntime is the checkpoint handler called by runtime
func CheckpointHandlerLibruntime(checkpointID string) ([]byte, error) {
	return state.GetStateByte()
}

// RecoverHandlerLibruntime is the recover handler called by runtime based on multi libruntime
func RecoverHandlerLibruntime(stateData []byte, libruntimeAPI api.LibruntimeAPI) error {
	var err error
	log.SetupLoggerLibruntime(libruntimeAPI.GetFormatLogger())
	log.GetLogger().Infof("trigger: faascontroller.RecoverHandler")
	if err = state.SetState(stateData); err != nil {
		return fmt.Errorf("recover faaS controller error:%s", err.Error())
	}
	state.InitState(config.GetFaaSControllerConfig().SchedulerExclusivity)
	if err = config.RecoverConfig(); err != nil {
		return fmt.Errorf("recover config error:%s", err.Error())
	}
	if err = config.InitEtcd(stopCh); err != nil {
		log.GetLogger().Errorf("failed to init etcd ,err:%s", err.Error())
		return err
	}
	if err = setupFaaSControllerLibruntime(libruntimeAPI); err != nil {
		return fmt.Errorf("restart faaS controller error:%s", err.Error())
	}
	state.Update(config.GetFaaSControllerConfig())
	return nil
}

// ShutdownHandlerLibruntime is the shutdown handler called by runtime based on multi libruntime
func ShutdownHandlerLibruntime(gracePeriodSecond uint64) error {
	log.GetLogger().Infof("trigger: faascontroller.ShutdownHandlerLibruntime")
	utils.SafeCloseChannel(stopCh)
	log.GetLogger().Infof("faascontrolerLibruntime exit")
	log.GetLogger().Sync()
	return nil
}

// SignalHandlerLibruntime is the signal handler called by runtime
func SignalHandlerLibruntime(signal int, payload []byte) error {
	log.GetLogger().Infof("trigger: faascontroller.SignalHandlerLibruntime signal:%d", signal)
	handler, ok := signalHandlerMap[int32(signal)]
	if !ok {
		log.GetLogger().Errorf("signal: %d, not found handler", signal)
		return fmt.Errorf("not found signal handler libruntime")
	}
	err := handler(payload)
	if err != nil {
		return err
	}
	return nil
}
