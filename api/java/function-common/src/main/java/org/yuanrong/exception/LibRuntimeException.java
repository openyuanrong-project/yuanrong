/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

package org.yuanrong.exception;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;

import lombok.Data;
import lombok.EqualsAndHashCode;

/**
 * The type Lib runtime exception.
 *
 * @since 2023/11/6
 */
@Data
@EqualsAndHashCode(callSuper = false)
public class LibRuntimeException extends Exception {
    private String message;

    private ErrorCode errorCode;

    private ModuleCode moduleCode;

    /**
     * Instantiates a new Lib runtime exception.
     *
     * @param errorCode  the error code
     * @param moduleCode the module code
     * @param message    the message
     */
    public LibRuntimeException(ErrorCode errorCode, ModuleCode moduleCode, String message) {
        this.errorCode = errorCode;
        this.moduleCode = moduleCode;
        this.message = message;
    }

    /**
     * Instantiates a new Lib runtime exception.
     *
     * @param errorCode  the error code
     * @param moduleCode the module code
     * @param message    the message
     */
    public LibRuntimeException(String message) {
        this(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, message);
    }
}
