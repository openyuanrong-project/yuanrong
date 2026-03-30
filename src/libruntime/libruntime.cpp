/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include "src/libruntime/libruntime.h"
#include <iostream>
#include "invoke_order_manager.h"
#include "re2/re2.h"
#include "src/dto/config.h"
#include "src/dto/data_object.h"
#include "src/dto/status.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/fsclient/fs_client.h"
#include "src/libruntime/generator/stream_generator_notifier.h"
#include "src/libruntime/generator/stream_generator_receiver.h"
#include "src/libruntime/gwclient/gw_client.h"
#include "src/libruntime/invokeadaptor/request_manager.h"
#include "src/libruntime/metricsadaptor/metrics_adaptor.h"
#include "src/libruntime/objectstore/memory_store.h"
#include "src/libruntime/utils/ak_sk_manager.h"
#include "src/libruntime/utils/serializer.h"
#include "src/utility/id_generator.h"
#include "src/utility/string_utility.h"

namespace YR {
namespace Libruntime {
const std::string HETERO_PREFIX = "hetero-dev-buf-";
const int MAX_INS_ID_LENGTH = 128;
const std::string DELEGATE_DIRECTORY_QUOTA = "DELEGATE_DIRECTORY_QUOTA";
const std::string DELEGATE_DIRECTORY_INFO = "DELEGATE_DIRECTORY_INFO";
const std::string DEFALUT_DELEGATE_DIRECTORY_INFO = "/tmp";
const std::string FAAS_INSTANCE_TYPE = "faas";
const std::string ACTOR_INSTANCE_TYPE = "actor";
const char *DEFAULT_DELEGATE_DIRECTORY_QUOTA = "512";  // 512MB
const int MAX_DELEGATE_DIRECTORY_QUOTA = 1024 * 1024;  // 1TB
const std::string QUOTA_NO_LIMIT = "-1";
const re2::RE2 POD_LABELS_KEY_REGEX("^[a-zA-Z0-9]([-a-zA-Z0-9]{0,61}[a-zA-Z0-9])?$");
const re2::RE2 POD_LABELS_VALUE_REGEX("^[a-zA-Z0-9]([-a-zA-Z0-9]{0,61}[a-zA-Z0-9])?$|^$");
const std::string DISPATCHER = "dis";
const size_t NUM_DISPATCHER = 2;
thread_local std::string threadLocalTraceId;
thread_local std::string threadLocalRequestId;
thread_local std::string threadLocalInstanceId;

Libruntime::Libruntime(std::shared_ptr<LibruntimeConfig> librtCfg, std::shared_ptr<ClientsManager> clientsMgr,
                       std::shared_ptr<MetricsAdaptor> metricsAdaptor, std::shared_ptr<Security> security,
                       std::shared_ptr<DomainSocketClient> socketClient)
    : config(librtCfg),
      clientsMgr(clientsMgr),
      metricsAdaptor(metricsAdaptor),
      security_(security),
      socketClient_(socketClient)
{
    invokeOrderMgr = std::make_shared<InvokeOrderManager>(librtCfg);
    messageCoder_ = std::make_shared<MessageCoder>();
}

ErrorInfo Libruntime::Init(std::shared_ptr<FSClient> fsClient, YR::Libruntime::DatasystemClients &datasystemClients,
                           FinalizeCallback cb)
{
    this->runtimeContext = std::make_shared<RuntimeContext>(config->jobId);
    this->dispatcherThread_ = std::make_shared<ThreadPool>();
    this->dispatcherThread_->Init(NUM_DISPATCHER, config->jobId + "." + DISPATCHER);
    this->dsClients.dsObjectStore = datasystemClients.dsObjectStore;
    this->dsClients.dsStreamStore = datasystemClients.dsStreamStore;
    this->dsClients.dsStateStore = datasystemClients.dsStateStore;
    this->dsClients.dsHeteroStore = datasystemClients.dsHeteroStore;
    this->waitingObjectManager = std::make_shared<WaitingObjectManager>(config->checkSignals_);
    this->memStore = std::make_shared<MemoryStore>();
    this->memStore->Init(dsClients.dsObjectStore, waitingObjectManager);
    this->waitingObjectManager->SetMemoryStore(this->memStore);
    this->dependencyResolver = std::make_shared<DependencyResolver>(memStore);
    this->objectIdPool = std::make_shared<ObjectIdPool>(memStore);
    auto mapper = std::make_shared<GeneratorIdMap>();
    this->generatorNotifier_ = std::make_shared<StreamGeneratorNotifier>(datasystemClients.dsStreamStore, mapper);
    this->generatorReceiver_ =
        std::make_shared<StreamGeneratorReceiver>(config, datasystemClients.dsStreamStore, this->memStore);
    this->rGroupManager_ = std::make_shared<ResourceGroupManager>();
    auto functionId = config->functionIds[config->selfLanguage];
    if (functionId.empty()) {
        functionId = config->functionName;
    }
    this->downgrade_ = std::make_shared<YR::scene::DowngradeController>(functionId, clientsMgr, security_);
    TenantAKSKManager::GetInstance().Initialize(clientsMgr, security_, config);
    this->invokeAdaptor = std::make_shared<InvokeAdaptor>(
        config, dependencyResolver, fsClient, memStore, runtimeContext, cb, waitingObjectManager, invokeOrderMgr,
        clientsMgr, metricsAdaptor, mapper, generatorReceiver_, generatorNotifier_, downgrade_);
    invokeAdaptor->SetRGroupManager(rGroupManager_);
    auto setTenantIdCb = [this]() { this->SetTenantIdWithPriority(); };
    invokeAdaptor->SetCallbackOfSetTenantId(setTenantIdCb);
    auto [serverVersion, err] = invokeAdaptor->Init(*runtimeContext, security_);
    if (!err.OK()) {
        return err;
    }
    this->config->serverVersion = serverVersion;
    if (config->logToDriver) {
        this->driverLogReceiver_ = std::make_shared<DriverLogReceiver>();
        this->driverLogReceiver_->Init(datasystemClients.dsStreamStore, config->jobId, config->dedupLogs);
    }
    return ErrorInfo();
}

void Libruntime::FinalizeHandler()
{
    this->Finalize(false);
}

void Libruntime::ReceiveRequestLoop(void)
{
    invokeAdaptor->ReceiveRequestLoop();
    YRLOG_INFO("Request loop exited");
}

std::string Libruntime::GetServerVersion()
{
    return this->config->serverVersion;
}

ErrorInfo Libruntime::CheckSpec(std::shared_ptr<InvokeSpec> spec)
{
    size_t concurrency = DEFAULT_CONCURRENCY;
    auto iter = spec->opts.customExtensions.find(CONCURRENCY);
    if (iter != spec->opts.customExtensions.end()) {
        try {
            concurrency = static_cast<unsigned int>(std::stoull(iter->second));
        } catch (const std::exception &e) {
            return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "invalid opts concurrency" + iter->second);
        }
    }
    if (concurrency > MAX_CONCURRENCY || concurrency < MIN_CONCURRENCY) {
        auto errMsg = "invalid opts concurrency, concurrency: " + iter->second +
                      ", please set the concurrency range between 1 and 1000";
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME, errMsg);
    }
    if (spec->opts.podLabels.size() > MAX_PODLABELS) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                         "The number of pod labels is invalid, please set the pod labels less than and equal to 5");
    }
    for (auto &iter : spec->opts.podLabels) {
        if (!RE2::FullMatch(iter.first, POD_LABELS_KEY_REGEX)) {
            return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                             "The pod label key is invalid, please set the pod label key with letters, digits and '-' "
                             "which cannot start or end with '-' and cannot exceed 63 characters.");
        }
        if (!RE2::FullMatch(iter.second, POD_LABELS_VALUE_REGEX)) {
            return ErrorInfo(
                YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                "The pod label value is invalid, please set the pod label value with letters, digits and '-' which "
                "cannot start or end with '-' and cannot exceed 63 characters. And empty string can also be set as pod "
                "label value too");
        }
    }
    if (spec->opts.customExtensions.find(DELEGATE_DIRECTORY_QUOTA) != spec->opts.customExtensions.end()) {
        auto quota = spec->opts.customExtensions[DELEGATE_DIRECTORY_QUOTA];
        re2::RE2 pattern(R"(^[0-9]+$)");
        if (quota != QUOTA_NO_LIMIT && !RE2::FullMatch(quota, pattern)) {
            return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                             "The DELEGATE_DIRECTORY_QUOTA value: {" + quota + "} is invalid, not composed of numbers");
        }
        if (quota != QUOTA_NO_LIMIT && (std::stoi(quota) > MAX_DELEGATE_DIRECTORY_QUOTA || std::stoi(quota) <= 0)) {
            return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                             "The DELEGATE_DIRECTORY_QUOTA value:{" + quota +
                                 "} is invalid, exceeding the maximum value of 1TB or less than 0M");
        } else {
            spec->opts.customExtensions[DELEGATE_DIRECTORY_QUOTA] = std::to_string(std::stoi(quota));
        }
    } else {
        spec->opts.customExtensions[DELEGATE_DIRECTORY_QUOTA] = DEFAULT_DELEGATE_DIRECTORY_QUOTA;
    }

    if (spec->opts.recoverRetryTimes < 0) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                         "The recoverRetryTimes: {" + std::to_string(spec->opts.recoverRetryTimes) +
                             "} is invalid, which must be non-nagative");
    }
    auto err = CheckInstanceRange(spec);
    if (!err.OK()) {
        return err;
    }
    err = CheckRGroupOpts(spec);
    if (!err.OK()) {
        return err;
    }
    auto insId = spec->GetNamedInstanceId(this->config);
    if (insId.size() > MAX_INS_ID_LENGTH) {
        return ErrorInfo(
            YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
            "The instance ID size is " + std::to_string(insId.size()) + ", exceeds the maximum length of 128 bytes");
    }
    return ErrorInfo();
}

ErrorInfo Libruntime::CheckInstanceRange(std::shared_ptr<InvokeSpec> spec)
{
    if (InstanceRangeEnabled(spec->opts.instanceRange)) {
        if (spec->opts.instanceRange.step <= 0) {
            return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                             "invalid instanceRange step, step is: " + std::to_string(spec->opts.instanceRange.step) +
                                 ", please set the step > 0.");
        }
        if (spec->opts.instanceRange.rangeOpts.timeout < NO_TIMEOUT) {
            return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                             "invalid instanceRange timeout, timeout is: " +
                                 std::to_string(spec->opts.instanceRange.rangeOpts.timeout) +
                                 ", please set the timeout >= -1.");
        }
    }
    return ErrorInfo();
}

