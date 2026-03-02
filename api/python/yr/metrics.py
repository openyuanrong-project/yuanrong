# coding=utf-8
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

import re

from dataclasses import field
from typing import Dict

from yr.base_runtime import AlarmInfo
from yr.config_manager import ConfigManager
from yr.runtime_holder import global_runtime
from yr.common.utils import GaugeData, UInt64CounterData, DoubleCounterData

_METRIC_NAME_RE = re.compile(r'^[a-zA-Z_:][a-zA-Z0-9_:]*$')
_METRIC_LABEL_NAME_RE = re.compile(r'^[a-zA-Z_][a-zA-Z0-9_]*$')
_RESERVED_METRIC_LABEL_NAME_RE = re.compile(r'^__.*$')


class Metrics:
    @classmethod
    def _check_name(cls, name: str):
        if len(name) == 0:
            raise ValueError("invalid metric name, should not be empty")
        if not _METRIC_NAME_RE.match(name):
            raise ValueError(f'invalid metric name: {name}')

    @classmethod
    def _check_label(cls, labels: dict):
        if not isinstance(labels, dict):
            raise ValueError(f'invalid metrics label, {type(labels)} not dict type')
        for key, value in labels.items():
            if not _METRIC_LABEL_NAME_RE.match(key):
                raise ValueError(f'invalid metrics label name: {key}')
            if _RESERVED_METRIC_LABEL_NAME_RE.match(key):
                raise ValueError(f'reserved metrics label name: {key}')
            if not isinstance(value, str):
                raise ValueError(f'invalid metrics label value: {value}, {type(value)} not str')


class Gauge(Metrics):
    """
    Used for reporting metrics data.

    Note:
        1. Prometheus-like data structure.
        2. Billing information is reported to the backend openYuanRong collector.
        3. Not usable within the driver.

    Args:
        name (str): name.
        description (str): description.
        unit (str, optional): Unit.
        labels (Dict[str, str], optional): Label. e.g. ``labels = {"request_id": "abc"}``.

    """
    def __init__(self,
                 name: str,
                 description: str,
                 unit: str = '',
                 labels: Dict[str, str] = {},
                 ):
        self.__name = name
        self.__description = description
        self.__unit = unit
        self.__labels = labels
        self._check_name(name)
        self._check_label(labels)

    def add_labels(self, labels: Dict[str, str]) -> None:
        """
        add label.

        Args:
            labels (Dict[str,str]): Both labels and key-value pairs must be strings,
            and keys cannot begin with special characters like `*`.

        Raises:
            ValueError: Labels are missing, or the data does not conform to Prometheus standard requirements.

        Example:
            >>> import yr
            >>> config = yr.Config(enable_metrics=True)
            >>> yr.init(config)

            >>> @yr.instance
            ... class GaugeActor:
            ...     def __init__(self):
            ...         self.labels = {"key1": "value1"} 
            ...         self.gauge = yr.Gauge(
            ...             name="test",
            ...             description="",
            ...             unit="ms"
            ...         )
            ...         self.gauge.add_labels(self.labels)
            ...         self.gauge.set(50)
            ...         print("Initial labels:", self.labels)
            ...
            ...     def set_value(self, value):
            ...         self.gauge.set(value)
            ...         return {
            ...             "value": value,
            ...             "labels": self.labels
            ...         }
            >>>
            >>> actor = GaugeActor.options(name="gauge_actor").invoke()
            >>> result = actor.set_value.invoke(75)
            >>> print("Driver got:", yr.get(result))
            
        """
        self._check_label(labels)
        if len(labels) == 0:
            raise ValueError("invalid metrics labels, should not be empty")
        old = self.__labels
        self.__labels = {**old, **labels}

    def set(self, value: float) -> None:
        """
        set value.

        Args:
            value (float): Target value.

        Raises:
            ValueError: Invoked in the driver or fails to meet the data type requirements
                (such as reporting data types other than float).

        Example:
            >>> import yr
            >>> config = yr.Config(enable_metrics=True)
            >>> yr.init(config)
            >>>
            >>> @yr.instance
            ... class GaugeActor:
            ...     def __init__(self):
            ...         self.labels = {"key1": "value1"}
            ...         self.gauge = yr.Gauge(
            ...             name="test",
            ...             description="",
            ...             unit="ms",
            ...             labels=self.labels
            ...         )
            ...         self.gauge.set(50)
            ...
            ...         print("Initial labels:", self.labels)
            ...
            ...     def set_value(self, value):
            ...         self.gauge.set(value)
            ...         return {
            ...             "value": value,
            ...             "labels": self.labels
            ...         }
            >>>
            >>> actor = GaugeActor.options(name="gauge_actor").invoke()
            >>>
            >>> result = actor.set_value.invoke(75)
            >>> print(yr.get(result))
        """
        fvalue = float(value)
        if ConfigManager().is_driver:
            raise ValueError("gauge metrics report not support in driver")
        data = GaugeData(name=self.__name, description=self.__description,
                         unit=self.__unit, labels=self.__labels, value=fvalue)
        global_runtime.get_runtime().report_gauge(data)


