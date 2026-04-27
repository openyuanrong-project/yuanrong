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

package loadbalance

import "sync"

type hashCache struct {
	hashes sync.Map
}

func createHashCache() *hashCache {
	return &hashCache{
		hashes: sync.Map{},
	}
}

func (cache *hashCache) getHash(key string) uint32 {
	hashIf, ok := cache.hashes.Load(key)
	if ok {
		hash, ok := hashIf.(uint32)
		if ok {
			return hash
		}
		return 0
	}
	hash := getHashKeyCRC32([]byte(key))
	cache.hashes.Store(key, hash)
	return hash
}
