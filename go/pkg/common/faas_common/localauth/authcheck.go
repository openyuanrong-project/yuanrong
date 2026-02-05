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

// Package localauth authenticates requests by local configmaps
package localauth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"yuanrong.org/kernel/pkg/common/faas_common/constant"
	"yuanrong.org/kernel/pkg/common/faas_common/logger/log"
	"yuanrong.org/kernel/pkg/common/faas_common/utils"
)

const (
	modeSDK = "SDKMode"
	modeHWS = "HWSMode"
	// the difference limit of a timestamp
	defaultTimestampDiffLimit = 5
	// 7 days
	maxTimestampDiffLimit = 10080
	maxHeaderLength       = 20
	minLengthOfAuthValue  = 2
	base                  = 10
	bitSize               = 64
)

const (
	// AuthPrefixHmacSha256 -
	AuthPrefixHmacSha256 = "HmacSha256"
	defaultSignFieldNum  = 4
	signExpirationTime   = int64(5 * 60 * 1000)
	signFieldSize        = 2
)

var timestampDiffLimit = getTimestampDiffLimit()

type modeOptions struct {
	authHeaderPrefix string
	timeFormat       string
	shortTimeFormat  string
	terminalString   string
	name             string
	date             string
}

var modeOption = &modeOptions{
	authHeaderPrefix: "",
	timeFormat:       "",
	shortTimeFormat:  "",
	terminalString:   "",
	name:             "",
	date:             "",
}

// Signer is a struct of
type Signer struct {
	signTime    time.Time
	serviceName string
	region      string
}

// AuthConfig represents configurations of local auth
type AuthConfig struct {
	AKey     string `json:"aKey" yaml:"aKey" valid:"optional"`
	SKey     string `json:"sKey" yaml:"sKey" valid:"optional"`
	Duration int    `json:"duration" yaml:"duration" valid:"optional"`
}

// Authentication represents aKey and sKey Decrypted from ak and sk
type Authentication struct {
	AKey []byte
	SKey []byte
}

// SignRequest -
type SignRequest struct {
	Method           string
	Path             string
	Query            string
	Body             []byte
	Headers          map[string]string
	SignedHeaderKeys []string
	Timestamp        string
	AK               string
	SK               string
}

// signLocalAuthRequest returns the authentication header
func signLocalAuthRequest(rawURL, timeStamp, appID string, key *Authentication, data []byte) (string, []byte) {
	signer := getSigner("SDKMode", "", "")
	timeStampInt, err := strconv.ParseInt(timeStamp, base, bitSize)
	if err != nil {
		log.GetLogger().Errorf("failed to parse the timestamp string")
		return "", data
	}
	signer.signTime = time.Unix(timeStampInt, 0)
	// default text of data
	if len(data) == 0 {
		data = []byte(`signature verification`)
	}
	header := make(map[string][]string, maxHeaderLength)
	header["Content-Type"] = []string{"application/json"}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		log.GetLogger().Errorf("failed to parse a URL")
		return "", data
	}
	request := &http.Request{Method: "POST", URL: parsedURL, Header: header}
	signerHeader := signer.sign(request, key.AKey, key.SKey, data, appID)
	return signerHeader["X-Identity-Sign"], data
}

func getSigner(mode, serviceName, region string) *Signer {
	if mode == modeSDK {
		setSDKMode()
	} else {
		setHWSMode()
	}
	return &Signer{
		signTime:    time.Now(),
		serviceName: serviceName,
		region:      region,
	}
}

func setSDKMode() {
	modeOption = &modeOptions{
		authHeaderPrefix: "SDK-HMAC-SHA256",
		timeFormat:       "20060102T150405Z",
		shortTimeFormat:  "20060102",
		terminalString:   "sdk_request",
		name:             "SDK",
		date:             "X-Sdk-Date",
	}
}

func setHWSMode() {
	modeOption = &modeOptions{
		authHeaderPrefix: "HWS-HMAC-SHA256",
		timeFormat:       "20060102T150405Z",
		shortTimeFormat:  "20060102",
		terminalString:   "hws_request",
		name:             "HWS",
		date:             "X-Hws-Date",
	}
}

