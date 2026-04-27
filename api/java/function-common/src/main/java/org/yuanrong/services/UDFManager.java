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

package org.yuanrong.services;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * The type Udf manager.
 *
 * @since 2022-07-20
 */
public class UDFManager {
    /**
     * Keep UDF key and object instance
     */
    private final Map<String, Object> udfMap = new ConcurrentHashMap<>();

    /**
     * UDF handler class
     */
    private Class<?> userClass;

    private static class SingletonUDFManager {
        private static final UDFManager UDF_MANAGER = new UDFManager();
    }

    /**
     * Gets udf manager.
     *
     * @return the udf manager
     */
    public static UDFManager getUDFManager() {
        return SingletonUDFManager.UDF_MANAGER;
    }

    /**
     * Register.
     *
     * @param functionKey the function key
     * @param instance    the instance
     */
    public void registerInstance(String functionKey, Object instance) {
        udfMap.put(functionKey, instance);
    }

    /**
     * Register class.
     *
     * @param userClass the user class
     */
    public void registerClass(Class<?> userClass) {
        this.userClass = userClass;
    }

    /**
     * Load object.
     *
     * @param functionKey the function key
     * @return the object
     */
    public Object loadInstance(String functionKey) {
        return udfMap.get(functionKey);
    }

    /**
     * Load class class.
     *
     * @return the class
     */
    public Class<?> loadClass() {
        return this.userClass;
    }
}
