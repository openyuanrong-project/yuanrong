/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

#include "clibruntime.h"

#include <new>
#include "securec.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/libruntime/utils/utils.h"
using namespace std;
using libruntime::SubscriptionType;
using YR::Libruntime::Affinity;
using namespace std::placeholders;
using YR::Libruntime::Buffer;
using YR::Libruntime::CacheType;
using YR::Libruntime::ConsistencyType;
using YR::Libruntime::CreateParam;
using YR::Libruntime::DataObject;
using YR::Libruntime::Element;
using YR::Libruntime::ErrorCode;
using YR::Libruntime::ErrorInfo;
using YR::Libruntime::ExistenceOpt;
using YR::Libruntime::FunctionMeta;
using YR::Libruntime::InstanceAllocation;
using YR::Libruntime::InvokeArg;
using YR::Libruntime::InvokeOptions;
using YR::Libruntime::LabelDoesNotExistOperator;
using YR::Libruntime::LabelExistsOperator;
using YR::Libruntime::LabelInOperator;
using YR::Libruntime::LabelNotInOperator;
using YR::Libruntime::LabelOperator;
using YR::Libruntime::LibruntimeConfig;
using YR::Libruntime::LibruntimeManager;
using YR::Libruntime::MSetParam;
using YR::Libruntime::NativeBuffer;
using YR::Libruntime::ProducerConf;
using YR::Libruntime::SetParam;
using YR::Libruntime::StackTraceElement;
using YR::Libruntime::StackTraceInfo;
using YR::Libruntime::StreamConsumer;
using YR::Libruntime::StreamProducer;
using YR::Libruntime::SubscriptionConfig;
using YR::Libruntime::WriteMode;

