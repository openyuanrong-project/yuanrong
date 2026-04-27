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

package logmanager

import (
	"sync"
)

// LogIndex is an index to speed up query (by some key)
type LogIndex struct {
	indexKeyByItem  func(entry *LogEntry) string
	index           map[string]*LogEntries
	indexKeyByQuery func(query logDBQuery) string
	mtx             sync.RWMutex // TO protect index
}

// NewLogIndex -
func NewLogIndex(indexKeyByItem func(entry *LogEntry) string, indexKeyByQuery func(query logDBQuery) string) LogIndex {
	return LogIndex{
		indexKeyByItem:  indexKeyByItem,
		indexKeyByQuery: indexKeyByQuery,
		index:           map[string]*LogEntries{},
	}
}

// Put an entry in the index
func (g *LogIndex) Put(item *LogEntry) {
	if item == nil || item.LogItem == nil {
		return
	}
	key := g.indexKeyByItem(item)
	g.mtx.Lock()
	defer g.mtx.Unlock()
	if _, ok := g.index[key]; !ok {
		g.index[key] = NewLogEntries()
	}
	g.index[key].Put(item, putOptionIfExistsReplace)
}

// Remove an entry from index
func (g *LogIndex) Remove(item *LogEntry) {
	if item == nil || item.LogItem == nil {
		return
	}
	key := g.indexKeyByItem(item)
	g.mtx.Lock()
	defer g.mtx.Unlock()
	if _, ok := g.index[key]; ok {
		g.index[key].Delete(item.ID())
	}
}

// Query by key
func (g *LogIndex) Query(queryIndexKey string) *LogEntries {
	g.mtx.RLock()
	defer g.mtx.RUnlock()
	if e, ok := g.index[queryIndexKey]; ok {
		return e
	}
	return NewLogEntries()
}
