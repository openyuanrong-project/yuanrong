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

package logcollector

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"yuanrong.org/kernel/pkg/collector/common"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

const (
	maxFileChanSize = 100
)

// reportedLogFiles will not remove elements even if the target process exits
var reportedLogFiles = struct {
	sync.Mutex
	hashmap map[string]struct{}
}{
	hashmap: make(map[string]struct{}),
}

var logFileRegexMap = map[logservice.LogTarget]*regexp.Regexp{
	logservice.LogTarget_USER_STD: regexp.MustCompile(`^runtime([-_][a-z0-9]+)+.(out|err)`),
}

var runtimeRegex = regexp.MustCompile(`runtime([-_][a-z0-9]+)+`)

func tryReportLog(name string) bool {
	if item, ok := parseLogFileName(name); ok {
		reportedLogFiles.Lock()
		if _, exists := reportedLogFiles.hashmap[item.Filename]; exists {
			log.GetLogger().Debugf("%s is already reported", item.Filename)
			reportedLogFiles.Unlock()
			return false
		}

		log.GetLogger().Infof("find log file to report: %s", item.Filename)
		log.GetLogger().Debugf("log item details to report: %+v", item)
		reportedLogFiles.hashmap[item.Filename] = struct{}{}
		reportedLogFiles.Unlock()
		if err := connectLogService(item, reportLog); err != nil {
			log.GetLogger().Errorf("failed to report log file %s, error: %v", item.Filename, err)
			return false
		}
		return true
	}
	return false
}

func getRuntimeID(filename string) string {
	return runtimeRegex.FindString(filename)
}

func parseLogFileName(filePath string) (*logservice.LogItem, bool) {
	filename := filepath.Base(filePath)
	for target, regex := range logFileRegexMap {
		if regex.MatchString(filename) {
			runtimeID := getRuntimeID(filename)
			log.GetLogger().Debugf("%s matches %v, runtimeID: %s", filePath, target, runtimeID)
			return &logservice.LogItem{
				Filename:    filePath,
				CollectorID: common.CollectorConfigs.CollectorID,
				Target:      target,
				RuntimeID:   runtimeID,
			}, true
		}
	}
	return nil, false
}

func reportLog(client logservice.LogManagerServiceClient, ctx context.Context,
	item *logservice.LogItem) (*logservice.ReportLogResponse, error) {
	return client.ReportLog(ctx, &logservice.ReportLogRequest{
		Items: []*logservice.LogItem{item},
	})
}

func removeLog(client logservice.LogManagerServiceClient, ctx context.Context,
	item *logservice.LogItem) (*logservice.ReportLogResponse, error) {
	return client.RemoveLog(ctx, &logservice.ReportLogRequest{
		Items: []*logservice.LogItem{item},
	})
}

func connectLogService(item *logservice.LogItem, logFunc func(logservice.LogManagerServiceClient, context.Context,
	*logservice.LogItem) (*logservice.ReportLogResponse, error)) error {
	retryInterval := constant.GetRetryInterval()
	maxRetryTimes := constant.GetMaxRetryTimes()
	client := GetLogServiceClient()
	if client == nil {
		log.GetLogger().Errorf("failed to get log service client")
		return fmt.Errorf("failed to get log service client")
	}
	ctx, cancel := context.WithTimeout(context.Background(), common.DefaultGrpcTimeoutS)
	defer cancel()

	for i := 0; i < maxRetryTimes; i++ {
		log.GetLogger().Infof("start to report log %s, attempt: %d", item.Filename, i)
		response, err := logFunc(client, ctx, item)
		if err != nil {
			log.GetLogger().Errorf("failed to report log %s, error: %v", item.Filename, err)
			time.Sleep(retryInterval)
			continue
		}
		if response.Code != 0 {
			log.GetLogger().Errorf("failed to report log %s, error: %d, message: %s", item.Filename, response.Code,
				response.Message)
			time.Sleep(retryInterval)
			continue
		}
		log.GetLogger().Infof("success to report log %s", item.Filename)
		return nil
	}
	return fmt.Errorf("failed to report log: exceeds max retry time: %d", maxRetryTimes)
}

