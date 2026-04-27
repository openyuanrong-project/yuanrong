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

// Package etcd3 -
package etcd3

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

const (
	cacheMetaFilePrefix = "etcdCacheMeta_"
	cacheDataFilePrefix = "etcdCacheData_"
	backupFileSuffix    = "_backup"
	cacheDataSplitNum   = 3
)

var (
	// ErrInvalidCacheMeta -
	ErrInvalidCacheMeta = errors.New("invalid cache meta")
	// ErrCacheDataNotExist -
	ErrCacheDataNotExist = errors.New("cache data not exist")
	// ErrCacheDataMD5Mismatch -
	ErrCacheDataMD5Mismatch = errors.New("cache data md5 mismatch")
	cacheDataSeparator      = "|"
	cacheDataLineFeed       = []byte("\n")
)

// ETCDCacheMeta -
type ETCDCacheMeta struct {
	Revision int64  `json:"revision"`
	CacheMD5 string `json:"cacheMD5"`
}

func (ew *EtcdWatcher) setCacheFilePath() {
	cacheMetaFileName := fmt.Sprintf("%s%s", cacheMetaFilePrefix, strings.ReplaceAll(ew.key, "/", "#"))
	cacheDataFileName := fmt.Sprintf("%s%s", cacheDataFilePrefix, strings.ReplaceAll(ew.key, "/", "#"))
	ew.cacheConfig.MetaFilePath = fmt.Sprintf("%s/%s", ew.cacheConfig.PersistPath, cacheMetaFileName)
	ew.cacheConfig.DataFilePath = fmt.Sprintf("%s/%s", ew.cacheConfig.PersistPath, cacheDataFileName)
	ew.cacheConfig.BackupFilePath = fmt.Sprintf("%s%s", ew.cacheConfig.DataFilePath, backupFileSuffix)
}

func (ew *EtcdWatcher) processETCDCache() {
	log.GetLogger().Infof("start processing ETCD cache")
	ew.setCacheFilePath()
	persistInterval := ew.cacheConfig.FlushInterval
	ticker := time.NewTicker(time.Minute * time.Duration(persistInterval))
	defer ticker.Stop()
	// only record event with latest revision which is easier for flushCacheFile
	eventCache := make(map[string]*Event, ew.cacheConfig.FlushThreshold)
	for {
		select {
		case <-ticker.C:
			log.GetLogger().Infof("ticker triggers, flushing cache now")
			if err := ew.flushCacheToFile(eventCache); err == nil {
				eventCache = make(map[string]*Event, ew.cacheConfig.FlushThreshold)
			}
		case event := <-ew.CacheChan:
			log.GetLogger().Infof("threshold triggers, flushing cache now")
			preEvent, exist := eventCache[event.Key]
			if !exist || (exist && preEvent.Rev < event.Rev) {
				eventCache[event.Key] = event
			}
			if len(eventCache) > ew.cacheConfig.FlushThreshold {
				if err := ew.flushCacheToFile(eventCache); err == nil {
					eventCache = make(map[string]*Event, ew.cacheConfig.FlushThreshold)
				}
			}
		case <-ew.configCh:
			log.GetLogger().Infof("cache config changed, new config is %+v", ew.cacheConfig)
			if !ew.cacheConfig.EnableCache {
				log.GetLogger().Warnf("etcd cache disabled, stop processing cache")
				return
			}
			if ew.cacheConfig.FlushInterval != persistInterval {
				persistInterval = ew.cacheConfig.FlushInterval
				ticker.Reset(time.Minute * time.Duration(persistInterval))
			}
		case <-ew.stopCh:
			log.GetLogger().Warnf("etcd watcher stopped, stop processing cache")
			return
		}
	}
}

