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

import asyncio
import time
import os

import pytest

import yr


@pytest.mark.skip(reason="not support now")
def test_actor_generator(init_yr):
    @yr.instance
    class Demo():
        def __init__(self):
            print("start test generator")
            self.sum = 10

        def syncget(self):
            return self.sum

        async def get(self, msg):
            print(msg)
            await asyncio.sleep(0.2)
            return self.sum

        def syncgen(self):
            for i in range(5):
                time.sleep(0.1)
                yield i

        async def gen(self, msg):
            for i in range(5):
                await asyncio.sleep(0.1)
                yield i

    opt = yr.InvokeOptions(concurrency=10)
    d = Demo.options(opt).invoke()

    def actor_sync_get():
        get = d.syncget.invoke()
        assert yr.get(get) == 10

    actor_sync_get()

    def actor_sync_gen():
        gen = d.syncgen.invoke()
        index = 0
        for ref in gen:
            res = yr.get(ref)
            assert res == index
            print(res)
            index = index + 1
        assert index == 5

    actor_sync_gen()

    async def actor_async_get(msg):
        await d.get.options(opt).invoke(msg)

    async def actor_async_gen(msg):
        gen = d.gen.invoke(msg)
        index = 0
        async for ref in gen:
            res = await ref
            assert res == index
            index = index + 1
        assert index == 5

    async def multi_gen():
        await asyncio.gather(actor_async_gen("1-task"), actor_async_gen("2-task"), actor_async_gen("3-task"))

    async def multi_get():
        await asyncio.gather(actor_async_get("1-task"), actor_async_get("2-task"), actor_async_get("3-task"))

    # 3 个异步函数在actor侧应该是并发的， 用时小于3 * 2
    s = time.time()
    asyncio.run(multi_get())
    cost = time.time() - s
    assert cost < 0.6

    # 3 个异步流在actor侧应该是并发的， 用时小于3 * 5
    s = time.time()
    asyncio.run(multi_gen())
    cost = time.time() - s
    assert cost < 1


@pytest.mark.skip(reason="not support now")
def test_actor_generator_exception(init_yr):
    @yr.instance
    class Demo():
        def __init__(self):
            print("start test generator")
            self.sum = 10

        def syncgen(self):
            for i in range(5):
                if i == 2:
                    raise Exception("moke sync generator exception")
                yield i

        async def gen(self, msg):
            for i in range(5):
                if i == 2:
                    raise Exception("moke async generator exception")
                yield i

    opt = yr.InvokeOptions(concurrency=10)
    d = Demo.options(opt).invoke()

    def actor_sync_gen():
        gen = d.syncgen.invoke()
        index = 0
        for ref in gen:
            is_except = False
            res = 0
            try:
                res = yr.get(ref)
                assert res == index
                print(res, index)
            except Exception as e:
                is_except = True

            if index != 2:
                assert res == index
            else:
                assert is_except
            index = index + 1

    actor_sync_gen()

    async def actor_async_gen(msg):
        gen = d.gen.invoke(msg)
        index = 0
        async for ref in gen:
            is_except = False
            res = 0
            try:
                res = yr.get(ref)
                assert res == index
                print(res)
            except Exception as e:
                is_except = True

            if index != 2:
                assert res == index
            else:
                assert is_except
            index = index + 1

    async def multi_gen():
        await asyncio.gather(actor_async_gen("1-task"))

    asyncio.run(multi_gen())

    def actor_sync_generator_next():
        gen = d.syncgen.invoke()
        assert yr.get(next(gen)) == 0
        assert yr.get(next(gen)) == 1
        try:
            yr.get(next(gen))
        except Exception as e:
            assert "moke sync generator exception" in str(e)
        else:
            assert False
        with pytest.raises(StopIteration):
            yr.get(next(gen))
        with pytest.raises(StopIteration):
            yr.get(next(gen))

    actor_sync_generator_next()

    async def actor_async_generator_next(msg):
        gen = d.gen.invoke(msg)
        res = await next(gen)
        assert res == 0

        res = await next(gen)
        assert res == 1

        try:
            res = await next(gen)
        except Exception as e:
            assert "moke async generator exception" in str(e)
        else:
            assert False
        try:
            await next(gen)
        except StopIteration:
            print("finished")
        try:
            await next(gen)
        except StopIteration:
            print("finished")

    async def multi_gen_next():
        await asyncio.gather(actor_async_generator_next("1-task"))

    asyncio.run(multi_gen_next())


@pytest.mark.skip(reason="not support now")
def test_actor_generator_task(init_yr):
    @yr.instance
    class Demo():
        async def gen(self, msg):
            task_name = asyncio.current_task().get_name()
            for i in range(5):
                assert (
                    task_name == asyncio.current_task().get_name()
                ), "failed"
                yield i
            assert task_name == asyncio.current_task().get_name()

    opt = yr.InvokeOptions(concurrency=10)
    d = Demo.options(opt).invoke()

    for ref in d.gen.invoke("task"):
        print(yr.get(ref))
