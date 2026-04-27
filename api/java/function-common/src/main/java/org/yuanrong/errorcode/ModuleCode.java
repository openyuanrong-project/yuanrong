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

package org.yuanrong.errorcode;

/**
 * The enum Module code.
 *
 * @since 2023/11/6
 */
public enum ModuleCode {
    /**
     * Core module code.
     */
    CORE(10),
    /**
     * Runtime module code.
     */
    RUNTIME(20),
    /**
     * Runtime create module code.
     */
    RUNTIME_CREATE(21),
    /**
     * Runtime invoke module code.
     */
    RUNTIME_INVOKE(22),
    /**
     * Runtime kill module code.
     */
    RUNTIME_KILL(23),
    /**
     * Datasystem module code.
     */
    DATASYSTEM(30);

    private int code;

    private ModuleCode(int code) {
        this.code = code;
    }

    /**
     * Gets code.
     *
     * @return the code
     */
    public int getCode() {
        return this.code;
    }

    @Override
    public String toString() {
        return String.valueOf(this.code);
    }
}