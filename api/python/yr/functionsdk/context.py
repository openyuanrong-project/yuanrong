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

"""faas context"""

from dataclasses import dataclass, field
import logging
import os
import time
from typing import Any, Dict

from yr import log
from yr.common import constants
from yr.functionsdk.utils import parse_json_data_to_dict, dump_data_to_json_str

_RUNTIME_MAX_RESP_BODY_SIZE = 6 * 1024 * 1024
_RUNTIME_CODE_ROOT = "/opt/function/code"
_RUNTIME_ROOT = "/home/snuser/runtime"
_RUNTIME_LOG_DIR = "/home/snuser/log"
_ENV_STORAGE = None

_HEADER_ACCESS_KEY: str = "X-Access-Key"
_HEADER_SECRET_KEY: str = "X-Secret-Key"
_HEADER_AUTH_TOKEN: str = "X-Auth-Token"
_HEADER_SECURITY_ACCESS_KEY: str = "X-Security-Access-Key"
_HEADER_SECURITY_SECRET_KEY: str = "X-Security-Secret-Key"
_HEADER_SECURITY_TOKE: str = "X-Security-Token"
_HEADER_REQUEST_ID: str = "X-Request-Id"


def load_context_meta(context_meta: dict):
    """
    load context meta
    """
    global _ENV_STORAGE
    _ENV_STORAGE = EnvStorage()
    _ENV_STORAGE.load_context_meta(context_meta)
    _ENV_STORAGE.load_user_data(_decrypt_user_data())


def init_context(stage: str):
    """
    create the context for init handler of user code
    configuration user function logger
    """
    logger = logging.getLogger(__name__)
    options = {
        "logger": logger,
        "stage": stage,
    }
    context = Context(options=options)
    return context


def init_context_invoke(stage: str, header: dict):
    """
    create the context for  call handler of user code
    configuration user function logger
    """
    global _ENV_STORAGE
    _ENV_STORAGE.update_user_agency(header)
    context = init_context(stage)
    if _HEADER_REQUEST_ID in header:
        context.set_trace_id(header[_HEADER_REQUEST_ID])
    return context


