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

// Package utils for common functions
package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// GetResourcePath Get Resource Path
func GetResourcePath() string {
	return getPath("ResourcePath", "resource")
}

// GetServicesPath Get Services Path
func GetServicesPath() string {
	return getPath("ServicesPath", "service-config")
}

func getPath(env, defaultPath string) string {
	envPath := os.Getenv(env)
	if envPath == "" {
		var err error
		cliPath, err := exec.LookPath(os.Args[0])
		if err != nil {
			return envPath
		}
		envPath, err = filepath.Abs(filepath.Dir(cliPath))
		// do not return this error
		if err != nil {
			fmt.Printf("GetResourcePath abs filepath dir error")
		}
		envPath = strings.Replace(envPath, "\\", "/", -1)
		envPath = path.Join(path.Dir(envPath), defaultPath)
	} else {
		envPath = strings.Replace(envPath, "\\", "/", -1)
	}

	return envPath
}
