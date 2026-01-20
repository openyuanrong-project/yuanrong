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

package org.yuanrong.function;

/**
 * Helper class to new cpp instance class.
 *
 * @since 2022/08/26
 */
public class CppInstanceClass {
    /**
     * the name of class.
     */
    public final String className;

    /**
     * the name of init function.
     */
    public final String initFunctionName;

    /**
     * Constructor of CppInstanceClass.
     *
     * @param className the name of class.
     * @param initFunctionName the name of init function.
     */
    CppInstanceClass(String className, String initFunctionName) {
        this.className = className;
        this.initFunctionName = initFunctionName;
    }

    /**
     * Java function creates C++ class instance.
     *
     * @param className Class name of C++ instance.
     * @param initFunctionName C++ class initialization function.
     * @return CppInstanceClass instance.
     *
     * @snippet{trimleft} CppFunctionExample.java CppInstanceClass 样例代码
     */
    public static CppInstanceClass of(String className, String initFunctionName) {
        return new CppInstanceClass(className, initFunctionName);
    }
}
