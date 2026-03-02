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

// Package faassdk for init and start
package faassdk

import (
	"fmt"
	"os"

	"yuanrong.org/kernel/runtime/libruntime"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common"
	"yuanrong.org/kernel/runtime/libruntime/config"
	"yuanrong.org/kernel/runtime/libruntime/pool"
)

// Run begins loop processing the received request.
func Run() {
	libruntime.ReceiveRequestLoop()
}

// InitRuntime init runtime
func InitRuntime() error {
	conf := common.GetConfig()
	intfs := newFaaSFuncExecutionIntfs()
	runtimeConf := config.Config{
		GrpcAddress:           conf.GrpcAddress,
		FunctionSystemAddress: conf.FSAddress,
		DataSystemAddress:     os.Getenv("DATASYSTEM_ADDR"),
		JobID:                 conf.JobID,
		RuntimeID:             conf.RuntimeID,
		InstanceID:            conf.InstanceID,
		FunctionName:          conf.FunctionName,
		LogDir:                conf.LogPath,
		LogLevel:              conf.LogLevel,
		InCluster:             true,
		IsDriver:              conf.DriverMode,
		EnableMTLS:            conf.EnableMTLS,
		PrivateKeyPath:        conf.PrivateKeyPath,
		CertificateFilePath:   conf.CertificateFilePath,
		VerifyFilePath:        conf.VerifyFilePath,
		PrivateKeyPaaswd:      conf.PrivateKeyPaaswd,
		Api:                   api.FaaSApi,
		Hooks: config.HookIntfs{
			LoadFunctionCb:      intfs.LoadFunction,
			FunctionExecutionCb: intfs.FunctionExecute,
			CheckpointCb:        intfs.Checkpoint,
			RecoverCb:           intfs.Recover,
			ShutdownCb:          intfs.Shutdown,
			SignalCb:            intfs.Signal,
			HealthCheckCb:       intfs.HealthCheck,
		},
		FunctionExectionPool:            pool.NewPool(pool.DefaultFuncExecPoolSize),
		SystemAuthAccessKey:             conf.SystemAuthAccessKey,
		SystemAuthSecretKey:             conf.SystemAuthSecretKey,
		EncryptPrivateKeyPasswd:         conf.EncryptPrivateKeyPasswd,
		PrimaryKeyStoreFile:             conf.PrimaryKeyStoreFile,
		StandbyKeyStoreFile:             conf.StandbyKeyStoreFile,
		EnableDsEncrypt:                 conf.EnableDsEncrypt,
		RuntimePublicKeyContextPath:     conf.RuntimePublicKeyContextPath,
		RuntimePrivateKeyContextPath:    conf.RuntimePrivateKeyContextPath,
		DsPublicKeyContextPath:          conf.DsPublicKeyContextPath,
		EnableEvent:                     conf.EnableEvent,
	}
	if err := libruntime.Init(runtimeConf); err != nil {
		fmt.Printf("failed to init libruntime, error %s\n", err.Error())
		return err
	}
	return nil
}
