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

package functionlog

import "yuanrong.org/kernel/runtime/faassdk/common/constants"

// LogRecorderOption -
type LogRecorderOption func(r *LogRecorder)

// WithLogOption -
func WithLogOption(option string) LogRecorderOption {
	return func(r *LogRecorder) {
		r.logOption = option
	}
}

// WithLogOptionTail -
func WithLogOptionTail() LogRecorderOption {
	return WithLogOption(constants.RuntimeLogOptTail)
}

// WithoutLineBreak -
func WithoutLineBreak() LogRecorderOption {
	return func(r *LogRecorder) {
		r.separatedWithLineBreak = false
	}
}

// WithLivedataID -
func WithLivedataID(livedataID string) LogRecorderOption {
	return func(r *LogRecorder) {
		r.livedataID = livedataID
	}
}

// WithLogHeaderFooterGenerator -
func WithLogHeaderFooterGenerator(g LogHeaderFooterGenerator) LogRecorderOption {
	return func(r *LogRecorder) {
		r.generator = g
	}
}
