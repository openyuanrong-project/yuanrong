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

// Package etcd3 client
package etcd3

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

var (
	routerEtcdClient     *EtcdClient
	metaEtcdClient       *EtcdClient
	caeMetaEtcdClient    *EtcdClient
	dataSystemEtcdClient *EtcdClient
)

const (
	// Router router etcd type
	Router = "route"

	// Meta meta etcd type
	Meta = "meta"

	// CAEMeta cae meta etcd type
	CAEMeta = "CAEMeta"

	// DataSystem cae meta etcd type
	DataSystem = "DataSystem"

	defaultEtcdLostContactTime = 5 * time.Minute
)

var (
	errInitRouterEtcd      = errors.New("failed to init router etcd client")
	errInitMetadataEtcd    = errors.New("failed to init metadata etcd client")
	errInitCAEMetadataEtcd = errors.New("failed to init CAE metadata etcd client")
	errInitDataSystemEtcd  = errors.New("failed to init dataSystem etcd client")
)

// GetRouterEtcdClient -
func GetRouterEtcdClient() *EtcdClient {
	return routerEtcdClient
}

// GetMetaEtcdClient -
func GetMetaEtcdClient() *EtcdClient {
	return metaEtcdClient
}

// GetCAEMetaEtcdClient -
func GetCAEMetaEtcdClient() *EtcdClient {
	return caeMetaEtcdClient
}

// GetDataSystemEtcdClient -
func GetDataSystemEtcdClient() *EtcdClient {
	return dataSystemEtcdClient
}

// GetEtcdStatusLostContact -
func (e *EtcdClient) GetEtcdStatusLostContact() bool {
	return e.etcdStatusAfterLostContact
}

// GetEtcdStatusNow -
func (e *EtcdClient) GetEtcdStatusNow() bool {
	return e.etcdStatusNow
}

// GetEtcdType -
func (e *EtcdClient) GetEtcdType() string {
	return e.etcdType
}

// InitParam -
func InitParam() *EtcdInitParam {
	return new(EtcdInitParam)
}

// WithRouteEtcdConfig -
func (e *EtcdInitParam) WithRouteEtcdConfig(config EtcdConfig) *EtcdInitParam {
	e.routeEtcdConfig = &config
	return e
}

// WithMetaEtcdConfig -
func (e *EtcdInitParam) WithMetaEtcdConfig(config EtcdConfig) *EtcdInitParam {
	e.metaEtcdConfig = &config
	return e
}

// WithCAEMetaEtcdConfig -
func (e *EtcdInitParam) WithCAEMetaEtcdConfig(config EtcdConfig) *EtcdInitParam {
	e.CAEMetaEtcdConfig = &config
	return e
}

// WithDataSystemEtcdConfig -
func (e *EtcdInitParam) WithDataSystemEtcdConfig(config EtcdConfig) *EtcdInitParam {
	e.DataSystemEtcdConfig = &config
	return e
}

// WithStopCh -
func (e *EtcdInitParam) WithStopCh(ch <-chan struct{}) *EtcdInitParam {
	e.stopCh = ch
	return e
}

// WithAlarmSwitch -
func (e *EtcdInitParam) WithAlarmSwitch(enableAlarm bool) *EtcdInitParam {
	e.enableAlarm = enableAlarm
	return e
}

// InitRouterEtcdClient -
func InitRouterEtcdClient(etcdConfig EtcdConfig, alarmConfig alarm.Config, stopCh <-chan struct{}) error {
	if err := InitParam().
		WithRouteEtcdConfig(etcdConfig).
		WithStopCh(stopCh).
		WithAlarmSwitch(alarmConfig.EnableAlarm).
		InitClient(); err != nil {
		return err
	}
	if routerClient := GetRouterEtcdClient(); routerClient != nil {
		if err := routerClient.EtcdHeatBeat(); err != nil {
			errInfo := fmt.Sprintf("failed to check etcd client conn, err: %s", err.Error())
			log.GetLogger().Errorf(errInfo)
			routerClient.reportOrClearAlarm(alarm.GenerateAlarmLog, errInfo, alarm.Level2)
			time.Sleep(DurationContextTimeout)
			return err
		}
	}
	return nil
}

