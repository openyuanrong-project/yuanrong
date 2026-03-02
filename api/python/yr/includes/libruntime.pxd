# distutils: language = c++
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

from libcpp cimport bool
from libc.stdint cimport uint32_t, uint64_t, uint8_t, int32_t, int64_t
from libcpp.list cimport list
from libcpp.pair cimport pair
from libcpp.string cimport string
from libcpp.vector cimport vector
from libcpp.memory cimport shared_ptr
from libcpp.optional cimport optional
from libcpp.unordered_set cimport unordered_set
from libcpp.unordered_map cimport unordered_map
from yr.includes.affinity cimport CAffinity

ctypedef void (*CWaitAsyncCallback) (const string & object_id, const CErrorInfo & error, void *userData)
ctypedef void (*CGetAsyncCallback) (shared_ptr[CDataObject] data, const CErrorInfo & error, void *userData)
ctypedef pair[CErrorInfo, shared_ptr[CBuffer]] (*CHandleReturnObjectCallback) (shared_ptr[CBuffer] buffer, int rank, string & objId)
cdef extern from * namespace "adapter" nogil:
    """
    namespace adapter {

    template <typename T>
    inline typename std::remove_reference<T>::type&& move(T& t) {
        return std::move(t);
    }

    template <typename T>
    inline typename std::remove_reference<T>::type&& move(T&& t) {
        return std::move(t);
    }

    }  // namespace adapter
    """
    cdef T move[T](T)

cdef extern from "msgpack.hpp" nogil:
    cdef cppclass sbuffer "msgpack::sbuffer":
        sbuffer(size_t initsz)
        void write(const char * buf, size_t len)
        char * data()
        size_t size() const


cdef extern from "src/libruntime/stacktrace/stack_trace_info.h" nogil:
    cdef cppclass CStackTraceInfo "YR::Libruntime::StackTraceInfo":
        CStackTraceInfo(string type, string message, string language)
        CStackTraceInfo()
        string Type() const
        string Message() const
        string Language() const

