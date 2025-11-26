#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

import time

import pytest
import importlib
import os
import yr
import subprocess
from common import SHUTDOWN_KEY, SHUTDOWN_VALUE, Add, AsyncAdd, init_in_process, Counter
from yr import Affinity, AffinityKind, AffinityType, LabelOperator, OperatorType


@pytest.mark.skip(
    reason="Check whether customextension has been written into the request body by viewing log file."
)
def test_invoke_instance_with_custom_extension(init_yr):
    """
    A function to test saving and loading the state of a stateful function.
    """
    opt = yr.InvokeOptions()
    opt.custom_extensions = {
        "endpoint": "CreateInstance1",
        "app_name": "CreateInstance2",
        "tenant_id": "CreateInstance3",
    }
    udf_class = Add.options(opt).invoke()

    opt.custom_extensions = {
        "endpoint": "InvokeInstance1",
        "app_name": "InvokeInstance2",
        "tenant_id": "InvokeInstance3",
    }
    ref = udf_class.add.options(opt).invoke()
    value = yr.get(ref)
    assert value == 1, f"member value after add one: {value}, 1 is expected."


@pytest.mark.skip(
    reason="Check whether 'antiOthers' has been written into the request body by viewing log file."
)
def test_anti_other_labels_success(init_yr):
    opt = yr.InvokeOptions()
    opt.preferred_anti_other_labels = True
    op1 = LabelOperator(OperatorType.LABEL_EXISTS, "label1")
    affinity = Affinity(AffinityKind.RESOURCE, AffinityType.PREFERRED, [op1])
    opt.schedule_affinities = [affinity]
    udf_class = Add.options(opt).invoke()
    ref = udf_class.add.options(opt).invoke()
    value = yr.get(ref)
    assert value == 1, f"member value after add one: {value}, 1 is expected."


@pytest.mark.smoke
def test_graceful_shutdown_with_terminate(init_yr):
    """test shutdown"""
    yr.kv_del(SHUTDOWN_KEY)
    instance = Add.invoke()
    ref = instance.get.invoke()
    yr.get(ref)
    instance.terminate()
    shutdown_value = yr.kv_read(SHUTDOWN_KEY, 30)
    fail_message = f"Shutdown failed. Expected value: {SHUTDOWN_VALUE}, Actual value: {shutdown_value}"
    assert shutdown_value == SHUTDOWN_VALUE, fail_message


@pytest.mark.skip(
    reason="Have to manually send SIGTERM signals to kill runtime processe during sleep."
)
def test_shutdown_with_manual_sigterm(init_yr):
    """test shutdown"""
    yr.kv_del(SHUTDOWN_KEY)
    Add.invoke()
    time.sleep(30)
    shutdown_value = yr.kv_read(SHUTDOWN_KEY, 30)
    fail_message = f"Shutdown failed. Expected value: {SHUTDOWN_VALUE}, Actual value: {shutdown_value}"
    assert shutdown_value == SHUTDOWN_VALUE, fail_message


@pytest.mark.smoke
def test_named_instance(init_yr):
    """test named instance"""
    opt = yr.InvokeOptions()
    opt.name = "actor1"
    opt.concurrency = 2
    instance = Add.options(opt).invoke()
    ref = instance.add.invoke()
    assert yr.get(ref) == 1
    instance2 = Add.options(opt).invoke()
    ref2 = instance2.add.invoke()
    assert yr.get(ref2) == 2
    opt2 = yr.InvokeOptions()
    opt2.name = "actor1"
    opt2.namespace = "ns1"
    opt2.concurrency = 2
    instance3 = Add.options(opt2).invoke()
    ref3 = instance3.add.invoke()
    assert yr.get(ref3) == 1
    instance4 = Add.options(opt2).invoke()
    ref4 = instance4.add.invoke()
    assert yr.get(ref4) == 2

    instance.terminate()
    instance3.terminate()


