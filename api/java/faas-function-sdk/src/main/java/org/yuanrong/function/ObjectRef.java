/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

package org.yuanrong.function;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.function.common.RspErrorCode;
import org.yuanrong.function.common.Util;
import org.yuanrong.function.runtime.exception.InvokeException;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.runtime.util.Constants;
import org.yuanrong.services.enums.FaasErrorCode;
import org.yuanrong.services.model.CallResponse;
import org.yuanrong.services.model.CallResponseJsonObject;
import org.yuanrong.services.model.Response;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.google.gson.JsonParseException;

import lombok.extern.slf4j.Slf4j;

import java.lang.reflect.Type;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Locale;

/**
 * ObjectRef class
 *
 * @since 2024-7-10
 */
@Slf4j
public class ObjectRef<T> {
    /**
     * GSON
     */
    protected Gson gson = new Gson();

    /**
     * flag represents whether the result has been got.
     */
    protected boolean hasFlag;

    /**
     * result is the result of the ObjectRef.
     */
    protected T result;

    private final String objectID;

    /**
     * Decrease reference flag of ObjectRef.
     */
    private boolean isReleased = false;

    protected ObjectRef(String objectID) {
        this.objectID = objectID;
        this.hasFlag = false;
        this.result = null;
    }

    private int checkAndGetTimeoutMs(int timeoutSec) throws InvokeException {
        if (timeoutSec < Constants.NO_TIMEOUT) {
            throw new InvokeException(RspErrorCode.INVALID_PARAMETER.getErrorCode(),
                "get config timeout (" + timeoutSec + ") is invalid");
        }
        int timeoutMs = Constants.NO_TIMEOUT;
        if (timeoutSec != Constants.NO_TIMEOUT) {
            timeoutMs = timeoutSec * Constants.SEC_TO_MS;
        }
        return timeoutMs;
    }

    /**
     * get result
     *
     * @param timeoutSec the timeoutSec
     * @return result of ObjectRef
     */
    public T get(int timeoutSec) {
        if (this.hasFlag) {
            return this.result;
        }
        int timeoutMs = checkAndGetTimeoutMs(timeoutSec);
        List<String> refIds = Arrays.asList(this.objectID);
        Pair<ErrorInfo, List<byte[]>> getResult;
        try {
            // This method may throw 'LibruntimeException' when timeout occurs.
            getResult = LibRuntime.Get(refIds, timeoutMs, true);
        } catch (LibRuntimeException e) {
            String message = String.format(Locale.ROOT, "failed to get result %s, error: %s", refIds, e.getMessage());
            throw new InvokeException(FaasErrorCode.FUNCTION_RUN_ERROR.getCode(), message);
        }
        Util.checkErrorAndThrow(getResult.getFirst(), "faas get");
        Response response;
        String responseStr = new String(getResult.getSecond().get(0), StandardCharsets.UTF_8);
        try {
            response = gson.fromJson(responseStr, CallResponseJsonObject.class);
        } catch (JsonParseException e) {
            try {
                response = gson.fromJson(responseStr, CallResponse.class);
            } catch (JsonParseException exception) {
                throw new InvokeException(FaasErrorCode.FUNCTION_RUN_ERROR.getCode(), exception.getMessage());
            }
        }
        String innerCode = response.getInnerCode();
        if (!ErrorCode.ERR_OK.toString().equals(innerCode)) {
            throw new InvokeException(Integer.parseInt(innerCode), String.valueOf(response.getBody()));
        }
        this.result = (T) response.getBody();
        this.hasFlag = true;
        return this.result;
    }

    /**
     * get result
     *
     * @return result of ObjectRef
     */
    public T get() {
        return this.get(Constants.NO_TIMEOUT);
    }

    /**
     * get result
     *
     * @param classType the classType
     * @param timeoutSec the timeoutSec
     * @return result of ObjectRef
     */
    public T get(Class<?> classType, int timeoutSec) {
        if (this.hasFlag) {
            return this.result;
        }
        int timeoutMs = checkAndGetTimeoutMs(timeoutSec);
        List<String> refIds = Arrays.asList(this.objectID);
        Pair<ErrorInfo, List<byte[]>> getRes;
        try {
            // This method may throw 'LibruntimeException' when timeout occurs.
            getRes = LibRuntime.Get(refIds, timeoutMs, true);
        } catch (LibRuntimeException e) {
            String message = String.format(Locale.ROOT, "failed to get result %s, error: %s", refIds, e.getMessage());
            throw new InvokeException(FaasErrorCode.FUNCTION_RUN_ERROR.getCode(), message);
        }
        Util.checkErrorAndThrow(getRes.getFirst(), "faas get");
        Response response;
        String responseString = new String(getRes.getSecond().get(0), StandardCharsets.UTF_8);
        try {
            if (JsonObject.class.equals(classType)) {
                response = gson.fromJson(responseString, CallResponseJsonObject.class);
            } else {
                response = gson.fromJson(responseString, CallResponse.class);
            }
        } catch (JsonParseException e) {
            try {
                response = gson.fromJson(responseString, CallResponse.class);
            } catch (JsonParseException exception) {
                throw new InvokeException(FaasErrorCode.FUNCTION_RUN_ERROR.getCode(), exception.getMessage());
            }
        }
        Object responseBody = response.getBody();
        String innerCode = response.getInnerCode();
        if (!ErrorCode.ERR_OK.toString().equals(innerCode)) {
            throw new InvokeException(Integer.parseInt(innerCode), String.valueOf(responseBody));
        }
        if (classType.isInstance(responseBody)) {
            this.result = (T) responseBody;
        } else {
            try {
                this.result = gson.fromJson(String.valueOf(responseBody), (Type) classType);
            } catch (JsonParseException e) {
                throw new InvokeException(RspErrorCode.INTERNAL_ERROR.getErrorCode(), e.getMessage());
            }
        }
        this.hasFlag = true;
        return this.result;
    }

    /**
     * get result
     *
     * @param classType the classType
     * @return result of ObjectRef
     */
    public T get(Class<?> classType) {
        return this.get(classType, Constants.NO_TIMEOUT);
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
