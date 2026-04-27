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

// Package logger rollingLog
package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"yuanrong.org/kernel/runtime/libruntime/common/logger/config"
)

const (
	megabyte        = 1024 * 1024
	defaultFileSize = 100
	defaultBackups  = 20
)

var logNameCache = struct {
	m map[string]string
	sync.Mutex
}{
	m:     make(map[string]string, 1),
	Mutex: sync.Mutex{},
}

type rollingLog struct {
	file         *os.File
	reg          *regexp.Regexp
	mu           sync.RWMutex
	sinks        []string
	dir          string
	nameTemplate string
	maxSize      int64
	size         int64
	maxBackups   int
	flag         int
	perm         os.FileMode
	isUserLog    bool
}

func initRollingLog(coreInfo config.CoreInfo, flag int, perm os.FileMode) (*rollingLog, error) {
	if coreInfo.FilePath == "" {
		return nil, errors.New("empty log file path")
	}
	log := &rollingLog{
		dir:          filepath.Dir(coreInfo.FilePath),
		nameTemplate: filepath.Base(coreInfo.FilePath),
		flag:         flag,
		perm:         perm,
		maxSize:      coreInfo.SingleSize * megabyte,
		maxBackups:   coreInfo.Threshold,
		isUserLog:    coreInfo.IsUserLog,
	}
	if log.maxBackups < 1 {
		log.maxBackups = defaultBackups
	}
	if log.maxSize < megabyte {
		log.maxSize = defaultFileSize * megabyte
	}
	if log.isUserLog {
		return log, log.tidySinks()
	}
	extension := filepath.Ext(log.nameTemplate)
	regExp := fmt.Sprintf(`^%s(?:(?:-|\.)\d*)?\%s$`,
		log.nameTemplate[:len(log.nameTemplate)-len(extension)], extension)
	reg, err := regexp.Compile(regExp)
	if err != nil {
		return nil, err
	}
	log.reg = reg
	return log, log.tidySinks()
}

func (r *rollingLog) tidySinks() error {
	if r.isUserLog || r.file != nil {
		return r.newSink()
	}
	// scan and reuse past log file when service restarted
	r.scanLogFiles()
	if len(r.sinks) > 0 {
		fullName := r.sinks[len(r.sinks)-1]
		info, err := os.Stat(fullName)
		if err != nil || info.Size() >= r.maxSize {
			return r.newSink()
		}
		file, err := os.OpenFile(fullName, r.flag, r.perm)
		if err == nil {
			r.file = file
			r.size = info.Size()
			return nil
		}
	}
	return r.newSink()
}

func (r *rollingLog) scanLogFiles() {
	dirEntrys, err := os.ReadDir(r.dir)
	if err != nil {
		fmt.Printf("failed to read dir: %s\n", r.dir)
		return
	}
	infos := make([]os.FileInfo, 0, r.maxBackups)
	for _, entry := range dirEntrys {
		if r.reg.MatchString(entry.Name()) {
			info, err := entry.Info()
			if err == nil {
				infos = append(infos, info)
			}
		}
	}
	if len(infos) > 0 {
		sort.Slice(infos, func(i, j int) bool {
			return infos[i].ModTime().Before(infos[j].ModTime())
		})
		for i := range infos {
			r.sinks = append(r.sinks, filepath.Join(r.dir, infos[i].Name()))
		}
		r.cleanRedundantSinks()
	}
}

func (r *rollingLog) cleanRedundantSinks() {
	if len(r.sinks) < r.maxBackups {
		return
	}
	curSinks := make([]string, 0, len(r.sinks))
	for _, name := range r.sinks {
		if isAvailable(name) {
			curSinks = append(curSinks, name)
		}

	}
	r.sinks = curSinks
	sinkNum := len(r.sinks)
	if sinkNum > r.maxBackups {
		removes := r.sinks[:sinkNum-r.maxBackups]
		go removeFiles(removes)
		r.sinks = r.sinks[sinkNum-r.maxBackups:]
	}
	return
}

func removeFiles(paths []string) {
	for _, path := range paths {
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("failed remove file %s\n", path)
		}
	}
}

func isAvailable(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (r *rollingLog) newSink() error {
	fullName := filepath.Join(r.dir, r.newName())
	if isAvailable(fullName) && r.file != nil && r.file.Name() == filepath.Base(fullName) {
		return errors.New("log file already opened: " + fullName)
	}
	file, err := os.OpenFile(fullName, r.flag, r.perm)
	if err != nil {
		return err
	}
	if r.file != nil {
		err = r.file.Close()
	}
	if err != nil {
		fmt.Printf("failed to close file: %s\n", err.Error())
	}
	r.file = file
	info, err := file.Stat()
	if err != nil {
		r.size = 0
	} else {
		r.size = info.Size()
	}
	r.sinks = append(r.sinks, fullName)
	r.cleanRedundantSinks()
	if r.isUserLog {
		logNameCache.Lock()
		logNameCache.m[r.nameTemplate] = fullName
		logNameCache.Unlock()
	}
	return nil
}

func (r *rollingLog) newName() string {
	if !r.isUserLog {
		timeNow := time.Now().Format("2006010215040506")
		ext := filepath.Ext(r.nameTemplate)
		return fmt.Sprintf("%s.%s%s", r.nameTemplate[:len(r.nameTemplate)-len(ext)], timeNow, ext)
	}
	if r.file == nil {
		return r.nameTemplate
	}
	timeNow := time.Now().Format("2006010215040506")
	var prefix, suffix string
	if index := strings.LastIndex(r.nameTemplate, "@") + 1; index <= len(r.nameTemplate) {
		prefix = r.nameTemplate[:index]
	}
	if index := strings.Index(r.nameTemplate, "#"); index >= 0 {
		suffix = r.nameTemplate[index:]
	}
	if prefix == "" || suffix == "" {
		return ""
	}
	return fmt.Sprintf("%s%s%s", prefix, timeNow, suffix)
}

// Write data to file and check whether to rotate log
func (r *rollingLog) Write(data []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r == nil || r.file == nil {
		return 0, errors.New("log file is nil")
	}
	n, err := r.file.Write(data)
	r.size += int64(n)
	if r.size > r.maxSize {
		r.tryRotate()
	}
	if syncErr := r.file.Sync(); syncErr != nil {
		fmt.Printf("failed to sync log err: %s\n", syncErr.Error())
	}
	return n, err
}

func (r *rollingLog) tryRotate() {
	if info, err := r.file.Stat(); err == nil && info.Size() < r.maxSize {
		return
	}
	err := r.tidySinks()
	if err != nil {
		fmt.Printf("failed to rotate log err: %s\n", err.Error())
	}
	return
}

// GetLogName get current log name when refreshing user log mod time
func GetLogName(nameTemplate string) string {
	logNameCache.Lock()
	name := logNameCache.m[nameTemplate]
	logNameCache.Unlock()
	return name
}
