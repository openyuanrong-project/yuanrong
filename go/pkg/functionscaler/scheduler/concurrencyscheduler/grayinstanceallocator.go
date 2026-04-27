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

package concurrencyscheduler

import (
	"hash/crc32"
	"math"
	"sort"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/functionscaler/selfregister"
)

// InstanceOperType -
type InstanceOperType int

const (
	// Add -
	Add InstanceOperType = iota
	// Del -
	Del
	// Find -
	Find
)

const (
	percentageBase = 100
)

// HashedInstance -
type HashedInstance struct {
	InsElem *instanceElement
	hash    uint32
}

// Hasher 哈希函数接口
type Hasher interface {
	Hash(key string) uint32
}

// CRC32Hasher -
type CRC32Hasher struct{}

// Hash -
func (*CRC32Hasher) Hash(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

// GrayInstanceAllocator 灰度分配器接口
type GrayInstanceAllocator interface {
	ComputeHash(instanceKey string) uint32
	UpdateRolloutRatio(rolloutPercent int)
	Partition(instances []*HashedInstance, isGrayNode bool) (self, other []*instanceElement)
	CheckSelf(isGrayNode bool, instanceKey string) bool
	ShouldReassign(operType InstanceOperType, instanceKey string) bool
	GetRolloutRatio() float64
}

// HashBasedInstanceAllocator 灰度分配器实现
type HashBasedInstanceAllocator struct {
	hasher       Hasher
	rolloutRatio float64
	boundaryHash uint32
	maxHashValue uint32
	grayCount    int
	notGrayCount int
}

// ShouldReassign -
func (h *HashBasedInstanceAllocator) ShouldReassign(operType InstanceOperType, instanceKey string) bool {
	if !selfregister.IsRollingOut {
		return false
	}
	h.modifyCount(operType, instanceKey)
	count := h.CountFloorGrayCount(h.grayCount + h.notGrayCount)
	// 多了: (9,1)->(8,1) 10% 需要调成(9,0)
	// 少了: (19,2)->(19,1)应该重新划分成(18,2)
	if h.grayCount != count {
		return true
	} else {
		return false
	}
}

func (h *HashBasedInstanceAllocator) checkSelfByBoundary(isGrayNode bool, hashValue uint32) bool {
	// [旧节点self 70% | 新节点self 30%] 从小到大
	if isGrayNode {
		return hashValue > h.boundaryHash
	}
	return hashValue <= h.boundaryHash
}

// CheckSelf -
func (h *HashBasedInstanceAllocator) CheckSelf(isGrayNode bool, instanceKey string) bool {
	hashValue := h.hasher.Hash(instanceKey)
	return h.checkSelfByBoundary(isGrayNode, hashValue)
}

// ModifyCount -
func (h *HashBasedInstanceAllocator) modifyCount(operType InstanceOperType, instanceKey string) {
	hashValue := h.hasher.Hash(instanceKey)
	switch operType {
	case Add:
		if hashValue <= h.boundaryHash {
			h.notGrayCount++
		} else {
			h.grayCount++
		}
	case Del:
		if hashValue <= h.boundaryHash {
			h.notGrayCount--
		} else {
			h.grayCount--
		}
	default:
	}
}

// NewHashBasedInstanceAllocator 创建新的基于哈希的灰度分配器
func NewHashBasedInstanceAllocator(rolloutRatio float64) GrayInstanceAllocator {
	var hasher Hasher = &CRC32Hasher{}

	return &HashBasedInstanceAllocator{
		hasher:       hasher,
		rolloutRatio: rolloutRatio,
		maxHashValue: math.MaxUint32,
		boundaryHash: math.MaxUint32,
	}
}

// CountFloorGrayCount 向下取整 灰度数量 9*10% = 0, 19*10% = 1
func (h *HashBasedInstanceAllocator) CountFloorGrayCount(totalCount int) int {
	targetGrayCount := int(math.Floor(float64(totalCount) * h.rolloutRatio))
	return targetGrayCount
}

// Partition 调整成2个队列，并且更新内部的graycount计数和boundary
func (h *HashBasedInstanceAllocator) Partition(instances []*HashedInstance, isGrayNode bool) (self,
	other []*instanceElement) {
	total := len(instances)
	if total == 0 {
		return
	}

	targetGrayCount := h.CountFloorGrayCount(total)
	targetNotGrayCount := total - targetGrayCount

	sort.Slice(instances, func(i, j int) bool {
		return instances[i].hash > instances[j].hash
	})

	selfCap, otherCap := targetNotGrayCount, targetGrayCount
	if isGrayNode {
		selfCap, otherCap = targetGrayCount, targetNotGrayCount
	}

	self = make([]*instanceElement, 0, selfCap)
	other = make([]*instanceElement, 0, otherCap)

	for i, item := range instances {
		addToSelf := isGrayNode && (i < targetGrayCount) || !isGrayNode && (i >= targetGrayCount)
		if addToSelf {
			self = append(self, item.InsElem)
		} else {
			other = append(other, item.InsElem)
		}
	}
	log.GetLogger().Infof("partition ratio %.2f (notGrayCount,grayCount): (%d,%d)->(%d,%d). isGrayNode: %t, "+
		"(self,other): (%d, %d)", h.rolloutRatio, h.notGrayCount, h.grayCount, targetNotGrayCount, targetGrayCount,
		isGrayNode, len(self), len(other))
	// 因为旧节点判断self的时候带==, 灰度100%时候边界应为0,否则旧节点添加、删除时会判断错误
	newBoundary := h.maxHashValue
	if targetGrayCount == total {
		newBoundary = 0
	} else {
		newBoundary = instances[targetGrayCount].hash
	}

	h.boundaryHash = newBoundary
	h.notGrayCount = targetNotGrayCount
	h.grayCount = targetGrayCount
	return
}

// ComputeHash -
func (h *HashBasedInstanceAllocator) ComputeHash(instanceKey string) uint32 {
	hashValue := h.hasher.Hash(instanceKey)
	return hashValue
}

// UpdateRolloutRatio -
func (h *HashBasedInstanceAllocator) UpdateRolloutRatio(rolloutPercent int) {
	log.GetLogger().Infof("update gray allocator ratio from %.2f to %d", h.rolloutRatio, rolloutPercent)
	h.rolloutRatio = float64(rolloutPercent) / percentageBase
	return
}

// GetRolloutRatio -
func (h *HashBasedInstanceAllocator) GetRolloutRatio() float64 {
	return h.rolloutRatio
}