func (sig *Signer) sign(request *http.Request, aKey, sKey []byte, body []byte,
	appID string) map[string]string {
	header := map[string]string{}
	request.Header.Add(modeOption.date, sig.signTime.UTC().Format(modeOption.timeFormat))
	contentSha256 := makeSha256Hex(body)
	canonicalString := sig.buildCanonicalRequest(request, contentSha256)
	stringToSign := sig.buildStringToSign(canonicalString)
	signatureStr := sig.buildSignature(sKey, stringToSign)
	credentialString := sig.buildCredentialString()
	signedHeaders := sig.buildSignedHeadersString(request)
	aKeyString := string(aKey)
	utils.ClearByteMemory(aKey)
	parts := []string{
		modeOption.authHeaderPrefix + " Credential=" + aKeyString + "/" + credentialString,
		"SignedHeaders=" + signedHeaders,
		"Signature=" + signatureStr,
	}
	if appID != "" {
		parts = append(parts, "appid="+appID)
	}
	utils.ClearStringMemory(aKeyString)

	signResult := strings.Join(parts, ", ")
	header["host"] = request.Host
	header[modeOption.date] = sig.signTime.UTC().Format(modeOption.timeFormat)
	header["Content-Type"] = "application/json;charset=UTF-8"
	header["Accept"] = "application/json"
	header["X-Identity-Sign"] = signResult
	return header
}

// buildSignature generate a signature with request and secret key
func (sig *Signer) buildSignature(sKey []byte, stringtoSign string) string {
	var secretBuf bytes.Buffer
	secretBuf.Write([]byte(modeOption.name))
	secretBuf.Write(sKey)
	utils.ClearByteMemory(sKey)
	sigTime := []byte(sig.signTime.UTC().Format(modeOption.shortTimeFormat))
	date := makeHmac(secretBuf.Bytes(), sigTime)
	secretBuf.Reset()
	region := makeHmac(date, []byte(sig.region))
	service := makeHmac(region, []byte(sig.serviceName))
	credentials := makeHmac(service, []byte(modeOption.terminalString))
	toSignature := makeHmac(credentials, []byte(stringtoSign))
	signature := hex.EncodeToString(toSignature)
	return signature
}

// buildStringToSign prepare data for building signature
func (sig *Signer) buildStringToSign(canonicalString string) string {
	stringToSign := strings.Join([]string{
		modeOption.authHeaderPrefix,
		sig.signTime.UTC().Format(modeOption.timeFormat),
		sig.buildCredentialString(),
		hex.EncodeToString(makeSha256([]byte(canonicalString))),
	}, "\n")
	return stringToSign
}

// buildCanonicalRequest converts the request info into canonical format
func (sig *Signer) buildCanonicalRequest(request *http.Request, hexbody string) string {
	canonicalHeadersOut := sig.buildCanonicalHeaders(request)
	signedHeaders := sig.buildSignedHeadersString(request)
	canonicalRequestStr := strings.Join([]string{
		request.Method,
		request.URL.Path + "/",
		request.URL.RawQuery,
		canonicalHeadersOut,
		signedHeaders,
		hexbody,
	}, "\n")
	return canonicalRequestStr
}

// buildCanonicalHeaders generate canonical headers
func (sig *Signer) buildCanonicalHeaders(request *http.Request) string {
	var headers []string

	for header := range request.Header {
		standardized := strings.ToLower(strings.TrimSpace(header))
		headers = append(headers, standardized)
	}
	sort.Strings(headers)

	for i, header := range headers {
		headers[i] = header + ":" + strings.Replace(request.Header.Get(header), "\n", " ", -1)
	}

	if len(headers) > 0 {
		return strings.Join(headers, "\n") + "\n"
	}

	return ""
}

// buildSignedHeadersString convert the header in request to a certain format
func (sig *Signer) buildSignedHeadersString(request *http.Request) string {
	var headers []string
	for header := range request.Header {
		headers = append(headers, strings.ToLower(header))
	}
	sort.Strings(headers)
	return strings.Join(headers, ";")
}

// buildCredentialString add date and several other information to signature header
func (sig *Signer) buildCredentialString() string {
	credentialString := strings.Join([]string{
		sig.signTime.UTC().Format(modeOption.shortTimeFormat),
		sig.region,
		sig.serviceName,
		modeOption.terminalString,
	}, "/")
	return credentialString
}

