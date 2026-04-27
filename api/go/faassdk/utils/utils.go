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

// Package utils
package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"yuanrong.org/kernel/runtime/libruntime/common/uuid"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

const (
	minPathPartLength = 2
)

var (
	once     sync.Once
	serverIP = ""
)

// UniqueID get unique ID
func UniqueID() string {
	return uuid.New().String()
}

// BytesToString convert []byte to string without memory alloc
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// GetServerIP -
func GetServerIP() (string, error) {
	var err error
	once.Do(func() {
		addr, errMsg := GetHostAddr()
		if errMsg != nil {
			err = errMsg
			return
		}
		serverIP = addr[0]
	})
	return serverIP, err
}

// GetHostAddr -
func GetHostAddr() ([]string, error) {
	name, err := os.Hostname()
	if err != nil {
		logger.GetLogger().Errorf("get hostname failed: %v", err)
		return nil, err
	}

	addrs, err := net.LookupHost(name)
	if err != nil || len(addrs) == 0 {
		logger.GetLogger().Errorf("look up host by name failed")
		return nil, fmt.Errorf("look up host by name failed")
	}
	return addrs, nil
}

// GetLibInfo will parse handler info, example:
// libPath="/tmp" libName="example.init"      --> fileName="/tmp/example.so" handlerName="init"
// libPath="/tmp" libName="test.example.init" --> fileName="/tmp/test/example.so" handlerName="init"
func GetLibInfo(libPath, libName string) (string, string) {
	path := libPath
	parts := strings.Split(libName, ".")
	length := len(parts)
	handlerName := parts[length-1]
	if length < minPathPartLength {
		return "", ""
	} else if length > minPathPartLength {
		tmpPath := filepath.Join(parts[:length-minPathPartLength]...)
		path = filepath.Join(path, tmpPath)
	}
	fileName := filepath.Join(path, parts[length-minPathPartLength]+".so")
	return fileName, handlerName
}

// DealEnv deal with environment and encrypted_user_data
func DealEnv() (error, map[string]string) {
	var err error
	delegateDecryptMap := make(map[string]string)
	environmentMap := make(map[string]string)
	rtUserDataMap := make(map[string]string)
	envMap := make(map[string]string)
	delegateDecrypt := os.Getenv("ENV_DELEGATE_DECRYPT")
	if delegateDecrypt == "" {
		return nil, nil
	}
	if err = json.Unmarshal([]byte(delegateDecrypt), &delegateDecryptMap); err != nil {
		return err, nil
	}
	environment, ok := delegateDecryptMap["environment"]
	if ok && environment != "" {
		if err = json.Unmarshal([]byte(environment), &environmentMap); err != nil {
			return err, nil
		}
	}
	for key, value := range environmentMap {
		if key != constants.LDLibraryPath {
			err = os.Setenv(key, value)
		}
		envMap[key] = value
	}
	if err != nil {
		return err, nil
	}
	rtUserData, ok := delegateDecryptMap["encrypted_user_data"]
	if ok && rtUserData != "" {
		if err = json.Unmarshal([]byte(rtUserData), &rtUserDataMap); err != nil {
			return err, nil
		}
	}

	for key, value := range rtUserDataMap {
		envMap[key] = value
	}
	return err, envMap
}

// ContainsConnRefusedErr -
func ContainsConnRefusedErr(err error) bool {
	const connRefusedStr = "connection refused"
	return strings.Contains(err.Error(), connRefusedStr)
}
