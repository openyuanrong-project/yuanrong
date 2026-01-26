# distutils: language = c++
# cython: embedsignature = True
# cython: language_level = 3
# cython: c_string_encoding = default
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
import concurrent
import inspect
import logging
import json
import os
import traceback
import threading
import traceback
import uuid
from pathlib import Path
from dataclasses import asdict
from typing import List, Tuple, Union, Callable
from cython.operator import dereference, postincrement
from cpython.exc cimport PyErr_CheckSignals

import yr
from yr.config import SchedulingAffinityType, FunctionGroupOptions, ResourceGroupOptions, GroupOptions
from yr.common.types import GroupInfo, CommonStatus, Resource, Resources, BundleInfo, Option, RgInfo, ResourceGroupUnit
from yr.common import constants
from yr.common.utils import generate_random_id, create_new_event_loop
from yr.config_manager import ConfigManager
from yr.err_type import ErrorCode, ModuleCode, ErrorInfo
from yr.executor.executor import Executor
from yr.executor.instance_manager import InstanceManager, InstancePackage
from yr.libruntime_pb2 import FunctionMeta, LanguageType, InvokeType, ApiType, Signal
from yr.object_ref import ObjectRef
from yr import log
from yr.common.utils import GaugeData, UInt64CounterData, DoubleCounterData
from yr.device import DataType, DeviceBufferParam, DataInfo
from yr.runtime import (ExistenceOpt, WriteMode, CacheType, ConsistencyType, SetParam, MSetParam, CreateParam,
                        GetParam, GetParams, AlarmInfo, AlarmSeverity)
from yr import runtime_env
from yr.exception import YRInvokeError
from yr.stream import (Element, ProducerConfig, SubscriptionConfig,
                       SubscriptionType)
from yr.accelerate.shm_broadcast import Handle, MessageQueue, decode, ResponseStatus
from yr.accelerate.executor import ACCELERATE_WORKER, Worker
from cpython cimport PyBytes_FromStringAndSize
from libc.stdint cimport uint64_t
from libcpp cimport bool
from libcpp.memory cimport make_shared, nullptr, shared_ptr, dynamic_pointer_cast
from libcpp.pair cimport pair
from libcpp.string cimport string
from libcpp.unordered_map cimport unordered_map
from libcpp.unordered_set cimport unordered_set
from libcpp.vector cimport vector

from yr.includes.libruntime cimport (CApiType, CSignal, CBuffer, CDataObject, CElement,
CErrorCode, CErrorInfo, CFunctionMeta, CInternalWaitResult, CInvokeArg,
CInvokeOptions, CInvokeType, CModuleCode,
CLanguageType, CLibruntimeConfig,
CLibruntimeManager,move,CLibruntime, CGroupOptions,
CProducerConf, CStreamConsumer,
CStreamProducer, CSubscriptionConfig,
CSubscriptionType, 
CExistenceOpt, CSetParam, CMSetParam, CCreateParam, CStackTraceInfo, CWriteMode, CCacheType, CConsistencyType,
CGetParam, CGetParams,
CMultipleReadResult, CDevice, CMultipleDelResult, CUInt64CounterData, CDoubleCounterData, NativeBuffer, StringNativeBuffer, CInstanceOptions, CGaugeData, CTensor, CDataType, CResourceUnit, CAlarmInfo, CAlarmSeverity, CFunctionGroupOptions, CBundleAffinity, CFunctionGroupRunningInfo, CFiberEvent,
CClusterAccessInfo, AutoGetClusterAccessInfo, CResourceGroupSpec, CResourceGroupOptions, CAccelerateMsgQueueHandle, QueryNamedInsResponse, CResourceGroupUnit, CRgInfo, CBundleInfo, CResources, CResource, CType, CScalar)

include "includes/affinity.pxi"
include "includes/buffer.pxi"
include "includes/serialization.pxi"

_logger = logging.getLogger(__name__)

_serialization_ctx = None
_eventloop_for_default_cg = None
_thread_for_default_cg = None
_actor_is_async = False

class GeneratorEndError(RuntimeError):
    """GeneratorEndError stream error"""
    def __init__(self, message):
        super().__init__(message)
        self.message = message

    def __str__(self):
        return f"generator stop : {self.message}"

cdef error_code_from_cpp(CErrorCode code):
    if code==CErrorCode.ERR_GENERATOR_FINISHED:
        return ErrorCode.ERR_GENERATOR_FINISHED

cdef CErrorCode error_code_from_py(error_code: ErrorCode):
    cdef CErrorCode c_error_code
    c_error_code = CErrorCode.ERR_OK
    if error_code == ErrorCode.ERR_USER_CODE_LOAD:
        c_error_code = CErrorCode.ERR_USER_CODE_LOAD
    elif error_code == ErrorCode.ERR_USER_FUNCTION_EXCEPTION:
        c_error_code = CErrorCode.ERR_USER_FUNCTION_EXCEPTION
    elif error_code == ErrorCode.ERR_EXTENSION_META_ERROR:
        c_error_code = CErrorCode.ERR_EXTENSION_META_ERROR
    return c_error_code

cdef CModuleCode module_code_from_py(module_code: ModuleCode):
    cdef CModuleCode c_module_code
    c_module_code = CModuleCode.RUNTIME
    if module_code == ModuleCode.CORE:
        c_module_code = CModuleCode.CORE
    return c_module_code

cdef error_info_from_cpp(CErrorInfo c_error_info):
    if c_error_info.OK():
        return ErrorInfo()
    return ErrorInfo(c_error_info.Code(), c_error_info.MCode(), c_error_info.Msg().decode())

cdef CErrorInfo error_info_from_py(error_info: ErrorInfo):
    cdef:
        CErrorCode error_code
        CModuleCode module_code
        string msg
        CErrorInfo c_error_info = CErrorInfo()

    if error_info.error_code == ErrorCode.ERR_OK:
        return c_error_info

    error_code = error_code_from_py(error_info.error_code)
    module_code = module_code_from_py(error_info.module_code)
    msg = error_info.msg.encode()
    c_error_info = CErrorInfo(error_code, module_code, msg)
    return c_error_info

cdef check_error_info(CErrorInfo c_error_info, mesg: str):
    if c_error_info.OK():
        return
    raise RuntimeError(
        f"{mesg}, "
        f"code: {c_error_info.Code()}, "
        f"module code {c_error_info.MCode()}, "
        f"msg: {c_error_info.Msg().decode()}"
    )

cdef api_type_from_cpp(const CApiType & c_api_type):
    api_type = ApiType.Function
    if c_api_type == CApiType.FAAS:
        api_type = ApiType.Faas
    elif c_api_type == CApiType.POSIX:
        api_type = ApiType.Posix
    elif c_api_type == CApiType.SERVE:
        api_type = ApiType.Serve
    return api_type

cdef language_type_from_cpp(const CLanguageType & c_language_type):
    language_type = LanguageType.Python
    if c_language_type == CLanguageType.LANGUAGE_CPP:
        language_type = LanguageType.Cpp
    elif c_language_type == CLanguageType.LANGUAGE_PYTHON:
        language_type = LanguageType.Python
    elif c_language_type == CLanguageType.LANGUAGE_JAVA:
        language_type = LanguageType.Java
    elif c_language_type == CLanguageType.LANGUAGE_GOLANG:
        language_type = LanguageType.Golang
    elif c_language_type == CLanguageType.LANGUAGE_NODE_JS:
        language_type = LanguageType.NodeJS
    elif c_language_type == LanguageType.CSharp:
        language_type = LanguageType.CSharp
    elif c_language_type == CLanguageType.LANGUAGE_PHP:
        language_type = LanguageType.Php
    return language_type

cdef CLanguageType language_type_from_py(language_type: LanguageType):
    cdef CLanguageType c_language_type = CLanguageType.LANGUAGE_PYTHON
    if language_type == LanguageType.Cpp:
        c_language_type = CLanguageType.LANGUAGE_CPP
    elif language_type == LanguageType.Python:
        c_language_type = CLanguageType.LANGUAGE_PYTHON
    elif language_type == LanguageType.Java:
        c_language_type = CLanguageType.LANGUAGE_JAVA
    elif language_type == LanguageType.Golang:
        c_language_type = CLanguageType.LANGUAGE_GOLANG
    elif language_type == LanguageType.NodeJS:
        c_language_type = CLanguageType.LANGUAGE_NODE_JS
    elif language_type == LanguageType.CSharp:
        c_language_type = CLanguageType.LANGUAGE_CSHARP
    elif language_type == LanguageType.Php:
        c_language_type = CLanguageType.LANGUAGE_PHP
    return c_language_type

cdef CUInt64CounterData uint64_counter_data_from_py(uint64_counter_data: UInt64CounterData):
    cdef:
        CUInt64CounterData c_uint64_counter_data
    c_uint64_counter_data.name = uint64_counter_data.name
    c_uint64_counter_data.description = uint64_counter_data.description
    c_uint64_counter_data.unit = uint64_counter_data.unit
    for key, value in uint64_counter_data.labels.items():
        c_uint64_counter_data.labels.insert(pair[string, string](key, value))
    c_uint64_counter_data.value = uint64_counter_data.value
    return c_uint64_counter_data

cdef CDoubleCounterData double_counter_data_from_py(double_counter_data: DoubleCounterData):
    cdef:
        CDoubleCounterData c_double_counter_data
    c_double_counter_data.name = double_counter_data.name
    c_double_counter_data.description = double_counter_data.description
    c_double_counter_data.unit = double_counter_data.unit
    for key, value in double_counter_data.labels.items():
        c_double_counter_data.labels.insert(pair[string, string](key, value))
    c_double_counter_data.value = double_counter_data.value
    return c_double_counter_data

cdef CGaugeData gauge_data_from_py(gauge_data: GaugeData):
    cdef:
        CGaugeData c_gauge_data
    c_gauge_data.name = gauge_data.name
    c_gauge_data.description = gauge_data.description
    c_gauge_data.unit = gauge_data.unit
    for key, value in gauge_data.labels.items():
        c_gauge_data.labels.insert(pair[string, string](key, value))
    c_gauge_data.value = gauge_data.value
    return c_gauge_data

cdef CAccelerateMsgQueueHandle message_queue_handle_from_py(handle: Handle):
    cdef:
        CAccelerateMsgQueueHandle c_msg_queue_handle
    c_msg_queue_handle.worldSize = handle.n_reader
    c_msg_queue_handle.rank = handle.rank
    c_msg_queue_handle.maxChunkBytes = handle.max_chunk_bytes
    c_msg_queue_handle.maxChunks = handle.max_chunks
    c_msg_queue_handle.name = handle.name.encode()
    c_msg_queue_handle.isAsync = handle.use_async_loop
    return c_msg_queue_handle

cdef CAlarmInfo alarm_info_from_py(info: AlarmInfo):
    cdef:
        CAlarmInfo c_alarm_info
    c_alarm_info.alarmName = info.alarm_name
    if info.alarm_severity == AlarmSeverity.OFF:
        c_alarm_info.alarmSeverity = CAlarmSeverity.OFF
    elif info.alarm_severity == AlarmSeverity.INFO:
        c_alarm_info.alarmSeverity = CAlarmSeverity.INFO
    elif info.alarm_severity == AlarmSeverity.MINOR:
        c_alarm_info.alarmSeverity = CAlarmSeverity.MINOR
    elif info.alarm_severity == AlarmSeverity.MAJOR:
        c_alarm_info.alarmSeverity = CAlarmSeverity.MAJOR
    elif info.alarm_severity == AlarmSeverity.CRITICAL:
        c_alarm_info.alarmSeverity = CAlarmSeverity.CRITICAL
    else:
        c_alarm_info.alarmSeverity = CAlarmSeverity.OFF
    c_alarm_info.locationInfo = info.location_info
    c_alarm_info.cause = info.cause
    c_alarm_info.startsAt = info.starts_at
    c_alarm_info.endsAt = info.ends_at
    c_alarm_info.timeout = info.timeout
    for key, value in info.custom_options.items():
        c_alarm_info.customOptions.insert(pair[string, string](key, value))
    return c_alarm_info

cdef CSetParam set_param_from_py(set_param: SetParam):
    cdef:
        CSetParam param
    if set_param.existence == ExistenceOpt.NONE:
        param.existence = CExistenceOpt.NONE
    else:
        param.existence = CExistenceOpt.NX

    if set_param.write_mode == WriteMode.NONE_L2_CACHE:
        param.writeMode = CWriteMode.NONE_L2_CACHE
    elif set_param.write_mode == WriteMode.WRITE_THROUGH_L2_CACHE:
        param.writeMode = CWriteMode.WRITE_THROUGH_L2_CACHE
    elif set_param.write_mode == WriteMode.WRITE_BACK_L2_CACHE:
        param.writeMode = CWriteMode.WRITE_BACK_L2_CACHE
    elif set_param.write_mode == WriteMode.NONE_L2_CACHE_EVICT:
        param.writeMode = CWriteMode.NONE_L2_CACHE_EVICT
    else:
        param.writeMode = CWriteMode.NONE_L2_CACHE
    param.ttlSecond = set_param.ttl_second

    if set_param.cache_type == CacheType.MEMORY:
        param.cacheType = CCacheType.MEMORY
    elif set_param.cache_type == CacheType.DISK:
        param.cacheType = CCacheType.DISK
    else:
        param.cacheType = CCacheType.MEMORY
    return param