def test_list_named_instance(init_yr):
    """test named instance"""
    opt = yr.InvokeOptions()
    opt.name = "actor1"
    instance = Add.options(opt).invoke()
    ref = instance.add.invoke()
    assert yr.get(ref) == 1

    opt2 = yr.InvokeOptions()
    opt2.name = "actor1"
    opt2.namespace = "ns1"
    instance2 = Add.options(opt2).invoke()
    ref2 = instance2.add.invoke()
    assert yr.get(ref2) == 1

    actors = yr.list_named_instances()
    assert actors == ['ns1-actor1', 'actor1']

    instance.terminate()
    instance2.terminate()


@pytest.mark.smoke
def test_async_instance(init_yr):
    """test async instance"""
    opt = yr.InvokeOptions()
    opt.concurrency = 1
    startTime = time.time()
    instance = AsyncAdd.options(opt).invoke()
    ref1 = instance.get.invoke()
    ref2 = instance.get.invoke()
    assert yr.get(ref1) == 1
    assert yr.get(ref2) == 1
    endTime = time.time()
    assert endTime - startTime < 4


@pytest.mark.smoke
def test_actor_get_opts_env_var(init_yr):
    opt = yr.InvokeOptions()
    opt.env_vars = {"A": "A_VARS"}
    instance = Add.options(opt).invoke()
    ref = instance.return_custom_envs.invoke("A")
    assert yr.get(ref) == "A_VARS"


@pytest.mark.smoke
def test_thread_safe(init_yr):
    """test thread safe"""
    @yr.instance
    class Instance:
        def a(self, num):
            return num

    @yr.invoke
    def hello():
        from concurrent.futures import ThreadPoolExecutor
        try:
            with ThreadPoolExecutor(max_workers=2) as executor:
                task1 = executor.submit(create_actor)
                task2 = executor.submit(create_actor)
                task1.result()
                task2.result()
                return "success"
        except Exception as e:
            return "failed:" + str(e)

    def create_actor():
        ins = Instance.invoke()
        res = ins.a.invoke(100)
        return yr.get(res)

    f = hello.invoke()
    assert "success" in yr.get(f)


@pytest.mark.skip(reason="tmp skip")
def test_cloud_init(init_yr):
    @yr.instance
    class Actor:
        def __init__(self):
            self.count = 1

        def init_oncloud(self, function_id, server_address, ds_address):
            import multiprocessing
            context = multiprocessing.get_context("spawn")
            p = context.Process(target=init_in_process, args=(
                function_id, server_address, ds_address))
            p.start()
            p.join()
            return "success" + str(p.exitcode)

    opts = yr.InvokeOptions()
    opts.name = 'counter-actor'
    counter = Counter.options(opts).invoke()
    res = yr.get(counter.add.invoke())
    assert res == 1

    actor = Actor.invoke()
    function_id = os.getenv("YR_PYTHON_FUNC_ID")
    server_address = os.getenv("YR_SERVER_ADDRESS")
    ds_address = os.getenv("YR_DS_ADDRESS")
    res = yr.get(actor.init_oncloud.invoke(
        function_id, server_address, ds_address))
    assert "success0" in res

    counter = yr.get_instance('counter-actor')
    res = yr.get(counter.add.invoke())
    assert res == 3


@pytest.mark.smoke
def test_static_method(init_yr):
    @yr.instance
    class Instance:
        def na(self, a):
            return f"{a} normal method"

        @classmethod
        def ca(cls, a):
            return f"{a} classmethod method"

        @staticmethod
        def sa(a):
            return f"{a} staticmethod method"

    opts = yr.InvokeOptions()
    opts.name = 'instance-actor'
    ins = Instance.options(opts).invoke()
    res = yr.get(ins.na.invoke("hello"))
    assert "hello normal method" in res

    res = yr.get(ins.ca.invoke("hello"))
    assert "hello classmethod method" in res

    res = yr.get(ins.sa.invoke("hello"))
    assert "hello staticmethod method" in res

    ins = yr.get_instance('instance-actor')
    res = yr.get(ins.na.invoke("hi"))
    assert "hi normal method" in res

    res = yr.get(ins.ca.invoke("hi"))
    assert "hi classmethod method" in res

    res = yr.get(ins.sa.invoke("hi"))
    assert "hi staticmethod method" in res


