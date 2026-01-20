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

package org.yuanrong.runtime.client;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.YRException;
import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.runtime.util.Utils;
import org.yuanrong.serialization.ObjectRefDeserializer;
import org.yuanrong.serialization.ObjectRefSerializer;
import org.yuanrong.serialization.strategy.Strategy;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;

import lombok.Data;

import java.util.Collections;
import java.util.List;

/**
 * ObjectRef class.
 *
 * @note 1.Yuanrong encourages users to store large objects in the data system using YR.put and obtain a unique
 *       ObjectRef (object reference). When invoking user functions, use the ObjectRef instead of the object itself as a
 *       function parameter to reduce the overhead of transmitting large objects between Yuanrong and user function
 *       components, ensuring efficient flow.\n
 *       2.The return value of each user function call will also be returned in the form of an ObjectRef, which the user
 *       can use as an input parameter for the next call or retrieve the corresponding object through YR.get.\n
 *       3.Currently, users cannot construct objectRef by themselves.
 *
 * @since 2023/11/6
 */
@Data
@JsonSerialize(using = ObjectRefSerializer.class)
@JsonDeserialize(using = ObjectRefDeserializer.class)
public class ObjectRef {
    /**
     * object type
     */
    private Class<?> type;

    /**
     * Whether the object is of type ByteBuffer.
     */
    private boolean isByteBuffer = false;

    /**
     * Object ID in the Yuanrong cluster.
     */
    private final String objectID;

    /**
     * Decrease reference flag of ObjectRef.
     */
    private boolean isReleased = false;

    /**
     * The constructor for ObjectRef.
     *
     * @param objectID object ID in the Yuanrong cluster.
     */
    public ObjectRef(String objectID) {
        this(objectID, null);
    }

    /**
     * The constructor for ObjectRef.
     *
     * @param objectID object ID in the Yuanrong cluster.
     * @param returnType object type.
     */
    public ObjectRef(String objectID, Class<?> returnType) {
        this.objectID = objectID;
        this.type = returnType;
    }

    /**
     * Get objectID.
     *
     * @return objectID: object ID in the Yuanrong cluster.
     */
    public String getObjId() {
        return this.objectID;
    }

    /**
     * Get results.
     *
     * @param classType Class Type.
     * @return The result of ObjectRef.
     * @throws YRException Unified exception types thrown.
     * @throws LibRuntimeException Unified exception types thrown.
     */
    public Object get(Class<?> classType) throws YRException, LibRuntimeException {
        LibRuntime.Wait(Collections.singletonList(this.objectID), 1, Constants.NO_TIMEOUT);
        Pair<ErrorInfo, List<byte[]>> getRes = LibRuntime.Get(Collections.singletonList(this.objectID),
            Constants.NO_TIMEOUT, false);
        Utils.checkErrorAndThrow(getRes.getFirst(), "get obj");
        List<byte[]> res = getRes.getSecond();
        List<Object> objects = Strategy.getObjects(res, Collections.singletonList(this));
        if (objects.isEmpty()) {
            throw new LibRuntimeException(ErrorCode.ERR_DATASYSTEM_FAILED, ModuleCode.DATASYSTEM, "get object failed");
        }
        return objects.get(0);
    }

    /**
     * Get results.
     *
     * @return The result of ObjectRef.
     * @throws YRException Unified exception types thrown.
     * @throws LibRuntimeException Unified exception types thrown.
     */
    public Object get() throws YRException, LibRuntimeException {
        LibRuntime.Wait(Collections.singletonList(this.objectID), 1, Constants.DEFAULT_GET_DATA_TIMEOUT_SEC);
        Pair<ErrorInfo, List<byte[]>> getRes = LibRuntime.Get(Collections.singletonList(this.objectID),
                Constants.DEFAULT_GET_DATA_TIMEOUT_SEC, false);
        Utils.checkErrorAndThrow(getRes.getFirst(), "get obj");
        List<byte[]> res = getRes.getSecond();
        List<Object> objects = Strategy.getObjects(res, Collections.singletonList(this));
        if (objects.isEmpty()) {
            throw new LibRuntimeException(ErrorCode.ERR_DATASYSTEM_FAILED, ModuleCode.DATASYSTEM, "get object failed");
        }
        return objects.get(0);
    }

    /**
     * Get results.
     *
     * @param timeout timeout duration.
     * @return The result of ObjectRef.
     * @throws YRException Unified exception types thrown.
     * @throws LibRuntimeException Unified exception types thrown.
     */
    public Object get(int timeout) throws YRException, LibRuntimeException {
        LibRuntime.Wait(Collections.singletonList(this.objectID), 1, timeout);
        Pair<ErrorInfo, List<byte[]>> getRes = LibRuntime.Get(Collections.singletonList(this.objectID), timeout,
                false);
        Utils.checkErrorAndThrow(getRes.getFirst(), "get obj");
        List<byte[]> res = getRes.getSecond();
        List<Object> objects = Strategy.getObjects(res, Collections.singletonList(this));
        if (objects.isEmpty()) {
            throw new LibRuntimeException(ErrorCode.ERR_DATASYSTEM_FAILED, ModuleCode.DATASYSTEM, "get object failed");
        }
        return objects.get(0);
    }

    @Override
    protected void finalize() throws Throwable {
        release();
    }

    /**
     * Release the ObjectRef, decrease reference.
     */
    public void release() {
        if (!isReleased && LibRuntime.IsInitialized()) {
            LibRuntime.DecreaseReference(Collections.singletonList(this.objectID));
            isReleased = true;
        }
    }
}
