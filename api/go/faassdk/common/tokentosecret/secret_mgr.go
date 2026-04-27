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

// Package tokentosecret
package tokentosecret

import (
	"crypto/sha512"
	"sync"

	"golang.org/x/crypto/pbkdf2"

	"yuanrong.org/kernel/runtime/libruntime/common"
)

// SecretMgr -
type SecretMgr struct {
	token string
	salt  string
	//	expireTime int
	sk   []byte
	lock sync.RWMutex
}

var mgr = &SecretMgr{
	lock: sync.RWMutex{},
}

// GetSecretMgr -
func GetSecretMgr() *SecretMgr {
	return mgr
}

// SetAuthContext -
func (s *SecretMgr) SetAuthContext(auth *common.AuthContext) {
	s.lock.Lock()
	if auth == nil || auth.Token == s.token || auth.Token == "" {
		s.lock.Unlock()
		return
	}

	s.token = auth.Token
	s.salt = auth.Salt
	s.lock.Unlock()
	s.generateSk(s.token, s.salt)
}

// GetSk -
func (s *SecretMgr) GetSk() []byte {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.sk
}

func (s *SecretMgr) generateSk(token string, salt string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.sk = pbkdf2.Key([]byte(token), []byte(salt), 25000, sha512.Size, sha512.New) // 25000 iter
}