class UInt64Counter(Metrics):
    """
    A class representing a 64-bit unsigned integer counter for metrics.

    Args:
        name (str): Name of the counter.
        description (str): Description of the counter.
        unit (str): Unit of the counter.
        labels (Dict[str, str]): Optional labels for the counter.

    """
    def __init__(self,
                 name: str,
                 description: str,
                 unit: str,
                 labels: Dict[str, str] = field(default_factory=dict),
                 ):
        self.__uint_counter_name = name
        self.__description = description
        self.__unit = unit
        self.__uint_counter_labels = labels
        self._check_name(name)
        self._check_label(labels)

    def add_labels(self, labels: dict) -> None:
        """
        add lable for metrics data.

        Args:
            labels (dict): A dictionary of labels where both keys and values must be strings.

        Raises:
            ValueError: If the label is empty or does not meet the data type requirements.

        Example:
            >>> import yr
            >>>
            >>> config = yr.Config(enable_metrics=True)
            >>> yr.init(config)
            >>>
            >>> @yr.instance
            >>> class Actor:
            ...     def __init__(self):
            ...         labels = {"key1": "value1", "key2": "value2"}
            ...         self.data = yr.UInt64Counter(
            ...             name="test",
            ...             description="",
            ...             unit="ms",
            ...             labels=labels
            ...         )
            ...         self.data.add_labels({"key3": "value3"})
            ...         print("Actor init done")
            ...
            ...     def run(self):
            ...         self.data.set(5)
            ...         self.data.add_labels({"phase": "run"})
            ...         msg = (
            ...             f"Actor run: {self.data._UInt64Counter__uint_counter_labels},
            ...             f"value: {self.data.get_value()}"
            ...         )
            ...         print(msg)
            ...         return msg
            >>>
            >>> actor1 = Actor.options(name="actor").invoke()
            >>> print(yr.get(actor1.run.invoke()))

        """
        self._check_label(labels)
        if len(labels) == 0:
            raise ValueError("invalid metrics labels, should not be empty")
        old = self.__uint_counter_labels
        self.__uint_counter_labels = {**old, **labels}

    def set(self, value: int) -> None:
        """
        Set uint64 counter to the given value.

        Args:
            value (int): The value to be set.

        Raises:
            ValueError: Invoked in the driver.

        Example:
            >>> import yr
            >>> config = yr.Config(enable_metrics=True)
            >>> yr.init(config)
            >>>
            >>> @yr.instance
            ... class MyActor:
            ...     def __init__(self):
            ...         self.data = yr.UInt64Counter(
            ...             "userFuncTime",
            ...             "user function cost time",
            ...             "ms",
            ...             {"runtime": "runtime1"}
            ...         )
            ...         self.data.add_labels({"stage": "init"})
            ...         self.data.set(100)
            ...
            ...     def add(self, n=1, labels=None):
            ...         current = self.data.get_value()
            ...         self.data.set(current + n)
            ...         if labels:
            ...             self.data.add_labels(labels)
            ...         return self.get()
            ...
            ...     def get(self):
            ...         return {
            ...             "labels": self.data._UInt64Counter__uint_counter_labels,
            ...             "value": self.data.get_value()
            ...         }
            >>> actor = MyActor.options(name="actor").invoke()
            >>> result = actor.add.invoke(5)
            >>> print(yr.get(result))
        """
        ivalue = int(value)
        if ConfigManager().is_driver:
            raise ValueError("uint64 counter metrics set not support in driver")
        data = UInt64CounterData(name=self.__uint_counter_name, description=self.__description,
                                 unit=self.__unit, labels=self.__uint_counter_labels, value=ivalue)
        global_runtime.get_runtime().set_uint64_counter(data)

    def reset(self) -> None:
        """Reset uint64 counter."""
        if ConfigManager().is_driver:
            raise ValueError("uint64 counter metrics reset not support in driver")
        data = UInt64CounterData(name=self.__uint_counter_name, description=self.__description,
                                 unit=self.__unit, labels=self.__uint_counter_labels)
        global_runtime.get_runtime().reset_uint64_counter(data)

    def increase(self, value: int) -> None:
        """Increase uint64 counter to the given value."""
        ivalue = int(value)
        if ConfigManager().is_driver:
            raise ValueError("uint64 counter metrics increase not support in driver")
        data = UInt64CounterData(name=self.__uint_counter_name, description=self.__description,
                                 unit=self.__unit, labels=self.__uint_counter_labels, value=ivalue)
        global_runtime.get_runtime().increase_uint64_counter(data)

    def get_value(self) -> int:
        """Get value of uint64 counter."""
        if ConfigManager().is_driver:
            raise ValueError("uint64 counter metrics get value not support in driver")
        data = UInt64CounterData(name=self.__uint_counter_name, description=self.__description,
                                 unit=self.__unit, labels=self.__uint_counter_labels)
        return global_runtime.get_runtime().get_value_uint64_counter(data)


