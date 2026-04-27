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

package common

import (
	"net"
	"sync"
	"testing"

	"google.golang.org/grpc"
)

func TestGetConnection(t *testing.T) {
	once = sync.Once{}
	connection = nil
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	go func() {
		grpc.NewServer().Serve(lis)
	}()

	CollectorConfigs.ManagerAddress = lis.Addr().String()

	conn := GetConnection()

	if conn == nil {
		t.Errorf("Expected a valid connection, but got nil")
	}

	conn.Close()
}
