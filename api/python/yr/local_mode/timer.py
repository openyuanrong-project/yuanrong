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

"""global timer"""
import logging
import sched
import time
from threading import Thread
from typing import Callable

from yr.common.singleton import Singleton

_logger = logging.getLogger(__name__)


@Singleton
class Timer:
    """timer for task schedule"""
    __slots__ = ["__scheduler", "__is_running", "__thread", "__interval"]

    def __init__(self):
        self.__scheduler = sched.scheduler(time.time, time.sleep)
        self.__is_running = False
        self.__thread = None
        self.__interval = 0.1

    def init(self, interval=0.1):
        """start thread to scheduler timer"""
        if self.__is_running and self.__thread.is_alive():
            return
        self.__interval = interval
        self.__thread = Thread(target=self.run, name="YRTimer", daemon=True)
        self.__is_running = True
        self.__thread.start()

    def clear(self):
        """clear tasks"""
        if self.__scheduler.empty():
            return
        for event in self.__scheduler.queue:
            try:
                self.__scheduler.cancel(event)
            except ValueError:
                continue

    def stop(self):
        """stop timer thread"""
        if self.__is_running:
            self.clear()
            self.__is_running = False
            self.__thread.join()

    def run(self) -> None:
        """timer loop"""
        while self.__is_running:
            try:
                self.__scheduler.run(blocking=False)
            except Exception as e:
                _logger.exception(e)
            time.sleep(self.__interval)

    def after(self, delay: float, action: Callable, *args, **kwargs) -> None:
        """run a task after some time"""
        self.__scheduler.enter(delay, 0, action, argument=args, kwargs=kwargs)