ErrorInfo Libruntime::CheckRGroupOpts(std::shared_ptr<InvokeSpec> spec)
{
    if (ResourceGroupEnabled(spec->opts.resourceGroupOpts)) {
        if (spec->opts.resourceGroupOpts.resourceGroupName == std::string(UNSUPPORTED_RGROUP_NAME)) {
            return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                             "invalid resource group name, name: " + spec->opts.resourceGroupOpts.resourceGroupName +
                                 ", please set the name other than primary.");
        }
        if (spec->opts.resourceGroupOpts.bundleIndex == -1) {
            SetResourceGroupAffinity(spec, std::string(RGROUP_NAME), spec->opts.resourceGroupOpts.resourceGroupName);
        } else if (spec->opts.resourceGroupOpts.bundleIndex >= 0) {
            SetResourceGroupAffinity(spec,
                                     std::string(RGROUP_BUNDLE_PREFIX) +
                                         spec->opts.resourceGroupOpts.resourceGroupName +
                                         std::string(RGROUP_BUNDLE_SUFFIX),
                                     std::to_string(spec->opts.resourceGroupOpts.bundleIndex));
        } else {
            return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                             "invalid resource group bundle index, index: " +
                                 std::to_string(spec->opts.resourceGroupOpts.bundleIndex) +
                                 ", please set the index >= -1.");
        }
    }
    return ErrorInfo();
}

void Libruntime::SetResourceGroupAffinity(std::shared_ptr<InvokeSpec> spec, const std::string &key,
                                          const std::string &value)
{
    std::list<std::shared_ptr<LabelOperator>> labelOperators;
    std::shared_ptr<LabelOperator> labelOperator = std::make_shared<LabelInOperator>();
    labelOperator->SetKey(key);
    labelOperator->SetValues({value});
    labelOperators.push_back(labelOperator);
    std::shared_ptr<Affinity> affinity = std::make_shared<ResourceRequiredAffinity>();
    affinity->SetLabelOperators(labelOperators);
    spec->opts.scheduleAffinities.push_back(affinity);
}

ErrorInfo Libruntime::PreProcessArgs(const std::shared_ptr<InvokeSpec> &spec)
{
    ErrorInfo err;
    std::vector<std::string> objIds;
    std::unordered_set<std::string> objIdSet;
    std::unordered_set<std::string> bigObjIdSet;
    std::vector<int> indexs;
    for (unsigned int i = 0; i < spec->invokeArgs.size(); i++) {
        indexs.emplace_back(i);
    }
    sort(indexs.begin(), indexs.end(), [spec](int a, int b) {
        if (spec->invokeArgs[a].dataObj && spec->invokeArgs[b].dataObj) {
            return spec->invokeArgs[a].dataObj->GetSize() < spec->invokeArgs[b].dataObj->GetSize();
        }
        return a < b;
    });

    uint64_t totalSize = 0;
    for (unsigned int i = 0; i < indexs.size(); i++) {
        auto &arg = spec->invokeArgs[indexs[i]];
        objIdSet.insert(arg.nestedObjects.begin(), arg.nestedObjects.end());
        if (arg.isRef) {
            objIdSet.insert(arg.objId);
        }
        uint64_t tmpTotalSize = totalSize + arg.dataObj->GetSize();
        if (tmpTotalSize < arg.dataObj->GetSize()) {
            return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "args size invalid");
        }
        if (tmpTotalSize > uint64_t(Config::Instance().MAX_ARGS_IN_MSG_BYTES())) {
            auto [err, objId] = Put(arg.dataObj, arg.nestedObjects);
            if (err.Code() != ErrorCode::ERR_OK) {
                YRLOG_ERROR("Put arg, code: {}, message: {}", fmt::underlying(err.Code()), err.Msg());
                return err;
            }
            YRLOG_DEBUG("Put arg, object ID: {}", objId);
            InvokeArg bigObjArg{};
            bigObjArg.isRef = true;
            bigObjArg.objId = objId;
            bigObjIdSet.insert(objId);
            arg = std::move(bigObjArg);
        } else {
            totalSize += arg.dataObj->GetSize();
        }
    }
    objIds.assign(objIdSet.begin(), objIdSet.end());
    if (!objIds.empty()) {
        SetTraceId();
        err = memStore->IncreaseObjRef(objIds);
        if (!err.OK()) {
            YRLOG_ERROR("increase ids[{}, ....] failed", objIds[0]);
            return err;
        }
    }
    std::copy(bigObjIdSet.begin(), bigObjIdSet.end(), std::back_inserter(objIds));
    // Used to retain object references in parameters
    memStore->BindObjRefInReq(spec->requestId, objIds);
    return err;
}

std::string Libruntime::ConstructTraceId(const YR::Libruntime::InvokeOptions &opts)
{
    if (opts.traceId.empty()) {
        return YR::utility::IDGenerator::GenTraceId(runtimeContext->GetJobId());
    }
    return opts.traceId;
}

ErrorInfo Libruntime::GenerateReturnObjectIds(const std::string &requestId,
                                              std::vector<YR::Libruntime::DataObject> &returnObjs)
{
    for (size_t i = 0; i < returnObjs.size(); i++) {
        returnObjs[i].id = YR::utility::IDGenerator::GenObjectId(requestId, i);
    }
    return ErrorInfo();
}

std::pair<ErrorInfo, std::string> Libruntime::CreateInstance(const YR::Libruntime::FunctionMeta &functionMeta,
                                                             std::vector<YR::Libruntime::InvokeArg> &invokeArgs,
                                                             YR::Libruntime::InvokeOptions &opts)
{
    std::string requestId = YR::utility::IDGenerator::GenRequestId();
    std::vector<DataObject> returnObjs{DataObject("")};
    auto err = GenerateReturnObjectIds(requestId, returnObjs);
    if (err.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("generate return obj id failed, req id: {}, code: {}, message: {}", requestId,
                    fmt::underlying(err.Code()), err.Msg());
        return std::make_pair(err, "");
    }
    std::string traceId = ConstructTraceId(opts);
    auto spec = std::make_shared<InvokeSpec>(runtimeContext->GetJobId(), functionMeta, returnObjs,
                                             std::move(invokeArgs), libruntime::InvokeType::CreateInstance,
                                             std::move(traceId), std::move(requestId), "", opts);
    err = this->CheckSpec(spec);
    if (err.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("check invoke spec failed, req id: {}, code: {}, message: {}", requestId,
                    fmt::underlying(err.Code()), err.Msg());
        return std::make_pair(err, "");
    }
    err = PreProcessArgs(spec);
    if (err.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("pre process args failed, req id: {}, code: {}, message: {}", requestId,
                    fmt::underlying(err.Code()), err.Msg());
        return std::make_pair(err, "");
    }

    invokeOrderMgr->CreateInstance(spec);
    memStore->AddReturnObject(spec->returnIds);
    dependencyResolver->ResolveDependencies(spec, [this, spec, returnObjs](const ErrorInfo &err) {
        if (err.OK()) {
            AddGeneratorReceiver(spec);
            if (PutRefArgToDs(spec)) {
                if (std::find(spec->opts.labels.begin(), spec->opts.labels.end(), FAAS_INSTANCE_TYPE) ==
                    spec->opts.labels.end()) {
                    spec->opts.labels.push_back(ACTOR_INSTANCE_TYPE);
                }
                spec->BuildInstanceCreateRequest(*config);
                this->invokeAdaptor->CreateInstance(spec);
            }
            return;
        }
        ErrorInfo dependencyErr(YR::Libruntime::ErrorCode::ERR_DEPENDENCY_FAILED, YR::Libruntime::ModuleCode::RUNTIME,
                                "dependency request failed, request id: " + spec->requestId +
                                    ", internal code: " + std::to_string(err.Code()) + ", internal msg: " + err.Msg(),
                                err.GetStackTraceInfos());
        ProcessErr(spec, dependencyErr);

        auto ids = memStore->UnbindObjRefInReq(spec->requestId);
        SetTraceId();
        auto errorInfo = memStore->DecreGlobalReference(ids);
        if (!errorInfo.OK()) {
            YRLOG_WARN("failed to decrease by requestid {}. Code: {}, MCode: {}, Msg: {}", spec->requestId,
                       fmt::underlying(errorInfo.Code()), fmt::underlying(errorInfo.MCode()), errorInfo.Msg());
        }
        invokeOrderMgr->RemoveInstance(spec);
    });
    return std::make_pair(ErrorInfo(), spec->returnIds[0].id);
}

bool Libruntime::PutRefArgToDs(std::shared_ptr<InvokeSpec> spec)
{
    ErrorInfo errInfo;
    for (unsigned int i = 0; i < spec->invokeArgs.size(); i++) {
        auto &arg = spec->invokeArgs[i];
        if (arg.isRef) {
            this->SetTenantId(arg.tenantId);
            SetTraceId();
            errInfo = this->memStore->AlsoPutToDS(arg.objId);
            if (!errInfo.OK()) {
                break;
            }
        }
        if (!arg.nestedObjects.empty()) {
            this->SetTenantId(arg.tenantId);
            SetTraceId();
            errInfo = this->memStore->AlsoPutToDS(arg.nestedObjects);
            if (!errInfo.OK()) {
                break;
            }
        }
    }
    if (!errInfo.OK()) {
        YRLOG_ERROR("put ref arg to ds failed, reqid is {}, err code is {}, err msg is {}", spec->requestId,
                    fmt::underlying(errInfo.Code()), errInfo.Msg());
        ProcessErr(spec, errInfo);
        return false;
    }
    return true;
}

