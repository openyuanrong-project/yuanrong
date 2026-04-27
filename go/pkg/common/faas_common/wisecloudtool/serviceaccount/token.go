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

// Package serviceaccount sign http request by jwttoken
package serviceaccount

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

var methodHash = crypto.SHA256

// Token -
type Token struct {
	Header map[string]interface{}
	Claims map[string]interface{}
}

// Sign sign jwt token string
func (t *Token) Sign(key interface{}) (string, error) {
	jsonHeader, err := json.Marshal(t.Header)
	if err != nil {
		return "", err
	}
	header := base64.RawURLEncoding.EncodeToString(jsonHeader)

	jsonClaims, err := json.Marshal(t.Claims)
	if err != nil {
		return "", err
	}
	claim := base64.RawURLEncoding.EncodeToString(jsonClaims)

	stringToBeSign := strings.Join([]string{header, claim}, ".")

	sig, err := t.getSig(stringToBeSign, key)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{stringToBeSign, sig}, "."), nil
}

func (t *Token) getSig(signingString string, key interface{}) (string, error) {
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("key is invalid")
	}
	if !methodHash.Available() {
		return "", errors.New("the requested hash function is unavailable")
	}
	hasher := methodHash.New()
	_, err := hasher.Write([]byte(signingString))
	if err != nil {
		return "", errors.New("hash write failed")
	}

	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, methodHash, hasher.Sum(nil))
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(sigBytes), nil
}
