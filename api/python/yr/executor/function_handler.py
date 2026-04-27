#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""function handler"""

import logging
import os
import traceback
import inspect
from typing import List, Tuple
import uuid

from yr.libruntime_pb2 import InvokeType
from yr.npu_object import NpuObject
import yr
from yr.code_manager import CodeManager
from yr.common.utils import err_to_str
from yr.err_type import ErrorCode, ErrorInfo, ModuleCode
from yr.exception import YRInvokeError
from yr.executor.handler_intf import HandlerIntf
from yr.executor.instance_manager import InstanceManager
from yr.signature import recover_args
from yr.ds_tensor_client_manager import get_tensor_client

try:
    import torch
    import torch_npu
    from torch.npu import current_device

    TORCH_IMPORTED = True
except ImportError as import_err:
    TORCH_IMPORTED = False
    _import_error = import_err

USER_SHUTDOWN_FUNC_NAME = "__yr_shutdown__"
USER_BEFORE_SNAPSHOT_FUNC_NAME = "__yr_before_snapshot__"
USER_AFTER_SNAPSTART_FUNC_NAME = "__yr_after_snapstart__"

_logger = logging.getLogger(__name__)


def _store_tensor_to_ds(results, instance_function_name):
    def dev_mset_to_ds(tensor: torch.Tensor, instance_function_name):
        key = str(uuid.uuid4())
        ds_client = get_tensor_client()
        _logger.info(
            f"start dev mset, tensor key is {key}, function name: {instance_function_name}")
        failed_keys = ds_client.dev_mset([key], [tensor])
        if failed_keys:
            _logger.error(
                f"dev mset failed, failed keys is {failed_keys}, function name is {instance_function_name}")
            raise RuntimeError(
                f"dev_mset failed, failed keys: {failed_keys}, function name: {instance_function_name}")
        InstanceManager().store_tensor_in_local(tensor, key)
        return NpuObject(key, tensor.size(), tensor.dtype, tensor.device, os.getenv("YR_DS_ADDRESS"),
                         os.getenv("INSTANCE_ID"))

    if isinstance(results, torch.Tensor):
        return dev_mset_to_ds(results, instance_function_name)
    if isinstance(results, list):
        return [dev_mset_to_ds(result, instance_function_name) for result in results]
    if isinstance(results, tuple):
        return tuple(dev_mset_to_ds(result, instance_function_name) for result in results)
    if isinstance(results, set):
        return {dev_mset_to_ds(result, instance_function_name) for result in results}
    if isinstance(results, dict):
        return {k: dev_mset_to_ds(v, instance_function_name) for k, v in results.items()}
    return results


def _get_tensor_from_ds(*args, **kwargs):
    new_args = []
    new_kwargs = {}

    def dev_mget_from_ds(arg):
        if isinstance(arg, NpuObject):
            if not TORCH_IMPORTED:
                raise RuntimeError(f"import err: {_import_error}")
            ds_client = get_tensor_client()
            out_tensor = torch.zeros(size=arg.size, dtype=arg.dtype, device=f'npu:{current_device()}')
            _logger.info(f"start get tensor from ds, arg id is {arg.id}, current device is {current_device()}")
            failed_keys = ds_client.dev_mget([arg.id], [out_tensor])
            if failed_keys:
                _logger.info(f"dev_mget failed, arg id is {arg.id}, failed key is {failed_keys}")
                raise RuntimeError(f"dev_mget failed, failed keys: {failed_keys}")
            return out_tensor
        elif isinstance(arg, list):
            return [dev_mget_from_ds(item) for item in arg]
        elif isinstance(arg, tuple):
            return tuple(dev_mget_from_ds(item) for item in arg)
        elif isinstance(arg, set):
            return {dev_mget_from_ds(item) for item in arg}
        elif isinstance(arg, dict):
            return {key: dev_mget_from_ds(value) for key, value in arg.items()}
        return arg

    for arg in args:
        new_args.append(dev_mget_from_ds(arg))
    for k, v in kwargs.items():
        new_kwargs[k] = dev_mget_from_ds(v)
    return new_args, new_kwargs


