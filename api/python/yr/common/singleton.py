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

"""singleton"""

import threading


class Singleton:
    """
    Used as metaclass to implement Singleton.
    """

    def __init__(self, cls):
        self._cls = cls
        self._instance = {}
        self.lock = threading.Lock()

    def __call__(self, *args, **kw):
        if self._cls not in self._instance:
            with self.lock:
                if self._cls not in self._instance:
                    self._instance[self._cls] = self._cls(*args, **kw)
        return self._instance.get(self._cls)
