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

// Package types -
package types

import (
	"k8s.io/apimachinery/pkg/types"

	"yuanrong.org/kernel/pkg/common/faas_common/alarm"
	"yuanrong.org/kernel/pkg/common/faas_common/crypto"
	"yuanrong.org/kernel/pkg/common/faas_common/etcd3"
	commonTypes "yuanrong.org/kernel/pkg/common/faas_common/types"
)

const (
	// KillSignalVal Kill instances of job
	KillSignalVal = 2 // Kill instances of job
)

type (
	// EventType defines registry event type
	EventType string
)

const (
	// SubEventTypeUpdate is update type of subscribe event
	SubEventTypeUpdate EventType = "update"
	// SubEventTypeDelete is delete type of subscribe event
	SubEventTypeDelete EventType = "delete"
	// PatPodRunningStatus is active pat pod
	PatPodRunningStatus string = "Active"
	// NetWorkDelegateKey is createoption's key of network info
	NetWorkDelegateKey = "DELEGATE_NETWORK_CONFIG"
	// PatIdlePodAnnotationKey is idlePod Key on annotation
	PatIdlePodAnnotationKey = "patservice.cap.io/idle-pods"
)

// ManagerConfig is the config used by faas frontend function
type ManagerConfig struct {
	HTTPSEnable          bool             `json:"httpsEnable" valid:"optional"`
	FunctionCapability   int              `json:"functionCapability" valid:"optional"`
	AuthenticationEnable bool             `json:"authenticationEnable" valid:"optional"`
	LeaseRenewMinute     int              `json:"leaseRenewMinute" valid:"optional"`
	EnableVPCManage      bool             `json:"enableVPCManage" valid:"optional"`
	EnableHealthCheck    bool             `json:"enableHealthCheck" valid:"optional"`
	RouterEtcd           etcd3.EtcdConfig `json:"routerEtcd" valid:"optional"`
	MetaEtcd             etcd3.EtcdConfig `json:"metaEtcd" valid:"optional"`
	AlarmConfig          alarm.Config     `json:"alarmConfig" valid:"optional"`
	SccConfig            crypto.SccConfig `json:"sccConfig" valid:"optional"`
}

// VPCEvent is vpc event def
type VPCEvent struct {
	EventType EventType
	InsInfo   *commonTypes.InstanceSpecification
	PatInfo   *Pat
}

// NetWorkConfig is instance's delegate network info
type NetWorkConfig struct {
	PatInstances []types.NamespacedName `json:"patInstances"`
}
