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

// Package logmanager - this file is about log db
package logmanager

import (
	"encoding/json"
	"sync"
)

type logDBQuery struct {
	Filename     string `form:"filename" json:"filename"`
	SubmissionID string `form:"submission_id" json:"submission_id"`
	InstanceID   string `form:"instance_id" json:"instance_id"`
	RuntimeID    string
	CollectorID  string
}

type putOptionIfExists int32

const (
	putOptionIfExistsReplace putOptionIfExists = iota
	putOptionIfExistsNoop
)

// LogDB interface
type LogDB interface {
	Put(item *LogEntry, replaceIfExists putOptionIfExists)
	Remove(item *LogEntry)
	GetLogEntries() *LogEntries
	Serialize() ([]byte, error)
	Deserialize([]byte) error

	// Query by union mode by default, will join all Entries satisfy at least one condition, treat empty query condition
	// as NO SELECTION : i.e. return all results
	Query(query logDBQuery) *LogEntries
}

type generalLogDBImpl struct {
	Entries  *LogEntries `json:"items"`
	allIndex []LogIndex

	mtx sync.RWMutex // To protect Entries
}

func newGeneralLogDBImpl() *generalLogDBImpl {
	return &generalLogDBImpl{
		Entries: NewLogEntries(),
		allIndex: []LogIndex{
			NewLogIndex(func(item *LogEntry) string { return item.InstanceID },
				func(query logDBQuery) string { return query.InstanceID }),
			NewLogIndex(func(item *LogEntry) string { return item.Filename },
				func(query logDBQuery) string { return query.Filename }),
			NewLogIndex(func(item *LogEntry) string { return item.JobID },
				func(query logDBQuery) string { return query.SubmissionID }),
			NewLogIndex(func(item *LogEntry) string { return item.RuntimeID },
				func(query logDBQuery) string { return query.RuntimeID }),
		},
	}
}

// Put an item into the db, should contain corresponding jobs/instance-id
func (g *generalLogDBImpl) Put(item *LogEntry, replaceIfExists putOptionIfExists) {
	if item == nil {
		return
	}
	g.mtx.Lock()
	defer g.mtx.Unlock()
	g.Entries.Put(item, replaceIfExists)
	for k, _ := range g.allIndex {
		g.allIndex[k].Put(item)
	}
}

// Remove an entry
func (g *generalLogDBImpl) Remove(item *LogEntry) {
	g.mtx.Lock()
	g.Entries.Delete(item.ID())
	g.mtx.Unlock()

	for k, _ := range g.allIndex {
		g.allIndex[k].Remove(item)
	}
}

// GetLogEntries -
func (g *generalLogDBImpl) GetLogEntries() *LogEntries {
	return g.Entries
}

// Serialize to byte
func (g *generalLogDBImpl) Serialize() ([]byte, error) {
	g.mtx.RLock()
	defer g.mtx.RUnlock()
	return json.Marshal(g.Entries)
}

// Deserialize from binary
func (g *generalLogDBImpl) Deserialize(data []byte) error {
	entries := NewLogEntries()
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	// re-put to reconstruct the index
	entries.Range(func(entry *LogEntry) {
		g.Put(entry, putOptionIfExistsReplace)
	})
	return nil
}

// Query make sure no nil will be returns, only empty possible
func (g *generalLogDBImpl) Query(query logDBQuery) *LogEntries {
	allMatches := []*LogEntries{g.Entries}
	for k, _ := range g.allIndex {
		if queryString := g.allIndex[k].indexKeyByQuery(query); queryString != "" {
			allMatches = append(allMatches, g.allIndex[k].Query(queryString))
		}
	}
	allSatisfiedEntries := logEntriesIntersection(allMatches...)
	return allSatisfiedEntries
}
