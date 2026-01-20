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

import org.yuanrong.call.CppFunctionHandler;
import org.yuanrong.call.CppInstanceCreator;
import org.yuanrong.call.FunctionHandler;
import org.yuanrong.call.GoFunctionHandler;
import org.yuanrong.call.GoInstanceCreator;
import org.yuanrong.call.InstanceCreator;
import org.yuanrong.call.JavaFunctionHandler;
import org.yuanrong.call.JavaInstanceCreator;
import org.yuanrong.call.VoidFunctionHandler;
import org.yuanrong.function.CppFunction;
import org.yuanrong.function.CppInstanceClass;
import org.yuanrong.function.GoFunction;
import org.yuanrong.function.GoInstanceClass;
import org.yuanrong.function.JavaFunction;
import org.yuanrong.function.JavaInstanceClass;
import org.yuanrong.function.YRFunc0;
import org.yuanrong.function.YRFunc1;
import org.yuanrong.function.YRFunc2;
import org.yuanrong.function.YRFunc3;
import org.yuanrong.function.YRFunc4;
import org.yuanrong.function.YRFunc5;
import org.yuanrong.function.YRFuncVoid0;
import org.yuanrong.function.YRFuncVoid1;
import org.yuanrong.function.YRFuncVoid2;
import org.yuanrong.function.YRFuncVoid3;
import org.yuanrong.function.YRFuncVoid4;
import org.yuanrong.function.YRFuncVoid5;

/**
 * The type YR Call.
 *
 * @since 2023/10/16
 */
public class YRCall extends YRGetInstance {
    /**
     * Function function handler.
     *
     * @param <R>  Return value type.
     * @param func Function name.
     * @return FunctionHandler Instance.
     *
     * @snippet{trimleft} FunctionExample.java function 样例代码
     */
    public static <R> FunctionHandler<R> function(YRFunc0<R> func) {
        return new FunctionHandler<>(func);
    }

    /**
     * Function function handler.
     *
     * @param <T0> Input parameter type.
     * @param <R>  Return value type.
     * @param func Function name.
     * @return FunctionHandler Instance.
     */
    public static <T0, R> FunctionHandler<R> function(YRFunc1<T0, R> func) {
        return new FunctionHandler<>(func);
    }

    /**
     * Function function handler.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <R>  Return value type.
     * @param func Function name.
     * @return FunctionHandler Instance.
     */
    public static <T0, T1, R> FunctionHandler<R> function(YRFunc2<T0, T1, R> func) {
        return new FunctionHandler<>(func);
    }

    /**
     * Function function handler.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <R>  Return value type.
     * @param func Function name.
     * @return FunctionHandler Instance.
     */
    public static <T0, T1, T2, R> FunctionHandler<R> function(YRFunc3<T0, T1, T2, R> func) {
        return new FunctionHandler<>(func);
    }

    /**
     * Function function handler.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T3> Input parameter type.
     * @param <R>  Return value type.
     * @param func Function name.
     * @return FunctionHandler Instance.
     */
    public static <T0, T1, T2, T3, R> FunctionHandler<R> function(YRFunc4<T0, T1, T2, T3, R> func) {
        return new FunctionHandler<>(func);
    }

    /**
     * Function function handler.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T3> Input parameter type.
     * @param <T4> Input parameter type.
     * @param <R>  Return value type.
     * @param func Function name.
     * @return FunctionHandler Instance.
     */
    public static <T0, T1, T2, T3, T4, R> FunctionHandler<R> function(YRFunc5<T0, T1, T2, T3, T4, R> func) {
        return new FunctionHandler<>(func);
    }

    /**
     * Function void function handler.
     *
     * @param func Function name.
     * @return VoidFunctionHandler Instance.
     */
    public static VoidFunctionHandler function(YRFuncVoid0 func) {
        return new VoidFunctionHandler(func);
    }

    /**
     * Function void function handler.
     *
     * @param <T0> Input parameter type.
     * @param func Function name.
     * @return VoidFunctionHandler Instance.
     */
    public static <T0> VoidFunctionHandler function(YRFuncVoid1<T0> func) {
        return new VoidFunctionHandler(func);
    }

    /**
     * Function void function handler.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param func Function name.
     * @return VoidFunctionHandler Instance.
     */
    public static <T0, T1> VoidFunctionHandler function(YRFuncVoid2<T0, T1> func) {
        return new VoidFunctionHandler(func);
    }

