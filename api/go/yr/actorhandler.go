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

/*
Package yr
This package provides methods to obtain the execution interface.
*/
package yr

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"

	"github.com/vmihailenco/msgpack"

	"yuanrong.org/kernel/runtime/libruntime"
	"yuanrong.org/kernel/runtime/libruntime/api"
	"yuanrong.org/kernel/runtime/libruntime/common/logger/log"
	"yuanrong.org/kernel/runtime/libruntime/config"
	"yuanrong.org/kernel/runtime/libruntime/execution"
)

type actorHandlers struct {
	execType execution.ExecutionType
	plugins  map[string]*plugin.Plugin
	instance *reflect.Value
}

var (
	actorHdlrs actorHandlers = actorHandlers{
		execType: execution.ExecutionTypeActor,
		plugins:  make(map[string]*plugin.Plugin),
		instance: nil,
	}
)

// parseArgsValue obtain the parameter type through reflection and deserialize the parameter.
func parseArgsValue(method reflect.Value, args []api.Arg) ([]reflect.Value, error) {
	if len(args) != method.Type().NumIn() {
		return make([]reflect.Value, 0), fmt.Errorf("args number is not valid")
	}

	values := make([]reflect.Value, 0, len(args))
	for i := 0; i < method.Type().NumIn(); i++ {
		value, err := parseArgValue(args[i], method.Type().In(i))
		if err != nil {
			return values, err
		}

		if method.Type().In(i).Kind() != reflect.Ptr {
			values = append(values, value.Elem())
			continue
		}
		values = append(values, value)
	}
	return values, nil
}

func parseArgValue(arg api.Arg, dataType reflect.Type) (reflect.Value, error) {
	if arg.Type == api.ObjectRef {
		// todo is ref
	}

	value := reflect.New(dataType).Interface()
	if unmarshalError := msgpack.Unmarshal(arg.Data, value); unmarshalError != nil {
		return reflect.Value{}, unmarshalError
	}

	return reflect.ValueOf(value), nil
}

// GetExecType get execution type
func (h *actorHandlers) GetExecType() execution.ExecutionType {
	return h.execType
}

// LoadFunction load user function
func (h *actorHandlers) LoadFunction(codePaths []string) error {
	for _, path := range codePaths {
		files, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		for _, file := range files {
			// not support load recursively
			if file.IsDir() {
				continue
			}
			if filepath.Ext(file.Name()) != ".so" {
				continue
			}
			pluginPath := filepath.Join(path, file.Name())
			p, err := plugin.Open(pluginPath)
			if err != nil {
				return err
			} else {
				h.plugins[pluginPath] = p
			}
		}
	}
	return nil
}

// invokeFunction invoke user function
func (h *actorHandlers) invokeFunction(funcMeta api.FunctionMeta, args []api.Arg) ([]reflect.Value, error) {
	for _, p := range h.plugins {
		symbol, err := p.Lookup(funcMeta.FuncName)
		if err != nil {
			continue
		}
		method := reflect.ValueOf(symbol)
		methodArgs, err := parseArgsValue(method, args)
		if err != nil {
			return make([]reflect.Value, 0), err
		}
		returnValues := method.Call(methodArgs)
		err = catchUserErr(returnValues)
		if err != nil {
			return make([]reflect.Value, 0), err
		}
		return returnValues, nil
	}

	return make([]reflect.Value, 0), fmt.Errorf("could not find function({})", funcMeta.FuncName)
}

func (h *actorHandlers) invokeMemberFunction(funcMeta api.FunctionMeta, args []api.Arg) ([]reflect.Value, error) {
	if h.instance == nil {
		return make([]reflect.Value, 0), fmt.Errorf("did not create instance firstly")
	}
	method := h.instance.MethodByName(funcMeta.FuncName)
	if !method.IsValid() {
		return make([]reflect.Value, 0), fmt.Errorf("can not find method: ", funcMeta.FuncName)
	}

	methodArgs, err := parseArgsValue(method, args)
	if err != nil {
		return make([]reflect.Value, 0), err
	}
	returnValues := method.Call(methodArgs)
	err = catchUserErr(returnValues)
	if err != nil {
		return make([]reflect.Value, 0), err
	}
	return returnValues, err
}

func catchUserErr(returnValues []reflect.Value) error {
	for i := 0; i < len(returnValues); i++ {
		if returnValues[i].Interface() == nil {
			continue
		}
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if returnValues[i].Type().Implements(errorInterface) {
			e, ok := returnValues[i].Interface().(error)
			if !ok {
				return nil
			}
			return e
		}
	}
	return nil
}

