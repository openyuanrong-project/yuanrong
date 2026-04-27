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
	"fmt"
	"strings"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

// UserAgencyRegistry watches user agency event of etcd
type UserAgencyRegistry struct {
	userAgencyCacheMap sync.Map
	watcher            etcd3.Watcher
	stopCh             <-chan struct{}
	sync.RWMutex
}

// NewUserAgencyRegistry will create UserAgencyRegistry
func NewUserAgencyRegistry(stopCh <-chan struct{}) *UserAgencyRegistry {
	userAgencyRegistry := &UserAgencyRegistry{
		stopCh: stopCh,
	}
	return userAgencyRegistry
}

func (ur *UserAgencyRegistry) initWatcher(etcdClient *etcd3.EtcdClient) {
	ur.watcher = etcd3.NewEtcdWatcher(
		userAgencyEtcdPrefix,
		ur.watcherFilter,
		ur.watcherHandler,
		ur.stopCh,
		etcdClient)
	ur.watcher.StartList()
}

// RunWatcher will start etcd watch process for instance event
func (ur *UserAgencyRegistry) RunWatcher() {
	go ur.watcher.StartWatch()
}

// GetUserAgencyByFuncMeta get user agency by function meta
func (ur *UserAgencyRegistry) GetUserAgencyByFuncMeta(
	funcMetaInfo *types.FunctionMetaInfo) types.UserAgency {
	userAgency := types.UserAgency{}
	if funcMetaInfo == nil {
		log.GetLogger().Errorf("funcMeta is nil and agency is empty")
		return userAgency
	}
	agencyID := ""
	if funcMetaInfo.ExtendedMetaData.Role.AppXRole != "" {
		agencyID = funcMetaInfo.ExtendedMetaData.Role.AppXRole
	} else if funcMetaInfo.ExtendedMetaData.Role.XRole != "" {
		agencyID = funcMetaInfo.ExtendedMetaData.Role.XRole
	}
	if agencyID != "" {
		tenantID := funcMetaInfo.FuncMetaData.TenantID
		domainID := funcMetaInfo.FuncMetaData.DomainID
		key := fmt.Sprintf("/sn/agency/business/yrk/tenant/%s/domain/%s/agency/%s", tenantID, domainID, agencyID)
		if value, ok := ur.userAgencyCacheMap.Load(key); ok {
			userAgency, ok = value.(types.UserAgency)
			if !ok {
				log.GetLogger().Errorf("not a valid userAgency cache")
			}
		}
		return userAgency
	}
	return userAgency
}

// watcherFilter will filter instance event from etcd event
func (ur *UserAgencyRegistry) watcherFilter(event *etcd3.Event) bool {
	items := strings.Split(event.Key, keySeparator)
	if len(items) != validEtcdKeyLenForAgency {
		return true
	}
	if items[agencyKeyIndex1] != "agency" || items[agencyBusinessKeyIndex] != "business" ||
		items[agencyTenantKeyIndex] != "tenant" || items[agencyDomainKeyIndex] != "domain" ||
		items[agencyKeyIndex2] != "agency" {
		return true
	}
	return false
}

func (ur *UserAgencyRegistry) watcherHandler(event *etcd3.Event) {
	log.GetLogger().Infof("handling user agency event type %s key %s", event.Type, event.Key)
	switch event.Type {
	case etcd3.PUT:
		var agency = types.UserAgency{}
		if err := json.Unmarshal(event.Value, &agency); err != nil {
			log.GetLogger().Errorf("failed to unmarshal the json, error: %s", err.Error())
			return
		}
		ur.userAgencyCacheMap.Delete(event.Key)
		ur.userAgencyCacheMap.Store(event.Key, agency)
	case etcd3.DELETE:
		if _, ok := ur.userAgencyCacheMap.Load(event.Key); ok {
			ur.userAgencyCacheMap.Delete(event.Key)
		}
		return
	case etcd3.SYNCED:
		log.GetLogger().Infof("userAgency registry ready to receive etcd kv")
	default:
		log.GetLogger().Warnf("unknown event type: %d", event.Type)
	}
}
