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

// Package crossclusterinvoke -
package crossclusterinvoke

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"

	"yuanrong.org/kernel/runtime/faassdk/common/constants"
	"yuanrong.org/kernel/runtime/faassdk/config"
	"yuanrong.org/kernel/runtime/faassdk/sts"
	"yuanrong.org/kernel/runtime/faassdk/types"
	"yuanrong.org/kernel/runtime/faassdk/utils"
	"yuanrong.org/kernel/runtime/faassdk/utils/signer"
	"yuanrong.org/kernel/runtime/faassdk/utils/urnutils"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/faas/logger"
)

const (
	crossHeaderKeyTraceID      = "X-Caas-Trace-Id"
	crossHeaderKeyCrossCluster = "X-System-Cross-Cluster"
	crossHeaderKeyClusterID    = "X-System-Clusterid"
	crossHeaderKeyTimestamp    = "X-System-Timestamp"
	crossHeaderKeySignature    = "X-System-Signature"
)

var (
	stsInitOnce sync.Once
	stsInitErr  error
)

// InvokeConfig -
type InvokeConfig struct {
	Enable           bool   `json:"enable" valid:"optional"`
	CrossClusterAddr string `json:"crossClusterAddr" valid:"optional"`
	ErrorCodes       string `json:"errorCodes" valid:"optional"`
	ErrCodeMap       map[int]struct{}
	AcquireTimeout   int `json:"acquireTimeout" valid:"optional"`
}

// AuthConfig -
type AuthConfig struct {
	AccessKey string `json:"accessKey" valid:"optional"`
	SecretKey string `json:"secretKey" valid:"optional"`
}

// Invoker -
type Invoker struct {
	StsServerConfig types.StsServerConfig
	FuncInfo        urnutils.BaseURN
	initErr         error
	InvokeConfig
	AuthConfig
}

// NewInvoker -
func NewInvoker(urn string) *Invoker {
	funcInfo, err := urnutils.GetFunctionInfo(urn)
	if err != nil {
		logger.GetLogger().Warnf("new cross cluster invoker error, err: %s", err.Error())
	}
	return &Invoker{
		InvokeConfig: getCrossClusterInvokeConfig(),
		AuthConfig:   getCrossClusterAuthConfig(),
		FuncInfo:     funcInfo,
		initErr:      err,
	}
}

// NeedCrossClusterInvoke -
func (invoker *Invoker) NeedCrossClusterInvoke(err error) bool {
	if !invoker.InvokeConfig.Enable {
		return false
	}
	snErr, ok := err.(api.ErrorInfo)
	if !ok {
		return false
	}
	if _, ok := invoker.ErrCodeMap[snErr.Code]; !ok {
		return false
	}
	return true
}

func (invoker *Invoker) getCalleeUrn(name, version string) string {
	calleeInfo := invoker.FuncInfo
	calleeInfo.Name = buildCalleeFullFuncName(invoker.FuncInfo.Name, name)
	calleeInfo.Version = version
	return calleeInfo.String()
}

