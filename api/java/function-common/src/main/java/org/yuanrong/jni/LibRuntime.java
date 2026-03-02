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

package org.yuanrong.jni;

import org.yuanrong.CreateParam;
import org.yuanrong.GetParams;
import org.yuanrong.GroupOptions;
import org.yuanrong.InvokeOptions;
import org.yuanrong.MSetParam;
import org.yuanrong.SetParam;
import org.yuanrong.api.InvokeArg;
import org.yuanrong.api.Node;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.libruntime.generated.Socket.FunctionLog;
import org.yuanrong.runtime.config.RuntimeContext;
import org.yuanrong.storage.InternalWaitResult;
import org.yuanrong.stream.ProducerConfig;
import org.yuanrong.stream.SubscriptionType;

import java.util.List;

/**
 * NativeLibRuntime represents methods of libRuntime
 * Generating command:
 * > cd MultiLanguageRuntime/java
 * > javah -classpath yr-jni/src/main/java/:function-common/src/main/java -d function-common/src/main/cpp/
 * -jni org.yuanrong.jni.LibRuntime
 *
 * @since 2023/10/23
 */
public class LibRuntime {
    static {
        LoadUtil.loadLibrary();
    }

    /**
     * get RuntimeContext
     *
     * @return RuntimeContext
     */
    public static String GetRuntimeContext() {
        return RuntimeContext.RUNTIME_CONTEXT.get();
    }

    /**
     * Init libruntime
     *
     * @param config the config
     * @return the ErrorInfo
     */
    public static native ErrorInfo Init(LibRuntimeConfig config);

    /**
     * Create Instance
     *
     * @param functionMeta functionMeta
     * @param args         create arguments
     * @param opt          invoke options
     * @return Pair of ErrorInfo and String
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native Pair<ErrorInfo, String> CreateInstance(FunctionMeta functionMeta, List<InvokeArg> args,
            InvokeOptions opt) throws LibRuntimeException;

    /**
     * Invoke instance by id
     *
     * @param funcMeta      functionMeta
     * @param instanceId    instanceId
     * @param args          invoke arguments
     * @param opt           invoke options
     * @return Pair of ErrorInfo and String
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native Pair<ErrorInfo, String> InvokeInstance(FunctionMeta funcMeta, String instanceId,
            List<InvokeArg> args, InvokeOptions opt) throws LibRuntimeException;

    /**
     * Get invoke results with ObjectRefIds
     *
     * @param ids          objectRefIds
     * @param timeoutMs    timeout in seconds, will be transferred to milliseconds
     *                     in Get interface
     * @param allowPartial whether to allow partial results
     * @return Pair of ErrorInfo and List<byte[]>
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native Pair<ErrorInfo, List<byte[]>> Get(List<String> ids, int timeoutMs, boolean allowPartial)
        throws LibRuntimeException;

    /**
     * Native method for put data
     *
     * @param data the data
     * @param nestObjIds the nestObjIds
     * @return Pair of ErrorInfo and String
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native Pair<ErrorInfo, String> Put(byte[] data, List<String> nestObjIds) throws LibRuntimeException;

    /**
     * Native method for put data
     *
     * @param data the data
     * @param nestObjIds the nestObjIds
     * @param createParam create param of datasystem
     * @return Pair of ErrorInfo and String
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native Pair<ErrorInfo, String> PutWithParam(byte[] data, List<String> nestObjIds,
            CreateParam createParam) throws LibRuntimeException;

    /**
     * Native method for Wait
     *
     * @param ids the ids
     * @param waitNum the waitNum
     * @param timeoutSec the timeoutSec
     * @return InternalWaitResult
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native InternalWaitResult Wait(List<String> ids, int waitNum, int timeoutSec)
        throws LibRuntimeException;

    /**
     * Native method for ReceiveRequestLoop
     */
    public static native void ReceiveRequestLoop();

    /**
     * Native method for FinalizeWithCtx
     *
     * @param runtimeCtx the runtimeCtx
     */
    public static native void FinalizeWithCtx(String runtimeCtx);

    /**
     * Native method for Finalize
     */
    public static native void Finalize();

    /**
     * Native method for Exit
     */
    public static native void Exit();

    /**
     * Native method for IsInitialized
     *
     * @return IsInitialized
     */
    public static native boolean IsInitialized();

    /**
     * Native method for setRuntimeContext
     * @param jobID jobID
     */
    public static native void setRuntimeContext(String jobID);

    /**
     * Native method for Kill instance
     *
     * @param instanceId the instanceId
     * @return ErrorInfo
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native ErrorInfo Kill(String instanceId) throws LibRuntimeException;

    /**
     * Native method for autoInitYR
     *
     * @param info info
     * @return YRAutoInitInfo
     */
    public static native YRAutoInitInfo AutoInitYR(YRAutoInitInfo info);

