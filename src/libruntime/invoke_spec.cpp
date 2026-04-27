/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#include "src/libruntime/invoke_spec.h"

#include <google/protobuf/util/json_util.h>

namespace YR {
namespace Libruntime {
const std::string LOW_RELIABILITY_TYPE = "low";
const std::string HIGH_RELIABILITY_TYPE = "high";
const char *STORAGE_TYPE = "storage_type";
const char *CODE_PATH = "code_path";
const char *WORKING_DIR = "working_dir";
const char *DELEGATE_DOWNLOAD = "DELEGATE_DOWNLOAD";
const char *ENABLE_DEBUG_KEY = "enable";
const char *ENABLE_DEBUG = "true";
const char *DEBUG_CONFIG_KEY = "debug_config";

void to_json(nlohmann::json &j, const CancelReqInfo &cancelReqInfo)
{
    j = json{{"requestId", cancelReqInfo.requestId}, {"instanceId", cancelReqInfo.instanceId}};
}

void from_json(const nlohmann::json &j, CancelReqInfo &cancelReqInfo)
{
    j.at("requestId").get_to(cancelReqInfo.requestId);
    j.at("instanceId").get_to(cancelReqInfo.instanceId);
}

InvokeSpec::InvokeSpec(const std::string &jobId, const FunctionMeta &functionMeta,
                       const std::vector<DataObject> &returnObjs, std::vector<InvokeArg> invokeArgs,
                       const libruntime::InvokeType invokeType, std::string traceId, std::string requestId,
                       const std::string &instanceId, InvokeOptions opts)
    : jobId(jobId),
      functionMeta(functionMeta),
      returnIds(returnObjs),
      invokeArgs(std::move(invokeArgs)),
      invokeType(invokeType),
      traceId(std::move(traceId)),
      requestId(std::move(requestId)),
      instanceId(instanceId),
      opts(std::move(opts)),
      requestInvoke(std::make_shared<InvokeMessageSpec>())
{
    schedulerInstanceIdMtx_ = std::make_shared<absl::Mutex>();
    createTimestamp =
        std::chrono::duration_cast<std::chrono::milliseconds>(std::chrono::system_clock::now().time_since_epoch())
            .count();
}

void InvokeSpec::ConsumeRetryTime()
{
    if (opts.retryTimes > 0) {
        --opts.retryTimes;
    }
}

bool InvokeSpec::ExceedMaxRetryTime()
{
    if (opts.maxRetryTime == -1) {
        return false;
    }
    if (opts.maxRetryTime == 0) {
        return true;
    }
    --opts.maxRetryTime;
    return false;
}

void InvokeSpec::IncrementSeq()
{
    seq++;
}

std::string InvokeSpec::ConstructRequestID()
{
    return YR::utility::IDGenerator::GenRequestId(this->requestId, this->seq);
}

std::string InvokeSpec::GetNamedInstanceId(std::shared_ptr<LibruntimeConfig> config)
{
    if (!functionMeta.name.empty()) {
        if (!functionMeta.ns.empty()) {
            return functionMeta.ns + "-" + functionMeta.name;
        }
        if (!config->ns.empty()) {
            return config->ns + "-" + functionMeta.name;
        }
        return DEFAULT_YR_NAMESPACE + "-" + functionMeta.name;
    }
    return "";
}

std::string InvokeSpec::GetInstanceId(std::shared_ptr<LibruntimeConfig> config)
{
    InitDesignatedInstanceId(*config);
    if (!designatedInstanceID.empty()) {
        return designatedInstanceID;
    }

    if (!returnIds.empty()) {
        return returnIds[0].id;
    }

    return "";
}

void InvokeSpec::InitDesignatedInstanceId(const LibruntimeConfig &config)
{
    if (!functionMeta.name.empty()) {
        auto ns = functionMeta.ns.empty() ? config.ns : functionMeta.ns;
        if (ns.empty()) {
            designatedInstanceID = functionMeta.name;
        } else {
            designatedInstanceID = ns + "-" + functionMeta.name;
        }
    }
}

std::vector<std::string> InvokeSpec::GetSchedulerInstanceIds()
{
    absl::ReaderMutexLock lock(schedulerInstanceIdMtx_.get());
    return this->opts.schedulerInstanceIds;
}

std::string InvokeSpec::GetSchedulerInstanceId()
{
    absl::ReaderMutexLock lock(schedulerInstanceIdMtx_.get());
    if (this->opts.schedulerInstanceIds.empty()) {
        return "";
    }
    return this->opts.schedulerInstanceIds[0];
}

void InvokeSpec::SetSchedulerInstanceId(const std::string &schedulerId)
{
    absl::WriterMutexLock lock(schedulerInstanceIdMtx_.get());
    if (this->opts.schedulerInstanceIds.empty()) {
        this->opts.schedulerInstanceIds.push_back(schedulerId);
    } else {
        this->opts.schedulerInstanceIds[0] = schedulerId;
    }
}

bool InvokeSpec::IsStaleDuplicateNotify(uint8_t sequence)
{
    if (sequence != this->seq) {  // stale duplicate notify
        YRLOG_INFO(
            "Received stale duplicate notify, invoke type: {}, raw requestId: {}, notify seq: {}, current seq: {}",
            static_cast<int>(this->invokeType), this->requestId, sequence, this->seq);
        return true;
    }
    return false;
}

void InvokeSpec::BuildRequestPbOptions(InvokeOptions &opts, const LibruntimeConfig &config, CreateRequest &request)
{
    BuildRequestPbScheduleOptions(opts, config, request);
    BuildRequestPbCreateOptions(opts, config, request);
}

void InvokeSpec::BuildRequestPbScheduleOptions(InvokeOptions &opts, const LibruntimeConfig &config,
                                               CreateRequest &request)
{
    SchedulingOptions *schedulingOps = request.mutable_schedulingops();
    schedulingOps->set_priority(opts.instancePriority);
    schedulingOps->set_scheduletimeoutms(opts.scheduleTimeoutMs);
    schedulingOps->set_preemptedallowed(opts.preemptedAllowed);
    auto *resourceMap = schedulingOps->mutable_resources();
    if (opts.cpu >= 0) {
        resourceMap->insert({"CPU", static_cast<double>(opts.cpu)});
    }
    if (opts.memory >= 0) {
        resourceMap->insert({"Memory", static_cast<double>(opts.memory)});
    }
    for (auto &resource : opts.customResources) {
        YRLOG_DEBUG("start insert custom resource into schedule opts, key is {}, value is {}", resource.first,
                    resource.second);
        resourceMap->insert({resource.first, static_cast<double>(resource.second)});
    }
    auto *extensionMap = schedulingOps->mutable_extension();
    for (auto &extention : opts.customExtensions) {
        extensionMap->insert({extention.first, extention.second});
    }
    if (ResourceGroupEnabled(opts.resourceGroupOpts)) {
        schedulingOps->set_rgroupname(opts.resourceGroupOpts.resourceGroupName);
    }
    static std::unordered_map<std::string, core_service::AffinityType> affinityMap = {
        {"PreferredAffinity", core_service::AffinityType::PreferredAffinity},
        {"PreferredAntiAffinity", core_service::AffinityType::PreferredAntiAffinity},
        {"RequiredAffinity", core_service::AffinityType::RequiredAffinity},
        {"RequiredAntiAffinity", core_service::AffinityType::RequiredAntiAffinity}};
    auto *affinities = schedulingOps->mutable_affinity();
    for (auto &affinity : opts.affinity) {
        auto iter = affinityMap.find(affinity.second);
        if (iter == affinityMap.end()) {
            YRLOG_ERROR("Invalid opts affinity, affinity: {}", affinity.second);
        }
        auto result = affinities->insert({affinity.first, affinityMap[affinity.second]});
        if (result.second == false) {
            std::cerr << affinity.first << " was already presented in affinities." << std::endl;
        }
    }

    auto scheduleAffinity = schedulingOps->mutable_scheduleaffinity();
    for (auto &affinity : opts.scheduleAffinities) {
        affinity->UpdatePbAffinity(scheduleAffinity);
    }

    if (InstanceRangeEnabled(opts.instanceRange)) {
        auto *instanceRange = schedulingOps->mutable_range();
        instanceRange->set_min(opts.instanceRange.min);
        instanceRange->set_max(opts.instanceRange.max);
        instanceRange->set_step(opts.instanceRange.step);
    }
}

void InvokeSpec::BuildRequestPbCreateOptions(InvokeOptions &opts, const LibruntimeConfig &config,
                                             CreateRequest &request)
{
    auto *createOptions = request.mutable_createoptions();
    createOptions->insert({"DATA_AFFINITY_ENABLED", opts.isDataAffinity ? "true" : "false"});
    for (auto &opt : opts.createOptions) {
        createOptions->insert({opt.first, opt.second});
    }
    for (auto &extension : opts.customExtensions) {
        if (extension.first == CONCURRENCY) {
            createOptions->insert({CONCURRENT_NUM, extension.second});
            continue;
        }
        createOptions->insert({extension.first, extension.second});
    }
    if (!opts.podLabels.empty()) {
        nlohmann::json podLabelsJson;
        for (auto &iter : opts.podLabels) {
            podLabelsJson[iter.first] = iter.second;
        }
        try {
            auto podLabelsValue = podLabelsJson.dump();
            createOptions->insert({DELEGATE_POD_LABELS, podLabelsValue});
        } catch (const std::exception &e) {
            YRLOG_WARN("json dump error: {}", e.what());
        }
    }
    if (opts.needOrder) {
        createOptions->insert({NEED_ORDER, ""});
    }
    createOptions->insert({RECOVER_RETRY_TIMES, std::to_string(opts.recoverRetryTimes)});
    auto workingDir = !opts.workingDir.empty() ? opts.workingDir : config.workingDir;
    if (!workingDir.empty()) {
        YRLOG_DEBUG("create meta workingDir: >{}<", workingDir);
        nlohmann::json delegateDownloadJson;
        delegateDownloadJson[std::string(STORAGE_TYPE)] = std::string(WORKING_DIR);
        delegateDownloadJson[std::string(CODE_PATH)] = workingDir;
        try {
            auto delegateDownloadValue = delegateDownloadJson.dump();
            createOptions->insert({std::string(DELEGATE_DOWNLOAD), delegateDownloadValue});
        } catch (const std::exception &e) {
            YRLOG_WARN("json dump error: {}", e.what());
        }
    }
    nlohmann::json envsJson;
    if (!config.customEnvs.empty()) {
        for (auto &pair : config.customEnvs) {
            envsJson[pair.first] = pair.second;
        }
    }
    if (!opts.envVars.empty()) {
        for (auto &pair : opts.envVars) {
            envsJson[pair.first] = pair.second;
        }
    }
    if (ResourceGroupEnabled(opts.resourceGroupOpts)) {
        envsJson["RG_NAME"] = opts.resourceGroupOpts.resourceGroupName;
    }
    try {
        createOptions->insert({DELEGATE_ENV_VAR, envsJson.dump()});
    } catch (const std::exception &e) {
        YRLOG_WARN("json dump error: {}", e.what());
    }
    if (config.isLowReliabilityTask) {
        std::string reliability = (invokeType == libruntime::InvokeType::CreateInstanceStateless)
                                      ? LOW_RELIABILITY_TYPE
                                      : HIGH_RELIABILITY_TYPE;
        createOptions->insert({RELIABILITY_TYPE, reliability});
    }
    if (opts.debug.enable) {
        nlohmann::json debugJson;
        debugJson[std::string(ENABLE_DEBUG_KEY)] = std::string(ENABLE_DEBUG);
        createOptions->insert({std::string(DEBUG_CONFIG_KEY), debugJson.dump()});
    }
}

void InvokeSpec::BuildInstanceCreateRequest(const LibruntimeConfig &config)
{
    BuildRequestPbOptions(this->opts, config, requestCreate);
    BuildRequestPbArgs(config, requestCreate, true);
    auto *createOptions = requestCreate.mutable_createoptions();
    createOptions->insert({TENANT_ID, config.tenantId});
    if (!functionMeta.functionId.empty()) {
        requestCreate.set_function(functionMeta.functionId);
    } else {
        if (config.functionIds.find(functionMeta.languageType) != config.functionIds.end()) {
            requestCreate.set_function(config.functionIds.at(functionMeta.languageType));
        }
    }
    requestCreate.set_requestid(this->ConstructRequestID());
    requestCreate.set_traceid(this->traceId);
    InitDesignatedInstanceId(config);
    if (!designatedInstanceID.empty()) {
        requestCreate.set_designatedinstanceid(designatedInstanceID);
    }
    for (auto &label : this->opts.labels) {
        requestCreate.add_labels(label);
    }
}

void InvokeSpec::BuildInstanceInvokeRequest(const LibruntimeConfig &config)
{
    BuildRequestPbArgs(config, requestInvoke->Mutable(), false);
    if (!functionMeta.functionId.empty()) {
        requestInvoke->Mutable().set_function(functionMeta.functionId);
    } else {
        if (config.functionIds.find(functionMeta.languageType) != config.functionIds.end()) {
            requestInvoke->Mutable().set_function(config.functionIds.at(functionMeta.languageType));
        }
    }
    requestInvoke->Mutable().set_requestid(this->ConstructRequestID());
    requestInvoke->Mutable().set_traceid(this->traceId);
    requestInvoke->Mutable().set_instanceid(this->invokeInstanceId);
    for (const auto &obj : returnIds) {
        requestInvoke->Mutable().add_returnobjectids(obj.id);
    }
    auto invokeOptions = requestInvoke->Mutable().mutable_invokeoptions();
    auto customTag = invokeOptions->mutable_customtag();
    for (auto &extension : opts.customExtensions) {
        if (extension.first == YR_ROUTE) {
            continue;
        }
        customTag->insert({extension.first, extension.second});
    }
    if (opts.instanceSession && !opts.instanceSession->sessionID.empty()) {
        customTag->insert({YR_AGENT_SESSION_ID, opts.instanceSession->sessionID});
    }
    if (!instanceRoute.empty()) {
        customTag->insert({YR_ROUTE, instanceRoute});
    }
}

std::string InvokeSpec::BuildCreateMetaData(const LibruntimeConfig &config, std::string &funcMetaStr)
{
    InitDesignatedInstanceId(config);
    libruntime::MetaData meta;
    meta.set_invoketype(this->invokeType);
    auto funcMeta = meta.mutable_functionmeta();
    funcMeta->set_applicationname(this->functionMeta.appName);
    funcMeta->set_apitype(this->functionMeta.apiType);
    funcMeta->set_classname(this->functionMeta.className);
    funcMeta->set_codeid(this->functionMeta.codeId);
    funcMeta->set_functionid(this->functionMeta.functionId);
    funcMeta->set_functionname(this->functionMeta.funcName);
    funcMeta->set_initializercodeid(this->functionMeta.initializerCodeId);
    funcMeta->set_isgenerator(this->functionMeta.isGenerator);
    funcMeta->set_isasync(this->functionMeta.isAsync);
    funcMeta->set_language(this->functionMeta.languageType);
    funcMeta->set_modulename(this->functionMeta.moduleName);
    funcMeta->set_signature(this->functionMeta.signature);
    funcMeta->set_code(this->functionMeta.code.data(), this->functionMeta.code.size());
    if (!this->functionMeta.name.empty()) {
        funcMeta->set_name(this->functionMeta.name);
    }
    if (!this->functionMeta.ns.empty()) {
        funcMeta->set_ns(this->functionMeta.ns);
    }
    if (!designatedInstanceID.empty()) {
        google::protobuf::util::JsonPrintOptions options;
        auto status = google::protobuf::util::MessageToJsonString(*funcMeta, &funcMetaStr, options);
        if (!status.ok()) {
            YRLOG_WARN("Failed to serialize function meta to json string, error message: {}", status.message());
        }
    }
    auto metaConfig = meta.mutable_config();
    config.BuildMetaConfig(*metaConfig);
    if (!this->opts.codePaths.empty()) {
        metaConfig->clear_codepaths();
        for (const auto &path : this->opts.codePaths) {
            metaConfig->add_codepaths(path);
        }
    }
    if (!this->functionMeta.functionId.empty()) {
        bool isSetValue = false;
        for (int i = 0; i < metaConfig->functionids_size(); i++) {
            if (metaConfig->functionids(i).language() == functionMeta.languageType) {
                auto funcId = metaConfig->mutable_functionids(i);
                funcId->set_functionid(functionMeta.functionId);
                isSetValue = true;
                break;
            }
        }
        if (!isSetValue) {
            auto funcId = metaConfig->add_functionids();
            funcId->set_language(functionMeta.languageType);
            funcId->set_functionid(functionMeta.functionId);
        }
    }
    if (!this->opts.schedulerInstanceIds.empty()) {
        for (const auto &id : this->opts.schedulerInstanceIds) {
            metaConfig->add_schedulerinstanceids(id);
        }
    }

    auto invocationMeta = meta.mutable_invocationmeta();
    if (config.runtimeId == "driver") {
        invocationMeta->set_invokerruntimeid(config.runtimeId + "_" + config.jobId);
    } else {
        invocationMeta->set_invokerruntimeid(config.runtimeId);
    }
    invocationMeta->set_invocationsequenceno(this->invokeSeqNo);
    invocationMeta->set_minunfinishedsequenceno(this->invokeUnfinishedSeqNo);
    YRLOG_DEBUG("create meta data: >{}<", meta.DebugString());
    return meta.SerializeAsString();
}

std::string InvokeSpec::BuildInvokeMetaData(const LibruntimeConfig &config)
{
    libruntime::MetaData meta;
    meta.set_invoketype(this->invokeType);
    auto funcMeta = meta.mutable_functionmeta();
    funcMeta->set_applicationname(this->functionMeta.appName);
    funcMeta->set_modulename(this->functionMeta.moduleName);
    funcMeta->set_functionname(this->functionMeta.funcName);
    funcMeta->set_classname(this->functionMeta.className);
    funcMeta->set_codeid(this->functionMeta.codeId);
    funcMeta->set_signature(this->functionMeta.signature);
    funcMeta->set_language(this->functionMeta.languageType);
    funcMeta->set_apitype(this->functionMeta.apiType);
    funcMeta->set_functionid(this->functionMeta.functionId);
    funcMeta->set_isgenerator(this->functionMeta.isGenerator);
    funcMeta->set_isasync(this->functionMeta.isAsync);
    funcMeta->set_tensortransporttarget(this->functionMeta.tensorTransportTarget);
    funcMeta->set_enabletensortransport(this->functionMeta.enableTensorTransport);
    funcMeta->set_code(this->functionMeta.code.data(), this->functionMeta.code.size());
    auto invocationMeta = meta.mutable_invocationmeta();
    if (config.runtimeId == "driver") {
        invocationMeta->set_invokerruntimeid(config.runtimeId + "_" + config.jobId);
    } else {
        invocationMeta->set_invokerruntimeid(config.runtimeId);
    }

    invocationMeta->set_invocationsequenceno(this->invokeSeqNo);
    invocationMeta->set_minunfinishedsequenceno(this->invokeUnfinishedSeqNo);
    YRLOG_DEBUG("invoke meta data: {}", meta.ShortDebugString());
    return meta.SerializeAsString();
}

bool RequestResource::operator==(const RequestResource &r) const
{
    if (r.opts.customResources.size() != opts.customResources.size()) {
        return false;
    }
    for (auto it = opts.customResources.begin(); it != opts.customResources.end(); ++it) {
        auto kvPair = r.opts.customResources.find(it->first);
        if (kvPair == r.opts.customResources.end() || fabs(kvPair->second - it->second) > FLOAT_EQUAL_RANGE) {
            return false;
        }
    }
    if (r.opts.scheduleAffinities.size() != opts.scheduleAffinities.size()) {
        return false;
    }
    std::size_t affinityHash = 0;
    for (auto it = opts.scheduleAffinities.begin(); it != opts.scheduleAffinities.end(); ++it) {
        std::size_t h1 = (*it)->GetAffinityHash();
        affinityHash = affinityHash ^ h1;
    }
    for (auto it = r.opts.scheduleAffinities.begin(); it != r.opts.scheduleAffinities.end(); ++it) {
        std::size_t h2 = (*it)->GetAffinityHash();
        affinityHash = affinityHash ^ h2;
    }
    if (opts.instanceSession || r.opts.instanceSession) {
        if (!opts.instanceSession || !r.opts.instanceSession) {
            return false;
        }
        if (opts.instanceSession->sessionID != r.opts.instanceSession->sessionID) {
            return false;
        }
        if (opts.instanceSession->sessionTTL != r.opts.instanceSession->sessionTTL) {
            return false;
        }
    }
    if (r.opts.invokeLabels.size() != opts.invokeLabels.size()) {
        return false;
    }
    for (auto it = opts.invokeLabels.begin(); it != opts.invokeLabels.end(); ++it) {
        auto kvPair = r.opts.invokeLabels.find(it->first);
        if (kvPair == r.opts.invokeLabels.end() || kvPair->second != it->second) {
            return false;
        }
    }
    if (r.opts.device.name != opts.device.name) {
        return false;
    }
    if (r.opts.debug.enable != opts.debug.enable) {
        return false;
    }
    if (r.opts.envVars.size() != opts.envVars.size()) {
        return false;
    }
    for (const auto &envPair : r.opts.envVars) {
        const auto &key = envPair.first;
        const auto &value = envPair.second;
        auto it = opts.envVars.find(key);
        if (it == opts.envVars.end() || it->second != value) {
            return false;
        }
    }
    return (functionMeta.languageType == r.functionMeta.languageType) &&
           (functionMeta.functionId == r.functionMeta.functionId) && (opts.cpu == r.opts.cpu) &&
           (opts.memory == r.opts.memory) && (concurrency == r.concurrency) &&
           (opts.resourceGroupOpts.resourceGroupName == r.opts.resourceGroupOpts.resourceGroupName) &&
           (opts.resourceGroupOpts.bundleIndex == r.opts.resourceGroupOpts.bundleIndex) && (affinityHash == 0);
}

std::size_t FaasInfoForBatchRenewFn::operator()(const FaasInfoForBatchRenew &i) const
{
    return std::hash<std::string>{}(i.schedulerFunctionID) ^ std::hash<std::string>{}(i.schedulerInstanceID) ^
           std::hash<std::string>{}(i.functionId) ^ std::hash<int64_t>{}(i.batchIndex);
}

bool FaasInfoForBatchRenew::operator==(const FaasInfoForBatchRenew &i) const
{
    return schedulerInstanceID == i.schedulerInstanceID && schedulerFunctionID == i.schedulerFunctionID &&
           functionId == i.functionId && batchIndex == i.batchIndex;
}

void RequestResource::Print(void) const
{
    YRLOG_DEBUG("function meta: {} {}", fmt::underlying(functionMeta.languageType), functionMeta.functionId);
}

std::size_t HashFn::operator()(const RequestResource &r) const
{
    std::size_t h1 = std::hash<libruntime::LanguageType>()(r.functionMeta.languageType);
    h1 ^= std::hash<std::string>()(r.functionMeta.functionId);
    std::size_t h2 = std::hash<int>()(r.opts.cpu);
    std::size_t h3 = std::hash<int>()(r.opts.memory);
    std::size_t h4 = std::hash<size_t>()(r.concurrency);
    std::size_t result = h1 ^ h2 ^ h3 ^ h4;
    for (auto &it : r.opts.customResources) {
        std::size_t h5 = std::hash<std::string>()(it.first);
        std::size_t h6 = std::hash<float>()(it.second);
        result = result ^ h5 ^ h6;
    }
    auto affinities = r.opts.scheduleAffinities;
    for (auto it = affinities.begin(); it != affinities.end(); it++) {
        std::size_t h7 = (*it)->GetAffinityHash();
        result = result ^ h7;
    }
    for (auto &it : r.opts.invokeLabels) {
        std::size_t h8 = std::hash<std::string>()(it.first);
        std::size_t h9 = std::hash<std::string>()(it.second);
        result = result ^ h8 ^ h9;
    }
    if (r.opts.instanceSession) {
        std::size_t h10 = std::hash<std::string>()(r.opts.instanceSession->sessionID);
        std::size_t h11 = std::hash<int>()(r.opts.instanceSession->sessionTTL);
        result = result ^ h10 ^ h11;
    }
    for (const auto &envPair : r.opts.envVars) {
        std::size_t h12 = std::hash<std::string>()(envPair.first);
        std::size_t h13 = std::hash<std::string>()(envPair.second);
        result = result ^ h12 ^ h13;
    }
    result = result ^ std::hash<DebugConfig>()(r.opts.debug);
    result = result ^ std::hash<ResourceGroupOptions>()(r.opts.resourceGroupOpts);
    return result;
}
}  // namespace Libruntime
}  // namespace YR