cdef CMSetParam m_set_param_from_py(m_set_param: MSetParam):
    cdef:
        CMSetParam param
    param.existence = CExistenceOpt.NX
    param.ttlSecond = m_set_param.ttl_second

    if m_set_param.write_mode == WriteMode.NONE_L2_CACHE:
        param.writeMode = CWriteMode.NONE_L2_CACHE
    elif m_set_param.write_mode == WriteMode.WRITE_THROUGH_L2_CACHE:
        param.writeMode = CWriteMode.WRITE_THROUGH_L2_CACHE
    elif m_set_param.write_mode == WriteMode.WRITE_BACK_L2_CACHE:
        param.writeMode = CWriteMode.WRITE_BACK_L2_CACHE
    elif m_set_param.write_mode == WriteMode.NONE_L2_CACHE_EVICT:
        param.writeMode = CWriteMode.NONE_L2_CACHE_EVICT
    else:
        param.writeMode = CWriteMode.NONE_L2_CACHE

    if m_set_param.cache_type == CacheType.MEMORY:
        param.cacheType = CCacheType.MEMORY
    elif m_set_param.cache_type == CacheType.DISK:
        param.cacheType = CCacheType.DISK
    else:
        param.cacheType = CCacheType.MEMORY
    return param

cdef CCreateParam create_param_from_py(create_param: CreateParam):
    cdef:
        CCreateParam param

    if create_param.consistency_type == ConsistencyType.PRAM:
        param.consistencyType = CConsistencyType.PRAM
    elif create_param.consistency_type == ConsistencyType.CAUSAL:
        param.consistencyType = CConsistencyType.CAUSAL
    else:
        param.consistencyType = CConsistencyType.PRAM

    if create_param.cache_type == CacheType.MEMORY:
        param.cacheType = CCacheType.MEMORY
    elif create_param.cache_type == CacheType.DISK:
        param.cacheType = CCacheType.DISK
    else:
        param.cacheType = CCacheType.MEMORY
    return param

cdef CGetParam get_param_from_py(param: GetParam):
    cdef:
        CGetParam c_get_param
    c_get_param.offset = param.offset
    c_get_param.size = param.size
    return c_get_param

cdef CGetParams get_params_from_py(params: GetParams):
    cdef:
        CGetParams c_get_params
    for get_param in params.get_params:
        c_get_param = get_param_from_py(get_param)
        c_get_params.getParams.push_back(c_get_param)
    return c_get_params

cdef CGroupOptions group_options_from_py(group_name: str, group_opts: GroupOptions):
    cdef:
        CGroupOptions c_group_options
    c_group_options.groupName = group_name.encode()
    c_group_options.timeout = group_opts.timeout
    c_group_options.sameLifecycle = group_opts.same_lifecycle
    c_group_options.strategy = group_opts.strategy
    return c_group_options

cdef CFunctionGroupOptions function_group_options_from_py(function_group_opts: FunctionGroupOptions, group_size: int):
    cdef:
        CFunctionGroupOptions c_function_group_options
    if function_group_opts.scheduling_affinity_type is None:
        c_function_group_options.bundleAffinity = CBundleAffinity.COMPACT
    elif function_group_opts.scheduling_affinity_type == SchedulingAffinityType.REQUIRED_AFFINITY_IN_EACH_BUNDLE:
        c_function_group_options.bundleAffinity = CBundleAffinity.COMPACT
    c_function_group_options.functionGroupSize = group_size
    if function_group_opts.scheduling_affinity_each_bundle_size is not None:
        c_function_group_options.bundleSize = function_group_opts.scheduling_affinity_each_bundle_size
    if function_group_opts.timeout is not None:
        c_function_group_options.timeout = function_group_opts.timeout
    return c_function_group_options

cdef CResourceGroupOptions resource_group_options_from_py(resource_group_opts: ResourceGroupOptions):
    cdef:
        CResourceGroupOptions c_resource_group_options
    c_resource_group_options.resourceGroupName = resource_group_opts.resource_group_name
    c_resource_group_options.bundleIndex = resource_group_opts.bundle_index
    return c_resource_group_options

cdef function_meta_from_py(CFunctionMeta & functionMeta, func_meta: FunctionMeta):
    cdef:
        string name
        string ns
    name = func_meta.name.encode()
    ns = func_meta.ns.encode()
    functionMeta.appName = func_meta.applicationName
    functionMeta.moduleName = func_meta.moduleName
    functionMeta.funcName = func_meta.functionName
    functionMeta.className = func_meta.className
    functionMeta.languageType = language_type_from_py(func_meta.language)
    functionMeta.codeId = func_meta.codeID
    functionMeta.signature = func_meta.signature
    functionMeta.apiType = func_meta.apiType
    functionMeta.name = name
    functionMeta.ns = ns
    functionMeta.functionId = func_meta.functionID
    functionMeta.initializerCodeId = func_meta.initializerCodeID
    functionMeta.isGenerator = func_meta.isGenerator
    functionMeta.isAsync = func_meta.isAsync
    functionMeta.tensorTransportTarget = func_meta.tensorTransportTarget
    functionMeta.enableTensorTransport = func_meta.enableTensorTransport

cdef function_meta_from_cpp(const CFunctionMeta & function):
    cdef:
        string emptyString
    emptyString = "".encode()
    func_meta = FunctionMeta(applicationName=function.appName.decode(),
                             moduleName=function.moduleName.decode(),
                             className=function.className.decode(),
                             functionName=function.funcName.decode(),
                             language=language_type_from_cpp(function.languageType),
                             codeID=function.codeId.decode(),
                             apiType=api_type_from_cpp(function.apiType),
                             signature=function.signature.decode(),
                             name=function.name.decode(),
                             ns=function.ns.decode(),
                             initializerCodeID=function.initializerCodeId.decode(),
                             isGenerator=function.isGenerator,
                             isAsync=function.isAsync,
                             tensorTransportTarget=function.tensorTransportTarget,
                             enableTensorTransport=function.enableTensorTransport)
    return func_meta

cdef function_group_running_info_from_cpp(const CFunctionGroupRunningInfo & info):
    context = yr.FunctionGroupContext()
    context.rank_id = info.instanceRankId
    context.world_size = info.worldSize
    context.device_name = info.deviceName.decode()
    for i in range(info.serverList.size()):
        server_info = yr.ServerInfo()
        server_info.server_id = info.serverList[i].serverId.decode()
        for j in range(info.serverList[i].devices.size()):
            device_info = yr.DeviceInfo()
            device_info.device_id = info.serverList[i].devices[j].deviceId
            device_info.device_ip = info.serverList[i].devices[j].deviceIp.decode()
            device_info.rank_id = info.serverList[i].devices[j].rankId
            server_info.devices.append(device_info)
        context.server_list.append(server_info)
    return context

cdef message_queue_handle_from_cpp(const CAccelerateMsgQueueHandle &c_handle):
    handle = Handle
    handle.n_reader = c_handle.worldSize
    handle.rank = c_handle.rank
    handle.max_chunk_bytes = c_handle.maxChunkBytes
    handle.max_chunks = c_handle.maxChunks
    handle.name = c_handle.name.decode()
    handle.use_async_loop = c_handle.isAsync
    return handle

cdef invoke_type_from_cpp(const CInvokeType & c_invoke_type):
    invoke_type = InvokeType.CreateInstance
    if c_invoke_type == CInvokeType.CREATE_INSTANCE:
        invoke_type = InvokeType.CreateInstance
    elif c_invoke_type == CInvokeType.INVOKE_MEMBER_FUNCTION:
        invoke_type = InvokeType.InvokeFunction
    elif c_invoke_type == CInvokeType.CREATE_NORMAL_FUNCTION_INSTANCE:
        invoke_type = InvokeType.CreateInstanceStateless
    elif c_invoke_type == CInvokeType.INVOKE_NORMAL_FUNCTION:
        invoke_type = InvokeType.InvokeFunctionStateless
    elif c_invoke_type == CInvokeType.GET_NAMED_INSTANCE_METADATA:
        invoke_type = InvokeType.GetNamedInstanceMeta
    elif c_invoke_type == CInvokeType.DELETE_REMOTE_TENSOR:
        invoke_type = InvokeType.DeleteRemoteTensor
    return invoke_type

cdef parse_rginfo_to_python(cRgInfoUnit: CResourceGroupUnit):
    """parse cpp rginfo into python"""
    cdef:
        CRgInfo c_rg_info
        string c_name
        CBundleInfo c_bundle

    python_resource_groups = {}
    for item in cRgInfoUnit.resourceGroups:
        c_name = item.first
        c_rg_info = item.second
        python_status = CommonStatus(
            code=c_rg_info.status.code,
            message=c_rg_info.status.message.decode()
        )
        python_opt = Option(
            priority=c_rg_info.opt.priority,
            groupPolicy=c_rg_info.opt.groupPolicy,
            extension={
                k.decode(): v.decode()
                for k, v in c_rg_info.opt.extension
            }
        )
        python_bundles = []
        for c_bundle in c_rg_info.bundles:
            python_resources_map = {}
            for resource_item in c_bundle.resources.resources:
                c_res = resource_item.second
                python_resource = Resource(
                    name=c_res.name.decode(),
                    type=Resource.Type(c_res.type),
                    scalar=Resource.Scalar(
                        value=c_res.scalar.value,
                        limit=c_res.scalar.limit
                    )
                )
                python_resources_map[resource_item.first.decode()] = python_resource
            python_resources = Resources(resources=python_resources_map)
            bundle_status = CommonStatus(
                code=c_bundle.status.code,
                message=c_bundle.status.message.decode()
            )
            python_kv_labels = {
                k.decode(): v.decode()
                for k, v in c_bundle.kvLabels
            }
            python_bundle = BundleInfo(
                bundleId=c_bundle.bundleId.decode(),
                rGroupName=c_bundle.rGroupName.decode(),
                parentRGroupName=c_bundle.parentRGroupName.decode(),
                functionProxyId=c_bundle.functionProxyId.decode(),
                functionAgentId=c_bundle.functionAgentId.decode(),
                tenantId=c_bundle.tenantId.decode(),
                resources=python_resources,
                labels=[label.decode() for label in c_bundle.labels],
                status=bundle_status,
                parentId=c_bundle.parentId.decode(),
                kvLabels=python_kv_labels
            )
            python_bundles.append(python_bundle)
        python_rg_info = RgInfo(
            name=c_rg_info.name.decode(),
            owner=c_rg_info.owner.decode(),
            appId=c_rg_info.appId.decode(),
            tenantId=c_rg_info.tenantId.decode(),
            bundles=python_bundles,
            status=python_status,
            parentId=c_rg_info.parentId.decode(),
            requestId=c_rg_info.requestId.decode(),
            traceId=c_rg_info.traceId.decode(),
            opt=python_opt
        )
        python_resource_groups[c_name.decode()] = python_rg_info
    return ResourceGroupUnit(resourceGroups=python_resource_groups)


def load_code_from_datasystem(code_id: str):
    """get code from datasystem"""
    cdef:
        vector[string] c_ids
        pair[CErrorInfo, vector[shared_ptr[CBuffer]]] ret
    cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
    if c_libruntime == nullptr:
        raise RuntimeError("already finalized")
    c_ids.push_back(code_id)
    _logger.debug("code id: %s", code_id)
    with nogil:
        c_libruntime.get().SetTenantIdWithPriority()
        c_libruntime.get().IncreaseReference(c_ids)
        ret = c_libruntime.get().GetBuffers(c_ids, 60000, False)
        c_libruntime.get().DecreaseReference(c_ids)
    if not ret.first.OK():
        raise RuntimeError(
            f"failed to get object, "
            f"code: {ret.first.Code()}, module code {ret.first.MCode()}, msg: {ret.first.Msg().decode()}")
    it = ret.second.begin()
    return yr.serialization.Serialization().deserialize(Buffer.make(dereference(it)))

def write_to_cbuffer(serialized_object: SerializedObject):
    # This method is for unit test suite 'test_serialization'
    cdef shared_ptr[CBuffer] c_buffer
    cdef int64_t data_size
    data_size = len(serialized_object)
    c_buffer = dynamic_pointer_cast[CBuffer, StringNativeBuffer](make_shared[StringNativeBuffer](data_size))
    buf = Buffer.make(c_buffer)
    serialized_object.write_to(buf)
    return buf

