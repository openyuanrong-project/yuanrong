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
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// MaxReadFileTime elapsed time allowed to read config file from disk
const MaxReadFileTime = 10

// ReadFileWithTimeout is to ReadFile and count timeConsumption at same time
func ReadFileWithTimeout(configFile string) ([]byte, error) {
	stopCh := make(chan struct{})
	go printTimeOut(stopCh)
	data, err := ioutil.ReadFile(configFile)
	close(stopCh)
	return data, err
}

// ReadFileInfoWithTimeout is to Read FileInfo and count timeConsumption at same time
func ReadFileInfoWithTimeout(filePath string) (os.FileInfo, error) {
	stopCh := make(chan struct{})
	go printTimeOut(stopCh)
	fileInfo, err := os.Stat(filePath)
	close(stopCh)
	return fileInfo, err
}

// printTimeOut print error info every 10s after timeout
func printTimeOut(stopCh <-chan struct{}) {
	if stopCh == nil {
		os.Exit(0)
		return
	}
	timer := time.NewTicker(time.Second * MaxReadFileTime)
	count := 0
	for {
		<-timer.C
		select {
		case _, ok := <-stopCh:
			if !ok {
				timer.Stop()
				return
			}
		default:
			count += MaxReadFileTime
			fmt.Printf("ReadFile Timeout: elapsed time %ds\n", count)
		}
	}
}