class Context:
    """Context information provided by the openyuanrong runtime."""

    def __init__(self, options: dict):
        self.__project_id = _ENV_STORAGE.env_project_id
        self.__package = _ENV_STORAGE.env_package
        self.__function_name = _ENV_STORAGE.env_function_name
        self.__function_version = _ENV_STORAGE.env_function_version
        self.__user_data = _ENV_STORAGE.env_user_data
        self.__timeout = _ENV_STORAGE.env_timeout
        self.__memory = _ENV_STORAGE.env_memory
        self.__cpu = _ENV_STORAGE.env_cpu
        self.__start_time = int(time.time() * 1000)
        self.__logger = options.get('logger', logging.getLogger(__name__))
        self.__request_id = options.get('requestId', "")
        self.__tenant_id = options.get('tenantId', _ENV_STORAGE.env_project_id)
        self.__access_key = options.get('accessKey', _ENV_STORAGE.env_access_key)
        self.__secret_key = options.get('secretKey', _ENV_STORAGE.env_secret_key)
        self.__auth_token = options.get('authToken', _ENV_STORAGE.env_auth_token)
        self.__security_access_key = options.get('securityAccessKey', _ENV_STORAGE.env_security_access_key)
        self.__security_secret_key = options.get('securitySecretKey', _ENV_STORAGE.env_security_secret_key)
        self.__security_token = options.get('securityToken', _ENV_STORAGE.env_security_token)
        self.__alias = options.get('alias', _ENV_STORAGE.env_alias)
        self.state = None
        self.instance_id = None
        self.invoke_property = None
        self.future_id = options.get('future_id', "")
        self.invoke_id = options.get('invoke_id', "")

    # Gets the request ID associated with the request.
    def getRequestID(self):
        """
        Get request ID.

        Returns:
            request ID.
        """
        return self.__request_id

    def getProjectID(self):
        """
        Get project ID.

        Returns:
            project ID.
        """
        return self.__project_id

    def getTenantID(self):
        """
        Get tenant ID.

        Returns:
            tenant ID.
        """
        return self.__tenant_id

    def getPackage(self):
        """
        Get function package.

        Returns:
            function package.
        """
        return self.__package

    # Gets name of the function
    def getFunctionName(self):
        """
        Get name of the function.

        Returns:
            function name.
        """
        return self.__function_name

    def getAlias(self):
        """
        Get the alias of the function.

        Returns:
            the alias of the function.
        """
        return self.__alias

    # Get version of the function
    def getVersion(self):
        """
        Get version of the function.

        Returns:
            version of the function.
        """
        return self.__function_version

    # Get the memory size distributed the running function
    def getMemorySize(self):
        """
        Get the memory size distributed the running function.

        Returns:
            The memory resources occupied by the function.
        """
        return self.__memory

    # Get the number of cpu distributed to the running function the cpu
    # number scale by millicores, one cpu cores equals 1000 millicores. In
    # function stage runtime, every function have base of 200 millicores,
    # and increased by memory size distributed to function. the offset is
    # about Memory Size(M)/128 * 100
    def getCPUNumber(self):
        """
        Get the number of cpu distributed to the running function the cpu number scale by millicores, 
        one cpu cores equals 1000 millicores. In function stage runtime, every function have base of 200 millicores,
        and increased by memory size distributed to function. The offset is about Memory Size(M)/128 * 100

        Returns:
            The CPU resources occupied by the function.
        """
        return self.__cpu

    def getAccessKey(self):
        """
        Get the user's AccessKey (valid for 24 hours).
        To use this method, you need to configure delegation permissions for the function.
        The current function workflow has ceased maintenance of the getAccessKey interface in the Runtime SDK, 
        and you will no longer be able to obtain the temporary AccessKey via this interface.

        Returns:
            user's delegated AccessKey.
        """  
        return self.__access_key

    def setAccessKey(self, access_key):
        """Method setAccessKey, not exposed"""
        self.__access_key = access_key

    def getSecretKey(self):
        """
        Get the user's SecretKey (valid for 24 hours).
        To use this method, you need to configure delegation permissions for the function.
        The current function workflow has ceased maintenance of the getSecretKey interface in the Runtime SDK, 
        and you will no longer be able to obtain the temporary SecretKey via this interface.

        Returns:
            user's delegated SecretKey.
        """  
        return self.__secret_key

    def setSecretKey(self, secret_key):
        """Method SetSecretKey, not exposed"""
        self.__secret_key = secret_key

    def getAuthToken(self):
        """
        Get the user's delegated token (valid for 24 hours).
        To use this method, you need to configure delegation for the function.

        Returns:
            user's delegated token.
        """  
        return self.__auth_token

    def setAuthToken(self, auth_token):
        """Method setToken, not exposed"""
        self.__auth_token = auth_token

    def getSecurityAccessKey(self):
        """
        Get the user's delegated SecurityAccessKey (valid for 24 hours) with a cache duration of 10 minutes, 
        meaning the same content is returned if retrieved again within 10 minutes. 
        Using this method requires configuring delegation for the function.

        Returns:
            user's delegated SecurityAccessKey.
        """  
        return self.__security_access_key

    def setSecurityAccessKey(self, security_access_key):
        """Method setAccessKey, not exposed"""
        self.__security_access_key = security_access_key

    def getSecuritySecretKey(self):
        """
        Get the user's delegated SecuritySecretKey (valid for 24 hours) with a cache duration of 10 minutes, 
        meaning the same content is returned if retrieved again within 10 minutes. 
        Using this method requires configuring delegation for the function.

        Returns:
            user's delegated SecuritySecretKey.
        """        
        return self.__security_secret_key

    def setSecuritySecretKey(self, security_secret_key):
        """Method SetSecretKey, not exposed"""
        self.__security_secret_key = security_secret_key

    def getSecurityToken(self):
        """
        Get the user's delegated SecurityToken (valid for 24 hours), with a cache duration of 10 minutes, 
        meaning the returned content is the same if retrieved again within 10 minutes. 
        Using this method requires configuring delegation for the function.

        Returns:
            user's delegated SecurityToken.
        """        
        return self.__security_token

    def setSecurityToken(self, security_token):
        """Method getSecurityToken, not exposed"""
        self.__security_token = security_token

    # Gets the user data,which saved in a map
    def getUserData(self, key, default=None):
        """
        Get the value passed in by the user through environment variables via the key.

        Args:
            key (string): The key of the environment configured by the user.
            default (string): The default value when the user obtains an empty environment variable.

        Returns:
            The value corresponding to the key of the environment variable configured by the user.
        """
        return self.__user_data.get(key, default)

    # Gets the time distributed to the running of the function, when exceed
    # the specified time, the running of the function would be stopped by force
    def getRunningTimeInSeconds(self):
        """
        Get the function timeout period, in seconds.

        Returns:
            function timeout period.
        """
        return self.__timeout

    # Gets the time remaining for this execution in milliseconds
    # Returns time before task is killed
    def getRemainingTimeInMilliSeconds(self):
        """
        Get the remaining running time of the function, in milliseconds.

        Returns:
            remaining running time.
        """
        now = int(time.time() * 1000)
        return self.__timeout + self.__start_time - now

    # Gets the logger for user to log out in standard output, The Logger
    # interface must be provided in SDK
    def getLogger(self):
        """
        Get the logger for user to log out in standard output, 
        The Logger interface must be provided in SDK

        Returns:
            logger.

        Examples:
            >>> log = context.getLogger()
            >>> log.info("test")
        """
        return self.__logger

    def set_state(self, state):
        """Method set_state, not exposed"""
        self.state = state

    def get_state(self):
        """Method get_state, not exposed"""
        return self.state

    def set_instance_id(self, instance_id):
        """Method set_instance_id, not exposed"""
        self.instance_id = instance_id

    def get_instance_id(self):
        """Method get_instance_id, not exposed"""
        return self.instance_id

    def get_invoke_id(self):
        """Method get_invoke_id, not exposed"""
        return self.invoke_id

    def get_trace_id(self):
        """Method get_trace_id, not exposed"""
        return self.__request_id

    def set_trace_id(self, request_id):
        """Method get_trace_id, not exposed"""
        self.__request_id = request_id

    def get_invoke_property(self):
        """Method get_invoke_property, not exposed"""
        return self.invoke_property


