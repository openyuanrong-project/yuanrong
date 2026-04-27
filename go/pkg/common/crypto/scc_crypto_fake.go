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

// Package crypto for auth
package crypto

// SCCInitialized -
func SCCInitialized() bool {
	return false
}

// GetSCCAlgorithm -
func GetSCCAlgorithm(algorithm string) int {
	return 0
}

// InitializeSCC -
func InitializeSCC(config SccConfig) bool {
	return false
}

// FinalizeSCC -
func FinalizeSCC() {
}

// SCCDecrypt -
func SCCDecrypt(cipher []byte) (string, error) {
	return "", nil
}

// SCCEncrypt -
func SCCEncrypt(plainInput string) ([]byte, error) {
	return []byte{}, nil
}
