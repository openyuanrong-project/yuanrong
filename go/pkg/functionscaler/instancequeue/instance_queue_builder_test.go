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
	"testing"

	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	commontypes "yuanrong.org/kernel/pkg/common/faas_common/types"
	wisecloudTypes "yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	"yuanrong.org/kernel/pkg/functionscaler/metrics"
	"yuanrong.org/kernel/pkg/functionscaler/requestqueue"
	"yuanrong.org/kernel/pkg/functionscaler/scaler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler/concurrencyscheduler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler/microservicescheduler"
	"yuanrong.org/kernel/pkg/functionscaler/scheduler/roundrobinscheduler"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

var testFuncSpec = &types.FunctionSpecification{
	InstanceMetaData: commontypes.InstanceMetaData{
		ConcurrentNum: 100,
	},
}

func TestBuildInstanceQueue(t *testing.T) {
	basicInsQueConfig := &InsQueConfig{
		InstanceType:     types.InstanceTypeScaled,
		FuncSpec:         &types.FunctionSpecification{},
		ResKey:           resspeckey.ResSpecKey{},
		MetricsCollector: &metrics.BucketCollector{},
		InsThdReqQueue:   requestqueue.NewInsAcqReqQueue("", 10),
	}
	insAcqReqQue := &requestqueue.InsAcqReqQueue{}
	metricsCollector := &metrics.BucketCollector{}
	insQue, err := BuildInstanceQueue(basicInsQueConfig, insAcqReqQue, metricsCollector)
	assert.Nil(t, err)
	typedInsQue, ok := insQue.(*ScaledInstanceQueue)
	assert.Equal(t, true, ok)
	assert.IsType(t, &concurrencyscheduler.ScaledConcurrencyScheduler{}, typedInsQue.instanceScheduler)

	basicInsQueConfig = &InsQueConfig{
		InstanceType:     types.InstanceTypeReserved,
		FuncSpec:         &types.FunctionSpecification{},
		ResKey:           resspeckey.ResSpecKey{},
		MetricsCollector: &metrics.BucketCollector{},
	}
	insQue, err = BuildInstanceQueue(basicInsQueConfig, insAcqReqQue, metricsCollector)
	assert.Nil(t, err)
	typedInsQue, ok = insQue.(*ScaledInstanceQueue)
	assert.Equal(t, true, ok)
	assert.IsType(t, &concurrencyscheduler.ReservedConcurrencyScheduler{}, typedInsQue.instanceScheduler)

	basicInsQueConfig = &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			InstanceMetaData: commontypes.InstanceMetaData{
				SchedulePolicy: types.InstanceSchedulePolicyRoundRobin,
			}},
		ResKey:           resspeckey.ResSpecKey{},
		MetricsCollector: &metrics.BucketCollector{},
	}
	insQue, err = BuildInstanceQueue(basicInsQueConfig, insAcqReqQue, metricsCollector)
	assert.Nil(t, err)
	typedInsQue, ok = insQue.(*ScaledInstanceQueue)
	assert.Equal(t, true, ok)
	assert.IsType(t, &roundrobinscheduler.RoundRobinScheduler{}, typedInsQue.instanceScheduler)

	basicInsQueConfig = &InsQueConfig{
		InstanceType: types.InstanceTypeReserved,
		FuncSpec: &types.FunctionSpecification{
			InstanceMetaData: commontypes.InstanceMetaData{
				SchedulePolicy: types.InstanceSchedulePolicyRoundRobin,
			}},
		ResKey:           resspeckey.ResSpecKey{},
		MetricsCollector: &metrics.BucketCollector{},
	}
	insQue, err = BuildInstanceQueue(basicInsQueConfig, insAcqReqQue, metricsCollector)
	assert.Nil(t, err)
	typedInsQue, ok = insQue.(*ScaledInstanceQueue)
	assert.Equal(t, true, ok)
	assert.IsType(t, &roundrobinscheduler.RoundRobinScheduler{}, typedInsQue.instanceScheduler)

	basicInsQueConfig = &InsQueConfig{
		InstanceType: types.InstanceTypeReserved,
		FuncSpec: &types.FunctionSpecification{
			InstanceMetaData: commontypes.InstanceMetaData{
				SchedulePolicy: types.InstanceSchedulePolicyRoundRobin,
				ScalePolicy:    types.InstanceScalePolicyPredict,
			}},
		ResKey:           resspeckey.ResSpecKey{},
		MetricsCollector: &metrics.BucketCollector{},
	}
	insQue, err = BuildInstanceQueue(basicInsQueConfig, insAcqReqQue, metricsCollector)
	assert.Nil(t, err)
	typedInsQue, ok = insQue.(*ScaledInstanceQueue)
	assert.Equal(t, true, ok)
	assert.IsType(t, &roundrobinscheduler.RoundRobinScheduler{}, typedInsQue.instanceScheduler)
	assert.IsType(t, &scaler.ReplicaScaler{}, typedInsQue.instanceScaler)

	basicInsQueConfig = &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			InstanceMetaData: commontypes.InstanceMetaData{
				SchedulePolicy: types.InstanceSchedulePolicyRoundRobin,
				ScalePolicy:    types.InstanceScalePolicyPredict,
			}},
		ResKey:           resspeckey.ResSpecKey{},
		MetricsCollector: &metrics.BucketCollector{},
	}
	insQue, err = BuildInstanceQueue(basicInsQueConfig, insAcqReqQue, metricsCollector)
	assert.Nil(t, err)
	typedInsQue, ok = insQue.(*ScaledInstanceQueue)
	assert.Equal(t, true, ok)
	assert.IsType(t, &roundrobinscheduler.RoundRobinScheduler{}, typedInsQue.instanceScheduler)
	assert.IsType(t, &scaler.PredictScaler{}, typedInsQue.instanceScaler)

	basicInsQueConfig = &InsQueConfig{
		InstanceType: types.InstanceTypeReserved,
		FuncSpec: &types.FunctionSpecification{
			InstanceMetaData: commontypes.InstanceMetaData{
				SchedulePolicy: types.InstanceSchedulePolicyMicroService,
				ScalePolicy:    types.InstanceScalePolicyConcurrency,
			}},
		ResKey:           resspeckey.ResSpecKey{},
		MetricsCollector: &metrics.BucketCollector{},
	}
	insQue, err = BuildInstanceQueue(basicInsQueConfig, insAcqReqQue, metricsCollector)
	assert.Nil(t, err)
	typedInsQue, ok = insQue.(*ScaledInstanceQueue)
	assert.Equal(t, true, ok)
	assert.IsType(t, &microservicescheduler.MicroServiceScheduler{}, typedInsQue.instanceScheduler)

	basicInsQueConfig = &InsQueConfig{
		InstanceType: types.InstanceTypeScaled,
		FuncSpec: &types.FunctionSpecification{
			InstanceMetaData: commontypes.InstanceMetaData{
				SchedulePolicy: types.InstanceSchedulePolicyMicroService,
				ScalePolicy:    types.InstanceScalePolicyConcurrency,
			}},
		ResKey:           resspeckey.ResSpecKey{},
		MetricsCollector: &metrics.BucketCollector{},
	}
	insQue, err = BuildInstanceQueue(basicInsQueConfig, insAcqReqQue, metricsCollector)
	assert.Nil(t, err)
	typedInsQue, ok = insQue.(*ScaledInstanceQueue)
	assert.Equal(t, true, ok)
	assert.IsType(t, &microservicescheduler.MicroServiceScheduler{}, typedInsQue.instanceScheduler)
}

