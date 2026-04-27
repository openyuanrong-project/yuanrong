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

// Package uuid for common functions
package uuid

import (
	"testing"
)

func TestNew(t *testing.T) {
	m := make(map[RandomUUID]bool)
	for x := 1; x < 32; x++ {
		s := New()
		if m[s] {
			t.Errorf("New returned duplicated RandomUUID %s", s)
		}
		m[s] = true
	}
}

func TestSHA1(t *testing.T) {
	uuid := NewSHA1(NameSpaceURL, []byte("python.org")).String()
	want := "7af94e2b-4dd9-50f0-9c9a-8a48519bdef0"
	if uuid != want {
		t.Errorf("SHA1: got %q expected %q", uuid, want)
	}
}

func Test_parseRandomUUID(t *testing.T) {
	type args struct {
		uuidStr string
	}
	validRandomUUID := "6ba7b811-9dad-11d1-80b4-00c04fd430c8"
	invalidFormatRandomUUID := "6ba7b8119dad11d180b400c04fd430c8"
	illegalCharRandomUUID := "6ba7b811-9dad-11d1-80b4-00c04fd430cG"
	shortRandomUUID := "6ba7b811-9dad-11d1-80b4"
	longRandomUUID := "6ba7b811-9dad-11d1-80b4-00c04fd430c8-extra"
	emptyRandomUUID := ""
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test_parseRandomUUID_with_validRandomUUID",
			args: args{
				uuidStr: validRandomUUID,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Test_parseRandomUUID_with_invalidFormatRandomUUID",
			args: args{
				uuidStr: invalidFormatRandomUUID,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Test_parseRandomUUID_with_illegalCharRandomUUID",
			args: args{
				uuidStr: illegalCharRandomUUID,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Test_parseRandomUUID_with_shortRandomUUID",
			args: args{
				uuidStr: shortRandomUUID,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Test_parseRandomUUID_with_longRandomUUID",
			args: args{
				uuidStr: longRandomUUID,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Test_parseRandomUUID_with_emptyRandomUUID",
			args: args{
				uuidStr: emptyRandomUUID,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUUID(tt.args.uuidStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRandomUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got.String() != tt.args.uuidStr) == tt.want {
				t.Errorf("parseRandomUUID() got = %v, want %v", got.String(), tt.args.uuidStr)
			}
		})
	}
}

func BenchmarkParseRandomUUID(b *testing.B) {
	uuidStr := "6ba7b811-9dad-11d1-80b4-00c04fd430c8"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseUUID(uuidStr)
		if err != nil {
			b.Fatal(err)
		}
	}
}
