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
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRSS(t *testing.T) {
	r := bytes.NewReader([]byte("367597 12113 4058 3810 0 47257 0\n"))
	buffer := make([]byte, KB)

	for i := 0; i < 10; i++ {
		rss, err := parseRSS(r, buffer)
		assert.Nil(t, err, "parseRSS should return no error")
		assert.Equal(t, uint64(12113*os.Getpagesize()), rss)
	}

	r = bytes.NewReader([]byte("123"))
	_, err := parseRSS(r, make([]byte, KB))
	assert.Error(t, err, "parseRSS should failed")

	r = bytes.NewReader([]byte("123 abcde 132"))
	_, err = parseRSS(r, make([]byte, KB))
	assert.Error(t, err, "parseRSS should failed")
}
