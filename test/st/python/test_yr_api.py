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
from common import CUSTOM_SHUTDOWN_OPT
from common import may_throw_exception


@pytest.mark.smoke
def test_task_invoke(init_yr):
    @yr.invoke
    def dis_sum(args):
        return sum(args)

    ref = dis_sum.invoke([1, 2, 3])
    num = yr.get(ref, 30)
    assert num == 6


@pytest.mark.smoke
def test_task_invoke_with_ref(init_yr):
    @yr.invoke
    def get_nums():
        return [1, 2, 3]

    @yr.invoke
    def dis_sum(args):
        return sum(args)

    ref = dis_sum.invoke(get_nums.invoke())
    num = yr.get(ref, 30)
    assert num == 6


@pytest.mark.smoke
def test_task_invoke_with_nested_ref(init_yr):
    @yr.invoke
    def get_num(x):
        return x

    @yr.invoke
    def dis_sum(args):
        print(args)
        import sys
        sys.stdout.flush()
        return sum(yr.get(args))

    ref = dis_sum.invoke([get_num.invoke(1), get_num.invoke(2), get_num.invoke(3)])
    num = yr.get(ref, 30)
    assert num == 6


@pytest.mark.smoke
def test_task_invoke_with_big_bytes(init_yr):
    @yr.invoke
    def get(x):
        return x

    arg = b"1" * 1024 * 200
    ref = get.invoke(arg)
    ret = yr.get(ref, 30)
    assert ret == arg


@pytest.mark.smoke
def test_return_nested_objref(init_yr):
    @yr.invoke
    def sum_list(refs):
        sum = 0
        for ref in refs:
            sum += yr.get(ref, 30)
        return sum

    @yr.invoke
    def remote_add(refs):
        result = sum_list.invoke(refs)
        return yr.get(result, 30)

    list_of_objref = [yr.put(1)]
    objref = remote_add.invoke(list_of_objref)
    num = yr.get(objref, 30)
    assert num == 1


@pytest.mark.slow
def test_task_invoke_with_recycle_time_1(init_yr_config):
    conf = init_yr_config
    conf.recycle_time = 1
    yr.init(conf)

    @yr.invoke
    def get_instance_id():
        import os
        return os.environ["INSTANCE_ID"]

    @yr.instance
    class Instance:
        def invoke_task(self):
            import time
            ret1 = yr.get(get_instance_id.invoke())
            time.sleep(1.2)
            ret2 = yr.get(get_instance_id.invoke())
            return ret1, ret2

    ins = Instance.invoke()
    ret = yr.get(ins.invoke_task.invoke())
    assert ret[0] != ret[1]


@pytest.mark.smoke
def test_task_invoke_with_return_nums(init_yr):
    @yr.invoke(return_nums=3)
    def func_returns():
        return 1, 2, 3

    obj_ref1, obj_ref2, obj_ref3 = func_returns.invoke()
    assert yr.get([obj_ref1, obj_ref2, obj_ref3]) == [1, 2, 3]

    @yr.invoke(return_nums=0)
    def func():
        return

    obj_ref = func.invoke()
    assert obj_ref == None


@pytest.mark.smoke
def test_actor_return_values(init_yr):
    @yr.instance
    class Actor:
        @yr.method(return_nums=0)
        def method0(self):
            return 1

        @yr.method(return_nums=1)
        def method1(self):
            return 1

        @yr.method(return_nums=2)
        def method2(self):
            return 1, 2

        @yr.method(return_nums=3)
        def method3(self):
            return 1, 2, 3

    actor = Actor.invoke()

    obj_ref = actor.method0.invoke()
    assert obj_ref == None

    obj_ref1 = actor.method1.invoke()
    assert yr.get(obj_ref1) == 1

    obj_ref1, obj_ref2 = actor.method2.invoke()
    assert yr.get([obj_ref1, obj_ref2]) == [1, 2]

    obj_ref1, obj_ref2, obj_ref3 = actor.method3.invoke()
    assert yr.get([obj_ref1, obj_ref2, obj_ref3]) == [1, 2, 3]

    actor.terminate()