ErrorInfo Libruntime::InvokeByInstanceId(const YR::Libruntime::FunctionMeta &funcMeta, const std::string &instanceId,
                                         std::vector<YR::Libruntime::InvokeArg> &invokeArgs,
                                         YR::Libruntime::InvokeOptions &opts, std::vector<DataObject> &returnObjs)
{
    std::string requestId = YR::utility::IDGenerator::GenRequestId();
    auto err = GenerateReturnObjectIds(requestId, returnObjs);
    if (err.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("generate return obj id failed, req id: {}, error code: {}, error message: {}", requestId,
                    fmt::underlying(err.Code()), err.Msg());
        return err;
    }
    std::string traceId = ConstructTraceId(opts);
    YRLOG_DEBUG("Invoke func: {}, instanceId: {}, request id: {}, trace id: {}", funcMeta.funcName, instanceId,
                requestId, traceId);
    auto spec = std::make_shared<InvokeSpec>(runtimeContext->GetJobId(), funcMeta, returnObjs, std::move(invokeArgs),
                                             libruntime::InvokeType::InvokeFunction, std::move(traceId),
                                             std::move(requestId), instanceId, opts);
    if (opts.forceInvoke) {
        auto invokeOptions = spec->requestInvoke->Mutable().mutable_invokeoptions();
        auto customTag = invokeOptions->mutable_customtag();
        customTag->insert({"ENABLE_FORCE_INVOKE", ""});
    }
    err = PreProcessArgs(spec);
    if (err.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("pre process failed, req id: {}, code: {}, message: {}", spec->requestId,
                    fmt::underlying(err.Code()), err.Msg());
        return err;
    }

    memStore->AddReturnObject(returnObjs);
    if (!this->config->inCluster) {
        std::vector<std::string> objIds;
        for (const auto &obj : returnObjs) {
            if (!obj.id.empty()) {
                objIds.push_back(obj.id);
            }
        }
        YRLOG_DEBUG("start increase ds global reference, req id is {}, obj ids: [{}, ...]", spec->requestId, objIds[0]);
        auto errInfo = memStore->IncreDSGlobalReference(objIds);
        if (!errInfo.OK()) {
            YRLOG_ERROR("failed to increase ds global reference, req id is {}, error code is {}, error msg is {}",
                        spec->requestId, fmt::underlying(errInfo.Code()), errInfo.Msg());
        }
    }
    invokeOrderMgr->Invoke(spec);
    this->invokeAdaptor->PushInvokeSpec(spec);
    auto func = [this, spec, returnObjs](const ErrorInfo &err) {
        if (err.OK()) {
            invokeOrderMgr->UpdateUnfinishedSeq(spec);
            AddGeneratorReceiver(spec);
            if (PutRefArgToDs(spec)) {
                auto namedId = spec->GetNamedInstanceId(this->config);
                if (spec->opts.isDeleteRemoteTensor) {
                    spec->invokeInstanceId = spec->instanceId;
                    spec->invokeType = libruntime::InvokeType::DeleteRemoteTensor;
                } else if (namedId.empty()) {
                    spec->invokeInstanceId = memStore->GetInstanceId(spec->instanceId);
                } else {
                    spec->invokeInstanceId = memStore->GetInstanceId(namedId);
                }
                spec->instanceRoute = memStore->GetInstanceRoute(spec->instanceId);
                spec->BuildInstanceInvokeRequest(*config);
                invokeAdaptor->InvokeInstanceFunction(spec);
            }
            return;
        }
        if (err.IsCreate()) {
            ErrorInfo dependencyErr(err.Code(), err.MCode(),
                                    "dependency instance create failed, request id: " + spec->requestId +
                                        ", internal code: " + std::to_string(err.Code()) +
                                        ", internal msg: " + err.Msg(),
                                    err.GetStackTraceInfos());
            ProcessErr(spec, dependencyErr);
        } else {
            ErrorInfo dependencyErr(
                YR::Libruntime::ErrorCode::ERR_DEPENDENCY_FAILED, YR::Libruntime::ModuleCode::RUNTIME,
                "dependency request failed, request id: " + spec->requestId +
                    ", internal code: " + std::to_string(err.Code()) + ", internal msg: " + err.Msg(),
                err.GetStackTraceInfos());
            ProcessErr(spec, dependencyErr);
        }
        auto ids = memStore->UnbindObjRefInReq(spec->requestId);
        auto errorInfo = memStore->DecreGlobalReference(ids);
        if (!errorInfo.OK()) {
            YRLOG_WARN("failed to decrease by requestid {}. Code: {}, MCode: {}, Msg: {}", spec->requestId,
                       fmt::underlying(errorInfo.Code()), fmt::underlying(errorInfo.MCode()), errorInfo.Msg());
        }
    };
    dependencyResolver->ResolveDependencies(spec, [func, dispatcherThread(dispatcherThread_)](const ErrorInfo &err) {
        if (dispatcherThread != nullptr) {
            dispatcherThread->Handle([func, err]() { func(err); }, "");
        }
    });
    return ErrorInfo();
}

std::string Libruntime::GetRealInstanceId(const std::string &objectId, int timeout)
{
    return memStore->GetInstanceId(objectId);
}

void Libruntime::SaveRealInstanceId(const std::string &objectId, const std::string &instanceId)
{
    memStore->AddReturnObject(objectId);
    memStore->SetInstanceId(objectId, instanceId);
    memStore->SetReady(objectId);
}

void Libruntime::SaveRealInstanceId(const std::string &objectId, const std::string &instanceId,
                                    const InstanceOptions &opts)
{
    memStore->AddReturnObject(objectId);
    memStore->SetInstanceId(objectId, instanceId);
    memStore->SetReady(objectId);
    if (opts.needOrder) {
        invokeOrderMgr->RegisterInstance(objectId);
    }
}

std::string Libruntime::GetGroupInstanceIds(const std::string &objectId, int timeout)
{
    auto [instanceIds, err] = memStore->GetInstanceIds(objectId, timeout);
    if (!err.OK()) {
        YRLOG_WARN("get group instance ids failed, error code: {}, error message: {}", fmt::underlying(err.Code()),
                   err.Msg());
        return "";
    }
    return YR::utility::Join(instanceIds, ";");
}

void Libruntime::SaveGroupInstanceIds(const std::string &objectId, const std::string &groupInsIds,
                                      const InstanceOptions &opts)
{
    std::vector<std::string> instanceIds;
    YR::utility::Split(groupInsIds, instanceIds, ';');
    for (size_t i = 0; i < instanceIds.size(); ++i) {
        YRLOG_DEBUG("save instance_{}, instance id is {}", i, instanceIds.at(i));
        if (opts.needOrder) {
            invokeOrderMgr->CreateGroupInstance(instanceIds.at(i));
        }
        memStore->AddReturnObject(instanceIds.at(i));
        memStore->SetInstanceId(instanceIds.at(i), instanceIds.at(i));
        memStore->SetReady(instanceIds.at(i));
    }
    memStore->AddReturnObject(objectId);
    memStore->SetInstanceIds(objectId, instanceIds);
    memStore->SetReady(objectId);
}

ErrorInfo Libruntime::InvokeByFunctionName(const YR::Libruntime::FunctionMeta &funcMeta,
                                           std::vector<YR::Libruntime::InvokeArg> &invokeArgs,
                                           YR::Libruntime::InvokeOptions &opts, std::vector<DataObject> &returnObjs)
{
    std::string requestId = YR::utility::IDGenerator::GenRequestId();
    auto err = GenerateReturnObjectIds(requestId, returnObjs);
    if (err.Code() != ErrorCode::ERR_OK) {
        return err;
    }
    std::string traceId = ConstructTraceId(opts);
    YRLOG_DEBUG("start invoke stateless function, request id: {}, obj id: {}, trace id: {}", requestId,
                returnObjs[0].id, traceId);
    auto spec = std::make_shared<InvokeSpec>(runtimeContext->GetJobId(), funcMeta, returnObjs, std::move(invokeArgs),
                                             libruntime::InvokeType::InvokeFunctionStateless, std::move(traceId),
                                             std::move(requestId), "", opts);
    err = this->CheckSpec(spec);
    if (err.Code() != ErrorCode::ERR_OK) {
        return err;
    }
    err = PreProcessArgs(spec);
    if (err.Code() != ErrorCode::ERR_OK) {
        return err;
    }
    if (this->config->enableEvent) {
        YRLOG_DEBUG("start to create timer for invocation, req id is {}", requestId);
        memStore->AddEventTimer(requestId, opts.timeout);
    }
    memStore->AddReturnObject(returnObjs);
    if (!this->config->inCluster) {
        std::vector<std::string> objIds;
        for (const auto &obj : returnObjs) {
            if (!obj.id.empty()) {
                objIds.push_back(obj.id);
            }
        }
        YRLOG_DEBUG("start increase ds global reference, req id is {} , obj ids: [{}, ...]", spec->requestId,
                    objIds[0]);
        memStore->IncreDSGlobalReference(objIds);
    }
    this->invokeAdaptor->PushInvokeSpec(spec);
    auto func = [this, spec](const ErrorInfo &err) {
        if (err.OK()) {
            AddGeneratorReceiver(spec);
            if (PutRefArgToDs(spec)) {
                spec->BuildInstanceInvokeRequest(*config);
                this->invokeAdaptor->SubmitFunction(spec);
            }
            return;
        }
        ErrorInfo dependencyErr(YR::Libruntime::ErrorCode::ERR_DEPENDENCY_FAILED, YR::Libruntime::ModuleCode::RUNTIME,
                                "dependency request failed, request id: " + spec->requestId +
                                    ", internal code: " + std::to_string(err.Code()) + ", internal msg: " + err.Msg(),
                                err.GetStackTraceInfos());
        ProcessErr(spec, dependencyErr);
        auto ids = memStore->UnbindObjRefInReq(spec->requestId);
        auto errorInfo = memStore->DecreGlobalReference(ids);
        if (!errorInfo.OK()) {
            YRLOG_WARN("failed to decrease by requestid {}. Code: {}, MCode: {}, Msg: {}", spec->requestId,
                       fmt::underlying(errorInfo.Code()), fmt::underlying(errorInfo.MCode()), errorInfo.Msg());
        }
    };
    dependencyResolver->ResolveDependencies(spec, [func, dispatcherThread(dispatcherThread_)](const ErrorInfo &err) {
        if (dispatcherThread != nullptr) {
            dispatcherThread->Handle([func, err]() { func(err); }, "");
        }
    });
    return ErrorInfo();
}

void Libruntime::ProcessErr(const std::shared_ptr<InvokeSpec> &spec, const ErrorInfo &errInfo)
{
    if (spec->functionMeta.isGenerator && generatorReceiver_) {
        generatorReceiver_->MarkEndOfStream(spec->returnIds[0].id, errInfo);
    }
    memStore->SetError(spec->returnIds, errInfo);
}

void Libruntime::AddGeneratorReceiver(std::shared_ptr<InvokeSpec> spec)
{
    if (spec->functionMeta.isGenerator && generatorReceiver_) {
        generatorReceiver_->Initialize();
        generatorReceiver_->AddRecord(spec->returnIds[0].id);
    }
}

void Libruntime::CreateInstanceRaw(std::shared_ptr<Buffer> reqRaw, RawCallback cb)
{
    this->invokeAdaptor->CreateInstanceRaw(reqRaw, cb);
}