cdef extern from "src/libruntime/err_type.h" nogil:
    cdef enum CModuleCode "YR::Libruntime::ModuleCode":
        CORE "YR::Libruntime::ModuleCode::CORE"
        RUNTIME "YR::Libruntime::ModuleCode::RUNTIME"
        RUNTIME_CREATE "YR::Libruntime::ModuleCode::RUNTIME_CREATE"
        RUNTIME_INVOKE "YR::Libruntime::ModuleCode::RUNTIME_INVOKE"
        RUNTIME_KILL "YR::Libruntime::ModuleCode::RUNTIME_KILL"
        DATASYSTEM "YR::Libruntime::ModuleCode::DATASYSTEM"

    cdef enum CErrorCode "YR::Libruntime::ErrorCode":
        ERR_OK "YR::Libruntime::ErrorCode::ERR_OK"
        ERR_PARAM_INVALID "YR::Libruntime::ErrorCode::ERR_PARAM_INVALID"
        ERR_RESOURCE_NOT_ENOUGH "YR::Libruntime::ErrorCode::ERR_RESOURCE_NOT_ENOUGH"
        ERR_INSTANCE_NOT_FOUND "YR::Libruntime::ErrorCode::ERR_INSTANCE_NOT_FOUND"
        ERR_INSTANCE_DUPLICATED "YR::Libruntime::ErrorCode::ERR_INSTANCE_DUPLICATED"
        ERR_INVOKE_RATE_LIMITED "YR::Libruntime::ErrorCode::ERR_INVOKE_RATE_LIMITED"
        ERR_RESOURCE_CONFIG_ERROR "YR::Libruntime::ErrorCode::ERR_RESOURCE_CONFIG_ERROR"
        ERR_INSTANCE_EXITED "YR::Libruntime::ErrorCode::ERR_INSTANCE_EXITED"
        ERR_EXTENSION_META_ERROR "YR::Libruntime::ErrorCode::ERR_EXTENSION_META_ERROR"

        ERR_USER_CODE_LOAD "YR::Libruntime::ErrorCode::ERR_USER_CODE_LOAD"
        ERR_USER_FUNCTION_EXCEPTION "YR::Libruntime::ErrorCode::ERR_USER_FUNCTION_EXCEPTION"

        ERR_REQUEST_BETWEEN_RUNTIME_BUS "YR::Libruntime::ErrorCode::ERR_REQUEST_BETWEEN_RUNTIME_BUS"
        ERR_INNER_COMMUNICATION "YR::Libruntime::ErrorCode::ERR_INNER_COMMUNICATION"
        ERR_INNER_SYSTEM_ERROR "YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR"
        ERR_DISCONNECT_FRONTEND_BUS "YR::Libruntime::ErrorCode::ERR_DISCONNECT_FRONTEND_BUS"
        ERR_ETCD_OPERATION_ERROR "YR::Libruntime::ErrorCode::ERR_ETCD_OPERATION_ERROR"
        ERR_BUS_DISCONNECTION "YR::Libruntime::ErrorCode::ERR_BUS_DISCONNECTION"
        ERR_REDIS_OPERATION_ERROR "YR::Libruntime::ErrorCode::ERR_REDIS_OPERATION_ERROR"

        ERR_INCORRECT_INIT_USAGE "YR::Libruntime::ErrorCode::ERR_INCORRECT_INIT_USAGE"
        ERR_INIT_CONNECTION_FAILED "YR::Libruntime::ErrorCode::ERR_INIT_CONNECTION_FAILED"
        ERR_DESERIALIZATION_FAILED "YR::Libruntime::ErrorCode::ERR_DESERIALIZATION_FAILED"
        ERR_INSTANCE_ID_EMPTY "YR::Libruntime::ErrorCode::ERR_INSTANCE_ID_EMPTY"
        ERR_GET_OPERATION_FAILED "YR::Libruntime::ErrorCode::ERR_GET_OPERATION_FAILED"
        ERR_INCORRECT_FUNCTION_USAGE "YR::Libruntime::ErrorCode::ERR_INCORRECT_FUNCTION_USAGE"
        ERR_INCORRECT_CREATE_USAGE "YR::Libruntime::ErrorCode::ERR_INCORRECT_CREATE_USAGE"
        ERR_INCORRECT_INVOKE_USAGE "YR::Libruntime::ErrorCode::ERR_INCORRECT_INVOKE_USAGE"
        ERR_INCORRECT_KILL_USAGE "YR::Libruntime::ErrorCode::ERR_INCORRECT_KILL_USAGE"

        ERR_ROCKSDB_FAILED "YR::Libruntime::ErrorCode::ERR_ROCKSDB_FAILED"
        ERR_SHARED_MEMORY_LIMITED "YR::Libruntime::ErrorCode::ERR_SHARED_MEMORY_LIMITED"
        ERR_OPERATE_DISK_FAILED "YR::Libruntime::ErrorCode::ERR_OPERATE_DISK_FAILED"
        ERR_INSUFFICIENT_DISK_SPACE "YR::Libruntime::ErrorCode::ERR_INSUFFICIENT_DISK_SPACE"
        ERR_CONNECTION_FAILED "YR::Libruntime::ErrorCode::ERR_CONNECTION_FAILED"
        ERR_KEY_ALREADY_EXIST "YR::Libruntime::ErrorCode::ERR_KEY_ALREADY_EXIST"
        ERR_DEPENDENCY_FAILED "YR::Libruntime::ErrorCode::ERR_DEPENDENCY_FAILED"
        ERR_DATASYSTEM_FAILED "YR::Libruntime::ErrorCode::ERR_DATASYSTEM_FAILED"
        ERR_GENERATOR_FINISHED "YR::Libruntime::ErrorCode::ERR_GENERATOR_FINISHED"
        ERR_CLIENT_TERMINAL_KILLED "YR::Libruntime::ErrorCode::ERR_CLIENT_TERMINAL_KILLED"

    cdef cppclass CErrorInfo "YR::Libruntime::ErrorInfo":
        CErrorInfo()
        CErrorInfo(CErrorCode errCode, CModuleCode moduleCode, string & errMsg)
        bool OK() const
        bool IsTimeout() const
        string GetExceptionMsg(const vector[string] & failIds, int timeoutMs) const
        string Msg() const
        CModuleCode MCode() const
        CErrorCode Code() const
        void SetErrorMsg(const string & errMsg)
        vector[CStackTraceInfo] GetStackTraceInfos() const

cdef extern from "src/dto/status.h" nogil:
    cdef cppclass CStatus "YR::Status":
        CStatus()
        int Code() const

cdef extern from "src/dto/function_group_running_info.h" nogil:
    cdef cppclass CDeviceInfo "YR::Libruntime::DeviceInfo":
        int deviceId
        string deviceIp
        int rankId

    cdef cppclass CServerInfo "YR::Libruntime::ServerInfo":
        vector[CDeviceInfo] devices
        string serverId

    cdef cppclass CFunctionGroupRunningInfo "YR::Libruntime::FunctionGroupRunningInfo":
        vector[CServerInfo] serverList
        int instanceRankId
        int worldSize
        string deviceName

cdef extern from "src/libruntime/objectstore/device_buffer.h" nogil:
    cdef enum CConsistencyType "YR::Libruntime::ConsistencyType":
        PRAM "YR::Libruntime::ConsistencyType::PRAM"
        CAUSAL "YR::Libruntime::ConsistencyType::CAUSAL"

    cdef cppclass CCreateParam "YR::Libruntime::CreateParam":
        CConsistencyType consistencyType
        CCacheType cacheType

