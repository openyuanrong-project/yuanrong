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

// Package process to run collector
package process

import (
	"yuanrong.org/kernel/pkg/collector/common"
	"yuanrong.org/kernel/pkg/collector/logcollector"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// StartCollector -
func StartCollector() {
	common.InitCmd()
	if err := common.InitEtcdClient(); err != nil {
		log.GetLogger().Errorf("failed to init etcd client %s", err.Error())
		return
	}
	ready := make(chan bool)
	go func() {
		<-ready
		err := logcollector.Register()
		if err != nil {
			log.GetLogger().Errorf("failed to register %s", err.Error())
			return
		}
		logcollector.StartLogReporter()
	}()
	err := logcollector.StartReadLogService(ready)
	if err != nil {
		log.GetLogger().Errorf("failed to start log service %s", err.Error())
		return
	}
}
