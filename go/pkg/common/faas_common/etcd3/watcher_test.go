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

package etcd3

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"

	mockUtils "yuanrong.org/kernel/pkg/common/faas_common/utils"
)

var watchChan clientv3.WatchChan
var resultCh chan *Event

type EtcdWatcherMock struct {
}

func (e EtcdWatcherMock) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	watchChan = make(chan clientv3.WatchResponse, 1)
	return watchChan
}

func (e EtcdWatcherMock) RequestProgress(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (e EtcdWatcherMock) Close() error {
	//TODO implement me
	panic("implement me")
}

func TestNewEtcdWatcher(t *testing.T) {
	prefix := ""
	filter := func(event *Event) bool { return true }
	handler := func(event *Event) {}
	stopCh := make(chan struct{})

	convey.Convey("Test NewEtcdWatcher", t, func() {

		convey.Convey("Test NewEtcdWatcher for success", func() {
			etcdClient := GetRouterEtcdClient()
			watcher := NewEtcdWatcher(prefix, filter, handler, stopCh, etcdClient)
			convey.So(watcher, convey.ShouldNotBeNil)
		})
	})
}

func TestEtcdList(t *testing.T) {
	convey.Convey("StartList", t, func() {
		stopCh := make(chan struct{})
		kv := &KvMock{}
		client := &clientv3.Client{KV: kv}
		etcdClient := &EtcdClient{Client: client, clientExitCh: make(chan struct{}), etcdStatusNow: true, cond: sync.NewCond(&sync.Mutex{})}
		resultCh = make(chan *Event, 2)
		watcher := NewEtcdWatcher("/xxx", etcdFilter, etcdHandler, stopCh, etcdClient)
		defer gomonkey.ApplyMethod(reflect.TypeOf(kv), "Get",
			func(_ *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				c := &clientv3.GetResponse{
					Header: &etcdserverpb.ResponseHeader{},
					Kvs:    []*mvccpb.KeyValue{{Key: []byte("/xxx1"), Value: []byte("value1")}},
				}
				c.Header.Revision = 1
				return c, nil
			}).Reset()
		watcher.StartList()
		event := <-resultCh
		convey.So(event.Type, convey.ShouldEqual, PUT)
		convey.So(event.Key, convey.ShouldEqual, "/xxx1")
		convey.So(string(event.Value), convey.ShouldEqual, "value1")
		event = <-resultCh
		convey.So(event.Type, convey.ShouldEqual, SYNCED)
		close(stopCh)
	})
}

func etcdFilter(event *Event) bool {
	return false
}

func etcdHandler(event *Event) {
	resultCh <- event
}

func erFail(t *testing.T) {
	convey.Convey("failed to watch", t, func() {
		convey.Convey("no connection with etcd", func() {
			stopCh := make(chan struct{})
			etcdClient := &EtcdClient{clientExitCh: make(chan struct{}), cond: sync.NewCond(&sync.Mutex{})}
			watcher := NewEtcdWatcher("/xxx", etcdFilter, etcdHandler, stopCh, etcdClient)
			watcher.StartWatch()
			watcher.watcher.cond.Broadcast()
			close(stopCh)
		})
		convey.Convey("recover watcher", func() {
			exitCh := make(chan struct{}, 1)
			stopCh := make(chan struct{}, 1)
			etcdClient := &EtcdClient{clientExitCh: exitCh, cond: sync.NewCond(&sync.Mutex{})}
			etcdClient.etcdStatusNow = true
			e := &EtcdWatcher{watcher: etcdClient, resultChanWG: &sync.WaitGroup{}, stopCh: stopCh, ResultChan: make(chan *Event)}
			e.resultChanWG.Add(1)
			go e.StartWatch()
			exitCh <- struct{}{}
			etcdClient.etcdStatusNow = true
			time.Sleep(1 * time.Second)
			close(stopCh)
			e.watcher.cond.Broadcast()
			e.resultChanWG.Done()
			_, ok := <-e.ResultChan
			convey.So(ok, convey.ShouldEqual, false)
		})
	})
}

func TestEtcdWatcher(t *testing.T) {
	stopCh := make(chan struct{})
	kv := &KvMock{}
	client := &clientv3.Client{KV: kv}
	etcdClient := &EtcdClient{Client: client, clientExitCh: make(chan struct{}), etcdStatusNow: true, cond: sync.NewCond(&sync.Mutex{})}
	resultCh = make(chan *Event, 1)
	watcher := NewEtcdWatcher("/xxx", etcdFilter, etcdHandler, stopCh, etcdClient)
	watcher.initialRev = 1
	patches := []*gomonkey.Patches{
		gomonkey.ApplyMethod(reflect.TypeOf(kv), "Get",
			func(_ *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				c := &clientv3.GetResponse{
					Header: &etcdserverpb.ResponseHeader{},
					Kvs:    []*mvccpb.KeyValue{},
				}
				c.Header.Revision = 1
				return c, nil
			}),
		gomonkey.ApplyFunc(clientv3.NewWatcher, func(c *clientv3.Client) clientv3.Watcher {
			return &EtcdWatcherMock{}
		}),
	}
	defer func() {
		for _, patch := range patches {
			patch.Reset()
		}
	}()
	convey.Convey("watch etcd", t, func() {
		go watcher.StartWatch()
		e := &Event{
			Type:      PUT,
			Key:       "/xxx",
			Value:     []byte("test"),
			PrevValue: nil,
			Rev:       0,
		}
		time.Sleep(500 * time.Millisecond)
		watcher.sendEvent(e)
		close(stopCh)
		event := <-resultCh
		convey.So(event, convey.ShouldEqual, e)
	})
}

type fakeKV struct {
	cache map[string]string
}

func (f *fakeKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	f.cache[key] = val
	return nil, nil
}

func (f *fakeKV) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	delete(f.cache, key)
	return nil, nil
}

