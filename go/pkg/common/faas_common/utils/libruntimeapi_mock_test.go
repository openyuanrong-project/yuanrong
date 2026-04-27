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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"yuanrong.org/kernel/runtime/libruntime/api"
)

func TestFakeLibruntimeSdkClient(t *testing.T) {
	fakeLibruntimeSdkClient := FakeLibruntimeSdkClient{}
	instanceID, err := fakeLibruntimeSdkClient.CreateInstance(api.FunctionMeta{}, []api.Arg{}, api.InvokeOptions{})
	assert.NotEqual(t, 0, len(instanceID))
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.InvokeByInstanceId(api.FunctionMeta{}, "", []api.Arg{}, api.InvokeOptions{})
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.InvokeByFunctionName(api.FunctionMeta{}, []api.Arg{}, api.InvokeOptions{})
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.AcquireInstance("", api.FunctionMeta{}, api.InvokeOptions{})
	assert.Equal(t, nil, err)

	err = fakeLibruntimeSdkClient.Kill("", 0, []byte{})
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.CreateInstanceRaw([]byte{})
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.InvokeByInstanceIdRaw([]byte{})
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.KillRaw([]byte{})
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.SaveState([]byte{})
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.LoadState("")
	assert.Equal(t, nil, err)

	err = fakeLibruntimeSdkClient.KVSet("", []byte{}, api.SetParam{})
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.KVSetWithoutKey([]byte{}, api.SetParam{})
	assert.Equal(t, nil, err)

	err = fakeLibruntimeSdkClient.KVMSetTx([]string{}, [][]byte{}, api.MSetParam{})
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.KVGet("", 1)
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.KVGetMulti([]string{}, 1)
	assert.Equal(t, nil, err)

	err = fakeLibruntimeSdkClient.KVDel("")
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.KVDelMulti([]string{})
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.CreateProducer("", api.ProducerConf{})
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.Subscribe("", api.SubscriptionConfig{})
	assert.Equal(t, nil, err)

	err = fakeLibruntimeSdkClient.DeleteStream("")
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.QueryGlobalProducersNum("")
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.QueryGlobalConsumersNum("")
	assert.Equal(t, nil, err)

	err = fakeLibruntimeSdkClient.SetTenantID("")
	assert.Equal(t, nil, err)

	err = fakeLibruntimeSdkClient.Put("", []byte{}, api.PutParam{}, "")
	assert.Equal(t, nil, err)

	err = fakeLibruntimeSdkClient.PutRaw("", []byte{}, api.PutParam{}, "")
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.Get([]string{}, 1)
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.GetRaw([]string{}, 1)
	assert.Equal(t, nil, err)

	_, _, aa := fakeLibruntimeSdkClient.Wait([]string{}, 1, 1)
	assert.Equal(t, map[string]error(map[string]error(nil)), aa)

	_, err = fakeLibruntimeSdkClient.GIncreaseRef([]string{}, "")
	assert.Equal(t, nil, err)

	_, err = fakeLibruntimeSdkClient.GDecreaseRefRaw([]string{}, "")
	assert.Equal(t, nil, err)

	bb := fakeLibruntimeSdkClient.GetFormatLogger()
	assert.Equal(t, nil, bb)

	_, err = fakeLibruntimeSdkClient.CreateClient(api.ConnectArguments{})
	assert.Equal(t, nil, err)

	err = fakeLibruntimeSdkClient.ReleaseGRefs("")
	assert.Equal(t, nil, err)

	err = fakeLibruntimeSdkClient.ReleaseGRefs("")
	assert.Equal(t, nil, err)

	credential := fakeLibruntimeSdkClient.GetCredential()
	assert.NotEqual(t, nil, credential)
}

func TestFakeStreamProducer(t *testing.T) {
	fakeStreamProducer := FakeStreamProducer{}
	err := fakeStreamProducer.Send(api.Element{})
	assert.Equal(t, nil, err)

	err = fakeStreamProducer.SendWithTimeout(api.Element{}, 1)
	assert.Equal(t, nil, err)

	err = fakeStreamProducer.Flush()
	assert.Equal(t, nil, err)

	err = fakeStreamProducer.Close()
	assert.Equal(t, nil, err)
}

func TestFakeStreamConsumer(t *testing.T) {
	fakeStreamConsumer := FakeStreamConsumer{}
	_, err := fakeStreamConsumer.ReceiveExpectNum(1, 1)
	assert.Equal(t, nil, err)

	_, err = fakeStreamConsumer.Receive(1)
	assert.Equal(t, nil, err)

	err = fakeStreamConsumer.Ack(1)
	assert.Equal(t, nil, err)

	err = fakeStreamConsumer.Close()
	assert.Equal(t, nil, err)
}