func (ew *EtcdWatcher) getCacheMeta() *ETCDCacheMeta {
	var cacheMeta *ETCDCacheMeta
	_, statErr := os.Stat(ew.cacheConfig.MetaFilePath)
	if os.IsNotExist(statErr) {
		return &ETCDCacheMeta{}
	}
	if statErr == nil {
		cacheMetaData, err := os.ReadFile(ew.cacheConfig.MetaFilePath)
		if err != nil {
			log.GetLogger().Errorf("failed to read cache meta file %s error %s", ew.cacheConfig.MetaFilePath,
				err.Error())
			return nil
		}
		cacheMeta = &ETCDCacheMeta{}
		if err = json.Unmarshal(cacheMetaData, cacheMeta); err != nil {
			log.GetLogger().Errorf("failed to unmarshal cache meta file %s error %s", ew.cacheConfig.MetaFilePath,
				err.Error())
			return nil
		}
		return cacheMeta
	}
	return nil
}

func (ew *EtcdWatcher) cleanCacheFile(cleanMeta, cleanData, cleanBackup bool) {
	if cleanMeta {
		if err := os.Remove(ew.cacheConfig.MetaFilePath); err != nil {
			log.GetLogger().Errorf("failed to remove cache meta file %s error %s", ew.cacheConfig.MetaFilePath,
				err.Error())
		}
	}
	if cleanData {
		if err := os.Remove(ew.cacheConfig.DataFilePath); err != nil {
			log.GetLogger().Errorf("failed to remove cache data file %s error %s", ew.cacheConfig.DataFilePath,
				err.Error())
		}
	}
	if cleanBackup {
		if err := os.Remove(ew.cacheConfig.BackupFilePath); err != nil {
			log.GetLogger().Errorf("failed to remove cache backup file %s error %s", ew.cacheConfig.BackupFilePath,
				err.Error())
		}
	}
}

// processDataBackup turns dataFile to backFile
func (ew *EtcdWatcher) processDataBackup(cacheMeta *ETCDCacheMeta) error {
	_, statDataFileErr := os.Stat(ew.cacheConfig.DataFilePath)
	_, statBackupFileErr := os.Stat(ew.cacheConfig.BackupFilePath)
	// need to handle backupFife if either dataFile or backupFile exists
	if statDataFileErr == nil || statBackupFileErr == nil {
		// if backupFile doesn't exist, it's the normal case, rename dataFile to backupFile if it exists.
		// if backupFile exists, it's the fault case where flush is interrupted, remove dataFile if it exists.
		if statDataFileErr == nil && os.IsNotExist(statBackupFileErr) {
			if err := os.Rename(ew.cacheConfig.DataFilePath, ew.cacheConfig.BackupFilePath); err != nil {
				log.GetLogger().Errorf("failed to rename cache file to %s error %s", ew.cacheConfig.BackupFilePath,
					err.Error())
				ew.cleanCacheFile(true, true, true)
				return err
			}
		} else if statDataFileErr == nil && statBackupFileErr == nil {
			if err := os.Remove(ew.cacheConfig.DataFilePath); err != nil {
				log.GetLogger().Errorf("failed to remove dirty cache file %s error %s", ew.cacheConfig.DataFilePath,
					err.Error())
				return err
			}
		}
		if utils.CalcFileMD5(ew.cacheConfig.BackupFilePath) != cacheMeta.CacheMD5 {
			log.GetLogger().Errorf("md5 mismatch for cache backup file %s", ew.cacheConfig.BackupFilePath)
			ew.cleanCacheFile(true, true, true)
			return ErrCacheDataMD5Mismatch
		}
	}
	return nil
}

