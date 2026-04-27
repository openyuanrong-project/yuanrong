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

package main

import (
	"errors"

	"yuanrong.org/kernel/runtime/yr"
)

type ExceptionCounter struct {
	Count int
}

func NewCounterException(init int) *ExceptionCounter {
	return &ExceptionCounter{Count: init}
}

func (c *ExceptionCounter) ReturnException(x int) (int, error) {
	return x + 1, errors.New("error from go newCounter 1")
}

func (c *ExceptionCounter) CallAddException10(x int) (int, error) {
	instance := yr.Instance(NewCounterException).Invoke(x)
	defer instance.Terminate()
	objs := instance.Function((*ExceptionCounter).ReturnException).Invoke(200)
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *ExceptionCounter) CallAddException9(x int) (int, error) {
	instance := yr.Instance(NewCounterException).Invoke(x)
	defer instance.Terminate()
	objs := instance.Function((*ExceptionCounter).CallAddException10).Invoke(200)
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *ExceptionCounter) CallAddException8(x int) (int, error) {
	instance := yr.Instance(NewCounterException).Invoke(x)
	defer instance.Terminate()
	objs := instance.Function((*ExceptionCounter).CallAddException9).Invoke(200)
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *ExceptionCounter) CallAddException7(x int) (int, error) {
	instance := yr.Instance(NewCounterException).Invoke(x)
	defer instance.Terminate()
	objs := instance.Function((*ExceptionCounter).CallAddException8).Invoke(200)
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *ExceptionCounter) CallAddException6(x int) (int, error) {
	instance := yr.Instance(NewCounterException).Invoke(x)
	defer instance.Terminate()
	objs := instance.Function((*ExceptionCounter).CallAddException7).Invoke(200)
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *ExceptionCounter) CallAddException5(x int) (int, error) {
	instance := yr.Instance(NewCounterException).Invoke(x)
	defer instance.Terminate()
	objs := instance.Function((*ExceptionCounter).CallAddException6).Invoke(200)
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *ExceptionCounter) CallAddException4(x int) (int, error) {
	instance := yr.Instance(NewCounterException).Invoke(x)
	defer instance.Terminate()
	objs := instance.Function((*ExceptionCounter).CallAddException5).Invoke(200)
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *ExceptionCounter) CallAddException3(x int) (int, error) {
	instance := yr.Instance(NewCounterException).Invoke(x)
	defer instance.Terminate()
	objs := instance.Function((*ExceptionCounter).CallAddException4).Invoke(200)
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *ExceptionCounter) CallAddException2(x int) (int, error) {
	instance := yr.Instance(NewCounterException).Invoke(x)
	defer instance.Terminate()
	objs := instance.Function((*ExceptionCounter).CallAddException3).Invoke(200)
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *ExceptionCounter) CallAddException1(x int) (int, error) {
	instance := yr.Instance(NewCounterException).Invoke(x)
	defer instance.Terminate()
	objs := instance.Function((*ExceptionCounter).CallAddException2).Invoke(200)
	value, err := yr.Get[int](objs[0], 30000)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// compile command:
// CGO_LDFLAGS="-L$(path which contains libcpplibruntime.so)" \
// go build -buildmode=plugin -o yrlibexception.so actor_example_exception.go