void Libruntime::InvokeByInstanceIdRaw(std::shared_ptr<Buffer> reqRaw, RawCallback cb)
{
    this->invokeAdaptor->InvokeByInstanceIdRaw(reqRaw, cb);
}

void Libruntime::KillRaw(std::shared_ptr<Buffer> reqRaw, RawCallback cb)
{
    this->invokeAdaptor->KillRaw(reqRaw, cb);
}

std::pair<ErrorInfo, std::string> Libruntime::Put(std::shared_ptr<DataObject> dataobj,
                                                  const std::unordered_set<std::string> &nestedIds,
                                                  const CreateParam &createParam)
{
    // small data -> MemoryStore
    // Get an id from pool
    auto [err, objId] = objectIdPool->Pop();
    if (!err.OK()) {
        return std::make_pair(err, objId);
    }
    SetTraceId();
    err = memStore->Put(dataobj->buffer, objId, nestedIds, createParam);
    return std::make_pair(err, objId);
}

ErrorInfo Libruntime::Put(const std::string &objId, std::shared_ptr<DataObject> dataObj,
                          const std::unordered_set<std::string> &nestedId, const CreateParam &createParam)
{
    SetTraceId();
    return memStore->Put(dataObj->buffer, objId, nestedId, createParam);
}

ErrorInfo Libruntime::Put(std::shared_ptr<Buffer> data, const std::string &objID,
                          const std::unordered_set<std::string> &nestedID, bool toDataSystem,
                          const CreateParam &createParam)
{
    SetTraceId();
    auto err = memStore->Put(data, objID, nestedID, toDataSystem, createParam);
    if (!err.OK()) {
        return err;
    }
    memStore->SetReady(objID);
    return err;
}

ErrorInfo Libruntime::PutRaw(const std::string &objId, std::shared_ptr<Buffer> data,
                             const std::unordered_set<std::string> &nestedId, const CreateParam &createParam)
{
    if (!dsClients.dsObjectStore) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "PutRaw dsClients.dsObjectStore is nullptr!");
    }
    SetTraceId();
    return dsClients.dsObjectStore->Put(data, objId, nestedId, createParam);
}

ErrorInfo Libruntime::IncreaseReference(const std::vector<std::string> &objIds, bool toDatasystem)
{
    SetTraceId();
    return memStore->IncreGlobalReference(objIds, toDatasystem);
}

std::pair<ErrorInfo, std::vector<std::string>> Libruntime::IncreaseReference(const std::vector<std::string> &objIds,
                                                                             const std::string &remoteId)
{
    SetTraceId();
    return memStore->IncreGlobalReference(objIds, remoteId);
}

ErrorInfo Libruntime::IncreaseReferenceRaw(const std::vector<std::string> &objIds)
{
    if (objIds.empty()) {
        return ErrorInfo();
    }
    if (!dsClients.dsObjectStore) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "IncreaseReferenceRaw dsObjectStore is nullptr!");
    }
    SetTraceId();
    return dsClients.dsObjectStore->IncreGlobalReference(objIds);
}

std::pair<ErrorInfo, std::vector<std::string>> Libruntime::IncreaseReferenceRaw(const std::vector<std::string> &objIds,
                                                                                const std::string &remoteId)
{
    if (objIds.empty()) {
        return std::make_pair(ErrorInfo(), std::vector<std::string>());
    }
    if (!dsClients.dsObjectStore) {
        return std::make_pair(
            ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "IncreaseReferenceRaw dsObjectStore is nullptr!"),
            std::vector<std::string>());
    }
    SetTraceId();
    return dsClients.dsObjectStore->IncreGlobalReference(objIds, remoteId);
}

void Libruntime::DecreaseReference(const std::vector<std::string> &objIds)
{
    if (memStore == nullptr) {
        std::cerr << "Libruntime::DecreaseReference memStore is nullptr." << std::endl;
        return;
    }
    SetTraceId();
    ErrorInfo err = memStore->DecreGlobalReference(objIds);
    if (err.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("ErrCode: {}, ModuleCode: {}, ErrMsg: {}", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg());
    }
    return;
}

std::pair<ErrorInfo, std::vector<std::string>> Libruntime::DecreaseReference(const std::vector<std::string> &objIds,
                                                                             const std::string &remoteId)
{
    SetTraceId();
    return memStore->DecreGlobalReference(objIds, remoteId);
}

void Libruntime::DecreaseReferenceRaw(const std::vector<std::string> &objIds)
{
    if (objIds.empty()) {
        return;
    }
    if (!dsClients.dsObjectStore) {
        YRLOG_ERROR("DecreaseReferenceRaw dsObjectStore is nullptr!");
        return;
    }
    SetTraceId();
    ErrorInfo err = dsClients.dsObjectStore->DecreGlobalReference(objIds);
    if (err.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("ErrCode: {}, ModuleCode: {}, ErrMsg: {}", fmt::underlying(err.Code()),
                    fmt::underlying(err.MCode()), err.Msg());
    }
    return;
}

std::pair<ErrorInfo, std::vector<std::string>> Libruntime::DecreaseReferenceRaw(const std::vector<std::string> &objIds,
                                                                                const std::string &remoteId)
{
    if (objIds.empty()) {
        return std::make_pair(ErrorInfo(), std::vector<std::string>());
    }
    if (!dsClients.dsObjectStore) {
        return std::make_pair(
            ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "DecreaseReferenceRaw dsObjectStore is nullptr!"),
            std::vector<std::string>());
    }
    SetTraceId();
    return dsClients.dsObjectStore->DecreGlobalReference(objIds, remoteId);
}

ErrorInfo Libruntime::ReleaseGRefs(const std::string &remoteId)
{
    SetTraceId();
    return memStore->ReleaseGRefs(remoteId);
}

// timeout < 0 : wait without timeout
std::shared_ptr<YR::InternalWaitResult> Libruntime::Wait(const std::vector<std::string> &objs, std::size_t waitNum,
                                                         int timeoutSec)
{
    int64_t timeoutMs = timeoutSec != NO_TIMEOUT ? timeoutSec * S_TO_MS : NO_TIMEOUT;
    return waitingObjectManager->WaitUntilReady(objs, waitNum, timeoutMs);
}

template <typename T>
std::pair<bool, std::string> CheckObjPartialResult(const std::vector<std::string> &ids,
                                                   std::vector<std::shared_ptr<T>> results, const ErrorInfo &errInfo,
                                                   int timeoutMs)
{
    std::vector<std::string> failIds;
    std::pair<bool, std::string> ret;
    bool isPartialResult = false;
    for (unsigned int i = 0; i < results.size(); i++) {
        if (results[i] == nullptr) {
            isPartialResult = true;
            failIds.push_back(ids[i]);
        }
    }
    if (isPartialResult) {
        ret.second = errInfo.GetExceptionMsg(failIds, timeoutMs);
    }
    ret.first = isPartialResult;
    return ret;
}

std::pair<ErrorInfo, int64_t> Libruntime::WaitBeforeGet(const std::vector<std::string> &ids, int timeoutMs,
                                                        bool allowPartial)
{
    auto beginningTs = GetCurrentTimestampMs();
    auto waitRes = waitingObjectManager->WaitUntilReady(ids, ids.size(), timeoutMs);
    if (waitRes->readyIds.size() == ids.size() || (waitRes->readyIds.size() > 0 && allowPartial)) {
        auto currentTs = GetCurrentTimestampMs();
        int64_t remainingTimePeriod = 0;
        if (timeoutMs == NO_TIMEOUT) {
            remainingTimePeriod = timeoutMs;
        } else if (beginningTs + timeoutMs > currentTs) {
            remainingTimePeriod = beginningTs + timeoutMs - currentTs;
        } else {
            remainingTimePeriod = 0;
        }
        return std::make_pair(ErrorInfo(), remainingTimePeriod);
    }

    ErrorInfo err;
    if (!waitRes->exceptionIds.empty()) {
        err = waitRes->exceptionIds.begin()->second;
        YRLOG_ERROR("WaitBeforeGet gets exceptionIds. exceptionIds: {}",
                    YR::utility::Join(std::vector<std::string>{waitRes->exceptionIds.begin()->first}, "..."));
    } else {
        err.SetErrorCode(YR::Libruntime::ErrorCode::ERR_GET_OPERATION_FAILED);
        err.SetErrorMsg("Get object timeout. allowPartial = " + std::to_string(allowPartial) + " Failed objects: [ " +
                        YR::utility::Join(waitRes->unreadyIds.empty()
                                              ? std::vector<std::string>{}
                                              : std::vector<std::string>{waitRes->unreadyIds[0]},
                                          "...") +
                        " ]");
        err.SetIsTimeout(true);
    }
    return std::make_pair(err, 0);
}

std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> Libruntime::MakeGetResult(
    MultipleResult &res, const std::vector<std::string> &ids, int timeoutMs, bool allowPartial)
{
    // all ids fail,need to throw Exception
    if (res.first.Code() == YR::Libruntime::ErrorCode::ERR_OK) {
        std::pair<bool, std::string> checkObjPartialResult =
            CheckObjPartialResult(ids, res.second, res.first, timeoutMs);
        // partial ids fail, need to throw Exception optionally
        if (!allowPartial && checkObjPartialResult.first) {
            // update exception msg
            res.first.SetErrCodeAndMsg(YR::Libruntime::ErrorCode::ERR_GET_OPERATION_FAILED,
                                       YR::Libruntime::ModuleCode::RUNTIME, checkObjPartialResult.second);
        }
    }
    return res;
}

// allowPartial=true means Get will return OK when getting partial object ref success
// allowPartial=false means Get will return ErrorInfo even if getting partial object ref success

std::pair<ErrorInfo, std::vector<std::shared_ptr<DataObject>>> Libruntime::Get(const std::vector<std::string> &ids,
                                                                               int timeoutMs, bool allowPartial)
{
    auto [err, remainingTimePeriod] = WaitBeforeGet(ids, timeoutMs, allowPartial);
    if (!err.OK()) {
        return std::make_pair(err, std::vector<std::shared_ptr<DataObject>>{});
    }
    SetTraceId();
    MultipleResult res = memStore->Get(ids, remainingTimePeriod);
    res = MakeGetResult(res, ids, timeoutMs, allowPartial);
    std::vector<std::shared_ptr<DataObject>> result(ids.size());
    for (size_t i = 0; i < res.second.size(); i++) {
        if (res.second[i]) {
            auto dataObj = std::make_shared<DataObject>(ids[i], res.second[i]);
            result[i] = dataObj;
        }
    }
    return std::make_pair(res.first, result);
}