class DoubleCounter(Metrics):
    """
    Initialize the double counter.

    Args:
        name (str): The name of the counter.
        description (str): The description of the counter.
        unit (str): unit.
        labels (Dict[str, str], optional): Labels for the counter, defaults to an empty dictionary.
    """
    def __init__(self,
                 name: str,
                 description: str,
                 unit: str,
                 labels: Dict[str, str] = field(default_factory=dict),
                 ):
        self.__double_counter_name = name
        self.__description = description
        self.__unit = unit
        self.__double_counter_labels = labels
        self._check_name(name)
        self._check_label(labels)

    def add_labels(self, labels: dict) -> None:
        """
        add lable for metrics data.

        Args:
            labels (dict): The dictionary of labels to add.

        Raises:
            ValueError: If the labels are empty.

        Example:
            >>> import yr
            >>>
            >>> config = yr.Config(enable_metrics=True)
            >>> yr.init(config)
            >>>
            >>> @yr.instance
            >>> class Actor:
            >>>     def __init__(self):
            >>>         self.data = yr.DoubleCounter(
            >>>             "userFuncTime",
            >>>             "user function cost time",
            >>>             "ms",
            >>>             {"runtime": "runtime1"}
            >>>         )
            >>>         self.data.add_labels({"stage": "init"})
            >>>         print("Actor init:", self.data._DoubleCounter__double_counter_labels)
            >>>
            >>>     def run(self):
            >>>         self.data.set(5)
            >>>         self.data.add_labels({"phase": "run"})
            >>>         msg = (
            ...             f"Actor run: {self.data._UInt64Counter__uint_counter_labels}, "
            ...             f"value: {self.data.get_value()}"
            ...         )
            >>>         print(msg)
            >>>         return msg 
            >>> actor1 = Actor.options(name="actor").invoke()
            >>> result = actor1.run.invoke()
            >>> print("run result:", yr.get(result))
        """
        self._check_label(labels)
        if len(labels) == 0:
            raise ValueError("invalid metrics labels, should not be empty")
        old = self.__double_counter_labels
        self.__double_counter_labels = {**old, **labels}

    def set(self, value: float) -> None:
        """
        Set double counter to the given value.

        Args:
            value (float): The value to set.

        Raises:
            ValueError: Invoked in the driver.

        Example:
            >>> import yr
            >>> config = yr.Config(enable_metrics=True)
            >>> yr.init(config)
            >>>
            >>> @yr.instance
            >>> class Actor:
            >>>     def __init__(self):
            >>>         try:
            >>>             self.data = yr.DoubleCounter(
            >>>                 "userFuncTime",
            >>>                 "user function cost time",
            >>>                 "ms",
            >>>                 {"runtime": "runtime1"}
            >>>            )
            >>>         except Exception as err:
            >>>             print(f"error: {err}")
            >>>
            >>>    def run(self, *args, **kwargs):
            >>>        try:
            >>>            self.data.set(1)
            >>>        except Exception as err:
            >>>            print(f"error: {err}")
        """
        fvalue = float(value)
        if ConfigManager().is_driver:
            raise ValueError("double counter metrics set not support in driver")
        data = DoubleCounterData(name=self.__double_counter_name, description=self.__description,
                                 unit=self.__unit, labels=self.__double_counter_labels, value=fvalue)
        global_runtime.get_runtime().set_double_counter(data)

    def reset(self) -> None:
        """Reset double counter."""
        if ConfigManager().is_driver:
            raise ValueError("double counter metrics reset not support in driver")
        data = DoubleCounterData(name=self.__double_counter_name, description=self.__description,
                                 unit=self.__unit, labels=self.__double_counter_labels)
        global_runtime.get_runtime().reset_double_counter(data)

    def increase(self, value: float) -> None:
        """Increase double counter to the given value."""
        fvalue = float(value)
        if ConfigManager().is_driver:
            raise ValueError("double counter metrics increase not support in driver")
        data = DoubleCounterData(name=self.__double_counter_name, description=self.__description,
                                 unit=self.__unit, labels=self.__double_counter_labels, value=fvalue)
        global_runtime.get_runtime().increase_double_counter(data)

    def get_value(self) -> float:
        """Get value of double counter."""
        if ConfigManager().is_driver:
            raise ValueError("double counter metrics get value not support in driver")
        data = DoubleCounterData(name=self.__double_counter_name, description=self.__description,
                                 unit=self.__unit, labels=self.__double_counter_labels)
        return global_runtime.get_runtime().get_value_double_counter(data)


