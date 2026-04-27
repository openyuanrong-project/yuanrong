/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package org.yuanrong.runtime.client;

import org.yuanrong.ExistenceOpt;
import org.yuanrong.GetParams;
import org.yuanrong.MSetParam;
import org.yuanrong.SetParam;
import org.yuanrong.api.YR;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.serialization.Serializer;
import org.yuanrong.utils.SdkUtils;

import com.fasterxml.jackson.core.JsonProcessingException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

/**
 * This class represents a Key-Value Manager that allows to store and retrieve
 * data using a key.
 *
 * @since 2023 /11/14
 */
public class KVManager {
    private static Logger LOGGER = LoggerFactory.getLogger(KVManager.class);

    private static final int THREAD_POOL_SIZE = 1;

    private static final boolean DEFAULT_ALLOW_PARTIAL = false;

    private static final ExecutorService EXECUTOR_SERVICE = Executors.newFixedThreadPool(THREAD_POOL_SIZE);

    static {
        // shutdown thread pool when the program exits
        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            EXECUTOR_SERVICE.shutdownNow();
        }));
    }

    private static class SingletonKVManager {
        private static final KVManager KV_MANAGER = new KVManager();
    }

    /**
     * Returns the instance of the Singleton KVManager.
     *
     * @return The Singleton KVManager instance.
     */
    public static KVManager getSingleton() {
        return SingletonKVManager.KV_MANAGER;
    }

    /**
     * Sets the value of the specified key in the key-value store.
     *
     * @param key       The key to set the value for
     * @param value     The byte array value to be set for the key
     * @param existence The existence option to determine how to handle existing
     *                  values for the key
     * @throws YRException If the value is null or if an error occurs while
     *                            writing to the key-value store
     */
    public void set(String key, byte[] value, ExistenceOpt existence) throws YRException {
        SetParam setParam = new SetParam.Builder().existence(existence).build();
        YR.getRuntime().KVWrite(key, value, setParam);
    }

    /**
     * Sets the value of a key in the key-value store.
     *
     * @param key       The key to set the value for
     * @param value     The byte array value to be set for the key
     * @param length    The length of the value to set
     * @param existence The existence option to determine how to handle existing
     *                  values for the key
     * @throws YRException If the value is null or if an error occurs while
     *                            writing to the key-value store
     */
    public void set(String key, byte[] value, Integer length, ExistenceOpt existence) throws YRException {
        byte[] newValue;
        try {
            newValue = Arrays.copyOfRange(value, 0, length);
        } catch (ArrayIndexOutOfBoundsException e) {
            LOGGER.error("Length of value({}) is smaller than then parameter 'length'({}), key: {}", value.length,
                length, key);
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, e);
        } catch (NullPointerException e) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Cannot set a null value to key: " + key);
        }
        SetParam setParam = new SetParam.Builder().existence(existence).build();
        YR.getRuntime().KVWrite(key, newValue, setParam);
    }

    /**
     * Sets the value of the specified key in the key-value store.
     *
     * @param key       The key to set the value for
     * @param value     The byte array value to be set for the key
     * @param setParam  The setParam option for datasystem stateClient, contain writeMode,
     *                  ttlSecond, existence and cacheType.
     * @throws YRException If the value is null or if an error occurs while
     *                            writing to the key-value store
     */
    public void set(String key, byte[] value, SetParam setParam) throws YRException {
        if (setParam.getTtlSecond() < 0) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Cannot set a negative value to setParam's ttlSecond: " + setParam.getTtlSecond());
        }
        YR.getRuntime().KVWrite(key, value, setParam);
    }

    /**
     * Sets the value of a key in the key-value store.
     *
     * @param key       The key to set the value for
     * @param value     The byte array value to be set for the key
     * @param length    The length of the value to set
     * @param setParam  The setParam option for datasystem stateClient, contain writeMode,
     *                  ttlSecond, existence and cacheType.
     * @throws YRException If the value is null or if an error occurs while
     *                            writing to the key-value store
     */
    public void set(String key, byte[] value, Integer length, SetParam setParam) throws YRException {
        byte[] newValue;
        try {
            newValue = Arrays.copyOfRange(value, 0, length);
        } catch (ArrayIndexOutOfBoundsException e) {
            LOGGER.error("Length of value({}) is smaller than then parameter 'length'({}), key: {}", value.length,
                length, key);
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, e);
        } catch (NullPointerException e) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Cannot set a null value to key: " + key);
        }
        if (setParam.getTtlSecond() < 0) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Cannot set a negative value to setParam's ttlSecond: " + setParam.getTtlSecond());
        }
        YR.getRuntime().KVWrite(key, newValue, setParam);
    }

    /**
     * Sets the value of the specified key in the key-value store.
     *
     * @param key   The key to set the value for
     * @param value The byte array value to be set for the key
     * @return True if the value is successfully set, false otherwise
     */
    public boolean set(String key, byte[] value) {
        try {
            SetParam setParam = new SetParam.Builder().existence(ExistenceOpt.NONE).build();
            YR.getRuntime().KVWrite(key, value, setParam);
            return true;
        } catch (YRException e) {
            LOGGER.error("Failed to set value for the key: {}, error:", key, e);
            return false;
        }
    }

    /**
     * Sets the value of a key in the key-value store and invokes the callback upon
     * completion.
     *
     * @param key        The key to set the value for
     * @param value      The value to be set for the key
     * @param kvCallback The callback to invoke upon completion
     * @return A Future<Boolean> representing the result of the operation
     */
    public Future<Boolean> set(String key, byte[] value, KVCallback kvCallback) {
        return EXECUTOR_SERVICE.submit(
                () -> {
                    try {
                        SetParam setParam = new SetParam.Builder().existence(ExistenceOpt.NONE).build();
                        YR.getRuntime().KVWrite(key, value, setParam);
                        kvCallback.onComplete();
                        return true;
                    } catch (YRException e) {
                        LOGGER.error("Failed to set value for the key:{}, error:", key, e);
                        return false;
                    }
                });
    }

    /**
     * Deprecated. Writes the value of the specified key in the key-value store.
     *
     * @param key       The key to set the value for
     * @param value     The value to be set for the key
     * @param existence The existence option to determine how to handle existing
     *                  values for the key
     * @throws YRException If the value is null or if an error occurs while
     *                            writing to the key-value store
     */
    @Deprecated
    public void Write(String key, Object value, ExistenceOpt existence) throws YRException {
        write(key, value, existence);
    }

    /**
     * Writes the value of the specified key in the key-value store.
     *
     * @param key       The key to set the value for.
     * @param value     The value to be set for the key.
     * @param existence The existence option to determine how to handle existing values for the key.
     * @throws YRException If the value is null or if an error occurs while writing to the key-value store.
     */
    public void write(String key, Object value, ExistenceOpt existence) throws YRException {
        byte[] serializedValue;
        try {
            serializedValue = Serializer.serialize(value);
        } catch (JsonProcessingException e) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, e);
        }
        SetParam setParam = new SetParam.Builder().existence(existence).build();
        YR.getRuntime().KVWrite(key, serializedValue, setParam);
    }

    /**
     * Deprecated. Writes the value of the specified key in the key-value store.
     *
     * @param key       The key to set the value for
     * @param value     The value to be set for the key
     * @param setParam  The setParam option for datasystem stateClient, contain writeMode,
     *                  ttlSecond, existence and cacheType.
     * @throws YRException If the value is null or if an error occurs while
     *                            writing to the key-value store
     */
    @Deprecated
    public void Write(String key, Object value, SetParam setParam) throws YRException {
        write(key, value, setParam);
    }

    /**
     * Writes the value of the specified key in the key-value store.
     *
     * @param key       The key to set the value for
     * @param value     The value to be set for the key
     * @param setParam  The setParam option for datasystem stateClient, contain writeMode,
     *                  ttlSecond, existence and cacheType.
     * @throws YRException If the value is null or if an error occurs while
     *                            writing to the key-value store
     */
    public void write(String key, Object value, SetParam setParam) throws YRException {
        byte[] serializedValue;
        try {
            serializedValue = Serializer.serialize(value);
        } catch (JsonProcessingException e) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, e);
        }
        YR.getRuntime().KVWrite(key, serializedValue, setParam);
    }

    /**
     * Set Multiple key-value pairs.
     * Operations on all key-value pairs must either succeed or fail completely.
     *
     * @param keys      The keys to set
     * @param vals      vals The values to set. Size of values should equal to size of keys.
     * @param mSetParam Optional. MSetParams for datasystem stateClient.
     * @throws YRException If an error occurs while writing to the key-value store
     */
    public void mSetTx(List<String> keys, List<byte[]> vals, MSetParam mSetParam) throws YRException {
        if (keys.size() != vals.size()) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Arguments vector size not equal. keys size is: " + keys.size() + ", vals size is: "
                    + vals.size());
        }
        SdkUtils.checkMSetParam(mSetParam);
        YR.getRuntime().KVMSetTx(keys, vals, mSetParam);
    }

    /**
     * Set Multiple key-value pairs.
     * Operations on all key-value pairs must either succeed or fail completely.
     *
     * @param keys      The keys to set
     * @param vals      vals The values to set. Size of values should equal to size of keys.
     * @param lengths   The length of each values. Size of lens should equal to size of keys.
     * @param mSetParam Optional. MSetParams for datasystem stateClient.
     * @throws YRException If an error occurs while writing to the key-value store
     */
    public void mSetTx(List<String> keys, List<byte[]> vals, List<Integer> lengths,
        MSetParam mSetParam) throws YRException {
        if (keys.size() != vals.size() || keys.size() != lengths.size()) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Arguments vector size not equal. keys size is: " + keys.size() + ", vals size is: "
                    + vals.size() + ", lengths size is: " + lengths.size());
        }
        SdkUtils.checkMSetParam(mSetParam);
        List<byte[]> newValues = new ArrayList<>();
        for (int i = 0; i < keys.size(); i++) {
            try {
                newValues.add(Arrays.copyOfRange(vals.get(i), 0, lengths.get(i)));
            } catch (ArrayIndexOutOfBoundsException e) {
                LOGGER.error("Length of value({}) is smaller than the parameter 'length'({}), key: {}",
                    vals.get(i).length, lengths.get(i), keys.get(i));
                throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, e);
            } catch (NullPointerException e) {
                throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                        "Cannot set a null value to key: " + keys.get(i));
            }
        }
        YR.getRuntime().KVMSetTx(keys, newValues, mSetParam);
    }

    /**
     * Set Multiple key-value pairs. Values will be serialized by msgpack.
     * Operations on all key-value pairs must either succeed or fail completely.
     *
     * @param keys      The keys to be set for the values
     * @param values    The values to be set for the keys
     * @param mSetParam Optional. MSetParams for datasystem stateClient.
     * @throws YRException If an error occurs while writing to the key-value store
     */
    public void mWriteTx(List<String> keys, List<Object> values, MSetParam mSetParam) throws YRException {
        if (keys.size() != values.size()) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Arguments vector size not equal. keys size is: " + keys.size() + ", values size is: "
                    + values.size());
        }
        SdkUtils.checkMSetParam(mSetParam);
        List<byte[]> newValues = new ArrayList<>();
        try {
            for (int i = 0; i < keys.size(); i++) {
                newValues.add(Serializer.serialize(values.get(i)));
            }
        } catch (JsonProcessingException e) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, e);
        }
        YR.getRuntime().KVMSetTx(keys, newValues, mSetParam);
    }

    /**
     * Retrieves the value associated with the specified key from the key-value
     * store, using default timeout 300 seconds.
     *
     * @param key The key to retrieve the value for
     * @return The value associated with the specified key
     * @throws YRException If an error occurs while retrieving the value
     */
    public byte[] get(String key) throws YRException {
        return YR.getRuntime().KVRead(key, Constants.DEFAULT_GET_DATA_TIMEOUT_MS);
    }

    /**
     * Retrieves the value associated with the specified key from the key-value
     * store.
     *
     * @param key       The key to retrieve the value for
     * @param timeoutSec The maximum amount of time to wait for the value to be
     *                  retrieved, in seconds. Keep waiting if timeoutSec = -1
     * @return The value associated with the specified key
     * @throws YRException If an error occurs while retrieving the value
     */
    public byte[] get(String key, int timeoutSec) throws YRException {
        int timeoutMS = (timeoutSec == Constants.NO_TIMEOUT) ? timeoutSec : timeoutSec * Constants.SEC_TO_MS;
        return YR.getRuntime().KVRead(key, timeoutMS);
    }

    /**
     * Retrieves the values associated with the specified keys from the key-value
     * store.
     *
     * @param keys         The list of keys to retrieve the values for
     * @param timeoutSec   The maximum time to wait for the operation to complete,
     *                     in seconds
     * @param allowPartial Specifies whether to allow partial reads or not
     * @return A list of values associated with the specified keys
     * @throws YRException if an error occurs while retrieving the values
     */
    public List<byte[]> get(List<String> keys, int timeoutSec, boolean allowPartial) throws YRException {
        if (keys.isEmpty()) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Cannot set values for empty key list");
        }
        int timeoutMS = (timeoutSec == Constants.NO_TIMEOUT) ? timeoutSec : timeoutSec * Constants.SEC_TO_MS;
        return YR.getRuntime().KVRead(keys, timeoutMS, allowPartial);
    }

    /**
     * Retrieves the values associated with the given keys.
     *
     * @param keys         The keys to retrieve values for
     * @param allowPartial Specifies whether to allow partial reads or not
     * @return A list of byte arrays representing the values associated with the
     *         keys
     * @throws YRException if an error occurs while retrieving the values
     */
    public List<byte[]> get(List<String> keys, boolean allowPartial) throws YRException {
        return get(keys, Constants.DEFAULT_GET_DATA_TIMEOUT_SEC, allowPartial);
    }

    /**
     * Retrieves the values associated with the given keys.
     *
     * @param keys The keys to retrieve values for
     * @return A list of byte arrays representing the values associated with the
     *         keys
     * @throws YRException if an error occurs while retrieving the values
     */
    public List<byte[]> get(List<String> keys) throws YRException {
        return get(keys, Constants.DEFAULT_GET_DATA_TIMEOUT_SEC, DEFAULT_ALLOW_PARTIAL);
    }

    /**
     * Retrieves the values associated with the given keys and get params.
     *
     * @param keys The keys to retrieve values for
     * @param params The get params for datasystem
     * @return A list of byte arrays representing the values associated with the
     *         keys
     * @throws YRException if an error occurs while retrieving the values
     */
    public List<byte[]> getWithParam(List<String> keys, GetParams params) throws YRException {
        return getWithParam(keys, params, Constants.DEFAULT_GET_DATA_TIMEOUT_SEC);
    }

    /**
     * Retrieves the values associated with the given keys and get params.
     *
     * @param keys The keys to retrieve values for
     * @param params The get params for datasystem
     * @param timeoutSec   The maximum time to wait for the operation to complete,
     *                     in seconds
     * @return A list of byte arrays representing the values associated with the
     *         keys
     * @throws YRException if an error occurs while retrieving the values
     */
    public List<byte[]> getWithParam(List<String> keys, GetParams params, int timeoutSec)
            throws YRException {
        if (timeoutSec < Constants.NO_TIMEOUT) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Parameter 'timeoutSec' should be greater than or equal to -1 (no timeout).");
        }
        if (params.getGetParams().isEmpty()) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Get params does not accept empty list.");
        }
        if (params.getGetParams().size() != keys.size()) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME,
                    "Get params size is not equal to keys size.");
        }
        int timeoutMS = (timeoutSec == Constants.NO_TIMEOUT) ? timeoutSec : timeoutSec * Constants.SEC_TO_MS;
        return YR.getRuntime().KVGetWithParam(keys, params, timeoutMS);
    }

    /**
     * Deprecated. Read return Object.
     *
     * @param key the key
     * @param timeoutSec the timeoutSec
     * @param type the type
     * @return Object
     * @throws YRException the YR exception
     */
    public Object Read(String key, int timeoutSec, Class<?> type) throws YRException {
        return read(key, timeoutSec, type);
    }

    /**
     * Read return Object.
     *
     * @param key the key
     * @param timeoutSec the timeoutSec
     * @param type the type
     * @return Object
     * @throws YRException the YR exception
     */
    public Object read(String key, int timeoutSec, Class<?> type) throws YRException {
        int timeoutMS = (timeoutSec == Constants.NO_TIMEOUT) ? timeoutSec : timeoutSec * Constants.SEC_TO_MS;
        byte[] serializedValue = YR.getRuntime().KVRead(key, timeoutMS);
        try {
            return Serializer.deserialize(serializedValue, type);
        } catch (IOException e) {
            throw new YRException(ErrorCode.ERR_DESERIALIZATION_FAILED, ModuleCode.RUNTIME,
                "Failed to deserialize the value associated with the key: " + key);
        }
    }

    /**
     * Deprecated. Read return list.
     *
     * @param keys the keys
     * @param timeoutSec the timeoutSec
     * @param types the types
     * @param allowPartial if allowPartial
     * @return list
     * @throws YRException the YR exception
     */
    @Deprecated
    public List<Object> Read(List<String> keys, int timeoutSec, List<Class<?>> types, boolean allowPartial)
            throws YRException {
        return read(keys, timeoutSec, types, allowPartial);
    }

    /**
     * Read return list.
     *
     * @param keys the keys
     * @param timeoutSec the timeoutSec
     * @param types the types
     * @param allowPartial if allowPartial
     * @return list
     * @throws YRException the YR exception
     */
    public List<Object> read(List<String> keys, int timeoutSec, List<Class<?>> types, boolean allowPartial)
        throws YRException {
        int timeoutMS = (timeoutSec == Constants.NO_TIMEOUT) ? timeoutSec : timeoutSec * Constants.SEC_TO_MS;
        List<byte[]> serializedValue = YR.getRuntime().KVRead(keys, timeoutMS, allowPartial);
        List<Object> objects = new ArrayList<Object>(keys.size());
        for (int i = 0; i < keys.size(); i++) {
            try {
                objects.add(Serializer.deserialize(serializedValue.get(i), types.get(i)));
            } catch (IOException e) {
                throw new YRException(ErrorCode.ERR_DESERIALIZATION_FAILED, ModuleCode.RUNTIME,
                        "Failed to deserialize the value associated with the key: " + keys.get(i));
            }
        }
        return objects;
    }

    /**
     * Deletes the value associated with the given key.
     *
     * @param key The key to delete the value for
     * @throws YRException If an error occurs while deleting the value
     */
    public void del(String key) throws YRException {
        YR.getRuntime().KVDel(key);
    }

    /**
     * Deletes the values associated with the given keys.
     *
     * @param keys The list of keys to delete the value for
     * @return A list of keys for values that failed to be deleted
     * @throws YRException If an error occurs while deleting the keys
     */
    public List<String> del(List<String> keys) throws YRException {
        return YR.getRuntime().KVDel(keys);
    }

    /**
     * Deletes the value associated with the given key, and invokes the provided
     * callback upon completion.
     *
     * @param key      The key to delete the value for
     * @param callback The callback to call upon completion
     * @return A Future object contains the key for value that failed to be deleted
     *         if there is any
     */
    public Future<String> del(String key, KVCallback callback) {
        return EXECUTOR_SERVICE.submit(
                () -> {
                    try {
                        YR.getRuntime().KVDel(key);
                        callback.onComplete();
                        return Constants.EMPTY_STRING;
                    } catch (YRException e) {
                        LOGGER.error("Failed to delete value for the key: {}, error:", key, e);
                        return key;
                    }
                });
    }

    /**
     * Deletes the values associated with the given keys, and calls the provided
     * callback upon completion.
     *
     * @param keys     The list of keys to delete the value for
     * @param callback The callback to call upon completion
     * @return A Future object contains the keys for values that failed to be
     *         deleted if there is any
     * @throws YRException If an error occurs while deleting the keys
     */
    public Future<List<String>> del(List<String> keys, KVCallback callback) throws YRException {
        return EXECUTOR_SERVICE.submit(
                () -> {
                    List<String> failedKeys;
                    try {
                        failedKeys = YR.getRuntime().KVDel(keys);
                    } catch (YRException e) {
                        LOGGER.error("Failed to delete all values for keys: {}", keys, e);
                        throw e;
                    }
                    if (!failedKeys.isEmpty()) {
                        LOGGER.warn("Failed to delete values for keys: {}", failedKeys);
                    }
                    if (failedKeys.size() < keys.size()) {
                        callback.onComplete();
                    }
                    return failedKeys;
                });
    }
}
