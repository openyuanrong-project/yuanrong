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

// Package yr for actor
package yr

import (
	"fmt"
	"sync"

	"github.com/vmihailenco/msgpack"

	"yuanrong.org/kernel/runtime/libruntime/api"
)

// IsInit Used to determine whether initialization has been completed
var IsInit = false

// Init Initialization entry, which is used to initialize the data system and function system.
func Init(yrConfig *Config) (*ClientInfo, error) {
	return InitWithFlags(yrConfig, &FlagsConfig{})
}

// InitWithFlags Init with flags args.
func InitWithFlags(yrConfig *Config, yrFlagsConfig *FlagsConfig) (*ClientInfo, error) {
	if IsInit {
		return GetConfigManager().GetClientInfo(), nil
	}
	GetConfigManager().Init(yrConfig, yrFlagsConfig)
	fmt.Println(GetConfigManager().Config)
	runtime := &ClusterModeRuntime{}
	if err := runtime.Init(); err != nil {
		return nil, err
	}
	GetRuntimeHolder().Init(runtime)
	IsInit = true
	return GetConfigManager().GetClientInfo(), nil
}

// Finalize release resources
func Finalize(all bool) {
	GetRuntimeHolder().GetRuntime().Finalize()
}

// Get to get obj from data system.
func Get[T any](objRef *ObjectRef, timeoutMs int) (res T, err error) {
	err = Wait(objRef, timeoutMs/1000)
	if err != nil {
		return
	}
	return get[T](objRef, timeoutMs)
}

func get[T any](objRef *ObjectRef, timeoutMs int) (res T, err error) {
	objIds := []string{objRef.objId}
	data, err := GetRuntimeHolder().GetRuntime().Get(objIds, timeoutMs)
	if err != nil {
		stacks := getStackTraceInfos()
		err = api.NewErrorInfoWithStackInfo(err, stacks)
		return
	}

	result := new(T)
	err = msgpack.Unmarshal(data[0], result)
	if err != nil {
		return
	}

	return *result, nil
}

// BatchGet to get objs from data system.
func BatchGet[T any](objRefs []*ObjectRef, timeoutMs int, allowPartial bool) (res []T, err error) {
	if len(objRefs) == 0 {
		return
	}

	objIds := make([]string, 0, len(objRefs))
	for _, ref := range objRefs {
		objIds = append(objIds, ref.objId)
	}
	datas, err := GetRuntimeHolder().GetRuntime().Get(objIds, timeoutMs)
	if err != nil {
		return
	}

	res = make([]T, 0, len(objRefs))
	for _, data := range datas {
		result := new(T)
		err = msgpack.Unmarshal(data, result)
		if err != nil {
			return
		}
		res = append(res, *result)
	}
	return
}