cdef extern from "src/dto/tensor.h" namespace "YR::Libruntime" nogil:
    cdef enum CDataType "YR::Libruntime::DataType":
        INT8 "YR::Libruntime::DataType::INT8"
        INT16 "YR::Libruntime::DataType::INT16"
        INT32 "YR::Libruntime::DataType::INT32"
        INT64 "YR::Libruntime::DataType::INT64"
        UINT8 "YR::Libruntime::DataType::UINT8"
        UINT16 "YR::Libruntime::DataType::UINT16"
        UINT32 "YR::Libruntime::DataType::UINT32"
        UINT64 "YR::Libruntime::DataType::UINT64"
        FLOAT16 "YR::Libruntime::DataType::FLOAT16"
        FLOAT32 "YR::Libruntime::DataType::FLOAT32"
        FLOAT64 "YR::Libruntime::DataType::FLOAT64"
        BFLOAT16 "YR::Libruntime::DataType::BFLOAT16"
        COMPLEX64 "YR::Libruntime::DataType::COMPLEX64"
        COMPLEX128 "YR::Libruntime::DataType::COMPLEX128"
    cdef cppclass CTensor "YR::Libruntime::Tensor":
        uint8_t *ptr
        uint64_t size
        uint64_t count
        CDataType dtype
        int32_t deviceIdx

cdef extern from "src/dto/buffer.h" namespace "YR::Libruntime" nogil:
    cdef cppclass CBuffer "YR::Libruntime::Buffer":
        CErrorInfo MemoryCopy(const void *data, uint64_t length)
        uint64_t GetSize() const
        const void *ImmutableData() const
        void *MutableData()
        CErrorInfo Seal(const unordered_set[string] & nestedIds)
        CErrorInfo WriterLatch()
        CErrorInfo WriterUnlatch()
        CErrorInfo ReaderLatch()
        CErrorInfo ReaderUnlatch()
        CErrorInfo Publish()

    cdef cppclass NativeBuffer(CBuffer):
        NativeBuffer(void *data, uint64_t size)
        NativeBuffer(uint64_t size)

    cdef cppclass StringNativeBuffer(CBuffer):
        StringNativeBuffer(uint64_t size)

    cdef cppclass SharedBuffer(CBuffer):
        SharedBuffer(void *data, uint64_t size)

cdef extern from "src/libruntime/libruntime_options.h" nogil:
    cdef struct CLibruntimeOptions "YR::Libruntime::LibruntimeOptions":
        (CErrorInfo(const CFunctionMeta & function, const CInvokeType invokeType,
                    const vector[shared_ptr[CDataObject]] & rawArgs,
                    vector[shared_ptr[CDataObject]] & returnValues)) functionExecuteCallback
        (CErrorInfo(const vector[string] & codePath)) loadFunctionCallback
        (CErrorInfo(const string & checkpointId, shared_ptr[CBuffer] & data)) checkpointCallback
        (CErrorInfo(shared_ptr[CBuffer] & data)) recoverCallback
        (CErrorInfo(uint64_t gracePeriodSeconds)) shutdownCallback
        (CErrorInfo(int sigNo, shared_ptr[CBuffer] & payload)) signalCallback
        (CErrorInfo(const CAccelerateMsgQueueHandle & intputHandle, CAccelerateMsgQueueHandle & outputHandle)) accelerateCallback

cdef extern from "src/libruntime/libruntime_config.h":
    cdef const int MAX_PASSWD_LENGTH

cdef extern from "src/libruntime/libruntime_config.h" nogil:
    cdef struct CLibruntimeConfig "YR::Libruntime::LibruntimeConfig":
        string functionSystemIpAddr
        int functionSystemPort
        string functionSystemRtServerIpAddr
        int functionSystemRtServerPort
        string dataSystemIpAddr
        int dataSystemPort
        bool isDriver
        string jobId
        string runtimeId
        unordered_map[CLanguageType, string] functionIds
        string logLevel
        string logDir
        uint32_t logFileSizeMax
        uint32_t logFileNumMax
        bool logToDriver
        bool dedupLogs
        int logFlushInterval
        CLibruntimeOptions libruntimeOptions
        string metaConfig
        int recycleTime
        int maxTaskInstanceNum
        int maxConcurrencyCreateNum
        bool enableMetrics
        vector[string] functionMasters
        uint32_t threadPoolSize
        vector[string] loadPaths
        CLanguageType selfLanguage
        bool enableMTLS
        string privateKeyPath
        string certificateFilePath
        string verifyFilePath
        char privateKeyPaaswd[MAX_PASSWD_LENGTH]
        shared_ptr[void] tlsContext
        uint32_t httpIocThreadsNum
        string serverName
        string clientId
        bool inCluster
        string ns
        int32_t rpcTimeout
        string tenantId
        unordered_map[string, string] customEnvs
        CFunctionGroupRunningInfo groupRunningInfo
        (CErrorInfo() nogil) checkSignals_
        string workingDir
        string runtimePublicKeyPath
        string runtimePrivateKeyPath
        string dsPublicKeyPath
        bool encryptEnable
        CLibruntimeConfig()
        void InitConfig(const CMetaConfig & config)
        void BuildMetaConfig(CMetaConfig & config)