func handleFile(watcher *fsnotify.Watcher, newFileChan chan string, removeFileChan chan string, directory string) {
	for {
		select {
		case file, ok := <-newFileChan:
			if !ok {
				log.GetLogger().Warnf("new file event chan is closed")
				return
			}
			log.GetLogger().Debugf("find new file %s", file)
			if relPath, err := filepath.Rel(directory, file); err == nil {
				tryReportLog(relPath)
			}
		case file, ok := <-removeFileChan:
			if !ok {
				log.GetLogger().Warnf("remove file event chan is closed")
				return
			}
			log.GetLogger().Debugf("find remove file %s", file)
			tryRemoveLog(file, directory)
		case err, ok := <-watcher.Errors:
			if !ok {
				log.GetLogger().Warnf("new file event chan is closed")
				return
			}
			log.GetLogger().Warnf("new file event chan error: %v", err)
		}
	}
}

func tryRemoveLog(file string, directory string) {
	name, err := filepath.Rel(directory, file)
	if err != nil {
		log.GetLogger().Errorf("failed to remove log %s, error: %v", file, err)
		return
	}
	if item, ok := parseLogFileName(name); ok {
		reportedLogFiles.Lock()
		log.GetLogger().Infof("find log file to remove: %s", item.Filename)
		log.GetLogger().Debugf("log item details to remove: %+v", item)
		delete(reportedLogFiles.hashmap, item.Filename)
		reportedLogFiles.Unlock()
		if err = connectLogService(item, removeLog); err != nil {
			log.GetLogger().Errorf("failed to remove log file %s, error: %v", item.Filename, err)
		}
	}
}

func monitorFile(watcher *fsnotify.Watcher, newFileChan chan string, removeFileChan chan string, directory string) {
	defer close(newFileChan)
	defer close(removeFileChan)
	defer watcher.Close()
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.GetLogger().Warnf("watch event chan for %s is closed ", directory)
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				log.GetLogger().Debugf("find a new file is created: %s", event.Name)
				newFileChan <- event.Name
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
				log.GetLogger().Debugf("find a file is created: %s %s", event.Name, event.Op)
				removeFileChan <- event.Name
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.GetLogger().Warnf("watch event chan for %s is closed ", directory)
				return
			}
			log.GetLogger().Warnf("watch event for %s error: %v", directory, err)
		}
	}
}

// createLogReporter starts two go routines that never end
func createLogReporter(directory string) error {
	log.GetLogger().Infof("create log report for %s", directory)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.GetLogger().Errorf("failed to create file watcher, error: %v", err)
		return err
	}

	err = watcher.Add(directory)
	if err != nil {
		log.GetLogger().Errorf("failed to create file watcher for %s, error: %v", directory, err)
		return err
	}

	newFileChan := make(chan string, maxFileChanSize)
	removeFileChan := make(chan string, maxFileChanSize)
	go handleFile(watcher, newFileChan, removeFileChan, directory)
	go monitorFile(watcher, newFileChan, removeFileChan, directory)
	return nil
}

func scanUserLog(directory string) {
	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.GetLogger().Warnf("failed to access file %s under %s", path, directory)
			return err
		}
		relPath, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		log.GetLogger().Debugf("find file %s under %s", relPath, directory)
		tryReportLog(relPath)
		return nil
	})

	if err != nil {
		log.GetLogger().Errorf("failed to iterate files under %s, error: %v", directory, err)
	}
}

// StartLogReporter starts file watcher to report logs
func StartLogReporter() {
	err := createLogReporter(common.CollectorConfigs.UserLogPath)
	if err != nil {
		log.GetLogger().Errorf("failed to create log reporter for %s, error: %s", common.CollectorConfigs.UserLogPath,
			err)
		return
	}
	scanUserLog(common.CollectorConfigs.UserLogPath)
}