class Alarm(Metrics):
    """
    Used to set and manage alarm information.

    Args:
        name (str): The name of the alarm.
        description (str): Description of the alarm.
    """
    def __init__(self,
                 name: str,
                 description: str,
                 ):
        self.__name = name
        self.__description = description
        self._check_name(name)

    def set(self, alarm_info: AlarmInfo) -> None:
        """
        Set alarm to the given info.

        Args:
            alarm_info (AlarmInfo): An object containing detailed alarm information.

        Raises:
            ValueError: Invoked in the driver.
            ValueError: If alarm_name is None.

        Example:
            >>> import yr
            >>> import time
            >>> config = yr.Config(enable_metrics=True)
            >>> config.log_level="DEBUG":378
            >>> yr.init(config)

            >>> @yr.instance
            >>> class Actor:
            >>>    def __init__(self):
            >>>        self.state = 0
            >>>        self.name = "aa"

            >>>    def add(self, value):
            >>>        self.state += value
            >>>        if self.state > 10:
            >>>            alarm_info = yr.AlarmInfo(alarm_name="aad")
            >>>            alarm = yr.Alarm(self.name, description="")
            >>>            alarm.set(alarm_info)
            >>>        return self.state
            >>>
            >>>    def get(self):
            >>>        return self.state
            >>>
            >>> actor1 = Actor.options(name="actor1").invoke()
            >>>
            >>> print("actor1 add 5:", yr.get(actor1.add.invoke(5)))
            >>> print("actor1 add 7:", yr.get(actor1.add.invoke(7)))
            >>> print("actor1 state:", yr.get(actor1.get.invoke()))
        """
        if ConfigManager().is_driver:
            raise ValueError("alarm metrics set not support in driver")
        if len(alarm_info.alarm_name) == 0:
            raise ValueError("invalid alarm name, should not be empty")
        global_runtime.get_runtime().set_alarm(self.__name, self.__description, alarm_info)
