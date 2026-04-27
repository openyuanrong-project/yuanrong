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

"""log"""

import json
import logging
import os
import socket
import stat
import sys
import time
from logging import Logger
import logging.config

from yr.common.singleton import Singleton

# MAX_ROW_SIZE max row size of a log
_MAX_ROW_SIZE = 1024 * 1024
# python runtime log location
_BASE_LOG_NAME = "yr"
_LOG_SUFFIX = "_runtime.log"


class CustomFilter(logging.Filterer):
    """
    CustomFilter custom filter
    """

    def filter(self, record):
        if len(record.msg) > _MAX_ROW_SIZE:
            record.msg = record.msg[:_MAX_ROW_SIZE]
        return True


@Singleton
class RuntimeLogger:
    """runtime logger"""

    def __init__(self) -> None:
        self.__logger: Logger = None
        self.__runtime_id = ""
        self.__runtime_log_location = ""

    def get_runtime_id(self):
        """get runtime id"""
        return self.__runtime_id

    def get_runtime_log_location(self):
        """get runtime log location"""
        return self.__runtime_log_location

    def init(self, is_driver: bool, runtime_id: str, log_level: str) -> None:
        """Initializes logger for on cloud logging or local logging."""
        if self.__logger is not None:
            return

        # Check if YR_ONLY_STDOUT is set to enable stdout-only logging
        only_stdout = os.getenv("YR_ONLY_STDOUT", "false").lower() in ("true", "1")

        if only_stdout or is_driver:
            self.__runtime_id = runtime_id
            self.__init_stream_logger(log_level)
        else:
            self.__runtime_id = runtime_id
            self.__init_file_logger()

    def get_logger(self) -> Logger:
        """get logger"""
        if self.__logger is None:
            self.__init_file_logger()
            self.__logger.warning(
                "Logger is not initialized. Using file logger without runtime ID."
            )
        return self.__logger

    def __init_file_logger(self) -> None:
        # logging will collect processes information by default. Set these values
        # to False can improve performance
        logging.logMultiprocessing = False
        logging.Formatter.default_msec_format = "%s.%03d"
        logging.Formatter.converter = time.localtime
        config = {
            "version": 1,
            "disable_existing_loggers": False,
            "formatters": {
                "extra": {
                    "format": (
                        "[%(asctime)s %(levelname)s %(filename)s:%(lineno)d] "
                        "[%(podname)s %(thread)d] [%(runtime_id)s] %(message)s"
                    )
                }
            },
            "handlers": {
                "file": {
                    "class": "logging.FileHandler",
                    "filename": "/home/snuser/log",
                    "formatter": "extra",
                }
            },
            "loggers": {"FileLogger": {"handlers": ["file"], "level": "INFO"}},
        }

        log_file_name = self.__adapt_log_file_name(config)
        # Call the 'dictConfig' to create a file before
        # performing the 'os.chmod' operation. Otherwise an error
        # would be raised.
        logging.config.dictConfig(config)
        os.chmod(log_file_name, stat.S_IWUSR | stat.S_IRUSR | stat.S_IRGRP)

        self.__logger = logging.getLogger("FileLogger")
        self.__logger.addFilter(CustomFilter())
        self.__logger = logging.LoggerAdapter(
            self.__logger,
            {"podname": socket.gethostname(), "runtime_id": self.__runtime_id},
        )

    def __init_stream_logger(self, log_level: str) -> None:
        self.__logger = logging.getLogger(_BASE_LOG_NAME)
        handler = logging.StreamHandler(sys.stdout)
        fmt = logging.Formatter(
            fmt="[%(asctime)s.%(msecs)03d %(levelname)s %(funcName)s %(filename)s:%(lineno)d %(thread)d] %(message)s",
            datefmt="%Y-%m-%d %H:%M:%S",
        )
        handler.setFormatter(fmt)
        self.__logger.setLevel(log_level)
        self.__logger.handlers.clear()
        self.__logger.addHandler(handler)
        self.__logger.propagate = False

    def __adapt_log_file_name(self, config: dict) -> str:
        log_file_name = config["handlers"]["file"]["filename"]
        if os.getenv("GLOG_log_dir") is not None:
            log_file_name = os.getenv("GLOG_log_dir")
        os.environ["DATASYSTEM_CLIENT_LOG_DIR"] = log_file_name
        self.__runtime_log_location = log_file_name
        os.makedirs(log_file_name, exist_ok=True)
        log_id = os.environ.get("YR_LOG_PREFIX", "")
        if len(log_id) != 0:
            log_file_name = os.path.join(log_file_name, log_id + _LOG_SUFFIX)
        else:
            log_file_name = os.path.join(log_file_name, self.__runtime_id + _LOG_SUFFIX)
        config["handlers"]["file"]["filename"] = log_file_name
        return log_file_name


def init_logger(on_cloud: bool, runtime_id: str = "", log_level: str = "DEBUG") -> None:
    """init log handler"""
    RuntimeLogger().init(on_cloud, runtime_id, log_level)


def get_logger() -> Logger:
    """Gets the runtime logger with basic config.

    Returns:
        Logger: The runtime logger
    """
    return RuntimeLogger().get_logger()
