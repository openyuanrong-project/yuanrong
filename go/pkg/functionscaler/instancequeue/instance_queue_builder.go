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

// Package instancequeue -
package instancequeue

import (
	"errors"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/requestqueue"
	"yuanrong.org/kernel/pkg/functionscaler/scaler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler/concurrencyscheduler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler/microservicescheduler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler/roundrobinscheduler"
	"yuanrong.org/kernel/pkg/functionscaler/types"
	"yuanrong.org/kernel/pkg/functionscaler/utils"
)

// ErrUnsupportedInstanceType is the error of unsupported instance type
var ErrUnsupportedInstanceType = errors.New("unsupported instance type")

// BuildInstanceQueue builds an instanceQueue
func BuildInstanceQueue(config *InsQueConfig, insAcqReqQue *requestqueue.InsAcqReqQueue,
	metricsCollector metrics.Collector,
) (InstanceQueue, error) {
	if config.InstanceType == types.InstanceTypeOnDemand {
		return NewOnDemandInstanceQueue(config), nil
	}
	instanceQueue := NewScaledInstanceQueue(config, metricsCollector)
	var err error
	err = AssembleScheduler(config.FuncSpec.InstanceMetaData.SchedulePolicy, instanceQueue, insAcqReqQue)
	if err != nil {
		log.GetLogger().Errorf("failed to assemble instanceScheduler for function %s", config.FuncSpec.FuncKey)
		return nil, err
	}
	switch config.FuncSpec.InstanceMetaData.ScalePolicy {
	case types.InstanceScalePolicyPredict:
		err = assembleScalerWithPredictPolicy(instanceQueue)
	case types.InstanceScalePolicyStaticFunction:
		err = assembleScalerWithStaticPolicy(instanceQueue, insAcqReqQue)
	default:
		err = assembleScalerWithConcurrencyPolicy(instanceQueue)
	}
	if err != nil {
		log.GetLogger().Errorf("failed to assemble instanceScaler for function %s err %v",
			config.FuncSpec.FuncKey, err)
		return nil, err
	}
	if instanceQueue.isFuncOwner {
		log.GetLogger().Infof("instance queue is funcOwner of function %s", config.FuncSpec.FuncKey)
		instanceQueue.instanceScheduler.HandleFuncOwnerUpdate(true)
	}
	return instanceQueue, nil
}

// AssembleScheduler assemble scheduler to queue
func AssembleScheduler(schedulePolicy string, instanceQueue *ScaledInstanceQueue,
	insAcqReqQue *requestqueue.InsAcqReqQueue,
) error {
	var err error
	switch schedulePolicy {
	case types.InstanceSchedulePolicyConcurrency:
		err = assembleSchedulerWithConcurrencyPolicy(instanceQueue, insAcqReqQue)
	case types.InstanceSchedulePolicyRoundRobin:
		err = assembleSchedulerWithRoundRobinPolicy(instanceQueue, insAcqReqQue)
	case types.InstanceSchedulePolicyMicroService:
		err = assembleSchedulerWithMicroService(instanceQueue)
	default:
		err = assembleSchedulerWithConcurrencyPolicy(instanceQueue, insAcqReqQue)
	}
	if err != nil {
		return err
	}
	return nil
}

func assembleSchedulerWithConcurrencyPolicy(instanceQueue *ScaledInstanceQueue,
	insThdReqQueue *requestqueue.InsAcqReqQueue,
) error {
	requestTimeout := utils.GetRequestTimeout(instanceQueue.funcSpec)
	var instanceScheduler scheduler.InstanceScheduler
	if instanceQueue.instanceType == types.InstanceTypeReserved {
		instanceScheduler = concurrencyscheduler.NewReservedConcurrencyScheduler(instanceQueue.funcSpec,
			instanceQueue.resKey, requestTimeout, insThdReqQueue)
		log.GetLogger().Infof("assemble with concurrency scheduler for %s instance queue of function %s",
			instanceQueue.instanceType, instanceQueue.funcKeyWithRes)
	} else if instanceQueue.instanceType == types.InstanceTypeScaled {
		instanceScheduler = concurrencyscheduler.NewScaledConcurrencyScheduler(instanceQueue.funcSpec,
			instanceQueue.resKey, insThdReqQueue)
		log.GetLogger().Infof("assemble with concurrency scheduler for %s instance queue of function %s",
			instanceQueue.instanceType, instanceQueue.funcKeyWithRes)
	} else {
		log.GetLogger().Errorf("unsupported instance type %s", instanceQueue.instanceType)
		return ErrUnsupportedInstanceType
	}
	instanceQueue.SetInstanceScheduler(instanceScheduler)
	return nil
}