@pytest.mark.skip(reason="not support k8s pattern")
def test_task_terminate_when_create_failed(init_yr_config):
    conf = init_yr_config
    yr.init(conf)

    @yr.instance
    class Instance:
        def invoke_task(self):
            return "hello, world"

    opt = yr.InvokeOptions(cpu=16001, memory=128)
    try:
        ins = Instance.options(opt).invoke()
        ret = yr.get(ins.invoke_task.invoke())
    except Exception as e:
        num = 1
        ins.terminate()
        assert num == 1


@pytest.mark.smoke
def test_task_with_concurrency_101(init_yr):
    @yr.invoke
    def f():
        import subprocess
        import os
        return int(subprocess.getoutput(f"ls -l /proc/{os.getpid()}/task|wc -l"))

    opts = yr.InvokeOptions()
    opts.concurrency = 101
    assert yr.get(f.options(opts).invoke()) > 101


@pytest.mark.smoke
def test_actor_with_pass_to_runtime(init_yr):
    @yr.instance
    class Counter:
        def __init__(self):
            self.cnt = 0

        def add(self):
            self.cnt += 1
            return self.cnt

        @classmethod
        def ca(cls, a):
            return a

        @staticmethod
        def sa(a):
            return a

    @yr.invoke
    def func(c):
        res = yr.get(c.add.invoke()) + yr.get(c.sa.invoke(10)) + yr.get(c.ca.invoke(10))
        return res

    counter = Counter.invoke()
    assert yr.get(counter.add.invoke()) == 1
    assert yr.get(func.invoke(counter)) == 22


def test_no_gpu_resources_func(init_yr):
    @yr.invoke
    def g(i):
        return i

    opts = yr.InvokeOptions()
    opts.custom_resources["nvidia.com/gpu"] = 1

    try:
        res = g.options(opts).invoke(2)
        yr.get(res, 30)
    except Exception as e:
        print(e)
        assert "code: 1006" in str(e)
        assert "message: invalid resource parameter" in str(e)


@pytest.mark.smoke
def test_invoke_outside_init(init_yr_config):
    conf = init_yr_config

    @yr.invoke
    def f(x):
        return x

    for i in range(10):
        yr.init(conf)
        assert yr.get(f.invoke(i)) == i
        yr.finalize()


@pytest.mark.smoke
def test_get_allow_partial(init_yr):
    @yr.invoke(invoke_options=CUSTOM_SHUTDOWN_OPT)
    def func():
        import time
        time.sleep(3)
        return 0

    obj1 = yr.put(0)
    obj2 = func.invoke()
    data = yr.get([obj1, obj2], 1, True)

    assert data == [0, None]


@pytest.mark.smoke
def test_invoke_with_nestid_in_return(init_yr):
    @yr.invoke(invoke_options=CUSTOM_SHUTDOWN_OPT)
    def f():
        return "good"

    @yr.invoke(invoke_options=CUSTOM_SHUTDOWN_OPT)
    def h():
        idlist = [f.invoke() for i in range(10)]
        return idlist

    ready_list = yr.get(h.invoke(), 30)
    assert yr.get(ready_list, 30) == yr.get([f.invoke() for i in range(10)])


@pytest.mark.smoke
def test_user_func_import_ssl(init_yr):
    @yr.invoke
    def import_ssl():
        import ssl
        return ssl.OPENSSL_VERSION

    ret = import_ssl.invoke()
    try:
        print(yr.get(ret))
    except Exception as e:
        print(e)
        assert False
    assert True


@pytest.mark.smoke
def test_invoke_with_raise_runtime_error(init_yr):
    @yr.invoke
    def raise_error():
        raise RuntimeError("aaa")

    @yr.invoke
    def func(x):
        return x

    with pytest.raises(RuntimeError):
        yr.get(raise_error.invoke())

    with pytest.raises(RuntimeError):
        yr.get(func.invoke([raise_error.invoke()]))