@pytest.mark.smoke
def test_py_yr_stateful_order_preserve(init_yr):
    @yr.instance
    class Instance:
        a = 0

        def f(self):
            self.a = self.a + 1
            return self.a

    ins = Instance.invoke()
    res = [ins.f.invoke() for i in range(1000)]
    print(yr.get(res))
    for i in range(1, 1001):
        assert yr.get(res)[i-1] == i, "assert failed."
    ins.terminate()


def test_py_get_instance_after_recover(init_yr):
    @yr.instance
    class Instance:
        def __init__(self) -> None:
            self.num = 1

        def add(self):
            self.num = self.num + 1
            return self.num

    opts = yr.InvokeOptions()
    opts.name = "actor_recover"
    opts.recover_retry_times = 3
    ins = Instance.options(opts).invoke()
    assert (yr.get(ins.add.invoke()), 2)

    command = "ps -ef | grep fnruntime | grep -v grep | awk {'print $2'} | xargs kill -9"
    subprocess.run(command, shell=True, check=False)

    get_ins = yr.get_instance("actor_recover")
    assert (yr.get(get_ins.add.invoke()), 2)


def test_py_get_async_instance_after_recover(init_yr):
    @yr.instance
    class Instance:
        def __init__(self) -> None:
            self.num = 1

        async def add(self):
            self.num = self.num + 1
            return self.num

    opts = yr.InvokeOptions()
    opts.name = "actor_recover"
    opts.recover_retry_times = 3
    ins = Instance.options(opts).invoke()
    assert (yr.get(ins.add.invoke()), 2)

    command = "ps -ef | grep fnruntime | grep -v grep | awk {'print $2'} | xargs kill -9"
    subprocess.run(command, shell=True, check=False)

    get_ins = yr.get_instance("actor_recover")
    assert (yr.get(get_ins.add.invoke()), 2)


def test_py_get_detached_async_instance_after_recover(init_yr):
    @yr.instance
    class Instance:
        def __init__(self) -> None:
            self.num = 1

        async def add(self):
            self.num = self.num + 1
            return self.num

    opt = yr.InvokeOptions()
    opt.recover_retry_times = 3
    opt.name = "detach_actor_only_one"
    opt.custom_extensions["lifecycle"] = "detached"
    ins = Instance.options(opt).invoke()
    obj = ins.add.invoke()
    assert (yr.get(obj), 2)

    command = "ps -ef | grep fnruntime | grep -v grep | awk {'print $2'} | xargs kill -9"
    subprocess.run(command, shell=True, check=False)

    ins_get = yr.get_instance("detach_actor_only_one", "", 60)
    res = yr.get(ins_get.add.invoke())
    assert (res, 2)


@pytest.mark.smoke
def test_create_and_invoke_in_different_method(init_yr):
    def create_actor():
        @yr.instance
        class Actor:
            def add(self, x):
                return x + 1
        actors = [Actor.invoke() for _ in range(2)]
        return actors

    def invoke_actor():
        actors = create_actor()
        objs = [actor.add.invoke(1) for actor in actors]
        res = yr.get(objs)
        assert (res[0], 2)
        assert (res[1], 2)

    invoke_actor()


@pytest.mark.smoke
def test_group(init_yr):
    group_opts = yr.GroupOptions()
    group_name = "test"
    g = yr.Group(group_name, group_opts)
    opts = yr.InvokeOptions()
    opts.group_name = group_name
    ins = Counter.options(opts).invoke()
    g.invoke()
    res = ins.add.invoke()
    assert (yr.get(res), 1)
    g.terminate()


@pytest.mark.smoke
def test_group_suspend_resume(init_yr):
    group_opts = yr.GroupOptions()
    group_name = "test"
    g = yr.Group(group_name, group_opts)
    opts = yr.InvokeOptions()
    opts.group_name = group_name
    ins = Counter.options(opts).invoke()
    g.invoke()
    res = ins.add.invoke()
    assert (yr.get(res), 1)
    g.suspend()
    g.resume()
    g.terminate()
