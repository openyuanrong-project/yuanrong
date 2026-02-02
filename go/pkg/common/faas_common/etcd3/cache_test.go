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
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

func TestProcessETCDCache(t *testing.T) {
	hackTicker := time.NewTicker(50 * time.Millisecond)
	resetDuration := time.Duration(0)
	patches := []*gomonkey.Patches{
		gomonkey.ApplyMethod(reflect.TypeOf(&time.Ticker{}), "Reset", func(_ *time.Ticker, d time.Duration) {
			resetDuration = d
		}),
	}
	defer func() {
		for _, p := range patches {
			p.Reset()
		}
	}()
	convey.Convey("ticker case", t, func() {
		patch := gomonkey.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
			hackTicker.Reset(50 * time.Millisecond)
			return hackTicker
		})
		os.Remove("etcdCacheMeta_#sn#function")
		os.Remove("etcdCacheData_#sn#function")
		os.Remove("etcdCacheData_#sn#function_backup")
		stopCh := make(chan struct{}, 1)
		ew := newEtcdWatcher()
		ew.stopCh = stopCh
		go ew.processETCDCache()
		time.Sleep(500 * time.Millisecond)
		ew.CacheChan <- &Event{
			Rev:   100,
			Type:  PUT,
			Key:   "/sn/function/123/hello/latest",
			Value: []byte(`{"name":"hello","version":"latest"}`),
		}
		time.Sleep(500 * time.Millisecond)
		stopCh <- struct{}{}
		time.Sleep(500 * time.Millisecond)
		data, err := os.ReadFile("etcdCacheData_#sn#function")
		convey.So(err, convey.ShouldBeNil)
		convey.So(string(data), convey.ShouldEqual, "/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n")
		data, err = os.ReadFile("etcdCacheMeta_#sn#function")
		convey.So(err, convey.ShouldBeNil)
		convey.So(string(data), convey.ShouldEqual, "{\"revision\":100,\"cacheMD5\":\"4fca8f1c736ca30135ed16538f4aebfc\"}")
		patch.Reset()
	})
	convey.Convey("threshold case", t, func() {
		os.Remove("etcdCacheMeta_#sn#function")
		os.Remove("etcdCacheData_#sn#function")
		os.Remove("etcdCacheData_#sn#function_backup")
		stopCh := make(chan struct{}, 1)
		ew := newEtcdWatcher()
		ew.stopCh = stopCh
		ew.cacheConfig.FlushThreshold = 0
		go ew.processETCDCache()
		time.Sleep(500 * time.Millisecond)
		ew.CacheChan <- &Event{
			Rev:   100,
			Type:  PUT,
			Key:   "/sn/function/123/hello/latest",
			Value: []byte(`{"name":"hello","version":"latest"}`),
		}
		time.Sleep(500 * time.Millisecond)
		stopCh <- struct{}{}
		time.Sleep(500 * time.Millisecond)
		data, err := os.ReadFile("etcdCacheData_#sn#function")
		convey.So(err, convey.ShouldBeNil)
		convey.So(string(data), convey.ShouldEqual, "/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n")
		data, err = os.ReadFile("etcdCacheMeta_#sn#function")
		convey.So(err, convey.ShouldBeNil)
		convey.So(string(data), convey.ShouldEqual, "{\"revision\":100,\"cacheMD5\":\"4fca8f1c736ca30135ed16538f4aebfc\"}")
	})
	convey.Convey("config update case", t, func() {
		os.Remove("etcdCacheMeta_#sn#function")
		os.Remove("etcdCacheData_#sn#function")
		os.Remove("etcdCacheData_#sn#function_backup")
		ew := newEtcdWatcher()
		go ew.processETCDCache()
		time.Sleep(500 * time.Millisecond)
		ew.cacheConfig.FlushInterval = 20
		ew.configCh <- struct{}{}
		time.Sleep(500 * time.Millisecond)
		convey.So(resetDuration, convey.ShouldEqual, 20*time.Minute)
		ew.cacheConfig.EnableCache = false
		ew.configCh <- struct{}{}
	})
	os.Remove("etcdCacheMeta_#sn#function")
	os.Remove("etcdCacheData_#sn#function")
	os.Remove("etcdCacheData_#sn#function_backup")
}

