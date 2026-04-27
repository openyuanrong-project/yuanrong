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
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
)

const defaultExp = 300 * time.Second

// GenerateJwtSignedHeaders put header authorization to request header
func GenerateJwtSignedHeaders(req *fasthttp.Request, body []byte, wiseCloudCtx types.NuwaRuntimeInfo,
	serviceAccountJwt *types.ServiceAccountJwt) error {
	headers := map[string]string{}
	req.Header.Set("x-wisecloud-site", wiseCloudCtx.WisecloudSite)
	req.Header.Set("x-wisecloud-service-id", wiseCloudCtx.WisecloudServiceId)
	req.Header.Set("x-wisecloud-environment-id", wiseCloudCtx.WisecloudEnvironmentId)
	headers = map[string]string{
		"x-wisecloud-site":           wiseCloudCtx.WisecloudSite,
		"x-wisecloud-service-id":     wiseCloudCtx.WisecloudServiceId,
		"x-wisecloud-environment-id": wiseCloudCtx.WisecloudEnvironmentId,
	}

	jwtToken, err := generateJWTToken(req, string(body), headers, serviceAccountJwt)
	if err != nil {
		return err
	}
	// Set headers
	req.Header.Set(constant.HeaderAuthorization, jwtToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-client-id", strconv.FormatInt(serviceAccountJwt.ClientId, 10)) // decimal notation
	return nil
}

func generateJWTToken(req *fasthttp.Request, body string, headers map[string]string,
	serviceAccountJwt *types.ServiceAccountJwt) (string, error) {
	return generateJWTTokenGeneric(&types.ServiceAccount{
		PrivateKey: serviceAccountJwt.PrivateKey,
		ClientId:   serviceAccountJwt.ClientId,
		KeyId:      serviceAccountJwt.KeyId,
	}, headers, buildQueryPayload(string(req.URI().Path()), string(req.Header.Method()),
		string(req.URI().QueryString()), body), serviceAccountJwt.OauthTokenUrl, "JWT-PRO2")
}

func buildQueryPayload(queryPath, method, queryString, body string) string {
	var payloadBuilder strings.Builder
	payloadBuilder.WriteString(body)
	if queryPath != "" {
		payloadBuilder.WriteString("\n")
		payloadBuilder.WriteString(queryPath)
	}
	if method != "" {
		payloadBuilder.WriteString("\n")
		payloadBuilder.WriteString(method)
	}
	if queryString != "" {
		payloadBuilder.WriteString("\n")
		payloadBuilder.WriteString(queryString)
	}
	return payloadBuilder.String()
}

func generateJWTTokenGeneric(sa *types.ServiceAccount, headers map[string]string,
	body string, aud string, jwtTokenType string) (string, error) {
	requestSign, err := getRequestSignature(headers, body)
	if err != nil {
		return "", err
	}
	iat := time.Now()
	exp := iat.Add(defaultExp)
	token := &Token{
		Header: map[string]interface{}{
			"typ":           jwtTokenType,
			"sdkVersion":    20200,
			"clientVersion": 2,
			"alg":           "RS256",
			"kid":           sa.KeyId,
		},
		Claims: map[string]interface{}{
			"aud":              aud,
			"iss":              strconv.FormatInt(sa.ClientId, 10),
			"exp":              exp.Unix(),
			"iat":              iat.Unix(),
			"signedHeaders":    getSignedHeaders1(headers),
			"requestSignature": requestSign,
		},
	}
	rsaPrikey, err := getRSAPrivateKey(sa.PrivateKey)
	if err != nil {
		return "", err
	}
	signToken, err := token.Sign(rsaPrikey)
	if err != nil {
		return "", err
	}
	return "Bearer " + signToken, nil
}

func getRequestSignature(headers map[string]string, payload string) (string, error) {
	canonicalHeaders := ""
	if headers != nil && len(headers) != 0 {
		var keys []string
		for k := range headers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			canonicalHeaders += strings.ToLower(k)
			canonicalHeaders += ":"
			canonicalHeaders += strings.TrimSpace(headers[k])
			canonicalHeaders += "\n"
		}
	}

	if len(canonicalHeaders) == 0 {
		canonicalHeaders += "\n"
	}
	if len(payload) != 0 {
		ch, err := sha256String(payload)
		if err != nil {
			return "", err
		}
		canonicalHeaders += hex.EncodeToString(ch)
	}

	ch, err := sha256String(canonicalHeaders)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(ch), nil
}

func getSignedHeaders1(headMap map[string]string) string {
	if headMap != nil && len(headMap) != 0 {
		var keyArray []string
		for key := range headMap {
			keyArray = append(keyArray, key)
		}

		sort.Strings(keyArray)
		return strings.Join(keyArray, ";")
	}

	return ""
}

func sha256String(input string) ([]byte, error) {
	h := sha256.New()
	_, err := h.Write([]byte(input))
	if err != nil {
		return nil, err
	}

	output := h.Sum(nil)
	return output, nil
}

func getRSAPrivateKey(privateKey string) (interface{}, error) {
	priKeyByte, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}
	private := []byte(fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----",
		base64.StdEncoding.EncodeToString(priKeyByte)))
	pkPem, _ := pem.Decode(private)

	privateRsa, err := x509.ParsePKCS8PrivateKey(pkPem.Bytes)
	if err != nil {
		return nil, err
	}

	return privateRsa, nil
}
