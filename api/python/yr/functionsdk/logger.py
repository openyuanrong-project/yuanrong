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

"""faas logger"""

import gzip
import logging
import logging.handlers
import os
import shutil

from yr.log import CustomFilter, RuntimeLogger

LOG_DEFAULT_MAX_BYTES = 50 * 1024 * 1024
LOG_DEFAULT_BACKUP_COUNT = 3
# user function log default log level
# user function log handler
USER_FUNCTION_LOGGER = None


class RotatingFileHandler(logging.handlers.RotatingFileHandler):
    """
    RotatingFileHandler
    """
    def __init__(self, file_name, mode="a", max_bytes=0, backup_count=0, log_name=""):
        super(RotatingFileHandler, self).__init__(file_name, mode, max_bytes, backup_count, None, False)
        self.log_name = log_name
        self.backup_count = backup_count
        self.log_dir = os.path.dirname(log_name)

    def doRollover(self) -> None:
        """
        go rollover
        """
        super(RotatingFileHandler, self).doRollover()
        to_compress = []
        for f in os.listdir(self.log_dir):
            if f.startswith(self.log_name) and not f.endswith((".gz", ".log")):
                to_compress.append(os.path.join(self.log_dir, f))

        self._update_zip_file_name()

        for f in to_compress:
            if os.path.exists(f):
                with open(f, "rb") as _old, gzip.open(f + ".gz", "wb") as _new:
                    shutil.copyfileobj(_old, _new)
                os.remove(f)

    def _update_zip_file_name(self):
        zip_log_files = {}
        for f in os.listdir(self.log_dir):
            if f.startswith(self.log_name) and f.endswith(".gz"):
                f_split = f.split(".")
                if len(f_split) < 3:
                    os.remove(os.path.join(self.log_name, f))
                    continue
                try:
                    index = int(f_split[-2])
                except IndexError:
                    os.remove(os.path.join(self.log_dir, f))
                    continue
                finally:
                    pass

                if index >= self.backup_count:
                    os.remove(os.path.join(self.log_dir, f))
                    continue

                zip_log_files[index] = os.path.join(self.log_dir, f)

        for i in range(self.backup_count - 1, 0, -1):
            if i in zip_log_files:
                f_split = zip_log_files.get(i).split(".")
                f_split[-2] = str(i + 1)
                os.rename(zip_log_files.get(i), ".".join(f_split))

        del zip_log_files


def init_user_function_log(loglevel, logger=None):
    """ initialize user function log """
    global USER_FUNCTION_LOGGER
    if logger is not None:
        USER_FUNCTION_LOGGER = logger
    else:
        USER_FUNCTION_LOGGER = logging.getLogger("user-function")
        USER_FUNCTION_LOGGER.setLevel(loglevel)
        USER_FUNCTION_LOGGER.addFilter(CustomFilter())
        log_file_name = os.path.join(RuntimeLogger().get_runtime_log_location(), RuntimeLogger().get_runtime_id() +
                                     "-user-function.log")
        handler = RotatingFileHandler(file_name=log_file_name, mode="a", max_bytes=LOG_DEFAULT_MAX_BYTES,
                                      backup_count=LOG_DEFAULT_BACKUP_COUNT, log_name="user-function")

        formatter = logging.Formatter('{"Level":"%(levelname)s",'
                                      '"log":"%(message)s",'
                                      '"projectId":"%(tenant_id)s",'
                                      '"podName":"%(pod_name)s",'
                                      '"package":"%(package)s",'
                                      '"function":"%(function_name)s",'
                                      '"version":"%(version)s",'
                                      '"stream":"%(stream)s",'
                                      '"instanceId":"%(instance_id)s",'
                                      '"requestId":"%(request_id)s",'
                                      '"time":"%(asctime)s.%(msecs)03d",'
                                      '"stage":"%(stage)s",'
                                      '"status":"%(status)s",'
                                      '"finishLog":"%(finish_log)s"}', datefmt='%Y-%m-%dT%H:%M:%S')
        handler.setFormatter(formatter)
        USER_FUNCTION_LOGGER.addHandler(handler)


def get_user_function_logger(log_level: int = logging.DEBUG):
    """
    get user function log
    return user function log, singleton mode
    """
    global USER_FUNCTION_LOGGER
    if USER_FUNCTION_LOGGER is None:
        init_user_function_log(log_level)
    return USER_FUNCTION_LOGGER