func (f *fakeKV) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	return clientv3.OpResponse{}, nil
}

func (f *fakeKV) Txn(ctx context.Context) clientv3.Txn {
	return nil
}

func (f *fakeKV) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	return nil, nil
}

func (f *fakeKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	if _, ok := f.cache[key]; !ok {
		return nil, fmt.Errorf("Doesn't exist")
	}
	return &clientv3.GetResponse{Count: 1, Kvs: []*mvccpb.KeyValue{
		&mvccpb.KeyValue{Value: []byte(f.cache[key])},
	}}, nil
}

func TestOptEtcd(t *testing.T) {
	ew := &EtcdWatcher{
		watcher: &EtcdClient{
			Client: &clientv3.Client{},
			cond:   sync.NewCond(&sync.Mutex{}),
		},
	}
	fakeKv := &fakeKV{cache: map[string]string{}}
	defer gomonkey.ApplyFunc(clientv3.NewKV, func(c *clientv3.Client) clientv3.KV {
		return fakeKv
	}).Reset()
	etcdCtx := EtcdCtxInfo{
		Cancel: func() {},
	}
	key1 := "etcdKey"
	val1 := "etcdValue"
	key2 := "etcdKey2"
	val2 := "etcdValue2"

	convey.Convey("Put", t, func() {
		err := ew.watcher.Put(etcdCtx, key1, val1)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("GetResponse", t, func() {
		resp, err := ew.watcher.GetResponse(etcdCtx, key1)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldResemble, &clientv3.GetResponse{Count: 1, Kvs: []*mvccpb.KeyValue{
			&mvccpb.KeyValue{Value: []byte(val1)},
		}})
	})
	convey.Convey("Delete", t, func() {
		err := ew.watcher.Delete(etcdCtx, key1)
		convey.So(err, convey.ShouldBeNil)

		resp, err := ew.watcher.GetValues(etcdCtx, key1)
		convey.So(err.Error(), convey.ShouldEqual, "Doesn't exist")
		convey.So(resp, convey.ShouldBeNil)
	})
	convey.Convey("GetValues", t, func() {
		err := ew.watcher.Put(etcdCtx, key2, val2)
		convey.So(err, convey.ShouldBeNil)
		resp, err := ew.watcher.GetValues(etcdCtx, key2)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldResemble, []string{val2})
	})
}

func TestCreateEtcdCtxInfoWithTimeout(t *testing.T) {
	convey.Convey("CreateEtcdCtxInfoWithTimeout", t, func() {
		ctxInfoWithTimeout := CreateEtcdCtxInfoWithTimeout(context.TODO(), time.Second)
		convey.So(ctxInfoWithTimeout, convey.ShouldNotBeNil)
	})
}

