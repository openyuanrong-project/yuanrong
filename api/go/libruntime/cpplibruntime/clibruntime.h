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

#pragma once

#include "stdint.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct tagCInvokeLabels {
    char *key;
    char *value;
} CInvokeLabels;

typedef struct tagCCustomExtension {
    char *key;
    char *value;
} CCustomExtension;

typedef struct tagStackTrace {
    char *className;
    char *methodName;
    char *fileName;
    int64_t lineNumber;
    CCustomExtension *extensions;
    int size_extensions;
} CStackTrace;

typedef struct tagStackTracesInfo {
    int code;
    int mcode;
    char *message;
    CStackTrace *stackTraces;
    int size_stackTraces;
} CStackTracesInfo;

typedef struct tagCErrorInfo {
    int code;
    char *message;
    CStackTracesInfo *stackTracesInfo;
    int size_stackTracesInfo;
    int dsStatusCode;
} CErrorInfo;

typedef enum tagCArgType {
    VALUE = 0,
    OBJECT_REF,
} CArgType;

typedef enum tagCApiType {
    ACTOR = 0,
    FAAS = 1,
    POSIX = 2,
} CApiType;

typedef struct tagCArg {
    CArgType type;
    char *data;
    uint64_t size;
} CArg;

typedef struct tagCFunctionMeta {
    char *appName;
    char *moduleName;
    char *funcName;
    char *className;
    int languageType;
    char *codeId;
    char *signature;
    char *poolLabel;
    CApiType apiType;
    char *functionId;
    char hasName;
    char *name;
    char hasNs;
    char *ns;
} CFunctionMeta;

typedef struct tagCInstanceAllocation {
    char *functionId;
    char *funcSig;
    char *instanceId;
    char *leaseId;
    int tLeaseInterval;
} CInstanceAllocation;

typedef struct tagCInstanceSession {
    char hasSessionId;
    char *sessionId;
    int sessionTtl;
    int concurrency;
} CInstanceSession;

typedef int CInvokeType;

typedef struct tagCBuffer {
    void *buffer;
    int64_t size_buffer;
    void *selfSharedPtrBuffer;
} CBuffer;

typedef struct tagCDataObject {
    char *id;
    CBuffer buffer;
    char **nestedObjIds;
    int size_nestedObjIds;
    void *selfSharedPtr;
} CDataObject;

typedef struct tagCLibruntimeConfig {
    char *functionSystemAddress;
    char *dataSystemAddress;
    char *grpcAddress;
    char *iamAddress;
    char *jobId;
    char *runtimeId;
    char *instanceId;
    char *functionName;
    char *logLevel;
    char *logDir;
    CApiType apiType;
    char inCluster;
    char isDriver;
    char enableMTLS;
    char *privateKeyPath;
    char *certificateFilePath;
    char *verifyFilePath;
    char *privateKeyPaaswd;
    char *functionId;
    char *systemAuthAccessKey;
    char *systemAuthSecretKey;
    int systemAuthSecretKeySize;
    char *systemAuthDataKey;
    int systemAuthDataKeySize;
    char *encryptPrivateKeyPasswd;
    char *primaryKeyStoreFile;
    char *standbyKeyStoreFile;
    char enableDsEncrypt;
    char *runtimePublicKeyContextPath;
    char *runtimePrivateKeyContextPath;
    char *dsPublicKeyContextPath;
    int maxConcurrencyCreateNum;
    char enableSigaction;
    char enableEvent;
} CLibruntimeConfig;

typedef struct tagCInvokeArg {
    // go memory, no need to free
    void *buf;
    int64_t size_buf;
    char isRef;
    char *objId;
    char *tenantId;
    char **nestedObjects;
    int size_nestedObjects;
} CInvokeArg;

typedef struct tagCCustomResource {
    char *name;
    float scalar;
} CCustomResource;

typedef struct tagCCreateOpt {
    char *key;
    char *value;
} CCreateOpt;

typedef enum tagCAffinityKind {
    RESOURCE = 0,
    INSTANCE = 1,
} CAffinityKind;

typedef enum tagCAffinityType {
    PREFERRED = 0,
    PREFERRED_ANTI = 1,
    REQUIRED = 2,
    REQUIRED_ANTI = 3,
} CAffinityType;

typedef enum tagCLabelOpType {
    IN = 0,
    NOT_IN = 1,
    EXISTS = 2,
    NOT_EXISTS = 3,
} CLabelOpType;

