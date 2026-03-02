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

// Package common for tools
package common

import (
	"flag"
	"os"
	"sync"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/config"
)

const (
	defaultConfigPath              = "/home/sn/config/runtime.json"
	defaultMaxConcurrencyCreateNum = 100
)

var (
	configSingleton = struct {
		sync.Once
		path string
		ac   *AuthContext
		cfg  *Configuration
	}{}
)

// Configuration to save config
type Configuration struct {
	RuntimeID                       string
	InstanceID                      string
	FunctionName                    string
	LogLevel                        string
	GrpcAddress                     string
	FSAddress                       string
	LogPath                         string
	JobID                           string
	DriverMode                      bool
	EnableMTLS                      bool
	PrivateKeyPath                  string
	CertificateFilePath             string
	VerifyFilePath                  string
	PrivateKeyPaaswd                string
	SystemAuthAccessKey             string
	SystemAuthSecretKey             string
	EncryptPrivateKeyPasswd         string
	PrimaryKeyStoreFile             string
	StandbyKeyStoreFile             string
	EnableDsEncrypt                 bool
	RuntimePublicKeyContextPath     string
	RuntimePrivateKeyContextPath    string
	DsPublicKeyContextPath          string
	MaxConcurrencyCreateNum         int
	EnableSigaction                 bool
	EnableEvent                     bool
}

func initConfig() {
	configSingleton.Do(
		func() {
			configSingleton.cfg = &Configuration{}
			flag.StringVar(&configSingleton.cfg.RuntimeID, "runtimeId", "", "")
			flag.StringVar(&configSingleton.cfg.InstanceID, "instanceId", "", "")
			flag.StringVar(&configSingleton.cfg.FunctionName, "functionName", "", "")
			flag.StringVar(&configSingleton.cfg.LogLevel, "logLevel", "", "")
			flag.StringVar(&configSingleton.cfg.GrpcAddress, "grpcAddress", "", "")
			flag.StringVar(&configSingleton.cfg.FSAddress, "functionSystemAddress", "", "")
			flag.StringVar(&configSingleton.path, "runtimeConfigPath", defaultConfigPath, "")
			flag.StringVar(&configSingleton.cfg.LogPath, "logPath", "", "")
			flag.StringVar(&configSingleton.cfg.JobID, "jobId", "12345678", "")
			flag.BoolVar(&configSingleton.cfg.DriverMode, "driverMode", false, "")
			flag.BoolVar(&configSingleton.cfg.EnableMTLS, "enableMTLS", false, "")
			flag.StringVar(&configSingleton.cfg.PrivateKeyPath, "privateKeyPath", "", "")
			flag.StringVar(&configSingleton.cfg.CertificateFilePath, "certificateFilePath", "", "")
			flag.StringVar(&configSingleton.cfg.VerifyFilePath, "verifyFilePath", "", "")
			flag.StringVar(&configSingleton.cfg.EncryptPrivateKeyPasswd, "encryptPrivateKeyPasswd", "", "")
			flag.StringVar(&configSingleton.cfg.PrimaryKeyStoreFile, "primaryKeyStoreFile", "", "")
			flag.StringVar(&configSingleton.cfg.StandbyKeyStoreFile, "standbyKeyStoreFile", "", "")
			flag.BoolVar(&configSingleton.cfg.EnableDsEncrypt, "enableDsEncrypt", false, "")
			flag.IntVar(&configSingleton.cfg.MaxConcurrencyCreateNum, "maxConcurrencyCreateNum",
				defaultMaxConcurrencyCreateNum, "")
			flag.BoolVar(&configSingleton.cfg.EnableEvent, "enableEvent", false, "")
			flag.Parse()
			setConfigSingletonCfg(&configSingleton.cfg.RuntimeID, "YR_RUNTIME_ID")
			setConfigSingletonCfg(&configSingleton.cfg.InstanceID, "INSTANCE_ID")
			setConfigSingletonCfg(&configSingleton.cfg.FunctionName, "FUNCTION_NAME")
			setConfigSingletonCfg(&configSingleton.cfg.LogLevel, "YR_LOG_LEVEL")
			setConfigSingletonCfg(&configSingleton.cfg.GrpcAddress, "POSIX_LISTEN_ADDR")
			setConfigSingletonCfg(&configSingleton.cfg.LogPath, "GLOG_log_dir")
			setConfigSingletonCfg(&configSingleton.cfg.JobID, "YR_JOB_ID")
			configSingleton.cfg.EnableSigaction = true
		},
	)
}

func setConfigSingletonCfg(fieldValue *string, envValue string) {
	if *fieldValue == "" {
		*fieldValue = os.Getenv(envValue)
	}
}

// GetConfig to parse args
func GetConfig() *Configuration {
	initConfig()
	config.LogLevelFromFlag = configSingleton.cfg.LogLevel
	return configSingleton.cfg
}