cdef shared_ptr[CBuffer] get_cbuffer(buf: Buffer):
    cdef shared_ptr[CBuffer] c_buffer
    c_buffer = buf.buffer
    return c_buffer

cdef CErrorInfo memory_copy(const shared_ptr[CBuffer] c_buffer, const char * data, uint64_t size) noexcept nogil:
    return c_buffer.get().MemoryCopy(<const void *> data, size)

cdef void memory_copy_string(char[] privateKeyPaaswd, const char * data, uint64_t size) noexcept nogil:
    if (0 < size < 100):
        for i in range(size):
            privateKeyPaaswd[i] = data[i]
    privateKeyPaaswd[size] = 0

cdef CErrorInfo write_str2buffer(const shared_ptr[CBuffer] c_buffer, const char * serialized_object,
                                 unordered_set[string] nest_ids_set) except*:
    cdef:
        CErrorInfo c_error_info

    with nogil:
        c_error_info = c_buffer.get().WriterLatch()
        if not c_error_info.OK():
            return c_error_info

    try:
        c_error_info = memory_copy(c_buffer, serialized_object, len(serialized_object))
        if not c_error_info.OK():
            return c_error_info

        with nogil:
            c_error_info = c_buffer.get().Seal(nest_ids_set)
            if not c_error_info.OK():
                return c_error_info
    finally:
        with nogil:
            c_error_info = c_buffer.get().WriterUnlatch()
            if not c_error_info.OK():
                return c_error_info
    return CErrorInfo()

cdef CErrorInfo write2buffer(const shared_ptr[CBuffer] c_buffer, SerializedObject serialized_object,
                             unordered_set[string] nest_ids_set) except*:
    cdef:
        CErrorInfo c_error_info

    with nogil:
        c_error_info = c_buffer.get().WriterLatch()
        if not c_error_info.OK():
            return c_error_info
    try:
        serialized_object.write_to(Buffer.make(c_buffer))
        with nogil:
            c_error_info = c_buffer.get().Seal(nest_ids_set)
            if not c_error_info.OK():
                return c_error_info
    finally:
        with nogil:
            c_error_info = c_buffer.get().WriterUnlatch()
            if not c_error_info.OK():
                return c_error_info

    return CErrorInfo()

cdef CErrorInfo write2DataObject(const shared_ptr[CDataObject] c_dataObj, SerializedObject serialized_object,
                                 unordered_set[string] nest_ids_set) except*:
    cdef:
        CErrorInfo c_error_info
        shared_ptr[CBuffer] c_buffer


    with nogil:
        c_buffer = c_dataObj.get().buffer
    return write2buffer(c_buffer, serialized_object, nest_ids_set)


def execute_streaming_generator_sync(generator_id, gen):
    index = 0
    try:
        for output in gen:
            notify_generator_result(generator_id, index, output, ErrorInfo())
            index = index + 1
    except Exception as e:
        err = ErrorInfo(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME,
                                     f"failed to execute user function, err: {str(e)}")
        notify_generator_result(generator_id, index, None, err)
        return
    notify_generator_finished(generator_id, index)


cdef CErrorInfo function_execute_callback_internal(const CFunctionMeta & functionMeta, const CInvokeType invokeType,
                                                   const vector[shared_ptr[CDataObject]] & rawArgs,
                                                   vector[shared_ptr[CDataObject]] & returnObjects) except*:
    cdef:
        shared_ptr[CBuffer] c_buffer
        uint64_t total_native_buffer_size = 0
        CErrorInfo c_error_info
        vector[string] nested_ids
        unordered_set[string] nested_ids_set
        shared_ptr[CInternalWaitResult] ret
        int c_index
        int64_t data_size
        int meta_size
        int nested_len
        int no_timeout = -1
        CFiberEvent genevent
    func_meta = function_meta_from_cpp(functionMeta)
    args = [Buffer.make(rawArgs.at(i).get().buffer) for i in range(rawArgs.size())]
    invoke_type = invoke_type_from_cpp(invokeType)
    return_nums = returnObjects.size()
    if ((invokeType == CInvokeType.CREATE_INSTANCE and func_meta.isAsync) or func_meta.apiType == ApiType.Serve):
        _logger.debug("start initlize _eventloop_for_default_cg")
        global _actor_is_async
        _actor_is_async = True
        global _eventloop_for_default_cg
        _eventloop_for_default_cg = create_new_event_loop()
        global _thread_for_default_cg
        _thread_for_default_cg = threading.Thread(
            target=lambda: _eventloop_for_default_cg.run_forever(),
            name="AsyncIO Thread: default")
        _thread_for_default_cg.start()
    cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
    if c_libruntime == nullptr:
        raise RuntimeError("already finalized")
    result_list = []
    error_info = ErrorInfo()
    cdef:
        CFiberEvent event
    is_async_execute = invokeType in (CInvokeType.INVOKE_MEMBER_FUNCTION, CInvokeType.CREATE_INSTANCE) or func_meta.apiType == ApiType.Serve
    if is_async_execute and _eventloop_for_default_cg is not None:
        async def function_executor():
            try:
                if func_meta.apiType == ApiType.Serve:
                    result_list, error_info = await Executor(func_meta, args, invoke_type, return_nums, _serialization_ctx, True).execute_serve()
                else:
                    result_list, error_info = Executor(func_meta, args, invoke_type, return_nums, _serialization_ctx, True).execute()
                real_result_list = []
                for index, value in enumerate(result_list):
                    if asyncio.iscoroutine(value):
                        res = await value
                        real_result_list.append(res)
                    else:
                        real_result_list.append(value)
                return real_result_list, error_info
            finally:
                event.Notify()

        future = asyncio.run_coroutine_threadsafe(function_executor(), _eventloop_for_default_cg)
        with nogil:
            (c_libruntime.get().WaitEvent(event))

        try:
            result_list, error_info = future.result()
        except concurrent.futures.CancelledError:
            error_info = ErrorInfo(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME, "exec async func err")
        except Exception as e:
            error_info = ErrorInfo(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME, "exec async func err")
            result_list = [YRInvokeError(e, traceback.format_exc())]
    else:
        result_list, error_info = Executor(func_meta, args, invoke_type, return_nums, _serialization_ctx, False).execute()
    if error_info.error_code == ErrorCode.ERR_USER_FUNCTION_EXCEPTION and result_list \
            and isinstance(result_list[0], YRInvokeError):
        serialized_object = _serialization_ctx.serialize(result_list[0])
        meta_size = constants.METALEN
        data_size = len(serialized_object) - constants.METALEN
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            c_error_info = c_libruntime.get().AllocReturnObject(
                returnObjects.at(0), meta_size, data_size, nested_ids, total_native_buffer_size)
        if not c_error_info.OK():
            return c_error_info
        c_buffer = returnObjects.at(0).get().buffer
        c_error_info = write2buffer(c_buffer, serialized_object, nested_ids_set)
        if not c_error_info.OK():
            return c_error_info
        return error_info_from_py(error_info)
    if error_info.error_code != ErrorCode.ERR_OK:
        return error_info_from_py(error_info)
    need_serialize = False
    if func_meta.apiType == ApiType.Function and \
            invoke_type in (InvokeType.InvokeFunction, InvokeType.InvokeFunctionStateless, InvokeType.GetNamedInstanceMeta, InvokeType.DeleteRemoteTensor):
        need_serialize = True
    if func_meta.isGenerator and invoke_type in (InvokeType.InvokeFunction, InvokeType.InvokeFunctionStateless):
        generator_id = returnObjects.at(0).get().id.decode()
        if len(generator_id) == 0:
            return CErrorInfo(CErrorCode.ERR_INNER_SYSTEM_ERROR, CModuleCode.RUNTIME, "generator_id should not be empty".encode())
        is_async_gen = inspect.isasyncgen(result_list[0])
        is_sync_gen = inspect.isgenerator(result_list[0])
        if (not is_async_gen and not is_sync_gen):
            print("should return a geneator, but get ", type(result_list))
            return CErrorInfo(CErrorCode.ERR_USER_FUNCTION_EXCEPTION, CModuleCode.RUNTIME, "should return a generator".encode())
        if is_sync_gen:
            execute_streaming_generator_sync(generator_id, result_list[0])
        else:
            async def async_generator():
                gen = result_list[0]
                index = 0
                try:
                    async for output in gen:
                        notify_generator_result(generator_id, index, output, ErrorInfo())
                        index = index + 1
                    notify_generator_finished(generator_id, index)
                except Exception as e:
                    err = ErrorInfo(ErrorCode.ERR_USER_FUNCTION_EXCEPTION, ModuleCode.RUNTIME,
                                     f"failed to execute user function, err: {str(e)}")
                    notify_generator_result(generator_id, index, None, err)
                finally:
                    genevent.Notify()
            future = asyncio.run_coroutine_threadsafe(async_generator(), _eventloop_for_default_cg)
            with nogil:
                (CLibruntimeManager.Instance().GetLibRuntime().get().WaitEvent(genevent))
            future.result()
        return CErrorInfo(CErrorCode.ERR_OK, CModuleCode.RUNTIME, "".encode())

    for i in range(return_nums):
        c_index = i
        nested_ids.clear()
        nested_ids_set.clear()
        total_bytes = 0
        if need_serialize:
            serialized_object = _serialization_ctx.serialize(result_list[i])
            total_bytes = len(serialized_object)
            nested_len = len(serialized_object.nested_refs)
            for nested_ref in serialized_object.nested_refs:
                nested_ids.push_back(nested_ref.id)
                nested_ids_set.insert(nested_ref.id)
            if nested_len != 0:
                with nogil:
                    ret = CLibruntimeManager.Instance().GetLibRuntime().get().Wait(nested_ids, nested_len, no_timeout)

                exception_ids = []
                exception_id = ret.get().exceptionIds.begin()
                while exception_id != ret.get().exceptionIds.end():
                    exception_ids.append(
                        f"{dereference(exception_id).first.decode()} "
                        f"code: {dereference(exception_id).second.Code()}, "
                        f"module code: {dereference(exception_id).second.MCode()} "
                        f"message: {dereference(exception_id).second.Msg().decode()}")
                    postincrement(exception_id)
                if len(exception_ids) != 0:
                    return CErrorInfo(CErrorCode.ERR_USER_FUNCTION_EXCEPTION, CModuleCode.RUNTIME, " ".join(exception_ids))
        else:
            serialized_object = result_list[i]
            if serialized_object:
                total_bytes = len(serialized_object)
        if serialized_object is None:
            continue
        meta_size = constants.METALEN
        data_size = total_bytes - constants.METALEN
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            c_error_info = c_libruntime.get().AllocReturnObject(
                returnObjects.at(c_index), meta_size, data_size, nested_ids, total_native_buffer_size)
        if not c_error_info.OK():
            return c_error_info
        c_buffer = returnObjects.at(i).get().buffer
        if need_serialize:
            c_error_info = write2buffer(c_buffer, serialized_object, nested_ids_set)
        else:
            c_error_info = write_str2buffer(c_buffer, serialized_object, nested_ids_set)
    return c_error_info


cdef CErrorInfo accelerate_execute_callback_internal(const CAccelerateMsgQueueHandle & input_handle, CAccelerateMsgQueueHandle & return_handle) except*:
    cdef CErrorInfo ret
    rpc_broadcast_mq = MessageQueue.create_from_handle(message_queue_handle_from_cpp(input_handle))
    fcc_max_chunk_bytes = int(os.environ.get("FCC_MAX_CHUNK_BYTES", 1024 * 1024 * 10))
    fcc_max_chunks = int(os.environ.get("FCC_MAX_CHUNKS", 10))
    worker_response_mq = MessageQueue(1, fcc_max_chunk_bytes, fcc_max_chunks)
    return_handle = message_queue_handle_from_py(worker_response_mq.export_handle())
    ACCELERATE_WORKER = Worker(rpc_broadcast_mq, worker_response_mq, input_handle.isAsync)
    ACCELERATE_WORKER.start()
    return ret

cdef CErrorInfo accelerate_execute_callback(const CAccelerateMsgQueueHandle & input_handle, CAccelerateMsgQueueHandle & return_handle) noexcept nogil:
    with gil:
        try:
            return accelerate_execute_callback_internal(input_handle, return_handle)
        except Exception as e:
            log.get_logger().debug(str(e))
            return CErrorInfo(CErrorCode.ERR_INNER_SYSTEM_ERROR, CModuleCode.RUNTIME, str(e))

cdef CErrorInfo function_execute_callback(const CFunctionMeta & functionMeta, const CInvokeType invokeType,
                                          const vector[shared_ptr[CDataObject]] & rawArgs,
                                          vector[shared_ptr[CDataObject]] & returnObjects) noexcept nogil:
    with gil:
        try:
            return function_execute_callback_internal(functionMeta, invokeType, rawArgs, returnObjects)
        except Exception as e:
            _logger.exception(e)
            return CErrorInfo(CErrorCode.ERR_INNER_SYSTEM_ERROR, CModuleCode.RUNTIME, str(e))

