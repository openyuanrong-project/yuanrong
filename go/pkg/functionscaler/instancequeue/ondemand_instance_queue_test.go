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
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestNewOnDemandInstanceQueue(t *testing.T) {
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeOnDemand,
		FuncSpec:     &types.FunctionSpecification{},
		ResKey:       resspeckey.ResSpecKey{},
	}
	q := NewOnDemandInstanceQueue(basicInsQueConfig)
	assert.NotNil(t, q)
}

func TestOnDemandAcquireInstance(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	var createErr error
	createInstanceFunc := func(traceID string, name string, _ types.InstanceType, _ resspeckey.ResSpecKey, _ []byte) (
		*types.Instance, error) {
		if createErr != nil {
			return nil, createErr
		}
		return &types.Instance{InstanceID: "testInstance1", InstanceName: name}, nil
	}
	deleteInstanceFunc := func(*types.Instance) error { return nil }
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeOnDemand,
		FuncSpec: &types.FunctionSpecification{
			FuncCtx: ctx,
		},
		ResKey:             resspeckey.ResSpecKey{},
		CreateInstanceFunc: createInstanceFunc,
		DeleteInstanceFunc: deleteInstanceFunc,
	}
	q := NewOnDemandInstanceQueue(basicInsQueConfig)
	ins, err := q.CreateInstance(&types.InstanceCreateRequest{InstanceName: "testInsName1"})
	assert.Nil(t, err)
	assert.Equal(t, "testInstance1", ins.InstanceID)
	insAlloc1, err := q.AcquireInstance(&types.InstanceAcquireRequest{InstanceName: "testInsName1"})
	assert.Nil(t, err)
	assert.Equal(t, "testInstance1", insAlloc1.AllocationID)
	insAlloc2, err := q.AcquireInstance(&types.InstanceAcquireRequest{DesignateInstanceID: "testInstance1"})
	assert.Nil(t, err)
	assert.Equal(t, "testInstance1", insAlloc2.AllocationID)
	cancel()
	_, err = q.AcquireInstance(&types.InstanceAcquireRequest{InstanceName: "testInsName1"})
	assert.NotNil(t, err)
}

func TestCreateInstance(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	var createErr error
	createInstanceFunc := func(traceID string, name string, _ types.InstanceType, _ resspeckey.ResSpecKey, _ []byte) (
		*types.Instance, error) {
		if createErr != nil {
			return nil, createErr
		}
		return &types.Instance{InstanceID: "testInstance1", InstanceName: name}, nil
	}
	deleteInstanceFunc := func(*types.Instance) error { return nil }
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeOnDemand,
		FuncSpec: &types.FunctionSpecification{
			FuncCtx: ctx,
		},
		ResKey:             resspeckey.ResSpecKey{},
		CreateInstanceFunc: createInstanceFunc,
		DeleteInstanceFunc: deleteInstanceFunc,
	}
	q := NewOnDemandInstanceQueue(basicInsQueConfig)
	ins, err := q.CreateInstance(&types.InstanceCreateRequest{InstanceName: "testInsName1"})
	assert.Nil(t, err)
	assert.Equal(t, "testInstance1", ins.InstanceID)
	createErr = errors.New("some error")
	ins, err = q.CreateInstance(&types.InstanceCreateRequest{InstanceName: "testInsName1"})
	assert.NotNil(t, err)
	assert.Nil(t, ins)
	cancel()
	ins, err = q.CreateInstance(&types.InstanceCreateRequest{InstanceName: "testInsName1"})
	assert.NotNil(t, err)
	assert.Nil(t, ins)
}

func TestDeleteInstance(t *testing.T) {
	createInstanceFunc := func(traceID string, name string, _ types.InstanceType, _ resspeckey.ResSpecKey, _ []byte) (
		*types.Instance, error) {
		return &types.Instance{InstanceID: "testInstance1", InstanceName: name}, nil
	}
	deleteInstanceFunc := func(*types.Instance) error { return nil }
	basicInsQueConfig := &InsQueConfig{
		InstanceType:       types.InstanceTypeOnDemand,
		FuncSpec:           &types.FunctionSpecification{FuncCtx: context.TODO()},
		ResKey:             resspeckey.ResSpecKey{},
		CreateInstanceFunc: createInstanceFunc,
		DeleteInstanceFunc: deleteInstanceFunc,
	}
	q := NewOnDemandInstanceQueue(basicInsQueConfig)
	ins, _ := q.CreateInstance(&types.InstanceCreateRequest{InstanceName: "testInsName1"})
	err := q.DeleteInstance(ins)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(q.instanceMap))
	assert.Equal(t, 0, len(q.insNameMap))
}

func TestOnDemandHandleInstanceDelete(t *testing.T) {
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeOnDemand,
		FuncSpec:     &types.FunctionSpecification{FuncCtx: context.TODO()},
		ResKey:       resspeckey.ResSpecKey{},
	}
	q := NewOnDemandInstanceQueue(basicInsQueConfig)
	instance := &types.Instance{
		InstanceID:   "testInstance1",
		InstanceName: "testInsName1",
	}
	q.instanceMap[instance.InstanceID] = instance
	q.insNameMap[instance.InstanceName] = instance
	q.HandleInstanceDelete(instance)
	assert.Equal(t, 0, len(q.instanceMap))
	assert.Equal(t, 0, len(q.insNameMap))
}

func TestOnDemandHandleFuncSpecUpdate(t *testing.T) {
	deleteInstanceFunc := func(*types.Instance) error { return nil }
	basicInsQueConfig := &InsQueConfig{
		InstanceType:       types.InstanceTypeOnDemand,
		FuncSpec:           &types.FunctionSpecification{FuncCtx: context.TODO()},
		ResKey:             resspeckey.ResSpecKey{},
		DeleteInstanceFunc: deleteInstanceFunc,
	}
	q := NewOnDemandInstanceQueue(basicInsQueConfig)
	instance := &types.Instance{
		InstanceID:   "testInstance1",
		InstanceName: "testInsName1",
	}
	q.instanceMap[instance.InstanceID] = instance
	q.insNameMap[instance.InstanceName] = instance
	q.HandleFuncSpecUpdate(&types.FunctionSpecification{FuncMetaSignature: "testFuncSig"})
	assert.Equal(t, 0, len(q.instanceMap))
	assert.Equal(t, 0, len(q.insNameMap))
}

func TestOnDemandGetInstanceNumber(t *testing.T) {
	basicInsQueConfig := &InsQueConfig{
		InstanceType: types.InstanceTypeOnDemand,
		FuncSpec:     &types.FunctionSpecification{FuncCtx: context.TODO()},
		ResKey:       resspeckey.ResSpecKey{},
	}
	q := NewOnDemandInstanceQueue(basicInsQueConfig)
	instance := &types.Instance{
		InstanceID:   "testInstance1",
		InstanceName: "testInsName1",
	}
	q.instanceMap[instance.InstanceID] = instance
	q.insNameMap[instance.InstanceName] = instance
	assert.Equal(t, 1, q.GetInstanceNumber(true))
}
