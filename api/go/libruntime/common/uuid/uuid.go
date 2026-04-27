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
	"crypto/rand"
	"encoding/hex"
	"io"
)

const (
	defaultByteNum   = 16
	indexFour        = 4
	indexSix         = 6
	indexEight       = 8
	indexNine        = 9
	indexTen         = 10
	indexThirteen    = 13
	indexFourteen    = 14
	indexEighteen    = 18
	indexNineteen    = 19
	indexTwentyThree = 23
	indexTwentyFour  = 24
	indexThirtySix   = 36
)

// RandomUUID -
type RandomUUID [defaultByteNum]byte

var (
	rander = rand.Reader // random function
)

// New -
func New() RandomUUID {
	return mustUUID(newRandom())
}

func mustUUID(uuid RandomUUID, err error) RandomUUID {
	if err != nil {
		return RandomUUID{}
	}
	return uuid
}

func newRandom() (RandomUUID, error) {
	return newRandomFromReader(rander)
}

func newRandomFromReader(r io.Reader) (RandomUUID, error) {
	var randomUUID RandomUUID
	_, err := io.ReadFull(r, randomUUID[:])
	if err != nil {
		return RandomUUID{}, err
	}
	randomUUID[indexSix] = (randomUUID[indexSix] & 0x0f) | 0x40     // Version 4
	randomUUID[indexEight] = (randomUUID[indexEight] & 0x3f) | 0x80 // Variant is 10
	return randomUUID, nil
}

// String-
func (uuid RandomUUID) String() string {
	var buf [indexThirtySix]byte
	encodeHex(buf[:], uuid)
	return string(buf[:])
}

func encodeHex(dstBuf []byte, uuid RandomUUID) {
	hex.Encode(dstBuf, uuid[:indexFour])
	dstBuf[indexEight] = '-'
	hex.Encode(dstBuf[indexNine:indexThirteen], uuid[indexFour:indexSix])
	dstBuf[indexThirteen] = '-'
	hex.Encode(dstBuf[indexFourteen:indexEighteen], uuid[indexSix:indexEight])
	dstBuf[indexEighteen] = '-'
	hex.Encode(dstBuf[indexNineteen:indexTwentyThree], uuid[indexEight:indexTen])
	dstBuf[indexTwentyThree] = '-'
	hex.Encode(dstBuf[indexTwentyFour:], uuid[indexTen:])
}