std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> Libruntime::GetRaw(const std::vector<std::string> &ids,
                                                                              int timeoutMs, bool allowPartial)
{
    if (!dsClients.dsObjectStore) {
        return std::make_pair(ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, "GetRaw dsObjectStore is nullptr!"),
                              std::vector<std::shared_ptr<Buffer>>());
    }
    SetTraceId();
    MultipleResult res = dsClients.dsObjectStore->Get(ids, timeoutMs);
    return MakeGetResult(res, ids, timeoutMs, allowPartial);
}

ErrorInfo Libruntime::AllocReturnObject(std::shared_ptr<DataObject> &returnObj, size_t metaSize, size_t dataSize,
                                        const std::vector<std::string> &nestedObjIds, uint64_t &totalNativeBufferSize)
{
    return this->AllocReturnObject(returnObj.get(), metaSize, dataSize, nestedObjIds, totalNativeBufferSize);
}

ErrorInfo Libruntime::AllocReturnObject(DataObject *returnObj, size_t metaSize, size_t dataSize,
                                        const std::vector<std::string> &nestedObjIds, uint64_t &totalNativeBufferSize)
{
    std::shared_ptr<Buffer> dataBuf;
    if (metaSize == 0) {
        metaSize = MetaDataLen;
    }
    auto bufferSize = metaSize + dataSize;
    if (returnObj->alwaysNative ||
        (nestedObjIds.empty() &&
         bufferSize + totalNativeBufferSize < YR::Libruntime::Config::Instance().MEM_STORE_SIZE_THRESHOLD())) {
        totalNativeBufferSize += bufferSize;
        dataBuf = std::make_shared<NativeBuffer>(bufferSize);
    } else {
        SetTraceId();
        auto err = memStore->IncreGlobalReference({returnObj->id}, true);
        if (!err.OK()) {
            return err;
        }
        ErrorInfo dsErr = memStore->AlsoPutToDS(nestedObjIds);
        if (dsErr.Code() != ErrorCode::ERR_OK) {
            YRLOG_ERROR("AlsoPutToDS for nestedIDs error.");
            return dsErr;
        }
        err = CreateBuffer(returnObj->id, bufferSize, dataBuf);
        if (!err.OK()) {
            YRLOG_ERROR("Failed to create return value, object Id: {}, data size: {}, ec : {}, err message: {}.",
                        returnObj->id, dataSize, fmt::underlying(err.Code()), err.Msg());
            return err;
        }
    }
    if (dataBuf) {
        returnObj->SetBuffer(dataBuf);
        returnObj->SetNestedIds(nestedObjIds);
        YRLOG_DEBUG("Succeed to alloc return object buffer, object Id: {}, data size: {}", returnObj->id, dataSize);
        return ErrorInfo();
    }
    YRLOG_ERROR("Empty return object buffer, object Id: {}, data size: {}", returnObj->id, dataSize);
    return ErrorInfo(ErrorCode::ERR_CREATE_RETURN_BUFFER, "data buffer empty");
}

ErrorInfo Libruntime::CreateBuffer(const std::string &objectId, size_t dataSize, std::shared_ptr<Buffer> &dataBuf)
{
    SetTraceId();
    return memStore->CreateBuffer(objectId, dataSize, dataBuf);
}

std::pair<ErrorInfo, std::string> Libruntime::CreateBuffer(size_t dataSize, std::shared_ptr<Buffer> &dataBuf)
{
    // small data -> MemoryStore
    // Get an id from pool
    SetTraceId();
    auto [err, objectId] = objectIdPool->Pop();
    if (!err.OK()) {
        return std::make_pair(err, objectId);
    }
    err = CreateBuffer(objectId, dataSize, dataBuf);
    return std::make_pair(err, objectId);
}

std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> Libruntime::GetBuffers(const std::vector<std::string> &ids,
                                                                                  int timeoutMs, bool allowPartial)
{
    auto [errWait, remainingTimePeriod] = WaitBeforeGet(ids, timeoutMs, allowPartial);
    if (!errWait.OK()) {
        YRLOG_DEBUG("Failed to WaitBeforeGet, ids: {}, code: {}, message: {}",
                    YR::utility::Join(std::vector<std::string>{ids[0]}, "..."), fmt::underlying(errWait.Code()),
                    errWait.Msg());
        return std::make_pair(errWait, std::vector<std::shared_ptr<Buffer>>{});
    }
    SetTraceId();
    auto [err, results] = memStore->GetBuffers(ids, remainingTimePeriod);
    if (err.Code() == YR::Libruntime::ErrorCode::ERR_OK) {
        YRLOG_DEBUG("Succeeded to GetBuffers, ids:{}, ids size: {}, results size: {}",
                    YR::utility::Join(std::vector<std::string>{ids[0]}, "..."), ids.size(), results.size());
        std::pair<bool, std::string> checkObjPartialResult =
            CheckObjPartialResult(ids, results, err, remainingTimePeriod);
        // partial ids fail, need to throw Exception optionally
        if (!allowPartial && checkObjPartialResult.first) {
            // update exception msg
            err.SetErrCodeAndMsg(YR::Libruntime::ErrorCode::ERR_GET_OPERATION_FAILED,
                                 YR::Libruntime::ModuleCode::RUNTIME, checkObjPartialResult.second);
        }
    } else {
        YRLOG_ERROR("Failed to GetBuffers, ids: {}, code: {}, message: {}",
                    YR::utility::Join(std::vector<std::string>{ids[0]}, "..."), fmt::underlying(err.Code()), err.Msg());
    }

    return std::make_pair(err, results);
}

std::pair<RetryInfo, std::vector<std::shared_ptr<DataObject>>> Libruntime::GetDataObjectsWithoutWait(
    const std::vector<std::string> &ids, int timeoutMS)
{
    auto [retryInfo, getBuffers] = this->GetBuffersWithoutWait(ids, timeoutMS);
    std::vector<std::shared_ptr<DataObject>> dataObjects(ids.size());
    for (size_t i = 0; i < getBuffers.size(); i++) {
        if (getBuffers[i]) {
            auto dataObj = std::make_shared<DataObject>(ids[i], getBuffers[i]);
            dataObjects[i] = dataObj;
        }
    }
    return std::make_pair(retryInfo, dataObjects);
}

std::pair<RetryInfo, std::vector<std::shared_ptr<Buffer>>> Libruntime::GetBuffersWithoutWait(
    const std::vector<std::string> &ids, int timeoutMS)
{
    SetTraceId();
    return memStore->GetBuffersWithoutRetry(ids, timeoutMS);
}

std::pair<ErrorInfo, std::string> Libruntime::CreateDataObject(size_t metaSize, size_t dataSize,
                                                               std::shared_ptr<DataObject> &dataObj,
                                                               const std::vector<std::string> &nestedObjIds,
                                                               const CreateParam &createParam)
{
    SetTraceId();
    auto [err, objId] = objectIdPool->Pop();
    if (!err.OK()) {
        return std::make_pair(err, objId);
    }
    err = CreateDataObject(objId, metaSize, dataSize, dataObj, nestedObjIds, createParam);
    return std::make_pair(err, objId);
}

ErrorInfo Libruntime::CreateDataObject(const std::string &objId, size_t metaSize, size_t dataSize,
                                       std::shared_ptr<DataObject> &dataObj,
                                       const std::vector<std::string> &nestedObjIds, const CreateParam &createParam)
{
    for (const auto &nestedId : nestedObjIds) {
        if (nestedId == objId) {
            return ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "check circular references detected, obj id: " + objId);
        }
    }
    if (nestedObjIds.size() != 0) {
        auto ret = Wait(nestedObjIds, nestedObjIds.size(), DEFAULT_TIMEOUT_SEC);
        if (!ret->unreadyIds.empty() || !ret->exceptionIds.empty()) {
            return ErrorInfo(ErrorCode::ERR_USER_FUNCTION_EXCEPTION, "wait nested objects timeout or exception");
        }
    }
    SetTraceId();
    ErrorInfo dsErr = memStore->AlsoPutToDS(nestedObjIds, createParam);
    if (dsErr.Code() != ErrorCode::ERR_OK) {
        YRLOG_ERROR("put nested obj to datasystem error");
        return dsErr;
    }
    std::shared_ptr<Buffer> buf;
    if (metaSize == 0) {
        metaSize = MetaDataLen;
    }
    if (WillSizeOverFlow(metaSize, dataSize)) {
        return ErrorInfo(
            ErrorCode::ERR_INNER_SYSTEM_ERROR,
            "data size overflow, metaSize: " + std::to_string(metaSize) + ", dataSize: " + std::to_string(dataSize));
    }
    auto err = memStore->CreateBuffer(objId, metaSize + dataSize, buf, createParam);
    if (!err.OK()) {
        YRLOG_ERROR("Failed to create dataObject, object Id: {}, data size: {}, code: {}, message: {}.", dataObj->id,
                    dataSize, fmt::underlying(err.Code()), err.Msg());
        return err;
    }
    if (buf) {
        dataObj->SetBuffer(buf);
        dataObj->SetNestedIds(nestedObjIds);
        return ErrorInfo();
    }
    YRLOG_ERROR("Empty return object buffer, object Id: {}, data size: {}", dataObj->id, dataSize);
    return ErrorInfo(ErrorCode::ERR_CREATE_RETURN_BUFFER, "data buffer empty");
}

std::pair<ErrorInfo, std::vector<std::shared_ptr<DataObject>>> Libruntime::GetDataObjects(
    const std::vector<std::string> &ids, int timeoutMs, bool allowPartial)
{
    SetTraceId();
    auto [err, buffers] = GetBuffers(ids, timeoutMs, allowPartial);
    if (!err.OK()) {
        YRLOG_DEBUG("Failed to GetDataObjects, ids: {}, code: {}, message: {}",
                    YR::utility::Join(std::vector<std::string>{ids[0]}, "..."), fmt::underlying(err.Code()), err.Msg());
        return std::make_pair(err, std::vector<std::shared_ptr<DataObject>>{});
    }
    std::vector<std::shared_ptr<DataObject>> result(ids.size());
    for (size_t i = 0; i < buffers.size(); i++) {
        if (buffers[i]) {
            auto dataObj = std::make_shared<DataObject>(ids[i], buffers[i]);
            result[i] = dataObj;
        }
    }
    return std::make_pair(err, result);
}

