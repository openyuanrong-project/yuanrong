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

"""Scheduler"""

from abc import abstractmethod, ABC
from typing import Optional, Iterable, List

from yr.local_mode.instance import Instance
from yr.local_mode.task_spec import TaskSpec


class Scorer(ABC):
    """abstract Scorer for scoring instances"""

    @staticmethod
    @abstractmethod
    def score(task: TaskSpec, instance: Instance) -> int:
        """scoring instance"""


class ConcurrencyScorer(Scorer):
    """Scorer for concurrency"""

    @staticmethod
    def score(task: TaskSpec, instance: Instance) -> int:
        """scoring instance"""
        if instance.task_count < instance.resource.concurrency:
            return 1
        return 0


class Scheduler(ABC):
    """abstract Scheduler"""

    @abstractmethod
    def schedule(self, task: TaskSpec, instances: Iterable[Instance]) -> Optional[Instance]:
        """select a instance"""


class NormalScheduler(Scheduler):
    """base scheduler implementation"""
    __slots__ = ["__scorers"]

    def __init__(self, scorers: List[Scorer] = None):
        if scorers is None:
            self.__scorers = []
        else:
            self.__scorers = scorers

    def schedule(self, task: TaskSpec, instances: Iterable[Instance]) -> Optional[Instance]:
        """select a instance"""
        max_score = 0
        max_instance = None
        for instance in instances:
            if not instance.is_activate:
                continue
            score = 0
            for scorer in self.__scorers:
                score += scorer.score(task, instance)
            if score > max_score:
                max_score = score
                max_instance = instance
        return max_instance
