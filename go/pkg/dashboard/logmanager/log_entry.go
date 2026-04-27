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
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
)

// LogEntry represents a log file, of one log exporter
type LogEntry struct {
	// reported by collectorClient
	*logservice.LogItem

	// matched by manager
	JobID      string
	InstanceID string
}

// Equal -
func (li *LogEntry) Equal(that *LogEntry) bool {
	if li == nil || that == nil {
		return li == that
	}
	return proto.Equal(li.LogItem, that.LogItem) && li.JobID == that.JobID && li.InstanceID == li.InstanceID
}

// String -
func (li *LogEntry) String() string {
	return li.LogItem.String() + fmt.Sprintf(`  instanceID:"%s"  jobID:"%s"`, li.InstanceID, li.JobID)
}

// NewLogEntry returns a log item pointer
func NewLogEntry(item *logservice.LogItem) *LogEntry {
	if item == nil {
		return nil
	}
	return &LogEntry{
		LogItem: item,
	}
}

// ID returns the key
func (li *LogEntry) ID() string {
	return fmt.Sprintf("%s//%s//%s", li.Filename, li.CollectorID, li.RuntimeID)
}

// LogEntries is a key-value map, key is item.ID(). This structure is to avoid complexity when delete some item from
// memory storage.
type LogEntries struct {
	Entries map[string]*LogEntry `json:"Entries"`
	mtx     sync.RWMutex
}

// NewLogEntries -
func NewLogEntries() *LogEntries {
	return &LogEntries{Entries: map[string]*LogEntry{}}
}

// FindFirst entry
func (l *LogEntries) FindFirst() *LogEntry {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	for _, v := range l.Entries {
		return v
	}
	return nil
}

// Get an entry by key
func (l *LogEntries) Get(k string) *LogEntry {
	l.mtx.RLock()
	defer l.mtx.RUnlock()
	if e, ok := l.Entries[k]; ok {
		return e
	}
	return nil
}

// Put an entry
func (l *LogEntries) Put(item *LogEntry, replaceIfExists putOptionIfExists) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	if _, ok := l.Entries[item.ID()]; ok && replaceIfExists == putOptionIfExistsNoop {
		return
	}
	l.Entries[item.ID()] = item
}

// Delete an entry
func (l *LogEntries) Delete(key string) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	delete(l.Entries, key)
}

// Range over all Entries, true to continue
func (l *LogEntries) Range(processor func(entry *LogEntry)) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	for _, e := range l.Entries {
		processor(e)
	}
}

// Len return length
func (l *LogEntries) Len() int {
	l.mtx.RLock()
	defer l.mtx.RUnlock()
	return len(l.Entries)
}

// String return string
func (l *LogEntries) String() string {
	l.mtx.RLock()
	defer l.mtx.RUnlock()
	var result string
	for k, e := range l.Entries {
		result += fmt.Sprintf(";k(%s),v(%#v)", k, e.String())
	}
	return result
}

func logEntriesIntersection(allEntries ...*LogEntries) *LogEntries {
	le := NewLogEntries()
	if len(allEntries) == 0 {
		return le
	}

	allEntries[0].mtx.RLock()
	defer allEntries[0].mtx.RUnlock()

	for k, e := range allEntries[0].Entries { // for each key
		matchedCnt := 0
		for _, each := range allEntries[1:] { // try match for each entry
			each.mtx.RLock()
			if _, ok := each.Entries[k]; ok {
				matchedCnt = matchedCnt + 1
			}
			each.mtx.RUnlock()
		}
		if matchedCnt == len(allEntries)-1 { // must match at least len(allEntries) - 1
			le.Entries[k] = e
		}
	}

	return le
}
