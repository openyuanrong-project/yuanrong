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
)

func TestInstanceCreateQueue(t *testing.T) {
	queue := NewInstanceCreateQueue()
	assert.NotNil(t, queue)
	queue.push(&InstanceCreateRequest{})
	req1 := queue.getForCreate()
	assert.NotNil(t, req1)
	req2 := queue.getForCancel()
	assert.Nil(t, req2)
	queue.push(&InstanceCreateRequest{})
	req3 := queue.getForCancel()
	assert.NotNil(t, req3)
	req4 := queue.getForCancel()
	assert.Nil(t, req4)
	queue.push(&InstanceCreateRequest{})
	assert.Equal(t, 1, queue.getQueLen())
}