#ifdef __cplusplus
extern "C" {
#endif

#define RETURN_ERR_WHEN_CONSUMER_ISNULL(consumer)                                                           \
    if (consumer == nullptr || *consumer == nullptr) {                                                      \
        if (consumer != nullptr) {                                                                          \
            delete consumer;                                                                                \
        }                                                                                                   \
        return ErrorInfoToCError(ErrorInfo(ErrorCode::ERR_PARAM_INVALID, "failed to get valid consumer.")); \
    }

std::tuple<std::shared_ptr<YR::Libruntime::Libruntime>, ErrorInfo> getLibRuntime()
{
    auto lrt = LibruntimeManager::Instance().GetLibRuntime();
    if (lrt == nullptr) {
        YRLOG_ERROR("GetLibRuntime empty");
        return {lrt, ErrorInfo(YR::Libruntime::ErrorCode::ERR_FINALIZED, YR::Libruntime::ModuleCode::RUNTIME,
                               "libRuntime empty")};
    }

    return {lrt, ErrorInfo(YR::Libruntime::ErrorCode::ERR_OK, YR::Libruntime::ModuleCode::RUNTIME, "")};
}

ErrorInfo CErrorToErrorInfo(CErrorInfo *cerr)
{
    ErrorInfo err;
    if (cerr == nullptr) {
        return err;
    }
    err.SetErrorCode(static_cast<ErrorCode>(cerr->code));
    if (cerr->message != nullptr) {
        err.SetErrorMsg(cerr->message);
        free(cerr->message);
    }
    if (cerr->size_stackTracesInfo == 0) {
        free(cerr);
        return err;
    }
    std::vector<StackTraceInfo> stackTraces;
    for (int i = 0; i < cerr->size_stackTracesInfo; i++) {
        std::vector<StackTraceElement> elements;
        for (int j = 0; j < cerr->stackTracesInfo[i].size_stackTraces; j++) {
            StackTraceElement element;
            element.className = cerr->stackTracesInfo[i].stackTraces[j].className;
            element.methodName = cerr->stackTracesInfo[i].stackTraces[j].methodName;
            element.fileName = cerr->stackTracesInfo[i].stackTraces[j].fileName;
            element.lineNumber = cerr->stackTracesInfo[i].stackTraces[j].lineNumber;
            for (int k = 0; k < cerr->stackTracesInfo[i].stackTraces[j].size_extensions; k++) {
                element.extensions.emplace(cerr->stackTracesInfo[i].stackTraces[j].extensions[k].key,
                                           cerr->stackTracesInfo[i].stackTraces[j].extensions[k].value);
                free(cerr->stackTracesInfo[i].stackTraces[j].extensions[k].key);
                free(cerr->stackTracesInfo[i].stackTraces[j].extensions[k].value);
            }
            elements.emplace_back(std::move(element));
            free(cerr->stackTracesInfo[i].stackTraces[j].className);
            free(cerr->stackTracesInfo[i].stackTraces[j].methodName);
            free(cerr->stackTracesInfo[i].stackTraces[j].fileName);
        }
        StackTraceInfo stackTrace;
        stackTrace.SetStackTraceElements(elements);
        stackTrace.SetMsg(cerr->stackTracesInfo[i].message);
        free(cerr->stackTracesInfo[i].message);
        free(cerr->stackTracesInfo[i].stackTraces);
        stackTraces.emplace_back(std::move(stackTrace));
    }
    err.SetStackTraceInfos(stackTraces);
    free(cerr->stackTracesInfo->stackTraces);
    free(cerr->stackTracesInfo);
    free(cerr);
    return err;
}

ErrorInfo CHealthCheckCodeToErrorInfo(CHealthCheckCode code)
{
    ErrorInfo err;
    if (code == CHealthCheckCode::HEALTHY) {
        err.SetErrorCode(ErrorCode::ERR_HEALTH_CHECK_HEALTHY);
    } else if (code == CHealthCheckCode::HEALTH_CHECK_FAILED) {
        err.SetErrorCode(ErrorCode::ERR_HEALTH_CHECK_FAILED);
    } else if (code == CHealthCheckCode::SUB_HEALTH) {
        err.SetErrorCode(ErrorCode::ERR_HEALTH_CHECK_SUBHEALTH);
    } else {
        err.SetErrorCode(ErrorCode::ERR_HEALTH_CHECK_HEALTHY);
    }
    return err;
}

char *CString(const std::string &str)
{
    char *cStr = (char *)malloc(str.size() + 1);
    (void)memcpy_s(cStr, str.size(), str.data(), str.size());
    cStr[str.size()] = 0;
    return cStr;
}

CErrorInfo ErrorInfoToCError(const ErrorInfo &err)
{
    CErrorInfo cErr{};
    cErr.message = nullptr;
    cErr.code = static_cast<int>(err.Code());
    cErr.dsStatusCode = err.GetDsStatusCode();
    if (!err.Msg().empty()) {
        cErr.message = CString(err.Msg());
    }
    std::vector<StackTraceInfo> stackTracesInfo = err.GetStackTraceInfos();
    if (stackTracesInfo.size() == 0) {
        return cErr;
    }
    cErr.size_stackTracesInfo = stackTracesInfo.size();
    cErr.stackTracesInfo = new (std::nothrow) CStackTracesInfo[stackTracesInfo.size()];

    for (size_t i = 0; i < stackTracesInfo.size(); i++) {
        std::vector<StackTraceElement> elements = stackTracesInfo[i].StackTraceElements();
        cErr.stackTracesInfo[i].stackTraces = new (std::nothrow) CStackTrace[elements.size()];
        cErr.stackTracesInfo[i].size_stackTraces = elements.size();
        cErr.stackTracesInfo[i].message = CString(stackTracesInfo[i].Message());
        for (size_t j = 0; j < elements.size(); j++) {
            cErr.stackTracesInfo[i].stackTraces[j].className = CString(elements[j].className);
            cErr.stackTracesInfo[i].stackTraces[j].fileName = CString(elements[j].fileName);
            cErr.stackTracesInfo[i].stackTraces[j].methodName = CString(elements[j].methodName);
            cErr.stackTracesInfo[i].stackTraces[j].lineNumber = static_cast<int>(elements[j].lineNumber);
            cErr.stackTracesInfo[i].stackTraces[j].size_extensions = elements[j].extensions.size();
            cErr.stackTracesInfo[i].stackTraces[j].extensions =
                new (std::nothrow) CCustomExtension[elements[j].extensions.size()];
            int k = 0;
            for (auto iter = elements[j].extensions.begin(); iter != elements[j].extensions.end(); iter++) {
                cErr.stackTracesInfo[i].stackTraces[j].extensions[k].key = CString(iter->first);
                cErr.stackTracesInfo[i].stackTraces[j].extensions[k].value = CString(iter->second);
                k++;
            }
        }
    }
    return cErr;
}

SetParam CSetParamToSetParam(CSetParam param)
{
    SetParam setParam;
    setParam.writeMode = static_cast<WriteMode>(param.writeMode);
    setParam.ttlSecond = param.ttlSecond;
    setParam.existence = static_cast<ExistenceOpt>(param.existence);
    setParam.cacheType = static_cast<CacheType>(param.cacheType);
    return setParam;
}

MSetParam CMSetParamToMSetParam(CMSetParam param)
{
    MSetParam mSetParam;
    mSetParam.writeMode = static_cast<WriteMode>(param.writeMode);
    mSetParam.ttlSecond = param.ttlSecond;
    mSetParam.existence = static_cast<ExistenceOpt>(param.existence);
    mSetParam.cacheType = static_cast<CacheType>(param.cacheType);
    return mSetParam;
}

CreateParam CCreateParamToCreateParam(CCreateParam param)
{
    CreateParam createParam;
    createParam.writeMode = static_cast<WriteMode>(param.writeMode);
    createParam.consistencyType = static_cast<ConsistencyType>(param.consistencyType);
    createParam.cacheType = static_cast<CacheType>(param.cacheType);
    return createParam;
}

void CheckNullAndAssignValue(const char *str, const int len, std::string &returnValue)
{
    if (str != nullptr && len > 0) {
        returnValue = std::string(str, len);
    }
}

char *StringToCString(const std::string &input)
{
    char *ret = nullptr;
    if (input.empty()) {
        return ret;
    }
    size_t destSize = input.size() + 1;
    ret = static_cast<char *>(malloc(destSize));
    if (ret != nullptr) {
        int err = memcpy_s(ret, destSize, input.data(), input.size());
        if (err == EOK) {
            ret[destSize - 1] = '\0';
        } else {
            free(ret);
            ret = nullptr;
            YRLOG_ERROR("StringToCString memcpy_s failed: {}", err);
            return nullptr;
        }
    }
    return ret;
}

CCredential ConverToCCredential(YR::Libruntime::Credential &credential)
{
    CCredential ccred{};
    ccred.ak = StringToCString(credential.ak);
    ccred.sk = StringToCString(credential.sk);
    ccred.sizeSk = credential.sk.length();
    ccred.dk = StringToCString(credential.dk);
    ccred.sizeDk = credential.dk.length();
    return ccred;
}

CCredential CredentialToCCre(YR::Libruntime::Credential &credential)
{
    return ConverToCCredential(credential);
}

void StringsToCStrings(const std::vector<std::string> &stringVec, char ***cStrings, int *cStringsLen)
{
    if (!stringVec.empty()) {
        *cStrings = (char **)malloc(stringVec.size() * sizeof(char **));
        for (size_t i = 0; i < stringVec.size(); i++) {
            (*cStrings)[i] = StringToCString(stringVec[i]);
        }
        *cStringsLen = stringVec.size();
    }
}

std::vector<std::string> CStringsToStrings(char **cKeys, int cKeysLen)
{
    std::vector<std::string> keys(cKeysLen);
    for (int i = 0; i < cKeysLen; i++) {
        keys[i] = cKeys[i];
    }
    return keys;
}

ErrorInfo LoadFunctionsWrapper(const std::vector<std::string> &codePaths)
{
    char **cPaths = new (std::nothrow) char *[codePaths.size()];
    for (size_t i = 0; i < codePaths.size(); i++) {
        cPaths[i] = const_cast<char *>(codePaths[i].c_str());
    }
    CErrorInfo *cerr = GoLoadFunctions(cPaths, codePaths.size());
    delete[] cPaths;
    return CErrorToErrorInfo(cerr);
}

void InsAllocationToCInsAllocation(const InstanceAllocation &alloc, CInstanceAllocation *cInsAlloc)
{
    cInsAlloc->functionId = CString(alloc.functionId);
    cInsAlloc->funcSig = CString(alloc.funcSig);
    cInsAlloc->instanceId = CString(alloc.instanceId);
    cInsAlloc->leaseId = CString(alloc.leaseId);
    cInsAlloc->tLeaseInterval = alloc.tLeaseInterval;
}

CFunctionMeta FunctionMetaToCFunctionMeta(const FunctionMeta &function)
{
    CFunctionMeta cFuncMeta{};
    cFuncMeta.appName = const_cast<char *>(function.appName.c_str());
    cFuncMeta.funcName = const_cast<char *>(function.funcName.c_str());
    cFuncMeta.functionId = const_cast<char *>(function.functionId.c_str());
    return cFuncMeta;
}

ErrorInfo FunctionExecutionWrapper(const FunctionMeta &function, const libruntime::InvokeType invokeType,
                                   const std::vector<std::shared_ptr<DataObject>> &rawArgs,
                                   std::vector<std::shared_ptr<DataObject>> &returnValues)
{
    CFunctionMeta cFuncMeta = FunctionMetaToCFunctionMeta(function);

    CArg *args = nullptr;
    if (!rawArgs.empty()) {
        args = new (std::nothrow) CArg[rawArgs.size()];
        for (size_t i = 0; i < rawArgs.size(); i++) {
            args[i].type = CArgType::VALUE;
            args[i].data = static_cast<char *>(rawArgs[i]->data->MutableData());
            args[i].size = rawArgs[i]->data->GetSize();
        }
    }

    CDataObject *retObjs = nullptr;
    if (!returnValues.empty()) {
        retObjs = new (std::nothrow) CDataObject[returnValues.size()];
        for (size_t i = 0; i < returnValues.size(); i++) {
            if (returnValues[i]->id.empty()) {
                retObjs[i].id = CString("empty");
            } else {
                retObjs[i].id = CString(returnValues[i]->id);
            }
            retObjs[i].selfSharedPtr = static_cast<void *>(returnValues[i].get());
            retObjs[i].nestedObjIds = nullptr;
            retObjs[i].size_nestedObjIds = 0;
            retObjs[i].buffer.buffer = nullptr;
            retObjs[i].buffer.selfSharedPtrBuffer = nullptr;
            retObjs[i].buffer.size_buffer = 0;
        }
    }

    CErrorInfo *cerr =
        GoFunctionExecution(&cFuncMeta, CInvokeType(invokeType), args, rawArgs.size(), retObjs, returnValues.size());
    for (size_t i = 0; i < returnValues.size(); i++) {
        free(retObjs[i].id);
    }
    delete[] args;
    delete[] retObjs;
    return CErrorToErrorInfo(cerr);
}

ErrorInfo CheckpointWrapper(const std::string &checkpointId, std::shared_ptr<Buffer> &data)
{
    char *cChkptId = const_cast<char *>(checkpointId.c_str());
    CBuffer buf = {0};
    CErrorInfo *cerr = GoCheckpoint(cChkptId, &buf);
    data = std::make_shared<NativeBuffer>(buf.buffer, buf.size_buffer, true);
    return CErrorToErrorInfo(cerr);
}

ErrorInfo RecoverWrapper(std::shared_ptr<Buffer> data)
{
    CBuffer buf = {0};
    buf.buffer = data->MutableData();
    buf.size_buffer = data->GetSize();
    CErrorInfo *cerr = GoRecover(&buf);
    return CErrorToErrorInfo(cerr);
}

ErrorInfo ShutdownWrapper(uint64_t gracePeriodSeconds)
{
    YRLOG_INFO("start execute go shutdown handler");
    CErrorInfo *cerr = GoShutdown(gracePeriodSeconds);
    YRLOG_INFO("end to execute shutdown handler");
    return CErrorToErrorInfo(cerr);
}

ErrorInfo SignalWrapper(int sigNo, std::shared_ptr<Buffer> payload)
{
    CBuffer buf = {0};
    buf.buffer = payload->MutableData();
    buf.size_buffer = payload->GetSize();
    CErrorInfo *cerr = GoSignal(sigNo, &buf);
    return CErrorToErrorInfo(cerr);
}

ErrorInfo HealthCheckWrapper(void)
{
    auto code = GoHealthCheck();
    return CHealthCheckCodeToErrorInfo(code);
}

void FuncExecSubmitHook(std::function<void(void)> &&f)
{
    auto funcPtr = new (std::nothrow) std::shared_ptr<std::function<void(void)>>();
    *funcPtr = std::make_shared<std::function<void(void)>>([f = std::move(f), funcPtr]() {
        f();
        delete funcPtr;
    });
    GoFunctionExecutionPoolSubmit(funcPtr);
}

void ParseAndSetBasicConfig(CLibruntimeConfig *config, LibruntimeConfig &librtCfg)
{
    YR::ParseIpAddr(config->functionSystemAddress, librtCfg.functionSystemIpAddr, librtCfg.functionSystemPort);
    YR::ParseIpAddr(config->grpcAddress, librtCfg.functionSystemRtServerIpAddr, librtCfg.functionSystemRtServerPort);
    YR::ParseIpAddr(config->dataSystemAddress, librtCfg.dataSystemIpAddr, librtCfg.dataSystemPort);
    librtCfg.iamAddress = config->iamAddress;
    librtCfg.runtimeId = config->runtimeId;
    librtCfg.instanceId = config->instanceId;
    librtCfg.functionName = config->functionName;
    librtCfg.logDir = config->logDir;
    librtCfg.logLevel = config->logLevel;
    librtCfg.selfApiType = static_cast<libruntime::ApiType>(config->apiType);
    librtCfg.inCluster = config->inCluster != 0;
    librtCfg.isDriver = config->isDriver != 0;
}

void SetMTLSConfig(CLibruntimeConfig *config, LibruntimeConfig &librtCfg)
{
    librtCfg.enableMTLS = config->enableMTLS != 0;
    librtCfg.privateKeyPath = config->privateKeyPath;
    librtCfg.certificateFilePath = config->certificateFilePath;
    librtCfg.verifyFilePath = config->verifyFilePath;
    librtCfg.primaryKeyStoreFile = config->primaryKeyStoreFile;
    librtCfg.standbyKeyStoreFile = config->standbyKeyStoreFile;
}

ErrorInfo SetEncryptConfig(CLibruntimeConfig *config, LibruntimeConfig &librtCfg)
{
    librtCfg.encryptEnable = config->enableDsEncrypt != 0;
    librtCfg.runtimePublicKeyPath = config->runtimePublicKeyContextPath;
    librtCfg.runtimePrivateKeyPath = config->runtimePrivateKeyContextPath;
    librtCfg.dsPublicKeyPath = config->dsPublicKeyContextPath;
    librtCfg.ak_ = config->systemAuthAccessKey;
    librtCfg.sk_ = datasystem::SensitiveValue(config->systemAuthSecretKey, config->systemAuthSecretKeySize);
    librtCfg.dk_ = datasystem::SensitiveValue(config->systemAuthDataKey, config->systemAuthDataKeySize);
    size_t len = strlen(config->privateKeyPaaswd) + 1;
    (void)memcpy_s(librtCfg.privateKeyPaaswd, YR::Libruntime::MAX_PASSWD_LENGTH, config->privateKeyPaaswd, len);
    return librtCfg.Decrypt();
}

void SetCallbacks(CLibruntimeConfig *config, LibruntimeConfig &librtCfg)
{
    librtCfg.libruntimeOptions.loadFunctionCallback = LoadFunctionsWrapper;
    librtCfg.libruntimeOptions.functionExecuteCallback = FunctionExecutionWrapper;
    librtCfg.libruntimeOptions.checkpointCallback = CheckpointWrapper;
    librtCfg.libruntimeOptions.recoverCallback = RecoverWrapper;
    librtCfg.libruntimeOptions.shutdownCallback = ShutdownWrapper;
    librtCfg.libruntimeOptions.signalCallback = SignalWrapper;
    if (GoHasHealthCheck() != 0) {
        librtCfg.libruntimeOptions.healthCheckCallback = HealthCheckWrapper;
    } else {
        librtCfg.libruntimeOptions.healthCheckCallback = nullptr;
    }
}

void SetFinalConfig(CLibruntimeConfig *config, LibruntimeConfig &librtCfg)
{
    librtCfg.functionIds[libruntime::LanguageType::Golang] = config->functionId;
    librtCfg.selfLanguage = libruntime::LanguageType::Golang;
    librtCfg.jobId = config->jobId;
    librtCfg.funcExecSubmitHook = FuncExecSubmitHook;
    librtCfg.maxConcurrencyCreateNum = config->maxConcurrencyCreateNum;
    librtCfg.enableSigaction = config->enableSigaction;
    librtCfg.enableEvent = config->enableEvent != 0;
}

CErrorInfo CInit(CLibruntimeConfig *config)
{
    LibruntimeConfig librtCfg{};
    ParseAndSetBasicConfig(config, librtCfg);
    SetMTLSConfig(config, librtCfg);
    auto decryptErr = SetEncryptConfig(config, librtCfg);
    if (!decryptErr.OK()) {
        return ErrorInfoToCError(decryptErr);
    }
    SetCallbacks(config, librtCfg);
    SetFinalConfig(config, librtCfg);
    auto err = LibruntimeManager::Instance().Init(librtCfg);
    return ErrorInfoToCError(err);
}

void CReceiveRequestLoop(void)
{
    LibruntimeManager::Instance().ReceiveRequestLoop();
}

void CExecShutdownHandler(int sigNum)
{
    LibruntimeManager::Instance().ExecShutdownCallback(sigNum, false);
}

static FunctionMeta BuildFunctionMeta(CFunctionMeta *cFuncMeta)
{
    FunctionMeta funcMeta;
    funcMeta.apiType = static_cast<libruntime::ApiType>(cFuncMeta->apiType);
    funcMeta.funcName = cFuncMeta->funcName;
    funcMeta.functionId = cFuncMeta->functionId;
    funcMeta.languageType = static_cast<libruntime::LanguageType>(cFuncMeta->languageType);
    funcMeta.signature = cFuncMeta->signature;
    funcMeta.poolLabel = cFuncMeta->poolLabel;
    if (cFuncMeta->hasName) {
        funcMeta.name = cFuncMeta->name;
    }
    if (cFuncMeta->hasNs) {
        funcMeta.ns = cFuncMeta->ns;
    }
    return funcMeta;
}

static std::vector<InvokeArg> BuildInvokeArgs(CInvokeArg *cInvokeArgs, int size_invokeArgs)
{
    std::vector<InvokeArg> invokeArgs;
    for (int i = 0; i < size_invokeArgs; i++) {
        CInvokeArg &cArg = cInvokeArgs[i];
        InvokeArg arg;
        if (cArg.isRef) {
            // need to process
        } else {
            arg.isRef = false;
            arg.dataObj = std::make_shared<YR::Libruntime::DataObject>(0, cArg.size_buf);
            // copy arg to DataObject->data
            arg.dataObj->data->MemoryCopy(cArg.buf, cArg.size_buf);
        }
        for (int i = 0; i < cArg.size_nestedObjects; i++) {
            arg.nestedObjects.emplace(cArg.nestedObjects[i]);
        }
        arg.tenantId = cArg.tenantId;
        invokeArgs.emplace_back(std::move(arg));
    }
    return invokeArgs;
}

static std::list<std::shared_ptr<LabelOperator>> BuildLabelOperators(CLabelOperator *cLabelOps, int cLabelOpsLen)
{
    std::list<std::shared_ptr<LabelOperator>> labelOps;
    for (int i = 0; i < cLabelOpsLen; i++) {
        std::shared_ptr<LabelOperator> labelOp;
        CLabelOperator &cLabelOp = cLabelOps[i];
        switch (cLabelOp.opType) {
            case CLabelOpType::IN:
                labelOp = std::make_shared<LabelInOperator>();
                break;
            case CLabelOpType::NOT_IN:
                labelOp = std::make_shared<LabelNotInOperator>();
                break;
            case CLabelOpType::EXISTS:
                labelOp = std::make_shared<LabelExistsOperator>();
                break;
            case CLabelOpType::NOT_EXISTS:
                labelOp = std::make_shared<LabelDoesNotExistOperator>();
                break;
            default:
                labelOp = std::make_shared<LabelInOperator>();
                break;
        }
        labelOp->SetKey(cLabelOp.labelKey);
        std::list<std::string> labelValues;
        for (int j = 0; j < cLabelOp.size_labelValues; j++) {
            labelValues.push_back(cLabelOp.labelValues[j]);
        }
        labelOp->SetValues(labelValues);

        labelOps.push_back(labelOp);
    }
    return labelOps;
}

static std::shared_ptr<Affinity> BuildScheduleAffinity(CAffinity &cAffinity)
{
    std::string kind;
    switch (cAffinity.affKind) {
        case CAffinityKind::RESOURCE:
            kind = YR::Libruntime::RESOURCE;
            break;
        case CAffinityKind::INSTANCE:
            kind = YR::Libruntime::INSTANCE;
            break;
        default:
            kind = YR::Libruntime::RESOURCE;
            break;
    }

    std::string type;
    switch (cAffinity.affType) {
        case CAffinityType::PREFERRED:
            type = YR::Libruntime::PREFERRED;
            break;
        case CAffinityType::PREFERRED_ANTI:
            type = YR::Libruntime::PREFERRED_ANTI;
            break;
        case CAffinityType::REQUIRED:
            type = YR::Libruntime::REQUIRED;
            break;
        case CAffinityType::REQUIRED_ANTI:
            type = YR::Libruntime::REQUIRED_ANTI;
            break;
        default:
            type = YR::Libruntime::PREFERRED;
            break;
    }
    using AffnitiyCreator = std::function<std::shared_ptr<Affinity>()>;
    const std::unordered_map<std::string, AffnitiyCreator> affinityMap{
        {"ResourcePreferredAffinity", [] { return std::make_shared<YR::Libruntime::ResourcePreferredAffinity>(); }},
        {"ResourcePreferredAntiAffinity",
         [] { return std::make_shared<YR::Libruntime::ResourcePreferredAntiAffinity>(); }},
        {"ResourceRequiredAffinity", [] { return std::make_shared<YR::Libruntime::ResourceRequiredAffinity>(); }},
        {"ResourceRequiredAntiAffinity",
         [] { return std::make_shared<YR::Libruntime::ResourceRequiredAntiAffinity>(); }},
        {"InstancePreferredAffinity", [] { return std::make_shared<YR::Libruntime::InstancePreferredAffinity>(); }},
        {"InstancePreferredAntiAffinity",
         [] { return std::make_shared<YR::Libruntime::InstancePreferredAntiAffinity>(); }},
        {"InstanceRequiredAffinity", [] { return std::make_shared<YR::Libruntime::InstanceRequiredAffinity>(); }},
        {"InstanceRequiredAntiAffinity",
         [] { return std::make_shared<YR::Libruntime::InstanceRequiredAntiAffinity>(); }}};
    std::string key = kind + type;
    std::shared_ptr<Affinity> aff;
    auto it = affinityMap.find(key);
    if (it != affinityMap.end()) {
        aff = it->second();
    } else {
        aff = std::make_shared<Affinity>(kind, type);
    }
    aff->SetLabelOperators(BuildLabelOperators(cAffinity.labelOps, cAffinity.size_labelOps));
    aff->SetPreferredPriority(cAffinity.preferredPrio == 0 ? false : true);
    aff->SetPreferredAntiOtherLabels(cAffinity.preferredAntiOtherLabels == 0 ? false : true);
    return aff;
}

static InvokeOptions BuildInvokeOptions(CInvokeOptions *cInvokeOpts)
{
    InvokeOptions invokeOpts;
    invokeOpts.cpu = cInvokeOpts->cpu;
    invokeOpts.memory = cInvokeOpts->memory;
    invokeOpts.retryTimes = cInvokeOpts->RetryTimes;
    invokeOpts.recoverRetryTimes = cInvokeOpts->RecoverRetryTimes;
    invokeOpts.timeout = cInvokeOpts->timeout;
    invokeOpts.acquireTimeout = cInvokeOpts->acquireTimeout;
    invokeOpts.scheduleTimeoutMs = cInvokeOpts->scheduleTimeoutMs;
    for (int i = 0; i < cInvokeOpts->size_customResources; i++) {
        invokeOpts.customResources.emplace(cInvokeOpts->customResources[i].name,
                                           cInvokeOpts->customResources[i].scalar);
    }
    for (int i = 0; i < cInvokeOpts->size_customExtensions; i++) {
        invokeOpts.customExtensions.emplace(cInvokeOpts->customExtensions[i].key,
                                            cInvokeOpts->customExtensions[i].value);
    }
    for (int i = 0; i < cInvokeOpts->size_createOpt; i++) {
        invokeOpts.createOptions.emplace(cInvokeOpts->createOpt[i].key, cInvokeOpts->createOpt[i].value);
    }
    for (int i = 0; i < cInvokeOpts->size_schedulerInstanceIds; i++) {
        invokeOpts.schedulerInstanceIds.emplace_back(cInvokeOpts->schedulerInstanceIds[i]);
    }
    for (int i = 0; i < cInvokeOpts->size_labels; i++) {
        invokeOpts.labels.emplace_back(cInvokeOpts->labels[i]);
    }
    for (int i = 0; i < cInvokeOpts->size_schedAffinities; i++) {
        invokeOpts.scheduleAffinities.push_back(BuildScheduleAffinity(cInvokeOpts->schedAffinities[i]));
    }
    for (int i = 0; i < cInvokeOpts->size_codePaths; i++) {
        invokeOpts.codePaths.emplace_back(cInvokeOpts->codePaths[i]);
    }
    if (cInvokeOpts->schedulerFunctionId != nullptr) {
        invokeOpts.schedulerFunctionId = cInvokeOpts->schedulerFunctionId;
    }
    if (cInvokeOpts->traceId != nullptr) {
        invokeOpts.traceId = cInvokeOpts->traceId;
    }
    if (cInvokeOpts->trafficLimited != 0) {
        invokeOpts.trafficLimited = true;
    }
    if (cInvokeOpts->forceInvoke != 0) {
        invokeOpts.forceInvoke = true;
    }
    for (int i = 0; i < cInvokeOpts->size_invokeLabels; i++) {
        invokeOpts.invokeLabels.emplace(cInvokeOpts->invokeLabels[i].key, cInvokeOpts->invokeLabels[i].value);
    }
    if (cInvokeOpts->instanceSession != nullptr) {
        invokeOpts.instanceSession = std::make_shared<YR::Libruntime::InstanceSession>();
        invokeOpts.instanceSession->sessionID = cInvokeOpts->instanceSession->sessionId;
        invokeOpts.instanceSession->sessionTTL = cInvokeOpts->instanceSession->sessionTtl;
        invokeOpts.instanceSession->concurrency = cInvokeOpts->instanceSession->concurrency;
    }
    return invokeOpts;
}

static ProducerConf BuildProducerConfig(CProducerConfig *config)
{
    ProducerConf producerConfig;
    producerConfig.delayFlushTime = config->delayFlushTime;
    producerConfig.pageSize = config->pageSize;
    producerConfig.maxStreamSize = config->maxStreamSize;
    std::string traceId(config->traceId);
    producerConfig.traceId = traceId;
    return producerConfig;
}

static SubscriptionConfig BuildSubscriptionConfig(CSubscriptionConfig *config)
{
    SubscriptionConfig subscriptionConfig;
    std::string subName(config->subscriptionName);
    subscriptionConfig.subscriptionName = subName;
    subscriptionConfig.subscriptionType = SubscriptionType::STREAM;
    std::string traceId(config->traceId);
    subscriptionConfig.traceId = traceId;
    return subscriptionConfig;
}

static Element BuildElement(uint8_t *ptr, uint64_t size, uint64_t id)
{
    Element ele;
    ele.ptr = ptr;
    ele.size = size;
    ele.id = id;
    return ele;
}

void FreeBuffers(CBuffer *cData, size_t size)
{
    for (size_t j = 0; j < size; j++) {
        cData[j].size_buffer = 0;
        if (cData[j].buffer != nullptr) {
            free(cData[j].buffer);
            cData[j].buffer = nullptr;
        }
    }
}

CErrorInfo CCreateInstance(CFunctionMeta *cFuncMeta, CInvokeArg *cInvokeArgs, int size_invokeArgs,
                           CInvokeOptions *cInvokeOpts, char **instanceId)
{
    auto funcMeta = BuildFunctionMeta(cFuncMeta);
    auto invokeArgs = BuildInvokeArgs(cInvokeArgs, size_invokeArgs);
    auto invokeOpts = BuildInvokeOptions(cInvokeOpts);

    auto [lrt, errorInfo] = getLibRuntime();
    if (!errorInfo.OK()) {
        return ErrorInfoToCError(errorInfo);
    }
    auto [err, instId] = lrt->CreateInstance(funcMeta, invokeArgs, invokeOpts);
    if (err.OK()) {
        *instanceId = CString(instId);
    }
    return ErrorInfoToCError(err);
}
CErrorInfo CInvokeByInstanceId(CFunctionMeta *cFuncMeta, char *cInstanceId, CInvokeArg *cInvokeArgs,
                               int size_invokeArgs, CInvokeOptions *cInvokeOpts, char **cReturnObjectId)
{
    auto funcMeta = BuildFunctionMeta(cFuncMeta);
    std::string instanceId(cInstanceId);
    auto invokeArgs = BuildInvokeArgs(cInvokeArgs, size_invokeArgs);
    auto invokeOpts = BuildInvokeOptions(cInvokeOpts);
    std::vector<DataObject> returnObjs{{""}};
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->InvokeByInstanceId(funcMeta, instanceId, invokeArgs, invokeOpts, returnObjs);
    if (err.OK()) {
        *cReturnObjectId = CString(returnObjs[0].id);
    }
    return ErrorInfoToCError(err);
}

CErrorInfo CInvokeByFunctionName(CFunctionMeta *cFuncMeta, CInvokeArg *cInvokeArgs, int size_invokeArgs,
                                 CInvokeOptions *cInvokeOpts, char **cReturnObjectId)
{
    auto funcMeta = BuildFunctionMeta(cFuncMeta);
    auto invokeArgs = BuildInvokeArgs(cInvokeArgs, size_invokeArgs);
    auto invokeOpts = BuildInvokeOptions(cInvokeOpts);
    std::vector<DataObject> returnObjs{{""}};
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->InvokeByFunctionName(funcMeta, invokeArgs, invokeOpts, returnObjs);
    if (err.OK()) {
        *cReturnObjectId = CString(returnObjs[0].id);
    }
    return ErrorInfoToCError(err);
}

CErrorInfo CAcquireInstance(char *stateId, CFunctionMeta *cFuncMeta, CInvokeOptions *cInvokeOpts,
                            CInstanceAllocation *cInsAlloc)
{
    auto funcMeta = BuildFunctionMeta(cFuncMeta);
    auto invokeOpts = BuildInvokeOptions(cInvokeOpts);
    std::string state(stateId);
    auto [lrt, err1] = getLibRuntime();
    if (!err1.OK()) {
        return ErrorInfoToCError(err1);
    }
    auto [instanceAllocation, err2] = lrt->AcquireInstance(state, funcMeta, invokeOpts);
    InsAllocationToCInsAllocation(instanceAllocation, cInsAlloc);
    return ErrorInfoToCError(err2);
}

CErrorInfo CReleaseInstance(CInstanceAllocation *insAlloc, char *cStateID, char cAbnormal, CInvokeOptions *cInvokeOpts)
{
    std::string leaseID(insAlloc->leaseId);
    std::string stateID(cStateID);
    auto invokeOpts = BuildInvokeOptions(cInvokeOpts);
    bool abnormal = cAbnormal == 0 ? false : true;
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->ReleaseInstance(leaseID, stateID, abnormal, invokeOpts);
    return ErrorInfoToCError(err);
}

void RawCallbackWrapper(const std::string context, const ErrorInfo &err, std::shared_ptr<Buffer> resultRaw)
{
    CBuffer cResult = {0};
    cResult.buffer = resultRaw->MutableData();
    cResult.size_buffer = resultRaw->GetSize();
    GoRawCallback(const_cast<char *>(context.c_str()), ErrorInfoToCError(err), cResult);
}

void CCreateInstanceRaw(CBuffer cReqRaw, char *cContext)
{
    auto reqRaw = std::make_shared<NativeBuffer>(cReqRaw.buffer, cReqRaw.size_buffer);
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return;  // 以后把报错抛出去
    }
    lrt->CreateInstanceRaw(reqRaw, std::bind(RawCallbackWrapper, std::string(cContext), _1, _2));
}

