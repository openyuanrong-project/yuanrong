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

"""stream"""

from dataclasses import dataclass, field
from enum import Enum
from typing import Dict, Union


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class ProducerConfig:
    """
    The configuration class created by the producer.
    """

    #: After Send, Flush will be triggered after a delay up to the specified duration.
    #: < 0: do not auto flush,
    #: = 0: flush immediately,
    #: > 0: delay duration in milliseconds before flushing.
    #: Default value is 5.
    delay_flush_time: int = 5
    #: Represents the buffer page size for the producer, in bytes (B). A flush is triggered when a page is full.
    #: The value must be greater than 0 and a multiple of 4 KB.
    #: Default is ``1`` MB (``1024 * 1024``).
    page_size: int = 1024 * 1024
    #: Specifies the maximum shared memory size that a stream can use on a worker, in bytes (B).
    #: The default is ``1`` GB ( ``1024 * 1024 * 1024``),
    #: and the valid range is [64 KB, size of the worker's shared memory].
    max_stream_size: int = 1024 * 1024 * 1024
    #: Specifies whether the stream enables the auto-cleanup feature. Default is ``false``.
    auto_clean_up: bool = False
    #: Specifies whether content encryption is enabled for the stream. Default is ``false`` (disabled).
    encrypt_stream: bool = False
    #: The data sent by the producer will be retained until received by the Nth consumer.
    #: The default value is ``0``, meaning that if there are no consumers when the producer sends data,
    #: the data will not be retained and may be missed when consumers are created later.
    retain_for_num_consumers: int = 0
    #: Represents the reserved memory size in bytes (B).
    #: When creating a producer, an attempt will be made to reserve ``reserve_size`` bytes of memory.
    #: If the reservation fails, creating the producer will raise an exception.
    #: ``reserve_size`` must be an integer multiple of ``page_size`` and within the range ``[0, max_stream_size]``.
    #: If ``reserve_size`` is ``0``, it will be set to `page_size`.
    #: Default value is ``0``.
    reserve_size: int = 0
    #: Extended configuration stored as a dictionary, allowing users to customize configuration items.
    #: Default value is an empty dictionary.
    extend_config: Dict[str, str] = field(default_factory=dict)


class SubscriptionType(Enum):
    """
    SubscriptionType

    Attributes:
        STREAM: default mode.
        ROUND_ROBIN: not support
        KEY_PARTITIONS: not support.
    """
    STREAM = 0
    ROUND_ROBIN = 1
    KEY_PARTITIONS = 2


@dataclass(init=True, repr=False, eq=False, order=False, unsafe_hash=False)
class SubscriptionConfig:
    """
    The configuration class subscribed by consumers.
    """
    #: Subscription name, used to identify subscriptions in the producer configuration.
    #: The value of this attribute is a string.
    subscription_name: str
    #: Subscription type, including ``STREAM``, ``ROUND_ROBIN``, and ``KEY_PARTITIONS``.
    #: ``STREAM`` means single consumer consumption within a subscription group,
    #: ``ROUND_ROBIN`` means multiple consumers in a subscription group share load in a round-robin manner,
    #: ``KEY_PARTITIONS`` means multiple consumers in a subscription group share load by key partitioning.
    #: Currently, only ``STREAM`` type is supported; other types are not supported.
    #: The default subscription type is ``STREAM``.
    subscriptionType: SubscriptionType = SubscriptionType.STREAM
    #: Extended configuration.
    #: stored in dictionary form, allows users to customize configuration items.
    #: The default value is an empty dictionary. the dictionary generated through ``field(default_factory=dict)``.
    extend_config: Dict[str, str] = field(default_factory=dict)


class Element:
    """
    Element class containing an element ID and data buffer.

    Args:
            value (Union[bytes, memoryview]): data to send.
            ele_id (int, optional): element id. Default to ``0``.
    """

    def __init__(self, value: Union[bytes, memoryview], ele_id: int = 0) -> None:
        self.data = value
        self.id = ele_id
