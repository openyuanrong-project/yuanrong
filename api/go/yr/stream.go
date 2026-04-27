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

/*
Package yr
This package encapsulates all cgo invocations.
*/
package yr

import "C"
import (
	"sync"
	"unsafe"

	"yuanrong.org/kernel/runtime/libruntime/api"
)

// SubscriptionType subscribe type
type SubscriptionType int

// Element data to be sent or receive
type Element struct {
	Data []byte
	Size int
	Id   int
}

// ProducerConf represents configuration information for a producer.
type ProducerConf struct {
	DelayFlushTime int
	PageSize       int
	MaxStreamSize  int
}

// Producer struct represents a producer of streaming data.
type Producer struct {
	producer api.StreamProducer
	mutex    *sync.RWMutex
}

// Consumer struct represents a consumer of streaming data.
type Consumer struct {
	consumer api.StreamConsumer
	mutex    *sync.RWMutex
}

// Send sends an element.
// This method can be used to send data to consumers.
func (producer *Producer) Send(data []byte) error {
	element := api.Element{
		Size: uint64(len(data)),
		Ptr:  (*uint8)(unsafe.Pointer(&data[0])),
	}
	producer.mutex.Lock()
	defer producer.mutex.Unlock()
	return producer.producer.Send(element)
}

// Close signals the producer to stop accepting new data and automatically flushes
// any pending data in the buffer. Once closed, the producer is no longer available.
func (producer *Producer) Close() error {
	producer.mutex.Lock()
	defer producer.mutex.Unlock()
	return producer.producer.Close()
}

// Receive retrieves data from the consumer with an optional timeout.
// Parameters:
// - expectNum: The expected number of elements to receive.
// - timeoutMs: Maximum time in milliseconds to wait for data before timing out.
// Returns:
// - []api.Element: The received data.
// - error: nil if data was received within the timeout, error otherwise.
func (consumer *Consumer) Receive(expectNum, timeoutMs uint32) ([]*Element, error) {
	datas, err := consumer.consumer.ReceiveExpectNum(expectNum, timeoutMs)
	if err != nil {
		return make([]*Element, 0), err
	}
	return consumer.receiveArr(datas)
}

func (consumer *Consumer) receiveArr(datas []api.Element) ([]*Element, error) {
	res := make([]*Element, 0, len(datas))
	for _, data := range datas {
		element := Element{
			Data: C.GoBytes(unsafe.Pointer(data.Ptr), C.int(data.Size)),
			Size: int(data.Size),
			Id:   int(data.Id)}
		res = append(res, &element)
	}
	return res, nil
}

// Ack confirms that the consumer has completed processing the element identified by elementID.
// This function signals to other workers whether the consumer has finished processing the element.
// If all consumers have acknowledged processing the element, it triggers internal memory reclamation
// for the corresponding page.
// Parameters:
// - elementID: The identifier of the element that has been consumed.
func (consumer *Consumer) Ack(elementId uint64) error {
	return consumer.consumer.Ack(elementId)
}

// Close closes the consumer, unsubscribing it from further data consumption.
// This method also acknowledges any unacknowledged elements on the consumer,
// ensuring that they are marked as processed before shutting down.
func (consumer *Consumer) Close() error {
	return consumer.consumer.Close()
}
