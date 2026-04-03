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

package monitor

import (
	"bufio"
	"reflect"
	"sync"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"yuanrong.org/kernel/pkg/common/faas_common/types"
)

func TestInitMemMonitor(t *testing.T) {
	convey.Convey("TestInitMemMonitor", t, func() {
		convey.Convey("success", func() {
			patches := [...]*Patches{
				ApplyFunc(readValue, func(path string) (uint64, error) {
					return uint64(100), nil
				}),
				ApplyFunc(NewCGroupMemoryParser, func() (*Parser, error) {
					return &Parser{
						f:      nil,
						reader: bufio.NewReader(nil),
						parser: cgroupMemoryParserFunc,
					}, nil
				}),
				ApplyMethod(reflect.TypeOf(new(Parser)), "Read", func(_ *Parser) (interface{}, error) {
					return uint64(100), nil
				}),
				ApplyMethod(reflect.TypeOf(new(Parser)), "Close", func() error {
					return nil
				}),
			}
			defer func() {
				for idx := range patches {
					patches[idx].Reset()
				}
			}()
			stopCh := make(chan struct{})
			err := InitMemMonitor(stopCh)
			assert.Nil(t, err)
			assert.Equal(t, uint64(0x0), memory.monitor.used)

			time.Sleep(2 * time.Second)
			assert.NotEqual(t, uint64(0x0), memory.monitor.used)
		})
	})

}

func TestMemMonitor_Allow(t *testing.T) {
	memMonitor := &memMonitor{enable: true, threshold: 1024, used: 10}
	result := memMonitor.Allow(1000)
	assert.Equal(t, true, result)
	result = memMonitor.Allow(15)
	assert.Equal(t, false, result)
}

func TestAllowByLowerThreshold(t *testing.T) {
	memMonitor := &memMonitor{
		enable:         true,
		threshold:      200000,
		lowerThreshold: 140000,
		memMapMutex:    sync.Mutex{},
		functionMemMap: map[string]uint64{},
	}

	allow := memMonitor.Allow(100)
	assert.Equal(t, true, allow)
	allow = memMonitor.AllowByLowerThreshold("1", "1", 100)
	assert.Equal(t, true, allow)

	allow = memMonitor.Allow(100000)
	assert.Equal(t, true, allow)
	allow = memMonitor.AllowByLowerThreshold("1", "1", 100000)
	assert.Equal(t, true, allow)

	allow = memMonitor.Allow(20000)
	assert.Equal(t, true, allow)
	allow = memMonitor.AllowByLowerThreshold("2", "2", 20000)
	assert.Equal(t, true, allow)

	allow = memMonitor.Allow(30000)
	assert.Equal(t, true, allow)
	allow = memMonitor.AllowByLowerThreshold("1", "1", 30000)
	assert.Equal(t, false, allow)

	memMonitor.ReleaseFunctionMem("1", 100000)
	memMonitor.used = memMonitor.used - 100000

	allow = memMonitor.Allow(100000)
	assert.Equal(t, true, allow)
	allow = memMonitor.AllowByLowerThreshold("1", "1", 100000)
	assert.Equal(t, true, allow)

	allow = memMonitor.Allow(20000)
	assert.Equal(t, true, allow)
	allow = memMonitor.AllowByLowerThreshold("2", "2", 20000)
	assert.Equal(t, true, allow)
}

func TestSetMemoryControlConfig(t *testing.T) {
	convey.Convey("TestSetMemoryControlConfig", t, func() {
		convey.Convey("nil config", func() {
			SetMemoryControlConfig(nil)
		})
		convey.Convey("SetMemoryControlConfig", func() {
			cfg := &types.MemoryControlConfig{
				LowerMemoryPercent:     0.5,
				BodyThreshold:          1024,
				MemDetectIntervalMs:    3,
				HighMemoryPercent:      0.5,
				StatefulHighMemPercent: 0.9,
			}
			SetMemoryControlConfig(cfg)
			convey.So(*config == *cfg, convey.ShouldEqual, true)
		})
	})
}

func Test_parseValue(t *testing.T) {
	convey.Convey("parseValue", t, func() {
		v, err := parseValue("100", 10, 64)
		convey.So(v, convey.ShouldEqual, 100)
		convey.So(err, convey.ShouldBeNil)
		v, err = parseValue("-100", 10, 64)
		convey.So(v, convey.ShouldEqual, 0)
		convey.So(err, convey.ShouldBeNil)
		v, err = parseValue("1.01", 10, 64)
		convey.So(v, convey.ShouldEqual, 0)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestIsAllowByMemory(t *testing.T) {
	convey.Convey("TestIsAllowByMemory", t, func() {
		memory.monitor = &memMonitor{
			enable:         true,
			threshold:      200000,
			lowerThreshold: 140000,
			memMapMutex:    sync.Mutex{},
			functionMemMap: map[string]uint64{},
		}
		memory.monitor.decreaseMemCnt(100)

		allow := IsAllowByMemory("1", 200001, "")
		convey.So(allow, convey.ShouldBeFalse)

		allow = IsAllowByMemory("1", 100000, "")
		convey.So(allow, convey.ShouldBeTrue)

		allow = IsAllowByMemory("2", 50000, "")
		convey.So(allow, convey.ShouldBeTrue)

		allow = IsAllowByMemory("1", 20000, "")
		convey.So(allow, convey.ShouldBeFalse)
	})
}
