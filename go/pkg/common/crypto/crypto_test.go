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

// This test file can also be used as a tool to create, encrypt and decrypt our secrets and cipher texts
package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAll tests all processes, including creating random numbers, encryption and decryption
func TestAll(t *testing.T) {
	rootKey := RootKey{}
	randNum := hex.EncodeToString(createRandNum())
	fmt.Println(randNum)
	rootKey.RootKey = []byte(randNum)
	content := "abcd"
	secret := hex.EncodeToString(createRandNum())
	fmt.Println(secret)
	cipherText, err := Encrypt(content, []byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(cipherText)
	result, err := Decrypt(cipherText, []byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, content, result)
}

// TestRandNum is also a tool to create the random number for encryption
func TestRandNum(t *testing.T) {
	randNum := hex.EncodeToString(createRandNum())
	fmt.Println("randNum: " + randNum)
}

// TestEncrypt is also a tool to generate a cipher text from a plain text and a secret
func TestEncrypt(t *testing.T) {
	content := "7b83a1e330ccb177048671182f5ce1fde59c4c1c8167e8cf56190c4a5dd2c434"
	secret := "f7de29fa800605cd7f490ff1d1607fffc1387f05ad8ca059868ab605d6bb6b6b"
	cipherText, err := Encrypt(content, []byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(cipherText))
}

// TestDecrypt is also a tool to decrypt a cipher text with a secret and get the plain text
func TestDecrypt(t *testing.T) {
	cipherText := "b53df10229eead59476ae034:1feb0793e5b021511f064681827dbb8660594b31dfd90e665fa9664fdf02f1aa64304b1db66328e0b87f19c188d9e0d6487049b19a3b3aab25e3c3dcdcd22d390e020dce27af51b94ac154d137a9ce19"
	secret := "f7de29fa800605cd7f490ff1d1607fffc1387f05ad8ca059868ab605d6bb6b6b"
	content, err := Decrypt([]byte(cipherText), []byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, content, "7b83a1e330ccb177048671182f5ce1fde59c4c1c8167e8cf56190c4a5dd2c434")
	fmt.Println(content)
}

func TestDecryptError(t *testing.T) {
	_, err := Decrypt([]byte("1A"), []byte("1C"))
	assert.NotNil(t, err)

	_, err = Decrypt([]byte("1A:1B"), []byte("1C"))
	assert.NotNil(t, err)

	_, err = Decrypt([]byte("1Z:1B"), []byte("1C"))
	assert.NotNil(t, err)

	_, err = Decrypt([]byte("1A:1Z"), []byte("1C"))
	assert.NotNil(t, err)

	_, err = Decrypt([]byte("1A:1B"), []byte("1Z"))
	assert.NotNil(t, err)
}

func createRandNum() []byte {
	var keyLengthAES256 = 32
	initNum := make([]byte, keyLengthAES256)
	_, err := rand.Read(initNum)
	if err != nil {
		return nil
	}
	return initNum
}

func TestDecryptByte(t *testing.T) {
	_, err := DecryptByte([]byte("1A"), []byte("1C"))
	assert.NotNil(t, err)

	_, err = DecryptByte([]byte("1A:1B"), []byte("1C"))
	assert.NotNil(t, err)
}

func Test_encryptPBKDF2WithSHA256(t *testing.T) {
	data3 := "0B6AA66FADD74F59F019109582E1AAED1EEEEA14CEDFAFCA6DB384D8C3360D5E34087FD513B16929A2567E5E184" +
		"AE2B49A71B9E25E6371C91227D8CE114957D3D383EBC4899DBA7C43F6D80273E57F60B8FC918C2474CA687F1C5DBD7A71" +
		"B1DC0A1EA455C7F2304A4846FD05FFD9FDD96B606546C51241A190EF8B70382ABE55"

	f := &RootKeyFactor{
		iterCount:      IterKeyFactoryIter,
		component3:     data3,
		component3byte: []byte(data3),
	}
	rKey := encryptPBKDF2WithSHA256(f)
	assert.NotNil(t, rKey.RootKey)

	rootKey = rKey

	s := &SecretWorkKey{
		Key: "123",
		Mac: "abc",
	}

	data, err := s.MarshalJSON()
	assert.Nil(t, err)

	err = s.UnmarshalJSON(data)
	assert.Nil(t, err)
}

func Test_EncryptGcmDataFromBody(t *testing.T) {
	encryptGcmDataFromBody([]byte{}, []byte{})
}
