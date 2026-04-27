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

// Package registry -
package registry

import (
	"context"
	"fmt"
	"strings"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	commonUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

// FunctionRegistry watches function event of etcd
type FunctionRegistry struct {
	userAgencyRegistry *UserAgencyRegistry
	watcher            etcd3.Watcher
	funcSpecs          map[string]*types.FunctionSpecification
	subscriberChans    []chan SubEvent
	listDoneCh         chan struct{}
	stopCh             <-chan struct{}
	sync.RWMutex
}

// NewFunctionRegistry will create FunctionRegistry
func NewFunctionRegistry(stopCh <-chan struct{}) *FunctionRegistry {
	functionRegistry := &FunctionRegistry{
		userAgencyRegistry: NewUserAgencyRegistry(stopCh),
		funcSpecs:          make(map[string]*types.FunctionSpecification, utils.DefaultMapSize),
		listDoneCh:         make(chan struct{}, 1),
		stopCh:             stopCh,
	}
	return functionRegistry
}

func (fr *FunctionRegistry) initWatcher(etcdClient *etcd3.EtcdClient) {
	fr.watcher = etcd3.NewEtcdWatcher(
		functionEtcdPrefix,
		fr.watcherFilter,
		fr.watcherHandler,
		fr.stopCh,
		etcdClient)
	fr.watcher.StartList()
	fr.WaitForETCDList()
	fr.userAgencyRegistry.initWatcher(etcdClient)
}

// WaitForETCDList -
func (fr *FunctionRegistry) WaitForETCDList() {
	log.GetLogger().Infof("start to wait function ETCD list")
	select {
	case <-fr.listDoneCh:
		log.GetLogger().Infof("receive list done, stop waiting ETCD list")
	case <-fr.stopCh:
		log.GetLogger().Warnf("registry is stopped, stop waiting ETCD list")
	}
}

// RunWatcher will start etcd watch process
func (fr *FunctionRegistry) RunWatcher() {
	go fr.watcher.StartWatch()
	fr.userAgencyRegistry.RunWatcher()
}

func (fr *FunctionRegistry) getFuncSpec(funcKey string) *types.FunctionSpecification {
	fr.RLock()
	funcSpec := fr.funcSpecs[funcKey]
	fr.RUnlock()
	return funcSpec
}

func (fr *FunctionRegistry) fetchSilentFuncSpec(funcKey string) *types.FunctionSpecification {
	tenantID, funcName, funcVersion := commonUtils.ParseFuncKey(funcKey)
	silentEtcdKey := fmt.Sprintf(constant.SilentFuncKey, tenantID, funcName, funcVersion)
	etcdValue, err := etcd3.GetValueFromEtcdWithRetry(silentEtcdKey, etcd3.GetMetaEtcdClient())
	if err != nil {
		log.GetLogger().Errorf("failed to get silent function, error: %s", err.Error())
		return nil
	}
	metaEtcdKey := fmt.Sprintf(constant.MetaFuncKey, tenantID, funcName, funcVersion)
	fr.Lock()
	defer fr.Unlock()
	funcSpec := fr.buildFuncSpec(metaEtcdKey, etcdValue, funcKey)
	if funcSpec == nil {
		return nil
	}
	fr.funcSpecs[funcKey] = funcSpec
	log.GetLogger().Infof("get silent function success,funcKey :%s", funcKey)
	return funcSpec
}

func (fr *FunctionRegistry) watcherFilter(event *etcd3.Event) bool {
	items := strings.Split(event.Key, keySeparator)
	if len(items) != validEtcdKeyLenForFunction {
		return true
	}
	if items[instanceKeyIndex] != "functions" || items[tenantKeyIndex] != "tenant" ||
		items[functionKeyIndex] != "function" {
		return true
	}

	return false
}

func (fr *FunctionRegistry) watcherHandler(event *etcd3.Event) {
	fr.Lock()
	defer fr.Unlock()
	log.GetLogger().Infof("handling function event type %s key %s", event.Type, event.Key)
	if event.Type == etcd3.SYNCED {
		log.GetLogger().Infof("publish function synced event")
		fr.publishEvent(SubEventTypeSynced, &types.FunctionSpecification{})
		return
	}
	funcKey := GetFuncKeyFromEtcdKey(event.Key)
	if len(funcKey) == 0 {
		log.GetLogger().Warnf("ignoring invalid etcd key of key %s", event.Key)
		return
	}
	switch event.Type {
	case etcd3.PUT:
		funcSpec := fr.buildFuncSpec(event.Key, event.Value, funcKey)
		if funcSpec == nil {
			return
		}
		fr.funcSpecs[funcKey] = funcSpec
		log.GetLogger().Infof("get function key :%s", funcKey)
		fr.publishEvent(SubEventTypeUpdate, funcSpec)
	case etcd3.DELETE:
		funcSpec, exist := fr.funcSpecs[funcKey]
		if !exist {
			log.GetLogger().Errorf("function %s doesn't exist in registry", funcKey)
			return
		}
		funcSpec.CancelFunc()
		delete(fr.funcSpecs, funcKey)
		fr.publishEvent(SubEventTypeDelete, funcSpec)
	default:
		log.GetLogger().Warnf("unsupported event: %v", event.Type)
	}
}

// buildFuncSpec without lock should lock outside
func (fr *FunctionRegistry) buildFuncSpec(etcdKey string, etcdValue []byte,
	funcKey string) *types.FunctionSpecification {
	funcMetaInfo := GetFuncMetaInfoFromEtcdValue(etcdValue)
	if funcMetaInfo == nil {
		log.GetLogger().Errorf("ignoring invalid etcd value of key %s", etcdKey)
		return nil
	}
	funcMetaInfo.ExtendedMetaData.UserAgency = fr.userAgencyRegistry.GetUserAgencyByFuncMeta(funcMetaInfo)
	funcSpec := createOrUpdateFuncSpec(fr.funcSpecs[funcKey], funcKey, funcMetaInfo)
	funcSpec.FuncSecretName = utils.GenerateStsSecretName(etcdKey)
	return funcSpec
}

func createOrUpdateFuncSpec(oldFuncSpec *types.FunctionSpecification, funcKey string,
	funcMetaInfo *commonTypes.FunctionMetaInfo) *types.FunctionSpecification {
	commonUtils.SetFuncMetaDynamicConfEnable(funcMetaInfo)
	var funcSpec *types.FunctionSpecification
	if oldFuncSpec == nil {
		funcCtx, cancelFunc := context.WithCancel(context.TODO())
		funcSpec = &types.FunctionSpecification{
			FuncCtx:    funcCtx,
			CancelFunc: cancelFunc,
			FuncKey:    funcKey,
			FuncMetaSignature: commonUtils.GetFuncMetaSignature(funcMetaInfo,
				config.GlobalConfig.RawStsConfig.StsEnable),
			FuncMetaData:     funcMetaInfo.FuncMetaData,
			S3MetaData:       funcMetaInfo.S3MetaData,
			CodeMetaData:     funcMetaInfo.CodeMetaData,
			EnvMetaData:      funcMetaInfo.EnvMetaData,
			StsMetaData:      funcMetaInfo.StsMetaData,
			ResourceMetaData: funcMetaInfo.ResourceMetaData,
			InstanceMetaData: funcMetaInfo.InstanceMetaData,
			ExtendedMetaData: funcMetaInfo.ExtendedMetaData,
			RootfsSpecMeta:   funcMetaInfo.RootfsSpecMeta,
		}
	} else {
		funcSpec = oldFuncSpec
		funcSpec.FuncMetaSignature = commonUtils.GetFuncMetaSignature(funcMetaInfo,
			config.GlobalConfig.RawStsConfig.StsEnable)
		funcSpec.FuncMetaData = funcMetaInfo.FuncMetaData
		funcSpec.S3MetaData = funcMetaInfo.S3MetaData
		funcSpec.CodeMetaData = funcMetaInfo.CodeMetaData
		funcSpec.EnvMetaData = funcMetaInfo.EnvMetaData
		funcSpec.StsMetaData = funcMetaInfo.StsMetaData
		funcSpec.ResourceMetaData = funcMetaInfo.ResourceMetaData
		funcSpec.InstanceMetaData = funcMetaInfo.InstanceMetaData
		funcSpec.ExtendedMetaData = funcMetaInfo.ExtendedMetaData
		funcSpec.RootfsSpecMeta = funcMetaInfo.RootfsSpecMeta
	}
	return funcSpec
}

func (fr *FunctionRegistry) addSubscriberChan(subChan chan SubEvent) {
	fr.Lock()
	fr.subscriberChans = append(fr.subscriberChans, subChan)
	fr.Unlock()
}

func (fr *FunctionRegistry) publishEvent(eventType EventType, funcSpec *types.FunctionSpecification) {
	for _, subChan := range fr.subscriberChans {
		if subChan != nil {
			subChan <- SubEvent{
				EventType: eventType,
				EventMsg:  funcSpec,
			}
		}
	}
}

// FinishEtcdList -
func (fr *FunctionRegistry) FinishEtcdList() {
	log.GetLogger().Infof("received function synced event")
	fr.listDoneCh <- struct{}{}
	return
}

// EtcdList -
func (fr *FunctionRegistry) EtcdList() []*types.FunctionSpecification {
	client := etcd3.GetMetaEtcdClient()
	if client == nil {
		return nil
	}
	if client.Client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), etcd3.DurationContextTimeout)
	defer cancel()
	res, err := client.Client.Get(ctx, functionEtcdPrefix, clientv3.WithPrefix())
	if err != nil {
		log.GetLogger().Errorf("get function meta failed, error: %v", err)
		return nil
	}
	var result []*types.FunctionSpecification
	for _, kv := range res.Kvs {
		e := &etcd3.Event{
			Key:   string(kv.Key),
			Value: kv.Value,
		}
		if fr.watcherFilter(e) {
			continue
		}
		funcKey := GetFuncKeyFromEtcdKey(e.Key)
		if len(funcKey) == 0 {
			log.GetLogger().Warnf("ignoring invalid etcd key of key %s", e.Key)
			continue
		}
		funcSpec := fr.buildFuncSpec(e.Key, e.Value, funcKey)
		if funcSpec == nil {
			continue
		}
		fr.funcSpecs[funcKey] = funcSpec
		result = append(result, funcSpec)
	}
	return result
}
