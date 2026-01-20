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

package org.yuanrong.instance;

import org.yuanrong.libruntime.generated.Libruntime.ApiType;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;
import org.yuanrong.libruntime.generated.Libruntime.LanguageType;

/**
 * The type Function meta.
 *
 * @since 2023/11/24
 */
public class FunctionMetaFactory {
    /**
     * The Function meta
     *
     * @param appName the appName
     * @param moduleName the modelName
     * @param funcName the funcName
     * @param className the className
     * @param language the language
     * @param apiType the apiType
     * @param signature the signature
     * @return FunctionMeta
     */
    public static FunctionMeta getFunctionMeta(String appName, String moduleName, String funcName, String className,
        LanguageType language, ApiType apiType, String signature) {
        return FunctionMeta.newBuilder()
            .setApplicationName(appName)
            .setFunctionName(funcName)
            .setClassName(className)
            .setLanguage(language)
            .setApiType(apiType)
            .setSignature(signature)
            .setModuleName(moduleName)
            .build();
    }
}
