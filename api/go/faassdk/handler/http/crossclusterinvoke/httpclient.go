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

// Package crossclusterinvoke -
package crossclusterinvoke

import (
	"sync"
	"sync/atomic"

	"github.com/valyala/fasthttp"
)

var httpClients []*fasthttp.Client
var httpClientsIndex uint32
var httpClientsOnce sync.Once
var httpClientNum uint32 = 5

// GetHttpClient -
func GetHttpClient() *fasthttp.Client {
	httpClientsOnce.Do(func() {
		httpClients = make([]*fasthttp.Client, httpClientNum, httpClientNum)
		for i := 0; i < int(httpClientNum); i++ {
			httpClients[i] = &fasthttp.Client{}
		}
		httpClientsIndex = 0
	})
	index := atomic.AddUint32(&httpClientsIndex, 1) % httpClientNum
	return httpClients[index]
}
