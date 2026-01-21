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
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	dsCommon "clients/common"
	"clients/stream"

	"yuanrong.org/kernel/pkg/collector/common"
	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/logservice"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

const (
	delayFlushTime int64  = 5
	pageSize       int64  = 1 * 1024 * 1024
	maxStreamSize  uint64 = 100 * 1024 * 1024

	maxRetryCount        = 3
	defaultRetryInterval = 1 * time.Second
	oomRetryInterval     = 10 * time.Second

	outOfMemory            = 6
	tryAgain               = 19
	rpcUnavailable         = 1002
	streamDeleteInProgress = 3007
)

var limitedRetryCode = map[int]struct{}{
	tryAgain:               struct{}{},
	rpcUnavailable:         struct{}{},
	streamDeleteInProgress: struct{}{},
}

var (
	streamClient        wrappedStreamClient
	streamClientPtr     streamClientInterface
	getStreamClientOnce sync.Once
)

type streamClientInterface interface {
	CreateProducerInterface(streamName string, delayFlushTimeMs int64, pageSize int64, maxStreamSize uint64,
		autoCleanup bool) (streamProducerInterface, dsCommon.Status)
	Init(reportWorkerLost bool) dsCommon.Status
	DestroyClient()
}

type wrappedStreamClient struct {
	stream.StreamClient
}

func (w *wrappedStreamClient) CreateProducerInterface(streamName string, delayFlushTimeMs int64, pageSize int64,
	maxStreamSize uint64, autoCleanup bool) (streamProducerInterface, dsCommon.Status) {
	return w.CreateProducer(streamName, delayFlushTimeMs, pageSize, maxStreamSize, autoCleanup)
}

type streamProducerInterface interface {
	Send(element stream.Element) dsCommon.Status
	Close() dsCommon.Status
}

func getStreamClient() streamClientInterface {
	getStreamClientOnce.Do(func() {
		arguments := dsCommon.ConnectArguments{
			Host: common.CollectorConfigs.IP,
			Port: common.CollectorConfigs.DatasystemPort,
		}
		streamClient = wrappedStreamClient{
			StreamClient: stream.CreateClient(arguments),
		}
		if status := streamClient.Init(false); status.IsError() {
			streamClient.DestroyClient()
			log.GetLogger().Errorf("failed to init stream client, error: %s", status.ToString())
			return
		}
		streamClientPtr = &streamClient
	})
	return streamClientPtr
}

// LogStreamPublisher -
type LogStreamPublisher struct {
	client                     streamClientInterface
	producer                   streamProducerInterface
	filename                   string
	streamName                 string
	runtimeID                  string
	offset                     int64
	healthy                    bool
	timestampForUnfinishedLine *time.Time
}

// NewLogStreamPublisher creates a new log stream publisher
func NewLogStreamPublisher(streamName string, item *logservice.LogItem, absoluteFilename string) (*LogStreamPublisher,
	error) {
	client := getStreamClient()
	if client == nil {
		return nil, errors.New("failed to get stream client")
	}
	producer, status := client.CreateProducerInterface(streamName, delayFlushTime, pageSize, maxStreamSize, true)
	if status.IsError() {
		log.GetLogger().Errorf("failed to create producer, error: %s", status.ToString())
		return nil, errors.New(status.ToString())
	}
	return &LogStreamPublisher{
		client:                     client,
		producer:                   producer,
		filename:                   absoluteFilename,
		streamName:                 streamName,
		runtimeID:                  item.RuntimeID,
		offset:                     0,
		healthy:                    true,
		timestampForUnfinishedLine: nil,
	}, nil
}

func checkNeedRetry(status dsCommon.Status) (bool, bool, time.Duration) {
	if status.Code == outOfMemory {
		return true, true, oomRetryInterval
	}
	if _, exists := limitedRetryCode[status.Code]; exists {
		return true, false, defaultRetryInterval
	}
	return false, false, 0 * time.Second
}

