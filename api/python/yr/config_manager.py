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

"""config manager"""

import logging

from yr.common import utils
from yr.common.singleton import Singleton
from yr.config import Config, DeploymentConfig, MetaConfig, MetaFunctionID

_DEFAULT_CLUSTER_PORT = "31222"
_DEFAULT_IN_CLUSTER_CLUSTER_PORT = "21003"
_DEFAULT_DS_PORT = "31501"
_DEFAULT_DS_PORT_OUTER = "31222"
_DEFAULT_HTTP_IOC_THREADS_NUM = 200
_DEFAULT_RPC_TIMOUT = 30 * 60
_URN_LENGTH = 7


def _check_function_urn(is_driver, value):
    if value is None:
        value = ''
    if not is_driver and value == '':
        return False
    items = value.split(':')
    if len(items) != _URN_LENGTH:
        raise ValueError(f"invalid function id {value}")
    return True


@Singleton
class ConfigManager:
    """
    Config manager singleton

    Attributes:
        server_address: System cluster address.
        ds_address: DataSystem address.
        is_driver: only False when initialize in runtime
        log_level: yr api log level, default: WARNING
    """

    def __init__(self):
        self.__server_address = ""
        self.__ds_address = ""
        self.__connection_nums = None
        self.__log_level = logging.WARNING
        self.__in_cluster = True
        self.__deployment_config = DeploymentConfig()
        self.tls_config = None
        self.meta_config = None
        self.log_file_num_max = 0
        self.log_flush_interval = 5
        self.__is_driver = True
        self.runtime_id = None
        self.invoke_timeout = None
        self.local_mode = False
        self.rt_server_address = ""
        self.log_dir = "./"
        self.log_file_size_max = 0
        self.load_paths = []
        self.custom_envs = {}
        self.rpc_timeout = _DEFAULT_RPC_TIMOUT
        self.tenant_id = ""
        self.enable_mtls = False
        self.private_key_path = ""
        self.certificate_file_path = ""
        self.verify_file_path = ""
        self.private_key_paaswd = ""
        self.http_ioc_threads_num = _DEFAULT_HTTP_IOC_THREADS_NUM
        self.server_name = ""
        self.ns = ""
        self.enable_metrics = True
        self.master_add_list = []
        self.working_dir = ""
        self.runtime_public_key_path = ""
        self.runtime_private_key_path = ""
        self.ds_public_key_path = ""
        self.enable_ds_encrypt = False
        self._num_cpus = 0
        self.runtime_env = ""
        self.namespace = ""
        self.log_to_driver = False
        self.dedup_logs = False

    @property
    def deployment_config(self) -> DeploymentConfig:
        """ when auto=True needed, use to define deployment detail. """
        return self.__deployment_config

    @property
    def in_cluster(self) -> bool:
        """if True will use datasystem in cluster client"""
        return self.__in_cluster

    @property
    def log_level(self):
        """
        YR api log level
        """
        return self.__log_level

    @log_level.setter
    def log_level(self, value):
        """
        YR api log level
        """
        if isinstance(value, str):
            value = value.upper()
        self.__log_level = value

    @property
    def server_address(self):
        """
        Get server address
        """
        return self.__server_address

    @server_address.setter
    def server_address(self, value: str):
        """
        Set server address

        Args:
            value (str): System cluster ip
        """
        if not self.__is_driver:
            return
        if utils.validate_ip(value):
            if self.__in_cluster:
                self.__server_address = value + ":" + _DEFAULT_IN_CLUSTER_CLUSTER_PORT
            else:
                self.__server_address = value + ":" + _DEFAULT_CLUSTER_PORT
            return
        _, _ = utils.validate_address(value)
        self.__server_address = value

    @property
    def ds_address(self):
        """
        Get datasystem address
        """
        return self.__ds_address

    @ds_address.setter
    def ds_address(self, value: str):
        """
        Set datasystem address

        Args:
            value (str): Datasystem worker address : <ip>:<port> or <ip>,
                default port : 31501
        """
        if utils.validate_ip(value):
            if not self.__is_driver or self.__in_cluster:
                self.__ds_address = value + ":" + _DEFAULT_DS_PORT
            else:
                self.__ds_address = value + ":" + _DEFAULT_DS_PORT_OUTER
            return
        _, _ = utils.validate_address(value)
        self.__ds_address = value

    @property
    def connection_nums(self):
        """
        Get connection_nums
        """
        return self.__connection_nums

    @connection_nums.setter
    def connection_nums(self, value: int):
        """
        Set connection_nums

        Args:
            value (int): max connection number
        """
        if not isinstance(value, int):
            raise TypeError(
                f"connection_nums {type(value)} type error, 'int' is expected.")
        if (value >= 1) is False:
            raise ValueError(
                f"invalid connection_nums value, expect connection_nums >= 1, actual {value}")

        self.__connection_nums = value

    @property
    def job_id(self):
        """
        Job id
        """
        return self.meta_config.jobID

    @property
    def is_driver(self) -> bool:
        """
        Return a boolean value indicating whether the instance is a driver or not.

        Returns:
            bool: True if the instance is a driver, False otherwise.
        """
        return self.__is_driver

    @property
    def num_cpus(self):
        """"""
        return self._num_cpus

    @num_cpus.setter
    def num_cpus(self, value):
        """"""
        self._num_cpus = value

    def init(self, conf: Config, is_init=False):
        """
        Init the ConfigManager

        Args:
            :param conf: The yr api config which set by user.
            :param is_init: init state
        """
        job_id = conf.job_id if conf.job_id != "" else utils.generate_job_id()
        if not conf.local_mode:
            _ = _check_function_urn(conf.is_driver, conf.function_id)
            if conf.cpp_function_id != "":
                _ = _check_function_urn(conf.is_driver, conf.cpp_function_id)
        function_id = MetaFunctionID(
            python=utils.get_function_from_urn(conf.function_id),
            cpp=utils.get_function_from_urn(conf.cpp_function_id))

        self.meta_config = MetaConfig(
            jobID=job_id,
            codePath=conf.code_dir,
            recycleTime=conf.recycle_time,
            functionID=function_id,
            enableMetrics=conf.enable_metrics,
            functionMasters=conf.master_addr_list,
            maxTaskInstanceNum=conf.max_task_instance_num)

        self.log_level = conf.log_level
        if is_init:
            # If the init method has been executed,
            # log_level and function_id are updated.
            return
        self.__deployment_config = conf.deployment_config
        self.tls_config = conf.tls_config
        self.connection_nums = conf.connection_nums
        self.invoke_timeout = conf.invoke_timeout
        self.__in_cluster = conf.in_cluster

        utils.set_job_id(self.job_id)
        self.runtime_id = conf.runtime_id
        self.local_mode = conf.local_mode
        if self.local_mode:
            return
        self.__is_driver = conf.is_driver
        self.server_address = conf.server_address
        if self.__in_cluster:
            self.ds_address = conf.ds_address
        self.rt_server_address = conf.rt_server_address
        self.log_dir = conf.log_dir
        self.log_file_size_max = conf.log_file_size_max
        self.log_file_num_max = conf.log_file_num_max
        self.log_flush_interval = conf.log_flush_interval
        self.load_paths = conf.load_paths
        self.custom_envs = conf.custom_envs
        self.rpc_timeout = conf.rpc_timeout
        self.tenant_id = conf.tenant_id
        self.enable_mtls = conf.enable_mtls
        self.private_key_path = conf.private_key_path
        self.private_key_paaswd = conf.private_key_paaswd
        conf.private_key_paaswd = ""
        self.certificate_file_path = conf.certificate_file_path
        self.verify_file_path = conf.verify_file_path
        self.http_ioc_threads_num = conf.http_ioc_threads_num
        self.server_name = conf.server_name
        self.ns = conf.ns
        self.working_dir = conf.working_dir
        self.enable_ds_encrypt = conf.enable_ds_encrypt
        self.ds_public_key_path = conf.ds_public_key_path
        self.runtime_public_key_path = conf.runtime_public_key_path
        self.runtime_private_key_path = conf.runtime_private_key_path
        self.num_cpus = conf.num_cpus
        self.runtime_env = conf.runtime_env
        self.log_to_driver = conf.log_to_driver
        self.dedup_logs = conf.dedup_logs

    def get_function_id_by_language(self, language):
        """
        Get function id by language from function id dict

        Args:
            language (str): The language of target function or class
        """
        return self.__all_function_id.get(language)
