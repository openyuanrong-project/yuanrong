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

// Package etcdkey contains etcd key definition and tools
package etcdkey

import (
	"fmt"
	"strings"

	"yuanrong.org/kernel/pkg/common/faas_common/urnutils"
)

const (
	keyFormat      = "/sn/%s/business/%s/tenant/%s/function/%s/version/%s/%s/%s"
	keySeparator   = "/"
	stateWorkerLen = 13

	cronTriggerTenantIndex      = 8
	functionMetadataTenantIndex = 4
)

// Index of element in etcd key
const (
	prefixIndex = iota + 1
	typeIndex
	businessIDKey
	businessIDIndex
	tenantIDKey
	tenantIDIndex
	functionKey
	functionIndex
	versionKey
	versionIndex
	zoneIndex
	keyIndex
)

// Index of instance element in etcd key
const (
	instancePrefixIndex = iota + 1
	instanceTypeIndex
	instanceBusinessIDKey
	instanceBusinessIDIndex
	instanceTenantIDKey
	instanceTenantIDIndex
	instanceZoneIndex
	instanceFunctionNameIndex
	instanceUrnVersionIndex = 10
	instanceInstanceIDIndex = 13
	instanceKeyLen          = 14
)

// EtcdKey etcd key interface definition
type EtcdKey interface {
	String() string
	ParseFrom(key string) error
}

// StateWorkersKey state workers key
type StateWorkersKey struct {
	TypeKey    string
	BusinessID string
	TenantID   string
	Function   string
	Version    string
	Zone       string
	StateID    string
}

// String serialize state workers key struct to string
func (s *StateWorkersKey) String() string {
	return fmt.Sprintf(keyFormat, s.TypeKey, s.BusinessID, s.TenantID, s.Function, s.Version, s.Zone, s.StateID)
}

// ParseFrom parse string to state workers key struct
func (s *StateWorkersKey) ParseFrom(etcdKey string) error {
	elements := strings.Split(etcdKey, keySeparator)
	if len(elements) != stateWorkerLen {
		return fmt.Errorf("failed to parse etcd key from %s", etcdKey)
	}
	s.TypeKey = elements[keyIndex]
	s.BusinessID = elements[businessIDIndex]
	s.TenantID = elements[tenantIDIndex]
	s.Function = elements[functionIndex]
	s.Version = elements[versionIndex]
	s.Zone = elements[zoneIndex]
	s.StateID = elements[keyIndex]

	return nil
}

// WorkerInstanceKey is the etcd key path of worker instance
type WorkerInstanceKey struct {
	TypeKey    string
	BusinessID string
	TenantID   string
	Function   string
	Version    string
	Zone       string
	Instance   string
}

// String serialize worker instance key struct to string
func (w *WorkerInstanceKey) String() string {
	return fmt.Sprintf(keyFormat, w.TypeKey, w.BusinessID, w.TenantID, w.Function, w.Version, w.Zone, w.Instance)
}

// ParseFrom parse string to worker instance key struct
func (w *WorkerInstanceKey) ParseFrom(etcdKey string) error {
	elements := strings.Split(etcdKey, keySeparator)
	if len(elements) != stateWorkerLen {
		return fmt.Errorf("failed to parse etcd key from %s", etcdKey)
	}
	w.TypeKey = elements[typeIndex]
	w.BusinessID = elements[businessIDIndex]
	w.TenantID = elements[tenantIDIndex]
	w.Function = elements[functionIndex]
	w.Version = elements[versionIndex]
	w.Zone = elements[zoneIndex]
	w.Instance = elements[keyIndex]
	return nil
}

// AnonymizeTenantCommonEtcdKey Anonymize tenant info in common etcd key
// /yr/functions/business/yrk/tenant/8e08d5cc0ad34032bba8d636040a278c/function/0-test1-addone/version/$latest
func AnonymizeTenantCommonEtcdKey(etcdKey string) string {
	elements := strings.Split(etcdKey, keySeparator)
	if len(elements) <= tenantIDIndex {
		return etcdKey
	}
	elements[tenantIDIndex] = urnutils.Anonymize(elements[tenantIDIndex])
	return strings.Join(elements, keySeparator)
}

// AnonymizeTenantCronTriggerEtcdKey Anonymize tenant info in cron trigger etcd key
// /sn/triggers/triggerType/CRON/business/yrk/tenant/i1fe539427b24702acc11fbb4e134e17/function/pytzip/version/$latest/398e2ca2-a160-4c22-bd05-94a90a5326e2
func AnonymizeTenantCronTriggerEtcdKey(etcdKey string) string {
	elements := strings.Split(etcdKey, keySeparator)
	if len(elements) <= cronTriggerTenantIndex {
		return etcdKey
	}
	elements[cronTriggerTenantIndex] = urnutils.Anonymize(elements[cronTriggerTenantIndex])
	return strings.Join(elements, keySeparator)
}

// AnonymizeTenantFunctionMetadataEtcdKey Anonymize tenant info in function metadata etcd key
// /repo/FunctionVersion/business/tenant/funcName/version/
func AnonymizeTenantFunctionMetadataEtcdKey(etcdKey string) string {
	elements := strings.Split(etcdKey, keySeparator)
	if len(elements) <= functionMetadataTenantIndex {
		return etcdKey
	}
	elements[functionMetadataTenantIndex] = urnutils.Anonymize(elements[functionMetadataTenantIndex])
	return strings.Join(elements, keySeparator)
}

// FunctionInstanceKey is the etcd key path of function instance
type FunctionInstanceKey struct {
	TypeKey      string
	BusinessID   string
	TenantID     string
	FunctionName string
	Version      string
	Zone         string
	InstanceID   string
}

// ParseFrom parse string to function instance key struct
// /sn/instance/business/yrk/tenant/default/function/0-defaultservice-py/version/$latest
// /defaultaz/8c9fa45600e5f44f00/10000000-0000-4000-b653-c11128589d17
func (f *FunctionInstanceKey) ParseFrom(etcdKey string) error {
	elements := strings.Split(etcdKey, keySeparator)
	if len(elements) != instanceKeyLen {
		return fmt.Errorf("failed to parse etcd key from %s: invalid key length", etcdKey)
	}
	f.TypeKey = elements[instanceTypeIndex]
	f.BusinessID = elements[instanceBusinessIDIndex]
	f.TenantID = elements[instanceTenantIDIndex]
	f.Zone = elements[instanceZoneIndex]
	f.InstanceID = elements[instanceInstanceIDIndex]
	return nil
}