void CInvokeByInstanceIdRaw(CBuffer cReqRaw, char *cContext)
{
    auto reqRaw = std::make_shared<NativeBuffer>(cReqRaw.buffer, cReqRaw.size_buffer);
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return;  // 以后把报错抛出去
    }
    lrt->InvokeByInstanceIdRaw(reqRaw, std::bind(RawCallbackWrapper, std::string(cContext), _1, _2));
}

void CKillRaw(CBuffer cReqRaw, char *cContext)
{
    auto reqRaw = std::make_shared<NativeBuffer>(cReqRaw.buffer, cReqRaw.size_buffer);
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return;  // 以后把报错抛出去
    }
    lrt->KillRaw(reqRaw, std::bind(RawCallbackWrapper, std::string(cContext), _1, _2));
}

ErrorInfo ToCBuffer(std::shared_ptr<Buffer> buf, CBuffer *data)
{
    data->size_buffer = buf->GetSize();
    if (data->size_buffer == 0) {
        data->buffer = nullptr;
        return ErrorInfo();
    }
    data->buffer = malloc(data->size_buffer);
    if (data->buffer != nullptr) {
        int err = memcpy_s(data->buffer, data->size_buffer, buf->ImmutableData(), buf->GetSize());
        if (err != EOK) {
            free(data->buffer);
            data->buffer = nullptr;
            YRLOG_ERROR("CGet memcpy_s failed: {}", err);
            return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                             "CGet memcpy_s failed");
        }
    } else {
        YRLOG_ERROR("CGet malloc failed");
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "CGet malloc failed");
    }
    return ErrorInfo();
}