func Test_run(t *testing.T) {
	convey.Convey("run", t, func() {
		defer gomonkey.ApplyFunc(clientv3.NewWatcher, func(c *clientv3.Client) clientv3.Watcher {
			return &EtcdWatcherMock{}
		}).Reset()
		convey.Convey("the channel received the result may be closed", func() {
			kv := &KvMock{}
			client := &clientv3.Client{KV: kv}
			etcdClient := &EtcdClient{Client: client, clientExitCh: make(chan struct{}), etcdStatusNow: true, cond: sync.NewCond(&sync.Mutex{})}
			receiveCh := make(chan *Event, 1)
			stopCh := make(chan struct{})
			e := &EtcdWatcher{watcher: etcdClient, ResultChan: receiveCh, stopCh: stopCh}
			e.initialRev = 1
			watchCh := make(chan clientv3.WatchResponse, 1)
			callCount := 0
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdWatcherMock{}), "Watch",
				func(e *EtcdWatcherMock, ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
					callCount++
					if callCount == 1 {
						return watchCh
					}
					return make(chan clientv3.WatchResponse, 1)
				}).Reset()
			defer gomonkey.ApplyMethod(reflect.TypeOf(e.watcher), "GetEtcdStatusNow", func(e *EtcdClient) bool {
				return false
			}).Reset()
			closeCount := 0
			defer gomonkey.ApplyMethod(reflect.TypeOf(kv), "Get",
				func(_ *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
					closeCount++
					return &clientv3.GetResponse{Header: &etcdserverpb.ResponseHeader{Revision: int64(closeCount)}}, nil
				}).Reset()
			go e.run()
			time.Sleep(100 * time.Millisecond)
			close(watchCh)
			time.Sleep(100 * time.Millisecond)
			convey.So(callCount, convey.ShouldEqual, 2)
			close(stopCh)
		})
		convey.Convey("sendEvent", func() {
			eventCh := make(chan clientv3.WatchResponse, 1)
			defer gomonkey.ApplyMethod(reflect.TypeOf(&EtcdWatcherMock{}), "Watch",
				func(e *EtcdWatcherMock, ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
					return eventCh
				}).Reset()
			stopCh := make(chan struct{})
			receiveCh := make(chan *Event, defaultEventChanSize)
			kv := &KvMock{}
			client := &clientv3.Client{KV: kv}
			etcdClient := &EtcdClient{Client: client, clientExitCh: make(chan struct{}), etcdStatusNow: true, cond: sync.NewCond(&sync.Mutex{})}
			e := &EtcdWatcher{watcher: etcdClient, stopCh: stopCh, ResultChan: receiveCh}
			e.initialRev = 1
			closeCount := 0
			defer gomonkey.ApplyMethod(reflect.TypeOf(kv), "Get",
				func(_ *KvMock, ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
					closeCount++
					return &clientv3.GetResponse{Header: &etcdserverpb.ResponseHeader{Revision: int64(closeCount)}}, nil
				}).Reset()
			go e.run()
			eventCh <- clientv3.WatchResponse{Events: []*clientv3.Event{{Kv: &mvccpb.KeyValue{Key: []byte("key1"), Value: []byte("value1")}}}}
			event := <-receiveCh
			convey.So(event.Key, convey.ShouldEqual, "key1")
			convey.So(string(event.Value), convey.ShouldEqual, "value1")
			close(stopCh)
		})
		convey.Convey("enable cache", func() {
			os.Remove("etcdCacheMeta_#sn#function")
			os.Remove("etcdCacheData_#sn#function")
			os.Remove("etcdCacheData_#sn#function_backup")
			defer gomonkey.ApplyMethod(reflect.TypeOf(&KvMock{}), "Get", func(_ *KvMock, ctx context.Context,
				key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return nil, errors.New("some error")
			}).Reset()
			stopCh := make(chan struct{})
			receiveCh := make(chan *Event, defaultEventChanSize)
			cacheCh := make(chan *Event, defaultEventChanSize)
			kv := &KvMock{}
			client := &clientv3.Client{KV: kv}
			etcdClient := &EtcdClient{Client: client, clientExitCh: make(chan struct{}), etcdStatusNow: true}
			e := &EtcdWatcher{watcher: etcdClient, stopCh: stopCh, ResultChan: receiveCh, CacheChan: cacheCh,
				key: "/sn/function", cacheConfig: EtcdCacheConfig{
					EnableCache:   true,
					PersistPath:   "./",
					FlushInterval: 10,
				}}
			os.WriteFile("./etcdCacheMeta_#sn#function", []byte(`{"revision":101,"cacheMD5":"5642747b723c9497e2b7324b49fb0513"}`), 0600)
			os.WriteFile("./etcdCacheData_#sn#function", []byte("/sn/function/123/goodbye/latest|101|{\"name\":\"goodbye\",\"version\":\"latest\"}\n/sn/function/123/hello/latest|100|{\"name\":\"hello\",\"version\":\"latest\"}\n"), 0600)
			go e.run()
			time.Sleep(500 * time.Millisecond)
			convey.So(len(e.ResultChan), convey.ShouldEqual, 3)
			event1 := <-e.ResultChan
			event2 := <-e.ResultChan
			convey.So(event1, convey.ShouldResemble, &Event{
				Rev:   101,
				Type:  PUT,
				Key:   "/sn/function/123/goodbye/latest",
				Value: []byte(`{"name":"goodbye","version":"latest"}`),
			})
			convey.So(event2, convey.ShouldResemble, &Event{
				Rev:   100,
				Type:  PUT,
				Key:   "/sn/function/123/hello/latest",
				Value: []byte(`{"name":"hello","version":"latest"}`),
			})
			close(stopCh)
		})
	})
	os.Remove("etcdCacheMeta_#sn#function")
	os.Remove("etcdCacheData_#sn#function")
	os.Remove("etcdCacheData_#sn#function_backup")
}