// Put put obj data to data system.
func Put(val any) (*ObjectRef, error) {
	data, err := msgpack.Marshal(val)
	if err != nil {
		return nil, err
	}

	ref := NewObjectRef()

	objectIds := []string{ref.objId}
	_, err = GetRuntimeHolder().GetRuntime().GIncreaseRef(objectIds, make([]string, 0)...)
	if err != nil {
		return nil, err
	}

	err = GetRuntimeHolder().GetRuntime().Put(ref.objId, data, api.PutParam{}, make([]string, 0)...)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

// Wait until result return or timeout
func Wait(objRef *ObjectRef, timeout int) error {
	objIds := []string{objRef.objId}
	readyIds, _, errs := GetRuntimeHolder().GetRuntime().Wait(objIds, 1, timeout)
	if errs != nil {
		return waitErr(errs)
	}

	if len(readyIds) != 1 {
		return fmt.Errorf("wait failed")
	}
	return nil
}

func waitErr(errs map[string]error) error {
	for _, v := range errs {
		stacks := getStackTraceInfos()
		return api.NewErrorInfoWithStackInfo(v, stacks)
	}
	return fmt.Errorf("wait failed")
}

// WaitNum specify a specific count to wait
func WaitNum(objRefs []*ObjectRef, waitNum uint64, timeout int) (readyObjRefs, unreadyObjRefs []*ObjectRef, err error) {
	objIds := make([]string, 0, len(objRefs))
	objIdMap := make(map[string]*ObjectRef)
	for _, ref := range objRefs {
		objIds = append(objIds, ref.objId)
		objIdMap[ref.objId] = ref
	}
	readyIds, unReadyIds, errors := GetRuntimeHolder().GetRuntime().Wait(objIds, waitNum, timeout)
	if errors != nil {
		err = fmt.Errorf("wait num failed")
		return
	}

	readyObjRefs = make([]*ObjectRef, 0, len(readyIds))
	unreadyObjRefs = make([]*ObjectRef, 0, len(unReadyIds))
	for _, id := range readyIds {
		readyObjRefs = append(readyObjRefs, objIdMap[id])
	}

	for _, id := range unReadyIds {
		unreadyObjRefs = append(unreadyObjRefs, objIdMap[id])
	}
	return
}

// SetKV save binary data to the data system.
func SetKV(key, value string) error {
	return GetRuntimeHolder().GetRuntime().KVSet(key, []byte(value), api.SetParam{})
}

// GetKV get binary data from data system.
func GetKV(key string, timeoutMs uint) (string, error) {
	data, err := GetRuntimeHolder().GetRuntime().KVGet(key, timeoutMs)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetKVs get multi binary data from data system.
func GetKVs(keys []string, timeoutMs uint, allowPartial bool) ([]string, error) {
	datas, err := GetRuntimeHolder().GetRuntime().KVGetMulti(keys, timeoutMs)
	if err != nil {
		return make([]string, 0), err
	}

	res := make([]string, 0, len(keys))
	for _, data := range datas {
		res = append(res, string(data))
	}
	return res, nil
}

// DelKV del data from data system.
func DelKV(key string) error {
	return GetRuntimeHolder().GetRuntime().KVDel(key)
}

// DelKVs del multi data from data system.
func DelKVs(keys []string) ([]string, error) {
	return GetRuntimeHolder().GetRuntime().KVDelMulti(keys)
}

// CreateProducer creates and return Producer
func CreateProducer(streamName string, opt ProducerConf) (*Producer, error) {
	conf := api.ProducerConf{
		DelayFlushTime: int64(opt.DelayFlushTime),
		PageSize:       int64(opt.PageSize),
		MaxStreamSize:  uint64(opt.MaxStreamSize),
	}
	producer, err := GetRuntimeHolder().GetRuntime().CreateProducer(streamName, conf)
	if err != nil {
		return nil, err
	}

	return &Producer{producer: producer, mutex: new(sync.RWMutex)}, nil
}

// Subscribe creates and return Consumer
func Subscribe(streamName, subscriptionName string, subscriptionType SubscriptionType) (*Consumer, error) {
	conf := api.SubscriptionConfig{
		SubscriptionName: subscriptionName,
		SubscriptionType: api.SubscriptionType(subscriptionType),
	}
	consumer, err := GetRuntimeHolder().GetRuntime().Subscribe(streamName, conf)
	if err != nil {
		return nil, err
	}

	return &Consumer{consumer: consumer, mutex: new(sync.RWMutex)}, nil
}

// DeleteStream Delete a data flow.
// When the number of global producers and consumers is 0,
// the data flow is no longer used and the metadata related to the data flow is deleted from each worker and the
// master. This function can be invoked on any host node.
func DeleteStream(streamName string) error {
	return GetRuntimeHolder().GetRuntime().DeleteStream(streamName)
}

// QueryGlobalProducersNum Specifies the flow name to query the number of all producers of the flow.
func QueryGlobalProducersNum(streamName string) (uint64, error) {
	return GetRuntimeHolder().GetRuntime().QueryGlobalProducersNum(streamName)
}

// QueryGlobalConsumersNum Specifies the flow name to query the number of all consumers of the flow.
func QueryGlobalConsumersNum(streamName string) (uint64, error) {
	return GetRuntimeHolder().GetRuntime().QueryGlobalConsumersNum(streamName)
}
