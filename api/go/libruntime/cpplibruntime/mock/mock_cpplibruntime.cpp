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

#include "../clibruntime.h"
#include "src/libruntime/libruntime_manager.h"
#include "src/libruntime/utils/utils.h"
using namespace std;
using namespace std::placeholders;
using YR::Libruntime::ErrorCode;
using YR::Libruntime::ErrorInfo;
using YR::Libruntime::StreamConsumer;
using YR::Libruntime::StreamProducer;
#ifdef __cplusplus
extern "C" {
#endif

char *CString(const std::string &str)
{
    char *cStr = (char *)malloc(str.size() + 1);
    memcpy_s(cStr, str.size(), str.data(), str.size());
    cStr[str.size()] = 0;
    return cStr;
}

CErrorInfo *GoLoadFunctions(char **codePaths, int size_codePaths)
{
    return nullptr;
}
CErrorInfo *GoFunctionExecution(CFunctionMeta *, CInvokeType, CArg *, int, CDataObject *, int)
{
    return nullptr;
}
CErrorInfo *GoCheckpoint(char *checkpointId, CBuffer *buffer)
{
    return nullptr;
}
CErrorInfo *GoRecover(CBuffer *buffer)
{
    return nullptr;
}
CErrorInfo *GoShutdown(uint64_t gracePeriodSeconds)
{
    return nullptr;
}
CErrorInfo *GoSignal(int sigNo, CBuffer *payload)
{
    return nullptr;
}
CHealthCheckCode GoHealthCheck(void)
{
    return CHealthCheckCode::HEALTHY;
}
char GoHasHealthCheck(void)
{
    return 0;
}

void GoFunctionExecutionPoolSubmit(void *ptr) {}

void GoRawCallback(char *cKey, CErrorInfo cErr, CBuffer cResultRaw) {}

void GoGetAsyncCallback(char *cObjectID, CBuffer cBuf, CErrorInfo *cErr, void *userData) {}

void GoWaitAsyncCallback(char *cObjectID, CErrorInfo *cErr, void *userData) {}

void GoGetEventCallback(char *cObjectID, CBuffer cBuf, CErrorInfo *cErr, void *userData) {}

CErrorInfo ErrorInfoToCError(const ErrorInfo &err)
{
    CErrorInfo cErr{};
    cErr.message = nullptr;
    if (!err.Msg().empty()) {
        cErr.message = CString(err.Msg());
    }
    cErr.code = static_cast<int>(err.Code());
    cErr.dsStatusCode = err.GetDsStatusCode();
    return cErr;
}

CErrorInfo CInit(CLibruntimeConfig *config)
{
    return ErrorInfoToCError(ErrorInfo());
}

void CReceiveRequestLoop(void)
{
    return;
}

void CExecShutdownHandler(int sigNum)
{
    return;
}

CErrorInfo CCreateInstance(CFunctionMeta *cFuncMeta, CInvokeArg *cInvokeArgs, int size_invokeArgs,
                           CInvokeOptions *cInvokeOpts, char **instanceId)
{
    return ErrorInfoToCError(ErrorInfo());
}
CErrorInfo CInvokeByInstanceId(CFunctionMeta *cFuncMeta, char *cInstanceId, CInvokeArg *cInvokeArgs,
                               int size_invokeArgs, CInvokeOptions *cInvokeOpts, char **cReturnObjectId)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CAcquireInstance(char *stateId, CFunctionMeta *cFuncMeta, CInvokeOptions *cInvokeOpts,
                            CInstanceAllocation *cInsAlloc)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CReleaseInstance(CInstanceAllocation *insAlloc, char *cStateID, char cAbnormal, CInvokeOptions *cInvokeOpts)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CInvokeByFunctionName(CFunctionMeta *cFuncMeta, CInvokeArg *cInvokeArgs, int size_invokeArgs,
                                 CInvokeOptions *cInvokeOpts, char **cReturnObjectId)
{
    return ErrorInfoToCError(ErrorInfo());
}

void CCreateInstanceRaw(CBuffer cReqRaw, char *cContext)
{
    return;
}

void CInvokeByInstanceIdRaw(CBuffer cReqRaw, char *cContext)
{
    return;
}

void CKillRaw(CBuffer cReqRaw, char *cContext)
{
    return;
}

CErrorInfo CGet(char *objId, int timeoutSec, CBuffer *data)
{
    return ErrorInfoToCError(ErrorInfo());
}

void CUpdateSchdulerInfo(char *schedulerName, char *schedulerId, char *option)
{
    return;
}

void CGetAsync(char *objectId, void *userData)
{
    return;
}

void CWaitAsync(char *objectId, void *userData)
{
    return;
}

void CGetEvent(char *objectId, void *userData)
{
    return;
}

CErrorInfo CKill(char *instanceId, int sigNo, CBuffer cData)
{
    if (sigNo == 128) {
        return ErrorInfoToCError(ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                           YR::Libruntime::ModuleCode::RUNTIME, "failed to kill"));
    }
    return ErrorInfoToCError(ErrorInfo());
}

char *CGetRealInstanceId(char *objectId, int timeout)
{
    std::string str = "InstanceID";
    char *cStr = (char *)malloc(str.size() + 1);
    memcpy_s(cStr, str.size(), str.data(), str.size());
    cStr[str.size()] = 0;
    return cStr;
}

void CExit(int code, char *message)
{
    return;
}