func TestEtcdWatcher_EtcdHistory(t *testing.T) {
	type fields struct {
		filter       EtcdWatcherFilter
		handler      EtcdWatcherHandler
		watcher      *EtcdClient
		ResultChan   chan *Event
		resultChanWG *sync.WaitGroup
		stopCh       <-chan struct{}
		key          string
		initialRev   int64
	}
	type args struct {
		revision int64
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		patchesFunc mockUtils.PatchesFunc
	}{
		{"case1", fields{watcher: &EtcdClient{cond: sync.NewCond(&sync.Mutex{})}}, args{revision: -1}, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(clientv3.NewWatcher, func(c *clientv3.Client) clientv3.Watcher {
					return &EtcdWatcherMock{}
				}),
				gomonkey.ApplyMethod(reflect.TypeOf(&EtcdWatcherMock{}), "Watch",
					func(e *EtcdWatcherMock, ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
						ch := make(chan clientv3.WatchResponse, 1)
						go close(ch)
						return ch
					})})
			return patches
		}},
		{"case2 watch chan nil", fields{watcher: &EtcdClient{cond: sync.NewCond(&sync.Mutex{})}}, args{revision: -1}, func() mockUtils.PatchSlice {
			patches := mockUtils.InitPatchSlice()
			patches.Append(mockUtils.PatchSlice{
				gomonkey.ApplyFunc(clientv3.NewWatcher, func(c *clientv3.Client) clientv3.Watcher {
					return &EtcdWatcherMock{}
				}),
				gomonkey.ApplyMethod(reflect.TypeOf(&EtcdWatcherMock{}), "Watch",
					func(e *EtcdWatcherMock, ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
						return nil
					})})
			return patches
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.patchesFunc()
			ew := &EtcdWatcher{
				filter:       tt.fields.filter,
				handler:      tt.fields.handler,
				watcher:      tt.fields.watcher,
				ResultChan:   tt.fields.ResultChan,
				resultChanWG: tt.fields.resultChanWG,
				stopCh:       tt.fields.stopCh,
				key:          tt.fields.key,
				initialRev:   tt.fields.initialRev,
			}
			ew.EtcdHistory(tt.args.revision)
			patches.ResetAll()
		})
	}
}

func TestEtcdClient_Get(t *testing.T) {
	e := &EtcdClient{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	ctxInfo := EtcdCtxInfo{Ctx: ctx, Cancel: cancel}

	key := "test-key"
	response := &clientv3.GetResponse{}

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFuncReturn(clientv3.NewKV, &clientv3.Client{})
	patches.ApplyMethodReturn(&clientv3.Client{}, "Get", response, nil)

	got, err := e.Get(ctxInfo, key)

	assert.NoError(t, err)
	assert.Equal(t, response, got)
}
