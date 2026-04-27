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
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/client-go/tools/cache"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/signals"
)

const (
	senderCacheDuration = 1 * time.Minute
)

// AuthCache cache interface
type AuthCache interface {
	GetSignForSender() (string, string, error)
	GetSignForReceiver(auth string) (string, bool, error)
	updateReceiver(string) error
	updateSender(string, string)
}

type authCache struct {
	// use an atomic value to promise concurrent safety, which stores the authorization token and time.
	senderCache *atomic.Value
	// sign-time
	receiverCache cache.Store
	appID         string
	AuthConfig
}

type senderValue struct {
	auth string
	time string
}

var localCache *authCache
var doOnce sync.Once

// GetLocalAuthCache you have to create it before you get it.
func GetLocalAuthCache(aKey, sKey, appID string, duration int) AuthCache {
	doOnce.Do(func() {
		var c cache.Store
		stopCh := signals.WaitForSignal()
		// cache.Store ttl the minimum valid value is 1 second. If this parameter is set to 0,
		// the cache does not need to be increased. Therefore, set the cache to nil.
		if duration == 0 {
			c = nil
		} else {
			c = cache.NewTTLStore(receiverCacheKey, time.Duration(duration)*time.Minute)
		}
		atom := &atomic.Value{}
		localCache = &authCache{
			senderCache:   atom,
			receiverCache: c,
		}
		localCache.appID = appID
		localCache.AKey = aKey
		localCache.SKey = sKey
		localCache.initSenderCache(stopCh)
		localCache.Duration = duration
		// clean expired keys by ticker could avoid worker-manager oom problem
		// because receiver cache clean expired keys is lazy by calling GetByKeys method or List method
		go localCache.startCleanExpiredKeysByTicker(stopCh, time.Duration(duration)*time.Minute)
	})
	if localCache == nil {
		return nil
	}
	return localCache
}

func (c *authCache) startCleanExpiredKeysByTicker(stopCh <-chan struct{}, duration time.Duration) {
	if stopCh == nil || c.receiverCache == nil {
		return
	}
	log.GetLogger().Infof("start to clean expired keys by ticker duration %s", duration.String())
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// call receiver cache list method will clean all expired keys
			length := len(c.receiverCache.List())
			log.GetLogger().Debugf("receiver cache length is %d after clean expired keys once by ticker", length)
		case <-stopCh:
			log.GetLogger().Infof("stop channel is closed")
			return
		}
	}
}

// GetSignForSender return time auth error
func (c *authCache) GetSignForSender() (string, string, error) {
	loaded := c.senderCache.Load()
	value, ok := loaded.(senderValue)
	if !ok {
		return "", "", errors.New("no sender cache")
	}
	if value.time == "" || value.auth == "" {
		return "", "", errors.New("no sender time")
	}
	return value.time, value.auth, nil
}

// GetSignForReceiver value exit error
func (c *authCache) GetSignForReceiver(auth string) (string, bool, error) {
	if c.receiverCache == nil {
		return "", false, nil
	}
	key, b, err := c.receiverCache.GetByKey(auth)
	if !b {
		key = ""
	}
	return key.(string), b, err
}

func (c *authCache) updateReceiver(sign string) error {
	if c.receiverCache == nil {
		return nil
	}
	err := c.receiverCache.Add(sign)
	if err != nil {
		return err
	}
	return nil
}

func (c *authCache) updateSender(auth, time string) {
	c.senderCache.Store(senderValue{
		auth: auth,
		time: time,
	})
}

func (c *authCache) waitForDoneSignal(stopCh <-chan struct{}) {
	if stopCh == nil {
		return
	}
	ticker := time.NewTicker(senderCacheDuration)
	for {
		select {
		case <-ticker.C:
			// update senderCache
			c.createAndUpdateSender()
		case <-stopCh:
			ticker.Stop()
			return
		}
	}
}

func (c *authCache) initSenderCache(stopCh <-chan struct{}) {
	c.createAndUpdateSender()
	go c.waitForDoneSignal(stopCh)
}

func (c *authCache) createAndUpdateSender() {
	var data []byte
	authorization, t := CreateAuthorization(
		c.AKey,
		c.SKey,
		"",
		c.appID,
		data,
	)
	c.updateSender(authorization, t)
	if c.Duration != 0 {
		log.GetLogger().Debugf("the length of receiver cache is: %d", len(c.receiverCache.ListKeys()))
	}
}

func receiverCacheKey(obj interface{}) (string, error) {
	return obj.(string), nil
}