typedef struct tagCLabelOperator {
    CLabelOpType opType;
    char *labelKey;
    char **labelValues;
    int size_labelValues;
} CLabelOperator;

typedef struct tagCAffinity {
    CAffinityKind affKind;
    CAffinityType affType;
    char preferredPrio;
    char preferredAntiOtherLabels;
    CLabelOperator *labelOps;
    int size_labelOps;
} CAffinity;

typedef struct TagCCredential {
    char *ak;
    char *sk;
    int sizeSk;
    char *dk;
    int sizeDk;
} CCredential;

typedef struct tagCInvokeOptions {
    int cpu;
    int memory;
    CCustomResource *customResources;
    int size_customResources;
    CCustomExtension *customExtensions;
    int size_customExtensions;
    CCreateOpt *createOpt;
    int size_createOpt;
    char **labels;
    int size_labels;
    CAffinity *schedAffinities;
    int RetryTimes;
    int RecoverRetryTimes;
    int size_schedAffinities;
    char **codePaths;
    int size_codePaths;
    char *schedulerFunctionId;
    char **schedulerInstanceIds;
    int size_schedulerInstanceIds;
    char *traceId;
    int timeout;
    int acquireTimeout;
    char trafficLimited;
    CInvokeLabels *invokeLabels;
    int size_invokeLabels;
    CInstanceSession *instanceSession;
    int64_t scheduleTimeoutMs;
    char forceInvoke;
} CInvokeOptions;

typedef struct tagCErrorObject {
    char *objectId;
    CErrorInfo *errorInfo;
} CErrorObject;

typedef struct tagCWaitResult {
    char **readyIds;
    int size_readyIds;
    char **unreadyIds;
    int size_unreadyIds;
    CErrorObject **errorIds;
    int size_errorIds;
} CWaitResult;

typedef enum tagCWriteMode {
    NONE_L2_CACHE = 0,
    WRITE_THROUGH_L2_CACHE = 1,  // sync write
    WRITE_BACK_L2_CACHE = 2,     // async write
    NONE_L2_CACHE_EVICT = 3,     // evictable objects
} CWriteMode;

typedef enum tagCExistenceOpt {
    NONE = 0,
    NX = 1,
} CExistenceOpt;

typedef enum tagCCacheType {
    MEMORY = 0,
    DISK = 1,
} CCacheType;

typedef struct tagCSetParam {
    CWriteMode writeMode;
    uint32_t ttlSecond;
    CExistenceOpt existence;
    CCacheType cacheType;
} CSetParam;

typedef struct tagCMSetParam {
    CWriteMode writeMode;
    uint32_t ttlSecond;
    CExistenceOpt existence;
    CCacheType cacheType;
} CMSetParam;

typedef enum tagCConsistencyType {
    PRAM = 0,
    CAUSAL = 1,
} CConsistencyType;

typedef struct tagCCreateParam {
    CWriteMode writeMode;
    CConsistencyType consistencyType;
    CCacheType cacheType;
} CCreateParam;

typedef struct tagCProducerConfig {
    int64_t delayFlushTime;
    int64_t pageSize;
    uint64_t maxStreamSize;
    char *traceId;
} CProducerConfig;

typedef enum tagCSubscriptionType {
    STREAM = 0,
    ROUND_ROBIN,
    KEY_PARTITIONS,
    UNKNOWN,
} CSubscriptionType;

typedef struct tagCSubscriptionConfig {
    char *subscriptionName;
    CSubscriptionType type;
    char *traceId;
} CSubscriptionConfig;

typedef void (*CGetAsyncCallback)(char *cObjectID, CErrorInfo *cErr, void *userData);

typedef void (*CGetEventCallback)(char *cObjectID, CErrorInfo *cErr, void *userData);

typedef struct tagCElement {
    uint8_t *ptr;
    uint64_t size;
    uint64_t id;
} CElement;

typedef struct tagConnectArguments {
    char *host;
    int port;
    int timeoutMs;
    char *token;
    int tokenLen;
    char *clientPublicKey;
    int clientPublicKeyLen;
    char *clientPrivateKey;
    int clientPrivateKeyLen;
    char *serverPublicKey;
    int serverPublicKeyLen;
    char *accessKey;
    int accessKeyLen;
    char *secretKey;
    int secretKeyLen;
    char *authClientID;
    int authClientIDLen;
    char *authClientSecret;
    int authClientSecretLen;
    char *authUrl;
    int authUrlLen;
    char *tenantID;
    int tenantIDLen;
    char enableCrossNodeConnection;
} CConnectArguments;

