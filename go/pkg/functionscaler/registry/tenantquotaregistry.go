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
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
	"yuanrong.org/kernel/pkg/functionscaler/tenantquota"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

// TenantQuotaRegistry watches tenant instance event of etcd
type TenantQuotaRegistry struct {
	watcher etcd3.Watcher
	stopCh  <-chan struct{}
	sync.RWMutex
}

// NewTenantQuotaRegistry will create TenantQuotaRegistry
func NewTenantQuotaRegistry(stopCh <-chan struct{}) *TenantQuotaRegistry {
	tenantQuotaRegistry := &TenantQuotaRegistry{
		stopCh: stopCh,
	}
	return tenantQuotaRegistry
}

func (tr *TenantQuotaRegistry) initWatcher(etcdClient *etcd3.EtcdClient) {
	tr.watcher = etcd3.NewEtcdWatcher(
		constant.TenantQuotaPrefix,
		tr.watcherFilter,
		tr.watcherHandler,
		tr.stopCh,
		etcdClient)
	tr.watcher.StartList()
}

// RunWatcher will start etcd watch process for tenant instance event
func (tr *TenantQuotaRegistry) RunWatcher() {
	go tr.watcher.StartWatch()
}

// watcherFilter will filter tenant instance event from etcd event
func (tr *TenantQuotaRegistry) watcherFilter(event *etcd3.Event) bool {
	if !isTenantQuota(event.Key) && !isDefaultTenantQuota(event.Key) {
		return true
	}
	return false
}

func isTenantQuota(etcdRef string) bool {
	// An example of a tenant quota key:
	// /sn/quota/cluster/<clusterID>/tenant/<tenantID>/instancemetadata
	strs := strings.Split(etcdRef, keySeparator)
	if len(strs) != validEtcdKeyLenForQuota2 {
		return false
	}
	if strs[quotaKeyIndex] != "quota" || strs[instanceMetadataKeyIndex2] != "instancemetadata" {
		return false
	}
	return true
}

func isDefaultTenantQuota(etcdRef string) bool {
	// An example of default tenant quota key:
	// /sn/quota/cluster/<clusterID>/default/instancemetadata
	strs := strings.Split(etcdRef, keySeparator)
	if len(strs) != validEtcdKeyLenForQuota1 {
		return false
	}
	if strs[quotaKeyIndex] != "quota" || strs[instanceMetadataKeyIndex1] != "instancemetadata" {
		return false
	}
	return true
}

// watcherHandler will handle tenant instance event from etcd
func (tr *TenantQuotaRegistry) watcherHandler(event *etcd3.Event) {
	log.GetLogger().Infof("handling TenantQuota event type %d key %s", event.Type, event.Key)
	var err error
	elements := strings.Split(event.Key, keySeparator)
	if isDefaultTenantQuota(event.Key) {
		log.GetLogger().Infof("default tenant quota key changed: %s", event.Key)
		err = handleDefaultQuotaEvent(event)
	} else if len(elements) == validEtcdKeyLenForQuota2 {
		log.GetLogger().Infof("different tenant quota key changed: %s", event.Key)
		tenantID := elements[tenantValueIndex]
		err = handleTenantQuotaEvent(event, tenantID)
	} else {
		log.GetLogger().Errorf("invalid tenant quota key: %s", event.Key)
		return
	}
	if err != nil {
		log.GetLogger().Errorf("failed to process event type: %s and key: %s, error %s",
			event.Type, event.Key, err.Error())
		return
	}
	return
}

func handleDefaultQuotaEvent(event *etcd3.Event) error {
	var (
		tenantMetaInfo = types.TenantMetaInfo{}
		err            error
	)

	// check value
	if event.Type == etcd3.DELETE {
		err = json.Unmarshal(event.PrevValue, &tenantMetaInfo)
	} else {
		err = json.Unmarshal(event.Value, &tenantMetaInfo)
	}
	if err != nil {
		log.GetLogger().Errorf("unmarshal default tenant quota info failed, key: %s and error: %s",
			event.Key, err.Error())
		return err
	}
	log.GetLogger().Infof("default tenant quota info: %+v", tenantMetaInfo)

	switch event.Type {
	case etcd3.PUT:
		log.GetLogger().Infof("default tenant quota update event type %s", event.Type)
		tenantquota.GetTenantCache().UpdateDefaultQuota(tenantMetaInfo)
	case etcd3.DELETE:
		log.GetLogger().Infof("default tenant quota delete event type %s, ignore", event.Type)
	default:
		log.GetLogger().Errorf("default tenant quota unsupported event type %s", event.Type)
		return errors.New("unsupported event type for tenant quota")
	}
	log.GetLogger().Infof("finished to process default tenant quota event, resource key: %s, type %d",
		event.Key, event.Type)
	return nil
}

func handleTenantQuotaEvent(event *etcd3.Event, tenantID string) error {
	var (
		tenantMetaInfo = types.TenantMetaInfo{}
		err            error
	)

	// check value
	if event.Type == etcd3.DELETE {
		err = json.Unmarshal(event.PrevValue, &tenantMetaInfo)
	} else {
		err = json.Unmarshal(event.Value, &tenantMetaInfo)
	}
	if err != nil {
		log.GetLogger().Errorf("unmarshal tenant quota info failed, key: %s and error: %s",
			urnutils.AnonymizeTenantMetadataEtcdKey(event.Key), err.Error())
		return err
	}

	switch event.Type {
	case etcd3.PUT:
		log.GetLogger().Infof("tenant quota update event type %s", event.Type)
		tenantquota.GetTenantCache().UpdateOrAddTenantQuota(tenantID, tenantMetaInfo)
	case etcd3.DELETE:
		log.GetLogger().Infof("tenant quota delete event type %s", event.Type)
		tenantquota.GetTenantCache().DeleteTenantQuota(tenantID)
	default:
		log.GetLogger().Errorf("tenant quota unsupported event type %s", event.Type)
		return errors.New("unsupported event type for tenant quota")
	}
	log.GetLogger().Infof("finished to process tenant quota event, resource key: %s, type %d",
		event.Key, event.Type)
	return nil
}
