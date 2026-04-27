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

"""faas logger manager"""

import logging
import socket
import threading
from queue import Queue

from yr.common.singleton import Singleton
from yr.functionsdk import logger


class Log:
    """ base log entity """

    def __init__(self, loglevel, msg):
        self.loglevel = loglevel
        self.msg = msg


class FaasLogger:
    """ The FaasLogger class provides a way to log messages with different levels,
    It enters logs into a queue so as to implement asynchronous printing. 'UserLogManager'
    will use global 'USER_FUNCTION_LOG' to write the log got from this queue.
    """

    def __init__(self):
        """
        log_level: should be the same as defined in logging.
            debug, info, warning, error, critical = logging.DEBUG, logging.INFO, ...
        """
        self.queue = None

    def set_queue(self, queue: Queue):
        """ set a queue """
        self.queue = queue

    def log(self, level, msg):
        """receive log message and put into the queue"""
        log = Log(level, msg)
        self.queue.put(log)

    def debug(self, msg: str):
        """log with level debug"""
        self.log(logging.DEBUG, msg)

    def info(self, msg: str):
        """log with level debug"""
        self.log(logging.INFO, msg)

    def warning(self, msg: str):
        """log with level debug"""
        self.log(logging.WARNING, msg)

    def error(self, msg: str):
        """log with level debug"""
        self.log(logging.ERROR, msg)


@Singleton
class UserLogManager:
    """
    UserLogManager provide user a logger
    It needs a queue and a user function logger
    It has a loop, receiving user log, running till shutdown() called
    """

    def __init__(self):
        self._user_function_log = None
        self.log_level = logging.INFO
        self.tenant_id = ""
        self.function_name = ""
        self.version = ""
        self.package = ""
        self.stream = ""
        self.instance_id = ""
        self.request_id = ""
        self.stage = ""
        self.finish_log = ""
        self.status = ""
        self.queue = None
        self.__running = True

    def start_user_log(self):
        """ start user function log with a new thread """
        self.__running = True
        t = threading.Thread(target=self.run, name="user_log_thread", daemon=True)
        t.start()

    def run(self) -> None:
        """ a loop runs till user function finished """
        self.insert_start_log()
        while self.__running:
            ret_log = self.queue.get()
            if ret_log is None:
                break
            self._write_log(ret_log.loglevel, ret_log.msg)
        while not self.queue.empty():
            ret_log = self.queue.get()
            if ret_log is None:
                break
            self._write_log(ret_log.loglevel, ret_log.msg)
        self.insert_end_log()

    def set_log(self, log):
        """ set a user function log """
        self._user_function_log = log

    def load_logger_config(self, cfg):
        """ load logger config when function initializing """
        self.log_level = cfg['log_level']
        self.tenant_id = cfg['tenant_id']
        self.function_name = cfg['function_name']
        self.version = cfg['version']
        self.package = cfg['package']
        self.stream = cfg['stream']
        self.instance_id = cfg['instance_id']
        self._user_function_log = logger.get_user_function_logger(cfg['log_level'])

    def set_stage(self, stage: str):
        """
        set stage
        """
        self.stage = stage

    def insert_start_log(self):
        """ insert a first log """
        self.status = "success"
        self.finish_log = "false"
        self._write_log(self.log_level, f"@@Start {self.stage} Reqeust")

    def insert_end_log(self, status="success"):
        """ insert a last log """
        self.status = status
        self.finish_log = "true"
        self._write_log(self.log_level, f"@@End {self.stage} Reqeust")

    def shutdown(self):
        """ stop the loop when user function finished """
        self.__running = False
        if self.queue:
            self.queue.put(None)

    def register_logger(self, log: FaasLogger):
        """
        register logger
        """
        self.queue = log.queue

    def _write_log(self, loglevel, message):
        extra_info = {
            "tenant_id": self.tenant_id,
            "pod_name": {'podname': socket.gethostname()},
            "package": self.package,
            "function_name": self.function_name,
            "version": self.version,
            "stream": self.stream,
            "instance_id": self.instance_id,
            "request_id": self.request_id,
            "stage": self.stage,
            "status": self.status,
            "finish_log": self.finish_log,
        }
        self._user_function_log.log(loglevel, message, extra=extra_info)
