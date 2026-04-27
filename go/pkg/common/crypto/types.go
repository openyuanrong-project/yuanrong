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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sync"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
	"yuanrong.org/kernel/pkg/common/reader"
)

const (
	// byteSize defines the key length
	byteSize          = 32
	queryElementLem   = 2
	saltDataMinLength = 16
	// IterKeyFactoryIter is the iter Count of Root Key Factor
	IterKeyFactoryIter = 10000
	apple              = "apple"
	boy                = "boy"
	cat                = "cat"
	dog                = "dog"
	egg                = "egg"
	fish               = "fish"
	wdo                = "wdo"
	KeyFactorNums      = 5
)

var (
	rootKeyOnce sync.Once
	workKeyOnce sync.Once

	workKey []byte
	rootKey *RootKey
)

// set root key factor
func buildRootKeyFactor(f *RootKeyFactor) error {
	resourcePath := utils.GetResourcePath()
	k1Path := path.Join(resourcePath, "rdo", "v1", apple, "a.txt")
	k2Path := path.Join(resourcePath, "rdo", "v1", boy, "b.txt")
	macPath := path.Join(resourcePath, "rdo", "v1", cat, "c.txt")
	saltPath := path.Join(resourcePath, "rdo", "v1", dog, "d.txt")

	// k1Data
	k1Data, err := reader.ReadFileWithTimeout(k1Path)
	if err != nil {
		return err
	}
	f.k1Data, err = hex.DecodeString(string(k1Data))
	if err != nil {
		return err
	}

	// k2Data
	k2Data, err := reader.ReadFileWithTimeout(k2Path)
	if err != nil {
		return err
	}
	f.k2Data, err = hex.DecodeString(string(k2Data))
	if err != nil {
		return err
	}

	// macData
	macData, err := reader.ReadFileWithTimeout(macPath)
	if err != nil {
		return err
	}
	f.macData, err = hex.DecodeString(string(macData))
	if err != nil {
		return err
	}

	// saltData
	saltData, err := reader.ReadFileWithTimeout(saltPath)
	if len(saltData) < saltDataMinLength {
		return fmt.Errorf("invalid salt data length of %d", len(saltData))
	}
	if err != nil {
		return err
	}
	if f.saltData, err = hex.DecodeString(string(saltData)); err != nil {
		return err
	}
	return nil
}

// LoadRootKey Load Root Key
func LoadRootKey() (*RootKey, error) {
	// k3
	resourcePath := utils.GetResourcePath()
	k3Path := path.Join(resourcePath, "rdo", "v1", egg, "e.txt")
	// k1Data
	k3Data, err := reader.ReadFileWithTimeout(k3Path)
	k3DataDecode, err := hex.DecodeString(string(k3Data))
	if err != nil {
		return nil, err
	}
	f := &RootKeyFactor{
		// 10000 is the iter Count of Root Key Factor
		iterCount:      IterKeyFactoryIter,
		component3:     string(k3DataDecode),
		component3byte: k3DataDecode,
	}
	err = buildRootKeyFactor(f)
	if err != nil {
		return nil, err
	}
	rootKey := encryptPBKDF2WithSHA256(f)
	return rootKey, nil
}

// LoadRootKeyWithKeyFactor Load Root Key With Key Factor
func LoadRootKeyWithKeyFactor(keyFactor []string) (*RootKey, error) {
	if len(keyFactor) < KeyFactorNums {
		return nil, errors.New("short key factors")
	}
	var err error
	k3Data := keyFactor[2]
	k3DataDecode, err := hex.DecodeString(k3Data)
	f := &RootKeyFactor{
		// 10000 is the iter Count of Root Key Factor
		iterCount:      IterKeyFactoryIter,
		component3:     string(k3DataDecode),
		component3byte: k3DataDecode,
	}
	f.k1Data, err = hex.DecodeString(keyFactor[0])
	if err != nil {
		return nil, err
	}
	f.k2Data, err = hex.DecodeString(keyFactor[1])
	if err != nil {
		return nil, err
	}
	f.macData, err = hex.DecodeString(keyFactor[3])
	if err != nil {
		return nil, err
	}
	if f.saltData, err = hex.DecodeString(keyFactor[4]); err != nil {
		return nil, err
	}
	rootKey := encryptPBKDF2WithSHA256(f)
	return rootKey, nil
}

