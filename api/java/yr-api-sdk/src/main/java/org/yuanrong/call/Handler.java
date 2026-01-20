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

package org.yuanrong.call;

import org.yuanrong.exception.YRException;
import org.yuanrong.function.YRFunc;
import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.libruntime.generated.Libruntime.LanguageType;
import org.yuanrong.utils.LambdaUtils;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.lang.invoke.SerializedLambda;
import java.util.Optional;

/**
 * The Handler class provides a static method to get the function metadata of a YRFunc object.
 *
 * @since 2023 /11/01
 */
public class Handler {
    private static final Logger LOGGER = LoggerFactory.getLogger(Handler.class);

    /**
     * Returns the FunctionMeta object of the given YRFunc object.
     *
     * @param func the YRFunc object to get the metadata from
     * @param apiType the API type of the function
     * @return the FunctionMeta object representing the metadata of the given YRFunc object
     * @throws YRException if there is an error getting the serialized lambda
     */
    public static FunctionMeta getFunctionMeta(YRFunc func, ApiType apiType) throws YRException {
        Optional<SerializedLambda> serializedLambdaOptional = LambdaUtils.getSerializedLambda(func);
        if (!serializedLambdaOptional.isPresent()) {
            LOGGER.warn("Failed to get SerializedLambda of function: {}", func);
            return FunctionMeta.newBuilder().build();
        }
        SerializedLambda serializedLambda = serializedLambdaOptional.get();
        final String className = serializedLambda.getImplClass().replace('/', '.');
        final String methodName = serializedLambda.getImplMethodName();
        final String signature = serializedLambda.getImplMethodSignature();
        return FunctionMeta.newBuilder().setClassName(className).setSignature(signature).setFunctionName(methodName)
                .setApiType(apiType).setLanguage(LanguageType.Java).build();
    }

    /**
     * Returns the FunctionMeta object of the given YRFunc object.
     *
     * @param func      the YRFunc object to get the metadata from
     * @param apiType   the API type of the function
     * @param namespace the user designated namespace of the function
     * @param name      the user designated name of the function
     * @return the FunctionMeta object representing the metadata of the given YRFunc
     *         object
     * @throws YRException if there is an error getting the serialized lambda
     */
    public static FunctionMeta getFunctionMeta(YRFunc func, ApiType apiType, String namespace, String name)
            throws YRException {
        FunctionMeta functionMeta = getFunctionMeta(func, apiType);
        String className = functionMeta.getClassName();
        String signature = functionMeta.getSignature();
        String methodName = functionMeta.getFunctionName();
        return FunctionMeta.newBuilder().setClassName(className).setSignature(signature).setFunctionName(methodName)
                .setApiType(apiType).setLanguage(LanguageType.Java).setName(name).setNs(namespace).build();
    }
}