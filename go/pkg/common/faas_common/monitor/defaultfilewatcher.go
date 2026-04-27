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

// Package monitor provide memory and file monitor
package monitor

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

var (
	hashRetry = 60
)

type defaultFileWatcher struct {
	watcher  *fsnotify.Watcher
	filename string
	callback FileChangedCallback
	hash     string
	stopCh   <-chan struct{}
}

func createDefaultFileWatcher(stopCh <-chan struct{}) (FileWatcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &defaultFileWatcher{
		watcher: fsWatcher,
		stopCh:  stopCh,
	}
	return w, nil
}

// RegisterCallback impl
func (w *defaultFileWatcher) RegisterCallback(filename string, callback FileChangedCallback) {
	realPath, err := w.getRealPath(filename)
	if err != nil {
		log.GetLogger().Errorf("filename %s getRealPath failed err %s", filename, err.Error())
		return
	}

	if callback == nil {
		log.GetLogger().Errorf("filename %s callback is nil", filename)
		return
	}

	hash := w.getFileHashRetry(filename)
	w.filename = filename
	w.callback = callback
	w.hash = hash
	if err := w.watcher.Add(realPath); err != nil {
		log.GetLogger().Warnf("watch file %s failed", filename)
	} else {
		log.GetLogger().Infof("file %s RegisterCallback, success", filename)
	}
}

func (w *defaultFileWatcher) getRealPath(filename string) (string, error) {
	realPath, err := filepath.EvalSymlinks(filename)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(realPath); err != nil {
		return "", err
	}
	return realPath, nil
}

func (w *defaultFileWatcher) handleFileRemove(event fsnotify.Event) {
	// remove old watcher
	w.watcher.Remove(event.Name)
	w.watcher.Remove(w.filename)

	// re-add new watcher
	realPath, err := w.getRealPath(w.filename)
	if err != nil {
		log.GetLogger().Warnf("filename %s getRealPath failed err %s", w.filename, err.Error())
	} else {
		if err := w.watcher.Add(realPath); err != nil {
			log.GetLogger().Warnf("re-add watcher %s failed", realPath)
		} else {
			log.GetLogger().Infof("re-add watcher %s success", realPath)
		}
	}

	if err := w.watcher.Add(w.filename); err != nil {
		log.GetLogger().Warnf("re-add watcher %s failed", w.filename)
	} else {
		log.GetLogger().Infof("re-add watcher %s success", w.filename)
	}
}

// Start impl
func (w *defaultFileWatcher) Start() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				log.GetLogger().Errorf("watcher event chan not ok")
				continue
			}
			w.invokeCallback(event)
			if event.Op == fsnotify.Remove {
				w.handleFileRemove(event)
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				log.GetLogger().Errorf("errors chan not ok, err %s", err.Error())
			}
		case <-w.stopCh:
			w.watcher.Close()
			return
		}
	}
}

func (w *defaultFileWatcher) invokeCallback(event fsnotify.Event) {
	newHash := w.getFileHashRetry(w.filename)
	if newHash != w.hash {
		begin := time.Now()
		log.GetLogger().Infof("file event %s happen, start invoke callback", event.String())
		w.hash = newHash
		w.callback(w.filename, OpType(event.Op))
		log.GetLogger().Infof("file event %s invoke callback success, cost %v",
			event.String(), time.Since(begin))
	}
}

func (w *defaultFileWatcher) getFileHashRetry(filename string) string {
	for i := 0; i < hashRetry; i++ {
		hash := w.getFileHash(filename)
		if len(hash) > 0 {
			return hash
		}
		time.Sleep(1 * time.Second)
	}
	return ""
}

func (w *defaultFileWatcher) getFileHash(filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}
	hash := sha256.New()
	_, err = hash.Write(content)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(hash.Sum(nil))
}
