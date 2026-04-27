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

// Package etcd implements engine.Engine
package etcd

import (
	"io"
	"yuanrong.org/kernel/pkg/common/engine"
)

type etcdStmt struct {
	fn      func() ([]interface{}, error)
	filters []engine.FilterFunc
}

// Filter implements engine.PrepareStmt
func (s *etcdStmt) Filter(filter engine.FilterFunc) engine.PrepareStmt {
	s.filters = append(s.filters, filter)
	return s
}

// Execute implements engine.PrepareStmt
func (s *etcdStmt) Execute() (engine.Stream, error) {
	vs, err := s.fn()
	if err != nil {
		return nil, err
	}

	var res []interface{}
outer:
	for _, v := range vs {
		for _, filter := range s.filters {
			if !filter(v) {
				continue outer
			}
		}
		res = append(res, v)
	}
	return &etcdStream{vs: res}, nil
}

type etcdStream struct {
	vs  []interface{}
	pos int
}

// Next implements engine.Stream
func (s *etcdStream) Next() (interface{}, error) {
	defer func() {
		s.pos++
	}()

	if s.pos == len(s.vs) {
		return nil, io.EOF
	}
	return s.vs[s.pos], nil
}
