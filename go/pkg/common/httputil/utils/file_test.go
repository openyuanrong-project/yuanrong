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

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	var cases = []struct {
		in       string // input
		expected bool   // expected result
	}{
		{"", false},
		{"./file.go", true},
		{"./notexists", false},
		{"/%$&*", false},
	}
	for _, c := range cases {
		actual := Exists(c.in)
		if actual != c.expected {
			t.Errorf("Exists(%s) = %v; expected %v", c.in, actual, c.expected)
		}
	}
}

func TestGetFileSize(t *testing.T) {
	ioutil.WriteFile("./test.txt", []byte("test"), 0666)
	var cases = []struct {
		in         string // input
		expectSize int64  // expected result
	}{
		{"./test.txt", 4},
		{"./test1.txt", 0},
	}
	for _, c := range cases {

		size := GetFileSize(c.in)
		assert.Equal(t, size, c.expectSize)
	}
	os.Remove("./test.txt")
}