CErrorInfo CGet(char *objId, int timeoutSec, CBuffer *data)
{
    auto [lrt, err1] = getLibRuntime();
    if (!err1.OK()) {
        return ErrorInfoToCError(err1);
    }
    auto [err2, res] = lrt->Get({objId}, timeoutSec, false);
    if (err2.OK()) {
        return ErrorInfoToCError(ToCBuffer(res[0]->data, data));
    }
    return ErrorInfoToCError(err2);
}

void CUpdateSchdulerInfo(char *scheduleName, char *schedulerId, char *option)
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return;  // 以后把报错抛出去
    }
    lrt->UpdateSchdulerInfo(scheduleName, schedulerId, option);
}

void CGetAsync(char *objectId, void *userData)
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return;  // 以后把报错抛出去
    }
    lrt->GetAsync(
        objectId,
        [](std::shared_ptr<DataObject> data, const ErrorInfo &err, void *userData) {
            auto cErr = ErrorInfoToCError(err);
            CBuffer cBuf = {0};
            if (err.OK()) {
                cErr = ErrorInfoToCError(ToCBuffer(data->data, &cBuf));
            }
            auto cObjectId = const_cast<char *>(data->id.c_str());
            GoGetAsyncCallback(cObjectId, cBuf, &cErr, userData);
        },
        userData);
}

