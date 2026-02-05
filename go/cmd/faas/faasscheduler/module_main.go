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

package main

import (
	"fmt"

	"github.com/valyala/fasthttp"

	"yuanrong.org/kernel/pkg/common/faas_common/autogc"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/signals"
	"yuanrong.org/kernel/pkg/common/faas_common/trafficlimit"
	"yuanrong.org/kernel/pkg/functionscaler"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/healthcheck"
	"yuanrong.org/kernel/pkg/functionscaler/httpserver"
	"yuanrong.org/kernel/pkg/functionscaler/instancequeue"
	"yuanrong.org/kernel/pkg/functionscaler/registry"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
	"yuanrong.org/kernel/pkg/functionscaler/state"
	"yuanrong.org/kernel/pkg/functionscaler/workermanager"
)

const (
	logFileName = "faas-scheduler"
	filePath    = "/home/sn/config/config.json"
	errChanSize = 2
)

func main() {
	defer func() {
		log.GetLogger().Sync()
	}()
	// init logger config
	err := log.InitRunLog(logFileName, true)
	if err != nil {
		fmt.Print("init logger error: " + err.Error())
		return
	}

	defer func() {
		if err != nil {
			fmt.Printf("panic, module: faasscheduler, err: %s\n", err.Error())
			log.GetLogger().Errorf("panic, module: scheduler, err: %s", err.Error())
		}
		log.GetLogger().Sync()
	}()
	err = config.InitModuleConfig()
	if err != nil {
		errMessage := fmt.Sprintf("init module config error: %s", err.Error())
		logAndPrintError(errMessage)
		return
	}
	autogc.InitAutoGOGC()
	stopCh := signals.WaitForSignal()
	if stopCh == nil {
		errMessage := "stopCh is nil"
		logAndPrintError(errMessage)
		return
	}
	if err = config.InitEtcd(stopCh); err != nil {
		errMessage := fmt.Sprintf("init etcd error: %s", err.Error())
		logAndPrintError(errMessage)
		return
	}
	if err = workermanager.InitLeaseInformer(stopCh); err != nil {
		errMessage := fmt.Sprintf("init lease informer error: %s", err.Error())
		logAndPrintError(errMessage)
		return
	}
	instancequeue.DisableCreateRetry()
	state.InitState()
	var stateByte []byte
	stateByte, err = state.GetStateByte()
	if err == nil && len(stateByte) != 0 {
		err = RecoverModuleScheduler(stateByte, stopCh)
		if err != nil {
			errMessage := fmt.Sprintf("failed to recover module scheduler ,err:%s", err.Error())
			logAndPrintError(errMessage)
			return
		}
	}
	if err = setupModuleScheduler(stopCh); err != nil {
		errMessage := fmt.Sprintf("failed to setup module scheduler,err:%s", err.Error())
		logAndPrintError(errMessage)
		return
	}
	registry.StartRegistry()
	config.ClearSensitiveInfo()
	errChan := make(chan error, errChanSize)
	httpServer, err := httpserver.StartHTTPServer(errChan)
	if err != nil {
		errMessage := fmt.Sprintf("failed to start http server, err: %s", err.Error())
		logAndPrintError(errMessage)
		return
	}
	err = healthcheck.StartHealthCheck(errChan)
	if err != nil {
		errMessage := fmt.Sprintf("failed to start health check, err: %s", err.Error())
		logAndPrintError(errMessage)
		return
	}
	if err = selfregister.RegisterToEtcd(stopCh); err != nil {
		errMessage := fmt.Sprintf("register to etcd error: %s", err.Error())
		logAndPrintError(errMessage)
	}
	waitShutdown(httpServer, stopCh, errChan)
}

func logAndPrintError(errMessage string) {
	log.GetLogger().Errorf(errMessage)
	fmt.Println(errMessage)
}

func setupModuleScheduler(stopCh <-chan struct{}) error {
	err := registry.InitRegistry(stopCh)
	if err != nil {
		return err
	}
	// WatchConfig failed do not return, just config hot load not enable
	if err := config.WatchConfig(filePath, stopCh, nil); err != nil {
		log.GetLogger().Warnf("WatchConfig %s failed, err %s", filePath, err.Error())
	}
	functionscaler.InitGlobalScheduler(stopCh)
	registry.ProcessETCDList()
	trafficlimit.SetFunctionLimitRate(config.GlobalConfig.FunctionLimitRate)
	return nil
}

// RecoverModuleScheduler -
func RecoverModuleScheduler(stateData []byte, stopCh <-chan struct{}) error {
	var err error
	log.GetLogger().Infof("trigger: RecoverModuleScheduler")
	if err = state.SetState(stateData); err != nil {
		return fmt.Errorf("module scheduler recover error:%s", err.Error())
	}
	state.RecoverStateRev()
	state.InitState()
	if err = setupModuleScheduler(stopCh); err != nil {
		log.GetLogger().Errorf("recover module scheduler error:%s", err.Error())
		return fmt.Errorf("module scheduler recover error:%s", err.Error())
	}
	if functionscaler.GetGlobalScheduler() != nil {
		functionscaler.GetGlobalScheduler().Recover()
	}
	registry.StartRegistry()
	config.ClearSensitiveInfo()
	errChan := make(chan error, errChanSize)
	httpServer, err := startServer(errChan)
	if err != nil {
		return err
	}
	if err = selfregister.RegisterToEtcd(stopCh); err != nil {
		return err
	}
	waitShutdown(httpServer, stopCh, errChan)
	return nil
}

func startServer(errChan chan error) (*fasthttp.Server, error) {
	var httpServer *fasthttp.Server
	var err error
	httpServer, err = httpserver.StartHTTPServer(errChan)
	if err != nil {
		return nil, fmt.Errorf("start fast http server error:%s", err.Error())
	}
	err = healthcheck.StartHealthCheck(errChan)
	if err != nil {
		return nil, fmt.Errorf("failed to start health check, err: %s", err.Error())
	}
	return httpServer, nil
}

func waitShutdown(server *fasthttp.Server, stopCh <-chan struct{}, errChan <-chan error) {
	if stopCh == nil || errChan == nil {
		errMessage := "input channel is nil"
		logAndPrintError(errMessage)
	}
	select {
	case <-stopCh:
		log.GetLogger().Infof("received termination signal")
		if server != nil {
			if err := server.Shutdown(); err != nil {
				errMessage := fmt.Sprintf("http server shutdowm error:%s", err.Error())
				logAndPrintError(errMessage)
			}
		}
	case err := <-errChan:
		errMessage := fmt.Sprintf("http server error:%s", err.Error())
		logAndPrintError(errMessage)
	}
}