    /**
     * Native method for retrieving the real instance ID of an object.
     *
     * @param objectID The ID of the object whose real instance ID is to be
     *                 retrieved
     * @return The real instance ID of the object as a String
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native String GetRealInstanceId(String objectID) throws LibRuntimeException;

    /**
     * Native method for saving the real instance ID of an object.
     *
     * @param objectID the object id
     * @param instanceID the instance id
     * @param opts the invoke options
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native void SaveRealInstanceId(String objectID, String instanceID, InvokeOptions opts)
        throws LibRuntimeException;

    /**
     * Native method for writing key-value pairs to a key-value store.
     *
     * @param key       The key to write.
     * @param value     The value to write.
     * @param setParam  The setParam option for datasystem stateClient, contain writeMode,
     *                  ttlSecond, existence and cacheType.
     * @return An ErrorInfo object containing information about any errors that
     *         occurred during the write operation.
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native ErrorInfo KVWrite(String key, byte[] value, SetParam setParam) throws LibRuntimeException;

    /**
     * Native method for writing key-value pairs to a key-value store.
     *
     * @param keys      The keys to write.
     * @param values    The values to write.
     * @param mSetParam The mSetParam option for datasystem stateClient, contain writeMode,
     *                  ttlSecond, existence and cacheType.
     * @return An ErrorInfo object containing information about any errors that
     *         occurred during the write operation.
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native ErrorInfo KVMSetTx(List<String> keys, List<byte[]> values, MSetParam mSetParam)
        throws LibRuntimeException;

    /**
     * Native method for reading a single key-value pair.
     *
     * @param key       The key to look up
     * @param timeoutMS The maximum amount of time in milliseconds to wait for the read operation
     * @return A Pair object containing the byte array value associated with the
     *         key, and an ErrorInfo object if there was an error.
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native Pair<byte[], ErrorInfo> KVRead(String key, int timeoutMS)
        throws LibRuntimeException;

    /**
     * Native method for reading multiple key-value pairs.
     *
     * @param keys         A list of keys to look up in the key-value store.
     * @param timeoutMS    The maximum amount of time to wait for the key-value store to
     *                     respond.
     * @param allowPartial A boolean flag indicating whether to return partial
     *                     results if some of the keys are not found.
     * @return A Pair object containing a list of byte array values associated with
     *         the keys, and an ErrorInfo object if there was an error.
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native Pair<List<byte[]>, ErrorInfo> KVRead(List<String> keys, int timeoutMS, boolean allowPartial)
        throws LibRuntimeException;

    /**
     * Native method for reading multiple key-value pairs with get params.
     *
     * @param keys         A list of keys to look up in the key-value store.
     * @param params       The get params for datasystem.
     * @param timeoutMS    The maximum amount of time to wait for the key-value store to
     *                     respond.
     * @return A Pair object containing a list of byte array values associated with
     *         the keys, and an ErrorInfo object if there was an error.
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native Pair<List<byte[]>, ErrorInfo> KVGetWithParam(List<String> keys,
            GetParams params, int timeoutMS) throws LibRuntimeException;

    /**
     * Native method for deleting the value associated with the given key from the
     * key-value store.
     *
     * @param key The key of the value to be deleted
     * @return An ErrorInfo object indicating the success or failure of the
     *         operation
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native ErrorInfo KVDel(String key) throws LibRuntimeException;

    /**
     * Native method for deleting the values associated with the given keys from the
     * key-value store.
     *
     * @param keys The keys of the values to be deleted
     * @return A Pair object containing a List of keys that were successfully
     *         deleted and an instance of ErrorInfo indicating the success or
     *         failure of the operation
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native Pair<List<String>, ErrorInfo> KVDel(List<String> keys) throws LibRuntimeException;

    /**
     * Native method for SaveState
     *
     * @param timeoutMs the timeoutMs
     * @return An ErrorInfo object indicating the success or failure of the operation
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native ErrorInfo SaveState(int timeoutMs) throws LibRuntimeException;

    /**
     * Native method for LoadState
     *
     * @param timeoutMs the timeoutMs
     * @return An ErrorInfo object indicating the success or failure of the operation
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native ErrorInfo LoadState(int timeoutMs) throws LibRuntimeException;

    /**
     * Native method for GroupCreate
     *
     * @param groupName the groupName
     * @param opts the opts
     * @return An ErrorInfo object indicating the success or failure of the operation
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native ErrorInfo GroupCreate(String groupName, GroupOptions opts) throws LibRuntimeException;

    /**
     * Native method for GroupTerminate
     *
     * @param groupName the groupName
     * @return An ErrorInfo object indicating the success or failure of the operation
     */
    public static native ErrorInfo GroupTerminate(String groupName);

