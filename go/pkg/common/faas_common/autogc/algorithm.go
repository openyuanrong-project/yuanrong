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

package autogc

// Algorithm is algorithm for adjusting GOGC
type Algorithm interface {
	Init(totalMemory, threshold uint64)
	NextGOGC(currentMemory uint64, preGOGC int) int
}

const (
	defaultMaxGOGC   = 500
	defaultMinGCStep = 50 * MB
	percent          = 100
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// DefaultAlg defines default algorithm to adjust GOGC
// when current memory <= threshold, it will adjust GOGC to match threshold, but not above defaultMaxGOGC (500)
// when current memory > threshold, it will adjust GOGC so that GC will trigger every defaultMinGCStep (50MB) heap alloc
type DefaultAlg struct {
	total     uint64
	threshold uint64
	maxGOGC   int
}

// Init initializes alg with total memory and memory threshold
func (da *DefaultAlg) Init(total, threshold uint64) {
	da.total = total
	da.threshold = threshold
	da.maxGOGC = defaultMaxGOGC
}

// NextGOGC calculates appropriated GOGC with current memory and previous GOGC
func (da *DefaultAlg) NextGOGC(currentMemory uint64, preGOGC int) int {
	if da.threshold >= currentMemory+defaultMinGCStep {
		return min(da.maxGOGC, int(percent*(float64(da.threshold)/float64(currentMemory)-1.0)))
	}
	return int(percent * defaultMinGCStep / currentMemory)
}