@pytest.mark.smoke
def test_cancel_cloud(init_yr):
    @yr.invoke
    def j(x):
        time.sleep(3)
        return x

    @yr.invoke
    def h(x):
        res_j = j.invoke(x)
        yr.cancel(res_j)
        return yr.get(res_j)

    @yr.invoke
    def g(x):
        res_h = h.invoke(x)
        return yr.get(res_h)

    @yr.invoke
    def f(x):
        res_g = g.invoke(x)
        return yr.get(res_g)

    res = f.invoke("test")
    try:
        yr.get(res)
    except RuntimeError as e:
        print(e)
        assert "the obj has been cancelled" in str(e)


@pytest.mark.smoke
def test_kv_read_write_del_success(init_yr):
    yr.kv_write("key1", b"value1")
    assert yr.kv_read("key1") == b"value1"

    @yr.invoke
    def func():
        yr.kv_write("key2", b"value2")
        v2 = yr.kv_read("key2")
        v1 = yr.kv_read("key1")
        yr.kv_del("key2")
        return v1, v2

    assert yr.get(func.invoke()) == (b"value1", b"value2")

    yr.kv_del("key1")
    with pytest.raises(RuntimeError):
        yr.kv_read("key1", 0)
    with pytest.raises(RuntimeError):
        yr.kv_read("key2", 0)


@pytest.mark.smoke
def test_actor_with_order(init_yr):
    @yr.instance
    class Counter:
        def __init__(self):
            self.cnt = 0

        def add(self):
            self.cnt += 1
            return self.cnt

    @yr.invoke
    def func(c):
        rets2 = [c.add.invoke() for _ in range(100)]
        return rets2

    counter = Counter.invoke()
    rets = [counter.add.invoke() for _ in range(100)]
    assert yr.get(rets) == [i for i in range(1, 101)]
    ret = func.invoke(counter)
    rets3 = yr.get(ret)
    assert yr.get(rets3) == [i for i in range(101, 201)]


def test_task_with_raise_error_when_dependency_error(init_yr):
    class CustomException(Exception):
        pass

    @yr.invoke
    def func1():
        raise CustomException("I am failed")

    @yr.invoke
    def func2(x):
        return x

    ret1 = func1.invoke()
    with pytest.raises(CustomException):
        yr.get(ret1)
    ret2 = func2.invoke(ret1)
    with pytest.raises(CustomException):
        yr.get(ret2)

    @yr.instance
    class Counter:
        def __init__(self):
            raise CustomException("I am failed.")

        def add(self):
            self.cnt += 1
            return self.cnt

    counter = Counter.invoke()
    ret = counter.add.invoke()
    with pytest.raises(CustomException):
        yr.get(ret)


def test_invoke_inconsistent_return_values(init_yr):
    @yr.invoke(return_nums=3)
    def func1():
        return 1, 2

    @yr.invoke(return_nums=3)
    def func2():
        return 1

    @yr.invoke(return_nums=2)
    def func3():
        return 1, 2, 3

    obj_ref1, obj_ref2, obj_ref3 = func1.invoke()

    try:
        yr.get(obj_ref1)
    except ValueError as e:
        print(e)
        assert "not enough values to unpack (expected 3, got 2)" in str(e)

    obj_ref = func2.invoke()
    try:
        yr.get(obj_ref)
    except TypeError as e:
        print(e)
        assert "cannot unpack non-iterable <class 'int'> object" in str(e)

    obj_ref4, obj_ref5 = func3.invoke()

    try:
        yr.get(obj_ref4)
    except ValueError as e:
        print(e)
        assert "not enough values to unpack (expected 2, got 3)" in str(e)


