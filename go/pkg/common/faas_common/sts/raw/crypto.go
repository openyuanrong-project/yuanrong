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

// Package raw use work key to encrypt and decrypt
package raw

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

const (
	defaultSaltSize = 12
)

// AesGCMDecrypt decrypt a cypher text using AES_GCM algorithm
func AesGCMDecrypt(secret, salt, cipherBytes []byte) ([]byte, error) {
	defer postRecover()
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}
	// salt 长度和 nonceSize 保持一致
	// cipher.NewGCM(block) 使用的是默认12字节的nonceSize,也代表盐值长度必须是12Byte;为了适应性强,我们使用自定义的 nonceSize
	gcm, err := cipher.NewGCMWithNonceSize(block, len(salt))
	if err != nil {
		return nil, err
	}
	plainBytes, err := gcm.Open(nil, salt, cipherBytes, nil)
	if err != nil {
		return nil, err
	}
	return plainBytes, nil
}

func postRecover() {
	var err error
	if r := recover(); r != nil {
		switch value := r.(type) {
		case string:
			err = fmt.Errorf("%s", value)
		case error:
			err = value
		default:
			err = fmt.Errorf("unexpect panic error: %w", err)
		}
		err = fmt.Errorf("panic error: %w", err)
	}
}

// AesGCMEncrypt will encrypt plainBytes to cipherBytes
func AesGCMEncrypt(secret, plainBytes []byte) ([]byte, []byte, error) {
	defer postRecover()
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, defaultSaltSize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed NewGCM: %w", err)
	}
	salt := make([]byte, gcm.NonceSize())
	_, err = rand.Read(salt)
	if err != nil {
		return nil, nil, err
	}
	cipherBytes := gcm.Seal(nil, salt, plainBytes, nil)
	return salt, cipherBytes, nil
}
