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

package datasystemclient

import (
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

const (
	// 节点已准备好对外服务
	dataSystemStatusReady = "ready"
	// 节点启动
	dataSystemStatusStart = "start"
	// 节点重启
	dataSystemStatusRestart = "restart"
	// 节点对账恢复中
	dataSystemStatusRecover = "recover"
	// etcd故障期间重启的状态
	dataSystemStatusDRst = "d_rst"
	// 节点退出（主动缩容）
	dataSystemStatusExiting = "exiting"
)

const readinessDuration = 15 * time.Second

var (
	localDataSystemStatusCache = &LocalDataSystemStatusCache{}
	shutdownFlag               = atomic.Bool{}
	streamEnable               = atomic.Bool{}
)

// LocalDataSystemStatusCache 本地数据系统状态缓存结构体
type LocalDataSystemStatusCache struct {
	status string
	lock   sync.RWMutex
}

// IsStatusReady -
func (d *LocalDataSystemStatusCache) IsStatusReady() bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if d.status != dataSystemStatusReady {
		log.GetLogger().Debugf("data system status is not ready, status: %s", d.status)
		return false
	}
	return true
}

// SetLocalDataSystemStatus -
func (d *LocalDataSystemStatusCache) SetLocalDataSystemStatus(ip, status string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	localNode := os.Getenv("NODE_IP")
	if localNode == "" {
		log.GetLogger().Debugf("get local node is empty")
		return
	}
	if ip != localNode {
		log.GetLogger().Debugf("node[%s] is not local data system node[%s]", ip, localNode)
		return
	}
	log.GetLogger().Infof("save local data system node[%s] status[%s]", ip, status)
	d.status = status
}

// GetLocalDataSystemStatus -
func (d *LocalDataSystemStatusCache) GetLocalDataSystemStatus() string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.status
}

// IsLocalDataSystemStatusReady -
func IsLocalDataSystemStatusReady() bool {
	return localDataSystemStatusCache.IsStatusReady()
}

// SetStreamEnable -
func SetStreamEnable(streamEnableConfig bool) {
	streamEnable.Store(streamEnableConfig)
}

func isShutdownFronted() bool {
	if !streamEnable.Load() {
		log.GetLogger().Infof("it's not stream scenario, skip shutdown frontend")
		return false
	}
	skipShutdownStatusMap := map[string]struct{}{
		dataSystemStatusReady:   {},
		dataSystemStatusStart:   {},
		dataSystemStatusRestart: {},
		dataSystemStatusRecover: {},
		dataSystemStatusDRst:    {},
	}
	status := localDataSystemStatusCache.GetLocalDataSystemStatus()
	if _, ok := skipShutdownStatusMap[status]; ok {
		log.GetLogger().Debugf("status is [%s], skip shutdown frontend", status)
		return false
	}
	return true
}

func destroy() {
	time.Sleep(readinessDuration)
	if shutdownFlag.Swap(true) {
		log.GetLogger().Infof("shutdown frontend has been triggered, skip this operation")
		return
	}
	defer func() {
		shutdownFlag.Store(false)
	}()
	log.GetLogger().Infof("local dataSystem status is not ready, prepare shutdown frontend")
	pid := os.Getpid()
	proc, err := os.FindProcess(pid)
	if err != nil {
		log.GetLogger().Errorf("get process pid failed, pid: %d, err: %v", pid, err)
		return
	}
	err = proc.Signal(syscall.SIGTERM)
	if err != nil {
		log.GetLogger().Errorf("send SIGTERM signal to the process failed, pid: %d, err: %v", pid, err)
		return
	}
	log.GetLogger().Infof("send SIGTERM signal to the process success, pid: %d", pid)
}
