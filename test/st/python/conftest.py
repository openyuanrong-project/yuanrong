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

import os
import pytest
import yr


@pytest.fixture()
def init_yr():
    function_id = os.getenv("YR_PYTHON_FUNC_ID")
    server_address = os.getenv("YR_SERVER_ADDRESS")
    ds_address = os.getenv("YR_DS_ADDRESS")
    master_address = os.getenv("YR_MASTER_ADDRESS")
    log_dir = os.getenv("GLOG_log_dir")

    conf = yr.Config(
        function_id=function_id,
        server_address=server_address,
        ds_address=ds_address,
        log_level="DEBUG",
        log_dir=log_dir,
        master_addr_list=[master_address]
    )

    print("server_address: " + server_address)
    print("ds_address: " + ds_address)
    print("function_id: " + function_id)

    yr.init(conf)
    yield
    yr.finalize()
    print("-----test over-----")


@pytest.fixture()
def init_yr_config():
    function_id = os.getenv("YR_PYTHON_FUNC_ID")
    server_address = os.getenv("YR_SERVER_ADDRESS")
    ds_address = os.getenv("YR_DS_ADDRESS")
    log_dir = os.getenv("GLOG_log_dir")

    conf = yr.Config(
        function_id=function_id,
        server_address=server_address,
        ds_address=ds_address,
        log_level="DEBUG",
        log_dir=log_dir,
    )

    yield conf
    yr.finalize()
    print("-----test over-----")


@pytest.fixture()
def init_yr_with_local_mode():
    log_dir = os.getenv("GLOG_log_dir")
    conf = yr.Config(
        local_mode=True,
        log_level="DEBUG",
        log_dir=log_dir,
    )

    print("init local mode")

    yr.init(conf)
    yield
    yr.finalize()
    print("-----test over-----")