// flushCacheToFile will modify the given eventCache during processing
func (ew *EtcdWatcher) flushCacheToFile(eventCache map[string]*Event) error {
	ew.Lock()
	if ew.cacheFlushing {
		ew.Unlock()
		return nil
	}
	ew.cacheFlushing = true
	defer func() {
		ew.Lock()
		ew.cacheFlushing = false
		ew.Unlock()
	}()
	ew.Unlock()
	cacheMeta := ew.getCacheMeta()
	if cacheMeta == nil {
		ew.cleanCacheFile(true, true, true)
		return ErrInvalidCacheMeta
	}
	// backup dataFile if it exists, will generate new dataFile from backupFile and eventCache
	if err := ew.processDataBackup(cacheMeta); err != nil {
		return err
	}
	var scanner *bufio.Scanner
	_, statBackupFileErr := os.Stat(ew.cacheConfig.BackupFilePath)
	if statBackupFileErr == nil {
		backupFile, err := os.OpenFile(ew.cacheConfig.BackupFilePath, os.O_RDONLY, 0600)
		if err != nil {
			log.GetLogger().Errorf("failed to open cache backup file %s error %s", ew.cacheConfig.BackupFilePath,
				err.Error())
			return err
		}
		defer func() {
			if err := backupFile.Close(); err != nil {
				log.GetLogger().Errorf("failed to close backup file %s error %s", ew.cacheConfig.BackupFilePath,
					err.Error())
			}
			if err := os.Remove(ew.cacheConfig.BackupFilePath); err != nil {
				log.GetLogger().Errorf("failed to remove backup file %s error %s", ew.cacheConfig.BackupFilePath,
					err.Error())
			}
		}()
		scanner = bufio.NewScanner(backupFile)
	}
	dataFile, err := os.OpenFile(ew.cacheConfig.DataFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 0600)
	if err != nil {
		log.GetLogger().Errorf("failed to open cache file %s error %s", ew.cacheConfig.DataFilePath, err.Error())
		return err
	}
	eventList := generateSortedCacheList(eventCache)
	offset := int64(0)
	for scanner != nil && scanner.Scan() {
		line := scanner.Text()
		items := strings.SplitN(line, cacheDataSeparator, cacheDataSplitNum)
		if len(items) != cacheDataSplitNum {
			log.GetLogger().Warnf("skip invalid data %s in cache file %s", line, ew.cacheConfig.BackupFilePath)
			continue
		}
		scanKey, scanValue := items[0], []byte(items[2])
		scanRevision, err := strconv.ParseInt(items[1], 10, 64)
		if err != nil {
			log.GetLogger().Errorf("invalid revision format of %s in line %s cache file %s", items[1], line,
				ew.cacheConfig.BackupFilePath)
			continue
		}
		if scanRevision > cacheMeta.Revision {
			cacheMeta.Revision = scanRevision
		}
		index := -1
		for i, event := range eventList {
			if event.Rev > cacheMeta.Revision {
				cacheMeta.Revision = event.Rev
			}
			// eventList keeps keys in lexicographical order which is also the order we set in cache data file, this
			// loop only handles eventKey <= scanKey scenario which contains two types of keys : 1. eventKey which goes
			// before scanKey with PUT type 2. eventKey equals to scanKey which will update or delete scanKey if it has
			// a newer revision
			if event.Key < scanKey {
				if event.Type == PUT {
					offset = flushEventToFile(dataFile, offset, []byte(event.Key), event.Value, event.Rev)
				}
				index = i
				continue
			}
			if event.Key == scanKey {
				// should not update or delete if event revision is older than cacheMeta
				if event.Rev > scanRevision && event.Type == PUT {
					scanValue = event.Value
					scanRevision = event.Rev
				} else if event.Rev > scanRevision && event.Type == DELETE {
					scanValue = nil
				}
				index = i
			}
			// here eventKey >= scanKey no need to go further
			break
		}
		if index != -1 {
			eventList = eventList[index+1:]
		}
		if scanValue != nil {
			offset = flushEventToFile(dataFile, offset, []byte(scanKey), scanValue, scanRevision)
		}
	}
	for _, event := range eventList {
		if event.Rev > cacheMeta.Revision {
			cacheMeta.Revision = event.Rev
		}
		if event.Type == PUT {
			offset = flushEventToFile(dataFile, offset, []byte(event.Key), event.Value, event.Rev)
		}
	}
	if err = dataFile.Close(); err != nil {
		log.GetLogger().Errorf("failed to close cache data file %s error %s", ew.cacheConfig.DataFilePath,
			err.Error())
	}
	if offset == 0 {
		log.GetLogger().Errorf("failed to write data file %s", ew.cacheConfig.DataFilePath)
		ew.cleanCacheFile(false, true, false)
		return errors.New("failed to write data file")
	}
	cacheMeta.CacheMD5 = utils.CalcFileMD5(ew.cacheConfig.DataFilePath)
	cacheMetaData, err := json.Marshal(cacheMeta)
	if err != nil {
		log.GetLogger().Errorf("failed to marshal cache meta error %s", err.Error())
		return err
	}
	if err = os.WriteFile(ew.cacheConfig.MetaFilePath, cacheMetaData, 0600); err != nil {
		log.GetLogger().Errorf("failed to write cache meta file %s error %s", ew.cacheConfig.MetaFilePath,
			err.Error())
		ew.cleanCacheFile(false, true, false)
		return err
	}
	log.GetLogger().Infof("succeed to flush cache")
	return nil
}