    /**
     * Function void function handler.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param func Function name.
     * @return VoidFunctionHandler Instance.
     */
    public static <T0, T1, T2> VoidFunctionHandler function(YRFuncVoid3<T0, T1, T2> func) {
        return new VoidFunctionHandler(func);
    }

    /**
     * Function void function handler.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T4> Input parameter type.
     * @param func Function name.
     * @return VoidFunctionHandler Instance.
     */
    public static <T0, T1, T2, T4> VoidFunctionHandler function(YRFuncVoid4<T0, T1, T2, T4> func) {
        return new VoidFunctionHandler(func);
    }

    /**
     * Function void function handler.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T4> Input parameter type.
     * @param <T5> Input parameter type.
     * @param func Function name.
     * @return VoidFunctionHandler Instance.
     */
    public static <T0, T1, T2, T4, T5> VoidFunctionHandler function(YRFuncVoid5<T0, T1, T2, T4, T5> func) {
        return new VoidFunctionHandler(func);
    }

    /**
     * Instance instance creator.
     *
     * @param <A>  Return value type.
     * @param func Function name.
     * @return InstanceCreator Instance.
     *
     * @snippet{trimleft} InstanceExample.java Instance example
     */
    public static <A> InstanceCreator<A> instance(YRFunc0<A> func) {
        return new InstanceCreator<>(func);
    }