@dataclass
class EnvStorage:
    """
    env storage
    """
    env_project_id: str = ""
    env_package: str = ""
    env_function_name: str = ""
    env_function_version: str = ""
    env_user_data: Dict = field(default_factory=dict)
    env_timeout: int = 0
    env_cpu: int = 0
    env_memory: int = 0
    env_access_key: str = ""
    env_secret_key: str = ""
    env_auth_token: str = ""
    env_alias: str = ""
    env_pre_stop_handler: str = ""
    env_pre_stop_timeout: str = ""
    env_security_access_key: str = ""
    env_security_secret_key: str = ""
    env_security_token: str = ""

    initializer_handler: str = ""
    initializer_timeout: str = ""
    name: str = ""
    handler: str = ""

    def load_context_meta(self, context_meta: dict):
        """
        load context
        """
        func_meta_data = _check_map_value(context_meta, 'funcMetaData', {})
        resource_meta_data = _check_map_value(context_meta, 'resourceMetaData', {})
        extended_meta_data = _check_map_value(context_meta, 'extendedMetaData', {})
        initializer = _check_map_value(extended_meta_data, 'initializer', {})
        pre_stop = _check_map_value(extended_meta_data, "pre_stop", {})

        self.env_project_id = _check_map_value(func_meta_data, 'tenantId', "")
        self.env_package = _check_map_value(func_meta_data, 'service', "")
        self.env_function_name = _check_map_value(func_meta_data, 'func_name', "")
        self.env_function_version = _check_map_value(func_meta_data, 'version', "")
        self.env_timeout = int(_check_map_value(func_meta_data, 'timeout', "3"))
        self.env_cpu = int(_check_map_value(resource_meta_data, 'cpu', "0"))
        self.env_memory = int(_check_map_value(resource_meta_data, 'memory', "0"))
        self.env_alias = context_meta.get('alias', "")
        self.env_pre_stop_handler = str(_check_map_value(pre_stop, "pre_stop_handler", ""))
        self.env_pre_stop_timeout = str(_check_map_value(pre_stop, "pre_stop_timeout", ""))

        initializer_handler = str(_check_map_value(initializer, 'initializer_handler', ""))
        initializer_timeout = str(_check_map_value(initializer, 'initializer_timeout', ""))
        name = _check_map_value(func_meta_data, 'name', "")
        hander = _check_map_value(func_meta_data, 'handler', "")
        self.__write_env(initializer_handler, initializer_timeout, name, hander)

    def load_user_data(self, user_data: Dict):
        """
        load user data
        """
        self.env_access_key = user_data.get("ENV_ACCESS_KEY", "")
        self.env_secret_key = user_data.get("ENV_SECRET_KEY", "")
        self.env_auth_token = user_data.get("ENV_AUTH_TOKEN", "")
        self.env_security_access_key = user_data.get("ENV_SECURITY_ACCESS_KEY", "")
        self.env_security_secret_key = user_data.get("ENV_SECURITY_SECRET_KEY", "")
        self.env_security_token = user_data.get("ENV_SECURITY_TOKEN", "")

        user_data["ENV_ALIAS"] = self.env_alias
        self.env_user_data = user_data
        os.environ["RUNTIME_USERDATA"] = dump_data_to_json_str(user_data)

    def update_user_agency(self, header: dict):
        """update user agency"""
        if _HEADER_ACCESS_KEY in header:
            self.env_access_key = header[_HEADER_ACCESS_KEY]
        if _HEADER_SECRET_KEY in header:
            self.env_secret_key = header[_HEADER_SECRET_KEY]
        if _HEADER_AUTH_TOKEN in header:
            self.env_auth_token = header[_HEADER_AUTH_TOKEN]
        if _HEADER_SECURITY_ACCESS_KEY in header:
            self.env_security_access_key = header[_HEADER_SECURITY_ACCESS_KEY]
        if _HEADER_SECURITY_SECRET_KEY in header:
            self.env_security_secret_key = header[_HEADER_SECURITY_SECRET_KEY]
        if _HEADER_SECURITY_TOKE in header:
            self.env_security_token = header[_HEADER_SECURITY_TOKE]

    def __write_env(self, initializer_handler, initializer_timeout, name, hander):
        os.environ["RUNTIME_PROJECT_ID"] = self.env_project_id
        os.environ["RUNTIME_PACKAGE"] = self.env_package
        os.environ["RUNTIME_FUNC_NAME"] = self.env_function_name
        os.environ["RUNTIME_FUNC_VERSION"] = self.env_function_version
        os.environ["RUNTIME_TIMEOUT"] = str(self.env_timeout)
        os.environ["RUNTIME_CPU"] = str(self.env_cpu)
        os.environ["RUNTIME_MEMORY"] = str(self.env_memory)
        os.environ["RUNTIME_INITIALIZER_HANDLER"] = initializer_handler
        os.environ["RUNTIME_INITIALIZER_TIMEOUT"] = initializer_timeout
        os.environ["RUNTIME_SERVICE_FUNC_VERSION"] = name
        os.environ["RUNTIME_HANDLER"] = hander
        os.environ["RUNTIME_ROOT"] = _RUNTIME_ROOT
        os.environ["RUNTIME_CODE_ROOT"] = _RUNTIME_CODE_ROOT
        os.environ["RUNTIME_LOG_DIR"] = _RUNTIME_LOG_DIR
        os.environ["RUNTIME_MAX_RESP_BODY_SIZE"] = str(_RUNTIME_MAX_RESP_BODY_SIZE)
        os.environ["PRE_STOP_HANDLER"] = self.env_pre_stop_handler
        os.environ["PRE_STOP_TIMEOUT"] = self.env_pre_stop_timeout


