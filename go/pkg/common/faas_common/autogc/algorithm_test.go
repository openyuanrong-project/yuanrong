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

package autogc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAlg(t *testing.T) {
	assert.Equal(t, 4*GB, 4294967296)

	alg := DefaultAlg{}
	alg.Init(4*GB, 3200*MB)

	tests := []struct {
		current  uint64
		excepted int
	}{
		{
			current:  40 * MB,
			excepted: defaultMaxGOGC,
		},
		{
			current:  3200 * MB,
			excepted: 1,
		},
		{
			current:  3201 * MB,
			excepted: 1,
		},
		{
			current:  3100 * MB,
			excepted: 3,
		},
		{
			current:  2000 * MB,
			excepted: 60,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.excepted, alg.NextGOGC(test.current, 0))
	}
}
