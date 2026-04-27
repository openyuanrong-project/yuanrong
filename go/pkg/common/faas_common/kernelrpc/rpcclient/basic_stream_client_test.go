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

package rpcclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	rtapi "yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
)

func TestBasicSteamClient_Create(t *testing.T) {
	client := &BasicSteamClient{}

	funcKey := "testFuncKey"
	args := []*rtapi.Arg{}
	createParams := CreateParams{}
	callback := func(result []byte, err snerror.SNError) {
	}

	result, err := client.Create(funcKey, args, createParams, callback)

	assert.Equal(t, "", result)
	assert.Equal(t, ErrUnsupportedMethod, err)
}

func TestBasicSteamClient_SaveState(t *testing.T) {
	client := &BasicSteamClient{}

	state := []byte("test state")

	result, err := client.SaveState(state)

	assert.Equal(t, "", result, "Expected empty string as result")
	assert.Equal(t, ErrUnsupportedMethod, err, "Expected ErrUnsupportedMethod error")
}

func TestBasicSteamClient_LoadState(t *testing.T) {
	client := &BasicSteamClient{}

	checkpointID := "testCheckpointID"

	result, err := client.LoadState(checkpointID)

	assert.Nil(t, result, "Expected nil as result")
	assert.Equal(t, ErrUnsupportedMethod, err, "Expected ErrUnsupportedMethod error")
}

func TestBasicSteamClient_Kill(t *testing.T) {
	client := &BasicSteamClient{}

	instanceID := "testInstanceID"
	signal := int32(9)
	payload := []byte("test payload")

	err := client.Kill(instanceID, signal, payload)

	assert.Equal(t, ErrUnsupportedMethod, err, "Expected ErrUnsupportedMethod error")
}
