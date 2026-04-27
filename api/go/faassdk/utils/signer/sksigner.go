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

// Package signer -
package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"strings"
)

const (
	// TimestampKey timestamp
	TimestampKey = "timestamp="

	// AccessIDKey accessId
	AccessIDKey = "accessId="

	// SignatureKey signature
	SignatureKey = "signature="

	// Algorithm Signature algorithm
	Algorithm = "SDK-HMAC-SHA256"
)

var hexCodes = []rune("0123456789abcdef")

// Sign hmac sha256 signature
func Sign(sk, content []byte) []byte {
	h := hmac.New(sha256.New, sk)
	_, err := h.Write(content)
	if err != nil {
		return nil
	}
	return h.Sum(nil)
}

// EncodeHex encode hex
func EncodeHex(data []byte) string {
	if data == nil || len(data) == 0 {
		return ""
	}
	l := len(data)
	out := make([]rune, l<<1)
	j := 0
	for i := 0; i < l; i++ {
		if j >= l<<1 {
			return ""
		}
		out[j] = hexCodes[(data[i]>>4)&0xF] // magic number
		j++
		if j >= l<<1 {
			return ""
		}
		out[j] = hexCodes[(data[i] & 0xF)] // magic number
		j++
	}
	return string(out)
}

// BuildAuthorization Splicing Signature
func BuildAuthorization(ak, ts, sign string) string {
	var builder strings.Builder
	builder.WriteString(Algorithm)
	builder.WriteString(" ")
	builder.WriteString(AccessIDKey + ak)
	builder.WriteString(",")
	builder.WriteString(TimestampKey + ts)
	builder.WriteString(",")
	builder.WriteString(SignatureKey + sign)
	return builder.String()
}
