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

package org.yuanrong;

import org.yuanrong.call.CppInstanceCreator;
import org.yuanrong.call.InstanceCreator;
import org.yuanrong.function.CppInstanceClass;
import org.yuanrong.function.YRFunc0;
import org.yuanrong.function.YRFunc1;
import org.yuanrong.function.YRFunc2;
import org.yuanrong.function.YRFunc3;
import org.yuanrong.function.YRFunc4;
import org.yuanrong.function.YRFunc5;

/**
 * The type YRGetInstance.
 * To Inheritance the old Java runtime interface, 'getinstance' is
 * retained. If all these interfaces are depreciated, this file can be deleted
 * directly.
 *
 * @since 2023 /12/11
 */
public class YRGetInstance {
    /**
     * Gets instance.
     *
     * @param <A>  the type parameter
     * @param func the func
     * @param name the name
     * @return the instance
     */
    public static <A> InstanceCreator<A> getInstance(YRFunc0<A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Gets instance.
     *
     * @param <A>       the type parameter
     * @param func      the func
     * @param name      the name
     * @param nameSpace the name space
     * @return the instance
     */
    public static <A> InstanceCreator<A> getInstance(YRFunc0<A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Gets instance.
     *
     * @param <T0> the type parameter
     * @param <A>  the type parameter
     * @param func the func
     * @param name the name
     * @return the instance
     */
    public static <T0, A> InstanceCreator<A> getInstance(YRFunc1<T0, A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Gets instance.
     *
     * @param <T0>      the type parameter
     * @param <A>       the type parameter
     * @param func      the func
     * @param name      the name
     * @param nameSpace the name space
     * @return the instance
     */
    public static <T0, A> InstanceCreator<A> getInstance(YRFunc1<T0, A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Gets instance.
     *
     * @param <T0> the type parameter
     * @param <T1> the type parameter
     * @param <A>  the type parameter
     * @param func the func
     * @param name the name
     * @return the instance
     */
    public static <T0, T1, A> InstanceCreator<A> getInstance(YRFunc2<T0, T1, A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Gets instance.
     *
     * @param <T0>      the type parameter
     * @param <T1>      the type parameter
     * @param <A>       the type parameter
     * @param func      the func
     * @param name      the name
     * @param nameSpace the name space
     * @return the instance
     */
    public static <T0, T1, A> InstanceCreator<A> getInstance(YRFunc2<T0, T1, A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Gets instance.
     *
     * @param <T0> the type parameter
     * @param <T1> the type parameter
     * @param <T2> the type parameter
     * @param <A>  the type parameter
     * @param func the func
     * @param name the name
     * @return the instance
     */
    public static <T0, T1, T2, A> InstanceCreator<A> getInstance(YRFunc3<T0, T1, T2, A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Gets instance.
     *
     * @param <T0>      the type parameter
     * @param <T1>      the type parameter
     * @param <T2>      the type parameter
     * @param <A>       the type parameter
     * @param func      the func
     * @param name      the name
     * @param nameSpace the name space
     * @return the instance
     */
    public static <T0, T1, T2, A> InstanceCreator<A> getInstance(
            YRFunc3<T0, T1, T2, A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Gets instance.
     *
     * @param <T0> the type parameter
     * @param <T1> the type parameter
     * @param <T2> the type parameter
     * @param <T3> the type parameter
     * @param <A>  the type parameter
     * @param func the func
     * @param name the name
     * @return the instance
     */
    public static <T0, T1, T2, T3, A> InstanceCreator<A> getInstance(YRFunc4<T0, T1, T2, T3, A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Gets instance.
     *
     * @param <T0>      the type parameter
     * @param <T1>      the type parameter
     * @param <T2>      the type parameter
     * @param <T3>      the type parameter
     * @param <A>       the type parameter
     * @param func      the func
     * @param name      the name
     * @param nameSpace the name space
     * @return the instance
     */
    public static <T0, T1, T2, T3, A> InstanceCreator<A> getInstance(
            YRFunc4<T0, T1, T2, T3, A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Gets instance.
     *
     * @param <T0> the type parameter
     * @param <T1> the type parameter
     * @param <T2> the type parameter
     * @param <T3> the type parameter
     * @param <T4> the type parameter
     * @param <A>  the type parameter
     * @param func the func
     * @param name the name
     * @return the instance
     */
    public static <T0, T1, T2, T3, T4, A> InstanceCreator<A> getInstance(
            YRFunc5<T0, T1, T2, T3, T4, A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Gets instance.
     *
     * @param <T0> the type parameter
     * @param <T1> the type parameter
     * @param <T2> the type parameter
     * @param <T3> the type parameter
     * @param <T4> the type parameter
     * @param <A> the type parameter
     * @param func the func
     * @param name the name
     * @param nameSpace the name space
     * @return the instance
     */
    public static <T0, T1, T2, T3, T4, A> InstanceCreator<A> getInstance(
            YRFunc5<T0, T1, T2, T3, T4, A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Gets cpp instance.
     *
     * @param cppInstanceClass the cpp instance class
     * @param name the name
     * @return the cpp instance
     */
    public static CppInstanceCreator getCppInstance(CppInstanceClass cppInstanceClass, String name) {
        return new CppInstanceCreator(cppInstanceClass, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Gets cpp instance.
     *
     * @param cppInstanceClass the cpp instance class
     * @param name             the name
     * @param nameSpace        the name space
     * @return the cpp instance
     */
    public static CppInstanceCreator getCppInstance(CppInstanceClass cppInstanceClass, String name, String nameSpace) {
        return new CppInstanceCreator(cppInstanceClass, name, nameSpace);
    }
}