func TestAssembleWithConcurrencyScaler(t *testing.T) {
	err := assembleScalerWithConcurrencyPolicy(&ScaledInstanceQueue{})
	assert.Equal(t, "missing instanceScheduler in instanceQueue", err.Error())
	err = assembleScalerWithConcurrencyPolicy(&ScaledInstanceQueue{instanceScheduler: &fakeInstanceScheduler{}})
	assert.Contains(t, "unsupported instance type", err.Error())
	err = assembleScalerWithConcurrencyPolicy(&ScaledInstanceQueue{funcSpec: testFuncSpec,
		instanceType: types.InstanceTypeScaled, instanceScheduler: &fakeInstanceScheduler{}})
	assert.Equal(t, nil, err)
}

func TestAssembleScalerWithPredictPolicy(t *testing.T) {
	err := assembleScalerWithPredictPolicy(&ScaledInstanceQueue{})
	assert.Equal(t, "missing instanceScheduler in instanceQueue", err.Error())
	err = assembleScalerWithPredictPolicy(&ScaledInstanceQueue{instanceScheduler: &fakeInstanceScheduler{}})
	assert.Contains(t, "unsupported instance type", err.Error())
	err = assembleScalerWithPredictPolicy(&ScaledInstanceQueue{funcSpec: testFuncSpec,
		instanceType: types.InstanceTypeScaled, instanceScheduler: &fakeInstanceScheduler{}})
	assert.Equal(t, nil, err)
}

func TestAssembleScalerWithStaticPolicy(t *testing.T) {
	config.GlobalConfig.ServiceAccountJwt = wisecloudTypes.ServiceAccountJwt{
		ServiceAccount: &wisecloudTypes.ServiceAccount{},
		TlsConfig:      &wisecloudTypes.TLSConfig{},
	}
	err := assembleScalerWithStaticPolicy(&ScaledInstanceQueue{}, requestqueue.NewInsAcqReqQueue("", 10))
	assert.Equal(t, "missing instanceScheduler in instanceQueue", err.Error())
	err = assembleScalerWithStaticPolicy(&ScaledInstanceQueue{instanceScheduler: &fakeInstanceScheduler{}}, requestqueue.NewInsAcqReqQueue("", 10))
	assert.Contains(t, "unsupported instance type", err.Error())
	err = assembleScalerWithStaticPolicy(&ScaledInstanceQueue{funcSpec: testFuncSpec,
		instanceType: types.InstanceTypeScaled, instanceScheduler: &fakeInstanceScheduler{}}, requestqueue.NewInsAcqReqQueue("", 10))
	assert.Equal(t, nil, err)
	err = assembleScalerWithStaticPolicy(&ScaledInstanceQueue{funcSpec: testFuncSpec,
		instanceType: types.InstanceTypeReserved, instanceScheduler: &fakeInstanceScheduler{}}, requestqueue.NewInsAcqReqQueue("", 10))
	assert.Equal(t, nil, err)
}