@pytest.mark.smoke
def test_kv_write_and_get_with_new_arguments_success(init_yr):
    # Test ExistenceOpt
    yr.kv_write("key1", b"value1", yr.ExistenceOpt.NONE, yr.WriteMode.NONE_L2_CACHE_EVICT, 0)
    assert yr.kv_read("key1") == b"value1"
    yr.kv_write("key1", b"value1", yr.ExistenceOpt.NONE, yr.WriteMode.NONE_L2_CACHE_EVICT, 0)
    with pytest.raises(RuntimeError):
        yr.kv_write("key1", b"value1", yr.ExistenceOpt.NX, yr.WriteMode.NONE_L2_CACHE_EVICT, 0)
    yr.kv_del("key1")

    # Test ttl_second
    yr.kv_write("key1", b"value1", yr.ExistenceOpt.NONE, yr.WriteMode.NONE_L2_CACHE_EVICT, 1)
    assert yr.kv_read("key1") == b"value1"
    time.sleep(1.5)
    with pytest.raises(RuntimeError):
        yr.kv_read("key1", 0)


@pytest.mark.smoke
def test_kv_set_and_get(init_yr):
    yr.kv_set("key1", b"value1")
    assert yr.kv_get("key1") == b"value1"
    get_param = yr.GetParam()
    get_param.offset = 0
    get_param.size = 0
    params = yr.GetParams()
    params.get_params = [get_param]
    res = yr.kv_get_with_param(["key1"], params)
    assert res[0] == b"value1"
    yr.kv_del("key1")


@pytest.mark.smoke
def test_get_with_pram_partial(init_yr):
    key1 = "kv-py-id-testGetWithParamPartial0"
    key2 = "kv-py-id-testGetWithParamPartial1"
    get_param = yr.GetParam()
    get_param.offset = 0
    get_param.size = 0
    params = yr.GetParams()
    params.get_params = [get_param, get_param]
    yr.kv_set(key1, b"value")
    res = yr.kv_get_with_param([key1, key2], params, 4)
    count = 0
    for item in res:
        if item is None:
            continue
        assert item == b"value"
        count += 1
    assert count == 1
    yr.kv_del(key1)


@pytest.mark.smoke
def test_get_with_pram_failed(init_yr):
    key1 = "kv-id-testFailed"
    get_param = yr.GetParam()
    get_param.offset = 0
    get_param.size = 0
    params = yr.GetParams()
    params.get_params = [get_param]
    try:
        yr.kv_get_with_param([key1], params, 4)
    except RuntimeError as e:
        print(e)
        assert "Get timeout 4000ms from datasystem, all failed" in str(e)
    try:
        yr.kv_get_with_param([], params, 4)
    except ValueError as e:
        print(e)
        assert "Get params size is not equal to keys size" in str(e)
    try:
        yr.kv_get_with_param([key1], yr.GetParams(), 4)
    except ValueError as e:
        print(e)
        assert "Get params does not accept empty list" in str(e)

    get_param2 = yr.GetParam()
    get_param2.offset = -1
    get_param2.size = 0
    params.get_params = [get_param2]
    try:
        yr.kv_get_with_param([key1], params, 4)
    except OverflowError as e:
        print(e)
        assert "can't convert negative value to uint64_t" in str(e)


@pytest.mark.smoke
def test_get_empty_list(init_yr):
    ret = yr.get([], 30)
    assert ret == []


@pytest.mark.smoke
def test_custom_envs_config(init_yr_config):
    conf = init_yr_config
    key = "LD_LIBRARY_PATH"
    value = "${LD_LIBRARY_PATH}:${YR_FUNCTION_LIB_PATH}/depend"
    conf.custom_envs = {key: value}
    yr.init(conf)

    @yr.invoke
    def return_custom_envs(k):
        v = os.getenv(k)
        print("custom env: ", v)
        return v

    ref = return_custom_envs.invoke(key)
    env = yr.get(ref, 30)
    print("custom env: ", env)
    assert ("depend" in env)
    assert ("YR_FUNCTION_LIB_PATH" not in env)
    assert ("LD_LIBRARY_PATH" not in env)


@pytest.mark.smoke
def test_invoke_cpp_task(init_yr):
    cppf = yr.cpp_function("PlusOne", "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stcpp:$latest")
    opt = yr.InvokeOptions()
    res = cppf.options(opt).invoke(10)
    cppres = yr.get(res)
    assert cppres == 11


