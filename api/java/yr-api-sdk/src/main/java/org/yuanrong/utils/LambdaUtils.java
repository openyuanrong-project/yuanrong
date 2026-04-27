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

package org.yuanrong.utils;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.exception.YRException;

import java.io.Serializable;
import java.lang.invoke.SerializedLambda;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.util.Optional;

/**
 * The type Lambda utils.
 *
 * @since 2022 /08/30
 */
public final class LambdaUtils {
    /**
     * Get the SerializedLambda representation of a Serializable lambda function.
     *
     * @param lambda the Serializable lambda function.
     * @return an Optional containing the SerializedLambda if successful, or an empty Optional if not.
     * @throws YRException if there is an error during the serialization process.
     */
    public static Optional<SerializedLambda> getSerializedLambda(Serializable lambda) throws YRException {
        try {
            Method method = lambda.getClass().getDeclaredMethod("writeReplace");
            method.setAccessible(true);
            Object obj = method.invoke(lambda);
            if (obj instanceof SerializedLambda) {
                return Optional.of((SerializedLambda) obj);
            }
            return Optional.empty();
        } catch (NoSuchMethodException | InvocationTargetException | IllegalAccessException e) {
            throw new YRException(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME_CREATE, e);
        }
    }
}