ErrorInfo Libruntime::KVWrite(const std::string &key, std::shared_ptr<Buffer> value, SetParam setParam)
{
    SetTraceId();
    return dsClients.dsStateStore->Write(key, value, setParam);
}

ErrorInfo Libruntime::KVMSetTx(const std::vector<std::string> &keys, const std::vector<std::shared_ptr<Buffer>> &vals,
                               const MSetParam &mSetParam)
{
    SetTraceId();
    return dsClients.dsStateStore->MSetTx(keys, vals, mSetParam);
}

SingleReadResult Libruntime::KVRead(const std::string &key, int timeoutMS)
{
    SetTraceId();
    return dsClients.dsStateStore->Read(key, timeoutMS);
}

MultipleReadResult Libruntime::KVRead(const std::vector<std::string> &keys, int timeoutMS, bool allowPartial)
{
    SetTraceId();
    return dsClients.dsStateStore->Read(keys, timeoutMS, allowPartial);
}

MultipleReadResult Libruntime::KVGetWithParam(const std::vector<std::string> &keys, const GetParams &params,
                                              int timeoutMs)
{
    SetTraceId();
    return dsClients.dsStateStore->GetWithParam(keys, params, timeoutMs);
}

ErrorInfo Libruntime::KVDel(const std::string &key)
{
    SetTraceId();
    return dsClients.dsStateStore->Del(key);
}

MultipleDelResult Libruntime::KVDel(const std::vector<std::string> &keys)
{
    SetTraceId();
    return dsClients.dsStateStore->Del(keys);
}

MultipleExistResult Libruntime::KVExist(const std::vector<std::string> &keys)
{
    return dsClients.dsStateStore->Exist(keys);
}

ErrorInfo Libruntime::DevDelete(const std::vector<std::string> &objectIds, std::vector<std::string> &failedObjectIds)
{
    SetTraceId();
    return dsClients.dsHeteroStore->DevDelete(objectIds, failedObjectIds);
}

ErrorInfo Libruntime::DevLocalDelete(const std::vector<std::string> &objectIds,
                                     std::vector<std::string> &failedObjectIds)
{
    SetTraceId();
    return dsClients.dsHeteroStore->DevLocalDelete(objectIds, failedObjectIds);
}

ErrorInfo Libruntime::DevSubscribe(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                                   std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec)
{
    SetTraceId();
    return dsClients.dsHeteroStore->DevSubscribe(keys, blob2dList, futureVec);
}

ErrorInfo Libruntime::DevPublish(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                                 std::vector<std::shared_ptr<YR::Libruntime::HeteroFuture>> &futureVec)
{
    SetTraceId();
    return dsClients.dsHeteroStore->DevPublish(keys, blob2dList, futureVec);
}

ErrorInfo Libruntime::DevMSet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                              std::vector<std::string> &failedKeys)
{
    SetTraceId();
    return dsClients.dsHeteroStore->DevMSet(keys, blob2dList, failedKeys);
}

ErrorInfo Libruntime::DevMGet(const std::vector<std::string> &keys, const std::vector<DeviceBlobList> &blob2dList,
                              std::vector<std::string> &failedKeys, int32_t timeoutSec)
{
    SetTraceId();
    return dsClients.dsHeteroStore->DevMGet(keys, blob2dList, failedKeys, timeoutSec * S_TO_MS);
}

ErrorInfo Libruntime::CreateStreamProducer(const std::string &streamName, ProducerConf producerConf,
                                           std::shared_ptr<StreamProducer> &producer)
{
    producer = std::make_shared<StreamProducer>();
    if (dsClients.dsStreamStore == nullptr) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "the datasystem stream store is empty, outcluster operations are not supported by now");
    }
    SetTraceId(producerConf.traceId);
    return dsClients.dsStreamStore->CreateStreamProducer(streamName, producer, producerConf);
}

ErrorInfo Libruntime::CreateStreamConsumer(const std::string &streamName, const SubscriptionConfig &config,
                                           std::shared_ptr<StreamConsumer> &consumer, bool autoAck)
{
    consumer = std::make_shared<StreamConsumer>();
    if (dsClients.dsStreamStore == nullptr) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "the datasystem stream store is empty, outcluster operations are not supported by now");
    }
    SetTraceId(config.traceId);
    return dsClients.dsStreamStore->CreateStreamConsumer(streamName, config, consumer, autoAck);
}

ErrorInfo Libruntime::DeleteStream(const std::string &streamName)
{
    if (dsClients.dsStreamStore == nullptr) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "the datasystem stream store is empty, outcluster operations are not supported by now");
    }
    SetTraceId();
    return dsClients.dsStreamStore->DeleteStream(streamName);
}

ErrorInfo Libruntime::QueryGlobalProducersNum(const std::string &streamName, uint64_t &gProducerNum)
{
    if (dsClients.dsStreamStore == nullptr) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "the datasystem stream store is empty, outcluster operations are not supported by now");
    }
    SetTraceId();
    return dsClients.dsStreamStore->QueryGlobalProducersNum(streamName, gProducerNum);
}

ErrorInfo Libruntime::QueryGlobalConsumersNum(const std::string &streamName, uint64_t &gConsumerNum)
{
    if (dsClients.dsStreamStore == nullptr) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "the datasystem stream store is empty, outcluster operations are not supported by now");
    }
    SetTraceId();
    return dsClients.dsStreamStore->QueryGlobalConsumersNum(streamName, gConsumerNum);
}

std::string Libruntime::GetInvokingRequestId(void)
{
    return runtimeContext->GetInvokingRequestId();
}

ErrorInfo Libruntime::Cancel(const std::vector<std::string> &objids, bool isForce, bool isRecursive)
{
    return invokeAdaptor->Cancel(objids, isForce, isRecursive);
}

void Libruntime::Exit(void)
{
    // exit data system
    invokeAdaptor->Exit(0, "");
}

void Libruntime::Exit(const int code, const std::string &message)
{
    // exit data system
    invokeAdaptor->Exit(code, message);
}

ErrorInfo Libruntime::Kill(const std::string &instanceId, int sigNo)
{
    auto realInsId = memStore->GetInstanceId(instanceId);
    invokeAdaptor->EraseFsIntf(realInsId);
    return invokeAdaptor->Kill(realInsId, "", sigNo);
}

ErrorInfo Libruntime::Kill(const std::string &instanceId, int sigNo, std::shared_ptr<Buffer> data)
{
    auto realInsId = memStore->GetInstanceId(instanceId);
    return invokeAdaptor->Kill(realInsId, std::string(static_cast<char *>(data->MutableData()), data->GetSize()),
                               sigNo);
}

void Libruntime::Finalize(bool isDriver)
{
    if (generatorNotifier_ != nullptr) {
        generatorNotifier_.reset();
    }
    if (generatorReceiver_ != nullptr) {
        generatorReceiver_->Stop();
        generatorReceiver_.reset();
    }
    if (driverLogReceiver_ != nullptr) {
        driverLogReceiver_.reset();
    }
    if (memStore) {
        memStore->Clear();
    }
    if (dsClients.dsObjectStore != nullptr) {
        dsClients.dsObjectStore.reset();
    }
    if (dsClients.dsStateStore != nullptr) {
        dsClients.dsStateStore.reset();
    }
    if (dsClients.dsStreamStore != nullptr) {
        dsClients.dsStreamStore.reset();
    }
    if (invokeAdaptor != nullptr) {
        invokeAdaptor->Finalize(isDriver);
    }
    if (!config->inCluster) {
        auto err = clientsMgr->ReleaseHttpClient(config->functionSystemIpAddr, config->functionSystemPort);
        if (!err.OK()) {
            YRLOG_ERROR("failed to release http client, message({})", err.Msg());
        }
    } else {
        auto err = clientsMgr->ReleaseDsClient(config->dataSystemIpAddr, config->dataSystemPort);
        if (!err.OK()) {
            YRLOG_ERROR("failed to release data system client, message({})", err.Msg());
        }
    }
    // If there are service requirements, the plaintext authentication credential can be stored in the memory. However,
    // the plaintext authentication credential needs to be cleared when an abnormal branch or exit is complete.
    config->ClearPaaswd();
    security_->ClearPrivateKey();
}

void Libruntime::WaitAsync(const std::string &objectId, WaitAsyncCallback callback, void *userData)
{
    this->memStore->AddReadyCallback(
        objectId, [objectId, callback, userData](const ErrorInfo &err) { callback(objectId, err, userData); });
}

void Libruntime::GetAsync(const std::string &objectId, GetAsyncCallback callback, void *userData)
{
    this->memStore->AddReadyCallbackWithData(
        objectId, [objectId, callback, userData](const ErrorInfo &err, std::shared_ptr<Buffer> buf) {
            std::shared_ptr<DataObject> dataObj;
            if (buf) {
                dataObj = std::make_shared<DataObject>(objectId, buf);
            } else {
                // make a fake buffer ptr to avoid accessing invalid memory
                dataObj = std::make_shared<DataObject>(0, 0);
                dataObj->id = objectId;
            }
            callback(dataObj, err, userData);
        });
}

void Libruntime::GetEvent(const std::string &objectId, GetEventCallback callback, void *userData)
{
    if (callback == nullptr) {
        YRLOG_ERROR("GetEventCallback is nullptr");
        return;
    }
    std::string requestId = YR::utility::IDGenerator::GetRequestIdFromObj(objectId);
    YRLOG_DEBUG("begin to GetEvent, reqId is {}", requestId);
    this->memStore->AddEventCallbackWithData(
        requestId, [objectId, callback, userData](const ErrorInfo &err, std::shared_ptr<Buffer> buffer) {
            YRLOG_DEBUG("begin to excute EventCallbackWithData, objectId is {}", objectId);
            std::shared_ptr<DataObject> dataObj;
            if (buffer) {
                YRLOG_DEBUG("buffer is not null, buffer size is {}", buffer->GetSize());
                dataObj = std::make_shared<DataObject>(objectId, buffer);
            } else {
                dataObj = std::make_shared<DataObject>(0, 0);
                dataObj->id = objectId;
            }
            YRLOG_DEBUG("begin to excute go callback, objectId is {}", objectId);

            callback(dataObj, err, userData);
        });
}

bool Libruntime::IsObjectExistingInLocal(const std::string &objId)
{
    return this->memStore->IsExistedInLocal(objId);
}

ErrorInfo Libruntime::GroupCreate(const std::string &groupName, GroupOpts &opts)
{
    YRLOG_DEBUG("group name is {}, timeout is {}", groupName, opts.timeout);
    return this->invokeAdaptor->GroupCreate(groupName, opts);
}