void CFinalize(void)
{
    return;
}

CErrorInfo CPutCommon(char *objectId, CBuffer data, char **nestedIds, int size_nestedIds, char isPutRaw,
                      CCreateParam param)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CGetMultiCommon(char **cObjIds, int size_cObjIds, int timeoutMs, char allowPartial, CBuffer *cData,
                           char isRaw)
{
    return ErrorInfoToCError(ErrorInfo());
}

void CWait(char **objIds, int size_objIds, int waitNum, int timeoutSec, CWaitResult *result)
{
    return;
}

CErrorInfo CIncreaseReferenceCommon(char **cObjIds, int size_cObjIds, char *cRemoteId, char ***cFailedIds,
                                    int *size_cFailedIds, char isRaw)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CDecreaseReferenceCommon(char **cObjIds, int size_cObjIds, char *cRemoteId, char ***cFailedIds,
                                    int *size_cFailedIds, char isRaw)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CKVWrite(char *key, CBuffer data, CSetParam param)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CKVMSetTx(char **key, int sizeKeys, CBuffer *data, CMSetParam param)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CKVRead(char *key, int timeoutMs, CBuffer *cData)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CKVMultiRead(char **cKeys, int size_cKeys, int timeoutMs, char allowPartial, CBuffer *cData)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CKVDel(char *key)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CKMultiVDel(char **cKeys, int size_cKeys, char ***cFailedKeys, int *size_cFailedKeys)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CCreateStreamProducer(char *cStreamName, CProducerConfig *config, Producer_p *producer)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CCreateStreamConsumer(char *cStreamName, CSubscriptionConfig *config, Consumer_p *consumer)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CDeleteStream(char *cStreamName)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CQueryGlobalProducersNum(char *cStreamName, uint64_t *num)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CQueryGlobalConsumersNum(char *cStreamName, uint64_t *num)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CProducerSend(Producer_p producerPtr, uint8_t *ptr, uint64_t size, uint64_t id)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CProducerSendWithTimeout(Producer_p producerPtr, uint8_t *ptr, uint64_t size, uint64_t id, int64_t timeoutMs)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CProducerClose(Producer_p producerPtr)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CConsumerReceive(Consumer_p consumerPtr, uint32_t timeoutMs, CElement **elements, uint64_t *count)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CConsumerReceiveExpectNum(Consumer_p consumerPtr, uint32_t expectNum, uint32_t timeoutMs,
                                     CElement **elements, uint64_t *count)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CConsumerAck(Consumer_p consumerPtr, uint64_t elementId)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CConsumerClose(Consumer_p consumerPtr)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CAllocReturnObject(CDataObject *cObject, int dataSize, char **cNestedIds, int size_cNestedIds,
                              uint64_t *totalNativeBufferSize)
{
    return ErrorInfoToCError(ErrorInfo());
}

void CSetReturnObject(CDataObject *cObject, int dataSize)
{
    return;
}

CErrorInfo CWriterLatch(CBuffer *cBuffer)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CMemoryCopy(CBuffer *cBuffer, void *cSrc, uint64_t size_cSrc)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CSeal(CBuffer *cBuffer)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CWriterUnlatch(CBuffer *cBuffer)
{
    return ErrorInfoToCError(ErrorInfo());
}

void CParseCErrorObjectPointer(CErrorObject *object, int *code, char **errMessage, char **objectId,
                               CStackTracesInfo *stackTracesInfo)
{
    return;
}

void CFunctionExecution(void *ptr) {}

CErrorInfo CCreateStateStore(CConnectArguments *arguments, CStateStorePtr *stateStorePtr)
{
    *stateStorePtr = static_cast<CStateStorePtr>(malloc(sizeof(CStateStorePtr)));
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CSetTraceId(const char *cTraceId, int cTraceIdLen)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CSetTenantId(const char *cTenantId, int cTenantIdLen)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CGenerateKey(CStateStorePtr stateStorePtr, char **cKey, int *cKeyLen)
{
    return ErrorInfoToCError(ErrorInfo());
}

void CDestroyStateStore(CStateStorePtr stateStorePtr)
{
    free(stateStorePtr);
    return;
}

CErrorInfo CSetByStateStore(CStateStorePtr stateStorePtr, char *key, CBuffer data, CSetParam param)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CSetValueByStateStore(CStateStorePtr stateStorePtr, CBuffer data, CSetParam param, char **cKey, int *cKeyLen)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CGetByStateStore(CStateStorePtr stateStorePtr, char *key, CBuffer *data, int timeoutMs)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CGetArrayByStateStore(CStateStorePtr stateStorePtr, char **cKeys, int cKeysLen, CBuffer *data, int timeoutMs)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CQuerySizeByStateStore(CStateStorePtr stateStorePtr, char **cKeys, int sizeCKeys, uint64_t *cSize)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CDelByStateStore(CStateStorePtr stateStorePtr, char *key)
{
    return ErrorInfoToCError(ErrorInfo());
}

CErrorInfo CDelArrayByStateStore(CStateStorePtr stateStorePtr, char **cKeys, int cKeysLen, char ***cFailedKeys,
                                 int *cFailedKeysLen)
{
    return ErrorInfoToCError(ErrorInfo());
}

CCredential CGetCredential()
{
    return CCredential();
}

int CIsHealth()
{
    return 1;
}

int CIsDsHealth()
{
    return 1;
}

#ifdef __cplusplus
}
#endif
