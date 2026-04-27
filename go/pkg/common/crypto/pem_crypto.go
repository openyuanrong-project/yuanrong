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

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"strings"

	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
)

// PEMCipher -
type PEMCipher int

// Possible values for the EncryptPEMBlock encryption algorithm.
const (
	_ PEMCipher = iota
	PEMCipherAES128
	PEMCipherAES192
	PEMCipherAES256
)

const (
	saltLength = 8
	aes128Cbc  = "AES-128-CBC"
	aes192Cbc  = "AES-192-CBC"
	aes256Cbc  = "AES-256-CBC"
)

// cipherUnit holds a method for enciphering a PEM block.
type cipherUnit struct {
	cipher     PEMCipher
	name       string
	cipherFunc func(key []byte) (cipher.Block, error)
	keySize    int
	blockSize  int
}

// cipherUnits holds a slice of cipherUnit.
var cipherUnits = []cipherUnit{{
	name:       aes256Cbc,
	cipher:     PEMCipherAES256,
	cipherFunc: aes.NewCipher,
	keySize:    32,
	blockSize:  aes.BlockSize,
}, {
	name:       aes192Cbc,
	cipher:     PEMCipherAES192,
	cipherFunc: aes.NewCipher,
	keySize:    24,
	blockSize:  aes.BlockSize,
}, {
	name:       aes128Cbc,
	cipher:     PEMCipherAES128,
	cipherFunc: aes.NewCipher,
	keySize:    16,
	blockSize:  aes.BlockSize,
},
}

// deriveKey uses a key derivation function to stretch the password into a key with
// the number of bits our cipher requires.
func (c cipherUnit) deriveKey(password, salt []byte) []byte {
	hash := md5.New()
	out := make([]byte, c.keySize)
	var digest []byte

	for i := 0; i < len(out); i += len(digest) {
		hash.Reset()
		_, err := hash.Write(digest)
		if err != nil {
			log.GetLogger().Warnf("write digest failed, err: %s", err)
		}
		_, err = hash.Write(password)
		if err != nil {
			log.GetLogger().Warnf("write password failed, err: %s", err)
		}
		_, err = hash.Write(salt)
		if err != nil {
			log.GetLogger().Warnf("write salt failed, err: %s", err)
		}
		digest = hash.Sum(digest[:0])
		copy(out[i:], digest)
	}
	return out
}

func cipherByName(name string) *cipherUnit {
	for i := range cipherUnits {
		alg := &cipherUnits[i]
		if alg.name == name {
			return alg
		}
	}
	return nil
}

// IsEncryptedPEMBlock returns whether the PEM block is password encrypted according to RFC 1423.
func IsEncryptedPEMBlock(b *pem.Block) bool {
	_, ok := b.Headers["DEK-Info"]
	return ok
}

// DecryptPEMBlock takes a PEM block encrypted according to RFC 1423 and the password used to encrypt
// it and returns a slice of decrypted DER encoded bytes.
func DecryptPEMBlock(b *pem.Block, pwd []byte) ([]byte, error) {
	dekInfo, ok := b.Headers["DEK-Info"]
	if !ok {
		return nil, errors.New("crypto: no DEK-Info header in block")
	}

	mode, hexIV, ok := strings.Cut(dekInfo, ",")
	if !ok {
		return nil, errors.New("crypto: malformed DEK-Info header")
	}

	ciph := cipherByName(mode)
	if ciph == nil {
		return nil, errors.New("crypto: unknown encryption mode")
	}
	iv, err := hex.DecodeString(hexIV)
	if err != nil {
		return nil, err
	}
	if len(iv) != ciph.blockSize {
		return nil, errors.New("crypto: incorrect IV size")
	}

	key := ciph.deriveKey(pwd, iv[:saltLength])
	block, err := ciph.cipherFunc(key)
	if err != nil {
		return nil, err
	}

	if len(b.Bytes)%block.BlockSize() != 0 {
		return nil, errors.New("crypto: encrypted PEM data is not a multiple of the block size")
	}

	data := make([]byte, len(b.Bytes))
	dec := cipher.NewCBCDecrypter(block, iv)
	dec.CryptBlocks(data, b.Bytes)

	dataLen := len(data)
	if dataLen == 0 || dataLen%ciph.blockSize != 0 {
		return nil, errors.New("crypto: invalid padding")
	}
	last := int(data[dataLen-1])
	if dataLen < last {
		return nil, errors.New("crypto: decryption password incorrect")
	}
	if last == 0 || last > ciph.blockSize {
		return nil, errors.New("crypto: decryption password incorrect")
	}
	for _, val := range data[dataLen-last:] {
		if int(val) != last {
			return nil, errors.New("crypto: decryption password incorrect")
		}
	}
	return data[:dataLen-last], nil
}