@pytest.mark.smoke
def test_invoke_cpp_task_with_ref(init_yr):
    obj = yr.put(10)
    cppf = yr.cpp_function("PlusOne", "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stcpp:$latest")
    opt = yr.InvokeOptions()
    res = cppf.options(opt).invoke(obj)
    cppres = yr.get(res)
    assert cppres == 11


@pytest.mark.smoke
def test_invoke_cpp_actor(init_yr):
    opt = yr.InvokeOptions()
    cpp_function_id = "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stcpp:$latest"
    cpp_ins = yr.cpp_instance_class("Counter", "Counter::FactoryCreate", cpp_function_id).options(opt).invoke(1)
    add_res = cpp_ins.Add.invoke(10)
    assert yr.get(add_res) == 11


@pytest.mark.smoke
def test_invoke_cpp_actor_with_ref(init_yr):
    obj = yr.put(10)
    opt = yr.InvokeOptions()
    cpp_function_id = "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stcpp:$latest"
    cpp_ins = yr.cpp_instance_class("Counter", "Counter::FactoryCreate", cpp_function_id).options(opt).invoke(1)
    add_res = cpp_ins.Add.invoke(obj)
    assert yr.get(add_res) == 11


@pytest.mark.smoke
def test_invoke_java_task(init_yr):
    javaf = yr.java_function("com.yuanrong.testutils.TestUtils", "returnInt",
                             "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stjava:$latest")
    res = javaf.invoke(10)
    ret = yr.get(res)
    assert ret == 10


@pytest.mark.smoke
def test_invoke_java_actor(init_yr):
    java_cls = yr.java_instance_class("com.yuanrong.testutils.Counter",
                                      "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-stjava:$latest")
    java_ins = java_cls.invoke()
    res = java_ins.addOne.invoke()
    ret = yr.get(res)
    assert ret == 1

    opt = yr.InvokeOptions()
    opt.name = "actor1"
    opt.concurrency = 2
    java_ins2 = java_cls.options(opt).invoke()
    res2 = java_ins2.addOne.invoke()
    assert yr.get(res2) == 1
    java_ins3 = java_cls.options(opt).invoke()
    res3 = java_ins3.addOne.invoke()
    assert yr.get(res3) == 2


@pytest.mark.smoke
def test_get_instance(init_yr_config):
    conf = init_yr_config
    yr.init(conf)

    @yr.instance
    class Instance:
        def __init__(self):
            self.num = 1

        def add(self):
            return self.num

    opt = yr.InvokeOptions()
    opt.name = "actor"
    opt.namespace = "ns"
    opt.concurrency = 1
    ins = Instance.options(opt).invoke()
    obj1 = ins.add.invoke()
    ret1 = yr.get(obj1)
    ins_get = yr.get_instance("actor", "ns")
    obj2 = ins_get.add.invoke()
    ret2 = yr.get(obj2)
    assert ret1 == ret2

    opt2 = yr.InvokeOptions()
    opt2.name = "actor"
    ins = Instance.options(opt2).invoke()
    obj3 = ins.add.invoke()
    ret3 = yr.get(obj3)
    ins_get_2 = yr.get_instance("actor")
    obj4 = ins_get_2.add.invoke()
    ret4 = yr.get(obj4)
    assert ret3 == ret4

    opt3 = yr.InvokeOptions()
    opt3.name = "actor_not_exist"
    try:
        ins_get_3 = yr.get_instance("actor_not_exist")
    except Exception as e:
        assert "failed to get instance" in str(e)

    ins_recreate = Instance.options(opt3).invoke()
    obj_recreate = ins_recreate.add.invoke()
    ret_recreate = yr.get(obj_recreate)
    assert ret_recreate == 1


