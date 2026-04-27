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

// Package main for example
package main

import (
	"yuanrong.org/kernel/runtime/faassdk"
	"yuanrong.org/kernel/runtime/faassdk/faas-sdk/go-api/context"
)

func initEntry(ctx context.RuntimeContext) {
	ctx.GetLogger().Infof("info log in initHandler")
}

func preStopEntry(ctx context.RuntimeContext) {
	ctx.GetLogger().Infof("info log in preStopHandler")
}

func callEntry(_ []byte, _ context.RuntimeContext) (interface{}, error) {
	return "hello world", nil
}

func main() {
	faassdk.RegisterInitializerFunction(initEntry)
	faassdk.RegisterPreStopFunction(preStopEntry)
	faassdk.Register(callEntry)
}
