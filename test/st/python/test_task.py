#!/usr/bin/env python3
# coding=UTF-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

import pytest
import yr
from yr import Affinity, AffinityKind, AffinityType, LabelOperator, OperatorType
from common import get, return_custom_envs

@pytest.mark.skip(reason="Check whether customextension has been written into the request body by viewing log file.")
def test_invoke_function_with_custom_extension(init_yr):
    opt = yr.InvokeOptions()
    opt.custom_extensions = {
        "endpoint": "InvokeFunction1",
        "app_name": "InvokeFunction2",
        "tenant_id": "InvokeFunction3",
    }
    ref = get.options(opt).invoke()
    assert yr.get(ref) == 1, "actual return of method get: {value}, 1 is expected."


@pytest.mark.skip(reason="Check whether 'antiOthers' has been written into the request body by viewing log file.")
def test_anti_other_labels_success(init_yr):
    opt = yr.InvokeOptions()
    opt.preferred_anti_other_labels = True
    op1 = LabelOperator(OperatorType.LABEL_EXISTS, "label1")
    affinity = Affinity(AffinityKind.RESOURCE, AffinityType.PREFERRED, [op1])
    opt.schedule_affinities = [affinity]
    ref = get.options(opt).invoke()
    assert yr.get(ref) == 1, "actual return of method get: {value}, 1 is expected."


@pytest.mark.smoke
def test_task_get_opts_env_var(init_yr):
    opt = yr.InvokeOptions()
    opt.env_vars = {"A" : "A_VARS"}
    ref = return_custom_envs.options(opt).invoke("A")
    assert yr.get(ref) == "A_VARS"


@pytest.mark.smoke
def test_kv_write_read(init_yr):
    yr.kv_write("key1", b"value1")
    v1 = yr.kv_read("key1")
    assert v1 == b"value1"
    yr.kv_del("key1")


@pytest.mark.smoke
def test_kv_write_read_with_param(init_yr):
    set_param = yr.SetParam()
    set_param.existence = yr.ExistenceOpt.NX
    set_param.write_mode = yr.WriteMode.NONE_L2_CACHE_EVICT
    set_param.ttl_second = 100
    # Check whether shared disk is enabled.
    yr.kv_write_with_param("key1", b"value1", set_param)
    v1 = yr.kv_read("key1")
    assert v1 == b"value1"
    yr.kv_del("key1")


@pytest.mark.smoke
def test_kv_m_write_tx_read_with_param(init_yr):
    mset_param = yr.MSetParam()
    mset_param.existence = yr.ExistenceOpt.NX
    mset_param.write_mode = yr.WriteMode.NONE_L2_CACHE_EVICT
    mset_param.ttl_second = 100
    # Check whether shared disk is enabled.
    yr.kv_m_write_tx(["key1", "key2"], [b"value1", b"value2"], mset_param)
    v1 = yr.kv_read("key1")
    assert v1 == b"value1"
    v2 = yr.kv_read("key2")
    assert v2 == b"value2"
    yr.kv_del(["key1", "key2"])


@pytest.mark.smoke
def test_kv_write_failed(init_yr):
    try:
        set_param = yr.SetParam()
        set_param.ttl_second = -1
        yr.kv_write_with_param("key1", b"value1", set_param)
    except Exception as e:
        print(e)
        assert "can't convert negative value to uint32_t" in str(e)
    try:
        mset_param = yr.MSetParam()
        mset_param.existence = yr.ExistenceOpt.NONE
        mset_param.cache_type = yr.CacheType.DISK
        yr.kv_m_write_tx(["key1", "key2"], [b"value1", b"value2"], mset_param)
    except Exception as e:
        print(e)
        assert "existence should be NX" in str(e)
    try:
        mset_param = yr.MSetParam()
        yr.kv_m_write_tx(["key1", "key2"], [b"value1"], mset_param)
    except Exception as e:
        print(e)
        assert "arguments list size not equal" in str(e)
    try:
        mset_param = yr.MSetParam()
        yr.kv_m_write_tx([], [], mset_param)
    except Exception as e:
        print(e)
        assert "The keys should not be empty" in str(e)