func TestFlushCacheToFile(t *testing.T) {
	stopCh := make(chan struct{}, 1)
	ew := &EtcdWatcher{
		key:       "/sn/function",
		CacheChan: make(chan *Event, 10),
		configCh:  make(chan struct{}, 1),
		stopCh:    stopCh,
		cacheConfig: EtcdCacheConfig{
			EnableCache:    true,
			PersistPath:    "./",
			FlushInterval:  10,
			FlushThreshold: 10,
		},
	}
	ew.setCacheFilePath()
	convey.Convey("no dataFile and no backupFile", t, func() {
		os.Remove("etcdCacheMeta_#sn#function")
		os.Remove("etcdCacheData_#sn#function")
		os.Remove("etcdCacheData_#sn#function_backup")
		eventBuffer := map[string]*Event{
			"/sn/function/123/hello/latest": &Event{
				Rev:   100,
				Key:   "/sn/function/123/hello/latest",
				Value: []byte(`{"name":"hello","version":"latest"}`),
			},
		}
		err := ew.flushCacheToFile(eventBuffer)
		convey.So(err, convey.ShouldBeNil)
		_, stateMetaFileErr := os.Stat("./etcdCacheMeta_#sn#function")
		_, stateDataFileErr := os.Stat("./etcdCacheData_#sn#function")
		_, stateBackupFileErr := os.Stat("./etcdCacheData_#sn#function_backup")
		convey.So(stateMetaFileErr, convey.ShouldBeNil)
		convey.So(stateDataFileErr, convey.ShouldBeNil)
		convey.So(os.IsNotExist(stateBackupFileErr), convey.ShouldEqual, true)
	})
	convey.Convey("no backupFile", t, func() {
		os.Remove("etcdCacheMeta_#sn#function")
		os.Remove("etcdCacheData_#sn#function")
		os.Remove("etcdCacheData_#sn#function_backup")
		// dataFile exists and no metaFile
		os.WriteFile("./etcdCacheData_#sn#function", []byte("/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n"), 0600)
		err := ew.flushCacheToFile(nil)
		convey.So(err, convey.ShouldNotBeNil)
		_, errStatMeta := os.Stat("./etcdCacheMeta_#sn#function")
		_, errStatData := os.Stat("./etcdCacheData_#sn#function")
		_, errStatBackup := os.Stat("./etcdCacheData_#sn#function_backup")
		convey.So(os.IsNotExist(errStatMeta), convey.ShouldEqual, true)
		convey.So(os.IsNotExist(errStatData), convey.ShouldEqual, true)
		convey.So(os.IsNotExist(errStatBackup), convey.ShouldEqual, true)
		// dataFile exists and metaFile exists
		os.WriteFile("./etcdCacheMeta_#sn#function", []byte(`{"revision":101,"cacheMD5":"726eb6f3140438ac1cbe334777e1a272"}`), 0600)
		os.WriteFile("./etcdCacheData_#sn#function", []byte("/sn/function/123/goodbye/latest|101|{\"name\":\"goodbye\",\"version\":\"latest\"}\n/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n/sn/function/123/invalid/latest|xxx|{\"name\":\"invalid\",\"version\":\"latest\"}\nThisIsInvalidData\n"), 0600)
		eventBuffer := map[string]*Event{
			"/sn/function/123/goodbye/latest": &Event{
				Rev:   102,
				Type:  PUT,
				Key:   "/sn/function/123/goodbye/latest",
				Value: []byte(`{"name":"goodbye","version":"v1"}`),
			},
			"/sn/function/123/hello/latest": &Event{
				Rev:   103,
				Type:  DELETE,
				Key:   "/sn/function/123/hello/latest",
				Value: []byte(`{"name":"hello","version":"latest"}`),
			},
			"/sn/function/123/echo/latest": &Event{
				Rev:   104,
				Type:  PUT,
				Key:   "/sn/function/123/echo/latest",
				Value: []byte(`{"name":"echo","version":"latest"}`),
			},
		}
		err = ew.flushCacheToFile(eventBuffer)
		convey.So(err, convey.ShouldBeNil)
		data, err := os.ReadFile("./etcdCacheMeta_#sn#function")
		convey.So(err, convey.ShouldBeNil)
		meta := &ETCDCacheMeta{}
		err = json.Unmarshal(data, meta)
		convey.So(err, convey.ShouldBeNil)
		convey.So(meta.Revision, convey.ShouldEqual, 104)
		convey.So(meta.CacheMD5, convey.ShouldEqual, "006731eddc832c067f9814b64ae12833")
		convey.So(utils.CalcFileMD5("./etcdCacheData_#sn#function"), convey.ShouldEqual, "006731eddc832c067f9814b64ae12833")
		_, stateBackupFileErr := os.Stat("./etcdCacheData_#sn#function_backup")
		convey.So(os.IsNotExist(stateBackupFileErr), convey.ShouldEqual, true)
	})
	convey.Convey("backupFile exists", t, func() {
		os.Remove("etcdCacheMeta_#sn#function")
		os.Remove("etcdCacheData_#sn#function")
		os.Remove("etcdCacheData_#sn#function_backup")
		// backupFile mismatch with metaFile
		os.WriteFile("./etcdCacheMeta_#sn#function", []byte(`{"revision":100,"cacheMD5":"4f4449c598ec58854d7104c4a64e979f"}`), 0600)
		os.WriteFile("./etcdCacheData_#sn#function_backup", []byte("/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n"), 0600)
		err := ew.flushCacheToFile(nil)
		convey.So(err, convey.ShouldNotBeNil)
		_, errStatMeta := os.Stat("./etcdCacheMeta_#sn#function")
		_, errStatData := os.Stat("./etcdCacheData_#sn#function")
		_, errStatBackup := os.Stat("./etcdCacheData_#sn#function_backup")
		convey.So(os.IsNotExist(errStatMeta), convey.ShouldEqual, true)
		convey.So(os.IsNotExist(errStatData), convey.ShouldEqual, true)
		convey.So(os.IsNotExist(errStatBackup), convey.ShouldEqual, true)
		// backupFile exists and dataFile exists
		os.WriteFile("./etcdCacheMeta_#sn#function", []byte(`{"revision":100,"cacheMD5":"4fca8f1c736ca30135ed16538f4aebfc"}`), 0600)
		os.WriteFile("./etcdCacheData_#sn#function", []byte("/sn/function/123/goodbye/latest|101|{\"name\":\"goodbye\",\"version\":\"latest\"}\n"), 0600)
		os.WriteFile("./etcdCacheData_#sn#function_backup", []byte("/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n"), 0600)
		err = ew.flushCacheToFile(nil)
		convey.So(err, convey.ShouldBeNil)
		data, err := os.ReadFile("./etcdCacheMeta_#sn#function")
		convey.So(err, convey.ShouldBeNil)
		meta := &ETCDCacheMeta{}
		err = json.Unmarshal(data, meta)
		convey.So(err, convey.ShouldBeNil)
		convey.So(meta.Revision, convey.ShouldEqual, 100)
		convey.So(meta.CacheMD5, convey.ShouldEqual, "4fca8f1c736ca30135ed16538f4aebfc")
		convey.So(utils.CalcFileMD5("./etcdCacheData_#sn#function"), convey.ShouldEqual, "4fca8f1c736ca30135ed16538f4aebfc")
		_, stateBackupFileErr := os.Stat("./etcdCacheData_#sn#function_backup")
		convey.So(os.IsNotExist(stateBackupFileErr), convey.ShouldEqual, true)
	})
	convey.Convey("file close fail", t, func() {
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(&os.File{}), "Close", func(f *os.File) error {
			return errors.New("some error")
		})
		fileHack, _ := os.OpenFile("./xxx", os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_SYNC, 0600)
		patch2 := gomonkey.ApplyFunc(os.OpenFile, func(name string, flag int, perm os.FileMode) (*os.File, error) {
			return fileHack, nil
		})
		os.Remove("etcdCacheMeta_#sn#function")
		os.Remove("etcdCacheData_#sn#function")
		os.Remove("etcdCacheData_#sn#function_backup")
		os.WriteFile("./etcdCacheMeta_#sn#function", []byte(`{"revision":100,"cacheMD5":"4fca8f1c736ca30135ed16538f4aebfc"}`), 0600)
		os.WriteFile("./etcdCacheData_#sn#function", []byte("/sn/function/123/goodbye/latest|101|{\"name\":\"goodbye\",\"version\":\"latest\"}\n"), 0600)
		os.WriteFile("./etcdCacheData_#sn#function_backup", []byte("/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n"), 0600)
		err := ew.flushCacheToFile(nil)
		convey.So(err, convey.ShouldNotBeNil)
		os.Remove("etcdCacheMeta_#sn#function")
		err = ew.flushCacheToFile(nil)
		convey.So(err, convey.ShouldNotBeNil)
		patch1.Reset()
		patch2.Reset()
		fileHack.Close()
		os.Remove("./xxx")
	})
	os.Remove("etcdCacheMeta_#sn#function")
	os.Remove("etcdCacheData_#sn#function")
	os.Remove("etcdCacheData_#sn#function_backup")
}

