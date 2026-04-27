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

"""E2E validation: InstanceProxy.real_id property"""

import yr


@yr.instance
class Counter:
    def __init__(self):
        self.count = 0

    def inc(self):
        self.count += 1
        return self.count


def main():
    yr.init()
    try:
        ins = Counter.invoke()

        logic_id = ins.instance_id
        real_id = ins.real_id

        print(f"logic_id : {logic_id}")
        print(f"real_id  : {real_id}")

        if not isinstance(real_id, str) or len(real_id) == 0:
            raise RuntimeError("real_id should be a non-empty string")
        print("real_id type/length check passed")

        result = yr.get(ins.inc.invoke())
        if result != 1:
            raise RuntimeError(f"expected 1, got {result}")
        print(f"actor method invoke passed: count={result}")

        ins.terminate()
        print("PASS: InstanceProxy.real_id e2e validation succeeded")
    finally:
        yr.finalize()


if __name__ == "__main__":
    main()
