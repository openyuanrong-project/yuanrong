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

package localauth

import (
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/cache"
)

func Test_receiverCacheKey(t *testing.T) {
	type args struct {
		obj interface{}
	}
	var a args
	a.obj = "aaa"
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"case1", a, "aaa", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := receiverCacheKey(tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("receiverCacheKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("receiverCacheKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_authCache_GetSignForSender(t *testing.T) {
	type fields struct {
		senderCache   *atomic.Value
		receiverCache cache.Store
		appID         string
		AuthConfig    AuthConfig
	}
	var f fields
	senderCache := &atomic.Value{}
	f.senderCache = senderCache

	var f2 fields
	senderCache2 := &atomic.Value{}
	senderCache2.Store(senderValue{
		time: "aaa",
		auth: "aaa",
	})
	f2.senderCache = senderCache2
	tests := []struct {
		name    string
		fields  fields
		wantS   string
		wantD   string
		wantErr bool
	}{
		{"case1", f, "", "", true},
		{"case2", f2, "aaa", "aaa", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &authCache{
				senderCache:   tt.fields.senderCache,
				receiverCache: tt.fields.receiverCache,
				appID:         tt.fields.appID,
				AuthConfig:    tt.fields.AuthConfig,
			}
			gotS, gotD, err := c.GetSignForSender()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSignForSender() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotS != tt.wantS {
				t.Errorf("GetSignForSender() gotS = %v, want %v", gotS, tt.wantS)
			}
			if gotD != tt.wantD {
				t.Errorf("GetSignForSender() gotD = %v, want %v", gotD, tt.wantD)
			}
		})
	}
}

func Test_authCache_updateReceiver(t *testing.T) {
	type fields struct {
		senderCache   *atomic.Value
		receiverCache cache.Store
		appID         string
		AuthConfig    AuthConfig
	}
	type args struct {
		sign string
	}
	var f fields
	var a args
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"case1", f, a, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &authCache{
				senderCache:   tt.fields.senderCache,
				receiverCache: tt.fields.receiverCache,
				appID:         tt.fields.appID,
				AuthConfig:    tt.fields.AuthConfig,
			}
			if err := c.updateReceiver(tt.args.sign); (err != nil) != tt.wantErr {
				t.Errorf("updateReceiver() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_authCache_createAndUpdateSender(t *testing.T) {
	receiverCache := cache.NewTTLStore(receiverCacheKey, time.Duration(5)*time.Second)
	patches := [...]*gomonkey.Patches{
		gomonkey.ApplyMethod(reflect.TypeOf(receiverCache), "ListKeys",
			func(_ *cache.ExpirationCache) []string {
				return []string{}
			}),
		gomonkey.ApplyFunc(DecryptKeys, func(inputAKey string, inputSKey string) ([]byte, []byte, error) {
			return []byte("aaa"), []byte("aaa"), nil
		}),
	}

	defer func() {
		for idx := range patches {
			patches[idx].Reset()
		}
	}()
	type fields struct {
		senderCache   *atomic.Value
		receiverCache cache.Store
		appID         string
		AuthConfig    AuthConfig
	}
	var f fields
	f.AuthConfig.Duration = 1
	f.receiverCache = receiverCache
	f.senderCache = &atomic.Value{}
	f.senderCache.Store(senderValue{
		time: "aaa",
		auth: "aaa",
	})
	tests := []struct {
		name   string
		fields fields
	}{
		{"case1", f},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &authCache{
				senderCache:   tt.fields.senderCache,
				receiverCache: tt.fields.receiverCache,
				appID:         tt.fields.appID,
				AuthConfig:    tt.fields.AuthConfig,
			}
			c.createAndUpdateSender()
		})
	}
}

func TestWaitForDoneSignal(t *testing.T) {
	c := &authCache{
		senderCache: &atomic.Value{},
	}
	c.senderCache.Store(senderValue{
		time: "aaa",
		auth: "bbb",
	})
	c.waitForDoneSignal(nil)
	stopChan := make(chan struct{})
	go c.waitForDoneSignal(stopChan)
	close(stopChan)
	assert.NotEqual(t, c, nil)
}

func TestGetSignForReceiver(t *testing.T) {
	c := &authCache{
		senderCache:   &atomic.Value{},
		receiverCache: cache.NewTTLStore(receiverCacheKey, time.Duration(1)*time.Minute),
	}
	c.senderCache.Store(senderValue{
		time: "aaa",
		auth: "aaa",
	})
	c.updateReceiver("sign")
	_, _, err := c.GetSignForReceiver("auth")
	assert.Equal(t, err, nil)
}