void CWaitAsync(char *objectId, void *userData)
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return;  // 以后把报错抛出去
    }
    lrt->WaitAsync(
        objectId,
        [](const std::string &objId, const ErrorInfo &err, void *userData) {
            auto cErr = ErrorInfoToCError(err);
            auto cObjectId = const_cast<char *>(objId.c_str());
            GoWaitAsyncCallback(cObjectId, &cErr, userData);
        },
        userData);
}

void CGetEvent(char *objectId, void *userData)
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return;  // 以后把报错抛出去
    }
    lrt->GetEvent(
        objectId,
        [](std::shared_ptr<DataObject> data, const ErrorInfo &err, void *userData) {
            auto cErr = ErrorInfoToCError(err);
            CBuffer cBuf = {0};
            if (err.OK()) {
                cErr = ErrorInfoToCError(ToCBuffer(data->buffer, &cBuf));
            }
            auto cObjectId = const_cast<char *>(data->id.c_str());
            GoGetEventCallback(cObjectId, cBuf, &cErr, userData);
        },
        userData);
}

CErrorInfo CKill(char *instanceId, int sigNo, CBuffer cData)
{
    std::shared_ptr<NativeBuffer> data;
    if (cData.buffer != nullptr) {
        data = std::make_shared<NativeBuffer>(cData.buffer, cData.size_buffer);
    }
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    if (data) {
        err = lrt->Kill(instanceId, sigNo, data);
    } else {
        err = lrt->Kill(instanceId, sigNo);
    }
    return ErrorInfoToCError(err);
}

char *CGetRealInstanceId(char *objectId, int timeout)
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return CString("");
    }
    auto instId = lrt->GetRealInstanceId(objectId, timeout);
    return CString(instId);
}

void CExit(int code, char *message)
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return;  // 以后把报错抛出去
    }
    lrt->Exit(code, message);
}

void CFinalize(void)
{
    LibruntimeManager::Instance().Finalize();
}

CErrorInfo CSetTenantId(const char *cTenantId, int cTenantIdLen)
{
    std::string tenantId = "";
    CheckNullAndAssignValue(cTenantId, cTenantIdLen, tenantId);
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->SetTenantId(tenantId, true);
    return ErrorInfoToCError(err);
}

CErrorInfo CPutCommon(char *cObjectId, CBuffer cData, char **cNestedIds, int size_cNestedIds, char isPutRaw,
                      CCreateParam param)
{
    std::unordered_set<std::string> nestedIds;
    for (int i = 0; i < size_cNestedIds; i++) {
        nestedIds.emplace(cNestedIds[i]);
    }
    CreateParam createParam = CCreateParamToCreateParam(param);
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    if (isPutRaw) {
        auto data = std::make_shared<YR::Libruntime::NativeBuffer>(cData.size_buffer);
        data->MemoryCopy(cData.buffer, cData.size_buffer);
        err = lrt->PutRaw(cObjectId, data, nestedIds, createParam);
        return ErrorInfoToCError(err);
    }
    auto dataObj = std::make_shared<YR::Libruntime::DataObject>(0, cData.size_buffer);
    dataObj->data->MemoryCopy(cData.buffer, cData.size_buffer);
    err = lrt->Put(cObjectId, dataObj, nestedIds, createParam);
    return ErrorInfoToCError(err);
}

CErrorInfo CGetMultiCommon(char **cObjIds, int size_cObjIds, int timeoutMs, char allowPartial, CBuffer *cData,
                           char isRaw)
{
    std::vector<std::string> objectIds(size_cObjIds);
    for (size_t i = 0; i < objectIds.size(); i++) {
        objectIds[i] = cObjIds[i];
    }
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }

    bool allowPart = allowPartial == 0 ? false : true;
    std::pair<ErrorInfo, std::vector<std::shared_ptr<Buffer>>> getRet;
    if (isRaw) {
        getRet = lrt->GetRaw(objectIds, timeoutMs, allowPart);
    } else {
        std::pair<ErrorInfo, std::vector<std::shared_ptr<DataObject>>> getDataObjRet;
        getDataObjRet = lrt->Get(objectIds, timeoutMs, allowPart);
        getRet.second.resize(getDataObjRet.second.size());
        for (size_t i = 0; i < getDataObjRet.second.size(); i++) {
            if (getDataObjRet.second[i]) {
                getRet.second[i] = getDataObjRet.second[i]->data;
            }
        }
    }
    auto [err1, data] = getRet;

    for (size_t i = 0; i < data.size(); i++) {
        if (data[i] == nullptr) {
            continue;
        }
        auto errInfo = ToCBuffer(data[i], &cData[i]);
        if (!errInfo.OK()) {
            FreeBuffers(cData, i);
            return ErrorInfoToCError(errInfo);
        }
    }
    return ErrorInfoToCError(err1);
}

