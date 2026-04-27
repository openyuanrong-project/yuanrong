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

// Package loadbalance provides consistent hash algorithm
package loadbalance

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LBTestSuite struct {
	suite.Suite
	LoadBalance
	lbType    LBType
	m         sync.RWMutex
	emptyNode interface{}
}

func (lbs *LBTestSuite) SetupSuite() {
	switch lbs.lbType {
	case RoundRobinNginx, RoundRobinLVS:
		lbs.emptyNode = nil
	case ConsistentHashGeneric:
		lbs.emptyNode = ""
	default:
		lbs.emptyNode = ""
	}
}

func (lbs *LBTestSuite) SetupTest() {
	lbs.m = sync.RWMutex{}
	lbs.LoadBalance = LBFactory(lbs.lbType)
}

func (lbs *LBTestSuite) TearDownTest() {
	lbs.LoadBalance = nil
}

func (lbs *LBTestSuite) AddToLB(workerInstance interface{}, weight int) {
	switch lbs.lbType {
	case RoundRobinNginx, RoundRobinLVS:
		lbs.m.Lock()
		lbs.Add(workerInstance, weight)
		lbs.Reset()
		lbs.m.Unlock()
	case ConsistentHashGeneric:
		lbs.Add(workerInstance, 0)
	default:
	}
}

func (lbs *LBTestSuite) DelFromLB(workerInstance interface{}) {
	switch lbs.lbType {
	case RoundRobinNginx, RoundRobinLVS:
		lbs.m.Lock()
		lbs.Remove(workerInstance)
		lbs.Reset()
		defer lbs.m.Unlock()
	case ConsistentHashGeneric:
		lbs.Remove(workerInstance)
	default:
	}
}

func (lbs *LBTestSuite) TestAdd() {
	lbs.AddToLB("new-node-01", 0)
	lbs.AddToLB("new-node-01", 1) // test duplicate
	lbs.AddToLB("new-node-02", 2)
	lbs.AddToLB("new-node-03", 5)
	lbs.AddToLB("", 6)
	lbs.AddToLB(nil, 4)
	next := lbs.Next("fn-urn-01", true)
	assert.NotEqual(lbs.T(), lbs.emptyNode, next)
	lbs.Reset()
	next = lbs.Next("fn-urn-01", true)
	assert.NotEqual(lbs.T(), lbs.emptyNode, next)
}

func (lbs *LBTestSuite) TestNext() {
	var wg sync.WaitGroup
	next := lbs.Next("fn-urn-01", false)
	assert.Equal(lbs.T(), lbs.emptyNode, next)

	lbs.AddToLB("new-node-01", 5)
	next = lbs.Next("fn-urn-01", true)
	assert.Equal(lbs.T(), "new-node-01", next)

	for i := 2; i < 5; i++ {
		wg.Add(1)
		go func(i int, wg *sync.WaitGroup) {
			lbs.AddToLB("new-node-0"+strconv.Itoa(i), 5)
			wg.Done()
		}(i, &wg)
	}
	wg.Wait()
	next = lbs.Next("fn-urn-01", true)
	assert.NotEqual(lbs.T(), lbs.emptyNode, next)
}

func (lbs *LBTestSuite) TestRemove() {
	var wg sync.WaitGroup
	for i := 1; i < 5; i++ {
		wg.Add(1)
		go func(i int, wg *sync.WaitGroup) {
			lbs.AddToLB("new-node-0"+strconv.Itoa(i), 5)
			wg.Done()
		}(i, &wg)
	}
	wg.Wait()
	for i := 1; i < 4; i++ {
		wg.Add(1)
		go func(i int, wg *sync.WaitGroup) {
			lbs.DelFromLB("new-node-0" + strconv.Itoa(i))
			wg.Done()
		}(i, &wg)
	}
	wg.Wait()
	next := lbs.Next("fn-urn-01", true)
	assert.Equal(lbs.T(), "new-node-04", next)
}

func (lbs *LBTestSuite) TestRemoveAll() {
	var wg sync.WaitGroup
	for i := 1; i < 5; i++ {
		wg.Add(1)
		go func(i int, wg *sync.WaitGroup) {
			lbs.Add("new-node-0"+strconv.Itoa(i), 5)
			wg.Done()
		}(i, &wg)
	}
	wg.Wait()
	lbs.RemoveAll()
	next := lbs.Next("fn-urn-01", true)
	assert.Equal(lbs.T(), lbs.emptyNode, next)
}

func TestLBTestSuite(t *testing.T) {
	suite.Run(t, &LBTestSuite{lbType: ConsistentHashGeneric})
}

func TestSimpleCHGeneric_Add(t *testing.T) {
	con := NewSimpleCHGeneric()
	con.Add("n1", 0)
	con.Add("n2", 0)

	next := con.Next("n1", false)
	assert.Equal(t, "n2", next)

	con.Remove("n2")
	con.RemoveAll()
	con.Reset()

	next = con.Next("n1", false)
	assert.Equal(t, "", next)
}

func TestLimiterCHGeneric(t *testing.T) {
	limiter := NewLimiterCHGeneric(5 * time.Second)
	limiter.Add("n1", 0)
	limiter.Add("n2", 0)
	limiter.Add("n3", 0)

	next := limiter.Next("func1", false)
	assert.Equal(t, "n1", next)

	limiter.SetStain("func1", "n1")

	next = limiter.Next("func1", false)
	assert.Equal(t, "n3", next)

	limiter.SetStain("func1", "n3")

	next = limiter.Next("func1", false)
	assert.Equal(t, "n2", next)

	limiter.SetStain("func1", "n2")

	next = limiter.Next("func1", false)
	assert.Equal(t, nil, next)

	time.Sleep(5 * time.Second)

	next = limiter.Next("func1", false)
	assert.Equal(t, "n2", next)

	limiter.Remove("n2")

	next = limiter.Next("func1", false)
	assert.Equal(t, "n1", next)

	limiter.RemoveAll()

	next = limiter.Next("func1", false)
	assert.Equal(t, "", next)

	limiter.Reset()
}

func TestLBFactory(t *testing.T) {
	convey.Convey("LBFactory", t, func() {
		convey.Convey("RoundRobinNginx", func() {
			factory := LBFactory(LBType(0))
			convey.So(factory, convey.ShouldNotBeNil)
		})
		convey.Convey("ConsistentHashGeneric", func() {
			factory := LBFactory(LBType(2))
			convey.So(factory, convey.ShouldNotBeNil)
		})
		convey.Convey("ConcurrentConsistentHashGeneric", func() {
			factory := LBFactory(LBType(3))
			convey.So(factory, convey.ShouldNotBeNil)
		})
		convey.Convey("default", func() {
			factory := LBFactory(LBType(1))
			convey.So(factory, convey.ShouldNotBeNil)
		})
	})
}
