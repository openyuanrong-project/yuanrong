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

// Package autogc adjusts GOGC automatically inspired by
// https://eng.uber.com/how-we-saved-70k-cores-across-30-mission-critical-services/
package autogc

import (
	"os"
	"runtime"
	"runtime/debug"
	"strconv"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

var (
	gcChannel    = make(chan struct{}, 1)
	gcAlg        Algorithm
	previousGOGC = 100
)

const (
	defaultMemoryThreshold = 80
)

// InitAutoGOGC starts to adjust GOGC automatically
func InitAutoGOGC() {
	currentThreshold, err := strconv.Atoi(os.Getenv("AUTO_GC_MEMORY_THRESHOLD"))
	if err != nil {
		currentThreshold = defaultMemoryThreshold
		log.GetLogger().Warnf("failed to get AUTO_GC_MEMORY_THRESHOLD, use default threshold, %s", err.Error())
	} else if currentThreshold <= 0 || currentThreshold > percent {
		currentThreshold = defaultMemoryThreshold
	}
	log.GetLogger().Infof("current auto gc memory threshold: %d", currentThreshold)
	limit, err := parseCGroupMemoryLimit()
	if err != nil {
		log.GetLogger().Errorf("failed to read cgroup memory limit, err: %s", err.Error())
		return
	}
	log.GetLogger().Infof("cgroup memory limit is %d, memory %d", limit, uint64(currentThreshold)*limit/percent)

	gcAlg = &DefaultAlg{}
	if percent == 0 {
		return
	}
	gcAlg.Init(limit, uint64(currentThreshold)*limit/percent)

	newCycleRefObj()

	go runAutoGOGC()
}

func runAutoGOGC() {
	file, err := os.Open(memPath)
	if err != nil {
		log.GetLogger().Errorf("failed to open statm file")
		return
	}
	defer file.Close()
	buffer := make([]byte, KB)
	for range gcChannel {
		rss, err := parseRSS(file, buffer)
		if err != nil {
			log.GetLogger().Errorf("failed to parse RSS, err: %s", err.Error())
			return
		}
		previousGOGC = debug.SetGCPercent(gcAlg.NextGOGC(rss, previousGOGC))
	}
}

type finalizer struct {
	ref *finalizerRef
}

type finalizerRef struct {
	parent *finalizer
}

func finalizerHandler(f *finalizerRef) {
	select {
	case gcChannel <- struct{}{}:
	default:
	}
	runtime.SetFinalizer(f, finalizerHandler)
}

func newCycleRefObj() *finalizer {
	f := &finalizer{}
	f.ref = &finalizerRef{parent: f}
	runtime.SetFinalizer(f.ref, finalizerHandler)
	f.ref = nil
	return f
}