func (p *LogStreamPublisher) sendElementWithRetry(element stream.Element) error {
	retryCount := 0
	var status dsCommon.Status
	for retryCount < maxRetryCount {
		status = p.producer.Send(element)
		if !status.IsError() {
			log.GetLogger().Debugf("success to send element to %s", p.streamName)
			return nil
		}
		log.GetLogger().Warnf("failed to send element after %d times, error: %s", retryCount, status.ToString())
		shouldRetry, isUnlimitedRetry, interval := checkNeedRetry(status)
		if !shouldRetry {
			break
		}
		time.Sleep(interval)
		if !isUnlimitedRetry {
			retryCount++
		}
	}
	log.GetLogger().Errorf("will not retry to send elements, error: %s", status.ToString())
	p.healthy = false
	return errors.New(status.ToString())
}

func (p *LogStreamPublisher) preparePrefix() []byte {
	prefix := ""
	if len(p.runtimeID) > 0 {
		prefix += "(" + p.runtimeID + ") "
	}
	return []byte(prefix)
}

func buildBuffer(buffer *bytes.Buffer, prefixBytes []byte, line []byte) error {
	buffer.Grow(len(prefixBytes) + len(line))
	if _, err := buffer.Write(prefixBytes); err != nil {
		return err
	}
	if _, err := buffer.Write(line); err != nil {
		return err
	}
	return nil
}

func (p *LogStreamPublisher) shouldSendLine(line []byte, err error) bool {
	if len(line) > 0 && line[len(line)-1] != '\n' && errors.Is(err, io.EOF) {
		// no timestamp or timestamp is not expired yet
		if p.timestampForUnfinishedLine == nil {
			p.timestampForUnfinishedLine = &time.Time{}
			*p.timestampForUnfinishedLine = time.Now()
			return false
		} else if time.Since(*p.timestampForUnfinishedLine) <= constant.GetMaxTimeForUnfinishedLine() {
			*p.timestampForUnfinishedLine = time.Now()
			return false
		} else {
			p.timestampForUnfinishedLine = nil
			return true
		}
	}
	p.timestampForUnfinishedLine = nil
	return true
}

func (p *LogStreamPublisher) internalPublish() bool {
	if !p.healthy {
		log.GetLogger().Errorf("publisher is not healthy")
		return false
	}
	file, err := os.OpenFile(p.filename, os.O_RDONLY, 0)
	if err != nil {
		log.GetLogger().Errorf("failed to open %s, error: %s", p.filename, err)
		return false
	}
	defer file.Close()

	if _, err := file.Seek(p.offset, 0); err != nil {
		log.GetLogger().Errorf("failed to seek file %s to %d, error: %s", p.filename, p.offset, err)
		return false
	}

	reader := bufio.NewReader(file)
	prefixBytes := p.preparePrefix()
	for {
		line, err := reader.ReadBytes('\n')
		if !p.shouldSendLine(line, err) {
			log.GetLogger().Warnf("read %s until EOF but find an unfinished line", p.filename)
			return true
		}
		if len(line) > 0 {
			buffer := &bytes.Buffer{}
			if err := buildBuffer(buffer, prefixBytes, line); err != nil {
				log.GetLogger().Errorf("failed to write line to buffer, error: %v", err)
				return false
			}
			element := stream.Element{Ptr: &buffer.Bytes()[0], Size: uint64(buffer.Len())}
			if err := p.sendElementWithRetry(element); err != nil {
				return false
			}
			p.offset += int64(len(line))
		}
		if errors.Is(err, io.EOF) {
			break
		} else if errors.Is(err, bufio.ErrBufferFull) {
			log.GetLogger().Warnf("published an unfinished long line because buffer is full, length: %d", len(line))
			continue
		} else if err != nil {
			log.GetLogger().Errorf("failed to read file %s from %d, error: %s", p.filename, p.offset, err)
			return false
		}
	}
	return false
}

