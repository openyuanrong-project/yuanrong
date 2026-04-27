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

// Package http -
package http

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"

	"yuanrong.org/kernel/runtime/faassdk/handler"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/libruntime/api"
	log "yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

const (
	bootstrapCommand  = "/bin/bash"
	bootstrapFilename = "bootstrap"
)

type processState struct {
	exitCode int
	message  string
}

// HttpHandler handles http function initialization and invocation
type HttpHandler struct {
	*basicHandler
	process  *os.Process
	waitChan chan processState
}

// NewHttpHandler creates HttpHandler
func NewHttpHandler(funcSpec *types.FuncSpec, client api.LibruntimeAPI) handler.ExecutorHandler {
	return &HttpHandler{
		basicHandler: newBasicHandler(funcSpec, client),
		waitChan:     make(chan processState, 1),
	}
}

func (hh *HttpHandler) bootstrap() error {
	delegateDownloadPath, err := handler.GetUserCodePath()
	if err != nil {
		return err
	}
	bootstrapPath := path.Join(delegateDownloadPath, bootstrapFilename)
	_, err = os.Stat(bootstrapPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.GetLogger().Errorf("bootstrap file %s not exist for http function", bootstrapPath)
			return errors.New("bootstrap file not exist")
		}
		log.GetLogger().Errorf("failed to check stat of bootstrap file %s for http function", bootstrapPath)
		return errors.New("failed to check stat of bootstrap")
	}
	cmd := exec.Command(bootstrapCommand, bootstrapPath)
	// make suer bootstrap can find http server binary
	cmd.Dir = delegateDownloadPath
	// make sure subprocess can be killed
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		log.GetLogger().Errorf("failed to execute bootstrap for http function error %s", err.Error())
		return err
	}
	hh.process = cmd.Process
	go hh.waitProcess()
	return nil
}

func (hh *HttpHandler) waitProcess() {
	defer close(hh.waitChan)
	if hh.process == nil {
		log.GetLogger().Errorf("failed to wait process, process is nil")
		return
	}
	state, err := hh.process.Wait()
	if err != nil {
		log.GetLogger().Errorf("failed to wait process error %s", err.Error())
		return
	}
	message := fmt.Sprintf("process exits with %d", state.ExitCode())
	if state.ExitCode() != 0 {
		message += fmt.Sprintf(" error: %s", state.String())
	}
	log.GetLogger().Warnf(message)
	hh.monitor.ErrChan <- errors.New(message)
}

func (hh *HttpHandler) killAllProcesses() {
	log.GetLogger().Warnf("killing all processes of http function %s", hh.funcSpec.FuncMetaData.FunctionName)
	if hh.process == nil {
		log.GetLogger().Errorf("failed to kill process, process is nil")
		return
	}
	gid, err := syscall.Getpgid(hh.process.Pid)
	if err != nil {
		log.GetLogger().Errorf("failed to get gid of process %d", hh.process.Pid)
		gid = hh.process.Pid
	}
	err = syscall.Kill(-gid, syscall.SIGKILL)
	if err != nil {
		log.GetLogger().Errorf("failed to kill http function process error %s", err.Error())
	}
}

// InitHandler will bring up user's http server and wait until its ready then send init request
// args[0]: function specification
// args[1]: create params
func (hh *HttpHandler) InitHandler(args []api.Arg, dsClient api.LibruntimeAPI) ([]byte, error) {
	hh.setBootstrapFunc(hh.bootstrap)
	rsp, err := hh.basicHandler.InitHandler(args, dsClient)
	if err != nil {
		hh.killAllProcesses()
	}
	return rsp, err
}

// ShutDownHandler handles shutdown
func (hh *HttpHandler) ShutDownHandler(gracePeriodSecond uint64) error {
	err := hh.basicHandler.ShutDownHandler(gracePeriodSecond)
	hh.killAllProcesses()
	return err
}

// HealthCheckHandler handles health check
func (hh *HttpHandler) HealthCheckHandler() (api.HealthType, error) {
	return api.Healthy, nil
}