// DoInvoke -
func (invoker *Invoker) DoInvoke(request types.InvokeRequest, response *types.GetFutureResponse, timeout time.Duration,
	logger api.FormatLogger) bool {
	if invoker.initErr != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = "crossClusterInvoker not ready, err: " + invoker.initErr.Error()
		return false
	}
	defer func() {
		if response.StatusCode != constants.NoneError {
			response.ErrorMessage = "do cross cluster invoke failed, " + response.GetErrorMessage()
		}
	}()
	if timeout <= 0 {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = "no time left"
		return false
	}
	calleeUrn := request.FuncUrn
	if calleeUrn == "" {
		calleeUrn = invoker.getCalleeUrn(request.FuncName, request.FuncVersion)
	}
	logger = logger.With(zap.Any("calleeUrn", calleeUrn), zap.Any("timeout", timeout.Milliseconds()),
		zap.Any("host", invoker.CrossClusterAddr))
	invokeUrl := fmt.Sprintf("/serverless/v1/functions/%s/invocations", calleeUrn)

	httpReq := fasthttp.AcquireRequest()
	httpRsp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(httpReq)
	defer fasthttp.ReleaseResponse(httpRsp)

	err := invoker.setHeader(httpReq, request, logger)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("setHeader failed, err: %s", err)
		return false
	}
	httpReq.SetRequestURI(invokeUrl)
	httpReq.URI().SetScheme("http")
	httpReq.SetBody([]byte(request.Payload))
	httpClient := GetHttpClient()
	err = httpClient.DoTimeout(httpReq, httpRsp, timeout)

	if needTryLocalCluster(err, httpRsp, logger) {
		logger.Infof("cross cluster is not ready or upgrading")
		return true
	}

	if err != nil {
		logger.Errorf("do invoke failed, err: %s", err)
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("do invoke failed, err: %s", err)
		return false
	}
	if httpRsp.StatusCode()/100 != 2 { // 2xx is ok
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("do invoke failed, statuscode:%d is not accepted", httpRsp.StatusCode())
		return false
	}
	handleHttpResponse(httpRsp, response)
	return false
}

func needTryLocalCluster(err error, resp *fasthttp.Response, logger api.FormatLogger) bool {
	if err != nil {
		if strings.Contains(err.Error(), fasthttp.ErrDialTimeout.Error()) || utils.ContainsConnRefusedErr(err) {
			logger.Errorf("do cross invoke failed, cross cluster is in error: %s", err.Error())
			return true
		}
		return false
	}
	if resp == nil {
		return false
	}
	if resp.StatusCode() == fasthttp.StatusNotFound {
		return true
	}
	if resp.StatusCode()/100 != 2 { // 2xx is ok
		return false
	}
	callResponse := &struct {
		Code         int             `json:"code"`
		Message      string          `json:"message"`
		UserResponse json.RawMessage `json:"userResponse"`
	}{}
	err = json.Unmarshal(resp.Body(), callResponse)
	if err != nil {
		return false
	}
	if callResponse.Code == constants.ClusterIsUpgrading {
		logger.Errorf("do cross invoke failed, cross cluster is in upgrading")
		return true
	}
	return false
}

func handleHttpResponse(httpResp *fasthttp.Response, response *types.GetFutureResponse) {
	callResponse := &struct {
		Code         int             `json:"code"`
		Message      string          `json:"message"`
		UserResponse json.RawMessage `json:"userResponse"`
	}{}
	err := json.Unmarshal(httpResp.Body(), callResponse)
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("call response unmarshal error %s", err.Error())
		return
	}
	if callResponse.Code != constants.NoneError {
		response.StatusCode = callResponse.Code
		response.ErrorMessage = callResponse.Message
		return
	}
	contentBytes, err := callResponse.UserResponse.MarshalJSON()
	if err != nil {
		response.StatusCode = constants.FaaSError
		response.ErrorMessage = fmt.Sprintf("call response content unmarshal error %s", err.Error())
		return
	}
	response.StatusCode = constants.NoneError

	response.Content = string(contentBytes)
}

func (invoker *Invoker) setHeader(httpReq *fasthttp.Request, request types.InvokeRequest,
	logger api.FormatLogger) error {
	httpReq.Header.ResetConnectionClose()
	httpReq.Header.SetHost(invoker.CrossClusterAddr)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(crossHeaderKeyTraceID, request.TraceID)
	httpReq.Header.Set(crossHeaderKeyCrossCluster, "true")
	httpReq.Header.Set(crossHeaderKeyClusterID, os.Getenv("CLUSTER_ID"))
	httpReq.Header.SetMethod(fasthttp.MethodPost)
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10) // int64 -> string

	httpReq.Header.Set(crossHeaderKeyTimestamp, timeStamp)
	ak, sk, err := invoker.parseAuthConfig()
	if err != nil {
		logger.Errorf("parse secret error %s", err.Error())
		return err
	}
	signature := buildSignature(timeStamp, []byte(request.Payload), string(ak))
	sign := signer.Sign(sk, signature)
	signStr := signer.EncodeHex(sign)

	buildAuth := signer.BuildAuthorization(string(ak), timeStamp, signStr)
	httpReq.Header.Set(crossHeaderKeySignature, buildAuth)
	return nil
}