void CWait(char **objIds, int size_objIds, int waitNum, int timeoutSec, CWaitResult *result)
{
    std::vector<std::string> objectIds(size_objIds);
    for (size_t i = 0; i < objectIds.size(); i++) {
        objectIds[i] = objIds[i];
    }

    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return;  // 以后把报错抛出去
    }
    auto waitResults = lrt->Wait(objectIds, waitNum, timeoutSec);
    if (!waitResults->readyIds.empty()) {
        StringsToCStrings(waitResults->readyIds, &(result->readyIds), &(result->size_readyIds));
    }

    if (!waitResults->unreadyIds.empty()) {
        StringsToCStrings(waitResults->unreadyIds, &(result->unreadyIds), &(result->size_unreadyIds));
    }
    if (!waitResults->exceptionIds.empty()) {
        result->errorIds = new (std::nothrow) CErrorObject *[waitResults->exceptionIds.size()];
        int index = 0;
        for (auto iter = waitResults->exceptionIds.begin(); iter != waitResults->exceptionIds.end(); iter++) {
            result->errorIds[index] = new (std::nothrow) CErrorObject();

            result->errorIds[index]->objectId = new (std::nothrow) char[iter->first.size()];
            memcpy_s(result->errorIds[index]->objectId, iter->first.size(), iter->first.c_str(), iter->first.size());

            result->errorIds[index]->errorInfo = new (std::nothrow) CErrorInfo();
            auto cError = ErrorInfoToCError(iter->second);
            result->errorIds[index]->errorInfo->code = cError.code;
            result->errorIds[index]->errorInfo->message = cError.message;
            if (cError.size_stackTracesInfo == 0) {
                return;
            }
            result->errorIds[index]->errorInfo->stackTracesInfo = cError.stackTracesInfo;
            result->errorIds[index]->errorInfo->size_stackTracesInfo = cError.size_stackTracesInfo;
            result->size_errorIds++;
            index++;
        }
    }
}

CErrorInfo CIncreaseReferenceCommon(char **cObjIds, int size_cObjIds, char *cRemoteId, char ***cFailedIds,
                                    int *size_cFailedIds, char isRaw)
{
    std::vector<std::string> objectIds(size_cObjIds);
    for (size_t i = 0; i < objectIds.size(); i++) {
        objectIds[i] = cObjIds[i];
    }
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    if (cRemoteId == nullptr) {
        if (isRaw) {
            err = lrt->IncreaseReferenceRaw(objectIds);
            return ErrorInfoToCError(err);
        } else {
            err = lrt->IncreaseReference(objectIds);
            return ErrorInfoToCError(err);
        }
    }

    std::pair<ErrorInfo, std::vector<std::string>> increRet;
    if (isRaw) {
        increRet = lrt->IncreaseReferenceRaw(objectIds, cRemoteId);
    } else {
        increRet = lrt->IncreaseReference(objectIds, std::string(cRemoteId));
    }
    auto [err1, failedIds] = increRet;
    StringsToCStrings(failedIds, cFailedIds, size_cFailedIds);
    return ErrorInfoToCError(err1);
}

CErrorInfo CDecreaseReferenceCommon(char **cObjIds, int size_cObjIds, char *cRemoteId, char ***cFailedIds,
                                    int *size_cFailedIds, char isRaw)
{
    std::vector<std::string> objectIds(size_cObjIds);
    for (size_t i = 0; i < objectIds.size(); i++) {
        objectIds[i] = cObjIds[i];
    }

    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    std::pair<ErrorInfo, std::vector<std::string>> decreRet;
    if (cRemoteId) {
        if (isRaw) {
            decreRet = lrt->DecreaseReferenceRaw(objectIds, cRemoteId);
        } else {
            decreRet = lrt->DecreaseReference(objectIds, cRemoteId);
        }
    } else {
        if (isRaw) {
            lrt->DecreaseReferenceRaw(objectIds);
        } else {
            lrt->DecreaseReference(objectIds);
        }
    }
    auto [err1, failedIds] = decreRet;
    StringsToCStrings(failedIds, cFailedIds, size_cFailedIds);
    return ErrorInfoToCError(err1);
}

CErrorInfo CReleaseGRefs(char *cRemoteId)
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->ReleaseGRefs(cRemoteId);
    return ErrorInfoToCError(err);
}

CErrorInfo CKVWrite(char *key, CBuffer data, CSetParam param)
{
    auto mData = std::make_shared<YR::Libruntime::NativeBuffer>(data.size_buffer);
    mData->MemoryCopy(data.buffer, data.size_buffer);
    SetParam setParam = CSetParamToSetParam(param);
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->KVWrite(key, mData, setParam);
    return ErrorInfoToCError(err);
}

CErrorInfo CKVMSetTx(char **cKeys, int sizeCKeys, CBuffer *data, CMSetParam param)
{
    std::vector<std::string> keys(sizeCKeys);
    for (size_t i = 0; i < keys.size(); i++) {
        keys[i] = cKeys[i];
    }
    std::vector<std::shared_ptr<Buffer>> vals(sizeCKeys);
    for (size_t i = 0; i < keys.size(); i++) {
        auto mData = std::make_shared<YR::Libruntime::NativeBuffer>(data[i].size_buffer);
        mData->MemoryCopy(data[i].buffer, data[i].size_buffer);
        vals[i] = mData;
    }
    MSetParam mSetParam = CMSetParamToMSetParam(param);

    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->KVMSetTx(keys, vals, mSetParam);
    return ErrorInfoToCError(err);
}

CErrorInfo CKVRead(char *key, int timeoutMs, CBuffer *cData)
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    auto [value, err1] = lrt->KVRead(key, timeoutMs);
    if (err1.OK()) {
        return ErrorInfoToCError(ToCBuffer(value, cData));
    }
    return ErrorInfoToCError(err1);
}

CErrorInfo CKVMultiRead(char **cKeys, int size_cKeys, int timeoutMs, char allowPartial, CBuffer *cData)
{
    std::vector<std::string> keys(size_cKeys);
    for (size_t i = 0; i < keys.size(); i++) {
        keys[i] = cKeys[i];
    }
    bool allowPart = allowPartial == 0 ? false : true;
    auto [lrt, err0] = getLibRuntime();
    if (!err0.OK()) {
        return ErrorInfoToCError(err0);
    }
    auto [values, err] = lrt->KVRead(keys, timeoutMs, allowPart);
    for (int i = 0; i < size_cKeys; i++) {
        if (values[i] == nullptr) {
            continue;
        }
        auto errInfo = ToCBuffer(values[i], &cData[i]);
        if (!errInfo.OK()) {
            FreeBuffers(cData, i);
            return ErrorInfoToCError(errInfo);
        }
    }
    return ErrorInfoToCError(err);
}

CErrorInfo CKVDel(char *key)
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->KVDel(key);
    return ErrorInfoToCError(err);
}

CErrorInfo CKMultiVDel(char **cKeys, int size_cKeys, char ***cFailedKeys, int *size_cFailedKeys)
{
    std::vector<std::string> keys(size_cKeys);
    for (size_t i = 0; i < keys.size(); i++) {
        keys[i] = cKeys[i];
    }
    auto [lrt, err0] = getLibRuntime();
    if (!err0.OK()) {
        return ErrorInfoToCError(err0);
    }
    auto [failedKeys, err] = lrt->KVDel(keys);
    StringsToCStrings(failedKeys, cFailedKeys, size_cFailedKeys);
    return ErrorInfoToCError(err);
}

CErrorInfo CCreateStreamProducer(char *cStreamName, CProducerConfig *config, Producer_p *producer)
{
    std::string streamName(cStreamName);
    auto producerConfig = BuildProducerConfig(config);
    std::shared_ptr<StreamProducer> outProducer;
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->CreateStreamProducer(streamName, producerConfig, outProducer);
    if (err.OK()) {
        std::unique_ptr<std::shared_ptr<StreamProducer>> pOutProducer =
            std::make_unique<std::shared_ptr<StreamProducer>>(std::move(outProducer));
        *producer = reinterpret_cast<void *>(pOutProducer.release());
    }
    return ErrorInfoToCError(err);
}

CErrorInfo CCreateStreamConsumer(char *cStreamName, CSubscriptionConfig *config, Consumer_p *consumer)
{
    std::string streamName(cStreamName);
    auto subscriptionConfig = BuildSubscriptionConfig(config);
    std::shared_ptr<StreamConsumer> outConsumer;
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->CreateStreamConsumer(streamName, subscriptionConfig, outConsumer);
    if (err.OK()) {
        *consumer = reinterpret_cast<void *>(new shared_ptr<StreamConsumer>(std::move(outConsumer)));
    }
    return ErrorInfoToCError(err);
}