cdef extern from "src/proto/libruntime.pb.h" nogil:
    cdef enum CLanguageType "libruntime::LanguageType":
        LANGUAGE_CPP "libruntime::LanguageType::Cpp"
        LANGUAGE_PYTHON "libruntime::LanguageType::Python"
        LANGUAGE_JAVA "libruntime::LanguageType::Java"
        LANGUAGE_GOLANG "libruntime::LanguageType::Golang"
        LANGUAGE_NODE_JS "libruntime::LanguageType::NodeJS"
        LANGUAGE_CSHARP "libruntime::LanguageType::CSharp"
        LANGUAGE_PHP "libruntime::LanguageType::Php"

    cdef enum CInvokeType "libruntime::InvokeType":
        CREATE_INSTANCE "libruntime::InvokeType::CreateInstance"
        INVOKE_MEMBER_FUNCTION "libruntime::InvokeType::InvokeFunction"
        CREATE_NORMAL_FUNCTION_INSTANCE "libruntime::InvokeType::CreateInstanceStateless"
        INVOKE_NORMAL_FUNCTION "libruntime::InvokeType::InvokeFunctionStateless"
        GET_NAMED_INSTANCE_METADATA "libruntime::InvokeType::GetNamedInstanceMeta"
        DELETE_REMOTE_TENSOR "libruntime::InvokeType::DeleteRemoteTensor"

    cdef enum CSubscriptionType "libruntime::SubscriptionType":
        STREAM "libruntime::SubscriptionType::STREAM"
        ROUND_ROBIN "libruntime::SubscriptionType::ROUND_ROBIN"
        KEY_PARTITIONS "libruntime::SubscriptionType::KEY_PARTITIONS"
        UNKNOWN "libruntime::SubscriptionType::UNKNOWN"

    cdef enum CApiType "libruntime::ApiType":
        ACTOR "libruntime::ApiType::Function",
        FAAS "libruntime::ApiType::Faas"
        POSIX "libruntime::ApiType::Posix"
        SERVE "libruntime::ApiType::Serve"

    cdef enum CSignal "libruntime::Signal":
        DEFAULTSIGNAL "libruntime::Signal::DefaultSignal",
        KILLINSTANCE "libruntime::Signal::KillInstance"
        KILLALLINSTANCES "libruntime::Signal::KillAllInstances"
        KILLINSTANCESYNC "libruntime::Signal::killInstanceSync"
        KILLGROUPINSTANCE "libruntime::Signal::KillGroupInstance"
        UPDATEALIAS "libruntime::Signal::UpdateAlias"
        UPDATEFRONTEND "libruntime::Signal::UpdateFrontend"
        UPDATESCHEDULER "libruntime::Signal::UpdateScheduler"
        CANCEL "libruntime::Signal::Cancel"

cdef extern from "src/dto/device.h" nogil:
    cdef cppclass CDevice "YR::Libruntime::Device":
        CDevice()
        string name
        size_t batch_size

cdef extern from "src/dto/accelerate.h" nogil:
    cdef cppclass CAccelerateMsgQueueHandle "YR::Libruntime::AccelerateMsgQueueHandle":
        int worldSize
        int rank
        int maxChunkBytes
        int maxChunks
        string name
        bool isAsync