func (invoker *Invoker) parseAuthConfig() ([]byte, []byte, error) {
	var accessKey []byte
	var decodeKey []byte
	return accessKey, decodeKey, nil
}

func buildSignature(timeStamp string, body []byte, tenantId string) []byte {
	signValues := [][]byte{
		[]byte(tenantId),
		[]byte(timeStamp),
		body,
	}
	return bytes.Join(signValues, []byte("&"))
}

func buildCalleeFullFuncName(callerFullFuncName string, calleeFuncName string) string {
	defaultPrefix := "0@default@"
	splits := strings.Split(callerFullFuncName, "@") // "@" separator
	if len(splits) != 3 {                            // example: 0@default@hello
		return defaultPrefix + calleeFuncName
	}
	splits[2] = calleeFuncName       // callerFuncName -> calleeFuncName
	return strings.Join(splits, "@") // separator
}

func initSTS(serverCfg types.StsServerConfig) error {
	stsInitOnce.Do(func() {
		stsInitErr = sts.InitStsSDK(serverCfg)
		if stsInitErr != nil {
			logger.GetLogger().Errorf("failed to init sts sdk, err: %s", stsInitErr.Error())
		} else {
			logger.GetLogger().Infof("succeeded in initing sts sdk")
		}
	})
	return stsInitErr
}

func getCrossClusterInvokeConfig() InvokeConfig {
	var crossClusterInvokeConfig InvokeConfig
	content, err := os.ReadFile(config.DefaultRuntimeJsonFilePath)
	if err != nil {
		logger.GetLogger().Errorf("read crossClusterInvokeConfig failed, err: %s, "+
			"filePath: %s", err.Error(), config.DefaultRuntimeJsonFilePath)
		return crossClusterInvokeConfig
	}
	configStruct := struct {
		InvokeConfig `json:"crossClusterInvokeConfig"`
	}{}
	err = json.Unmarshal(content, &configStruct)
	if err != nil {
		logger.GetLogger().Errorf("unmarshal crossClusterInvokeConfig env failed, "+
			"content: %s, err: %v", string(content), err.Error())
		return crossClusterInvokeConfig
	}
	crossClusterInvokeConfig = configStruct.InvokeConfig
	splits := strings.Split(crossClusterInvokeConfig.ErrorCodes, ",")
	crossClusterInvokeConfig.ErrCodeMap = make(map[int]struct{})
	for _, v := range splits {
		errCode, err := strconv.Atoi(v)
		if err != nil {
			logger.GetLogger().Errorf("parse errCode failed, v: %s, err: %v", v, err)
			continue
		}
		crossClusterInvokeConfig.ErrCodeMap[errCode] = struct{}{}
	}
	logger.GetLogger().Infof("show crossclusterinvokeconfig: %v", crossClusterInvokeConfig)
	return crossClusterInvokeConfig
}

func getCrossClusterAuthConfig() AuthConfig {
	var crossClusterAuthConfig AuthConfig
	csac := os.Getenv("CROSS_CLUSTER_AUTH_CONFIG")
	err := json.Unmarshal([]byte(csac), &crossClusterAuthConfig)
	if err != nil {
		logger.GetLogger().Errorf("unmarshal crossClusterAuthConfig failed, config:%s  err: %s",
			csac, err.Error())
		return crossClusterAuthConfig
	}
	logger.GetLogger().Infof("get crossclusterAuthConfig success")
	return crossClusterAuthConfig
}
