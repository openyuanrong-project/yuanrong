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

// Package reader provides ReadFile with timeConsumption
package reader

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/assert"
)

func TestReadFileWithTimeout(t *testing.T) {
	patch := gomonkey.ApplyFunc(ioutil.ReadFile, func(string) ([]byte, error) {
		return nil, nil
	})
	data, _ := ReadFileWithTimeout("/sn/home")
	assert.Nil(t, data)
	patch.Reset()
}

func TestReadFileInfoWithTimeout(t *testing.T) {
	patch := gomonkey.ApplyFunc(os.Stat, func(string) (os.FileInfo, error) {
		return nil, nil
	})
	fileInfo, _ := ReadFileInfoWithTimeout("/sn/home")
	assert.Nil(t, fileInfo)
	patch.Reset()
}

func TestPrintTimeout(t *testing.T) {
	stopCh := make(chan struct{})
	go printTimeOut(stopCh)
	time.Sleep(time.Second * 15)
	close(stopCh)
}
