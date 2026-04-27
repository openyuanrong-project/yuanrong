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

"""Local object store"""
import logging
import threading
from typing import List

from yr.common.singleton import Singleton

_logger = logging.getLogger(__name__)


@Singleton
class LocalObjectStore:
    """Local object store"""
    __object_map = {}
    __obj_cnt_map = {}
    __obj_future_map = {}
    __lock = threading.RLock()

    def clear(self):
        """
        clear
        :return: None
        """
        with self.__lock:
            self.__object_map.clear()
            self.__obj_cnt_map.clear()
            self.__obj_future_map.clear()

    def put(self, key, value):
        """put object"""
        self.__object_map[key] = value

    def get(self, key):
        """get object"""
        if isinstance(key, list):
            return [self.__object_map.get(k) for k in key]
        return self.__object_map.get(key)

    def set_return_obj(self, key, future):
        """
        set return obj
        :param key: object id
        :param future: future
        :return: None
        """
        self.__obj_future_map[key] = future

    def get_future(self, key):
        """
        get future
        :param key: object id
        :return: future
        """
        return self.__obj_future_map.get(key, None)

    def set_result(self, key, value):
        """
        set result
        :param key: object id
        :param value: value
        :return: None
        """
        self.__lock.acquire()
        self.__object_map[key] = value
        if key in self.__obj_future_map:
            future = self.__obj_future_map.get(key)
            self.__lock.release()
            future.set_result(value)
        else:
            self.__lock.release()
            _logger.debug("object %s not existed", key)

    def set_exception(self, key, exception):
        """
        set exception
        :param key: object id
        :param exception: exception
        :return: None
        """
        self.__lock.acquire()
        self.__object_map[key] = exception
        if key in self.__obj_future_map:
            future = self.__obj_future_map.get(key)
            self.__lock.release()
            future.set_exception(exception)
        else:
            self.__lock.release()
            _logger.debug("object %s not existed", key)

    def add_done_callback(self, key, callback):
        """
        add done callback
        :param key: object id
        :param callback: user callback
        :return: None
        """
        self.__lock.acquire()
        if key not in self.__obj_future_map:
            if key in self.__object_map:
                self.__lock.release()
                callback(self.__object_map[key])
                return
            self.__lock.release()
            raise RuntimeError("object not existing")
        future = self.__obj_future_map.get(key)
        self.__lock.release()
        future.add_done_callback(callback)

    def is_object_existing_in_local(self, key):
        """is object existing in local"""
        return key in self.__object_map

    def increase_global_reference(self, object_ids: List[str]) -> None:
        """
        increase global reference for object
        :param object_ids: objects
        :return: None
        """
        with self.__lock:
            for i in object_ids:
                if i in self.__obj_cnt_map:
                    self.__obj_cnt_map[i] += 1
                else:
                    self.__obj_cnt_map[i] = 0

    def decrease_global_reference(self, object_ids: List[str]) -> None:
        """
        decrease global reference for object
        :param object_ids: objects
        :return: None
        """
        with self.__lock:
            for i in object_ids:
                if i in self.__obj_cnt_map:
                    self.__obj_cnt_map[i] -= 1
                    if self.__obj_cnt_map[i] == 0:
                        self.__object_map.pop(i)
                        self.__obj_future_map.pop(i)
                        self.__obj_cnt_map.pop(i)
                else:
                    _logger.debug("object %s not existed in local store.", i)

    def release(self, key):
        """
        del key
        """
        if key in self.__object_map:
            del self.__object_map[key]