// InitMetaEtcdClient -
func InitMetaEtcdClient(etcdConfig EtcdConfig, alarmConfig alarm.Config, stopCh <-chan struct{}) error {
	if err := InitParam().
		WithMetaEtcdConfig(etcdConfig).
		WithAlarmSwitch(alarmConfig.EnableAlarm).
		WithStopCh(stopCh).
		InitClient(); err != nil {
		return err
	}
	if metaClient := GetMetaEtcdClient(); metaClient != nil {
		if err := metaClient.EtcdHeatBeat(); err != nil {
			errInfo := fmt.Sprintf("failed to check etcd client conn, err: %s", err.Error())
			log.GetLogger().Errorf(errInfo)
			metaClient.reportOrClearAlarm(alarm.GenerateAlarmLog, errInfo, alarm.Level2)
			time.Sleep(DurationContextTimeout)
			return err
		}
	}
	return nil
}

// InitCAEMetaEtcdClient -
func InitCAEMetaEtcdClient(etcdConfig EtcdConfig, alarmConfig alarm.Config, stopCh <-chan struct{}) error {
	if err := InitParam().
		WithCAEMetaEtcdConfig(etcdConfig).
		WithAlarmSwitch(alarmConfig.EnableAlarm).
		WithStopCh(stopCh).
		InitClient(); err != nil {
		return err
	}
	if metaClient := GetCAEMetaEtcdClient(); metaClient != nil {
		if err := metaClient.EtcdHeatBeat(); err != nil {
			errInfo := fmt.Sprintf("failed to check etcd client conn, err: %s", err.Error())
			log.GetLogger().Errorf(errInfo)
			metaClient.reportOrClearAlarm(alarm.GenerateAlarmLog, errInfo, alarm.Level2)
			time.Sleep(DurationContextTimeout)
			return err
		}
	}
	return nil
}

// InitDataSystemEtcdClient -
func InitDataSystemEtcdClient(etcdConfig EtcdConfig, alarmConfig alarm.Config, stopCh <-chan struct{}) error {
	if err := InitParam().
		WithDataSystemEtcdConfig(etcdConfig).
		WithAlarmSwitch(alarmConfig.EnableAlarm).
		WithStopCh(stopCh).
		InitClient(); err != nil {
		return err
	}
	if etcdClient := GetDataSystemEtcdClient(); etcdClient != nil {
		if err := etcdClient.EtcdHeatBeat(); err != nil {
			errInfo := fmt.Sprintf("failed to check etcd client conn, err: %s", err.Error())
			log.GetLogger().Errorf(errInfo)
			etcdClient.reportOrClearAlarm(alarm.GenerateAlarmLog, errInfo, alarm.Level2)
			time.Sleep(DurationContextTimeout)
			return err
		}
	}
	return nil
}

// InitClient initialize etcdClient based on initialization parameters.
func (e *EtcdInitParam) InitClient() error {
	if e.routeEtcdConfig != nil && e.initRouteEtcdClient() != nil {
		return errInitRouterEtcd
	}
	if e.metaEtcdConfig != nil && e.initMetadataEtcdClient() != nil {
		return errInitMetadataEtcd
	}
	if e.CAEMetaEtcdConfig != nil && e.initCAEMetadataEtcdClient() != nil {
		return errInitCAEMetadataEtcd
	}
	if e.DataSystemEtcdConfig != nil && e.initDataSystemEtcdClient() != nil {
		return errInitDataSystemEtcd
	}
	return nil
}

func (e *EtcdInitParam) initRouteEtcdClient() error {
	if routerEtcdClient != nil {
		return nil
	}
	var err error
	if routerEtcdClient, err = newClient(e.routeEtcdConfig, e.stopCh, e.enableAlarm, Router); err != nil {
		log.GetLogger().Errorf("failed to new router etcd client with error: %s", err.Error())
		return err
	}
	return nil
}

func (e *EtcdInitParam) initMetadataEtcdClient() error {
	if metaEtcdClient != nil {
		return nil
	}
	var err error
	log.GetLogger().Infof("new meta etcd client")
	if metaEtcdClient, err = newClient(e.metaEtcdConfig, e.stopCh, e.enableAlarm, Meta); err != nil {
		log.GetLogger().Errorf("failed to new metadata etcd client with error: %s", err.Error())
		return err
	}
	return nil
}

