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
	"errors"
	"path/filepath"
	"strings"
)

// ValidateFilePath verify the legitimacy of the file path
func ValidateFilePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil || !strings.HasPrefix(path, absPath) {
		return errors.New("invalid file path, expect to be configured as an absolute path")
	}
	return nil
}

// GetFuncNameFromFuncLibPath parses function name out of FUNCTION_LIB_PATH
func GetFuncNameFromFuncLibPath(funcLibPath string) string {
	if funcLibPath == "" {
		return ""
	}
	funcLibPathSplits := strings.Split(funcLibPath, "/")
	return funcLibPathSplits[len(funcLibPathSplits)-1]
}