    /**
     * Instance instance creator.
     *
     * @param <A>  Return value type.
     * @param func Function name.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @return InstanceCreator Instance.
     *
     * @snippet{trimleft} InstanceExample.java Instance name example
     */
    public static <A> InstanceCreator<A> instance(YRFunc0<A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Instance instance creator.
     *
     * @param <A> Return value type.
     * @param func Function name.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @param nameSpace Namespace of the named instance. When both name and nameSpace exist, the instance name is
     *                  concatenated into nameSpace-name. This field is currently used only for concatenation.
     * @return InstanceCreator Instance.
     *
     * @snippet{trimleft} InstanceExample.java Instance nameSpace example
     */
    public static <A> InstanceCreator<A> instance(YRFunc0<A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Instance instance creator.
     *
     * @param <T0> Input parameter type.
     * @param <A>  Return value type.
     * @param func Function name.
     * @return InstanceCreator Instance.
     */
    public static <T0, A> InstanceCreator<A> instance(YRFunc1<T0, A> func) {
        return new InstanceCreator<>(func);
    }

    /**
     * Instance instance creator.
     *
     * @param <T0> Input parameter type.
     * @param <A>  Input parameter type.
     * @param func Function name.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @return InstanceCreator Instance.
     */
    public static <T0, A> InstanceCreator<A> instance(YRFunc1<T0, A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Instance instance creator.
     *
     * @param <T0>      Input parameter type.
     * @param <A>       Return value type.
     * @param func      Function name.
     * @param name      The instance name of the named instance, the second parameter.
     *                  When only name exists, the instance name will be set to name.
     * @param nameSpace Namespace of the named instance. When both name and nameSpace exist, the instance name is
     *                  concatenated into nameSpace-name. This field is currently used only for concatenation.
     * @return InstanceCreator Instance.
     */
    public static <T0, A> InstanceCreator<A> instance(YRFunc1<T0, A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Instance instance creator.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <A>  Return value type.
     * @param func Function name.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, A> InstanceCreator<A> instance(YRFunc2<T0, T1, A> func) {
        return new InstanceCreator<>(func);
    }

    /**
     * Instance instance creator.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <A>  Return value type.
     * @param func Function name.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, A> InstanceCreator<A> instance(YRFunc2<T0, T1, A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Instance instance creator.
     *
     * @param <T0>      Input parameter type.
     * @param <T1>      Input parameter type.
     * @param <A>       Return value type.
     * @param func      Function name.
     * @param name      The instance name of the named instance, the second parameter.
     *                  When only name exists, the instance name will be set to name.
     * @param nameSpace Namespace of the named instance. When both name and nameSpace exist, the instance name is
     *                  concatenated into nameSpace-name.
     *                  This field is currently used only for concatenation.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, A> InstanceCreator<A> instance(YRFunc2<T0, T1, A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Instance instance creator.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <A>  Return value type.
     * @param func Function name.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, T2, A> InstanceCreator<A> instance(YRFunc3<T0, T1, T2, A> func) {
        return new InstanceCreator<>(func);
    }

    /**
     * Instance instance creator.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <A>  Return value type.
     * @param func Function name.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, T2, A> InstanceCreator<A> instance(YRFunc3<T0, T1, T2, A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Instance instance creator.
     *
     * @param <T0>      Input parameter type.
     * @param <T1>      Input parameter type.
     * @param <T2>      Input parameter type.
     * @param <A>       Return value type.
     * @param func      Function name.
     * @param name      The instance name of the named instance, the second parameter.
     *                  When only name exists, the instance name will be set to name.
     * @param nameSpace Namespace of the named instance. When both name and nameSpace exist, the instance name is
     *                  concatenated into nameSpace-name.
     *                  This field is currently used only for concatenation.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, T2, A> InstanceCreator<A> instance(
            YRFunc3<T0, T1, T2, A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Instance instance creator.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T3> Input parameter type.
     * @param <A>  Return value type.
     * @param func Function name.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, T2, T3, A> InstanceCreator<A> instance(YRFunc4<T0, T1, T2, T3, A> func) {
        return new InstanceCreator<>(func);
    }

    /**
     * Instance instance creator.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T3> Input parameter type.
     * @param <A>  Return value type.
     * @param func Function name.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, T2, T3, A> InstanceCreator<A> instance(YRFunc4<T0, T1, T2, T3, A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Instance instance creator.
     *
     * @param <T0>      Input parameter type.
     * @param <T1>      Input parameter type.
     * @param <T2>      Input parameter type.
     * @param <T3>      Input parameter type.
     * @param <A>       Return value type.
     * @param func      Function name.
     * @param name      The instance name of the named instance, the second parameter.
     *                  When only name exists, the instance name will be set to name.
     * @param nameSpace Namespace of the named instance. When both name and nameSpace exist, the instance name is
     *                  concatenated into nameSpace-name.
     *                  This field is currently used only for concatenation.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, T2, T3, A> InstanceCreator<A> instance(
            YRFunc4<T0, T1, T2, T3, A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Instance instance creator.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T3> Input parameter type.
     * @param <T4> Input parameter type.
     * @param <A>  Return value type.
     * @param func Function name.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, T2, T3, T4, A> InstanceCreator<A> instance(YRFunc5<T0, T1, T2, T3, T4, A> func) {
        return new InstanceCreator<>(func);
    }

    /**
     * Instance instance creator.
     *
     * @param <T0> Input parameter type.
     * @param <T1> Input parameter type.
     * @param <T2> Input parameter type.
     * @param <T3> Input parameter type.
     * @param <T4> Input parameter type.
     * @param <A>  Return value type.
     * @param func Function name.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, T2, T3, T4, A> InstanceCreator<A> instance(
            YRFunc5<T0, T1, T2, T3, T4, A> func, String name) {
        return new InstanceCreator<>(func, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Instance instance creator.
     *
     * @param <T0>      Input parameter type.
     * @param <T1>      Input parameter type.
     * @param <T2>      Input parameter type.
     * @param <T3>      Input parameter type.
     * @param <T4>      Input parameter type.
     * @param <A>       Return value type.
     * @param func      Function name.
     * @param name      The instance name of the named instance, the second parameter.
     *                  When only name exists, the instance name will be set to name.
     * @param nameSpace Namespace of the named instance. When both name and nameSpace exist, the instance name is
     *                  concatenated into nameSpace-name.
     *                  This field is currently used only for concatenation.
     * @return InstanceCreator Instance.
     */
    public static <T0, T1, T2, T3, T4, A> InstanceCreator<A> instance(
            YRFunc5<T0, T1, T2, T3, T4, A> func, String name, String nameSpace) {
        return new InstanceCreator<>(func, name, nameSpace);
    }

    /**
     * Function java function handler.
     *
     * @param <R> Return value type.
     * @param javaFunction Java Function name.
     * @return JavaFunctionHandler Instance.
     */
    public static <R> JavaFunctionHandler<R> function(JavaFunction<R> javaFunction) {
        return new JavaFunctionHandler<>(javaFunction);
    }

    /**
     * Instance java instance creator.
     *
     * @param javaInstanceClass the java function instance class.
     * @return JavaInstanceCreator Instance.
     */
    public static JavaInstanceCreator instance(JavaInstanceClass javaInstanceClass) {
        return new JavaInstanceCreator(javaInstanceClass, "", ConfigManager.getInstance().getNs());
    }

    /**
     * Instance java instance creator.
     *
     * @param javaInstanceClass the java instance class.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @return JavaInstanceCreator Instance.
     */
    public static JavaInstanceCreator instance(JavaInstanceClass javaInstanceClass, String name) {
        return new JavaInstanceCreator(javaInstanceClass, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Instance java instance creator.
     *
     * @param javaInstanceClass the java instance class.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @param nameSpace Namespace of the named instance. When both name and nameSpace exist, the instance name is
     *                  concatenated into nameSpace-name.
     *                  This field is currently used only for concatenation.
     * @return JavaInstanceCreator Instance.
     */
    public static JavaInstanceCreator instance(JavaInstanceClass javaInstanceClass, String name, String nameSpace) {
        return new JavaInstanceCreator(javaInstanceClass, name, nameSpace);
    }

    /**
     * Function cpp function handler.
     *
     * @param <R> Return value type.
     * @param cppFunction C++ Function name.
     * @return CppFunctionHandler Instance.
     */
    public static <R> CppFunctionHandler<R> function(CppFunction<R> cppFunction) {
        return new CppFunctionHandler<>(cppFunction);
    }

    /**
     * Instance cpp instance creator.
     *
     * @param cppInstanceClass the cpp function instance class.
     * @return CppInstanceCreator Instance.
     */
    public static CppInstanceCreator instance(CppInstanceClass cppInstanceClass) {
        return new CppInstanceCreator(cppInstanceClass, "", ConfigManager.getInstance().getNs());
    }

    /**
     * Instance cpp instance creator.
     *
     * @param cppInstanceClass the cpp instance class.
     * @param name             The instance name of the named instance, the second parameter.
     *                         When only name exists, the instance name will be set to name.
     * @return CppInstanceCreator Instance.
     */
    public static CppInstanceCreator instance(CppInstanceClass cppInstanceClass, String name) {
        return new CppInstanceCreator(cppInstanceClass, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Instance cpp instance creator.
     *
     * @param cppInstanceClass the cpp instance class.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @param nameSpace Namespace of the named instance. When both name and nameSpace exist, the instance name is
     *                  concatenated into nameSpace-name.
     *                  This field is currently used only for concatenation.
     * @return CppInstanceCreator Instance.
     */
    public static CppInstanceCreator instance(CppInstanceClass cppInstanceClass, String name, String nameSpace) {
        return new CppInstanceCreator(cppInstanceClass, name, nameSpace);
    }

    /**
     * Function go function handler.
     *
     * @param <R> Return value type.
     * @param goFunction  go Function name.
     * @return GoFunctionHandler Instance.
     */
    public static <R> GoFunctionHandler<R> function(GoFunction<R> goFunction) {
        return new GoFunctionHandler<>(goFunction);
    }

    /**
     * Instance go instance creator.
     *
     * @param goInstanceClass the go function instance class.
     * @return GoInstanceCreator Instance.
     */
    public static GoInstanceCreator instance(GoInstanceClass goInstanceClass) {
        return new GoInstanceCreator(goInstanceClass, "", ConfigManager.getInstance().getNs());
    }

    /**
     * Instance go instance creator.
     *
     * @param goInstanceClass the go instance class.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @return GoInstanceCreator Instance.
     */
    public static GoInstanceCreator instance(GoInstanceClass goInstanceClass, String name) {
        return new GoInstanceCreator(goInstanceClass, name, ConfigManager.getInstance().getNs());
    }

    /**
     * Instance go instance creator.
     *
     * @param goInstanceClass the go instance class.
     * @param name The instance name of the named instance, the second parameter.
     *             When only name exists, the instance name will be set to name.
     * @param nameSpace Namespace of the named instance. When both name and nameSpace exist, the instance name is
     *                  concatenated into nameSpace-name.
     *                  This field is currently used only for concatenation.
     * @return GoInstanceCreator Instance.
     */
    public static GoInstanceCreator instance(GoInstanceClass goInstanceClass, String name, String nameSpace) {
        return new GoInstanceCreator(goInstanceClass, name, nameSpace);
    }
}
