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

// Package common for token
package common

import (
	"sync"
	"time"

	"yuanrong.org/kernel/runtime/libruntime/common/logger/log"
)

const (
	updateTokenInterval = 60 * time.Second
)

// TokenMgr -
type TokenMgr struct {
	token string
	salt  string
	lock  sync.RWMutex

	callback func(ctx *AuthContext)
}

var tokenMgr = &TokenMgr{
	lock: sync.RWMutex{},
}

// SetCallback -
func (t *TokenMgr) SetCallback(f func(ctx *AuthContext)) {
	t.lock.Lock()
	t.callback = f

	if t.token != "" {
		f(&AuthContext{
			Token: t.token,
			Salt:  t.salt,
		})
	}
	t.lock.Unlock()
	log.GetLogger().Infof("set callback ok")
}

// GetTokenMgr -
func GetTokenMgr() *TokenMgr {
	return tokenMgr
}

// GetToken -
func (t *TokenMgr) GetToken() string {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.token
}

// UpdateToken -
func (t *TokenMgr) UpdateToken(auth *AuthContext) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if auth == nil || auth.Token == t.token || auth.Token == "" {
		return
	}
	log.GetLogger().Infof("recv token")
	t.token = auth.Token
	t.salt = auth.Salt
	if t.callback != nil {
		t.callback(auth)
	}
}

// UpdateTokenLoop -
func (t *TokenMgr) UpdateTokenLoop(stopCh <-chan struct{}) {
	ticker := time.NewTicker(updateTokenInterval)
	defer ticker.Stop()
	for {
		select {
		case _, ok := <-stopCh:
			if !ok {
				log.GetLogger().Errorf("updateToken is stopping")
				return
			}
		case <-ticker.C:
		}
	}
}