func (e *EtcdInitParam) initCAEMetadataEtcdClient() error {
	if caeMetaEtcdClient != nil {
		return nil
	}
	var err error
	log.GetLogger().Infof("new CAE meta etcd client")
	if caeMetaEtcdClient, err = newClient(e.CAEMetaEtcdConfig, e.stopCh, e.enableAlarm, CAEMeta); err != nil {
		log.GetLogger().Errorf("failed to new CAE metadata etcd client with error: %s", err.Error())
		return err
	}
	return nil
}

func (e *EtcdInitParam) initDataSystemEtcdClient() error {
	if dataSystemEtcdClient != nil {
		return nil
	}
	var err error
	log.GetLogger().Infof("new DataSystem etcd client")
	if dataSystemEtcdClient, err = newClient(e.DataSystemEtcdConfig, e.stopCh, e.enableAlarm, DataSystem); err != nil {
		log.GetLogger().Errorf("failed to new DataSystem etcd client with error: %s", err.Error())
		return err
	}
	return nil
}

func newClient(config *EtcdConfig, stopCh <-chan struct{}, enableAlarm bool,
	etcdType string) (*EtcdClient, error) {
	if stopCh == nil {
		return nil, errors.New("etcd stopCh should not be nil")
	}
	client, err := buildClient(config)
	if err != nil {
		log.GetLogger().Errorf("failed to new %s etcd client, %s", etcdType, err.Error())
		return nil, err
	}
	client.stopCh = stopCh
	client.config = config
	client.etcdType = etcdType
	client.isAlarmEnable = enableAlarm
	client.etcdStatusAfterLostContact = true
	client.etcdStatusNow = true

	go client.keepConnAlive()
	return client, nil
}

func buildClient(config *EtcdConfig) (*EtcdClient, error) {
	cfg, err := GetEtcdAuthType(*config).GetEtcdConfig()
	if err != nil {
		log.GetLogger().Errorf("failed to create shared etcd client error %s", err.Error())
		return nil, err
	}
	cfg.DialTimeout = etcdDialTimeout
	cfg.DialKeepAliveTime = etcdKeepaliveTime
	cfg.DialKeepAliveTimeout = etcdKeepaliveTimeout
	cfg.Endpoints = config.Servers
	etcdClient, err := clientv3.New(*cfg)
	if err != nil {
		log.GetLogger().Errorf("failed to create shared etcd client error %s", err.Error())
		return nil, err
	}
	return &EtcdClient{
		Client:       etcdClient,
		clientExitCh: make(chan struct{}),
		cond:         sync.NewCond(&sync.Mutex{}),
	}, nil
}

func (e *EtcdClient) keepConnAlive() {
	timer := time.NewTimer(keepConnAliveTTL)
	for {
		select {
		case <-timer.C:
			e.checkConnState()
			timer.Reset(keepConnAliveTTL)
		case _, ok := <-e.stopCh:
			if !ok {
				log.GetLogger().Warnf("stop channel is closed and quits keep %s etcd conn alive task", e.etcdType)
			}
			e.cond.Broadcast()
			timer.Stop()
			return
		}
	}
}

// EtcdHeatBeat -
func (e *EtcdClient) EtcdHeatBeat() error {
	ctx, cancel := context.WithTimeout(context.Background(), keepConnAliveTTL)
	defer cancel()
	_, err := e.Client.Get(ctx, "alive", clientv3.WithKeysOnly())
	return err
}