ErrorInfo Libruntime::GroupWait(const std::string &groupName)
{
    return this->invokeAdaptor->GroupWait(groupName);
}

void Libruntime::GroupTerminate(const std::string &groupName)
{
    return this->invokeAdaptor->GroupTerminate(groupName);
}

ErrorInfo Libruntime::GroupSuspend(const std::string &groupName)
{
    return this->invokeAdaptor->GroupSuspend(groupName);
}

ErrorInfo Libruntime::GroupResume(const std::string &groupName)
{
    return this->invokeAdaptor->GroupResume(groupName);
}

std::pair<std::vector<std::string>, ErrorInfo> Libruntime::GetInstances(const std::string &objId, int timeoutSec)
{
    return this->memStore->GetInstanceIds(objId, timeoutSec);
}

std::pair<std::vector<std::string>, ErrorInfo> Libruntime::GetInstances(const std::string &objId,
                                                                        const std::string &groupName)
{
    return this->invokeAdaptor->GetInstanceIds(objId, groupName);
}

std::string Libruntime::GenerateGroupName()
{
    return YR::utility::IDGenerator::GenGroupId(runtimeContext->GetJobId());
}

ErrorInfo Libruntime::SaveState(const std::shared_ptr<Buffer> data, const int &timeout)
{
    return this->invokeAdaptor->SaveState(data, timeout);
}

ErrorInfo Libruntime::LoadState(std::shared_ptr<Buffer> &data, const int &timeout)
{
    return this->invokeAdaptor->LoadState(data, timeout);
}

ErrorInfo Libruntime::CreateStateStore(const DsConnectOptions &opts, std::shared_ptr<StateStore> &stateStore)
{
    auto client = std::make_shared<DSCacheStateStore>();
    auto err = client->Init(opts);
    if (err.OK()) {
        stateStore = std::move(client);
    }
    return err;
}

ErrorInfo Libruntime::SetTraceId(const std::string &traceId)
{
    datasystem::Status rc;
    if (!traceId.empty()) {
        rc = datasystem::Context::SetTraceId(traceId);
    } else if (!threadLocalTraceId.empty()) {
        rc = datasystem::Context::SetTraceId(threadLocalTraceId);
    }
    if (rc.IsError()) {
        return ErrorInfo(ConvertDatasystemErrorToCore(rc.GetCode()), YR::Libruntime::ModuleCode::DATASYSTEM,
                         rc.ToString());
    }
    return ErrorInfo();
}

ErrorInfo Libruntime::SetTenantId(const std::string &tenantId, bool isReturnErrWhenTenantIDEmpty)
{
    if (!config->enableAuth && config->inCluster) {
        return ErrorInfo();
    }
    if (isReturnErrWhenTenantIDEmpty && tenantId.empty()) {
        auto msg = "tenant id is empty, please set the correct tenant id or function urn in config.";
        YRLOG_ERROR("failed to set tenantId, err: {}", msg);
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME, msg);
    }
    if (!dsClients.dsObjectStore) {
        return ErrorInfo(ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "failed to set tenantId, err: datasystem client is empty, please check whether runtime is "
                         "initialized or exiting gracefully.");
    }
    this->dsClients.dsObjectStore->SetTenantId(tenantId);
    this->config->tenantId = tenantId;
    YRLOG_DEBUG("succeed to set tenant id, tenant id is {}", tenantId);
    return ErrorInfo();
}

void Libruntime::SetTenantIdWithPriority()
{
    auto tenantId = GetTenantId();
    SetTenantId(tenantId);
}

std::string Libruntime::GetTenantId()
{
    // get tenantId with priority: config tenant id > seturn tenant id > init tenant id
    auto tenantId = !this->config->tenantId.empty()
                        ? this->config->tenantId
                        : this->config->functionIds[this->config->selfLanguage].substr(
                              0, this->config->functionIds[this->config->selfLanguage].find_first_of('/'));
    return tenantId;
}

ErrorInfo Libruntime::GenerateKeyByStateStore(std::shared_ptr<StateStore> stateStore, std::string &returnKey)
{
    SetTraceId();
    return stateStore->GenerateKey(returnKey);
}

ErrorInfo Libruntime::SetByStateStore(std::shared_ptr<StateStore> stateStore, const std::string &key,
                                      std::shared_ptr<ReadOnlyNativeBuffer> nativeBuffer, SetParam setParam)
{
    SetTraceId();
    return stateStore->Write(key, nativeBuffer, setParam);
}

ErrorInfo Libruntime::SetValueByStateStore(std::shared_ptr<StateStore> stateStore,
                                           std::shared_ptr<ReadOnlyNativeBuffer> nativeBuffer, SetParam setParam,
                                           std::string &returnKey)
{
    SetTraceId();
    return stateStore->Write(nativeBuffer, setParam, returnKey);
}

SingleReadResult Libruntime::GetByStateStore(std::shared_ptr<StateStore> stateStore, const std::string &key,
                                             int timeoutMs)
{
    SetTraceId();
    return stateStore->Read(key, timeoutMs);
}

MultipleReadResult Libruntime::GetArrayByStateStore(std::shared_ptr<StateStore> stateStore,
                                                    const std::vector<std::string> &keys, int timeoutMs,
                                                    bool allowPartial)
{
    SetTraceId();
    return stateStore->Read(keys, timeoutMs, allowPartial);
}

ErrorInfo Libruntime::QuerySizeByStateStore(std::shared_ptr<StateStore> stateStore,
                                            const std::vector<std::string> &keys, std::vector<uint64_t> &outSizes)
{
    SetTraceId();
    return stateStore->QuerySize(keys, outSizes);
}

ErrorInfo Libruntime::DelByStateStore(std::shared_ptr<StateStore> stateStore, const std::string &key)
{
    SetTraceId();
    return stateStore->Del(key);
}

MultipleDelResult Libruntime::DelArrayByStateStore(std::shared_ptr<StateStore> stateStore,
                                                   const std::vector<std::string> &keys)
{
    SetTraceId();
    return stateStore->Del(keys);
}

ErrorInfo Libruntime::ExecShutdownCallback(uint64_t gracePeriodSec)
{
    if (this->invokeAdaptor == nullptr) {
        YRLOG_ERROR("Failed to call ExecShutdownCallback, invokeAdaptor is nullptr.");
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "invokeAdaptor is nullptr.");
    }
    return this->invokeAdaptor->ExecShutdownCallback(gracePeriodSec);
}

ErrorInfo Libruntime::SetUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    return metricsAdaptor->SetUInt64Counter(data);
}

ErrorInfo Libruntime::ResetUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    return metricsAdaptor->ResetUInt64Counter(data);
}

ErrorInfo Libruntime::IncreaseUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    return metricsAdaptor->IncreaseUInt64Counter(data);
}

std::pair<ErrorInfo, uint64_t> Libruntime::GetValueUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    return metricsAdaptor->GetValueUInt64Counter(data);
}

ErrorInfo Libruntime::SetDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    return metricsAdaptor->SetDoubleCounter(data);
}

ErrorInfo Libruntime::ResetDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    return metricsAdaptor->ResetDoubleCounter(data);
}

ErrorInfo Libruntime::IncreaseDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    return metricsAdaptor->IncreaseDoubleCounter(data);
}

std::pair<ErrorInfo, double> Libruntime::GetValueDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    return metricsAdaptor->GetValueDoubleCounter(data);
}

ErrorInfo Libruntime::ReportGauge(const YR::Libruntime::GaugeData &gauge)
{
    auto err = metricsAdaptor->ReportGauge(gauge);
    return err;
}

ErrorInfo Libruntime::SetAlarm(const std::string &name, const std::string &description,
                               const YR::Libruntime::AlarmInfo &alarmInfo)
{
    auto err = metricsAdaptor->SetAlarm(name, description, alarmInfo);
    return err;
}

std::pair<InstanceAllocation, ErrorInfo> Libruntime::AcquireInstance(const std::string &stateId,
                                                                     const FunctionMeta &functionMeta,
                                                                     InvokeOptions &opts)
{
    return invokeAdaptor->AcquireInstance(stateId, functionMeta, opts);
}

ErrorInfo Libruntime::ReleaseInstance(const std::string &leaseId, const std::string &stateId, bool abnormal,
                                      InvokeOptions &opts)
{
    return invokeAdaptor->ReleaseInstance(leaseId, stateId, abnormal, opts);
}

ErrorInfo Libruntime::ProcessLog(FunctionLog &functionLog)
{
    functionLog.set_instanceid(Config::Instance().INSTANCE_ID());
    auto socketMsg = messageCoder_->GenerateSocketMsg(MAGIC_NUMBER, X_VERSION, MESSAGE_REQUEST_BYTE,
                                                      YR::utility::IDGenerator::GenPacketId(), functionLog);
    auto toSend = messageCoder_->Encode(socketMsg);
    return socketClient_->Send(toSend);
}

void Libruntime::WaitEvent(FiberEventNotify &event)
{
    boost::this_fiber::yield();
    event.Wait();
}

void Libruntime::NotifyEvent(FiberEventNotify &event)
{
    event.Notify();
}

std::pair<ErrorInfo, std::string> Libruntime::PeekObjectRefStream(const std::string &generatorId, bool blocking)
{
    return memStore->GetOutput(generatorId, blocking);
}

ErrorInfo Libruntime::NotifyGeneratorResult(const std::string &generatorId, int index,
                                            std::shared_ptr<DataObject> resultObj, const ErrorInfo &resultErr)
{
    return generatorNotifier_->NotifyResult(generatorId, index, resultObj, resultErr);
}

ErrorInfo Libruntime::NotifyGeneratorFinished(const std::string &generatorId, int numResults)
{
    return generatorNotifier_->NotifyFinished(generatorId, numResults);
}

FunctionGroupRunningInfo Libruntime::GetFunctionGroupRunningInfo()
{
    return config->groupRunningInfo;
}

std::pair<ErrorInfo, ResourceGroupUnit> Libruntime::GetResourceGroupTable(const std::string &resourceGroupId)
{
    return invokeAdaptor->GetResourceGroupTable(resourceGroupId);
}

std::pair<ErrorInfo, std::vector<ResourceUnit>> Libruntime::GetResources(void)
{
    return invokeAdaptor->GetResources();
}