cdef CErrorInfo load_function_callback(const vector[string] & codePath) noexcept nogil:
    cdef CErrorInfo errorInfo
    with gil:
        code_paths = [codePath.at(i).decode() for i in range(codePath.size())]
        error_info = yr.code_manager.CodeManager().load_functions(code_paths)
        errorInfo = error_info_from_py(error_info)
        return errorInfo

cdef CErrorInfo shutdown_function_callback(gracePeriodSecond: uint64_t) noexcept nogil:
    with gil:
        _logger.debug("Start to execute graceful exit.")
        error_info = Executor.shutdown(gracePeriodSecond)
        _logger.info("Execution of graceful exit finished.")
        return error_info_from_py(error_info)

cdef CErrorInfo build_invoke_arg(arg, vector[CInvokeArg]& invokeArgs, string tenantId):
    cdef:
        CInvokeArg c_arg
        shared_ptr[CBuffer] data_buf
        CErrorInfo c_error_info
        cdef int64_t data_size
    c_arg.dataObj = make_shared[CDataObject]()
    if arg.serialized_obj is not None and isinstance(arg.serialized_obj, SerializedObject):
        data_buf = get_cbuffer(write_to_cbuffer(arg.serialized_obj))
    elif arg.buf is not None:
        data_buf = dynamic_pointer_cast[CBuffer, StringNativeBuffer](make_shared[StringNativeBuffer](len(arg.buf)))
        c_arg.dataObj.get().SetBuffer(data_buf)
        if len(arg.buf) != 0:
            c_error_info = memory_copy(data_buf, arg.buf, len(arg.buf))
            if not c_error_info.OK():
                raise RuntimeError(
                    f"failed to build invoke args, "
                    f"code: {c_error_info.Code()}, module code {c_error_info.MCode()}, msg: {c_error_info.Msg().decode()}")
    c_arg.dataObj.get().SetBuffer(data_buf)
    c_arg.isRef = arg.is_ref
    c_arg.objId = arg.obj_id
    c_arg.tenantId = tenantId
    for obj in arg.nested_objects:
        c_arg.nestedObjects.insert(obj)
    invokeArgs.push_back(move(c_arg))

cdef vector[CInvokeArg] parse_invoke_args(args, vector[CInvokeArg]& invokeArgs, string tenantId):
    for arg in args:
        build_invoke_arg(arg, invokeArgs, tenantId)

def get_conda_bin_executable(executable_name: str) -> str:
    conda_home = os.environ.get("YR_CONDA_HOME")
    if conda_home:
        return conda_home
    else:
        raise ValueError(
            "please configure YR_CONDA_HOME environment variable which contain a bin subdirectory"
        )

cdef parse_invoke_opts(CInvokeOptions & opts, opt: yr.InvokeOptions, group_info: GroupInfo = None):
    cdef:
        string concurrency_key = "Concurrency".encode()
        shared_ptr[CAffinity] c_affinity
    create_opt = runtime_env.parse_runtime_env(opt)
    if runtime_env.WORKING_DIR_KEY in create_opt:
        opts.workingDir = create_opt.pop(runtime_env.WORKING_DIR_KEY)
    for key, value in create_opt.items():
        opts.createOptions.insert(pair[string, string](key, value))
    opts.cpu = opt.cpu
    opts.memory = opt.memory
    opts.customExtensions.insert(pair[string, string](concurrency_key, str(opt.concurrency)))
    for key, value in opt.custom_resources.items():
        opts.customResources.insert(pair[string, float](key, value))
    for key, value in opt.custom_extensions.items():
        opts.customExtensions.insert(pair[string, string](key, value))
    for key, value in opt.pod_labels.items():
        opts.podLabels.insert(pair[string, string](key, value))
    for arg in opt.labels:
        opts.labels.push_back(arg)
    for key, value in opt.affinity.items():
        opts.affinity.insert(pair[string, string](key, value))
    for key, value in opt.alias_params.items():
        opts.aliasParams.insert(pair[string, string](key, value))
    opts.retryTimes = opt.retry_times
    opts.device = CDevice()
    opts.device.name = opt.device.name
    opts.maxInvokeLatency = opt.max_invoke_latency
    opts.minInstances = opt.min_instances
    opts.maxInstances = opt.max_instances
    opts.isDataAffinity = opt.is_data_affinity
    opts.resourceGroupOpts = resource_group_options_from_py(opt.resource_group_options)
    opts.groupName = opt.group_name.encode()
    if group_info is not None:
        opts.functionGroupOpts = function_group_options_from_py(opt.function_group_options, group_info.group_size)
        opts.groupName = group_info.group_name.encode()
    for affinity in opt.schedule_affinities:
        c_affinity = affinity_from_py_to_cpp(affinity, opt.preferred_priority, opt.required_priority,
                                             opt.preferred_anti_other_labels)
        if c_affinity == nullptr:
            raise ValueError("Failed to convert affinity to cpp affinity.")
        opts.scheduleAffinities.push_back(c_affinity)
    opts.recoverRetryTimes = opt.recover_retry_times
    opts.needOrder = opt.need_order
    opts.traceId = opt.trace_id
    for key, value in opt.env_vars.items():
        opts.envVars.insert(pair[string, string](key, value))
    opts.preemptedAllowed = opt.preempted_allowed
    opts.instancePriority = opt.instance_priority
    opts.scheduleTimeoutMs = opt.schedule_timeout_ms
    opts.isDeleteRemoteTensor = opt.is_delete_remote_tensor

