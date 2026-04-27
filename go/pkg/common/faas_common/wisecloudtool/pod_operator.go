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

// Package wisecloudtool -
package wisecloudtool

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/serviceaccount"
	"yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
)

const (
	queryRetryTime     = 10
	queryRetryDuration = 200 * time.Millisecond // 初始等待时间
	queryRetryFactor   = 4                      // 倍数因子（每次翻4倍）
	queryRetryJitter   = 0.5                    // 随机抖动系数
	queryRetryCap      = 20 * time.Second       // 最大等待时间上限
)

var (
	coldStartBackoff = wait.Backoff{
		Duration: queryRetryDuration,
		Factor:   queryRetryFactor,
		Jitter:   queryRetryJitter,
		Steps:    queryRetryTime,
		Cap:      queryRetryCap,
	}
)

// PodOperator -
type PodOperator struct {
	nuwaConsoleAddr string //
	nuwaGatewayAddr string
	*types.ServiceAccountJwt
	*fasthttp.Client
	logger api.FormatLogger
}

// NewColdStarter -
func NewColdStarter(serviceAccountJwt *types.ServiceAccountJwt, logger api.FormatLogger) *PodOperator {
	return &PodOperator{
		nuwaConsoleAddr:   serviceAccountJwt.NuwaRuntimeAddr,
		nuwaGatewayAddr:   serviceAccountJwt.NuwaGatewayAddr,
		ServiceAccountJwt: serviceAccountJwt,
		Client: &fasthttp.Client{
			TLSConfig: &tls.Config{
				InsecureSkipVerify: serviceAccountJwt.TlsConfig.HttpsInsecureSkipVerify,
				CipherSuites:       serviceAccountJwt.TlsConfig.TlsCipherSuites,
				MinVersion:         tls.VersionTLS12,
			},
			MaxIdemponentCallAttempts: 3,
		},
		logger: logger,
	}
}

// ColdStart -
func (p *PodOperator) ColdStart(funcKeyWithRes string, resSpec resspeckey.ResSpecKey,
	info *types.NuwaRuntimeInfo) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(fmt.Sprintf("%s/activator/coldstart", p.nuwaConsoleAddr))
	req.Header.SetMethod(fasthttp.MethodPost)
	createInstanceReq := types.NuwaColdCreateInstanceReq{
		RuntimeId:   info.WisecloudRuntimeId,
		RuntimeType: "Function",
		PoolType:    "noPool",
		Memory:      resSpec.Memory,
		CPU:         resSpec.CPU,
		EnvLabel:    info.EnvLabel,
	}
	logger := p.logger.With(zap.Any("funcKeyWithRes", funcKeyWithRes), zap.Any("resKey", resSpec.String()))

	body, err := jsoniter.Marshal(createInstanceReq)
	if err != nil {
		return err
	}
	err = serviceaccount.GenerateJwtSignedHeaders(req, body, *info, p.ServiceAccountJwt)
	if err != nil {
		return err
	}
	req.SetBodyRaw(body)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	backoffErr := wait.ExponentialBackoff(
		coldStartBackoff, func() (bool, error) {
			err = p.Client.Do(req, resp)
			if err != nil {
				return false, nil
			}
			return true, nil
		})
	if backoffErr != nil {
		logger.Warnf("cold start error, backoffErr: %s", backoffErr.Error())
		return backoffErr
	}
	if err != nil {
		logger.Warnf("cold start error, backoffErr: %s", err.Error())
		return err
	}
	if resp.StatusCode()/100 != 2 { // resp http code != 2xx
		logger.Warnf("cold start error, code: %d, body: %s", resp.StatusCode(), string(resp.Body()))
		return fmt.Errorf("failed to cold start")
	}
	logger.Infof("cold start %s succeed", info.WisecloudRuntimeId)
	return nil
}

// DelPod will send a req to erase runtime pod
func (p *PodOperator) DelPod(nuwaRuntimeInfo *types.NuwaRuntimeInfo, deploymentName string,
	podId string) error {
	p.logger.Infof("delete nuwa pod %s", podId)
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(fmt.Sprintf("%s/runtime/instance", p.NuwaGatewayAddr))
	req.Header.SetMethod(fasthttp.MethodDelete)
	destroyInsReq := types.NuwaDestroyInstanceReq{
		RuntimeType:  "Function",
		RuntimeId:    nuwaRuntimeInfo.WisecloudRuntimeId,
		InstanceId:   podId,
		WorkLoadName: deploymentName,
	}

	reqBody, err := jsoniter.Marshal(destroyInsReq)
	if err != nil {
		return err
	}
	err = serviceaccount.GenerateJwtSignedHeaders(req, reqBody, *nuwaRuntimeInfo, p.ServiceAccountJwt)
	if err != nil {
		return err
	}
	req.SetBodyRaw(reqBody)

	logger := p.logger.With(zap.Any("deployment", deploymentName), zap.Any("podId", podId))
	rsp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(rsp)
	backoffError := wait.ExponentialBackoff(
		coldStartBackoff, func() (bool, error) {
			err = p.Client.Do(req, rsp)
			if err != nil {
				return false, nil
			}
			return true, nil
		})
	if backoffError != nil {
		logger.Warnf("delete runtime pod error, backoffErr: %s", backoffError.Error())
		return backoffError
	}
	if err != nil {
		logger.Warnf("delete runtime pod error, err: %s", err.Error())
		return err
	}
	if rsp.StatusCode()/100 != 2 { // resp http code != 2xx
		logger.Warnf("delete runtime pod error, code: %d, body: %s", rsp.StatusCode(), string(rsp.Body()))
		return fmt.Errorf("failed to delete runtime pod")
	}
	logger.Infof("succeed to delete runtime pod %s", podId)
	return nil
}