// makeHmac convert data into sha256 format with certain key
func makeHmac(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	_, err := hash.Write(data)
	if err != nil {
		log.GetLogger().Errorf("failed to write in makeHmac, error: %s", err.Error())
	}
	return hash.Sum(nil)

}

// makeHmac convert data into sha256 format
func makeSha256(data []byte) []byte {
	hash := sha256.New()
	_, err := hash.Write(data)
	if err != nil {
		log.GetLogger().Errorf("failed to write in makeSha256, error: %s", err.Error())
	}
	return hash.Sum(nil)
}

// makeHmac convert data into Hex format
func makeSha256Hex(data []byte) string {
	hash := sha256.New()
	_, err := hash.Write(data)
	if err != nil {
		log.GetLogger().Errorf("failed to write in makeSha256Hex, error: %s", err.Error())
	}
	md := hash.Sum(nil)
	hexBody := hex.EncodeToString(md)
	return hexBody
}

func getTimestampDiffLimit() float64 {
	var tsDiffLimit float64
	envTimestampDiffLimit, err := strconv.Atoi(os.Getenv("AUTH_VALID_TIME_MINUTE"))
	if err == nil && envTimestampDiffLimit > 0 && envTimestampDiffLimit <= maxTimestampDiffLimit {
		tsDiffLimit = float64(envTimestampDiffLimit)
	} else {
		tsDiffLimit = float64(defaultTimestampDiffLimit)
	}
	log.GetLogger().Infof("current timestampDiffLimit is %f", tsDiffLimit)
	return tsDiffLimit
}

// AuthCheckLocally authenticates requests by local auth
func AuthCheckLocally(ak string, sk string, requestSign string, timestamp string, duration int) error {
	if len(requestSign) == 0 {
		return fmt.Errorf("authentication string is nil")
	}
	curTime := time.Now().Unix()
	timeUnix, err := strconv.ParseInt(timestamp, base, bitSize)
	if err != nil {
		return fmt.Errorf("invalid timestamp")
	}
	// the default timestamp limit is 5 minutes
	if math.Abs(float64(curTime-timeUnix)) >= timestampDiffLimit*time.Minute.Seconds() {
		return fmt.Errorf("the request is timeout")
	}
	appID, err := getAppIDFromRequestSign(requestSign)
	if err != nil {
		return err
	}
	_, exist, err := GetLocalAuthCache(ak, sk, appID, duration).GetSignForReceiver(requestSign)
	if err != nil {
		log.GetLogger().Errorf("failed to get sign from receiver cache")
		return err
	}
	if exist {
		return nil
	}
	aKey, sKey, err := DecryptKeys(ak, sk)
	if err != nil {
		utils.ClearByteMemory(aKey)
		utils.ClearByteMemory(sKey)
		return err
	}
	key := &Authentication{
		AKey: aKey,
		SKey: sKey,
	}
	var data []byte
	signature, _ := signLocalAuthRequest("", timestamp, appID, key, data)
	utils.ClearByteMemory(aKey)
	utils.ClearByteMemory(sKey)
	if signature == "" || signature != requestSign {
		return fmt.Errorf("auth check failed")
	}
	if err := GetLocalAuthCache(ak, sk, appID, duration).updateReceiver(signature); err != nil {
		log.GetLogger().Errorf("failed to update receiver cache")
		return err
	}
	return nil
}

func getAppIDFromRequestSign(sign string) (string, error) {
	arrays := strings.Split(sign, "appid=")
	if len(arrays) < minLengthOfAuthValue {
		return "", fmt.Errorf("failed to parse authorization appid= %s", "*****")
	}
	arrays = strings.Split(arrays[1], ", ")
	return arrays[0], nil
}

// SignLocally makes signatures by local auth
func SignLocally(ak, sk, appID string, duration int) (string, string) {
	t, auth, err := GetLocalAuthCache(ak, sk, appID, duration).GetSignForSender()
	if err != nil {
		var data []byte
		log.GetLogger().Warnf("failed to get sender cache: %s", err.Error())
		return CreateAuthorization(ak, sk, "", appID, data)
	}
	return auth, t
}

// SignOMSVC make signatures for request send to OMSVC
func SignOMSVC(ak, sk, url string, data []byte) (string, string) {
	return CreateAuthorization(ak, sk, url, "", data)
}

