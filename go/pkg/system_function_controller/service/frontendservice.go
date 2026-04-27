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

// Package service -
package service

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	"yuanrong.org/kernel/pkg/common/faas_common/k8sclient"
	"yuanrong.org/kernel/pkg/system_function_controller/config"
	"yuanrong.org/kernel/pkg/system_function_controller/constant"
)

var backoff = wait.Backoff{
	Steps:    5,
	Duration: 10000000, // 10 ms
	Factor:   1.5,
	Jitter:   0.3,
}

// CreateFrontendService -
func CreateFrontendService() error {
	nameSpace := config.GetFaaSControllerConfig().NameSpace
	if nameSpace == "" {
		nameSpace = constant.NamespaceDefault
	}
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constant.FuncNameFaasfrontend,
			Namespace: nameSpace,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				constant.SystemFuncName: constant.FuncNameFaasfrontend,
			},
			Ports: []v1.ServicePort{
				{
					Name:     "http",
					Protocol: "TCP",
					Port:     constant.ServiceFrontendPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: constant.ServiceFrontendTargetPort,
					},
					NodePort: constant.ServiceFrontendNodePort,
				},
			},
			Type: v1.ServiceTypeNodePort,
		},
	}

	createService := func() error {
		// create Service
		return k8sclient.GetkubeClient().CreateK8sService(service)
	}
	// Used to determine whether an error can be retried
	isRetriable := func(err error) bool {
		// always retry
		return true
	}
	return retry.OnError(backoff, isRetriable, createService)
}