class FunctionHandler(HandlerIntf):
    """
    Function handler class
    """

    def execute_function(
            self, func_meta, args: List, invoke_type, return_num: int,
            is_actor_async: bool) -> Tuple[List[object], ErrorInfo]:
        """execute function"""
        result = None
        try:
            if invoke_type == InvokeType.CreateInstance:
                self.__create_instance(func_meta, args)
            elif invoke_type == InvokeType.InvokeFunction:
                result = self.__invoke_instance(func_meta, args, is_actor_async)
            elif invoke_type == InvokeType.CreateInstanceStateless:
                self.__initialize_stateless_instance(func_meta)
            elif invoke_type == InvokeType.InvokeFunctionStateless:
                result = self.__invoke_stateless_function(func_meta, args)
            elif invoke_type == InvokeType.GetNamedInstanceMeta:
                result = InstanceManager().class_code
            elif invoke_type == InvokeType.DeleteRemoteTensor:
                result = self.__delete_tensors(args)
            else:
                msg = f"invalid invoke type {invoke_type}"
                _logger.warning(msg)
                return [], ErrorInfo(ErrorCode.ERR_EXTENSION_META_ERROR, ModuleCode.RUNTIME, msg)
        except Exception as err:
            _logger.warning("failed to execute user function, err: %s", err_to_str(err))
            if isinstance(err, YRInvokeError):
                result = [YRInvokeError(err.cause, traceback.format_exc())]
            else:
                result = [YRInvokeError(err, traceback.format_exc())]
            return result, ErrorInfo(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME,
                                     f"failed to execute user function, err: {repr(err)}")

        try:
            result_new = self.__check_return_list(result, return_num)
        except (TypeError, ValueError) as e:
            result_new = [YRInvokeError(e, traceback.format_exc())]
            _logger.error("Errors found during checking return values list, error: %s", e)
            return result_new, ErrorInfo(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME,
                                         f"failed to execute user function, err: {repr(e)}")

        return result_new, ErrorInfo()

    def shutdown(self, grace_period_second: int) -> ErrorInfo:
        """shutdown"""
        _logger.debug("Start to call user shutdown function __yr_shutdown__")
        instance = InstanceManager().instance()
        if instance is None:
            return ErrorInfo(
                ErrorCode.ERR_INNER_SYSTEM_ERROR,
                ModuleCode.RUNTIME,
                f"Failed to invoke instance function [{USER_SHUTDOWN_FUNC_NAME}], instance has not been initialized",
            )

        shutdown_func = None
        try:
            shutdown_func = getattr(instance, USER_SHUTDOWN_FUNC_NAME)
        except AttributeError:
            return ErrorInfo()
        try:
            shutdown_func(grace_period_second)
            _logger.info("Succeeded to call user shutdown function __yr_shutdown__")
        except Exception as e:
            _logger.exception(e)
            return ErrorInfo(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME, err_to_str(e))
        return ErrorInfo()

    def before_snapshot(self) -> ErrorInfo:
        """
        Public method to trigger snapshot hook, called by libruntime before snapshot.

        Returns:
            ErrorInfo: Error information if hook execution failed.
        """
        _logger.debug("Trigger before_snapshot hook")
        instance = InstanceManager().instance()
        return self.__before_snapshot(instance)

    def after_snapstart(self) -> ErrorInfo:
        """
        Public method to trigger snapstart hook, called by libruntime after restore.

        Returns:
            ErrorInfo: Error information if hook execution failed.
        """
        _logger.debug("Trigger after_snapstart hook")
        instance = InstanceManager().instance()
        return self.__after_snapstarted(instance)

    def __before_snapshot(self, instance) -> ErrorInfo:
        """
        Call user-defined __yr_before_snapshot__ hook before taking snapshot.

        Args:
            instance: The user instance object.

        Returns:
            ErrorInfo: Error information if hook execution failed.
        """
        if instance is None:
            msg = (
                f"Failed to invoke instance function [{USER_BEFORE_SNAPSHOT_FUNC_NAME}], "
                "instance has not been initialized"
            )
            return ErrorInfo(
                ErrorCode.ERR_INNER_SYSTEM_ERROR,
                ModuleCode.RUNTIME,
                msg,
            )

        # Check if user defined the hook method
        if not hasattr(instance, USER_BEFORE_SNAPSHOT_FUNC_NAME):
            _logger.debug(f"User hook {USER_BEFORE_SNAPSHOT_FUNC_NAME} not defined, skipping")
            return ErrorInfo()

        snapshot_hook = getattr(instance, USER_BEFORE_SNAPSHOT_FUNC_NAME)
        if not callable(snapshot_hook):
            _logger.debug(f"User hook {USER_BEFORE_SNAPSHOT_FUNC_NAME} is not callable, skipping")
            return ErrorInfo()

        _logger.debug(f"Start to call user snapshot hook {USER_BEFORE_SNAPSHOT_FUNC_NAME}")
        try:
            snapshot_hook()
            _logger.info(
                f"Succeeded to call user snapshot hook {USER_BEFORE_SNAPSHOT_FUNC_NAME}"
            )
        except Exception as e:
            _logger.exception(e)
            return ErrorInfo(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME, err_to_str(e))
        return ErrorInfo()

    def __after_snapstarted(self, instance) -> ErrorInfo:
        """
        Call user-defined __yr_after_snapstart__ hook after restoring from snapshot.

        Args:
            instance: The user instance object.

        Returns:
            ErrorInfo: Error information if hook execution failed.
        """
        if instance is None:
            msg = (
                f"Failed to invoke instance function [{USER_AFTER_SNAPSTART_FUNC_NAME}], "
                "instance has not been initialized"
            )
            return ErrorInfo(
                ErrorCode.ERR_INNER_SYSTEM_ERROR,
                ModuleCode.RUNTIME,
                msg,
            )

        # Check if user defined the hook method
        if not hasattr(instance, USER_AFTER_SNAPSTART_FUNC_NAME):
            _logger.debug(f"User hook {USER_AFTER_SNAPSTART_FUNC_NAME} not defined, skipping")
            return ErrorInfo()

        snapstart_hook = getattr(instance, USER_AFTER_SNAPSTART_FUNC_NAME)
        if not callable(snapstart_hook):
            _logger.debug(f"User hook {USER_AFTER_SNAPSTART_FUNC_NAME} is not callable, skipping")
            return ErrorInfo()

        _logger.debug(f"Start to call user snapstart hook {USER_AFTER_SNAPSTART_FUNC_NAME}")
        try:
            snapstart_hook()
            _logger.info(f"Succeeded to call user snapstart hook {USER_AFTER_SNAPSTART_FUNC_NAME}")
        except Exception as e:
            _logger.exception(e)
            return ErrorInfo(ErrorCode.ERR_INNER_SYSTEM_ERROR, ModuleCode.RUNTIME, err_to_str(e))
        return ErrorInfo()

    def __create_instance(self, func_meta, args) -> None:
        _logger.info("%s" % func_meta)
        class_code = CodeManager().load_code(func_meta, True)
        if class_code is None:
            raise RuntimeError(f"Failed to load code from data system, code id: [{func_meta.codeID}]")
        InstanceManager().class_code = class_code
        InstanceManager().set_code_ref(func_meta.codeID, False)
        InstanceManager().is_async = func_meta.isAsync

        args, kwargs = self.__get_param(args)
        instance = class_code(*args, **kwargs)
        InstanceManager().init(instance)

    def __invoke_instance(self, func_meta, args, is_actor_async) -> object:
        enable_tensor_transport = func_meta.enableTensorTransport
        tensor_transport_target = func_meta.tensorTransportTarget
        args, kwargs = self.__get_param(args)

        if tensor_transport_target == 'npu':
            args, kwargs = _get_tensor_from_ds(*args, **kwargs)

        instance_function_name = func_meta.functionName
        instance = InstanceManager().instance()
        if instance is None:
            raise RuntimeError(
                f"Failed to invoke instance function [{instance_function_name}], instance has not been initialized")

        def sync_to_async(func):
            if inspect.iscoroutinefunction(func) or inspect.isasyncgenfunction(func):
                return func

            async def wrapper(*args, **kwargs):
                res = func(*args, **kwargs)
                if inspect.isawaitable(res):
                    return await res
                return res

            return wrapper

        func = getattr(instance, instance_function_name)
        if not func_meta.isAsync and is_actor_async:
            func = sync_to_async(func)
        results = func(*args, **kwargs)
        if enable_tensor_transport:
            return _store_tensor_to_ds(results, instance_function_name)
        return results

    def __delete_tensors(self, args):
        args, _ = self.__get_param(args)
        return InstanceManager().erase_tensors(args)

    def __initialize_stateless_instance(self, func_meta):
        if func_meta.initializerCodeID:
            initializer_code = CodeManager().load_code_from_datasystem(func_meta.initializerCodeID)
            initializer_code()

    def __invoke_stateless_function(self, func_meta, args):
        code = CodeManager().load_code(func_meta)
        if code is None:
            raise RuntimeError(f"Failed to load code from data system, code id: [{func_meta.codeID}]")

        args, kwargs = self.__get_param(args)
        result = code(*args, **kwargs)
        return result

    def __get_param(self, args: List) -> Tuple[object, object]:
        params = yr.serialization.Serialization().deserialize(args)
        args, kwargs = recover_args(params)
        return args, kwargs

    def __check_return_list(self, result, return_num: int):
        if isinstance(result, YRInvokeError) or return_num == 1:
            return [result] * return_num
        if return_num > 1:
            if not hasattr(result, "__len__"):
                raise TypeError(f"cannot unpack non-iterable {type(result)} object")
            if len(result) != return_num:
                raise ValueError(f"not enough values to unpack (expected {return_num}, got {len(result)})")
        return result
