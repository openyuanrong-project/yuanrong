/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2026. All rights reserved.
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

package org.yuanrong.services.runtime.action;

import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.exception.LibRuntimeException;
import org.yuanrong.exception.YRException;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.runtime.util.Utils;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonObject;

import lombok.Setter;

import java.nio.charset.StandardCharsets;

/**
 * The stream handler of context, used to do some stream operation in FaaS mode.
 *
 * @since 2025-09-11
 */
public class Stream {
    private static final Gson GSON = new GsonBuilder().serializeNulls().create();

    @Setter
    private int maxWriteSize = 4 * 1024 * 1024;

    @Setter
    private String requestId = "";

    @Setter
    private String instanceId = "";

    /**
     * write jsonObject.
     *
     * @param jsonObject jsonObject
     * @throws YRException exception
     */
    public void write(JsonObject jsonObject) throws YRException {
        String jsonStr = GSON.toJson(jsonObject);
        int writeSizeInBytes = 0;
        if (jsonStr != null) {
            writeSizeInBytes = jsonStr.getBytes(StandardCharsets.UTF_8).length;
        }
        if (writeSizeInBytes > maxWriteSize) {
            throw new IllegalArgumentException("Write data size is larger than " + maxWriteSize);
        }
        ErrorInfo errorInfo;
        try {
            errorInfo = LibRuntime.streamWrite(jsonStr, requestId, instanceId);
        } catch (LibRuntimeException e) {
            throw new YRException(e.getErrorCode(), e.getModuleCode(), e.getMessage());
        }
        Utils.checkErrorAndThrow(errorInfo, "Write data failed");
    }
}
