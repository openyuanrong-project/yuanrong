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
	"os"
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

func TestInitAutoGOGC(t *testing.T) {
	InitAutoGOGC()
	runtime.GC()
	assert.Equal(t, 100, previousGOGC)
}

func TestInitAutoGOGC2(t *testing.T) {
	patches := utils.InitPatchSlice()
	patches.Append(utils.PatchSlice{
		gomonkey.ApplyFunc(debug.SetGCPercent,
			func(percent int) int {
				return 100
			})})
	defer patches.ResetAll()
	os.Setenv("AUTO_GC_MEMORY_THRESHOLD", "120")
	InitAutoGOGC()
	runtime.GC()
	assert.Equal(t, 100, previousGOGC)
}
