/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

package scaler

import (
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"k8s.io/apimachinery/pkg/util/wait"

	"yuanrong.org/kernel/pkg/common/faas_common/resspeckey"
	"yuanrong.org/kernel/pkg/common/faas_common/types"
	wisecloudTypes "yuanrong.org/kernel/pkg/common/faas_common/wisecloudtool/types"
	"yuanrong.org/kernel/pkg/functionscaler/config"
	types2 "yuanrong.org/kernel/pkg/functionscaler/types"
)

func TestNewWiseCloudScaler(t *testing.T) {
	config.GlobalConfig.ServiceAccountJwt = wisecloudTypes.ServiceAccountJwt{
		NuwaRuntimeAddr:      "https://127.0.0.1/v1/test",
		NuwaGatewayAddr:      "https://127.0.0.1/v1/test",
		OauthTokenUrl:        "https://oauthtokenurl.com/oauthtoken",
		ServiceAccountKeyStr: "",
		ServiceAccount: &wisecloudTypes.ServiceAccount{
			PrivateKey: "xxxxxxxxxxxxxxxxxxx",
			ClientId:   123456,
			KeyId:      "0xxxxxxxxxxxxxxxxxxx4d",
			PublicKey:  "test",
			UserId:     123,
			Version:    1,
		},
		TlsConfig: &wisecloudTypes.TLSConfig{
			HttpsInsecureSkipVerify: false,
			TlsCipherSuitesStr:      nil,
			TlsCipherSuites:         nil,
		},
	}
	done := make(chan struct{})
	defer gomonkey.ApplyFunc(wait.ExponentialBackoff, func(backoff wait.Backoff, condition wait.ConditionFunc) error {
		done <- struct{}{}
		return fmt.Errorf("some error")
	}).Reset()
	scaler := NewWiseCloudScaler("testfunc-label1", resspeckey.ResSpecKey{
		CPU:                 500,
		Memory:              600,
		EphemeralStorage:    0,
		CustomResources:     "",
		CustomResourcesSpec: "",
		InvokeLabel:         "label1222",
	}, true, nil)
	scaler.SetEnable(true)
	scaler.TriggerScale()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		assert.Error(t, fmt.Errorf("cold start error"))
	}
	scaler.HandleInsThdUpdate(0, 100)
	scaler.HandleFuncSpecUpdate(&types2.FunctionSpecification{InstanceMetaData: types.InstanceMetaData{ConcurrentNum: 1}})
	scaler.HandleInsConfigUpdate(nil)
	scaler.SetEnable(false)
	scaler.UpdateCreateMetrics(time.Nanosecond)
	scaler.CheckScaling()
	assert.Equal(t, 100, scaler.GetExpectInstanceNumber())
	scaler.HandleCreateError(nil)
	scaler.Destroy()
}

func TestNewWiseCloudScalerSuccess(t *testing.T) {
	config.GlobalConfig.ServiceAccountJwt = wisecloudTypes.ServiceAccountJwt{
		NuwaRuntimeAddr:      "https://127.0.0.1/v1/test",
		NuwaGatewayAddr:      "https://127.0.0.1/v1/test",
		OauthTokenUrl:        "https://oauthtokenutl.com/oauthtoken",
		ServiceAccountKeyStr: "",
		ServiceAccount: &wisecloudTypes.ServiceAccount{
			PrivateKey: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			ClientId:   2010000826,
			KeyId:      "yyyyyyyyyyyyyyyyyyyyyyyyy",
			PublicKey:  "zzzz",
			UserId:     123,
			Version:    1,
		},
		TlsConfig: &wisecloudTypes.TLSConfig{
			HttpsInsecureSkipVerify: false,
			TlsCipherSuitesStr:      []string{"aaaaa"},
			TlsCipherSuites:         nil,
		},
	}
	done := make(chan struct{})
	defer gomonkey.ApplyFunc((*fasthttp.Client).Do, func(_ *fasthttp.Client, req *fasthttp.Request, resp *fasthttp.Response) error {
		done <- struct{}{}
		return nil
	}).Reset()
	scaler := NewWiseCloudScaler("testfunc-label1", resspeckey.ResSpecKey{
		CPU:                 500,
		Memory:              700,
		EphemeralStorage:    0,
		CustomResources:     "",
		CustomResourcesSpec: "",
		InvokeLabel:         "label1",
	}, true, nil)
	scaler.SetEnable(true)
	scaler.TriggerScale()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		assert.Error(t, fmt.Errorf("cold start error"))
	}
}

func TestNewWiseCloudScalerSetInsFetal(t *testing.T) {
	convey.Convey("test", t, func() {
		config.GlobalConfig.ServiceAccountJwt = wisecloudTypes.ServiceAccountJwt{
			NuwaRuntimeAddr:      "https://127.0.0.1/v1/test",
			NuwaGatewayAddr:      "https://127.0.0.1/v1/test",
			OauthTokenUrl:        "https://oauthtokenurl.com/oauthtoken",
			ServiceAccountKeyStr: "",
			ServiceAccount: &wisecloudTypes.ServiceAccount{
				PrivateKey: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				ClientId:   123456,
				KeyId:      "yyyyyyyyyyyyyyyyyyyyyyyyyyyyyy",
				PublicKey:  "zzzzz",
				UserId:     123,
				Version:    1,
			},
			TlsConfig: &wisecloudTypes.TLSConfig{
				HttpsInsecureSkipVerify: false,
				TlsCipherSuitesStr:      []string{"aaaaa"},
				TlsCipherSuites:         nil,
			},
		}
		defer gomonkey.ApplyFunc((*fasthttp.Client).Do, func(_ *fasthttp.Client, req *fasthttp.Request, resp *fasthttp.Response) error {
			return nil
		}).Reset()
		wisescaler := NewWiseCloudScaler("testfunc-label1", resspeckey.ResSpecKey{
			CPU:                 500,
			Memory:              700,
			EphemeralStorage:    0,
			CustomResources:     "",
			CustomResourcesSpec: "",
			InvokeLabel:         "label1",
		}, true, nil)
		wisescaler.SetEnable(true)
		ins := types2.Instance{InstanceStatus: types.InstanceStatus{Code: 11, ErrorCode: 3015}, PodID: "podnamespace:podname"}
		scale, ok := wisescaler.(*WiseCloudScaler)
		print("scale is nil")
		print(scale == nil)
		var err error
		if ok && scale != nil {
			scale.nuwaRuntimeInfo = &wisecloudTypes.NuwaRuntimeInfo{WisecloudRuntimeId: "WisecloudRuntimeId"}
			err = scale.DelNuwaPod(&ins)
		}
		convey.So(err, convey.ShouldBeNil)
	})
}
