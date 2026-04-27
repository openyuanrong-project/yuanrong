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

package org.yuanrong.executor;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.handler.traceback.StackTraceInfo;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.List;

/**
 * The type Return type.
 *
 * @since 2023/11/6
 */
public class ReturnType {
    /**
     * The Code.
     */
    public ErrorCode code;

    /**
     * The Error info.
     */
    public ErrorInfo errorInfo;

    /**
     * The Buf.
     */
    public byte[] buf;

    /**
     * Instantiates a new Return type.
     *
     * @param code the code
     * @param buf the buf
     */
    ReturnType(ErrorCode code, ByteBuffer buf) {
        this.code = code;
        this.buf = buf.array();
        this.errorInfo = new ErrorInfo(code, ModuleCode.RUNTIME, "");
    }

    /**
     * Instantiates a new Return type.
     *
     * @param code the code
     * @param inputBuf the input buf
     */
    ReturnType(ErrorCode code, byte[] inputBuf) {
        this(code, ModuleCode.RUNTIME, "", inputBuf);
    }

    /**
     * Instantiates a new Return type.
     *
     * @param code the code
     * @param errorMsg the error msg
     */
    ReturnType(ErrorCode code, String errorMsg) {
        this(code, ModuleCode.RUNTIME, errorMsg, errorMsg.getBytes(StandardCharsets.UTF_8));
    }

    /**
     * Instantiates a new Return type.
     *
     * @param code the code
     */
    ReturnType(ErrorCode code) {
        this(code, ModuleCode.RUNTIME, "", ByteBuffer.allocate(0).array());
    }

    /**
     * Instantiates a new Return type.
     *
     * @param code the code
     * @param mCode the m code
     * @param msg the msg
     */
    ReturnType(ErrorCode code, ModuleCode mCode, String msg) {
        this(code, mCode, msg, msg.getBytes(StandardCharsets.UTF_8));
    }

    /**
     * Instantiates a new Return type.
     *
     * @param code the code
     * @param mCode the m code
     * @param msg the msg
     * @param stackTraceInfo the stackTraceInfo
     */
    ReturnType(ErrorCode code, ModuleCode mCode, String msg, List<StackTraceInfo> stackTraceInfo) {
        this.code = code;
        this.errorInfo = new ErrorInfo(code, mCode, msg, stackTraceInfo);
        this.buf = msg.getBytes(StandardCharsets.UTF_8);
    }

    /**
     * Instantiates a new Return type.
     *
     * @param code the code
     * @param mCode the m code
     * @param msg the msg
     * @param buf the buf
     */
    ReturnType(ErrorCode code, ModuleCode mCode, String msg, byte[] buf) {
        this.code = code;
        this.errorInfo = new ErrorInfo(code, mCode, msg);
        this.buf = buf;
    }

    /**
     * Gets code.
     *
     * @return the code
     */
    public ErrorCode getCode() {
        return this.code;
    }

    /**
     * Get bytes byte [ ].
     *
     * @return the byte [ ]
     */
    public byte[] getBytes() {
        return this.buf;
    }

    /**
     * Gets error info.
     *
     * @return the error info
     */
    public ErrorInfo getErrorInfo() {
        return this.errorInfo;
    }
}