cdef class Producer:
    """
    Producer interface class.

    Examples:
        >>> try:
        ...     producer_config = ProducerConfig(
        ...         delay_flush_time=5,
        ...         auto_clean_up=True,
        ...     )
        ...     producer = yr.create_stream_producer("streamName", producer_config)
        ...     # .......
        ...     data = b"hello"
        ...     element = yr.Element(data=data, id=0)
        ...     producer.send(element)
        ...     producer.flush()
        ...     producer.close()
        ... except RuntimeError as exp:
        ...     # .......
        ...     pass
    """
    cdef:
        shared_ptr[CStreamProducer] producer

    @staticmethod
    cdef make(shared_ptr[CStreamProducer] producer):
        self = Producer()
        self.producer = producer
        return self

    def __init__(self):
        """
        Initialize a Producer instance.
        """
        pass

    def send(self, element: Element, timeout_ms: int = None) -> None:
        """
        The Producer sends data, which is first placed into a buffer.
        The buffer is then flushed either according to the configured automatic flush strategy
        (after a certain interval or when the buffer is full),
        or by manually invoking the flush operation to make the data accessible to the consumer.

        Args:
            element (Element): The Element data to be sent.
            timeout_ms (int, optional): The timeout in milliseconds. Defaults to ``None``.

        Raises:
            RuntimeError: If the send operation fails.
        """
        cdef:
            CElement e = CElement(element.data, len(element.data), element.id)
            CErrorInfo ret
            int c_timeout_ms
        if timeout_ms == None:
            with nogil:
                ret = self.producer.get().Send(e)
        else:
            c_timeout_ms = timeout_ms
            with nogil:
                ret = self.producer.get().Send(e, c_timeout_ms)
        if not ret.OK():
            raise RuntimeError(
                f"failed to send, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

    def close(self) -> None:
        """
        Closing the producer will trigger an automatic flush of the data buffer and
        indicate that the data buffer is no longer in use. Once closed, the producer cannot be used again.
        """
        cdef CErrorInfo ret
        ret = self.producer.get().Close()
        if not ret.OK():
            raise RuntimeError(
                f"failed to close, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

cdef class SharedBuffer:
    cdef:
        shared_ptr[CBuffer] c_buffer
    buf: memoryview
    @staticmethod
    cdef make(shared_ptr[CBuffer] c_buffer):
        self = SharedBuffer()
        self.c_buffer = c_buffer
        self.buf = memoryview(Buffer.make(c_buffer))
        return self
    def get_buf(self) -> memoryview:
        return self.buf

    def set_buf(self, buf: memoryview):
        self.buf = buf

cdef class Consumer:
    """
    Consumer interface class.

    Examples:
        >>> try:
        ...     config = yr.SubscriptionConfig("subName", yr.SubscriptionConfig.subscriptionType.STREAM)
        ...     consumer = yr.create_stream_consumer("streamName", config)
        ...     # .......
        ...     elements = consumer.receive(6000, 1)
        ... except RuntimeError as exp:
        ...     # .......
        ...     pass
    """
    cdef:
        shared_ptr[CStreamConsumer] consumer

    @staticmethod
    cdef make(shared_ptr[CStreamConsumer] consumer):
        self = Consumer()
        self.consumer = consumer
        return self

    def receive(self, timeout_ms: int, expect_num: int = None) -> List[Element]:
        """
        Consumer receives data with subscription support.
        The call waits until either the expected number of elements `expect_num` is received
        or the timeout `timeout_ms` is reached.

        Args:
            timeout_ms (int): Timeout in milliseconds.
            expect_num (int, optional): Expected number of elements to receive.

        Raises:
            RuntimeError: If receiving data fails.

        Return:
            The actual list of received elements.
            Data Type is List[Element].
        """
        cdef:
            vector[CElement] out_elements
            CErrorInfo ret
            int c_expect_num
            int c_timeout_ms = timeout_ms
        result = []
        if expect_num:
            c_expect_num = expect_num
            with nogil:
                ret = self.consumer.get().Receive(c_expect_num, c_timeout_ms, out_elements)
        else:
            with nogil:
                ret = self.consumer.get().Receive(c_timeout_ms, out_elements)
        if not ret.OK():
            raise RuntimeError(
                f"failed to receive, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")
        it = out_elements.begin()
        while it != out_elements.end():
            result.append(Element(
                PyBytes_FromStringAndSize(<const char *> dereference(it).ptr, dereference(it).size),
                dereference(it).id))
            postincrement(it)
        return result

    def ack(self, element_id: int) -> None:
        """
        After the consumer finishes using the element identified by `element_id`,
        it must acknowledge (ack) the element. This allows all workers to determine
        whether all consumers have finished processing the data. Once a page has been
        fully consumed, internal memory reclamation may be triggered.

        Args:
            element_id (int): The ID of the element to acknowledge.

        Raises:
             RuntimeError: If acknowledging the element fails.

        Return:
            None.

        """
        cdef CErrorInfo ret
        ret = self.consumer.get().Ack(element_id)
        if not ret.OK():
            raise RuntimeError(
                f"failed to ack, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

    def close(self):
        """
        Turning off consumer will automatically trigger Ack.

        Raises:
            RuntimeError: If failed to close.
        """
        cdef CErrorInfo ret
        ret = self.consumer.get().Close()
        if not ret.OK():
            raise RuntimeError(
                f"failed to close, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

cdef CErrorInfo dump_instance(const string & checkpointId, shared_ptr[CBuffer] & data) noexcept with gil:
    cdef:
        shared_ptr[CBuffer] c_buffer
        int64_t data_size
        SerializedObject serialized_inst
    _logger.debug("Starts to serialze the instance for checkpoint, instance ID: %s", checkpointId.decode())
    try:
        serialized_inst = _serialization_ctx.serialize(InstanceManager().get_instance_package())
    except Exception as e:
        print(f'failed to dump instance {str(e)}')
        return CErrorInfo(CErrorCode.ERR_INNER_SYSTEM_ERROR, CModuleCode.RUNTIME, str(e))
    data_size = len(serialized_inst)
    c_buffer = dynamic_pointer_cast[CBuffer, NativeBuffer](make_shared[NativeBuffer](data_size))
    py_buffer = Buffer.make(c_buffer)
    serialized_inst.write_to(py_buffer)
    data.swap(c_buffer)
    _logger.debug("Succeeded to serialze the instance for checkpoint, instance ID: %s", checkpointId.decode())
    return CErrorInfo()

cdef CErrorInfo load_instance(shared_ptr[CBuffer] data) noexcept with gil:
    _logger.debug("Starts to deserialze the instance buffer for recover")
    py_buffer = Buffer.make(data)
    try:
        InstanceManager().init_from_inspackage(_serialization_ctx.deserialize(py_buffer))
        if InstanceManager().is_async and _eventloop_for_default_cg is None:
            _logger.debug("instance is async, reinitlize _eventloop_for_default_cg")
            global _actor_is_async
            _actor_is_async = True
            global _eventloop_for_default_cg
            _eventloop_for_default_cg = asyncio.new_event_loop()
            global _thread_for_default_cg
            _thread_for_default_cg = threading.Thread(
                target=lambda: _eventloop_for_default_cg.run_forever(),
                name="AsyncIO Thread: default")
            _thread_for_default_cg.start()
    except Exception as e:
        print(f'failed to load instance {str(e)}')
        return CErrorInfo(CErrorCode.ERR_INNER_SYSTEM_ERROR, CModuleCode.RUNTIME, str(e))
    _logger.info("Succeeded to deserialze the instance buffer for recover")
    return CErrorInfo()

cdef void async_callback(shared_ptr[CDataObject] data, const CErrorInfo & error, void *user_callback_ptr) noexcept with gil:
    err =  error_info_from_cpp(error)
    try:
        user_callback = <object>user_callback_ptr
        buf = Buffer.make(data.get().buffer)
        user_callback(data.get().id, err, buf)
    except Exception as e:
        print(f"failed to run async callback (user func) {e}")
    finally:
        cpython.Py_DECREF(user_callback)

cdef pair[CErrorInfo, shared_ptr[CBuffer]] handle_return_object_callback(shared_ptr[CBuffer] buffer, int rank, string & objId) noexcept with gil:
    cdef:
        CErrorInfo c_error_info
        string c_exception
        uint64_t total_native_buffer_size = 0
        shared_ptr[CBuffer] c_buffer
        unordered_set[string] nested_ids_set
    buf = memoryview(Buffer.make(buffer))
    status, (task_id, result) = decode(buf[2:])
    obj_id = task_id + "_" + str(rank)
    objId = obj_id.encode()
    if status == ResponseStatus.SUCCESS:
        serialized_object = _serialization_ctx.serialize(result)
        total_native_buffer_size = len(serialized_object)
        c_buffer = dynamic_pointer_cast[CBuffer, NativeBuffer](make_shared[NativeBuffer](total_native_buffer_size))
        c_error_info = write2buffer(c_buffer, serialized_object, nested_ids_set)
    elif status == ResponseStatus.FAILURE:
        c_exception = result.encode()
        c_error_info = CErrorInfo(CErrorCode.ERR_USER_FUNCTION_EXCEPTION, CModuleCode.RUNTIME, c_exception)
    return pair[CErrorInfo, shared_ptr[CBuffer]](c_error_info, c_buffer)


cdef CErrorInfo check_signals() noexcept nogil:
    cdef:
        string errmsg
    with gil:
        try:
            PyErr_CheckSignals()
        except KeyboardInterrupt:
            errmsg = "recieve teminal kill signal".encode()
            return CErrorInfo(CErrorCode.ERR_CLIENT_TERMINAL_KILLED, CModuleCode.RUNTIME, errmsg)
    return CErrorInfo()

cdef class Fnruntime:
    def __cinit__(self, functionSystemIpAddr, functionSystemPort, functionSystemRtServerIpAddr,
                  functionSystemRtServerPort):
        cdef CLibruntimeConfig config = CLibruntimeConfig()
        cdef CErrorInfo ret
        config.libruntimeOptions.functionExecuteCallback = function_execute_callback
        config.libruntimeOptions.loadFunctionCallback = load_function_callback
        config.libruntimeOptions.shutdownCallback = shutdown_function_callback
        config.libruntimeOptions.checkpointCallback = dump_instance
        config.libruntimeOptions.recoverCallback = load_instance
        config.libruntimeOptions.accelerateCallback = accelerate_execute_callback
        config.functionSystemIpAddr = functionSystemIpAddr
        config.functionSystemPort = functionSystemPort
        config.functionSystemRtServerIpAddr = functionSystemRtServerIpAddr
        config.functionSystemRtServerPort = functionSystemRtServerPort
        if ConfigManager().in_cluster:
            config.dataSystemIpAddr = ConfigManager().ds_address.split(":")[0]
            config.dataSystemPort = int(ConfigManager().ds_address.split(":")[1])
        config.metaConfig = ConfigManager().meta_config.to_json()
        config.isDriver = ConfigManager().is_driver
        config.jobId = ConfigManager().job_id
        config.runtimeId = ConfigManager().runtime_id
        config.functionIds.insert(pair[CLanguageType, string](CLanguageType.LANGUAGE_PYTHON,
                                                              ConfigManager().meta_config.functionID.python))
        config.functionIds.insert(pair[CLanguageType, string](CLanguageType.LANGUAGE_CPP,
                                                              ConfigManager().meta_config.functionID.cpp))
        config.functionIds.insert(pair[CLanguageType, string](CLanguageType.LANGUAGE_JAVA,
                                                              ConfigManager().meta_config.functionID.java))
        config.logLevel = ConfigManager().log_level
        config.logDir = ConfigManager().log_dir
        config.logFileSizeMax = ConfigManager().log_file_size_max
        config.logFileNumMax = ConfigManager().log_file_num_max
        config.logFlushInterval = ConfigManager().log_flush_interval
        config.recycleTime = ConfigManager().meta_config.recycleTime
        config.maxTaskInstanceNum = ConfigManager().meta_config.maxTaskInstanceNum
        config.maxConcurrencyCreateNum = ConfigManager().meta_config.maxConcurrencyCreateNum
        config.enableMetrics = ConfigManager().meta_config.enableMetrics
        config.functionMasters = ConfigManager().meta_config.functionMasters
        config.selfLanguage = CLanguageType.LANGUAGE_PYTHON
        config.loadPaths = ConfigManager().load_paths
        config.rpcTimeout = ConfigManager().rpc_timeout
        config.tenantId = ConfigManager().tenant_id
        config.enableMTLS = ConfigManager().enable_mtls
        config.privateKeyPath = ConfigManager().private_key_path
        config.certificateFilePath = ConfigManager().certificate_file_path
        config.verifyFilePath = ConfigManager().verify_file_path
        memory_copy_string(config.privateKeyPaaswd, ConfigManager().private_key_paaswd,
                           len(ConfigManager().private_key_paaswd))
        config.httpIocThreadsNum = ConfigManager().http_ioc_threads_num
        config.serverName = ConfigManager().server_name
        config.inCluster = ConfigManager().in_cluster
        config.ns = ConfigManager().ns
        config.checkSignals_ = check_signals
        config.workingDir = ConfigManager().working_dir
        config.runtimePublicKeyPath = ConfigManager().runtime_public_key_path
        config.runtimePrivateKeyPath = ConfigManager().runtime_private_key_path
        config.dsPublicKeyPath = ConfigManager().ds_public_key_path
        config.encryptEnable = ConfigManager().enable_ds_encrypt
        config.ns = ConfigManager().ns
        config.logToDriver = ConfigManager().log_to_driver
        config.dedupLogs = ConfigManager().dedup_logs
        for key, value in ConfigManager().custom_envs.items():
            config.customEnvs.insert(pair[string, string](key, value))
        with nogil:
            ret = CLibruntimeManager.Instance().Init(config)
        if not ret.OK():
            raise ValueError(
                f"failed to init, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

    def init(self, ctx):
        global _serialization_ctx
        _serialization_ctx = ctx
        yr.code_manager.CodeManager().register_load_code_from_datasystem_func(load_code_from_datasystem)

    def receive_request_loop(self):
        with nogil:
            CLibruntimeManager.Instance().ReceiveRequestLoop()

    def finalize(self):
        with nogil:
            CLibruntimeManager.Instance().Finalize()

    def put(self, serialized_object: SerializedObject, create_param: CreateParam):
        """
        Put data to ds
        :param data: serialized object data
        :param nested_object_ids: nested object in data
        :param create_param create param to datasystem
        :return: object id
        """
        cdef:
            CCreateParam param
            shared_ptr[CDataObject] c_dataObj
            vector[string] nested_ids
            unordered_set[string] nested_ids_set
            pair[CErrorInfo, string] ret
            int64_t data_size
            int64_t meta_size

        param = create_param_from_py(create_param)
        for nested_ref in serialized_object.nested_refs:
            nested_ids.push_back(nested_ref.id)
            nested_ids_set.insert(nested_ref.id)
        meta_size = constants.METALEN
        data_size = len(serialized_object) - constants.METALEN
        c_dataObj = make_shared[CDataObject]()
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            ret = c_libruntime.get().CreateDataObject(meta_size, data_size, c_dataObj,
                                                                                       nested_ids, param)

        if not ret.first.OK():
            raise RuntimeError(
                f"failed to put object, "
                f"code: {ret.first.Code()}, module code {ret.first.MCode()}, msg: {ret.first.Msg().decode()}")

        c_error_info = write2DataObject(c_dataObj, serialized_object, nested_ids_set)

        if not c_error_info.OK():
            raise RuntimeError(
                f"failed to put object, "
                f"code: {c_error_info.Code()}, module code {c_error_info.MCode()}, msg: {c_error_info.Msg().decode()}")
        return ret.second.decode("utf-8")

    def get(self, ids: List[str], timeout_ms: int, allow_partial: bool) -> List[Buffer]:
        """
        Get data from ds
        :param ids: ids which need to get
        :param timeout_ms: timeout
        :param allow_partial: allow partial
        :return: data which get from ds
        """
        cdef:
            vector[string] c_ids
            pair[CErrorInfo, vector[shared_ptr[CDataObject]]] ret
            int c_timeout_ms = timeout_ms
            shared_ptr[CDataObject] c_dataObj
            shared_ptr[CBuffer] c_buffer
            vector[CStackTraceInfo] c_stack_trace_info
        for i in ids:
            c_ids.push_back(i)
        c_dataObj = make_shared[CDataObject]()
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            ret = c_libruntime.get().GetDataObjects(c_ids, c_timeout_ms, allow_partial)
        result = []
        if not ret.first.OK():
            if error_code_from_cpp(ret.first.Code()) == ErrorCode.ERR_GENERATOR_FINISHED:
                raise GeneratorEndError(
                    f"gernerator stop, "
                    f"code: {ret.first.Code()}, module code {ret.first.MCode()}, msg: {ret.first.Msg().decode()}")
            if ret.first.Code() == CErrorCode.ERR_USER_FUNCTION_EXCEPTION or \
                    ret.first.Code() == CErrorCode.ERR_DEPENDENCY_FAILED:
                c_stack_trace_info = ret.first.GetStackTraceInfos()
                it_sti = c_stack_trace_info.begin()
                while it_sti != c_stack_trace_info.end():
                    if dereference(it_sti).Type().decode() == "YRInvokeError":
                        c_buffer = dynamic_pointer_cast[CBuffer, NativeBuffer](
                            make_shared[NativeBuffer](dereference(it_sti).Message().size()))
                        memory_copy(c_buffer, dereference(it_sti).Message().data(),
                                    dereference(it_sti).Message().size())
                        result.append(Buffer.make(c_buffer))
                        return result
                    postincrement(it_sti)
            mesg = ("failed to get object, "
                    f"code: {ret.first.Code()}, module code: {ret.first.MCode()}, message: {ret.first.Msg().decode()}")
            if ret.first.IsTimeout():
                raise TimeoutError(mesg)
            raise RuntimeError(mesg)

        result = []
        for it in ret.second:
            if it == NULL or dereference(it).buffer == NULL:
                result.append(None)
            else:
                result.append(Buffer.make(dereference(it).buffer))
        return result

    def wait(self, objs: List[str], wait_num: int, timeout: int) -> Tuple[List[str], List[str], List[str]]:
        """
        wait objects
        :param objs: objs which need to wait
        :param wait_num: wait number
        :param timeout: timeout
        :return: ready objects, unready objects, exception of objects
        """
        cdef:
            vector[string] c_ids
            shared_ptr[CInternalWaitResult] ret
            int cwaitNum = wait_num
            int ctimeout = timeout
        for i in objs:
            c_ids.push_back(i)
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            ret = c_libruntime.get().Wait(c_ids, cwaitNum, ctimeout)
        ready_ids = []
        ready_id = ret.get().readyIds.begin()
        while ready_id != ret.get().readyIds.end():
            ready_ids.append(dereference(ready_id).decode())
            postincrement(ready_id)
        unready_ids = []
        unready_id = ret.get().unreadyIds.begin()
        while unready_id != ret.get().unreadyIds.end():
            unready_ids.append(dereference(unready_id).decode())
            postincrement(unready_id)
        exception_ids = []
        exception_id = ret.get().exceptionIds.begin()
        while exception_id != ret.get().exceptionIds.end():
            exception_ids.append(f"{dereference(exception_id).first.decode()}")
            postincrement(exception_id)
        if len(exception_ids) != 0:
            _logger.debug("Gets exception IDs while waiting: %s", exception_ids)
        return ready_ids, unready_ids, exception_ids

    def get_async(self, object_id: str, callback: Callable):
        """
        get async
        :param object_id: object id
        :param callback: user callback
        :return: None
        """
        cpython.Py_INCREF(callback)
        cdef:
            c_object_id = object_id.encode()

        CLibruntimeManager.Instance().GetLibRuntime().get().GetAsync(c_object_id, async_callback, <void*>callback)

    def kv_write(self, key: str, value: bytes, set_param: SetParam) -> None:
        """
        store value to ds
        :param key: key
        :param value: value
        :param set_param: set param
        :return: None
        """
        cdef:
            CSetParam param
            shared_ptr[CBuffer] c_buffer
            CErrorInfo ret
            string c_str = key.encode()

        param = set_param_from_py(set_param)
        c_buffer = dynamic_pointer_cast[CBuffer, NativeBuffer](make_shared[NativeBuffer](len(value)))
        ret = memory_copy(c_buffer, value, len(value))
        if not ret.OK():
            raise RuntimeError(
                f"failed to kv_write, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            ret = c_libruntime.get().KVWrite(c_str, c_buffer, param)
        if not ret.OK():
            raise RuntimeError(
                f"failed to kv_write, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

    def kv_m_write_tx(self, keys: List[str], values: List[bytes], m_set_param: MSetParam) -> None:
        """
        store multiple key-value pairs to ds
        :param keys: the keys to set
        :param values: the values to set. Size of values should equal to size of keys.
        :param m_set_param: MSetParams for datasystem stateClient.
        :return: None
        """
        cdef:
            CMSetParam param
            vector[string] c_ids
            vector[shared_ptr[CBuffer]] c_buffers
            CErrorInfo ret

        param = m_set_param_from_py(m_set_param)
        for k_id in keys:
            c_ids.push_back(k_id)

        for value in values:
            c_buffers.push_back(dynamic_pointer_cast[CBuffer, NativeBuffer](make_shared[NativeBuffer](len(value))))
            ret = memory_copy(c_buffers.back(), value, len(value))
            if not ret.OK():
                raise RuntimeError(
                    f"failed to kv_m_write_tx, "
                    f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            ret = c_libruntime.get().KVMSetTx(c_ids, c_buffers, param)
        if not ret.OK():
            raise RuntimeError(
                f"failed to kv_m_write_tx, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

    def kv_read(self, key: Union[str, List[str]], timeout_ms: int, allow_partial: bool = False) -> List[memoryview]:
        """
        read value from ds
        :param key: key
        :param timeout_ms: timeout
        :param allow_partial: allow_partial
        :return: value
        """
        cdef:
            vector[string] c_ids
            int c_timeout_ms = timeout_ms
            CMultipleReadResult ret
        if isinstance(key, str):
            c_ids.push_back(key)
        else:
            for k_id in key:
                c_ids.push_back(k_id)
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            ret = c_libruntime.get().KVRead(c_ids, c_timeout_ms, allow_partial)
        if not ret.second.OK():
            raise RuntimeError(
                f"failed to kv_read, "
                f"code: {ret.second.Code()}, module code {ret.second.MCode()}, msg: {ret.second.Msg().decode()}")
        it = ret.first.begin()
        result = []
        while it != ret.first.end():
            result.append(memoryview(Buffer.make(dereference(it))))
            postincrement(it)
        return result

    def kv_get_with_param(self, keys: List[str], get_params: GetParams, timeout_ms: int) -> List[memoryview]:
        """
        read value from ds
        :param key: key
        :param get_params: get params for datasystem
        :param timeout_ms: timeout
        :return: value
        """
        cdef:
            vector[string] c_ids
            CGetParams c_get_params
            int c_timeout_ms = timeout_ms
            CMultipleReadResult ret
        for k_id in keys:
            c_ids.push_back(k_id)
        c_get_params = get_params_from_py(get_params)
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            ret = c_libruntime.get().KVGetWithParam(c_ids, c_get_params, c_timeout_ms)
        if not ret.second.OK():
            raise RuntimeError(
                f"failed to kv_get_with_param, "
                f"code: {ret.second.Code()}, module code {ret.second.MCode()}, msg: {ret.second.Msg().decode()}")
        result = []
        for it in ret.first:
            if it == NULL:
                result.append(None)
            else:
                result.append(memoryview(Buffer.make(it)))
        return result

    def kv_del(self, key: Union[str, List[str]]) -> None:
        """
        del value in ds
        :param key: key
        :return: None
        """
        cdef:
            CMultipleDelResult ret
            vector[string] c_ids
        if isinstance(key, str):
            c_ids.push_back(key)
        else:
            for k_id in key:
                c_ids.push_back(k_id)
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            ret = c_libruntime.get().KVDel(c_ids)
        if not ret.second.OK():
            raise RuntimeError(
                f"failed to kv_del, "
                f"code: {ret.second.Code()}, module code {ret.second.MCode()}, msg: {ret.second.Msg().decode()}")

    def increase_global_reference(self, object_ids: List[str]) -> None:
        """
        increase global reference for object
        :param object_ids: objects
        :return: None
        """
        cdef:
            vector[string] c_ids
            CErrorInfo ret
        for i in object_ids:
            c_ids.push_back(i)
        with nogil:
            CLibruntimeManager.Instance().GetLibRuntime().get().SetTenantIdWithPriority()
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().IncreaseReference(c_ids)
        if not ret.OK():
            raise RuntimeError(
                f"failed to increase, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

    def decrease_global_reference(self, object_ids: List[str]) -> None:
        """
        decrease global reference for object
        :param object_ids: objects
        :return: None
        """
        cdef vector[string] c_ids
        for i in object_ids:
            c_ids.push_back(i)
        with nogil:
            runtime = CLibruntimeManager.Instance().GetLibRuntime().get()
            if runtime != nullptr:
                runtime.SetTenantIdWithPriority()
                runtime.DecreaseReference(c_ids)

    def invoke_by_name(self, func_meta: FunctionMeta, args: List, opt,
                       return_nums: int, group_info: GroupInfo = None) -> List[str]:
        """
        invoke instance by name
        :param func_meta: function meta
        :param args: args
        :param opt: options
        :param return_objs: return objects
        :param group_info: group info
        :return: return_objs: List[str]
        """
        cdef:
            CErrorInfo error_info
            CFunctionMeta functionMeta
            vector[CInvokeArg] invokeArgs
            CInvokeOptions opts
            vector[CDataObject] returnObjs
            CDataObject returnObj

        # TODO add fundamental type
        for _ in range(return_nums):
            returnObj = CDataObject()
            returnObj.id = "".encode()
            returnObjs.push_back(returnObj)
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            tenantId = c_libruntime.get().GetTenantId()
        function_meta_from_py(functionMeta, func_meta)
        parse_invoke_args(args, invokeArgs, tenantId)
        parse_invoke_opts(opts, opt, group_info)
        with nogil:
            libruntime = c_libruntime.get()
            libruntime.SetTenantIdWithPriority()
            error_info = libruntime.InvokeByFunctionName(functionMeta, invokeArgs, opts, returnObjs)
        if not error_info.OK():
            raise ValueError(
                f"failed to invoke instance, "
                f"code: {error_info.Code()}, module code {error_info.MCode()}, msg: {error_info.Msg().decode()}")

        _logger.debug("Succeed to invoke function: %s", func_meta)
        return_objs = []
        for i in range(returnObjs.size()):
            return_objs.append(returnObjs.at(i).id.decode())
        return return_objs

    def create_instance(self, func_meta: FunctionMeta, args: List, opt, group_info: GroupInfo = None) -> str:
        """
        create instance
        :param func_meta: function_meta
        :param args: args
        :param opt: options
        :param group_info: group info
        :return: instance id
        """
        cdef:
            pair[CErrorInfo, string] ret
            CFunctionMeta functionMeta
            vector[CInvokeArg] invokeArgs
            CInvokeOptions opts
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            tenantId = c_libruntime.get().GetTenantId()
        function_meta_from_py(functionMeta, func_meta)
        parse_invoke_args(args, invokeArgs, tenantId)
        parse_invoke_opts(opts, opt, group_info)
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            ret = c_libruntime.get().CreateInstance(functionMeta, invokeArgs, opts)
        if not ret.first.OK():
            raise ValueError(
                f"failed to create instance, "
                f"code: {ret.first.Code()}, module code {ret.first.MCode()}, msg: {ret.first.Msg().decode()}")
        return ret.second.decode()

    def invoke_instance(self, func_meta: FunctionMeta, instance_id: str, args: List,
                        opt, return_nums: int) -> List[str]:
        """
        invoke instance by id
        :param func_meta: function meta
        :param instance_id: instance id
        :param args: args
        :param opt: options
        :return: object ids
        """
        cdef:
            CErrorInfo ret
            CFunctionMeta functionMeta
            vector[CInvokeArg] invokeArgs
            CInvokeOptions opts
            vector[CDataObject] returnObjs
            CDataObject returnObj
            string cinstanceID = instance_id.encode()
        # TODO add fundamental type
        for _ in range(return_nums):
            returnObj = CDataObject()
            returnObj.id = "".encode()
            returnObjs.push_back(returnObj)
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            tenantId = c_libruntime.get().GetTenantId()
        function_meta_from_py(functionMeta, func_meta)
        parse_invoke_args(args, invokeArgs, tenantId)
        parse_invoke_opts(opts, opt)
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            ret = c_libruntime.get().InvokeByInstanceId(functionMeta, cinstanceID,
                                                                                         invokeArgs, opts, returnObjs)
        if not ret.OK():
            raise ValueError(
                f"failed to invoke instance, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")
        return_objs = []
        for i in range(returnObjs.size()):
            return_objs.append(returnObjs.at(i).id.decode())
        return return_objs

    def cancel(self, object_ids: List[str], allow_force: bool, allow_recursive: bool) -> None:
        """
        cancel task
        :param object_id: object id which need to cancel
        :return: None
        """
        cdef vector[string] c_ids
        for i in object_ids:
            c_ids.push_back(i)
        with nogil:
            CLibruntimeManager.Instance().GetLibRuntime().get().Cancel(c_ids, allow_force, allow_recursive)

    def terminate_instance(self, instance_id: str) -> None:
        """
        terminate instance
        :param instance_id: instance id which need to terminate
        :return: None
        """
        cdef:
            CErrorInfo ret
            string c_id = instance_id
            int sigNo = CSignal.KILLINSTANCE
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            ret = c_libruntime.get().Kill(c_id, sigNo)
        if not ret.OK():
            raise RuntimeError(
                f"failed to kill instance, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

    def terminate_instance_sync(self, instance_id: str) -> None:
        """
        terminate instance sync
        :param instance_id: instance id which need to terminate
        :return: None
        """
        cdef:
            CErrorInfo ret
            string c_id = instance_id
            int sigNo = CSignal.KILLINSTANCESYNC
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            ret = c_libruntime.get().Kill(c_id, sigNo)
        if not ret.OK():
            raise RuntimeError(
                f"failed to kill instance sync, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

    def terminate_group(self, group_name: str) -> None:
        """
        terminate group
        :param group_name: group which need to terminate
        :return: None
        """
        cdef:
            string c_group_name = group_name
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_libruntime.get().GroupTerminate(c_group_name)

    def exit(self) -> None:
        """
        exit current instance. only support in runtime.
        :return: None
        """
        with nogil:
            CLibruntimeManager.Instance().GetLibRuntime().get().Exit()

    def create_stream_producer(self, stream_name: str, config: ProducerConfig) -> Producer:
        """
        create stream producer
        :param stream_name: stream name
        :param config: ProducerConfig
        :return: producer
        """
        cdef:
            CProducerConf cfg
            shared_ptr[CStreamProducer] producer
            CErrorInfo ret
            string cstreamName = stream_name.encode()
        cfg.delayFlushTime = config.delay_flush_time
        cfg.pageSize = config.page_size
        cfg.maxStreamSize = config.max_stream_size
        cfg.autoCleanup = config.auto_clean_up
        cfg.encryptStream = config.encrypt_stream
        cfg.retainForNumConsumers = config.retain_for_num_consumers
        cfg.reserveSize = config.reserve_size
        for key, value in config.extend_config.items():
            cfg.extendConfig.insert(pair[string, string](key, value))
        with nogil:
            CLibruntimeManager.Instance().GetLibRuntime().get().SetTenantIdWithPriority()
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().CreateStreamProducer(cstreamName, cfg, producer)
        if not ret.OK():
            raise RuntimeError(
                f"failed to create stream producer, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")
        return Producer.make(producer)

    def create_stream_consumer(self, stream_name: str, config: SubscriptionConfig) -> Consumer:
        """
        create stream consumer
        :param stream_name: stream name
        :param config: SubscriptionConfig
        :return: consumer
        """
        cdef:
            CSubscriptionConfig cfg
            shared_ptr[CStreamConsumer] consumer
            CErrorInfo ret
            string cstreamName = stream_name.encode()
        cfg.subscriptionName = config.subscription_name
        if config.subscriptionType == SubscriptionType.STREAM:
            cfg.subscriptionType = CSubscriptionType.STREAM
        elif config.subscriptionType == SubscriptionType.ROUND_ROBIN:
            cfg.subscriptionType = CSubscriptionType.ROUND_ROBIN
        elif config.subscriptionType == SubscriptionType.KEY_PARTITIONS:
            cfg.subscriptionType = CSubscriptionType.KEY_PARTITIONS
        else:
            cfg.subscriptionType = CSubscriptionType.UNKNOWN
        for key, value in config.extend_config.items():
            cfg.extendConfig.insert(pair[string, string](key, value))
        with nogil:
            CLibruntimeManager.Instance().GetLibRuntime().get().SetTenantIdWithPriority()
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().CreateStreamConsumer(cstreamName, cfg, consumer)
        if not ret.OK():
            raise RuntimeError(
                f"failed to create stream consumer, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")
        return Consumer.make(consumer)

    def delete_stream(self, stream_name: str) -> None:
        """
        delete stream
        :param stream_name: stream name
        :return: None
        """
        cdef:
            CErrorInfo ret
            string cstreamName = stream_name.encode()
        with nogil:
            CLibruntimeManager.Instance().GetLibRuntime().get().SetTenantIdWithPriority()
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().DeleteStream(cstreamName)
        if not ret.OK():
            raise RuntimeError(
                f"failed to delete stream, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

    def query_global_producers_num(self, stream_name: str) -> int:
        """
        query global producers num
        :param stream_name: stream name
        :return: producers num
        """
        cdef:
            uint64_t num = 0
            CErrorInfo ret
            string cstreamName = stream_name.encode()
        with nogil:
            CLibruntimeManager.Instance().GetLibRuntime().get().SetTenantIdWithPriority()
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().QueryGlobalProducersNum(cstreamName, num)
        if not ret.OK():
            raise RuntimeError(
                f"failed to query global producers num, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")
        return num

    def query_global_consumers_num(self, stream_name: str) -> int:
        """
        query global consumers num
        :param stream_name: stream name
        :return: consumers num
        """
        cdef:
            uint64_t num = 0
            CErrorInfo ret
            string c_stream_name = stream_name.encode()
        with nogil:
            CLibruntimeManager.Instance().GetLibRuntime().get().SetTenantIdWithPriority()
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().QueryGlobalConsumersNum(c_stream_name, num)
        if not ret.OK():
            raise RuntimeError(
                f"failed to query global consumers num, "
                f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")
        return num

    def save_real_instance_id(self, instance_id: str, need_order: bool) -> None:
        """
        save real instance id
        :param instance_id: instance id
        :param need_order: need order
        :return: None
        """
        cdef:
            string c_instance_id = instance_id.encode()
            CInstanceOptions opts
        opts.needOrder = need_order
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_libruntime.get().SaveRealInstanceId(c_instance_id, c_instance_id, opts)

    def get_real_instance_id(self, instance_id: str) -> str:
        """
        get real instance id
        :param instance_id: instance key for instance id
        :return: real instance id
        """
        cdef:
            string c_real_instance_id
            string c_instance_id = instance_id.encode()
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_real_instance_id = c_libruntime.get().GetRealInstanceId(c_instance_id)

        real_instance_id = c_real_instance_id.decode()
        return real_instance_id

    def is_object_existing_in_local(self, object_id: str) -> bool:
        """
        is object existed in local
        :param object_id: object id
        :return: is existed
        """
        cdef:
            bool ret
            string c_object_id = object_id.encode()
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            ret = c_libruntime.get().IsObjectExistingInLocal(c_object_id)

        return ret

    def save_state(self, timeout_ms: int):
        """
        save actor state
        :param timeout_sec: timeout milliseconds
        :return: None
        """
        cdef:
            int c_timeout_ms = timeout_ms;
            shared_ptr[CBuffer] data
            CErrorInfo c_errror_info

        c_errror_info = dump_instance("".encode(), data)
        check_error_info(c_errror_info, "Failed to save state")
        if data == nullptr:
            raise RuntimeError("Failed to save state. Instance buffer is nullptr.")
        with nogil:
            c_errror_info = CLibruntimeManager.Instance().GetLibRuntime().get().SaveState(data, c_timeout_ms)
        check_error_info(c_errror_info, "Failed to save state")

    def load_state(self, timeout_ms: int):
        """
        load actor state
        :param timeout_ms: timeout milliseconds
        :return: None
        """
        cdef:
            int c_timeout_ms = timeout_ms;
            shared_ptr[CBuffer] data;
            CErrorInfo c_errror_info
        with nogil:
            c_errror_info = CLibruntimeManager.Instance().GetLibRuntime().get().LoadState(data, c_timeout_ms)
        check_error_info(c_errror_info, "Failed to load state")
        c_errror_info = load_instance(data)
        check_error_info(c_errror_info, "Failed to load state")

    def create_resource_group(self, vector[unordered_map[string, double]] bundles, string name, string strategy) -> str:
        """
        create resource group
        :param bundles: The bundles in resource group.
        :name: The name of resource group
        :return: Request id
        """
        cdef:
            CResourceGroupSpec c_resource_group_spec
            CErrorInfo c_errror_info
            string c_request_id
        c_resource_group_spec.name = name
        c_resource_group_spec.strategy = strategy
        c_resource_group_spec.bundles = bundles
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_errror_info = c_libruntime.get().CreateResourceGroup(c_resource_group_spec,
                                                                                                     c_request_id)
        check_error_info(c_errror_info, "Failed to create resource group")
        return c_request_id.decode()

    def remove_resource_group(self, name: str):
        """
        remove resource group
        :name: The name of resource group
        :return: None
        """
        cdef:
            string c_resource_group_name = name.encode()
            CErrorInfo c_errror_info
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_errror_info = c_libruntime.get().RemoveResourceGroup(c_resource_group_name)
        check_error_info(c_errror_info, "Failed to remove resource group")

    def wait_resource_group(self, name: str, request_id: str, timeout_seconds: int):
        """
        wait resource group
        :name: The name of resource group
        :request_id: The request id of resource group
        :timeout_seconds: Timeout for waiting create resource group response.
        :return: None
        """
        cdef:
            string c_resource_group_name = name.encode()
            string c_request_id = request_id.encode()
            int c_timeout_seconds
            CErrorInfo c_errror_info
        c_timeout_seconds = timeout_seconds
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_errror_info = c_libruntime.get().WaitResourceGroup(c_resource_group_name,
                                                                                                   c_request_id,
                                                                                                   c_timeout_seconds)
        check_error_info(c_errror_info, "Failed to wait resource group")

    def set_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        set uint64 counter metric
        :param data: UInt64CounterData
        :return: None
        """
        cdef:
            CErrorInfo error_info
            CUInt64CounterData uint64_counter_data
        uint64_counter_data = uint64_counter_data_from_py(data)
        with nogil:
            error_info = CLibruntimeManager.Instance().GetLibRuntime().get().SetUInt64Counter(uint64_counter_data)
        check_error_info(error_info, "Failed to set uint64 counter")

    def reset_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        reset uint64 counter metric
        :param data: UInt64CounterData
        :return: None
        """
        cdef:
            CErrorInfo error_info
            CUInt64CounterData uint64_counter_data
        uint64_counter_data = uint64_counter_data_from_py(data)
        with nogil:
            error_info = CLibruntimeManager.Instance().GetLibRuntime().get().ResetUInt64Counter(uint64_counter_data)
        check_error_info(error_info, "Failed to set uint64 counter")

    def increase_uint64_counter(self, data: UInt64CounterData) -> None:
        """
        increase uint64 counter metric
        :param data: UInt64CounterData
        :return: None
        """
        cdef:
            CErrorInfo error_info
            CUInt64CounterData uint64_counter_data
        uint64_counter_data = uint64_counter_data_from_py(data)
        with nogil:
            error_info = CLibruntimeManager.Instance().GetLibRuntime().get().IncreaseUInt64Counter(uint64_counter_data)
        check_error_info(error_info, "Failed to set uint64 counter")

    def get_value_uint64_counter(self, data: UInt64CounterData) -> int:
        """
        get value of uint64 counter metric
        :param data: UInt64CounterData
        :return: value
        """
        cdef:
            pair[CErrorInfo, int] ret
            CUInt64CounterData uint64_counter_data
        uint64_counter_data = uint64_counter_data_from_py(data)
        with nogil:
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().GetValueUInt64Counter(uint64_counter_data)
        check_error_info(ret.first, "Failed to set uint64 counter")
        return ret.second

    def set_double_counter(self, data: DoubleCounterData) -> None:
        """
        set double counter
        :param data: DoubleCounterData
        :return: None
        """
        cdef:
            CErrorInfo error_info
            CDoubleCounterData double_counter_data
        double_counter_data = double_counter_data_from_py(data)
        with nogil:
            error_info = CLibruntimeManager.Instance().GetLibRuntime().get().SetDoubleCounter(double_counter_data)
        check_error_info(error_info, "Failed to set double counter")

    def reset_double_counter(self, data: DoubleCounterData) -> None:
        """
        reset double counter
        :param data: DoubleCounterData
        :return: None
        """
        cdef:
            CErrorInfo error_info
            CDoubleCounterData double_counter_data
        double_counter_data = double_counter_data_from_py(data)
        with nogil:
            error_info = CLibruntimeManager.Instance().GetLibRuntime().get().ResetDoubleCounter(double_counter_data)
        check_error_info(error_info, "Failed to set double counter")

    def increase_double_counter(self, data: DoubleCounterData) -> None:
        """
        increase double counter
        :param data: DoubleCounterData
        :return: None
        """
        cdef:
            CErrorInfo error_info
            CDoubleCounterData double_counter_data
        double_counter_data = double_counter_data_from_py(data)
        with nogil:
            error_info = CLibruntimeManager.Instance().GetLibRuntime().get().IncreaseDoubleCounter(double_counter_data)
        check_error_info(error_info, "Failed to set double counter")

    def get_value_double_counter(self, data: DoubleCounterData) -> float:
        """
        get value of double counter
        :param data: DoubleCounterData
        :return: value
        """
        cdef:
            pair[CErrorInfo, float] ret
            CDoubleCounterData double_counter_data
        double_counter_data = double_counter_data_from_py(data)
        with nogil:
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().GetValueDoubleCounter(double_counter_data)
        check_error_info(ret.first, "Failed to set double counter")
        return ret.second

    def report_gauge(self, data: GaugeData) -> None:
        """
        report gauge metric
        :param data: GaugeData
        :return: None
        """
        cdef:
            CErrorInfo error_info
            CGaugeData gauge_data
        gauge_data = gauge_data_from_py(data)
        with nogil:
            error_info = CLibruntimeManager.Instance().GetLibRuntime().get().ReportGauge(gauge_data)
        check_error_info(error_info, "Failed to report gauge")

    def set_alarm(self, name: str, description: str, info: AlarmInfo) -> None:
        """
        set alarm
        :param name: The name of alarm metrics
        :param description: The description of alarm metrics
        :param info: The alarm info
        : return: None
        """
        cdef:
            string c_name = name.encode()
            string c_description = description.encode()
            CErrorInfo error_info
            CAlarmInfo alarm_info
        alarm_info = alarm_info_from_py(info)
        with nogil:
            error_info = CLibruntimeManager.Instance().GetLibRuntime().get().SetAlarm(c_name, c_description,
                                                                                      alarm_info)
        check_error_info(error_info, "Failed to set alarm")

    def peek_object_ref_stream(self, generator_id: str, blocking: bool, timeout_ms: int):
        """
        peek_object_ref_stream
        :param timeout_ms
        :param generator_id
        :return: object_id
        """
        cdef:
            pair[CErrorInfo, string] ret
            string c_generator_id = generator_id.encode()
            shared_ptr[CBuffer] c_buffer
        with nogil:
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().PeekObjectRefStream(c_generator_id, blocking)

        if not ret.first.OK():
            if error_code_from_cpp(ret.first.Code()) == ErrorCode.ERR_GENERATOR_FINISHED:
                self.decrease_global_reference([ret.second.decode()])
                raise GeneratorEndError(
                    f"gernerator stop, "
                    f"code: {ret.first.Code()}, module code {ret.first.MCode()}, msg: {ret.first.Msg().decode()}")
            else:
                result = []
                if ret.first.Code() == CErrorCode.ERR_USER_FUNCTION_EXCEPTION:
                    c_stack_trace_info = ret.first.GetStackTraceInfos()
                    it_sti = c_stack_trace_info.begin()
                    while it_sti != c_stack_trace_info.end():
                        if dereference(it_sti).Type().decode() == "YRInvokeError":
                            c_buffer = dynamic_pointer_cast[CBuffer, NativeBuffer](
                                make_shared[NativeBuffer](dereference(it_sti).Message().size()))
                            memory_copy(c_buffer, dereference(it_sti).Message().data(),
                                        dereference(it_sti).Message().size())
                            result.append(Buffer.make(c_buffer))
                            return result
                        postincrement(it_sti)
                raise RuntimeError(
                    f"failed to peek object, "
                    f"code: {ret.first.Code()}, module code {ret.first.MCode()}, msg: {ret.first.Msg().decode()}")
        obj_id = ret.second.decode()
        return obj_id

    def generate_group_name(self) -> str:
        """
        generate group name.
        Returns:
            group name
        """
        cdef:
            string c_group_name
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            c_group_name = c_libruntime.get().GenerateGroupName()
        group_name = c_group_name.decode()
        return group_name

    def get_instances(self, obj_id, group_name) -> List[str]:
        """
        get instance id list.
        Returns:
            instance id list
        """
        cdef:
            string c_obj_id = obj_id.encode()
            string c_group_name = group_name.encode()
            pair[vector[string], CErrorInfo] ret
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            ret = c_libruntime.get().GetInstances(c_obj_id, c_group_name)
        if not ret.second.OK():
            raise RuntimeError(
                f"failed to get instances, "
                f"code: {ret.second.Code()}, module code {ret.second.MCode()}, msg: {ret.second.Msg().decode()}")
        return [c_instance_id.decode() for c_instance_id in ret.first]

    def resource_group_table(self, resource_group_id):
        """
        query resource group info
        """
        cdef:
            pair[CErrorInfo, CResourceGroupUnit] ret
            string id = resource_group_id.encode()
            CRgInfo rg_info
            CBundleInfo bundle
            CResource resource
            CScalar scalar
            size_t i
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        ret = c_libruntime.get().GetResourceGroupTable(id)
        if not ret.first.OK():
            raise RuntimeError(
                f"failed to get resources, "
                f"code: {ret.first.Code()}, module code {ret.first.MCode()}, msg: {ret.first.Msg().decode()}")

        return parse_rginfo_to_python(ret.second)

    def resources(self):
        """
        get resources info list.
        Returns:
            resources info list
        """
        cdef:
            pair[CErrorInfo, vector[CResourceUnit]] ret
            vector[string] label_values
        ret = CLibruntimeManager.Instance().GetLibRuntime().get().GetResources()
        if not ret.first.OK():
            raise RuntimeError(
                f"failed to get resources, "
                f"code: {ret.first.Code()}, module code {ret.first.MCode()}, msg: {ret.first.Msg().decode()}")
        capacity = {}
        result = []
        for it in ret.second:
            res = {}
            res['id'] = it.id.decode()
            res['status'] = it.status
            capacity = {}
            for r in it.capacity:
                name = r[0].decode()
                capacity[name] = r[1]
            res['capacity'] = capacity
            allocatable = {}
            for r in it.allocatable:
                name = r[0].decode()
                allocatable[name] = r[1]
            res['allocatable'] = allocatable
            labels = {}
            for r in it.nodeLabels:
                key = r[0].decode()
                label_values = r[1]
                labels[key] = []
                for value in label_values:
                    labels[key].append(value.decode())
            res['labels'] = labels
            result.append(res)
        return result

    def query_named_instances(self):
        """
        """
        cdef pair[CErrorInfo, QueryNamedInsResponse] ret
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        ret = c_libruntime.get().QueryNamedInstances()

        if not ret.first.OK():
            raise RuntimeError(
                f"failed to get queryInstances, "
                f"code: {ret.first.Code()}, module code {ret.first.MCode()}, msg: {ret.first.Msg().decode()}"
            )

        cdef QueryNamedInsResponse rsp = ret.second
        name_result = [rsp.names(i).decode() for i in range(rsp.names_size())]
        return name_result

    def get_node_ip_address(self):
        """
        get node ip address info list.
        Returns:
            node ip address
        """
        cdef:
            pair[CErrorInfo, string] ret
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        ret = c_libruntime.get().GetNodeIpAddress()
        if not ret.first.OK():
            raise RuntimeError(
                f"failed to get node ip address, "
                f"code: {ret.first.Code()}, module code {ret.first.MCode()}, msg: {ret.first.Msg().decode()}")
        return ret.second.decode()

    def get_node_id(self):
        cdef pair[CErrorInfo, string] ret
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        ret = c_libruntime.get().GetNodeId()

        if not ret.first.OK():
            raise RuntimeError(
                f"failed to get node_id, "
                f"code: {ret.first.Code()}, module code {ret.first.MCode()}, msg: {ret.first.Msg().decode()}"
            )

        return ret.second.decode()

    def get_namespace(self):
        cdef string ret
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        ret = c_libruntime.get().GetNameSpace()
        return ret.decode()

    def get_function_group_context(self) -> yr.FunctionGroupContext:
        cdef:
            CFunctionGroupRunningInfo info
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            info = c_libruntime.get().GetFunctionGroupRunningInfo()
        return function_group_running_info_from_cpp(info)

    def get_instance_by_name(self, name, ns, timeout):
        cdef:
            pair[CFunctionMeta, CErrorInfo] ret
            string cinstanceID
            int sigNo = CSignal.KILLINSTANCESYNC
            string cname = name.encode()
            string cns = ns.encode()
            int ctimeout = timeout
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        yr_ns = self.get_namespace()
        cinstanceID = (ns + "-" + name if ns else yr_ns + "-" + name).encode()
        with nogil:
            c_libruntime.get().SetTenantIdWithPriority()
            ret = c_libruntime.get().GetInstance(cname, cns, ctimeout)
        if not ret.second.OK():
            if ret.second.IsTimeout():
                raise TimeoutError(ret.second.Msg().decode())
            raise ValueError(
                f"failed to invoke instance, "
                f"code: {ret.second.Code()}, module code {ret.second.MCode()}, msg: {ret.second.Msg().decode()}")
        return function_meta_from_cpp(ret.first)

    def is_local_instances(self, instance_ids: List[str]) -> bool:
        cdef:
            vector[string] c_instance_ids
            bool ret
        for instance_id in instance_ids:
            c_instance_ids.push_back(instance_id.encode())
        cdef shared_ptr[CLibruntime] c_libruntime = CLibruntimeManager.Instance().GetLibRuntime()
        if c_libruntime == nullptr:
            raise RuntimeError("already finalized")
        with nogil:
            ret = c_libruntime.get().IsLocalInstances(c_instance_ids)
        return ret

    def create_buffer(self, buffer_size: int) -> Tuple[str, SharedBuffer]:
        cdef:
            uint64_t c_size = buffer_size
            shared_ptr[CBuffer] c_buffer
        with nogil:
            CLibruntimeManager.Instance().GetLibRuntime().get().SetTenantIdWithPriority()
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().CreateBuffer(c_size, c_buffer)

        if not ret.first.OK():
            raise RuntimeError(f"failed to create buffer, code: {ret.first.Code()}, "
                               f"module code: {ret.first.MCode()}, msg: {ret.first.Msg().decode()}")
        with nogil:
            result = c_buffer.get().Publish()
        if not result.OK():
            raise RuntimeError(f"failed to publish, code: {result.Code()}, "
                               f"module code: {result.MCode()}, msg: {result.Msg().decode()}")
        return ret.second.decode("utf-8"), SharedBuffer.make(c_buffer)

    def get_buffer(self, obj_id: str)-> SharedBuffer:
        cdef:
            vector[string] c_ids
            int c_timeout = -1
        c_ids.push_back(obj_id)
        with nogil:
            CLibruntimeManager.Instance().GetLibRuntime().get().SetTenantIdWithPriority()
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().GetBuffers(c_ids, c_timeout, False)
        if not ret.first.OK():
            raise RuntimeError(f"failed to get buffer, code: {ret.first.Code()}, "
                               f"module code: {ret.first.MCode()}, msg: {ret.first.Msg().decode()}")
        return SharedBuffer.make(ret.second.front())

    def accelerate(self, group_name: str, handle: Handle):
        cdef:
            vector[string] c_instance_ids
            CAccelerateMsgQueueHandle c_handle
            string c_group_name = group_name.encode()

        c_handle = message_queue_handle_from_py(handle)
        with nogil:
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().Accelerate(c_group_name, c_handle, handle_return_object_callback)
        if not ret.OK():
            raise RuntimeError(
                f"failed to accelerate, code: {ret.Code()}, "
                f"module code: {ret.MCode()}, msg: {ret.Msg().decode()}")

    def add_return_object(self, obj_ids: List[str]):
        cdef:
            vector[string] c_obj_ids
            bool ret
        for obj_id in obj_ids:
            c_obj_ids.push_back(obj_id.encode())
        with nogil:
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().AddReturnObject(c_obj_ids)
        return ret
    
    def create_group(self, group_name: str, group_opts: GroupOptions):
        cdef:
            string c_group_name = group_name.encode()
            CGroupOptions c_group_options
        c_group_options = group_options_from_py(group_name, group_opts)
        with nogil:
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().GroupCreate(c_group_name, c_group_options)
        if not ret.OK():
            raise RuntimeError(
                f"failed to create group, code: {ret.Code()}, "
                f"module code: {ret.MCode()}, msg: {ret.Msg().decode()}")

    def wait_group(self, group_name: str):
        cdef:
            string c_group_name = group_name.encode()
        with nogil:
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().GroupWait(c_group_name)
        if not ret.OK():
            raise RuntimeError(
                f"failed to wait group, code: {ret.Code()}, "
                f"module code: {ret.MCode()}, msg: {ret.Msg().decode()}")

    def suspend_group(self, group_name: str):
        cdef:
            string c_group_name = group_name.encode()
        with nogil:
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().GroupSuspend(c_group_name)
        if not ret.OK():
            raise RuntimeError(
                f"failed to suspend group, code: {ret.Code()}, "
                f"module code: {ret.MCode()}, msg: {ret.Msg().decode()}")

    def resume_group(self, group_name: str):
        cdef:
            string c_group_name = group_name.encode()
        with nogil:
            ret = CLibruntimeManager.Instance().GetLibRuntime().get().GroupResume(c_group_name)
        if not ret.OK():
            raise RuntimeError(
                f"failed to resume group, code: {ret.Code()}, "
                f"module code: {ret.MCode()}, msg: {ret.Msg().decode()}")


def notify_generator_result(generator_id, index, output, error_info):
    """
    notify_generator_result
    :param generator_id
    :param index
    :param output
    :return ErrorInfo
    """
    object_id = generate_random_id()

    cdef:
        CErrorInfo ret
        int c_index= index
        string c_generator_id = generator_id.encode()
        string c_object_id = object_id.encode()
        shared_ptr[CDataObject] dataObj
        shared_ptr[CBuffer] c_buffer
        CErrorInfo c_error_info
        CErrorInfo c_result_err
        vector[string] nested_ids
        unordered_set[string] nested_ids_set
        uint64_t total_native_buffer_size = 0

    c_result_err = error_info_from_py(error_info)
    dataObj = make_shared[CDataObject](c_object_id)
    dataObj.get().alwaysNative = True
    serialized_object = _serialization_ctx.serialize(output)
    data_bytes = serialized_object.to_bytes()
    meta_size = constants.METALEN
    data_size = serialized_object.total_bytes - constants.METALEN

    c_error_info = CLibruntimeManager.Instance().GetLibRuntime().get().AllocReturnObject(
            dataObj, meta_size, data_size, nested_ids, total_native_buffer_size)

    if c_error_info.OK():
        c_buffer = dataObj.get().buffer
        c_error_info = write2buffer(c_buffer, serialized_object, nested_ids_set)
    if c_error_info.OK():
        c_error_info = CLibruntimeManager.Instance().GetLibRuntime().get().NotifyGeneratorResult(c_generator_id, c_index, dataObj, c_result_err)
    if not c_error_info.OK():
        raise RuntimeError(
            f"failed to notify gennerator result, "
            f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")



def notify_generator_finished(generator_id, number_result):
    """
    notify_generator_finished
    :param generator_id
    :param number_result
    :return ErrorInfo
    """

    cdef:
        string c_generator_id = generator_id.encode()
        int c_number_result = number_result;
        CErrorInfo ret

    with nogil:
        ret = CLibruntimeManager.Instance().GetLibRuntime().get().NotifyGeneratorFinished(c_generator_id, c_number_result)

    if not ret.OK():
        raise RuntimeError(
            f"failed to notify generator finished, "
            f"code: {ret.Code()}, module code {ret.MCode()}, msg: {ret.Msg().decode()}")

cdef cluster_access_info_cpp_to_py(const CClusterAccessInfo & c_cluster_info):
    return {
        "serverAddr": c_cluster_info.serverAddr.decode(),
        "dsAddr": c_cluster_info.dsAddr.decode(),
        "inCluster": c_cluster_info.inCluster
    }

cdef CClusterAccessInfo cluster_access_info_py_to_cpp(dict info):
    cdef CClusterAccessInfo c_cluster_info = CClusterAccessInfo()
    if 'serverAddr' in info:
        c_cluster_info.serverAddr = info['serverAddr'].encode('utf-8')
    if 'dsAddr' in info:
        c_cluster_info.dsAddr = info['dsAddr'].encode('utf-8')
    if 'inCluster' in info:
        c_cluster_info.inCluster = info['inCluster']
    return c_cluster_info


def auto_get_cluster_access_info(dict info, args):
    cdef:
        vector[string] c_args
        CClusterAccessInfo c_info
        CClusterAccessInfo result
    for arg in args:
        c_args.push_back(arg.encode())
    c_info = cluster_access_info_py_to_cpp(info)
    result = AutoGetClusterAccessInfo(c_info, c_args)
    return cluster_access_info_cpp_to_py(result)
