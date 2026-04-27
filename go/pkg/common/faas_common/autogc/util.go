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

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	cgroupMemLimitPath = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	rssValueFieldIndex = 1
	base               = 10
	bitSize            = 64
)

// constants of memory unit
const (
	B = 1 << (10 * iota)
	KB
	MB
	GB
)

var (
	pageSize = uint64(os.Getpagesize())
	memPath  = fmt.Sprintf("/proc/%d/statm", os.Getpid())
)

func parseCGroupMemoryLimit() (uint64, error) {
	v, err := ioutil.ReadFile(cgroupMemLimitPath)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(v)), base, bitSize)
}

func parseRSS(f io.ReadSeeker, buffer []byte) (uint64, error) {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}
	_, err = f.Read(buffer)
	if err != nil && err != io.EOF {
		return 0, err
	}
	fields := strings.Split(string(buffer), " ")
	if len(fields) < (rssValueFieldIndex + 1) {
		return 0, errors.New("invalid statm fields")
	}
	rss, err := strconv.ParseUint(fields[rssValueFieldIndex], base, bitSize)
	if err != nil {
		return 0, err
	}
	return rss * pageSize, nil
}
