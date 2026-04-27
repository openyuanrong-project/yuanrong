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

"""code manager"""

import importlib.util
import logging
import os
import re
import sys
import threading
from typing import Callable, List

from yr.common import constants, utils
from yr.common.singleton import Singleton
from yr.err_type import ErrorCode, ErrorInfo, ModuleCode
from yr.functionsdk.error_code import FaasErrorCode
from yr.libruntime_pb2 import LanguageType

_DEFAULT_ADMIN_FUNC_PATH = "/adminfunc/"
_MAX_FAAS_ENTRY_NUMS = 3
_MIN_FAAS_ENTRY_NUMS = 2

_logger = logging.getLogger(__name__)


def _are_faas_entries(code_paths: List[str]) -> bool:
    if len(code_paths) > _MAX_FAAS_ENTRY_NUMS or len(code_paths) < _MIN_FAAS_ENTRY_NUMS:
        return False
    match_re = re.match(constants.PATTERN_FAAS_ENTRY, code_paths[constants.INDEX_SECOND]) is not None
    # For example, '/dcache/init.d/bucket-id' is also not a valid faas entry.
    return match_re


@Singleton
class CodeManager:
    """
    CodeManager provides code loading service
    it will try to cache the code and load from the cache as possible
    so it will only load
    * one module at most once (identified by the module file path)
    * one function at most once (identified by the module and function name)
    """

    def __init__(self):
        # code_key : code
        self.code_map = {}
        # code_key: code_path
        self.entry_map = {}
        self.module_cache = {}
        self.__lock = threading.Lock()
        self.custom_handler = os.environ.get(constants.ENV_KEY_ENV_DELEGATE_DOWNLOAD)
        self.deploy_dir = os.environ.get(constants.ENV_KEY_FUNCTION_LIBRARY_PATH)
        self.load_code_from_datasystem_func: Callable = None
        self.load_code_from_bytes: Callable = None

    def clear(self):
        """clear"""
        self.code_map.clear()
        self.entry_map.clear()
        self.module_cache.clear()

    def register_load_code_from_datasystem_func(self, function: Callable):
        """
        register load code from datasystem
        """
        self.load_code_from_datasystem_func = function

    def register_load_code_from_bytes_func(self, function: Callable):
        """
        register load code from bytes
        """
        self.load_code_from_bytes = function
        
    def register(self, function_key: str, function_obj: Callable):
        """
        register function code to code manager
        """
        with self.__lock:
            self.code_map[function_key] = function_obj

    def load(self, function_key):
        """
        load function code
        """
        with self.__lock:
            if function_key not in self.code_map:
                _logger.info("Code has not been registered in cache, code key: %s", function_key)
                return
            code = self.code_map.get(function_key)

        if code is None:
            # Although the 'register' method has raised an error if code is None,
            # the 'register' method and the 'load' method are not in the same thread,
            # the 'load' method will be executed even though the 'register' raise error.
            _logger.error("Failed to load code from local cache, code is None. Code key: %s", function_key)
            return

        _logger.debug("Succeeded to load code. Code key: %s", function_key)
        return code

    def get_code_path(self, code_key):
        """
        get code path
        """
        with self.__lock:
            return self.entry_map.get(code_key, None)

    def load_code(self, func_meta, is_class=False):
        """
        load user code
        """
        if len(func_meta.code) != 0:
            code = self.load_code_from_bytes(func_meta.code)
            if code is not None:
                if is_class:
                    self.register_class_lookup_alias(func_meta, code)
                return code
        if len(func_meta.codeID) != 0:
            code = self.load_code_from_datasystem(func_meta.codeID)
            if is_class:
                self.register_class_lookup_alias(func_meta, code)
            return code
        _logger.debug(f'there is no code id in meta, try to load function from local {func_meta}')
        code_name = func_meta.className if is_class else func_meta.functionName
        code = self.load_code_from_local(self.deploy_dir, func_meta.moduleName, code_name)
        _logger.debug(f'finished load function from local {self.deploy_dir}, module name '
                               f'{func_meta.moduleName}, code name {code_name}')
        return code

    def load_code_from_datasystem(self, code_id: str, language: LanguageType = LanguageType.Python):
        """
        load code from datasystem
        """
        if language != LanguageType.Python:
            # add cross language
            _logger.error("Intermodulation of functions in different languages is not supported.")

        code = self.load(code_id)
        if code is not None:
            return code

        if self.load_code_from_datasystem_func is None:
            msg = (
                f"Failed to load code from data system, code ID: [{code_id}]."
                "Function for getting code from data system has not been not registered."
            )
            _logger.error(msg)
            raise RuntimeError(msg)
        code = self.load_code_from_datasystem_func(code_id)
        if code is None:
            msg = f"Failed to import function from datasystem by {code_id}"
            _logger.error(msg)
            raise ImportError(msg)

        _logger.debug("Succeeded to load code from datasystem. Code key: %s", code_id)
        self.register(code_id, code)
        return code

    def load_code_from_local(self, code_dir, module_name, entry_name, code_key: str = ""):
        """
        get code from local,
        throw error if module not exists
        return None if module exists but entry not exists
        """
        # For SDK built-in modules (starting with 'yr.'), use direct import
        # instead of file path loading to support cross-version compatibility
        if module_name.startswith("yr."):
            code_dir = None

        _logger.debug("get python code [%s] from local file [%s/%s.py]",
                               entry_name, code_dir, module_name)
        code_key = module_name + "%%" + entry_name if code_key == "" else code_key
        code = self.load(code_key)
        if code is not None:
            return code

        module = self.__load_module(code_dir, module_name)
        code = getattr(module, entry_name, None)

        # Handle YRInstance proxy objects - extract the original user class
        # When a class is decorated with @yr.instance, it gets wrapped in a proxy
        # For skip_serialize mode, we need the original unwrapped class
        if code is not None and hasattr(code, "get_original_cls"):
            # Use public method to get the original class
            code = code.get_original_cls()
        elif code is not None and hasattr(code, "__user_class__"):
            code = code.__user_class__

        if code is None:
            msg = f"Failed to import function {entry_name} from {module_name}"
            _logger.error(msg)
            raise ImportError(msg)
        self.register(code_key, code)
        _logger.debug("Succeeded to load code from file %s/%s.py. code_key: %s, entry_name: %s",
                               code_dir, module_name, code_key, entry_name)
        return code

    def load_module_impl_for_testing(self, code_dir, module_name):
        """
        Public interface for testing _load_module_impl.

        This method provides access to the protected _load_module_impl method for testing purposes.
        """
        return self._load_module_impl(code_dir, module_name)

    def register_class_lookup_alias(self, func_meta, class_obj):
        """
        Register moduleName%%className so get_instance_by_name can resolve the class when
        runtime returns meta without inline code or codeID (e.g. nested class qualnames are
        not importable via getattr on the module).
        """
        if class_obj is None or not isinstance(class_obj, type):
            return
        if not func_meta.moduleName or not func_meta.className:
            return
        key = func_meta.moduleName + "%%" + func_meta.className
        self.register(key, class_obj)

    def load_functions(self, code_paths: List[str]) -> ErrorInfo:
        """
        Load code paths and return ErrorInfo object.

        Args:
            code_paths (List[str]): List of paths to load.

        Returns:
            ErrorInfo: Error information for any errors encountered during loading.
        """
        if _are_faas_entries(code_paths):
            # to judge 'code_paths' are FaaS entries.
            for code_path, code_key in zip(code_paths, [constants.KEY_USER_INIT_ENTRY, constants.KEY_USER_CALL_ENTRY,
                                                        constants.KEY_USER_SHUT_DOWN_ENTRY]):
                with self.__lock:
                    self.entry_map[code_key] = code_path
                error_info = self.__load_faas_entry(code_path, code_key)
                if error_info.error_code != ErrorCode.ERR_OK:
                    return error_info
        else:
            for code_path in code_paths:
                sys.path.insert(constants.INDEX_FIRST, code_path)
        return ErrorInfo()

    def __load_faas_entry(self, user_entry: str, code_key: str):
        """
        load module and the entry code, throw RuntimeError if failed.
        """
        user_hook_length = 2
        _logger.debug("Faas load module and entry [%s] from [%s]", user_entry, self.custom_handler)
        user_hook_splits = user_entry.rsplit(".", maxsplit=1) if isinstance(user_entry, str) else None
        if len(user_hook_splits) != user_hook_length:
            if code_key == constants.KEY_USER_INIT_ENTRY:
                return ErrorInfo()
            msg = convert_response_to_jsonstr(
                "User hook not satisfy requirement, expect: xxx.xxx",
                FaasErrorCode.INIT_FUNCTION_FAIL)
            return ErrorInfo(ErrorCode.ERR_USER_CODE_LOAD, ModuleCode.RUNTIME, msg)

        user_module, user_entry = user_hook_splits[0], user_hook_splits[1]
        _logger.debug("User module: %s, entry: %s", user_module, user_entry)

        try:
            user_code = self.load_code_from_local(self.custom_handler, user_module, user_entry, code_key)
        except ValueError as exp:
            _logger.error("Missing user module: %s, exception: %s", user_entry, exp)
            msg = convert_response_to_jsonstr(f"Missing user module: {user_entry}, err: {exp}",
                                              FaasErrorCode.ENTRY_EXCEPTION)
            return ErrorInfo(ErrorCode.ERR_USER_CODE_LOAD, ModuleCode.RUNTIME, msg)
        except ImportError as exp:
            _logger.error("Failed to import user module: %s, exception: %s", user_entry, exp)
            msg = convert_response_to_jsonstr(f"Failed to import user module: {user_entry}, err: {exp}",
                                              FaasErrorCode.ENTRY_EXCEPTION)
            return ErrorInfo(ErrorCode.ERR_USER_CODE_LOAD, ModuleCode.RUNTIME, msg)
        except SyntaxError as exp:
            _logger.error("Failed to load user code: %s, exception: %s", user_entry, exp)
            msg = convert_response_to_jsonstr("Failed to load user code. There is syntax error in user code: "
                                              f"{user_entry}, err: {exp}", FaasErrorCode.ENTRY_EXCEPTION)
            return ErrorInfo(ErrorCode.ERR_USER_CODE_LOAD, ModuleCode.RUNTIME, msg)
        except Exception as exp:
            _logger.error("Failed to load user code: %s, exception: %s", user_entry, exp)
            msg = convert_response_to_jsonstr(f"Failed to load user code: {user_entry}, err: {exp}",
                                              FaasErrorCode.ENTRY_EXCEPTION)
            return ErrorInfo(ErrorCode.ERR_USER_CODE_LOAD, ModuleCode.RUNTIME, msg)

        if user_code is None:
            _logger.error("Missing user entry. %s", user_entry)
            msg = convert_response_to_jsonstr(f"Missing user entry. {user_entry}", FaasErrorCode.ENTRY_EXCEPTION)
            return ErrorInfo(ErrorCode.ERR_USER_CODE_LOAD, ModuleCode.RUNTIME, msg)
        return ErrorInfo()

    def __load_module(self, code_dir, module_name):
        """load module using cache"""
        return self._load_module_impl(code_dir, module_name)

    def _load_module_impl(self, code_dir, module_name):
        """load module using cache (implementation for testing)"""
        if code_dir is None:
            _logger.debug(
                "code dir is None, import from module name %s directly.", module_name)
            module = self.module_cache.get(module_name, None)
            if module is not None:
                return module
            module = importlib.import_module(module_name)
            self.module_cache[module_name] = module
            return module

        file_path = os.path.join(code_dir, module_name.replace(".", os.sep) + ".py")
        if not os.path.exists(file_path):
            admin_path = os.path.join(_DEFAULT_ADMIN_FUNC_PATH, module_name.replace(".", os.sep) + ".py")
            if not os.path.exists(admin_path):
                raise ValueError("entry file does not exist: [{}]".format(file_path))
            file_path = admin_path

        module = self.module_cache.get(file_path, None)
        if module is not None:
            _logger.debug("successfully load module [%s] from cache", file_path)
            return module

        _logger.debug("loading module [%s] from file system", file_path)
        module_spec = importlib.util.spec_from_file_location(module_name, file_path)
        try:
            module = importlib.util.module_from_spec(module_spec)
        except ImportError as exp:
            _logger.warning("failed to import user python module, %s", str(exp))
            raise

        try:
            # In case of user code errors, the exec_module operation does not throw errors
            # in non-main threads, user code would be None when execute.
            module_spec.loader.exec_module(module)
        except (ImportError, SyntaxError) as exp:
            _logger.warning("Failed to exec_module user module, please check user code. %s", exp)
            raise
        self.module_cache[file_path] = module
        return module


def convert_response_to_jsonstr(message: str, status_code: FaasErrorCode) -> str:
    """Method transform_call_response_to_str"""
    message = "" if message is None else message
    result = dict(
        message=message,
        errorCode=str(status_code.value)
    )
    return utils.to_json_string(result)