func TestRestoreCacheFromFile(t *testing.T) {
	stopCh := make(chan struct{}, 1)
	ew := &EtcdWatcher{
		key:        "/sn/function",
		ResultChan: make(chan *Event, 10),
		CacheChan:  make(chan *Event, 10),
		configCh:   make(chan struct{}, 1),
		stopCh:     stopCh,
		cacheConfig: EtcdCacheConfig{
			EnableCache:    true,
			PersistPath:    "./",
			FlushInterval:  10,
			FlushThreshold: 10,
		},
	}
	convey.Convey("no dataFile", t, func() {
		err := ew.restoreCacheFromFile()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(len(ew.ResultChan), convey.ShouldEqual, 0)
	})
	convey.Convey("backupFile exists", t, func() {
		os.Remove("etcdCacheMeta_#sn#function")
		os.Remove("etcdCacheData_#sn#function")
		os.Remove("etcdCacheData_#sn#function_backup")
		// invalid metaFile
		os.WriteFile("./etcdCacheMeta_#sn#function", []byte(`this is a invalid json`), 0600)
		os.WriteFile("./etcdCacheData_#sn#function_backup", []byte("/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n"), 0600)
		err := ew.restoreCacheFromFile()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(len(ew.ResultChan), convey.ShouldEqual, 0)
		_, errStatMeta := os.Stat("./etcdCacheMeta_#sn#function")
		_, errStatData := os.Stat("./etcdCacheData_#sn#function")
		_, errStatBackup := os.Stat("./etcdCacheData_#sn#function_backup")
		convey.So(os.IsNotExist(errStatMeta), convey.ShouldEqual, true)
		convey.So(os.IsNotExist(errStatData), convey.ShouldEqual, true)
		convey.So(os.IsNotExist(errStatBackup), convey.ShouldEqual, true)
		// backupFile mismatches with metaFile
		os.WriteFile("./etcdCacheMeta_#sn#function", []byte(`{"revision":100,"cacheMD5":"4f4449c598ec58854d7104c4a64e979f"}`), 0600)
		os.WriteFile("./etcdCacheData_#sn#function_backup", []byte("/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n"), 0600)
		err = ew.restoreCacheFromFile()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(len(ew.ResultChan), convey.ShouldEqual, 0)
		_, errStatMeta = os.Stat("./etcdCacheMeta_#sn#function")
		_, errStatData = os.Stat("./etcdCacheData_#sn#function")
		_, errStatBackup = os.Stat("./etcdCacheData_#sn#function_backup")
		convey.So(os.IsNotExist(errStatMeta), convey.ShouldEqual, true)
		convey.So(os.IsNotExist(errStatData), convey.ShouldEqual, true)
		convey.So(os.IsNotExist(errStatBackup), convey.ShouldEqual, true)
	})
	convey.Convey("dataFile exist and no backupFile", t, func() {
		os.Remove("etcdCacheMeta_#sn#function")
		os.Remove("etcdCacheData_#sn#function")
		os.Remove("etcdCacheData_#sn#function_backup")
		// dataFile mismatches with metaFile
		os.WriteFile("./etcdCacheMeta_#sn#function", []byte(`{"revision":100,"cacheMD5":"4f4449c598ec58854d7104c4a64e979f"}`), 0600)
		os.WriteFile("./etcdCacheData_#sn#function", []byte("/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n"), 0600)
		err := ew.restoreCacheFromFile()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(len(ew.ResultChan), convey.ShouldEqual, 0)
		_, errStatMeta := os.Stat("./etcdCacheMeta_#sn#function")
		_, errStatData := os.Stat("./etcdCacheData_#sn#function")
		_, errStatBackup := os.Stat("./etcdCacheData_#sn#function_backup")
		convey.So(os.IsNotExist(errStatMeta), convey.ShouldEqual, true)
		convey.So(os.IsNotExist(errStatData), convey.ShouldEqual, true)
		convey.So(os.IsNotExist(errStatBackup), convey.ShouldEqual, true)
		// dataFile matches with metaFile
		os.WriteFile("./etcdCacheMeta_#sn#function", []byte(`{"revision":100,"cacheMD5":"03d9ff29f229e0123e427a1c84ad5afb"}`), 0600)
		os.WriteFile("./etcdCacheData_#sn#function", []byte("/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\nThisIsInvalidData\n"), 0600)
		err = ew.restoreCacheFromFile()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(ew.ResultChan), convey.ShouldEqual, 1)
		event := <-ew.ResultChan
		convey.So(event, convey.ShouldResemble, &Event{
			Rev:   100,
			Key:   "/sn/function/123/hello/latest",
			Value: []byte(`{"name":"hello","version":"latest"}`),
		})
	})
	os.Remove("etcdCacheMeta_#sn#function")
	os.Remove("etcdCacheData_#sn#function")
	os.Remove("etcdCacheData_#sn#function_backup")
}

func newEtcdWatcher() *EtcdWatcher {
	return &EtcdWatcher{
		key:       "/sn/function",
		CacheChan: make(chan *Event, 10),
		configCh:  make(chan struct{}, 1),
		cacheConfig: EtcdCacheConfig{
			EnableCache:    true,
			PersistPath:    "./",
			FlushInterval:  10,
			FlushThreshold: 10,
		},
	}
}
