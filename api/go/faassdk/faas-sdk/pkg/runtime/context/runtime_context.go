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

// Package context
package context

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	defaultTimeout = "3"
)

func atoi(input string) int {
	result, err := strconv.Atoi(input)
	if err != nil {
		log.Println("execute strconv.atoi failed")
	}
	return result
}

func getCurrentTime() int {
	return (int)(time.Now().UnixNano() / int64(time.Millisecond))
}

// InitializeContext initialize context from env
func (p *Provider) InitializeContext() error {
	if err := p.CtxEnv.initializeContextEnv(); err != nil {
		return err
	}

	return p.CtxHTTPHead.initializeInvokeContext()
}

func (ic *InvokeContext) initializeInvokeContext() error {
	delegateDecrypt := os.Getenv("ENV_DELEGATE_DECRYPT")
	if delegateDecrypt == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(delegateDecrypt), ic); err != nil {
		return fmt.Errorf("initializeContext failed to Unmarshal ENV_DELEGATE_DECRYPT, error: %s", err)
	}
	return nil
}

func (e *Env) initializeContextEnv() error {
	e.rtStartTime = getCurrentTime()
	timeout := os.Getenv("RUNTIME_TIMEOUT")
	if timeout == "" {
		timeout = defaultTimeout
	}
	e.rtTimeout = atoi(timeout)
	rtProjectID := os.Getenv("RUNTIME_PROJECT_ID")
	if rtProjectID != "" {
		e.rtProjectID = rtProjectID
	}
	rtPackage := os.Getenv("RUNTIME_PACKAGE")
	if rtPackage != "" {
		e.rtPackage = rtPackage
	}
	rtFcName := os.Getenv("RUNTIME_FUNC_NAME")
	if rtFcName != "" {
		e.rtFcName = rtFcName
	}
	rtFcVersion := os.Getenv("RUNTIME_FUNC_VERSION")
	if rtFcVersion != "" {
		e.rtFcVersion = rtFcVersion
	}
	rtMemory := os.Getenv("RUNTIME_MEMORY")
	if rtMemory != "" {
		e.rtMemory = atoi(rtMemory)
	}
	rtCPU := os.Getenv("RUNTIME_CPU")
	if rtCPU != "" {
		e.rtCPU = atoi(rtCPU)
	}
	rtUserData := os.Getenv("RUNTIME_USERDATA")
	if rtUserData != "" {
		err := json.Unmarshal([]byte(rtUserData), &e.rtUserData)
		if err != nil {
			return fmt.Errorf("initializeContext failed to Unmarshal Userdata, error: %s", err)
		}
	}
	return nil
}