@pytest.mark.smoke
def test_get_instance_timeout(init_yr_config):
    conf = init_yr_config
    yr.init(conf)

    @yr.instance
    class Instance:

        def __init__(self):
            time.sleep(5)
            self.num = 1

    opt = yr.InvokeOptions()
    opt.name = "actor_timeout"
    opt.namespace = "ns_timeout"
    ins = Instance.options(opt).invoke()
    time.sleep(1)
    value = 1
    try:
        ins_get = yr.get_instance("actor_timeout", "ns_timeout", 2)
    except TimeoutError as e:
        value = 2
    assert value == 2


@pytest.mark.smoke
def test_get_async_instance(init_yr_config):
    conf = init_yr_config
    yr.init(conf)

    @yr.instance
    class AysncInstance:
        def __init__(self):
            self.num = 1

        async def add(self):
            await asyncio.sleep(1)
            self.num += 1
            return self.num

        def get(self):
            return self.add()

    opt = yr.InvokeOptions()
    opt.name = "actor"
    opt.namespace = "ns"
    opt.concurrency = 1
    ins = AysncInstance.options(opt).invoke()
    obj1 = ins.get.invoke()
    ret1 = yr.get(obj1)
    ins_get = yr.get_instance("actor", "ns")
    obj2 = ins_get.get.invoke()
    ret2 = yr.get(obj2)
    ins.terminate()
    assert ret1 + 1 == ret2


@pytest.mark.smoke
def test_get_async_instance_in_actor_proxy(init_yr_config):
    conf = init_yr_config
    yr.init(conf)

    @yr.instance
    class Counter:
        def __init__(self):
            self.num = 1

        def get(self):
            return self.num

    @yr.instance
    class CounterProxy:
        def __init__(self):
            self.num = 1

        def InvokeCounterGet(self):
            counter = yr.get_instance("actor2", "ns")
            obj = counter.get.invoke()
            return yr.get(obj)

    opt = yr.InvokeOptions()
    opt.name = "actor2"
    opt.namespace = "ns"
    counter = Counter.options(opt).invoke()
    obj1 = counter.get.invoke()
    ret1 = yr.get(obj1)

    counter_proxy = CounterProxy.invoke()
    obj2 = counter_proxy.InvokeCounterGet.invoke()
    ret2 = yr.get(obj2)
    counter.terminate()
    counter_proxy.terminate()
    assert ret1 == ret2


@pytest.mark.skip(
    reason="This use case will now fail; let's temporarily disable it and resolve it later through an issue ID64WF.")
def test_get_order_instance_after_teminate(init_yr_config):
    conf = init_yr_config
    yr.init(conf)

    @yr.instance
    class Counter:
        def add(self, x):
            return x + 1

    for i in range(3):
        ins_list = []
        ref_list = []
        for j in range(3):
            print("start inner cycle: ", j)
            opts = yr.InvokeOptions()
            opts.cpu = 300
            opts.memory = 128
            opts.name = "actor-" + str(j)
            ins = Counter.options(opts).invoke()
            assert yr.get(ins.add.invoke(1)) == 2
            ins_list.append(ins)

            ins_get = yr.get_instance(opts.name)
            ref = ins_get.add.invoke(1)
            ref_list.append(ref)
        for j in range(len(ins_list)):
            assert yr.get(ref_list[j]) == 2
        for j in range(len(ins_list)):
            ins_list[j].terminate(True)


@pytest.mark.smoke
def test_invoke_with_redefine(init_yr):
    @yr.invoke
    def get_num():
        return 1

    assert yr.get(get_num.invoke()) == 1

    @yr.invoke
    def get_num():
        return 2

    assert yr.get(get_num.invoke()) == 2

    @yr.instance
    class Counter:
        def __init__(self):
            self.num = 1

        def get(self):
            return self.num

    counter = Counter.invoke()
    assert yr.get(counter.get.invoke()) == 1

    @yr.instance
    class Counter:
        def __init__(self):
            self.num = 2

        def get(self):
            return self.num

    counter = Counter.invoke()
    assert yr.get(counter.get.invoke()) == 2