// CreateAuthorization create Authentication Information
func CreateAuthorization(ak, sk, url, appID string, data []byte) (string, string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), base)
	aKey, sKey, err := DecryptKeys(ak, sk)
	if err != nil {
		utils.ClearByteMemory(aKey)
		utils.ClearByteMemory(sKey)
		log.GetLogger().Errorf("failed to decrypt SKey when create auth, error: %s", err.Error())
		return "", ""
	}
	key := &Authentication{
		AKey: aKey,
		SKey: sKey,
	}
	authorization, _ := signLocalAuthRequest(url, timestamp, appID, key, data)
	utils.ClearByteMemory(aKey)
	utils.ClearByteMemory(sKey)
	return authorization, timestamp
}

// SignWithHmacSha256 make signatures for request with system ak/sk
func SignWithHmacSha256(req *fasthttp.Request, ak, sk string) error {
	headers, signedHeaderKeys := getHeaders(req)
	signReq := &SignRequest{
		Method:           string(req.Header.Method()),
		Path:             string(req.URI().Path()),
		Query:            getQuery(req),
		Body:             req.Body(),
		Headers:          headers,
		SignedHeaderKeys: signedHeaderKeys,
		Timestamp:        strconv.FormatInt(time.Now().UnixMilli(), base),
		AK:               ak,
		SK:               sk,
	}
	signature := sign(signReq)
	if len(signature) == 0 {
		return fmt.Errorf("sign failed")
	}
	auth := fmt.Sprintf("%s timestamp=%s,ak=%s,signature=%s", AuthPrefixHmacSha256, signReq.Timestamp,
		signReq.AK, signature)
	req.Header.Set(constant.HeaderAuthorization, auth)
	return nil
}

// VerifySignWithHmacSha256 check signatures for request with system ak/sk
func VerifySignWithHmacSha256(ctx *fasthttp.RequestCtx, ak, sk string) error {
	authHeader := string(ctx.Request.Header.Peek(constant.HeaderAuthorization))
	signatures := strings.FieldsFunc(authHeader, func(r rune) bool {
		return r == ' ' || r == ','
	})
	if len(signatures) != defaultSignFieldNum {
		return fmt.Errorf("the signature field is missing")
	}
	timestamp, err := checkTimestamp(signatures[1])
	if err != nil {
		return err
	}
	err = checkSignField("ak", signatures[2], ak)
	if err != nil {
		return fmt.Errorf("%w, request value: %s, actual value: %s", err, signatures[2], ak)
	}
	headers, signedHeaderKeys := getHeaders(&ctx.Request)
	signReq := &SignRequest{
		Method:           string(ctx.Request.Header.Method()),
		Path:             string(ctx.Request.URI().Path()),
		Query:            getQuery(&ctx.Request),
		Body:             ctx.Request.Body(),
		Headers:          headers,
		SignedHeaderKeys: signedHeaderKeys,
		Timestamp:        timestamp,
		AK:               ak,
		SK:               sk,
	}
	actualSignature := sign(signReq)
	err = checkSignField("signature", signatures[3], actualSignature)
	if err != nil {
		return err
	}
	return nil
}

// checkTimestamp requestTimestamp eq: "timestamp=1767611707333"
func checkTimestamp(requestTimestamp string) (string, error) {
	timestampSlice := strings.FieldsFunc(requestTimestamp, func(r rune) bool {
		return r == '='
	})
	if len(timestampSlice) != signFieldSize {
		return "", fmt.Errorf("the timestamp is error, request value: %s", requestTimestamp)
	}
	// 将时间戳字符串转换为int64
	timestamp, err := strconv.ParseInt(timestampSlice[1], base, bitSize)
	if err != nil {
		return "", fmt.Errorf("parse timestamp: %s failed, err: %w", timestampSlice[1], err)
	}
	// 校验时间有效性
	now := time.Now().UnixMilli()
	diff := now - timestamp
	if diff < 0 {
		diff = -diff
	}
	if diff > signExpirationTime {
		return "", fmt.Errorf("timestamp is invaild, timestamp: %d, now: %d", timestamp, now)
	}
	return timestampSlice[1], nil
}