func (p *LogStreamPublisher) publish(watcher *fsnotify.Watcher, event fsnotify.Event) {
	findUnfinishedLine := p.internalPublish()
	if findUnfinishedLine {
		log.GetLogger().Infof("found unfinished line; read file again after %v", constant.GetMaxTimeForUnfinishedLine())
		time.AfterFunc(constant.GetMaxTimeForUnfinishedLine(), func() {
			watcher.Events <- event
		})
	}
}

func handleFileUpdates(watcher *fsnotify.Watcher, publisher *LogStreamPublisher, done chan bool,
	item *logservice.LogItem) {
	defer close(done)
	defer publisher.producer.Close()
	defer watcher.Close()

	publisher.publish(watcher, fsnotify.Event{
		Name: "inner-trigger-update-event",
		Op:   fsnotify.Write,
	})

	for {
		select {
		case <-done:
			log.GetLogger().Debugf("watch for %s update is done", item.Filename)
			return
		case event, ok := <-watcher.Events:
			if !ok {
				log.GetLogger().Warnf("watch event chan for %s is closed", item.Filename)
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				publisher.publish(watcher, event)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				log.GetLogger().Warnf("watch event chan for %s is closed ", item.Filename)
				return
			}
			log.GetLogger().Warnf("watch event for %s error: %v", item.Filename, err)
		}
	}
}

func addToStreamMap(streamName string, filename string, done chan bool) error {
	streamControlChans.Lock()
	if streamMap, exists := streamControlChans.hashmap[streamName]; exists {
		if _, opened := streamMap[filename]; opened {
			streamControlChans.Unlock()
			return errors.New("file" + filename + "is already opened for log stream")
		}
	} else {
		streamControlChans.hashmap[streamName] = make(map[string]chan bool)
	}
	streamControlChans.hashmap[streamName][filename] = done
	streamControlChans.Unlock()
	return nil
}

func removeFromStreamMap(streamName string, filename string) {
	streamControlChans.Lock()
	defer streamControlChans.Unlock()
	if streamMap, exists := streamControlChans.hashmap[streamName]; exists {
		if _, opened := streamMap[filename]; opened {
			delete(streamMap, filename)
		}
	}
}

// CreateLogStreamPublisher creates a new log stream publisher for the incoming request
func CreateLogStreamPublisher(streamName string, item *logservice.LogItem) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.GetLogger().Errorf("failed to create file watcher, error: %v", err)
		return err
	}
	var absoluteFilename string
	absoluteFilename, err = GetAbsoluteFilePath(item)
	if err != nil {
		log.GetLogger().Errorf("failed to get absolute path for %s, error: %v", item.Filename, err)
		return err
	}

	done := make(chan bool)
	if err := addToStreamMap(streamName, item.Filename, done); err != nil {
		return err
	}

	err = watcher.Add(absoluteFilename)
	if err != nil {
		log.GetLogger().Errorf("failed to create file watcher for %s, error: %v", item.Filename, err)
		removeFromStreamMap(streamName, item.Filename)
		watcher.Close()
		return err
	}

	publisher, err := NewLogStreamPublisher(streamName, item, absoluteFilename)
	if err != nil {
		log.GetLogger().Errorf("failed to create new log stream publisher, error: %v", err)
		removeFromStreamMap(streamName, item.Filename)
		watcher.Close()
		return err
	}

	go handleFileUpdates(watcher, publisher, done, item)
	return nil
}

// CloseLogStreamPublisher close log stream publisher based on its stream name
func CloseLogStreamPublisher(streamName string) error {
	streamControlChans.Lock()
	defer streamControlChans.Unlock()
	if streamMap, exists := streamControlChans.hashmap[streamName]; exists {
		for _, done := range streamMap {
			if done != nil {
				done <- true
			}
		}
		delete(streamControlChans.hashmap, streamName)
		return nil
	}
	err := errors.New("stream name" + streamName + "is not found")
	log.GetLogger().Errorf(err.Error())
	return err
}