def _decrypt_user_data() -> dict:
    """Decrypts user data from environment variables and returns it as a dictionary.

    Args:
        alias (str): The alias of user function.

    Returns:
        dict: A dictionary containing the decrypted user data.
    """
    env_map = {}
    delegate_decrypt = parse_json_data_to_dict(os.environ.get('ENV_DELEGATE_DECRYPT', ""))
    # 'environment' could be None or '{}' string after parsing, to be compatible with these two cases,
    # the default value is '{}' (not {}) and still have to be parsed.
    environment = parse_json_data_to_dict(delegate_decrypt.get('environment', '{}'))
    encrypted_user_data = parse_json_data_to_dict(delegate_decrypt.get('encrypted_user_data', '{}'))

    log.get_logger().debug(
        f"Succeeded to read from ENV_DELEGATE_DECRYPT, delegate_decrypt={delegate_decrypt}, "
        f"environment={environment}, encrypted_user_data={encrypted_user_data}")

    # write environment values
    for key in environment:
        if key == constants.ENV_KEY_LD_LIBRARY_PATH:
            new_path = encrypted_user_data.get(
                constants.ENV_KEY_LD_LIBRARY_PATH,
                environment.get(constants.ENV_KEY_LD_LIBRARY_PATH, ""))
            env_map[key] = os.environ.get(key, "") + f":{new_path}"
            os.environ[key] = os.environ.get(key, "") + f":{new_path}"
        else:
            os.environ[key] = str(environment[key])
            env_map[key] = str(environment[key])

    for key in encrypted_user_data:
        env_map[key] = str(encrypted_user_data[key])

    return env_map


def _check_map_value(check_map: dict, key: str, default: Any) -> Any:
    value = check_map.get(key)
    if value in ("", {}, None, "{}"):
        log.get_logger().warning("%s is %s, using default value: %s", key, value, default)
        return default
    return value
