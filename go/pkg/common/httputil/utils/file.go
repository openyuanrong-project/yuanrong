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
	"os"
	"path/filepath"

	"yuanrong.org/kernel/pkg/common/reader"
)

// Exists exists Whether the path exists
func Exists(path string) bool {
	if _, err := filepath.Abs(path); err != nil {
		return false
	}

	if _, err := reader.ReadFileInfoWithTimeout(path); err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}

	return true
}

// GetFileSize 获取文件大小
func GetFileSize(path string) int64 {
	if !Exists(path) {
		return 0
	}
	fileInfo, err := reader.ReadFileInfoWithTimeout(path)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}