func (e *EtcdClient) checkConnState() {
	e.rwMutex.RLock()
	err := e.EtcdHeatBeat()
	e.rwMutex.RUnlock()

	if err != nil {
		if e.etcdTimer == nil {
			e.abnormalContinuouslyTimes++
			e.etcdTimer = time.AfterFunc(defaultEtcdLostContactTime, func() {
				e.etcdStatusAfterLostContact = false
				errInfo := fmt.Sprintf("etcd %s lost contact over %v, etcdStatusAfterLostContact is %v",
					e.etcdType, defaultEtcdLostContactTime, e.etcdStatusAfterLostContact)
				e.reportOrClearAlarm(alarm.GenerateAlarmLog, errInfo, alarm.Level3)
				log.GetLogger().Warnf(errInfo)
			})
		}
		e.etcdStatusNow = false
		e.exitOnce.Do(func() {
			close(e.clientExitCh)
		})
		errInfo := fmt.Sprintf("failed to check etcd client conn, err: %s", err.Error())
		log.GetLogger().Errorf(errInfo)
		e.reportOrClearAlarm(alarm.GenerateAlarmLog, errInfo, alarm.Level2)
		if err = e.restart(); err != nil {
			log.GetLogger().Errorf("failed to restart etcd client, %s", err.Error())
		}
		return
	}
	if e.etcdStatusAfterLostContact == false {
		e.reportOrClearAlarm(alarm.ClearAlarmLog, "Clear critical alarm, "+
			"The connection to etcd has been restored", alarm.Level3)
	}
	if e.abnormalContinuouslyTimes > 0 {
		e.reportOrClearAlarm(alarm.ClearAlarmLog, "Clear major alarm, "+
			"The connection to etcd has been restored", alarm.Level2)
		e.abnormalContinuouslyTimes = 0
	}

	if e.etcdTimer != nil {
		e.etcdTimer.Stop()
		e.etcdTimer = nil
		e.etcdStatusAfterLostContact = true
		log.GetLogger().Infof("reconnect to %s etcd", e.etcdType)
	}
	if !e.etcdStatusNow {
		e.clientExitCh = make(chan struct{})
		e.exitOnce = sync.Once{}
		e.cond.Broadcast()
	}
	e.etcdStatusNow = true
}

func (e *EtcdClient) reportOrClearAlarm(opType string, detail string, alarmLevel string) {
	if e.isAlarmEnable {
		alarmDetail := &alarm.Detail{
			SourceTag: os.Getenv(constant.PodNameEnvKey) + "|" + os.Getenv(constant.PodIPEnvKey) +
				"|" + os.Getenv(constant.ClusterName) + "|MetadataEtcdConnection",
			OpType:         opType,
			Details:        detail,
			StartTimestamp: 0,
			EndTimestamp:   0,
		}
		if alarmDetail.OpType == alarm.GenerateAlarmLog {
			alarmDetail.StartTimestamp = int(time.Now().Unix())
		} else {
			alarmDetail.EndTimestamp = int(time.Now().Unix())
		}
		alarmInfo := &alarm.LogAlarmInfo{
			AlarmID:    alarm.MetadataEtcdConnection00001,
			AlarmName:  "MetadataEtcdConnection",
			AlarmLevel: alarmLevel,
		}
		if e.etcdType == Router {
			alarmDetail.SourceTag = os.Getenv(constant.PodNameEnvKey) + "|" + os.Getenv(constant.PodIPEnvKey) +
				"|" + os.Getenv(constant.ClusterName) + "|RouterEtcdConnection"
			alarmInfo.AlarmID = alarm.RouterEtcdConnection00001
			alarmInfo.AlarmName = "RouterEtcdConnection"
		}
		if e.etcdType == CAEMeta {
			alarmDetail.SourceTag = os.Getenv(constant.PodNameEnvKey) + "|" + os.Getenv(constant.PodIPEnvKey) +
				"|" + os.Getenv(constant.ClusterName) + "|CAEMetadataEtcdConnection"
			alarmInfo.AlarmID = alarm.RouterEtcdConnection00001
			alarmInfo.AlarmName = "CAEMetadataEtcdConnection"
		}
		alarm.ReportOrClearAlarm(alarmInfo, alarmDetail)
	}
}

func (e *EtcdClient) restart() error {
	log.GetLogger().Infof("start to rebuild %s etcd client", e.etcdType)
	recreatedClient, err := buildClient(e.config)
	if err != nil {
		log.GetLogger().Errorf("failed to recreate %s etcd client, %s", e.etcdType, err.Error())
		return err
	}
	e.rwMutex.Lock()
	e.stop()
	e.Client = recreatedClient.Client
	e.rwMutex.Unlock()
	return nil
}

func (e *EtcdClient) stop() {
	if err := e.Client.Close(); err != nil {
		log.GetLogger().Errorf("failed to close %s etcd client, %s", e.etcdType, err.Error())
	}
}

// AttachAZPrefix -
func (e *EtcdClient) AttachAZPrefix(key string) string {
	if e.config != nil && len(e.config.AZPrefix) != 0 {
		return fmt.Sprintf("/%s%s", e.config.AZPrefix, key)
	}
	return key
}

// DetachAZPrefix -
func (e *EtcdClient) DetachAZPrefix(key string) string {
	if e.config != nil && len(e.config.AZPrefix) != 0 {
		return strings.TrimPrefix(key, fmt.Sprintf("/%s", e.config.AZPrefix))
	}
	return key
}
