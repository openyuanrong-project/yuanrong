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

package org.yuanrong.function;

/**
 * Helper class to new java instance class.
 *
 * @since 2024/04/16
 */
public class JavaInstanceClass {
    /**
     * the name of class.
     */
    public final String className;

    /**
     * Constructor of JavaInstanceClass.
     *
     * @param className the name of class.
     */
    JavaInstanceClass(String className) {
        this.className = className;
    }

    /**
     * Java function creates Java class instance.
     *
     * @param className Class name of Java instance.
     * @return JavaInstanceClass Instance.
     *
     * @snippet{trimleft} JavaInstanceExample.java JavaInstanceClass 样例代码
     */
    public static JavaInstanceClass of(String className) {
        return new JavaInstanceClass(className);
    }
}
