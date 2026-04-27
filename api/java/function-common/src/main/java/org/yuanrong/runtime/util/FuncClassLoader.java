/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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

package org.yuanrong.runtime.util;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.URL;
import java.net.URLClassLoader;
import java.util.ArrayList;
import java.util.List;

/**
 * Description:
 * for one thing, this class is avoid duplicate load between UDF and runtime.
 * For another, it load UDF first to make sure its customised setting works, especially for LOG.
 * define a function class loader, which is isolated from the Runtime Lib between class loader.
 * This loader follows parental delegation.
 * The parent classloader is Ext ClassLoader, which is at the same level as AppClassLoader.
 *
 * @since 2023/10/23
 */
public class FuncClassLoader extends URLClassLoader {
    private static final Logger log = LoggerFactory.getLogger(FuncClassLoader.class);

    private static final List<String> YR_SYSTEM_STARTER =
            new ArrayList<String>() {
                {
                    add("org.yuanrong");
                    add("com.services");
                    add("com.function");
                    add("com.datasystem");
                    add("com.faas");
                    add("com.google.protobuf");
                    add("com.google.gson");
                }
            };

    static {
        ClassLoader.registerAsParallelCapable();
    }

    public FuncClassLoader(URL[] urls, ClassLoader parent) {
        super(urls, parent);
    }

    /**
     * Search based on the parental delegation mechanism.
     * If the Bootstrap ClassLoader and Ext ClassLoader cannot be loaded, run the Classloader findClass method.
     * Order that Class Loader Lookup :
     *                           Bootstrap ClassLoader
     *                                     |
     *                              Ext Classloader
     *           not start with      /           \  start whit com.faas || org.yuanrong.runtime
     *                   FuncClassLoader   SystemClassLoader
     *                              |             |
     *                   SystemClassLoader     FuncClassLoader
     *
     * @param name    class name
     * @return Class class obj
     * @throws ClassNotFoundException class not exit
     */
    @Override
    protected Class<?> findClass(String name) throws ClassNotFoundException {
        Class<?> clz;
        for (String yrSystem : YR_SYSTEM_STARTER) {
            if (name.startsWith(yrSystem)) {
                try {
                    clz = ClassLoader.getSystemClassLoader().loadClass(name);
                } catch (ClassNotFoundException exp) {
                    clz = super.findClass(name);
                }
                return clz;
            }
        }
        try {
            clz = super.findClass(name);
        } catch (ClassNotFoundException exp) {
            clz = ClassLoader.getSystemClassLoader().loadClass(name);
        }
        return clz;
    }
}