func assembleSchedulerWithRoundRobinPolicy(instanceQueue *ScaledInstanceQueue,
	insThdReqQueue *requestqueue.InsAcqReqQueue,
) error {
	requestTimeout := utils.GetRequestTimeout(instanceQueue.funcSpec)
	var instanceScheduler scheduler.InstanceScheduler
	if instanceQueue.instanceType == types.InstanceTypeReserved {
		instanceScheduler = roundrobinscheduler.NewRoundRobinScheduler(instanceQueue.funcKeyWithRes, true, requestTimeout)
		log.GetLogger().Infof("assemble with round-robin scheduler for %s instance queue of function %s",
			instanceQueue.instanceType, instanceQueue.funcKeyWithRes)
	} else if instanceQueue.instanceType == types.InstanceTypeScaled {
		instanceScheduler = roundrobinscheduler.NewRoundRobinScheduler(instanceQueue.funcKeyWithRes, false, requestTimeout)
		log.GetLogger().Infof("assemble with round-robin scheduler for %s instance queue of function %s",
			instanceQueue.instanceType, instanceQueue.funcKeyWithRes)
	} else {
		log.GetLogger().Errorf("unsupported instance type %s", instanceQueue.instanceType)
		return ErrUnsupportedInstanceType
	}
	instanceQueue.SetInstanceScheduler(instanceScheduler)
	return nil
}

func assembleSchedulerWithMicroService(instanceQueue *ScaledInstanceQueue) error {
	var instanceScheduler scheduler.InstanceScheduler
	if instanceQueue.instanceType == types.InstanceTypeReserved {
		instanceScheduler = microservicescheduler.NewMicroServiceScheduler(instanceQueue.funcKeyWithRes,
			config.GlobalConfig.MicroServiceSchedulingPolicy)
		log.GetLogger().Infof("assemble with microService for %s instance queue of function %s",
			instanceQueue.instanceType, instanceQueue.funcKeyWithRes)
	} else if instanceQueue.instanceType == types.InstanceTypeScaled {
		instanceScheduler = microservicescheduler.NewMicroServiceScheduler(instanceQueue.funcKeyWithRes,
			config.GlobalConfig.MicroServiceSchedulingPolicy)
		log.GetLogger().Infof("assemble with microService for %s instance queue of function %s",
			instanceQueue.instanceType, instanceQueue.funcKeyWithRes)
	} else {
		log.GetLogger().Errorf("unsupported instance type %s", instanceQueue.instanceType)
		return ErrUnsupportedInstanceType
	}
	instanceQueue.SetInstanceScheduler(instanceScheduler)
	return nil
}

func assembleScalerWithStaticPolicy(instanceQueue *ScaledInstanceQueue,
	insAcqReqQue *requestqueue.InsAcqReqQueue,
) error {
	if instanceQueue.instanceScheduler == nil {
		return errors.New("missing instanceScheduler in instanceQueue")
	}
	var instanceScaler scaler.InstanceScaler
	// only reserve queue and func owner can trigger nuwa cold start
	if instanceQueue.instanceType == types.InstanceTypeReserved {
		instanceScaler = scaler.NewWiseCloudScaler(instanceQueue.funcKeyWithRes, instanceQueue.resKey, true,
			insAcqReqQue.HandleCreateError)
	} else if instanceQueue.instanceType == types.InstanceTypeScaled {
		instanceScaler = scaler.NewWiseCloudScaler(instanceQueue.funcKeyWithRes, instanceQueue.resKey, false,
			insAcqReqQue.HandleCreateError)
	} else {
		log.GetLogger().Errorf("unsupported instance type %s", instanceQueue.instanceType)
		return ErrUnsupportedInstanceType
	}
	// set concurrentNum for first time
	instanceScaler.HandleFuncSpecUpdate(instanceQueue.funcSpec)
	instanceQueue.instanceScheduler.ConnectWithInstanceScaler(instanceScaler)
	instanceQueue.SetInstanceScaler(instanceScaler)
	return nil
}