typedef void *Consumer_p;
typedef void *Producer_p;
typedef void *CStateStorePtr;

typedef enum tagCHealthCheckCode {
    HEALTHY = 0,
    HEALTH_CHECK_FAILED,
    SUB_HEALTH,
} CHealthCheckCode;

void CParseCErrorObjectPointer(CErrorObject *object, int *code, char **errMessage, char **objectId,
                               CStackTracesInfo *stackTracesInfo);

// function
CErrorInfo CCreateInstance(CFunctionMeta *funcMeta, CInvokeArg *invokeArgs, int size_invokeArgs, CInvokeOptions *opts,
                           char **instanceId);
CErrorInfo CInvokeByInstanceId(CFunctionMeta *funcMeta, char *instanceId, CInvokeArg *invokeArgs, int size_invokeArgs,
                               CInvokeOptions *opts, char **cReturnObjectId);
CErrorInfo CInvokeByFunctionName(CFunctionMeta *funcMeta, CInvokeArg *invokeArgs, int size_invokeArgs,
                                 CInvokeOptions *opts, char **cReturnObjectId);
CErrorInfo CAcquireInstance(char *stateId, CFunctionMeta *cFuncMeta, CInvokeOptions *cInvokeOpts,
                            CInstanceAllocation *cInsAlloc);

CErrorInfo CReleaseInstance(CInstanceAllocation *insAlloc, char *cStateID, char cAbnormal, CInvokeOptions *cInvokeOpts);
void CCreateInstanceRaw(CBuffer cReqRaw, char *cContext);
void CInvokeByInstanceIdRaw(CBuffer cReqRaw, char *cContext);
void CKillRaw(CBuffer cReqRaw, char *cContext);

extern void GoRawCallback(char *cKey, CErrorInfo cErr, CBuffer cResultRaw);

// object
CErrorInfo CSetTenantId(const char *cTenantId, int cTenantIdLen);
void CWait(char **objIds, int size_objIds, int waitNum, int timeoutSec, CWaitResult *result);
CErrorInfo CPutCommon(char *objectId, CBuffer data, char **nestedIds, int sizeNestedIds, char isPutRaw,
                      CCreateParam param);
CErrorInfo CGetMultiCommon(char **cObjIds, int size_cObjIds, int timeoutMs, char allowPartial, CBuffer *cData,
                           char isRaw);
CErrorInfo CGet(char *objId, int timeoutSec, CBuffer *data);
extern void GoGetAsyncCallback(char *cObjectID, CBuffer cBuf, CErrorInfo *cErr, void *userData);
extern void GoWaitAsyncCallback(char *cObjectID, CErrorInfo *cErr, void *userData);
extern void GoGetEventCallback(char *cObjectID, CBuffer cBuf, CErrorInfo *cErr, void *userData);
void CUpdateSchdulerInfo(char *scheduleName, char *schedulerId, char *option);
void CGetAsync(char *objectId, void *userData);
void CWaitAsync(char *objectId, void *userData);
void CGetEvent(char *objectId, void *userData);
CErrorInfo CIncreaseReferenceCommon(char **cObjIds, int size_cObjIds, char *cRemoteId, char ***cFailedIds,
                                    int *size_cFailedIds, char isRaw);
CErrorInfo CDecreaseReferenceCommon(char **cObjIds, int size_cObjIds, char *cRemoteId, char ***cFailedIds,
                                    int *size_cFailedIds, char isRaw);
CErrorInfo CReleaseGRefs(char *cRemoteId);
CErrorInfo CAllocReturnObject(CDataObject *object, int dataSize, char **nestedIds, int sizeNestedIds,
                              uint64_t *totalNativeBufferSize);
void CSetReturnObject(CDataObject *cObject, int dataSize);
void CCancel(char **objIds, int size_objIds, char force, char recursive);

// KV
CErrorInfo CKVWrite(char *key, CBuffer data, CSetParam param);
CErrorInfo CKVMSetTx(char **key, int sizeKeys, CBuffer *data, CMSetParam param);
CErrorInfo CKVRead(char *key, int timeoutMs, CBuffer *data);
CErrorInfo CKVMultiRead(char **keys, int size_keys, int timeoutMs, char allowPartial, CBuffer *data);
CErrorInfo CKVDel(char *key);
CErrorInfo CKMultiVDel(char **keys, int size_keys, char ***cFailedKeys, int *size_cFailedKeys);