func (h *actorHandlers) createInstance(
	funcMeta api.FunctionMeta, args []api.Arg, returnobjs []config.DataObject,
) error {
	fmt.Println("stateful create")
	var err error
	var returnValues []reflect.Value
	returnValues, err = h.invokeFunction(funcMeta, args)
	if err != nil {
		fmt.Println("Create stateful instance failed, error: ", err)
		return api.AddStack(err, GetStackTrace(h.instance, funcMeta.FuncName))
	}

	if len(returnValues) != 1 {
		fmt.Println("return value number is valid")
		err = fmt.Errorf("return value number is valid")
		return err
	}

	h.instance = &returnValues[0]
	return nil
}

func processInvokeRes(returnValues []reflect.Value, returnobjs []config.DataObject) error {
	if len(returnValues) == 0 {
		fmt.Println("return number is zero")
		return nil
	}

	packedReturnValues := make([][]byte, 0, len(returnValues))
	for _, returnValue := range returnValues {
		packedReturnValue, err := msgpack.Marshal(returnValue.Interface())
		if err != nil {
			fmt.Println("marshel failed ", err)
			return nil
		}
		packedReturnValues = append(packedReturnValues, packedReturnValue)
	}

	ret := packedReturnValues[0]
	var totalNativeBufferSize uint = 0
	var do *config.DataObject = &returnobjs[0]
	if err := libruntime.AllocReturnObject(do, uint(len(ret)), []string{}, &totalNativeBufferSize); err != nil {
		return err
	}

	if err := libruntime.WriterLatch(do); err != nil {
		return err
	}
	defer func() {
		if err := libruntime.WriterUnlatch(do); err != nil {
			log.GetLogger().Errorf("%v", err)
		}
	}()

	if err := libruntime.MemoryCopy(do, ret); err != nil {
		return err
	}

	if err := libruntime.Seal(do); err != nil {
		return err
	}

	return nil
}

func (h *actorHandlers) invokeInstanceStateless(
	funcMeta api.FunctionMeta, args []api.Arg, returnobjs []config.DataObject,
) error {
	fmt.Println("stateless invoke")
	returnValues, err := h.invokeFunction(funcMeta, args)
	if err != nil {
		fmt.Println("Invoke Instance Stateless failed, error: ", err)
		return api.AddStack(err, GetStackTrace(h.instance, funcMeta.FuncName))
	}
	err = processInvokeRes(returnValues, returnobjs)
	if err != nil {
		fmt.Println("failed to process invokeInstanceStateless result by runtime: ", err)
	}
	return err
}

func (h *actorHandlers) invokeInstance(
	funcMeta api.FunctionMeta, args []api.Arg, returnobjs []config.DataObject,
) error {
	fmt.Println("stateful invoke")
	returnValues, err := h.invokeMemberFunction(funcMeta, args)
	if err != nil {
		fmt.Println("Invoke Instance Stateless failed, error: ", err)
		return api.AddStack(err, GetStackTrace(h.instance, funcMeta.FuncName))
	}
	err = processInvokeRes(returnValues, returnobjs)
	if err != nil {
		fmt.Println("failed to process invokeInstance result by runtime: ", err)
	}
	return err
}

// FunctionExecute function execute hook
func (h *actorHandlers) FunctionExecute(
	funcMeta api.FunctionMeta, invokeType config.InvokeType, args []api.Arg, returnobjs []config.DataObject,
) error {
	switch invokeType {
	case config.CreateInstance:
		return h.createInstance(funcMeta, args, returnobjs)
	case config.CreateInstanceStateless:
		fmt.Println("stateless create")
		return nil
	case config.InvokeInstance:
		return h.invokeInstance(funcMeta, args, returnobjs)
	case config.InvokeInstanceStateless:
		return h.invokeInstanceStateless(funcMeta, args, returnobjs)
	default:
		fmt.Println("invalid invoke type: ", invokeType)
		return fmt.Errorf("invalid invoke type: %s", invokeType)
	}
}

// Checkpoint check point
func (h *actorHandlers) Checkpoint(checkpointID string) ([]byte, error) {
	return []byte{}, nil
}

// Recover recover hook
func (h *actorHandlers) Recover(state []byte) error {
	return nil
}

// Shutdown hook
func (h *actorHandlers) Shutdown(gracePeriod uint64) error {
	return nil
}

// Signal hook
func (h *actorHandlers) Signal(sig int, data []byte) error {
	return nil
}

func (h *actorHandlers) HealthCheck() (api.HealthType, error) {
	return api.Healthy, nil
}

func newActorFuncExecutionIntfs() execution.FunctionExecutionIntfs {
	runtime := &ClusterModeRuntime{}
	GetRuntimeHolder().Init(runtime)
	return &actorHdlrs
}