// checkSignField originalStr eq: ak=xxxx
func checkSignField(fieldName string, originalStr string, actualStr string) error {
	originals := strings.FieldsFunc(originalStr, func(r rune) bool {
		return r == '='
	})
	if len(originals) != signFieldSize {
		return fmt.Errorf("the %s is error", fieldName)
	}
	if originals[1] != actualStr {
		return fmt.Errorf("the %s is different, original value: %s, actual value: %s", fieldName, originals[1],
			actualStr)
	}
	return nil
}

func getHeaders(req *fasthttp.Request) (map[string]string, []string) {
	signedHeader := string(req.Header.Peek(constant.HeaderSignedHeader))
	if len(signedHeader) == 0 {
		return nil, nil
	}
	signedHeaderKeys := strings.FieldsFunc(signedHeader, func(r rune) bool {
		return r == ';'
	})
	size := len(signedHeaderKeys)
	headers := make(map[string]string, size)
	sortedHeaderKeys := make([]string, 0, size)
	for _, v := range signedHeaderKeys {
		headerValue := string(req.Header.Peek(v))
		if len(headerValue) == 0 {
			continue
		}
		headerKey := strings.ToLower(v)
		headers[headerKey] = headerValue
		sortedHeaderKeys = append(sortedHeaderKeys, headerKey)
	}
	sort.Strings(sortedHeaderKeys)
	return headers, sortedHeaderKeys
}

func getQuery(req *fasthttp.Request) string {
	originalQueryStr := string(req.URI().QueryString())
	if len(originalQueryStr) == 0 {
		return ""
	}
	querys := strings.FieldsFunc(originalQueryStr, func(r rune) bool {
		return r == '&'
	})
	sort.Strings(querys)
	var sortedQuery strings.Builder
	for i, value := range querys {
		if i > 0 {
			sortedQuery.WriteString("&")
		}
		sortedQuery.WriteString(value)
	}
	return sortedQuery.String()
}

func sign(req *SignRequest) string {
	canonicalString := buildCanonicalString(req)
	stringToSign := req.Timestamp + " " + makeSha256Hex([]byte(canonicalString))
	signatureStr := buildSignature(req.SK, stringToSign)
	return signatureStr
}

// buildCanonicalString converts the request info into canonical format
func buildCanonicalString(req *SignRequest) string {
	canonicalHeadersOut := buildCanonicalHeaders(req.Headers, req.SignedHeaderKeys)
	signedHeaders := buildSignedHeadersString(req.SignedHeaderKeys)
	hexBody := makeSha256Hex(req.Body)
	canonicalRequestStr := strings.Join([]string{
		req.Method,
		req.Path,
		req.Query,
		canonicalHeadersOut,
		signedHeaders,
		hexBody,
	}, "\n")
	return canonicalRequestStr
}

func buildCanonicalHeaders(headers map[string]string, keys []string) string {
	if headers == nil || keys == nil || len(keys) == 0 {
		return ""
	}
	var sb strings.Builder
	lowerHeaderConnection := strings.ToLower(constant.HeaderConnection)
	lowerHeaderAuthorization := strings.ToLower(constant.HeaderAuthorization)
	for _, v := range keys {
		headerKey := strings.ToLower(v)
		if headerKey == lowerHeaderConnection ||
			headerKey == lowerHeaderAuthorization {
			continue
		}
		headerValue := strings.TrimSpace(headers[headerKey])
		sb.WriteString(headerKey + ":" + headerValue + "\n")
	}
	return sb.String()
}

func buildSignedHeadersString(keys []string) string {
	if keys == nil || len(keys) == 0 {
		return ""
	}
	var sb strings.Builder
	first := true
	lowerHeaderConnection := strings.ToLower(constant.HeaderConnection)
	lowerHeaderAuthorization := strings.ToLower(constant.HeaderAuthorization)
	for _, v := range keys {
		headerKey := strings.ToLower(v)
		if headerKey == lowerHeaderConnection ||
			headerKey == lowerHeaderAuthorization {
			continue
		}
		if first {
			first = false
			sb.WriteString(headerKey)
		} else {
			sb.WriteString(";" + headerKey)
		}
	}
	return sb.String()
}

// buildSignature generate a signature by secret key
func buildSignature(sk string, data string) string {
	var secretBuf bytes.Buffer
	secretBuf.WriteString(sk)
	toSignature := makeHmac(secretBuf.Bytes(), []byte(data))
	signature := hex.EncodeToString(toSignature)
	return signature
}