// stream
CErrorInfo CCreateStreamProducer(char *streamName, CProducerConfig *config, Producer_p *producer);
CErrorInfo CCreateStreamConsumer(char *streamName, CSubscriptionConfig *config, Consumer_p *consumer);
CErrorInfo CDeleteStream(char *streamName);
CErrorInfo CQueryGlobalProducersNum(char *streamName, uint64_t *num);
CErrorInfo CQueryGlobalConsumersNum(char *streamName, uint64_t *num);
CErrorInfo CProducerSend(Producer_p producerPtr, uint8_t *ptr, uint64_t size, uint64_t id);
CErrorInfo CProducerSendWithTimeout(Producer_p producerPtr, uint8_t *ptr, uint64_t size, uint64_t id,
                                    int64_t timeoutMs);
CErrorInfo CProducerClose(Producer_p producerPtr);
CErrorInfo CConsumerReceive(Consumer_p consumerPtr, uint32_t timeoutMs, CElement **elements, uint64_t *count);
CErrorInfo CConsumerReceiveExpectNum(Consumer_p consumerPtr, uint32_t expectNum, uint32_t timeoutMs,
                                     CElement **elements, uint64_t *count);
CErrorInfo CConsumerAck(Consumer_p consumerPtr, uint64_t elementId);
CErrorInfo CConsumerClose(Consumer_p consumerPtr);

// management
void CExit(int code, char *message);
CErrorInfo CKill(char *instanceId, int sigNo, CBuffer cData);
void CFinalize(void);
CErrorInfo CInit(CLibruntimeConfig *config);
void CReceiveRequestLoop(void);
void CExecShutdownHandler(int sigNum);
char *CGetRealInstanceId(char *objectId, int timeout);
void CSaveRealInstanceId(char *objectId, char *instanceId);

// buffer
CErrorInfo CWriterLatch(CBuffer *cBuffer);
CErrorInfo CMemoryCopy(CBuffer *cBuffer, void *cSrc, uint64_t size_cSrc);
CErrorInfo CSeal(CBuffer *cBuffer);
CErrorInfo CWriterUnlatch(CBuffer *cBuffer);

// handlers
extern CErrorInfo *GoLoadFunctions(char **codePaths, int size_codePaths);
extern CErrorInfo *GoFunctionExecution(CFunctionMeta *, CInvokeType, CArg *, int, CDataObject *, int);
extern CErrorInfo *GoCheckpoint(char *checkpointId, CBuffer *buffer);
extern CErrorInfo *GoRecover(CBuffer *buffer);
extern CErrorInfo *GoShutdown(uint64_t gracePeriodSeconds);
extern CErrorInfo *GoSignal(int sigNo, CBuffer *payload);
extern CHealthCheckCode GoHealthCheck(void);
extern char GoHasHealthCheck(void);

// pool
extern void GoFunctionExecutionPoolSubmit(void *ptr);
void CFunctionExecution(void *ptr);

// kv client
extern CErrorInfo CCreateStateStore(CConnectArguments *arguments, CStateStorePtr *stateStorePtr);
extern CErrorInfo CSetTraceId(const char *cTraceId, int traceId);
extern CErrorInfo CGenerateKey(CStateStorePtr stateStorePtr, char **cKey, int *cKeyLen);
extern void CDestroyStateStore(CStateStorePtr stateStorePtr);
extern CErrorInfo CSetByStateStore(CStateStorePtr stateStorePtr, char *key, CBuffer data, CSetParam param);
extern CErrorInfo CSetValueByStateStore(CStateStorePtr stateStorePtr, CBuffer data, CSetParam param, char **cKey,
                                        int *cKeyLen);
extern CErrorInfo CGetByStateStore(CStateStorePtr stateStorePtr, char *key, CBuffer *data, int timeoutMs);
extern CErrorInfo CGetArrayByStateStore(CStateStorePtr stateStorePtr, char **cKeys, int cKeysLen, CBuffer *data,
                                        int timeoutMs);
extern CErrorInfo CQuerySizeByStateStore(CStateStorePtr stateStorePtr, char **cKeys, int sizeCKeys, uint64_t *cSize);
extern CErrorInfo CDelByStateStore(CStateStorePtr stateStorePtr, char *key);
extern CErrorInfo CDelArrayByStateStore(CStateStorePtr stateStorePtr, char **cKeys, int cKeysLen, char ***cFailedKeys,
                                        int *cFailedKeysLen);
extern CCredential CGetCredential();
extern int CIsHealth();
extern int CIsDsHealth();
#ifdef __cplusplus
}
#endif
