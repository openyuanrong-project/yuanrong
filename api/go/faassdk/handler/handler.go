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

// Package handler -
package handler

import (
	"fmt"
	"os"

	"yuanrong.org/kernel/runtime/faassdk/config"
	"yuanrong.org/kernel/runtime/libruntime/api"
	log "yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

// ExecutorHandler -
type ExecutorHandler interface {
	InitHandler(args []api.Arg, rt api.LibruntimeAPI) ([]byte, error)
	CallHandler(args []api.Arg, context map[string]string) ([]byte, error)
	CheckPointHandler(checkPointId string) ([]byte, error)
	RecoverHandler(state []byte) error
	ShutDownHandler(gracePeriodSecond uint64) error
	SignalHandler(signal int32, payload []byte) error
	HealthCheckHandler() (api.HealthType, error)
}

// GetUserCodePath will get user code downloaded in delegate download path
func GetUserCodePath() (string, error) {
	userCodePath := os.Getenv(config.DelegateDownloadPath)
	if userCodePath == "" {
		log.GetLogger().Errorf("%s not found", config.DelegateDownloadPath)
		return "", fmt.Errorf("%s not found", config.DelegateDownloadPath)
	}
	return userCodePath, nil
}