cdef extern from "src/dto/invoke_options.h" nogil:
    cdef cppclass CMetaFunctionID "YR::Libruntime::MetaFunctionID":
        string cpp
        string python
        string java

    cdef cppclass CFunctionMeta "YR::Libruntime::FunctionMeta":
        string appName
        string moduleName
        string funcName
        string className
        CLanguageType languageType
        string codeId
        string signature
        string functionId
        CApiType apiType
        string name
        string ns
        string initializerCodeId
        bool isGenerator
        bool isAsync
        string tensorTransportTarget
        bool enableTensorTransport

    cdef cppclass CGroupOptions "YR::Libruntime::GroupOpts":
        string groupName
        int timeout
        bool sameLifecycle
        string strategy
        
    cdef enum CBundleAffinity "YR::Libruntime::BundleAffinity":
        COMPACT "YR::Libruntime::BundleAffinity::COMPACT"
        DISCRETE "YR::Libruntime::BundleAffinity::DISCRETE"

    cdef cppclass CFunctionGroupOptions "YR::Libruntime::FunctionGroupOptions":
        int functionGroupSize
        int bundleSize
        CBundleAffinity bundleAffinity
        int timeout
        bool sameLifecycle

    cdef cppclass CResourceGroupOptions "YR::Libruntime::ResourceGroupOptions":
        string resourceGroupName
        int bundleIndex

    cdef cppclass CDebugConfig "YR:Libruntime::DebugConfig":
        bool enable

    cdef cppclass CInvokeOptions "YR::Libruntime::InvokeOptions":
        int cpu
        int memory
        unordered_map[string, float] customResources
        unordered_map[string, string] customExtensions
        unordered_map[string, string] createOptions
        unordered_map[string, string] podLabels
        unordered_map[string, string] aliasParams
        vector[string] labels
        unordered_map[string, string] affinity
        size_t retryTimes
        size_t priority
        vector[string] codePaths
        string schedulerFunctionId
        string schedulerInstanceIds
        CDevice device
        int maxInvokeLatency
        int minInstances
        int maxInstances
        bool isDataAffinity
        list[shared_ptr[CAffinity]] scheduleAffinities
        bool needOrder
        int recoverRetryTimes
        CFunctionGroupOptions functionGroupOpts
        CResourceGroupOptions resourceGroupOpts
        string groupName
        unordered_map[string, string] envVars
        int timeout
        bool isGetInstance
        bool isDeleteRemoteTensor
        string traceId
        string workingDir
        bool preemptedAllowed
        int instancePriority
        int64_t scheduleTimeoutMs
        CDebugConfig debug

    cdef cppclass CMetaConfig "YR::Libruntime::MetaConfig":
        string jobID
        string codePath
        int recycleTime
        int maxTaskInstanceNum
        int maxConcurrencyCreateNum
        bool enableMetrics
        vector[string] functionMasters
        uint32_t threadPoolSize
        CMetaFunctionID functionID

    cdef cppclass CInstanceOptions "YR::Libruntime::InstanceOptions":
        bool needOrder

    cdef cppclass CUInt64CounterData "YR::Libruntime::UInt64CounterData":
        string name
        string description
        string unit
        unordered_map[string, string] labels
        uint64_t value

    cdef cppclass CDoubleCounterData "YR::Libruntime::DoubleCounterData":
        string name
        string description
        string unit
        unordered_map[string, string] labels
        double value

    cdef cppclass CGaugeData "YR::Libruntime::GaugeData":
        string name
        string description
        string unit
        unordered_map[string, string] labels
        double value

    cdef enum CAlarmSeverity "YR::Libruntime::AlarmSeverity":
        OFF "YR::Libruntime::AlarmSeverity::OFF"
        INFO "YR::Libruntime::AlarmSeverity::INFO"
        MINOR "YR::Libruntime::AlarmSeverity::MINOR"
        MAJOR "YR::Libruntime::AlarmSeverity::MAJOR"
        CRITICAL "YR::Libruntime::AlarmSeverity::CRITICAL"

    cdef cppclass CAlarmInfo "YR::Libruntime::AlarmInfo":
        string alarmName
        CAlarmSeverity alarmSeverity
        string locationInfo
        string cause
        long startsAt
        long endsAt
        long timeout
        unordered_map[string, string] customOptions


cdef extern from "src/dto/invoke_arg.h" nogil:
    cdef cppclass CInvokeArg "YR::Libruntime::InvokeArg":
        CInvokeArg()
        shared_ptr[CDataObject] dataObj
        bool isRef
        string objId
        string tenantId
        unordered_set[string] nestedObjects

cdef extern from "src/dto/internal_wait_result.h" nogil:
    cdef cppclass CInternalWaitResult "YR::InternalWaitResult":
        vector[string] readyIds
        vector[string] unreadyIds
        unordered_map[string, CErrorInfo] exceptionIds

cdef extern from "src/dto/data_object.h" nogil:
    cdef cppclass CDataObject "YR::Libruntime::DataObject":
        CDataObject()
        CDataObject(const string & objId, bool isFundaType)
        CDataObject(const string & objId, shared_ptr[CBuffer] dataBuf)
        CDataObject(uint64_t meta_size, uint64_t data_size)
        void SetDataBuf(shared_ptr[CBuffer] dataBuf)
        void SetBuffer(shared_ptr[CBuffer] buf)
        string id
        bool alwaysNative
        shared_ptr[CBuffer] buffer
        shared_ptr[CBuffer] data
        shared_ptr[CBuffer] meta
        vector[string] nestedObjIds

cdef extern from "src/dto/types.h" nogil:
    cdef enum CWriteMode "YR::Libruntime::WriteMode":
        NONE_L2_CACHE "YR::Libruntime::WriteMode::NONE_L2_CACHE"
        WRITE_THROUGH_L2_CACHE "YR::Libruntime::WriteMode::WRITE_THROUGH_L2_CACHE"
        WRITE_BACK_L2_CACHE "YR::Libruntime::WriteMode::WRITE_BACK_L2_CACHE"
        NONE_L2_CACHE_EVICT "YR::Libruntime::WriteMode::NONE_L2_CACHE_EVICT"

    cdef enum CCacheType "YR::Libruntime::CacheType":
        MEMORY "YR::Libruntime::CacheType::MEMORY"
        DISK "YR::Libruntime::CacheType::DISK"

