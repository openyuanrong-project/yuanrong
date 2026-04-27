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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReadSeekerParser(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		parser   ParserFunc
		hasError bool
		expected uint64
	}{
		{
			name: "parse cgroup memory",
			content: []byte(`cache 10150707200
rss 880640
rss_huge 0
shmem 0
mapped_file 946176
dirty 135168
writeback 270336
swap 0
pgpgin 3158595
pgpgout 680215
pgfault 992277
pgmajfault 0
inactive_anon 0
active_anon 0
inactive_file 8343023616
active_file 1808744448
unevictable 0
hierarchical_memory_limit 9223372036854771712
hierarchical_memsw_limit 9223372036854771712
total_cache 21492334592
total_rss 9384980480
total_rss_huge 5515509760
total_shmem 654385152
total_mapped_file 2744586240
total_dirty 8110080
total_writeback 2027520
total_swap 0
total_pgpgin 1336448421
total_pgpgout 1354048239
total_pgfault 1405894809
total_pgmajfault 50622
total_inactive_anon 199806976
total_active_anon 8360579072
total_inactive_file 19150966784
total_active_file 3246854144
total_unevictable 0`),
			parser:   cgroupMemoryParserFunc,
			expected: 880640,
		},
		{
			name:     "parse cgroup memory no such line",
			content:  []byte(`880640`),
			parser:   cgroupMemoryParserFunc,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewReadSeekerParser(bytes.NewReader(tt.content), tt.parser)
			data, err := parser.Read()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.expected, data)
			}
			assert.Nil(t, parser.Close())
		})
	}
}
