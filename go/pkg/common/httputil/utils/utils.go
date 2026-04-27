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

package utils

import "github.com/gin-gonic/gin"

// ParseHeader 解析请求头
func ParseHeader(ctx *gin.Context) map[string]string {
	if ctx == nil || ctx.Request == nil || len(ctx.Request.Header) == 0 {
		return map[string]string{}
	}
	headers := make(map[string]string)
	for key, values := range ctx.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}