@pytest.mark.smoke
def test_task_wait_exception(init_yr):
    not_ready_ids = []
    for num in range(0, 10):
        not_ready_ids.append(may_throw_exception.invoke(num))
    ready_ids, not_ready_ids = yr.wait(not_ready_ids, wait_num=10)
    assert len(not_ready_ids) == 0
    assert len(ready_ids) == 10
    print(ready_ids)
    for i in range(len(ready_ids)):
        if i % 2 != 0:
            assert i == yr.get(ready_ids[i])
            continue
        with pytest.raises(RuntimeError):
            yr.get(ready_ids[i])


@pytest.mark.smoke
def test_task_wait_exception_with_num_1(init_yr):
    not_ready_ids = []
    total = 10
    for num in range(0, total):
        not_ready_ids.append(may_throw_exception.invoke(num))
    i = 0
    while i < total:
        ready_ids, not_ready_ids = yr.wait(not_ready_ids, wait_num=1)
        assert len(ready_ids) == 1
        assert len(not_ready_ids) == total - i - 1
        i += 1


@pytest.mark.smoke
def test_handle_resource_group_failed(init_yr):
    try:
        rg = yr.create_resource_group([], "rgname1")
    except Exception as e:
        print(e)
        assert "invalid bundles number, actual: 0" in str(e)
    try:
        rg = yr.create_resource_group("aaa", "rgname1")
    except Exception as e:
        print(e)
        assert "invalid bundles type, actual:" in str(e)
    try:
        rg = yr.create_resource_group(["cpu", "memory"], "rgname1")
    except Exception as e:
        print(e)
        assert "object has no attribute" in str(e)
    try:
        rg = yr.create_resource_group([{"cpu": "1", "memory": "2"}], "rgname1")
    except Exception as e:
        print(e)
        assert "must be real number, not str" in str(e)
    try:
        rg = yr.create_resource_group([{"cpu": 500, "memory": 500}], [])
    except Exception as e:
        print(e)
        assert "invalid name type:" in str(e)
    try:
        rg = yr.create_resource_group([{"cpu": 500, "memory": 500}], "primary")
    except Exception as e:
        print(e)
        assert "please set the name other than primary or empty" in str(e)
    try:
        rg = yr.create_resource_group([{"cpu": -1, "memory": 500}], "rgname1")
    except Exception as e:
        print(e)
        assert "bundle index: 0, please set the value of cpu >= 0." in str(e)
    try:
        rg = yr.create_resource_group([{"CPU": 500, "memory": 500}], "rgname1")
        rg.wait(-2)
    except Exception as e:
        print(e)
        assert "should be greater than or equal to -1" in str(e)
        yr.remove_resource_group("rgname1")
    try:
        rg = yr.remove_resource_group("")
    except Exception as e:
        print(e)
        assert "please set the name other than primary or empty" in str(e)


@pytest.mark.skip(reason="not support until fs support rgroup")
def test_handle_resource_group(init_yr):
    @yr.invoke
    def get_num(x):
        return x
    rg = yr.create_resource_group([{"CPU": 1000, "Memory": 1000}, {"NPU": 1}, {"CPU": 500, "NPU": 0}], "rgname1")
    rg.wait(10)
    rgp = yr.ResourceGroupOptions()
    rgp.resource_group_name = "rgname1"
    rgp.bundle_index = 0
    opts = yr.InvokeOptions()
    opts.resource_group_options = rgp
    ins = Instance.options(opts).invoke()
    results = [get_num.invoke(6) for i in range(1)]
    assert yr.get(results[0]) == 6
    yr.remove_resource_group("rgname1")
 
 
@pytest.mark.smoke
def test_task_invoke_with_return_small_and_big_obj(init_yr):
    @yr.invoke(return_nums=2)
    def func_returns():
        return 1, b"1" * 1024 * 20000
 
    obj_ref1, obj_ref2 = func_returns.invoke()
    time.sleep(1)
    assert yr.get([obj_ref1, obj_ref2]) == [1, b"1" * 1024 * 20000]