/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

package org.yuanrong;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;

import java.lang.reflect.Constructor;
import java.lang.reflect.Executable;
import java.lang.reflect.Method;
import java.util.Optional;

/**
 * The java function wrapper.
 *
 * @since 2023 /10/16
 */
public class FunctionWrapper {
    /**
     * The Executable.
     */
    public final Executable executable;

    /**
     * The Function descriptor.
     */
    public final FunctionMeta functionMeta;

    /**
     * Instantiates a new Yr function.
     *
     * @param executable the executable
     * @param functionMeta the function meta
     */
    public FunctionWrapper(Executable executable, FunctionMeta functionMeta) {
        this.executable = executable;
        this.functionMeta = functionMeta;
    }

    /**
     * Gets function descriptor.
     *
     * @return the function descriptor
     */
    public FunctionMeta getFunctionMeta() {
        return functionMeta;
    }

    /**
     * Gets method.
     *
     * @return the method
     * @throws YRException the YR exception
     */
    public Method getMethod() throws YRException {
        if (executable instanceof Method) {
            return (Method) executable;
        }
        throw new YRException(ErrorCode.ERR_INCORRECT_FUNCTION_USAGE, ModuleCode.RUNTIME_INVOKE,
            "get wrong Method");
    }

    /**
     * Gets return type.
     *
     * @return the return type
     * @throws YRException the YR exception
     */
    public Optional<Class<?>> getReturnType() throws YRException {
        if (executable instanceof Constructor || void.class.equals(getMethod().getReturnType())) {
            return Optional.empty();
        }
        return Optional.of((getMethod()).getReturnType());
    }
}