func (ew *EtcdWatcher) restoreCacheFromFile() error {
	ew.setCacheFilePath()
	_, statBackupFileErr := os.Stat(ew.cacheConfig.BackupFilePath)
	if statBackupFileErr == nil {
		// backupFile exists, it's the fault scenario, flushCacheToFile with nil to restore dataFile from backupFile
		ew.flushCacheToFile(nil)
	}
	_, statDataFileErr := os.Stat(ew.cacheConfig.DataFilePath)
	if os.IsNotExist(statDataFileErr) {
		return ErrCacheDataNotExist
	}
	cacheMeta := ew.getCacheMeta()
	if cacheMeta == nil {
		ew.cleanCacheFile(true, true, true)
		return ErrInvalidCacheMeta
	}
	if utils.CalcFileMD5(ew.cacheConfig.DataFilePath) != cacheMeta.CacheMD5 {
		log.GetLogger().Errorf("md5 mismatch for cache data file %s", ew.cacheConfig.DataFilePath)
		ew.cleanCacheFile(true, true, true)
		return ErrCacheDataMD5Mismatch
	}
	dataFile, err := os.OpenFile(ew.cacheConfig.DataFilePath, os.O_RDONLY, 0600)
	if err != nil {
		log.GetLogger().Errorf("failed to open cache backup file %s error %s", ew.cacheConfig.DataFilePath,
			err.Error())
		ew.cleanCacheFile(true, true, true)
		return err
	}
	scanner := bufio.NewScanner(dataFile)
	for scanner.Scan() {
		line := scanner.Text()
		items := strings.SplitN(line, cacheDataSeparator, cacheDataSplitNum)
		if len(items) != cacheDataSplitNum {
			log.GetLogger().Warnf("skip invalid data %s in cache file %s", line, ew.cacheConfig.DataFilePath)
			continue
		}
		scanKey, scanValue := items[0], []byte(items[2])
		scanRevision, err := strconv.ParseInt(items[1], 10, 64)
		if err != nil {
			log.GetLogger().Errorf("invalid revision format of %s in line %s file %s", items[1], line,
				ew.cacheConfig.DataFilePath)
			continue
		}
		ew.sendEvent(&Event{
			Type:  PUT,
			Key:   scanKey,
			Value: scanValue,
			Rev:   scanRevision,
		})
	}
	if err = dataFile.Close(); err != nil {
		log.GetLogger().Errorf("failed to close cache backup file %s error %s", ew.cacheConfig.DataFilePath,
			err.Error())
	}
	ew.initialRev = cacheMeta.Revision
	log.GetLogger().Infof("succeed to restore etcd cache to revision %d", cacheMeta.Revision)
	return nil
}

func flushEventToFile(f *os.File, offset int64, key, value []byte, revision int64) int64 {
	buffer := new(bytes.Buffer)
	buffer.Write(key)
	buffer.Write([]byte(cacheDataSeparator))
	buffer.Write([]byte(strconv.FormatInt(revision, 10)))
	buffer.Write([]byte(cacheDataSeparator))
	buffer.Write(value)
	buffer.Write(cacheDataLineFeed)
	_, err := f.WriteAt(buffer.Bytes(), offset)
	if err != nil {
		log.GetLogger().Errorf("failed to write content to cache file error %s", err.Error())
		return offset
	}
	return offset + int64(buffer.Len())
}

func generateSortedCacheList(cache map[string]*Event) []*Event {
	cacheList := make([]*Event, 0, len(cache))
	for _, v := range cache {
		cacheList = append(cacheList, v)
	}
	sort.Slice(cacheList, func(i, j int) bool {
		return cacheList[i].Key < cacheList[j].Key
	})
	return cacheList
}