    /**
     * Native method for GroupWait
     *
     * @param groupName the groupName
     * @return An ErrorInfo object indicating the success or failure of the operation
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native ErrorInfo GroupWait(String groupName) throws LibRuntimeException;

    /**
     * Native method for processLog
     *
     * @param functionLog the functionLog
     * @return An ErrorInfo object indicating the success or failure of the operation
     */
    public static native ErrorInfo processLog(FunctionLog functionLog);

    /**
     * Use jni to create producer.
     *
     * @param streamName               The name of the stream.
     * @param delayFlushTimeMs         The time used in automatic flush after send.
     * @param pageSizeByte             The size used in allocate page.
     * @param maxStreamSize            The max stream size in worker.
     * @param shouldCleanup            Should auto delete when last producer/consumer exit
     * @param encryptStream            The encrypt stream
     * @param retainForNumConsumers    The retain for num consumers
     * @param reserveSize              The reserve size
     * @return The Producer pointer.
     * @throws LibRuntimeException if there is an exception during creating stream producer
     */
    public static native long CreateStreamProducer(String streamName, long delayFlushTimeMs,
            long pageSizeByte, long maxStreamSize, boolean shouldCleanup, boolean encryptStream,
            long retainForNumConsumers, long reserveSize)
            throws LibRuntimeException;

    /**
     * Use jni to subscribe a new consumer onto master request
     *
     * @param streamName       The name of the stream.
     * @param subName          The name of subscription.
     * @param subscriptionType The type of SubscriptionType.
     * @param shouldAutoAck    Should AutoAck be enabled for this subscriber or not.
     * @return The Consumer pointer.
     * @throws LibRuntimeException if there is an exception during creating stream consumer
     */
    public static native long CreateStreamConsumer(String streamName, String subName,
            SubscriptionType subscriptionType, boolean shouldAutoAck) throws LibRuntimeException;

    /**
     * Use jni to delete the stream
     *
     * @param streamName The name of the target stream.
     * @return ErrorInfo
     */
    public static native ErrorInfo DeleteStream(String streamName);

    /**
     * Use jni to query numbers of producer in global worker node
     *
     * @param streamName The name of the target stream.
     * @return The query result.
     * @throws LibRuntimeException if there is an exception during quering global producersNum
     */
    public static native long QueryGlobalProducersNum(String streamName) throws LibRuntimeException;

    /**
     * Use jni to query numbers of consumer in global worker node
     *
     * @param streamName The name of the target stream.
     * @return The query result.
     * @throws LibRuntimeException if there is an exception during quering global consumersNum
     */
    public static native long QueryGlobalConsumersNum(String streamName) throws LibRuntimeException;

    /**
     * Native method for DecreaseReference
     *
     * @param ids the ids
     */
    public static native void DecreaseReference(List<String> ids);

    /**
     * Use jni to create producer.
     *
     * @param streamName               The name of the stream.
     * @param producerConfig           The producer config
     * @return The Producer pointer.
     * @throws LibRuntimeException if there is an exception during creating stream producer
     */
    public static native long CreateStreamProducerWithConfig(String streamName, ProducerConfig producerConfig)
            throws LibRuntimeException;

    /**
     * Native method for retrieving the instance route of an object.
     *
     * @param objectID The ID of the object whose instance route is to be retrieved
     * @return The instance route of the object as a String
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native String GetInstanceRoute(String objectID) throws LibRuntimeException;

    /**
     * Native method for saving the instance route of an object.
     *
     * @param objectID the object id
     * @param instanceRoute the instance route
     * @throws LibRuntimeException the LibRuntimeException
     */
    public static native void SaveInstanceRoute(String objectID, String instanceRoute) throws LibRuntimeException;

    /**
     * Native method for Kill instance sync
     *
     * @param instanceId the instanceId
     * @return ErrorInfo
     */
    public static native ErrorInfo KillSync(String instanceId);

    /**
     * Get node information in the cluster.
     *
     * @return Pair of ErrorInfo and list of node information.
     * @throws LibRuntimeException the LibRuntimeException.
     */
    public static native Pair<ErrorInfo, List<Node>> nodes() throws LibRuntimeException;

    /**
     * write stream data event.
     *
     * @param jsonData the json data of stream event.
     * @param requestId the request id.
     * @param instanceId the instance id.
     * @return ErrorInfo
     * @throws LibRuntimeException the LibRuntimeException.
     */
    public static native ErrorInfo streamWrite(String jsonData, String requestId, String instanceId)
            throws LibRuntimeException;

    /**
     * get requestId and instanceId.
     *
     * @return Pair of requestId and instanceId.
     */
    public static native Pair<String, String> getRequestAndInstanceID();
}