cdef extern from "src/dto/resource_group_spec.h" nogil:
    cdef cppclass CResourceGroupSpec "YR::Libruntime::ResourceGroupSpec":
        string name
        string strategy
        vector[unordered_map[string, double]] bundles

cdef extern from "src/libruntime/statestore/state_store.h" nogil:
    cdef enum CExistenceOpt "YR::Libruntime::ExistenceOpt":
        NONE "YR::Libruntime::ExistenceOpt::NONE"
        NX "YR::Libruntime::ExistenceOpt::NX"

    cdef cppclass CSetParam "YR::Libruntime::SetParam":
        CWriteMode writeMode
        uint32_t ttlSecond
        CExistenceOpt existence
        CCacheType cacheType

    cdef cppclass CMSetParam "YR::Libruntime::MSetParam":
        CWriteMode writeMode
        uint32_t ttlSecond
        CExistenceOpt existence
        CCacheType cacheType

    cdef cppclass CGetParam "YR::Libruntime::GetParam":
        uint64_t offset;
        uint64_t size;

    cdef cppclass CGetParams "YR::Libruntime::GetParams":
        vector[CGetParam] getParams;

    ctypedef pair[string, CErrorInfo] CSingleGetResult "YR::Libruntime::SingleGetResult"

    ctypedef pair[vector[string], CErrorInfo] CMultipleGetResult "YR::Libruntime::MultipleGetResult"

    ctypedef pair[vector[string], CErrorInfo] CMultipleDelResult "YR::Libruntime::MultipleDelResult"

    ctypedef pair[shared_ptr[CBuffer], CErrorInfo] CSingleReadResult "YR::Libruntime::SingleReadResult"

    ctypedef pair[vector[shared_ptr[CBuffer]], CErrorInfo] CMultipleReadResult "YR::Libruntime::MultipleReadResult"

cdef extern from "src/dto/stream_conf.h" nogil:
    cdef cppclass CElement "YR::Libruntime::Element":
        CElement(uint8_t *ptr, uint64_t size, uint64_t id)
        CElement()
        uint8_t *ptr
        uint64_t size
        uint64_t id

    cdef cppclass CProducerConf "YR::Libruntime::ProducerConf":
        int64_t delayFlushTime
        int64_t pageSize
        uint64_t maxStreamSize
        bool autoCleanup
        bool encryptStream
        uint64_t retainForNumConsumers
        uint64_t reserveSize
        unordered_map[string, string] extendConfig

    cdef cppclass CSubscriptionConfig "YR::Libruntime::SubscriptionConfig":
        string subscriptionName
        CSubscriptionType subscriptionType
        unordered_map[string, string] extendConfig
        SubscriptionConfig(string subName, const CSubscriptionType subType)
        SubscriptionConfig()

cdef extern from "src/dto/resource_unit.h" nogil:
    cdef cppclass CResourceUnit "YR::Libruntime::ResourceUnit":
        string id
        uint32_t status
        unordered_map[string, float] capacity
        unordered_map[string, float] allocatable
        unordered_map[string, vector[string]] nodeLabels

cdef extern from "src/dto/resource_unit.h" nogil:
    cdef cppclass CScalar "YR::Libruntime::Resource::Scalar":
        double value;
        double limit;

    cdef enum CType "YR::Libruntime::Resource::Type":
        SCALER "YR::Libruntime::Resource::Type::SCALER"

    cdef cppclass CResource "YR::Libruntime::Resource":
        string name
        CType type
        CScalar scalar

    cdef cppclass CResources "YR::Libruntime::Resources":
        unordered_map[string, CResource] resources;

    cdef cppclass CCommonSatus "YR::Libruntime::CofmmonSatus":
        int code;
        string message;

    cdef cppclass CBundleInfo "YR::Libruntime::BundleInfo":
        string bundleId;
        string rGroupName;
        string parentRGroupName;
        string functionProxyId;
        string functionAgentId;
        string tenantId;
        CResources resources;
        vector[string] labels;
        CCommonSatus status;
        string parentId;
        unordered_map[string, string] kvLabels;

    cdef cppclass COption "YR::Libruntime::COption":
        int priority;
        int groupPolicy;
        unordered_map[string, string] extension;

    cdef cppclass CRgInfo "YR::Libruntime::RgInfo":
        string name;
        string owner;
        string appId;
        string tenantId;
        vector[CBundleInfo] bundles;
        CCommonSatus status;
        string parentId;
        string requestId;
        string traceId;
        COption opt;

    cdef cppclass CResourceGroupUnit "YR::Libruntime::ResourceGroupUnit":
        unordered_map[string, CRgInfo] resourceGroups;