CErrorInfo CDeleteStream(char *cStreamName)
{
    std::string streamName(cStreamName);
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->DeleteStream(streamName);
    return ErrorInfoToCError(err);
}

CErrorInfo CQueryGlobalProducersNum(char *cStreamName, uint64_t *num)
{
    std::string streamName(cStreamName);
    uint64_t gProducerNum = 0;
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->QueryGlobalProducersNum(streamName, gProducerNum);
    *num = gProducerNum;
    return ErrorInfoToCError(err);
}

CErrorInfo CQueryGlobalConsumersNum(char *cStreamName, uint64_t *num)
{
    std::string streamName(cStreamName);
    uint64_t gConsumerNum = 0;
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->QueryGlobalConsumersNum(streamName, gConsumerNum);
    *num = gConsumerNum;
    return ErrorInfoToCError(err);
}

CErrorInfo CProducerSend(Producer_p producerPtr, uint8_t *ptr, uint64_t size, uint64_t id)
{
    auto element = BuildElement(ptr, size, id);
    auto producer = *reinterpret_cast<std::shared_ptr<StreamProducer> *>(producerPtr);
    auto err = producer->Send(element);
    return ErrorInfoToCError(err);
}

CErrorInfo CProducerSendWithTimeout(Producer_p producerPtr, uint8_t *ptr, uint64_t size, uint64_t id, int64_t timeoutMs)
{
    auto element = BuildElement(ptr, size, id);
    auto producer = *reinterpret_cast<std::shared_ptr<StreamProducer> *>(producerPtr);
    auto err = producer->Send(element, timeoutMs);
    return ErrorInfoToCError(err);
}

CErrorInfo CProducerClose(Producer_p producerPtr)
{
    auto producer = reinterpret_cast<std::shared_ptr<StreamProducer> *>(producerPtr);
    RETURN_ERR_WHEN_CONSUMER_ISNULL(producer);
    auto err = (*producer)->Close();
    delete producer;
    return ErrorInfoToCError(err);
}

CErrorInfo CConsumerReceive(Consumer_p consumerPtr, uint32_t timeoutMs, CElement **elements, uint64_t *count)
{
    std::vector<Element> eles;
    auto consumer = reinterpret_cast<std::shared_ptr<StreamConsumer> *>(consumerPtr);
    RETURN_ERR_WHEN_CONSUMER_ISNULL(consumer);
    auto err = (*consumer)->Receive(timeoutMs, eles);
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    *count = eles.size();
    std::unique_ptr<CElement[]> pEles = std::make_unique<CElement[]>(eles.size());
    for (size_t i = 0; i < eles.size(); i++) {
        pEles[i].id = eles[i].id;
        pEles[i].ptr = eles[i].ptr;
        pEles[i].size = eles[i].size;
    }
    *elements = pEles.release();
    return ErrorInfoToCError(err);
}

CErrorInfo CConsumerReceiveExpectNum(Consumer_p consumerPtr, uint32_t expectNum, uint32_t timeoutMs,
                                     CElement **elements, uint64_t *count)
{
    std::vector<Element> eles;
    auto consumer = reinterpret_cast<std::shared_ptr<StreamConsumer> *>(consumerPtr);
    RETURN_ERR_WHEN_CONSUMER_ISNULL(consumer);
    auto err = (*consumer)->Receive(expectNum, timeoutMs, eles);
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    *count = eles.size();
    std::unique_ptr<CElement[]> pEles = std::make_unique<CElement[]>(eles.size());
    for (size_t i = 0; i < eles.size(); i++) {
        pEles[i].id = eles[i].id;
        pEles[i].ptr = eles[i].ptr;
        pEles[i].size = eles[i].size;
    }
    *elements = pEles.release();
    return ErrorInfoToCError(err);
}

CErrorInfo CConsumerAck(Consumer_p consumerPtr, uint64_t elementId)
{
    auto consumer = reinterpret_cast<std::shared_ptr<StreamConsumer> *>(consumerPtr);
    RETURN_ERR_WHEN_CONSUMER_ISNULL(consumer);
    auto err = (*consumer)->Ack(elementId);
    return ErrorInfoToCError(err);
}

CErrorInfo CConsumerClose(Consumer_p consumerPtr)
{
    auto consumer = reinterpret_cast<std::shared_ptr<StreamConsumer> *>(consumerPtr);
    RETURN_ERR_WHEN_CONSUMER_ISNULL(consumer);
    auto err = (*consumer)->Close();
    delete consumer;
    return ErrorInfoToCError(err);
}

CErrorInfo CAllocReturnObject(CDataObject *cObject, int dataSize, char **cNestedIds, int size_cNestedIds,
                              uint64_t *totalNativeBufferSize)
{
    auto dataObject = static_cast<DataObject *>(cObject->selfSharedPtr);
    dataObject->alwaysNative = true;
    std::vector<std::string> nestedIds(size_cNestedIds);
    for (size_t i = 0; i < nestedIds.size(); i++) {
        nestedIds[i] = cNestedIds[i];
    }
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->AllocReturnObject(dataObject, 0, (size_t)dataSize, nestedIds, *totalNativeBufferSize);
    if (err.OK()) {
        cObject->buffer.buffer = dataObject->data->MutableData();
        cObject->buffer.size_buffer = dataObject->data->GetSize();
        cObject->buffer.selfSharedPtrBuffer = static_cast<void *>(dataObject->data.get());
    }
    return ErrorInfoToCError(err);
}

void CSetReturnObject(CDataObject *cObject, int dataSize)
{
    auto dataObject = static_cast<DataObject *>(cObject->selfSharedPtr);
    std::shared_ptr<Buffer> dataBuf = std::make_shared<NativeBuffer>(dataSize);
    dataObject->SetDataBuf(dataBuf);
    cObject->buffer.buffer = dataObject->data->MutableData();
    cObject->buffer.size_buffer = dataObject->data->GetSize();
    cObject->buffer.selfSharedPtrBuffer = static_cast<void *>(dataObject->data.get());
}

CErrorInfo CWriterLatch(CBuffer *cBuffer)
{
    Buffer *buffer = static_cast<Buffer *>(cBuffer->selfSharedPtrBuffer);
    auto err = buffer->WriterLatch();
    return ErrorInfoToCError(err);
}

CErrorInfo CMemoryCopy(CBuffer *cBuffer, void *cSrc, uint64_t size_cSrc)
{
    Buffer *buffer = static_cast<Buffer *>(cBuffer->selfSharedPtrBuffer);
    auto err = buffer->MemoryCopy(cSrc, size_cSrc);
    return ErrorInfoToCError(err);
}

CErrorInfo CSeal(CBuffer *cBuffer)
{
    Buffer *buffer = static_cast<Buffer *>(cBuffer->selfSharedPtrBuffer);
    std::unordered_set<std::string> nestedIds;  // support Seal with nestedIDs
    auto err = buffer->Seal(nestedIds);
    return ErrorInfoToCError(err);
}

CErrorInfo CWriterUnlatch(CBuffer *cBuffer)
{
    Buffer *buffer = static_cast<Buffer *>(cBuffer->selfSharedPtrBuffer);
    auto err = buffer->WriterUnlatch();
    return ErrorInfoToCError(err);
}

void CParseCErrorObjectPointer(CErrorObject *object, int *code, char **errMessage, char **objectId,
                               CStackTracesInfo *stackTracesInfo)
{
    if (object == nullptr) {
        return;
    }
    if (object->errorInfo == nullptr) {
        return;
    }
    *code = object->errorInfo->code;
    *errMessage = object->errorInfo->message;
    stackTracesInfo->size_stackTraces = object->errorInfo->stackTracesInfo[0].size_stackTraces;
    stackTracesInfo->stackTraces = object->errorInfo->stackTracesInfo[0].stackTraces;
    stackTracesInfo->message = object->errorInfo->stackTracesInfo->message;
    *objectId = object->objectId;
}

void CFunctionExecution(void *ptr)
{
    auto f = static_cast<std::shared_ptr<std::function<void(void)>> *>(ptr);
    if (f && *f && **f) {
        (**f)();
    }
}

