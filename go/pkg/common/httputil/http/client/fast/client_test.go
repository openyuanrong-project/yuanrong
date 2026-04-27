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

package fast

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"testing"

	"yuanrong.org/kernel/pkg/common/httputil/http"
)

func Test_parseFastResponse(t *testing.T) {
	response1 := &fasthttp.Response{}
	badResponse := map[string]interface{}{
		"code":    0,
		"message": "500 error",
	}
	bytes, _ := json.Marshal(badResponse)
	response1.SetStatusCode(fasthttp.StatusInternalServerError)
	response1.SetBody(bytes)

	response2 := &fasthttp.Response{}

	response3 := &fasthttp.Response{}
	response3.SetStatusCode(fasthttp.StatusBadRequest)

	tests := []struct {
		name     string
		response *fasthttp.Response
		want     *http.SuccessResponse
		wantErr  bool
	}{
		{
			name:     "test 500",
			response: response1,
			wantErr:  false,
		},
		{
			name:     "test 200",
			response: response2,
			wantErr:  false,
		},
		{
			name:     "test 400",
			response: response3,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFastResponse(tt.response)
			assert.Equal(t, err != nil, tt.wantErr)
		})
	}
}
