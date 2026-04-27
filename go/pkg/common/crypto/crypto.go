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

// Package crypto for auth
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"

	"golang.org/x/crypto/pbkdf2"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

const (
	randomNumberMaxLength = 16
	randomNumberMinLength = 12
	defaultSliceLen       = 1024
	cipherTextsLen        = 2
)

var (
	decryptAlgorithm      string = "GCM"
	decryptAlgorithmMutex sync.RWMutex
)

// SetDecryptAlgorithm -
func SetDecryptAlgorithm(algorithm string) {
	decryptAlgorithmMutex.Lock()
	defer decryptAlgorithmMutex.Unlock()
	decryptAlgorithm = algorithm
}

// GetDecryptAlgorithm returns global decryptAlgorithm
func GetDecryptAlgorithm() string {
	decryptAlgorithmMutex.RLock()
	defer decryptAlgorithmMutex.RUnlock()
	return decryptAlgorithm
}

// Encrypt encrypts data by GCM algorithm
func Encrypt(content string, secret []byte) ([]byte, error) {
	if GetDecryptAlgorithm() == "NO_CRYPTO" {
		log.GetLogger().Debug("decrypt algorithm is NO_CRYPTO, return plain text directly")
		return []byte(content), nil
	}
	textByte := []byte(content)
	cipherByte, salt, err := encryptGcmDataFromBody(textByte, secret)
	if err != nil {
		return nil, err
	}
	ciperText := fmt.Sprintf("%s:%s", salt, hex.EncodeToString(cipherByte))
	return []byte(ciperText), nil
}

func encryptPBKDF2WithSHA256(f *RootKeyFactor) *RootKey {
	minLen := math.Min(float64(len(f.k1Data)), math.Min(float64(len(f.k2Data)), float64(len(f.component3byte))))
	bytePsd := make([]byte, int(minLen), int(minLen))

	for i := 0; i < int(minLen); i++ {
		bytePsd[i] = f.k1Data[i] ^ f.k2Data[i] ^ f.component3[i]
	}

	rootKeyByte := pbkdf2.Key(bytePsd, f.saltData, f.iterCount, byteSize, sha256.New)
	sliceLen := len(rootKeyByte)
	if sliceLen <= 0 || sliceLen > defaultSliceLen {
		sliceLen = defaultSliceLen
	}

	byteMac := make([]byte, sliceLen)
	macSecretKeyByte := pbkdf2.Key(byteMac, f.macData, f.iterCount, byteSize, sha256.New)

	rootKey := &RootKey{}
	rootKey.RootKey = rootKeyByte
	rootKey.MacSecretKey = macSecretKeyByte

	return rootKey
}

func hmacHash(data []byte, key []byte) string {
	hm := hmac.New(sha256.New, key)
	_, err := hm.Write(data)
	if err != nil {
		log.GetLogger().Errorf("failed to hmacHash write data: %s ", err.Error())
		return ""
	}
	return hex.EncodeToString(hm.Sum([]byte{}))
}

// encryptGcmDataFromBody encrypts data
func encryptGcmDataFromBody(body []byte, secret []byte) ([]byte, string, error) {
	if len(body) == 0 {
		return nil, "", fmt.Errorf("body is empty")
	}
	secretBytes, err := hex.DecodeString(string(secret))
	if err != nil {
		return nil, "", err
	}

	aesBlock, err := aes.NewCipher(secretBytes)
	if err != nil {
		return nil, "", err
	}
	aesgcm, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, "", err
	}

	// generate salt value
	nonceSize := aesgcm.NonceSize()
	if nonceSize > randomNumberMaxLength || nonceSize < randomNumberMinLength {
		err = errors.New("nonceSize out of bound")
		return nil, "", err
	}
	salt := make([]byte, nonceSize, nonceSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, "", err
	}

	cipherByte := aesgcm.Seal(nil, salt, body, nil)
	return cipherByte, hex.EncodeToString(salt), nil
}

// Decrypt returns string cipher bytes by AES and GCM algorithms
func Decrypt(cipherText []byte, secret []byte) (string, error) {
	if GetDecryptAlgorithm() == "NO_CRYPTO" {
		log.GetLogger().Debug("decrypt algorithm is NO_CRYPTO, return plain text directly")
		return string(cipherText), nil
	}
	cipherTexts := strings.Split(string(cipherText), ":")
	if len(cipherTexts) != cipherTextsLen {
		return "", fmt.Errorf("wrong cipher text")
	}

	saltStr := cipherTexts[0]
	encryptStr := cipherTexts[1]

	salt, err := hex.DecodeString(saltStr)
	if err != nil {
		return "", err
	}

	encrypt, err := hex.DecodeString(encryptStr)
	if err != nil {
		return "", err
	}

	secretData := secret
	if utils.IsHexString(string(secret)) {
		var err error
		secretData, err = hex.DecodeString(string(secret))
		if err != nil {
			return "", err
		}
	}

	cipherBytes, err := decryptGcmData(encrypt, secretData, salt)
	if err != nil {
		return "", err
	}

	if cipherBytes == nil {
		return "", fmt.Errorf("decrypt error")
	}

	return string(cipherBytes), nil
}

// DecryptByte returns string cipher bytes by AES and GCM algorithms
func DecryptByte(cipherText []byte, secret []byte) ([]byte, error) {
	cipherTexts := strings.Split(string(cipherText), ":")
	if len(cipherTexts) != cipherTextsLen {
		return nil, fmt.Errorf("wrong cipher text")
	}

	saltStr := cipherTexts[0]
	encryptStr := cipherTexts[1]
	salt, err := hex.DecodeString(saltStr)
	if err != nil {
		return nil, err
	}

	encryptByte, err := hex.DecodeString(encryptStr)
	if err != nil {
		return nil, err
	}

	secretData := secret
	if utils.IsHexString(string(secret)) {
		var err error
		secretData, err = hex.DecodeString(string(secret))
		if err != nil {
			return nil, err
		}
	}

	cipherBytes, err := decryptGcmData(encryptByte, secretData, salt)
	if err != nil {
		return nil, err
	}

	if cipherBytes == nil {
		return nil, fmt.Errorf("decrypt error")
	}

	return cipherBytes, nil
}

// decryptGcmData decrypt data with aes gcm mode
func decryptGcmData(encrypt []byte, secret []byte, salt []byte) ([]byte, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}

	aesGcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	decrypted, err := aesGcm.Open(nil, salt, encrypt, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

// decryptWorkKey Decrypt Work Key
func decryptWorkKey(workKey string, workMac string, rootKey *RootKey) (string, error) {
	workKeyDecrypt, err := Decrypt([]byte(workKey), rootKey.RootKey)
	if err != nil {
		return "", err
	}

	workKeyMac := hmacHash([]byte(workKeyDecrypt), rootKey.MacSecretKey)
	if workKeyMac == workMac {
		return workKeyDecrypt, nil
	}

	return "", fmt.Errorf("workKey is changed")
}