CErrorInfo CCreateStateStore(CConnectArguments *arguments, CStateStorePtr *stateStorePtr)
{
    YR::Libruntime::DsConnectOptions opts;
    size_t hostLen = 0;
    if (arguments->host != nullptr) {
        hostLen = strlen(arguments->host);
    }
    CheckNullAndAssignValue(arguments->host, hostLen, opts.host);
    CheckNullAndAssignValue(arguments->token, arguments->tokenLen, opts.token);
    CheckNullAndAssignValue(arguments->clientPublicKey, arguments->clientPublicKeyLen, opts.clientPublicKey);
    CheckNullAndAssignValue(arguments->clientPrivateKey, arguments->clientPrivateKeyLen, opts.clientPrivateKey);
    CheckNullAndAssignValue(arguments->serverPublicKey, arguments->serverPublicKeyLen, opts.serverPublicKey);
    CheckNullAndAssignValue(arguments->accessKey, arguments->accessKeyLen, opts.accessKey);
    CheckNullAndAssignValue(arguments->secretKey, arguments->secretKeyLen, opts.secretKey);
    CheckNullAndAssignValue(arguments->authClientID, arguments->authClientIDLen, opts.oAuthClientId);
    CheckNullAndAssignValue(arguments->authClientSecret, arguments->authClientSecretLen, opts.oAuthClientSecret);
    CheckNullAndAssignValue(arguments->authUrl, arguments->authUrlLen, opts.oAuthUrl);
    CheckNullAndAssignValue(arguments->tenantID, arguments->tenantIDLen, opts.tenantId);
    opts.port = arguments->port;
    opts.connectTimeoutMs = arguments->timeoutMs;
    opts.enableCrossNodeConnection = arguments->enableCrossNodeConnection != 0;
    std::shared_ptr<YR::Libruntime::StateStore> sharedStore = nullptr;
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->CreateStateStore(opts, sharedStore);
    if (err.OK()) {
        std::unique_ptr<std::shared_ptr<YR::Libruntime::StateStore>> uniqueStorePtr(
            new std::shared_ptr<YR::Libruntime::StateStore>(std::move(sharedStore)));
        *stateStorePtr = reinterpret_cast<CStateStorePtr>(uniqueStorePtr.release());
    }
    return ErrorInfoToCError(err);
}

CErrorInfo CSetTraceId(const char *cTraceId, int cTraceIdLen)
{
    std::string traceId = "";
    CheckNullAndAssignValue(cTraceId, cTraceIdLen, traceId);
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->SetTraceId(traceId);
    return ErrorInfoToCError(err);
}

inline std::shared_ptr<YR::Libruntime::StateStore> *GetStateStoreClient(CStateStorePtr stateStorePtr)
{
    if (stateStorePtr == nullptr) {
        return nullptr;
    }
    auto stateStore = reinterpret_cast<std::shared_ptr<YR::Libruntime::StateStore> *>(stateStorePtr);
    if (stateStore == nullptr || *stateStore == nullptr) {
        return nullptr;
    }
    return stateStore;
}

CErrorInfo CGenerateKey(CStateStorePtr stateStorePtr, char **cKey, int *cKeyLen)
{
    auto stateStore = GetStateStoreClient(stateStorePtr);
    if (stateStore == nullptr) {
        return ErrorInfoToCError(ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                           YR::Libruntime::ModuleCode::RUNTIME, "the state store is empty"));
    }
    std::string key = "";
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->GenerateKeyByStateStore(*stateStore, key);
    *cKey = StringToCString(key);
    *cKeyLen = key.size();
    return ErrorInfoToCError(err);
}

void CDestroyStateStore(CStateStorePtr stateStorePtr)
{
    auto stateStore = GetStateStoreClient(stateStorePtr);
    if (stateStore != nullptr) {
        delete stateStore;
        stateStore = nullptr;
    }
}

CErrorInfo CSetByStateStore(CStateStorePtr stateStorePtr, char *key, CBuffer data, CSetParam param)
{
    auto stateStore = GetStateStoreClient(stateStorePtr);
    if (stateStore == nullptr) {
        return ErrorInfoToCError(ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                           YR::Libruntime::ModuleCode::RUNTIME, "the state store is empty"));
    }
    SetParam setParam = CSetParamToSetParam(param);
    auto nativeBuffer = std::make_shared<YR::Libruntime::ReadOnlyNativeBuffer>(data.buffer, data.size_buffer);
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->SetByStateStore(*stateStore, key, nativeBuffer, setParam);
    return ErrorInfoToCError(err);
}

CErrorInfo CSetValueByStateStore(CStateStorePtr stateStorePtr, CBuffer data, CSetParam param, char **cKey, int *cKeyLen)
{
    auto stateStore = GetStateStoreClient(stateStorePtr);
    if (stateStore == nullptr) {
        return ErrorInfoToCError(ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                           YR::Libruntime::ModuleCode::RUNTIME, "the state store is empty"));
    }
    SetParam setParam = CSetParamToSetParam(param);
    std::string key = "";
    auto nativeBuffer = std::make_shared<YR::Libruntime::ReadOnlyNativeBuffer>(data.buffer, data.size_buffer);
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->SetValueByStateStore(*stateStore, nativeBuffer, setParam, key);
    *cKey = StringToCString(key);
    *cKeyLen = key.size();
    return ErrorInfoToCError(err);
}

CErrorInfo CGetByStateStore(CStateStorePtr stateStorePtr, char *key, CBuffer *data, int timeoutMs)
{
    auto stateStore = GetStateStoreClient(stateStorePtr);
    if (stateStore == nullptr) {
        return ErrorInfoToCError(ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                           YR::Libruntime::ModuleCode::RUNTIME, "the state store is empty"));
    }
    auto [lrt, err0] = getLibRuntime();
    if (!err0.OK()) {
        return ErrorInfoToCError(err0);
    }
    auto [value, err] = lrt->GetByStateStore(*stateStore, key, timeoutMs);
    if (err.OK()) {
        return ErrorInfoToCError(ToCBuffer(value, data));
    }
    return ErrorInfoToCError(err);
}

CErrorInfo CGetArrayByStateStore(CStateStorePtr stateStorePtr, char **cKeys, int cKeysLen, CBuffer *data, int timeoutMs)
{
    auto stateStore = GetStateStoreClient(stateStorePtr);
    if (stateStore == nullptr) {
        return ErrorInfoToCError(ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                           YR::Libruntime::ModuleCode::RUNTIME, "the state store is empty"));
    }
    auto keys = CStringsToStrings(cKeys, cKeysLen);
    auto [lrt, err0] = getLibRuntime();
    if (!err0.OK()) {
        return ErrorInfoToCError(err0);
    }
    auto [values, err] = lrt->GetArrayByStateStore(*stateStore, keys, timeoutMs);
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    for (int i = 0; i < cKeysLen; i++) {
        if (values[i] == nullptr) {
            continue;
        }
        auto errInfo = ToCBuffer(values[i], &data[i]);
        if (!errInfo.OK()) {
            FreeBuffers(data, i);
            return ErrorInfoToCError(errInfo);
        }
    }
    return ErrorInfoToCError(err);
}

CErrorInfo CQuerySizeByStateStore(CStateStorePtr stateStorePtr, char **cKeys, int sizeCKeys, uint64_t *cSize)
{
    auto stateStore = GetStateStoreClient(stateStorePtr);
    if (stateStore == nullptr) {
        return ErrorInfoToCError(ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                           YR::Libruntime::ModuleCode::RUNTIME, "the state store is empty"));
    }
    std::vector<std::string> keys(sizeCKeys);
    for (size_t i = 0; i < keys.size(); i++) {
        keys[i] = cKeys[i];
    }
    std::vector<uint64_t> results{};
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->QuerySizeByStateStore(*stateStore, keys, results);
    for (size_t i = 0; i < results.size(); ++i) {
        cSize[i] = results[i];
    }
    return ErrorInfoToCError(err);
}

CErrorInfo CDelByStateStore(CStateStorePtr stateStorePtr, char *key)
{
    auto stateStore = GetStateStoreClient(stateStorePtr);
    if (stateStore == nullptr) {
        return ErrorInfoToCError(ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                           YR::Libruntime::ModuleCode::RUNTIME, "the state store is empty"));
    }
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return ErrorInfoToCError(err);
    }
    err = lrt->DelByStateStore(*stateStore, key);
    return ErrorInfoToCError(err);
}

CErrorInfo CDelArrayByStateStore(CStateStorePtr stateStorePtr, char **cKeys, int cKeysLen, char ***cFailedKeys,
                                 int *cFailedKeysLen)
{
    auto stateStore = GetStateStoreClient(stateStorePtr);
    if (stateStore == nullptr) {
        return ErrorInfoToCError(ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                           YR::Libruntime::ModuleCode::RUNTIME, "the state store is empty"));
    }
    auto keys = CStringsToStrings(cKeys, cKeysLen);
    auto [lrt, err0] = getLibRuntime();
    if (!err0.OK()) {
        return ErrorInfoToCError(err0);
    }
    auto [failedKeys, err] = lrt->DelArrayByStateStore(*stateStore, keys);
    StringsToCStrings(failedKeys, cFailedKeys, cFailedKeysLen);
    return ErrorInfoToCError(err);
}

CCredential CGetCredential()
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        auto credential = YR::Libruntime::Credential{};
        return CredentialToCCre(credential);
    }
    auto credential = lrt->GetCredential();
    return CredentialToCCre(credential);
}

int CIsHealth()
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return 0;
    }
    if (lrt->IsHealth()) {
        return 1;
    }
    return 0;
}

int CIsDsHealth()
{
    auto [lrt, err] = getLibRuntime();
    if (!err.OK()) {
        return 0;
    }
    if (lrt->IsDsHealth()) {
        return 1;
    }
    return 0;
}

#ifdef __cplusplus
}
#endif