cdef extern from "src/libruntime/streamstore/stream_producer_consumer.h" nogil:
    cdef cppclass CStreamProducer "YR::Libruntime::StreamProducer":
        CErrorInfo Send(const CElement & element)
        CErrorInfo Send(const CElement & element, int64_t timeoutMs)
        CErrorInfo Flush()
        CErrorInfo Close()

    cdef cppclass CStreamConsumer "YR::Libruntime::StreamConsumer":
        CErrorInfo Receive(uint32_t expectNum, uint32_t timeoutMs, vector[CElement] & outElements)
        CErrorInfo Receive(uint32_t timeoutMs, vector[CElement] & outElements)
        CErrorInfo Ack(uint64_t elementId)
        CErrorInfo Close()

cdef extern from "src/libruntime/libruntime.h" nogil:
    cdef cppclass CLibruntime "YR::Libruntime::Libruntime":
        CLibruntime(shared_ptr[CLibruntimeConfig] config)
        pair[CErrorInfo, string] CreateInstance(const CFunctionMeta & functionMeta, vector[CInvokeArg] & invokeArgs,
                                                CInvokeOptions & opts)
        CErrorInfo InvokeByFunctionName(const CFunctionMeta & funcMeta, vector[CInvokeArg] & args,
                                        CInvokeOptions & opt, vector[CDataObject] & returnObjs)
        CErrorInfo InvokeByInstanceId(const CFunctionMeta & funcMeta, const string & instanceId,
                                      vector[CInvokeArg] & args, CInvokeOptions & opt,
                                      vector[CDataObject] & returnObjs)
        shared_ptr[CInternalWaitResult] Wait(const vector[string] & objs, size_t waitNum, int timeout)
        pair[CErrorInfo, string] Put(shared_ptr[sbuffer] data, const unordered_set[string] & nestedId)
        CErrorInfo Put(shared_ptr[CBuffer] data, const string & objID, const unordered_set[string] & nestedID,
                               bool toDataSystem, const CCreateParam & createParam)
        pair[CErrorInfo, vector[shared_ptr[sbuffer]]] Get(const vector[string] & ids, int timeoutMs, bool allowPartial)
        CErrorInfo IncreaseReference(const vector[string] & objids)
        void DecreaseReference(const vector[string] & objids)
        CErrorInfo AllocReturnObject(shared_ptr[CDataObject] & returnObj, size_t metaSize, size_t dataSize,
                                     const vector[string] & nestedObjIds, uint64_t & totalNativeBufferSize)
        CErrorInfo CreateReturnValue(size_t dataSize, shared_ptr[CBuffer] & dataBuf)
        void Cancel(const vector[string] & objids, bool isForce, bool isRecursive)
        void Exit()
        CErrorInfo Kill(const string & instanceId, int sigNo)
        void GroupTerminate(const string & groupName)
        string GetRealInstanceId(const string & objectId)
        void SaveRealInstanceId(const string & objectId, const string & instanceId, const CInstanceOptions & opts)
        void Finalize()
        CErrorInfo Init()

        pair[CErrorInfo, string] CreateBuffer(uint64_t dataSize, shared_ptr[CBuffer] & dataBuf)
        pair[CErrorInfo, vector[shared_ptr[CBuffer]]] GetBuffers(const vector[string] & ids, int timeoutMs,
                                                                 bool allowPartial)
        pair[CErrorInfo, string] CreateDataObject(uint64_t metaSize, uint64_t dataSize,
                                                  shared_ptr[CDataObject] & dataObj,
                                                  const vector[string] & nestedObjIds,
                                                  const CCreateParam & createParam)
        pair[CErrorInfo, vector[shared_ptr[CDataObject]]] GetDataObjects(const vector[string] & ids, int timeoutMs,
                                                                         bool allowPartial)
        bool IsObjectExistingInLocal(const string & objId)
        CErrorInfo KVWrite(const string & key, shared_ptr[CBuffer] value, CSetParam setParam)
        CErrorInfo KVMSetTx(const vector[string] & keys, const vector[shared_ptr[CBuffer]] & values, CMSetParam mSetParam)
        CSingleReadResult KVRead(const string & key, int timeoutMS)
        CMultipleReadResult KVRead(const vector[string] & keys, int timeoutMS, bool allowPartial)
        CMultipleReadResult KVGetWithParam(const vector[string] & keys, const CGetParams & params, int timeoutMs);
        CErrorInfo KVDel(const string & key)
        CMultipleDelResult KVDel(const vector[string] & keys)

        CErrorInfo CreateStreamProducer(const string & streamName, CProducerConf producerConf,
                                        shared_ptr[CStreamProducer] producer)
        CErrorInfo CreateStreamConsumer(const string & streamName, const CSubscriptionConfig & config,
                                        shared_ptr[CStreamConsumer] consumer)
        CErrorInfo DeleteStream(const string & streamName)
        CErrorInfo QueryGlobalProducersNum(const string & streamName, uint64_t & gProducerNum)
        CErrorInfo QueryGlobalConsumersNum(const string & streamName, uint64_t & gConsumerNum)

        void SetTenantIdWithPriority()
        string GetTenantId()

        CErrorInfo SaveState(const shared_ptr[CBuffer] data, const int & timeout);
        CErrorInfo LoadState(shared_ptr[CBuffer] & data, const int & timeout);

        CErrorInfo CreateResourceGroup(const CResourceGroupSpec &resourceGroupSpec, string & requestId);
        CErrorInfo RemoveResourceGroup(const string &resourceGroupName);
        CErrorInfo WaitResourceGroup(const string & resourceGroupName, const string & requestId, int timeoutSec);

        CErrorInfo SetUInt64Counter(const CUInt64CounterData & data);
        CErrorInfo ResetUInt64Counter(const CUInt64CounterData & data);
        CErrorInfo IncreaseUInt64Counter(const CUInt64CounterData & data);
        pair[CErrorInfo, int] GetValueUInt64Counter(const CUInt64CounterData & data);
        CErrorInfo SetDoubleCounter(const CDoubleCounterData & data);
        CErrorInfo ResetDoubleCounter(const CDoubleCounterData & data);
        CErrorInfo IncreaseDoubleCounter(const CDoubleCounterData & data);
        pair[CErrorInfo, float] GetValueDoubleCounter(const CDoubleCounterData & data);
        CErrorInfo ReportGauge(const CGaugeData & gauge);
        CErrorInfo SetAlarm(const string & name, const string & description, const CAlarmInfo & alarmInfo);

        string GenerateGroupName();

        pair[CErrorInfo, string] PeekObjectRefStream(const string & generatorId, bool blocking);
        CErrorInfo NotifyGeneratorResult(const string & generatorId, int index, shared_ptr[CDataObject] & resultObj, CErrorInfo & resultErr);
        CErrorInfo NotifyGeneratorFinished(const string & generatorId, int numResults);
        void WaitAsync(const string & objectId, CWaitAsyncCallback callback, void *userData);
        void GetAsync(const string & objectId, CGetAsyncCallback callback, void *userData);

        pair[vector[string], CErrorInfo] GetInstances(const string & objId, const string & groupName);

        void WaitEvent(CFiberEvent &event);

        CFunctionGroupRunningInfo GetFunctionGroupRunningInfo();

        pair[CFunctionMeta, CErrorInfo] GetInstance(const string & name, const string & nameSpace, int timeoutSec);

        CErrorInfo ProcessLog(const string & functionLog);

        pair[CErrorInfo, vector[CResourceUnit]] GetResources();

        pair[CErrorInfo, QueryNamedInsResponse] QueryNamedInstances();

        pair[CErrorInfo, CResourceGroupUnit] GetResourceGroupTable(const string & id);

        pair[CErrorInfo, string] GetNodeIpAddress();

        pair[CErrorInfo, string] GetNodeId();

        string GetNameSpace();

        bool IsLocalInstances(const vector[string] & instanceIds);

        bool AddReturnObject(const vector[string] & objIds);

        CErrorInfo Accelerate(const string & group_name, const CAccelerateMsgQueueHandle & handle, CHandleReturnObjectCallback callback);

        bool SetError(const string &objId, const CErrorInfo &err);

        CErrorInfo GroupCreate(const string & groupName, CGroupOptions & opts);

        CErrorInfo GroupWait(const string & groupName);

        CErrorInfo GroupSuspend(const string & groupName);

        CErrorInfo GroupResume(const string & groupName);


