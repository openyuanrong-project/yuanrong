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

package healthlog

import "testing"

func TestPrintHealthLog(t *testing.T) {
	type args struct {
		stopCh   chan struct{}
		inputLog func()
		name     string
	}
	var a args
	a.stopCh = nil
	a.inputLog = func() {
		return
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "case1",
			args: args{
				stopCh: nil,
				inputLog: func() {
					return
				},
			},
		},
		{
			name: "case2",
			args: args{
				stopCh: make(chan struct{}),
				inputLog: func() {
					return
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.stopCh != nil {
				close(tt.args.stopCh)
			}
			PrintHealthLog(tt.args.stopCh, tt.args.inputLog, tt.args.name)
		})
	}
}
