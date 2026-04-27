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

// Package instancemanager -
package instancemanager

import (
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/system_function_controller/constant"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasfrontendmanager"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasfunctionmanager"
	"yuanrong.org/kernel/pkg/system_function_controller/instancemanager/faasschedulermanager"
	"yuanrong.org/kernel/pkg/system_function_controller/types"
)

// InstanceManager define InstanceManager management, which manage all system function instance
type InstanceManager struct {
	CommonSchedulerManager       *faasschedulermanager.SchedulerManager
	ExclusivitySchedulerManagers map[string]*faasschedulermanager.SchedulerManager
	FrontendManager              *faasfrontendmanager.FrontendManager
	FunctionManager              *faasfunctionmanager.FunctionManager
}

// HandleEventUpdate handle update event
func (im *InstanceManager) HandleEventUpdate(functionSpec *types.InstanceSpecification, kind types.EventKind) error {
	switch kind {
	case types.EventKindFrontend:
		im.FrontendManager.HandleInstanceUpdate(functionSpec)
	case types.EventKindScheduler:
		tenantID := ""
		if functionSpec.InstanceSpecificationMeta.CreateOptions != nil {
			tenantID = functionSpec.InstanceSpecificationMeta.CreateOptions[constant.SchedulerExclusivity]
		}
		if tenantID == "" {
			im.CommonSchedulerManager.HandleInstanceUpdate(functionSpec)
		} else {
			schedulerManager, ok := im.ExclusivitySchedulerManagers[tenantID]
			if ok && schedulerManager != nil {
				schedulerManager.HandleInstanceUpdate(functionSpec)
			} else {
				// delete scheduler when it has no scheduler manager
				im.CommonSchedulerManager.KillInstance(functionSpec.InstanceID)
			}
		}
	case types.EventKindManager:
		im.FunctionManager.HandleInstanceUpdate(functionSpec)
	default:
		log.GetLogger().Errorf("unknown event kind: %s", kind)
	}
	return nil
}

// HandleEventDelete handle delete event
func (im *InstanceManager) HandleEventDelete(functionSpec *types.InstanceSpecification, kind types.EventKind) error {
	switch kind {
	case types.EventKindFrontend:
		im.FrontendManager.HandleInstanceDelete(functionSpec)
	case types.EventKindScheduler:
		tenantID := ""
		if functionSpec.InstanceSpecificationMeta.CreateOptions != nil {
			tenantID = functionSpec.InstanceSpecificationMeta.CreateOptions[constant.SchedulerExclusivity]
		}
		if tenantID == "" {
			im.CommonSchedulerManager.HandleInstanceDelete(functionSpec)
		} else {
			schedulerManager, ok := im.ExclusivitySchedulerManagers[tenantID]
			if ok && schedulerManager != nil {
				schedulerManager.HandleInstanceDelete(functionSpec)
			}
		}
	case types.EventKindManager:
		im.FunctionManager.HandleInstanceDelete(functionSpec)
	default:
		log.GetLogger().Errorf("unknown event kind: %s", kind)
	}
	return nil
}

// HandleEventRecover handle recover event
func (im *InstanceManager) HandleEventRecover(functionSpec *types.InstanceSpecification, kind types.EventKind) error {
	switch kind {
	case types.EventKindFrontend:
		im.FrontendManager.RecoverInstance(functionSpec)
	case types.EventKindScheduler:
		tenantID := ""
		if functionSpec.InstanceSpecificationMeta.CreateOptions != nil {
			tenantID = functionSpec.InstanceSpecificationMeta.CreateOptions[constant.SchedulerExclusivity]
		}
		if tenantID == "" {
			im.CommonSchedulerManager.RecoverInstance(functionSpec)
		} else {
			schedulerManager, ok := im.ExclusivitySchedulerManagers[tenantID]
			if ok && schedulerManager != nil {
				schedulerManager.RecoverInstance(functionSpec)
			}
		}
	case types.EventKindManager:
		im.FunctionManager.RecoverInstance(functionSpec)
	default:
		log.GetLogger().Errorf("unknown event kind: %s", kind)
	}
	return nil
}