// RootKey include RootKey and MacSecretKey
type RootKey struct {
	RootKey      []byte
	MacSecretKey []byte
}

// RootKeyFactor include Root Key Factor
type RootKeyFactor struct {
	k1Data         []byte
	k2Data         []byte
	macData        []byte
	saltData       []byte
	iterCount      int
	component3     string
	component3byte []byte
}

// WorkKeys define Work Keys
type WorkKeys map[string]*SecretNamedWorkKeys

// GetKeyByName Get Key By Name
func (k *WorkKeys) GetKeyByName(name string) *SecretWorkKey {
	namedKey, exist := (*k)[name]
	if !exist {
		return nil
	}

	return namedKey.Keys
}

// SecretNamedWorkKeys include Keys and Description
type SecretNamedWorkKeys struct {
	Keys        *SecretWorkKey `json:"keys"`
	Description string         `json:"description"`
}

// SecretWorkKey include Key and Mac
type SecretWorkKey struct {
	Key string `json:"key"`
	Mac string `json:"mac"`
}

// MarshalJSON Marshal JSON
func (s *SecretWorkKey) MarshalJSON() ([]byte, error) {
	if rootKey == nil || rootKey.RootKey == nil || rootKey.MacSecretKey == nil {
		return nil, fmt.Errorf("rootKey is nil")
	}

	key, err := Encrypt(s.Key, []byte(hex.EncodeToString(rootKey.RootKey)))
	if err != nil {
		return nil, err
	}
	mac := hmacHash([]byte(s.Key), rootKey.MacSecretKey)

	type SecretWorkKeyJSON SecretWorkKey

	return json.Marshal(SecretWorkKeyJSON(SecretWorkKey{
		Key: string(key), Mac: mac}))
}

// UnmarshalJSON Unmarshal JSON
func (s *SecretWorkKey) UnmarshalJSON(data []byte) error {

	type SecretWorkKeyJSON SecretWorkKey

	err := json.Unmarshal(data, (*SecretWorkKeyJSON)(s))
	if err != nil {
		return err
	}

	key, err := decryptWorkKey(s.Key, s.Mac, rootKey)
	if err != nil {
		return err
	}

	s.Key = key

	return nil
}

// Signature define Signature
type Signature struct {
	Method      []byte
	Path        []byte
	QueryStr    string
	Body        []byte
	AppID       []byte
	CurTimeTamp []byte
}

// GetRootKey Get Root Key
func GetRootKey() []byte {
	rootKeyOnce.Do(func() {
		rk, err := LoadRootKey()
		if err != nil {
			log.GetLogger().Errorf("failed to load rootKey, err: %s", err.Error())
			return
		}
		rootKey = rk
	})

	if rootKey == nil {
		log.GetLogger().Errorf("root key is nil")
		return []byte{}
	}
	return []byte(hex.EncodeToString(rootKey.RootKey))
}

// LoadWorkKey Load work Key
func LoadWorkKey() ([]byte, error) {
	resourcePath := utils.GetResourcePath()
	workKeyPath := path.Join(resourcePath, "rdo", "v1", fish, "f.txt")
	workKey, err := reader.ReadFileWithTimeout(workKeyPath)
	return workKey, err
}

// GetWorkKey Get Work Key
func GetWorkKey() []byte {
	workKeyOnce.Do(func() {
		wk, err := LoadWorkKey()
		if err != nil {
			log.GetLogger().Errorf("failed to load workKey, err: %s", err.Error())
			return
		}
		workKey = wk
	})

	if workKey == nil {
		log.GetLogger().Errorf("work key is nil")
		return []byte{}
	}
	return workKey
}