std::pair<ErrorInfo, std::string> Libruntime::GetNodeIpAddress(void)
{
    if (!this->config->isDriver) {
        return std::make_pair(ErrorInfo(), Config::Instance().HOST_IP());
    }
    return this->invokeAdaptor->GetNodeIpAddress();
}

std::pair<ErrorInfo, QueryNamedInsResponse> Libruntime::QueryNamedInstances()
{
    auto ret = invokeAdaptor->QueryNamedInstances();
    return ret;
}

Credential Libruntime::GetCredential()
{
    if (!this->security_) {
        return Credential{};
    }
    return this->security_->GetCredential();
}

ErrorInfo Libruntime::CheckRGroupName(const std::string &rGroupName)
{
    if (rGroupName == std::string(UNSUPPORTED_RGROUP_NAME) || rGroupName == "") {
        auto errMsg =
            "invalid resource group name, name: " + rGroupName + ", please set the name other than primary or empty.";
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME, errMsg);
    }
    return ErrorInfo();
}

ErrorInfo Libruntime::CheckRGroupSpec(const ResourceGroupSpec &resourceGroupSpec)
{
    auto err = CheckRGroupName(resourceGroupSpec.name);
    if (!err.OK()) {
        return err;
    }
    auto bundleSize = resourceGroupSpec.bundles.size();
    for (size_t i = 0; i < bundleSize; ++i) {
        for (auto &pair : resourceGroupSpec.bundles[i]) {
            if (pair.first.empty()) {
                auto errMsg =
                    "invalid bundle, bundle index: " + std::to_string(i) + ", please set a non-empty and correct key.";
                return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                                 errMsg);
            }
            if (pair.second < 0) {
                auto errMsg = "invalid bundle, bundle index: " + std::to_string(i) + ", please set the value of " +
                              pair.first + " >= 0.";
                return ErrorInfo(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, YR::Libruntime::ModuleCode::RUNTIME,
                                 errMsg);
            }
        }
    }
    return ErrorInfo();
}

ErrorInfo Libruntime::CreateResourceGroup(const ResourceGroupSpec &resourceGroupSpec, std::string &requestId)
{
    requestId = YR::utility::IDGenerator::GenRequestId();
    auto err = CheckRGroupSpec(resourceGroupSpec);
    if (!err.OK()) {
        YRLOG_ERROR(
            "check resource group create options failed, name: {}, bundles size: {}, request id: {}, error code: {}, "
            "error message: {}.",
            resourceGroupSpec.name, resourceGroupSpec.bundles.size(), requestId, fmt::underlying(err.Code()),
            err.Msg());
        return err;
    }

    std::string traceId = ConstructTraceId({});
    YRLOG_DEBUG("start to create resource group, name: {}, bundles size: {}, request id: {}, trace id: {}.",
                resourceGroupSpec.name, resourceGroupSpec.bundles.size(), requestId, traceId);
    auto spec = std::make_shared<ResourceGroupCreateSpec>(resourceGroupSpec, requestId, traceId,
                                                          runtimeContext->GetJobId(), GetTenantId());
    spec->BuildCreateResourceGroupRequest();
    this->invokeAdaptor->CreateResourceGroup(spec);
    return ErrorInfo();
}

ErrorInfo Libruntime::RemoveResourceGroup(const std::string &resourceGroupName)
{
    auto err = CheckRGroupName(resourceGroupName);
    if (!err.OK()) {
        return err;
    }
    YRLOG_DEBUG("start to remove resource group, name: {}.", resourceGroupName);
    this->rGroupManager_->RemoveRGDetail(resourceGroupName);
    this->invokeAdaptor->KillAsync(resourceGroupName, "", libruntime::Signal::RemoveResourceGroup);
    return ErrorInfo();
}

ErrorInfo Libruntime::WaitResourceGroup(const std::string &resourceGroupName, const std::string &requestId,
                                        int timeoutSec)
{
    YRLOG_DEBUG("start to wait resource group create info, name: {}, request id: {}, timeout: {}.", resourceGroupName,
                requestId, timeoutSec);
    return this->rGroupManager_->GetRGCreateErrInfo(resourceGroupName, requestId, timeoutSec);
}

std::pair<YR::Libruntime::FunctionMeta, ErrorInfo> Libruntime::GetInstance(const std::string &name,
                                                                           const std::string &nameSpace, int timeoutSec)
{
    auto [meta, err] = this->invokeAdaptor->GetInstance(name, nameSpace, timeoutSec);
    if (!err.OK() &&
        (err.Code() == ErrorCode::ERR_INSTANCE_NOT_FOUND || err.Code() == ErrorCode::ERR_INSTANCE_EXITED)) {
        auto insId = nameSpace.empty() ? this->config->ns + "-" + name : nameSpace + "-" + name;
        YRLOG_WARN("instance： {} not exist, need kill directly", insId);
        this->invokeAdaptor->Kill(nameSpace.empty() ? this->config->ns + "-" + name : nameSpace + "-" + name, "",
                                       libruntime::Signal::killInstanceSync);
    }
    if (err.OK() && meta.needOrder) {
        this->invokeOrderMgr->RegisterInstance(nameSpace.empty() ? this->config->ns + "-" + name
                                                                 : nameSpace + "-" + name);
    }
    return std::make_pair<>(meta, err);
}

bool Libruntime::IsLocalInstances(const std::vector<std::string> &instanceIds)
{
    auto dsAddress = config->dataSystemIpAddr + ":" + std::to_string(config->dataSystemPort);
    std::vector<std::future<bool>> futures;
    for (const auto &instanceId : instanceIds) {
        auto promise = std::make_shared<std::promise<bool>>();
        futures.emplace_back(promise->get_future());
        memStore->AddReadyCallback(instanceId, [this, promise, instanceId, dsAddress](const ErrorInfo &err) {
            if (!err.OK()) {
                promise->set_value(false);
                return;
            }
            auto killErr = this->invokeAdaptor->Kill(instanceId, "", libruntime::Signal::QueryDsAddress);
            if (!killErr.OK()) {
                YRLOG_WARN("kill QueryDsAddress code: {}, msg: {}", fmt::underlying(err.Code()), err.Msg());
                promise->set_value(false);
                return;
            }
            if (killErr.Msg() != dsAddress) {
                YRLOG_DEBUG("not local instances, local ds address is {}, msg {}", dsAddress, killErr.Msg());
                promise->set_value(false);
                return;
            }
            promise->set_value(true);
        });
    }
    for (auto &future : futures) {
        if (!future.get()) {
            return false;
        }
    }
    YRLOG_DEBUG("all are local instances.");
    return true;
}

ErrorInfo Libruntime::Accelerate(const std::string &groupName, const AccelerateMsgQueueHandle &handle,
                                 HandleReturnObjectCallback callback)
{
    return invokeAdaptor->Accelerate(groupName, handle, callback);
}

bool Libruntime::AddReturnObject(const std::vector<std::string> &objIds)
{
    for (const auto &objId : objIds) {
        if (!this->memStore->AddReturnObject(objId)) {
            return false;
        }
    }
    return true;
}

bool Libruntime::SetError(const std::string &objId, const ErrorInfo &err)
{
    return this->memStore->SetError(objId, err);
}

void Libruntime::UpdateSchdulerInfo(const std::string &schedulerName, const std::string &schedulerId,
                                    const std::string &option)
{
    YRLOG_DEBUG("start update scheduler info, scheduler name is {}, scheduler id is {}, options is {}", schedulerName,
                schedulerId, option);
    return invokeAdaptor->UpdateSchdulerInfo(schedulerName, schedulerId, option);
}

std::string Libruntime::GetInstanceRoute(const std::string &objectId)
{
    return memStore->GetInstanceRoute(objectId);
}

void Libruntime::SaveInstanceRoute(const std::string &objectId, const std::string &instanceRoute)
{
    memStore->SetInstanceRoute(objectId, instanceRoute);
}

bool Libruntime::IsHealth()
{
    if (!invokeAdaptor) {
        return false;
    }
    return invokeAdaptor->IsHealth();
}

bool Libruntime::IsDsHealth()
{
    if (!dsClients.dsStreamStore && !dsClients.dsStateStore) {
        return true;
    }
    auto err = dsClients.dsStateStore->HealthCheck();
    if (err.OK()) {
        return true;
    }
    return false;
}

void Libruntime::KillAsync(const std::string &instanceId, int sigNo, std::function<void(const ErrorInfo &err)> cb)
{
    auto realInsId = memStore->GetInstanceId(instanceId);
    invokeAdaptor->EraseFsIntf(realInsId);
    this->invokeAdaptor->KillAsyncCB(realInsId, "", sigNo, cb);
}

ErrorInfo Libruntime::StreamWrite(const std::string &streamMessage, const std::string &requestId,
                                  const std::string &instanceId)
{
    YRLOG_DEBUG("start to write stream event, size of streamMessage is {}, requestId is {}, instanceId is {}",
                streamMessage.size(), requestId, instanceId);
    return invokeAdaptor->StreamWriteEvent(streamMessage, requestId, instanceId);
}

std::pair<std::string, ErrorInfo> Libruntime::LoadCurrentSession(const std::string &sessionId)
{
    if (invokeAdaptor == nullptr) {
        return {"", ErrorInfo(ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "invoke adaptor is nullptr")};
    }
    return invokeAdaptor->LoadCurrentSession(sessionId);
}

ErrorInfo Libruntime::UpdateCurrentSession(const std::string &sessionId, const std::string &sessionData)
{
    if (invokeAdaptor == nullptr) {
        return ErrorInfo(ERR_INNER_SYSTEM_ERROR, ModuleCode::RUNTIME, "invoke adaptor is nullptr");
    }
    return invokeAdaptor->UpdateCurrentSession(sessionId, sessionData);
}

std::pair<std::string, std::string> Libruntime::GetRequestAndInstanceID()
{
    if (threadLocalRequestId.empty() || threadLocalInstanceId.empty()) {
        YRLOG_DEBUG("requestId or instanceId is empty, threadLocalRequestId is {}, threadLocalInstanceId is {}",
                    threadLocalRequestId, threadLocalInstanceId);
        return {"", ""};
    }
    return {threadLocalRequestId, threadLocalInstanceId};
}

std::pair<ErrorInfo, std::string> Libruntime::GetNodeId()
{
    ErrorInfo err;
    if (!this->config->isDriver) {
        return std::make_pair(ErrorInfo(), Config::Instance().NODE_ID());
    }
    return this->invokeAdaptor->GetNodeId();
}

std::string Libruntime::GetNameSpace()
{
    return this->config->ns;
}

}  // namespace Libruntime
}  // namespace YR