cdef extern from "src/libruntime/libruntime_manager.h" nogil:
    cdef cppclass CLibruntimeManager "YR::Libruntime::LibruntimeManager":
        @ staticmethod
        CLibruntimeManager & Instance()
        CErrorInfo Init(const CLibruntimeConfig & config)
        @ staticmethod
        void Finalize()
        @ staticmethod
        shared_ptr[CLibruntime] GetLibRuntime()
        @ staticmethod
        bool IsInitialized()
        void ReceiveRequestLoop()

cdef extern from "src/libruntime/auto_init.h" namespace "YR::Libruntime" nogil:
    cdef cppclass CClusterAccessInfo "YR::Libruntime::ClusterAccessInfo":
        string serverAddr
        string dsAddr
        bool inCluster

        CClusterAccessInfo()

    CClusterAccessInfo AutoGetClusterAccessInfo(const CClusterAccessInfo& info, vector[string] args)

cdef extern from "src/libruntime/libruntime.h" nogil:
    cdef cppclass CFiberEvent "YR::Libruntime::FiberEventNotify":
        CFiberEvent()
        void Wait()
        void Notify()


cdef extern from "src/libruntime/fmclient/fm_client.h" namespace "YR::Libruntime":
    cdef cppclass QueryNamedInsResponse "messages::QueryNamedInsResponse":
        QueryNamedInsResponse() except +
        int names_size() const
        string names(int idx) const