func assembleScalerWithConcurrencyPolicy(instanceQueue *ScaledInstanceQueue) error {
	if instanceQueue.instanceScheduler == nil {
		return errors.New("missing instanceScheduler in instanceQueue")
	}
	var instanceScaler scaler.InstanceScaler
	if instanceQueue.instanceType == types.InstanceTypeReserved {
		instanceScaler = scaler.NewReplicaScaler(instanceQueue.funcKeyWithRes, instanceQueue.metricsCollector,
			instanceQueue.ScaleUpHandler, instanceQueue.ScaleDownHandler)
		log.GetLogger().Infof("assemble with replica scaler for %s instance queue of function %s",
			instanceQueue.instanceType, instanceQueue.funcKeyWithRes)
	} else if instanceQueue.instanceType == types.InstanceTypeScaled {
		autoScaleConfig := types.AutoScaleConfig{
			SLAQuota:      config.GlobalConfig.AutoScaleConfig.SLAQuota,
			ScaleDownTime: config.GlobalConfig.AutoScaleConfig.ScaleDownTime,
			BurstScaleNum: config.GlobalConfig.AutoScaleConfig.BurstScaleNum,
		}
		if instanceQueue.funcSpec.FuncMetaData.AutoScaleConfig.SLAQuota != -1 {
			autoScaleConfig.SLAQuota = instanceQueue.funcSpec.FuncMetaData.AutoScaleConfig.SLAQuota
		}
		if instanceQueue.funcSpec.FuncMetaData.AutoScaleConfig.ScaleDownTime != -1 {
			autoScaleConfig.ScaleDownTime = instanceQueue.funcSpec.FuncMetaData.AutoScaleConfig.ScaleDownTime
		}
		if instanceQueue.funcSpec.FuncMetaData.AutoScaleConfig.BurstScaleNum != -1 {
			autoScaleConfig.BurstScaleNum = instanceQueue.funcSpec.FuncMetaData.AutoScaleConfig.BurstScaleNum
		}
		instanceScheduler, ok := instanceQueue.instanceScheduler.(*concurrencyscheduler.ScaledConcurrencyScheduler)
		if ok {
			instanceScaler = scaler.NewAutoScaler(instanceQueue.funcKeyWithRes, instanceQueue.metricsCollector,
				instanceScheduler.GetReqQueLen, instanceQueue.ScaleUpHandler, instanceQueue.ScaleDownHandler, autoScaleConfig)
		} else {
			log.GetLogger().Warnf("missing concurrencyScheduler when build concurrencyScaler for scaled instance")
			instanceScaler = scaler.NewAutoScaler(instanceQueue.funcKeyWithRes, instanceQueue.metricsCollector,
				func() int { return 0 }, instanceQueue.ScaleUpHandler, instanceQueue.ScaleDownHandler, autoScaleConfig)
		}
		log.GetLogger().Infof("assemble with auto scaler for %s instance queue of function %s",
			instanceQueue.instanceType, instanceQueue.funcKeyWithRes)
	} else {
		log.GetLogger().Errorf("unsupported instance type %s", instanceQueue.instanceType)
		return ErrUnsupportedInstanceType
	}
	// set concurrentNum for first time
	instanceScaler.HandleFuncSpecUpdate(instanceQueue.funcSpec)
	instanceQueue.instanceScheduler.ConnectWithInstanceScaler(instanceScaler)
	instanceQueue.SetInstanceScaler(instanceScaler)
	return nil
}

func assembleScalerWithPredictPolicy(instanceQueue *ScaledInstanceQueue) error {
	if instanceQueue.instanceScheduler == nil {
		return errors.New("missing instanceScheduler in instanceQueue")
	}
	var instanceScaler scaler.InstanceScaler
	switch instanceQueue.instanceType {
	case types.InstanceTypeReserved:
		instanceScaler = scaler.NewReplicaScaler(instanceQueue.funcKeyWithRes, instanceQueue.metricsCollector,
			instanceQueue.ScaleUpHandler, instanceQueue.ScaleDownHandler)
		log.GetLogger().Infof("assemble with replica scaler for %s instance queue of function %s",
			instanceQueue.instanceType, instanceQueue.funcKeyWithRes)
	case types.InstanceTypeScaled:
		instanceScheduler, ok := instanceQueue.instanceScheduler.(*concurrencyscheduler.ScaledConcurrencyScheduler)
		if ok {
			instanceScaler = scaler.NewPredictScaler(instanceQueue.funcKeyWithRes, instanceQueue.metricsCollector,
				instanceScheduler.GetReqQueLen, instanceQueue.ScaleUpHandler, instanceQueue.ScaleDownHandler)
		} else {
			log.GetLogger().Warnf("missing concurrencyScheduler when build concurrencyScaler for scaled instance")
			instanceScaler = scaler.NewPredictScaler(instanceQueue.funcKeyWithRes, instanceQueue.metricsCollector,
				func() int { return 0 }, instanceQueue.ScaleUpHandler, instanceQueue.ScaleDownHandler)
		}
		log.GetLogger().Infof("assemble with predict scaler for %s instance queue of function %s",
			instanceQueue.instanceType, instanceQueue.funcKeyWithRes)
	default:
		log.GetLogger().Errorf("unsupported instance type %s", instanceQueue.instanceType)
		return ErrUnsupportedInstanceType
	}
	// set concurrentNum for first time
	instanceScaler.HandleFuncSpecUpdate(instanceQueue.funcSpec)
	instanceQueue.instanceScheduler.ConnectWithInstanceScaler(instanceScaler)
	instanceQueue.SetInstanceScaler(instanceScaler)
	return nil
}
