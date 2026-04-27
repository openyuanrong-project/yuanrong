#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

"""
End-to-end test for idle timeout eviction feature.

Creates a function with idle_timeout=5s, invokes it once to warm up an instance,
then waits for the instance to be evicted by the idle timer.
"""

import time
import yr

yr.init()


@yr.invoke(num_cpus=0.1, idle_timeout=5)
def idle_test_func(x):
    return x * 2

print("[1] Invoking idle_test_func to create an instance...")
ref = idle_test_func.invoke(10)
result = yr.get(ref)
print(f"[1] Result: {result}")
if result != 20:
    raise RuntimeError(f"Expected 20, got {result}")

print("[2] Waiting 12s for idle timeout eviction (timeout=5s)...")
time.sleep(12)

print("[3] Invoking again — should create a NEW instance after eviction...")
ref2 = idle_test_func.invoke(21)
result2 = yr.get(ref2)
print(f"[3] Result: {result2}")
if result2 != 42:
    raise RuntimeError(f"Expected 42, got {result2}")

print("\n[OK] idle timeout e2e test passed!")

yr.finalize()
