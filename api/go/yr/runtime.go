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

// Package yr for init and start
package yr

import (
	"os"

	"yuanrong.org/kernel/runtime/libruntime"
	"yuanrong.org/kernel/runtime/libruntime/common"
	"yuanrong.org/kernel/runtime/libruntime/config"
)

// InitRuntime init runtime
func InitRuntime() error {
	conf := common.GetConfig()
	intfs := newActorFuncExecutionIntfs()
	hooks := config.HookIntfs{
		LoadFunctionCb:      intfs.LoadFunction,
		FunctionExecutionCb: intfs.FunctionExecute,
		CheckpointCb:        intfs.Checkpoint,
		RecoverCb:           intfs.Recover,
		ShutdownCb:          intfs.Shutdown,
		SignalCb:            intfs.Signal,
	}

	yrConf := &Config{
		InCluster:      true,
		Mode:           ClusterMode,
		DataSystemAddr: os.Getenv("DATASYSTEM_ADDR"),
		LogLevel:       conf.LogLevel,
		AutoStart:      true,
	}

	yrFlagsConfig := &FlagsConfig{
		RuntimeID:  conf.RuntimeID,
		InstanceID: conf.InstanceID,
		LogPath:    conf.LogPath,
		JobID:      conf.JobID,
		Hooks:      hooks,
	}

	_, err := InitWithFlags(yrConf, yrFlagsConfig)
	return err
}

// Run begins loop processing the received request.
func Run() {
	libruntime.ReceiveRequestLoop()
}
