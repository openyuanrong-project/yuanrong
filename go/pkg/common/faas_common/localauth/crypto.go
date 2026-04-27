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

// Package localauth authenticates requests by local configmaps
package localauth

import (
	"sync"
)

var (
	algorithm     = "aeswithkey"
	once          sync.Once
)

func initCrypto() error {
	return nil
}

// Decrypt decrypts a cypher text using a certain algorithm
func Decrypt(src string) ([]byte, error) {
	var text []byte
	return text, nil
}

// Encrypt encrypts a cypher text using a certain algorithm
func Encrypt(src string) (string, error) {
	var ciperText string
	return ciperText, nil
}

// DecryptKeys decrypts a set of aKey and sKey
func DecryptKeys(inputAKey string, inputSKey string) ([]byte, []byte, error) {
	var aKey []byte
	var sKey []byte
	return aKey, sKey, nil
}
