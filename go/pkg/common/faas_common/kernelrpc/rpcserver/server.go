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

// Package rpcserver -
package rpcserver

import (
	"time"

	"yuanrong.org/kernel/runtime/libruntime/api"

	"yuanrong.org/kernel/pkg/common/faas_common/grpc/pb/common"
)

const (
	defaultSendTimeout = 5 * time.Second
)

// KernelInvokeHandler -
type KernelInvokeHandler func(args []*api.Arg, traceID string) (string, error)

// KernelServer defines basic POSIX server methods, currently only RegisterInvokeHandler is needed
type KernelServer interface {
	RegisterInvokeHandler(handler KernelInvokeHandler)
	Serve() error
	Stop()
}

func pb2Arg(args []*common.Arg) []*api.Arg {
	length := len(args)
	newArgs := make([]*api.Arg, 0, length)
	for _, value := range args {
		newArgs = append(newArgs, &api.Arg{Type: api.ArgType(value.Type), Data: value.Value})
	}
	return newArgs
}
