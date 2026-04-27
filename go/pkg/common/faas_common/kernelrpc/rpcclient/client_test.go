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
	"time"

	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/snerror"
)

func TestSyncInvoke(t *testing.T) {
	funcKey := "testFuncKey"
	instanceID := "testInstanceID"
	args := []string{"arg1", "arg2"}
	invokeParams := InvokeParams{
		TransportParams: TransportParams{
			Timeout:       500 * time.Millisecond,
			RetryInterval: 100 * time.Millisecond,
			RetryNumber:   2,
		},
		InvokeOptions: map[string]string{"option1": "value1"},
		RequestID:     "testRequestID",
		TraceID:       "testTraceID",
	}

	t.Run("Success without retries", func(t *testing.T) {
		asyncInvoke := func(funcKey, instanceID string, args []string, invokeParams InvokeParams,
			callback func(result []byte, err snerror.SNError)) snerror.SNError {
			go func() {
				time.Sleep(50 * time.Millisecond) // 模拟异步调用的延迟
				callback([]byte("success"), nil)
			}()
			return nil
		}

		resultData, resultError := SyncInvoke(asyncInvoke, funcKey, instanceID, args, invokeParams)

		assert.Nil(t, resultError)
		assert.Equal(t, []byte("success"), resultData)
	})

	t.Run("Error then success after retries", func(t *testing.T) {
		var callCount int
		asyncInvoke := func(funcKey, instanceID string, args []string, invokeParams InvokeParams,
			callback func(result []byte, err snerror.SNError)) snerror.SNError {
			go func() {
				time.Sleep(50 * time.Millisecond)
				if callCount < 1 {
					callCount++
					callback(nil, snerror.New(1, "temporary error"))
				} else {
					callback([]byte("recovered success"), nil)
				}
			}()
			return nil
		}

		resultData, resultError := SyncInvoke(asyncInvoke, funcKey, instanceID, args, invokeParams)

		assert.Nil(t, resultError)
		assert.Equal(t, []byte("recovered success"), resultData)
	})

	t.Run("Error with retries exhausted", func(t *testing.T) {
		asyncInvoke := func(funcKey, instanceID string, args []string, invokeParams InvokeParams,
			callback func(result []byte, err snerror.SNError)) snerror.SNError {
			go func() {
				time.Sleep(50 * time.Millisecond)
				callback(nil, snerror.New(1, "persistent error"))
			}()
			return nil
		}

		resultData, resultError := SyncInvoke(asyncInvoke, funcKey, instanceID, args, invokeParams)

		assert.NotNil(t, resultError)
		assert.Equal(t, "persistent error", resultError.Error())
		assert.Nil(t, resultData)
	})

	t.Run("Timeout without response", func(t *testing.T) {
		asyncInvoke := func(funcKey, instanceID string, args []string, invokeParams InvokeParams,
			callback func(result []byte, err snerror.SNError)) snerror.SNError {
			return nil
		}

		resultData, resultError := SyncInvoke(asyncInvoke, funcKey, instanceID, args, invokeParams)

		assert.NotNil(t, resultError)
		assert.Equal(t, ErrKernelClientTimeout, resultError)
		assert.Nil(t, resultData)
	})
}
