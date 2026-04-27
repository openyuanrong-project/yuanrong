#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

"""
Function decorator
"""
import inspect
import logging
import threading
import types
from functools import wraps
from typing import List, Union

import yr
from yr import signature
from yr.config import function_group_enabled
from yr.common.types import GroupInfo
from yr.common import utils
from yr.common.utils import CrossLanguageInfo, ObjectDescriptor
from yr.config import InvokeOptions
from yr.libruntime_pb2 import FunctionMeta, LanguageType
from yr.object_ref import ObjectRef
from yr.runtime_holder import global_runtime
from yr.generator import ObjectRefGenerator
from yr.serialization import Serialization

_logger = logging.getLogger(__name__)


class FunctionProxy:
    """
    Use to decorate user function.

    Examples:
        >>> import yr
        >>>
        >>> yr.init()
        >>>
        >>> @yr.invoke
        ... def add(a, b):
        ...     return a + b
        >>>
        >>> ret = add.invoke(1, 2)
        >>> print(yr.get(ret))
        >>>
        >>> yr.finalize()
    """

    def __init__(self, func, cross_language_info=None, invoke_options=None, return_nums=None, initializer=None):
        """
        Initialize a FunctionProxy instance.
        """
        self.cross_language_info = cross_language_info
        self.invoke_options = invoke_options
        self._code_ref = None
        self._code = None
        self._initializer_code_ref = None
        self.__original_func__ = func
        self._is_generator = inspect.isgeneratorfunction(func) or inspect.isasyncgenfunction(func)
        self._initializer = initializer
        if self.invoke_options is None:
            self.invoke_options = InvokeOptions()
        self.designated_urn = ""
        self.function_group_size = 0
        self.group_name = ""
        self.sig = None
        self.func = func
        if func is not None:
            self.sig = signature.get_signature(func)
        if self.cross_language_info is None and func is not None:
            self.function_descriptor = ObjectDescriptor.get_from_function(func)
        if self.cross_language_info is not None:
            self.function_descriptor = ObjectDescriptor(function_name=self.cross_language_info.function_name,
                                                        target_language=self.cross_language_info.target_language,
                                                        class_name=self.cross_language_info.class_name)
        self._lock = threading.Lock()
        if return_nums is None:
            self.return_nums = 1
        else:
            if not isinstance(return_nums, int):
                raise TypeError(f"invalid return_nums type: {type(return_nums)}, should be an int")
            if return_nums < 0 or return_nums > 100:
                raise RuntimeError(f"invalid return_nums: {return_nums}, should be an integer between 0 and 100")
            self.return_nums = return_nums
        if self.cross_language_info is not None:
            def _cross_invoke_proxy(*args, **kwargs):
                return self._invoke(func, args=args, kwargs=kwargs)

            self.invoke = _cross_invoke_proxy
            self.remote = _cross_invoke_proxy
        else:
            @wraps(func)
            def _invoke_proxy(*args, **kwargs):
                return self._invoke(func, args=args, kwargs=kwargs)

            self.invoke = _invoke_proxy

    def __call__(self, *args, **kwargs):
        """
        invalid call
        """
        raise RuntimeError("invoke function cannot be called directly")

    def __getstate__(self):
        attrs = self.__dict__.copy()
        del attrs["_lock"]
        del attrs["_code_ref"]
        del attrs["_code"]
        del attrs["_initializer_code_ref"]
        return attrs

    def __setstate__(self, state):
        self.__dict__.update(state)
        self.__dict__["_lock"] = threading.Lock()
        self.__dict__["_code_ref"] = None
        self.__dict__["_code"] = None
        self.__dict__["_initializer_code_ref"] = None

    def options(self, invoke_options: InvokeOptions) -> "FunctionProxy":
        """Modify the parameters of the decorated function dynamically.

        This method is used to dynamically modify the parameters of a remote function call and return the current
        function proxy.

        Note:
            This interface does not take effect in local mode.

        Args:
            invoke_options (InvokeOptions): The invoke options.

        Returns:
            The FunctionProxy object itself. Data type is FunctionProxy.

        Examples:
            >>> import yr
            >>>
            >>> yr.init()
            >>>
            >>> @yr.invoke
            ... def add(a, b):
            ...    return a + b
            >>>
            >>> opt = yr.InvokeOptions(cpu=1000, memory=1024)
            >>> ret = add.options(opt).invoke(1, 2)
            >>> print(yr.get(ret))
            >>>
            >>> yr.finalize()

        """
        for method in ("check_options_valid", "check_options_range"):
            if hasattr(invoke_options, method):
                getattr(invoke_options, method)()
        self.invoke_options = invoke_options
        return self

    def set_urn(self, urn: str) -> "FunctionProxy":
        """
        Set urn for this function

        Args:
            urn (str): The designated urn.

        Returns:
            The FunctionProxy object itself. Data type is FunctionProxy.
        """
        self.designated_urn = utils.get_function_from_urn(urn)
        return self

    def set_function_group_size(self, function_group_size: int) -> "FunctionProxy":
        """
        Set group size for this function.

        Args:
            function_group_size (int): The function group size.

        Returns:
            The FunctionProxy object itself. Data type is FunctionProxy.
        """
        self.function_group_size = function_group_size
        return self

    def get_original_func(self) -> any:
        """
        Get the unwrapped function

        Returns:
            The unwrapped function. Data type is any.
        """
        return self.__original_func__

    def invoke(self, *args, **kwargs):
        """
        Execute the decorated function remotely.

        This method triggers the execution of the decorated function on remote workers.
        The function is executed with the provided arguments and returns an ObjectRef
        that can be used to retrieve the result.

        Args:
            *args: Variable arguments to pass to the decorated function.
            **kwargs: Keyword arguments to pass to the decorated function.

        Returns:
            ObjectRef or List[ObjectRef]: A reference to the result object(s). 
            For functions with return_nums=1, returns a single ObjectRef.
            For functions with return_nums>1, returns a list of ObjectRefs.
            For generator functions, returns an ObjectRefGenerator.
            For functions with return_nums=0, returns None.

        Raises:
            TypeError: If the provided arguments don't match the function signature.
            RuntimeError: If the function execution fails or runtime is not initialized.

        Examples:
            >>> import yr
            >>>
            >>> yr.init()
            >>>
            >>> @yr.invoke
            ... def add(a, b):
            ...     return a + b
            >>>
            >>> result_ref = add.invoke(1, 2)
            >>> result = yr.get(result_ref)
            >>> print(result)  # Output: 3
            >>>
            >>> # For functions with multiple return values
            >>> @yr.invoke(return_nums=2)
            ... def divmod_func(a, b):
            ...     return divmod(a, b)
            >>>
            >>> quotient_ref, remainder_ref = divmod_func.invoke(10, 3)
            >>> print(yr.get(quotient_ref))  # Output: 3
            >>> print(yr.get(remainder_ref))  # Output: 1
            >>>
            >>> yr.finalize()
        """
        # This is a placeholder method for documentation purposes.
        # The actual implementation is dynamically assigned in __init__.
        raise NotImplementedError("This method is dynamically overridden in __init__")

    def create_opts_wrapper(self, opts: InvokeOptions):
        """
        Public interface to safely wrap invoke options.

        Args:
            opts (InvokeOptions): the invoke options to wrap

        Returns:
            An instance of FunctionProxyWrapper
        """
        return self._options_wrapper(opts)

    def invoke_function_for_testing(self, opts: InvokeOptions, func, args=None, kwargs=None) -> Union[
            "yr.ObjectRef", List["yr.ObjectRef"]]:
        """
        Invoke function for testing purposes.

        This is a public wrapper around _invoke_function for testing.

        Args:
            opts: The invoke options.
            func: The function to invoke.
            args: The positional arguments.
            kwargs: The keyword arguments.

        Returns:
            A reference to the data object or list of references.
        """
        return self._invoke_function(opts, func, args, kwargs)

    def _invoke_function(self, opts: InvokeOptions, func, args=None, kwargs=None) -> Union[
            "yr.ObjectRef", List["yr.ObjectRef"]]:
        """
        The real realization of the invoke function

        Returns:
            A reference to the data object.

        Raises:
            TypeError: This exception is thrown if the type of the passed parameter is incorrect.
        """
        function_id = self.designated_urn
        if self.cross_language_info is None:
            args_list = signature.package_args(self.sig, args, kwargs)
        else:
            function_id = self.cross_language_info.function_key
            args_list = utils.make_cross_language_args(args, kwargs)
        with self._lock:
            if self.cross_language_info is None and (
                    self._code_ref is None
                    or not global_runtime.get_runtime().is_object_existing_in_local(
                        self._code_ref.id
                        )):
                serialized_object = Serialization().serialize(func)
                if len(serialized_object) <= 102400:
                    self._code = serialized_object.to_bytes()
                    _logger.debug("[Reference Counting] pass code by request, functionName = %s", func.__qualname__)
                self._code_ref = ObjectRef(
                    global_runtime.get_runtime().put_serialized(serialized_object),
                    need_incre=False
                )
                _logger.debug("[Reference Counting] put code with id = %s, functionName = %s",
                              self._code_ref.id, func.__qualname__)
        with self._lock:
            if self._initializer and self._initializer_code_ref is None:
                self._initializer_code_ref = yr.put(self._initializer)

        initializer_code_id = ""
        if self._initializer_code_ref is not None:
            initializer_code_id = self._initializer_code_ref.id
        func_meta = FunctionMeta(functionID=function_id,  # if designated_urn is not set,
                                 # use function id in the config
                                 moduleName=self.function_descriptor.module_name,
                                 className=self.function_descriptor.class_name,
                                 functionName=self.function_descriptor.function_name,
                                 language=self.function_descriptor.target_language,
                                 codeID=self._code_ref.id if self._code_ref is not None else "",
                                 initializerCodeID=initializer_code_id,
                                 isGenerator=self._is_generator,
                                 code=self._code if self._code is not None else b"",
                                 )
        return_nums = 1 if (self.return_nums == 0 or self._is_generator) else self.return_nums
        group_enabled = function_group_enabled(self.invoke_options.function_group_options, self.function_group_size)
        self.group_name = global_runtime.get_runtime().generate_group_name() if group_enabled else ""
        return_nums = return_nums * self.function_group_size if group_enabled else return_nums
        obj_list = global_runtime.get_runtime().invoke_by_name(func_meta=func_meta,
                                                               args=args_list,
                                                               opt=opts,
                                                               return_nums=return_nums,
                                                               group_info=GroupInfo(group_size=self.function_group_size,
                                                                                    group_name=self.group_name))
        if self.return_nums == 0:
            return None
        objref_list = []
        for i in obj_list:
            objref_list.append(ObjectRef(i, need_incre=False))

        if self._is_generator:
            return ObjectRefGenerator(objref_list[0])

        return objref_list[0] if return_nums == 1 else objref_list

    def _options_wrapper(self, invoke_options: InvokeOptions):
        """
        options wrapper, Set user invoke options.
        """
        function_cls = self

        class FunctionProxyWrapper:
            """instance option wrapper"""

            def invoke(self, *args, **kwargs):
                """invoke"""
                return function_cls._invoke_function(opts=invoke_options,
                                                     func=function_cls.func,
                                                     args=args,
                                                     kwargs=kwargs or {},
                                                     )
        return FunctionProxyWrapper()

    def _invoke(self, func, args=None, kwargs=None):
        """
        Triggers the execution of the function with the given function, arguments, and keyword arguments.
        Calls the _invoke_function method to perform the function invocation.
        """
        return self._invoke_function(self.invoke_options, func, args, kwargs)


def make_decorator(invoke_options=None, return_nums=None, initializer=None) -> callable:
    """
    Make decorator for invoke function
    """

    def decorator(func):
        if isinstance(func, types.FunctionType):
            return FunctionProxy(func, invoke_options=invoke_options, return_nums=return_nums, initializer=initializer)
        raise RuntimeError("@yr.invoke decorator must be applied to a function")

    return decorator


def make_cpp_function_proxy(function_name, function_key):
    """
    Make proxy for invoke cpp function
    """
    return FunctionProxy(None, CrossLanguageInfo(function_name, function_key, LanguageType.Cpp))


def make_cross_language_function_proxy(function_name, function_urn, language):
    """
    Make decorator for invoke function
    """
    function_key = utils.get_function_from_urn(function_urn)

    return FunctionProxy(None, CrossLanguageInfo(function_name, function_key, language))


# Gradual migration: StatelessFunction is the new preferred name
# FunctionProxy is kept for backward compatibility
StatelessFunction = FunctionProxy
