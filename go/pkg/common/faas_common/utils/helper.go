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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	envPathSeparators = ":"
)

// IsFile returns true if the path is a file
func IsFile(path string) bool {
	file, err := os.Stat(path)
	if err != nil {
		return false
	}
	return file.Mode().IsRegular()
}

// IsDir returns true if the path is a dir
func IsDir(path string) bool {
	dir, err := os.Stat(path)
	if err != nil {
		return false
	}

	return dir.IsDir()
}

// FileExists returns true if the path exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

// IsHexString judge If Hex String
func IsHexString(str string) bool {

	str = strings.ToLower(str)

	for _, c := range str {
		if c < '0' || (c > '9' && c < 'a') || c > 'f' {
			return false
		}
	}

	return true
}

// ValidateFilePath verify the legitimacy of the file path
func ValidateFilePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil || !strings.HasPrefix(path, absPath) {
		return errors.New("invalid file path, expect to be configured as an absolute path")
	}
	return nil
}

// ValidEnvValuePath verify the legitimacy of the env path
func ValidEnvValuePath(envValues string) error {
	if envValues == "" {
		return nil
	}
	envByte := strings.Split(envValues, envPathSeparators)
	for _, envValue := range envByte {
		if err := ValidateFilePath(envValue); err != nil {
			return err
		}
	}
	return nil
}

// copyFile copies a single file from src to dst
func copyFile(srcPath, dstPath string) error {
	var err error
	var fromFd *os.File
	var toFd *os.File
	var fromFdInfo os.FileInfo

	if fromFd, err = os.Open(srcPath); err != nil {
		return err
	}
	defer func(fromFd *os.File) {
		if fromFd != nil {
			err = fromFd.Close()
		}
	}(fromFd)

	if fromFdInfo, err = os.Stat(srcPath); err != nil {
		return err
	}

	toFd, err = os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fromFdInfo.Mode())
	defer func(toFd *os.File) {
		if toFd != nil {
			err = toFd.Close()
		}
	}(toFd)

	if err != nil {
		return err
	}

	if _, err = io.Copy(toFd, fromFd); err != nil {
		return err
	}

	return err
}

// CopyDir copies a whole directory recursively
func CopyDir(srcPath string, dstPath string) error {
	var err error
	var dirFds []os.FileInfo
	var fromInfo os.FileInfo

	if fromInfo, err = os.Stat(srcPath); err != nil {
		return err
	}

	if err = os.MkdirAll(dstPath, fromInfo.Mode()); err != nil {
		return err
	}

	if dirFds, err = ioutil.ReadDir(srcPath); err != nil {
		return err
	}
	for _, fd := range dirFds {
		fromPath := path.Join(srcPath, fd.Name())
		toPath := path.Join(dstPath, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(fromPath, toPath); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyFile(fromPath, toPath); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
