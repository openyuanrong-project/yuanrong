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

package org.yuanrong.test;

import org.yuanrong.InvokeOptions;
import org.yuanrong.affinity.Affinity;
import org.yuanrong.affinity.AffinityKind;
import org.yuanrong.affinity.AffinityType;
import org.yuanrong.affinity.LabelOperator;
import org.yuanrong.affinity.OperatorType;
import org.yuanrong.api.YR;
import org.yuanrong.call.CppInstanceHandler;
import org.yuanrong.call.InstanceHandler;
import org.yuanrong.call.JavaInstanceHandler;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.CppInstanceClass;
import org.yuanrong.function.CppInstanceMethod;
import org.yuanrong.function.JavaInstanceClass;
import org.yuanrong.function.JavaInstanceMethod;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.testutils.Counter;
import org.yuanrong.testutils.TestUtils;

import org.junit.Ignore;
import org.junit.Test;
import org.testng.Assert;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.nio.charset.Charset;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * ActorTest
 */
public class ActorTest {
    /*case
     * @title: java通过函数名调用java 有状态函数
     * @precondition:
     * @step:   1.调用openYuanRong 有状态函数 returnInt
     * @expect: 1.预期无异常抛出
     */
    @Test
    public void test_java_task_invoke_successfully() throws Exception {
        TestUtils.initYR(false);
        try {
            JavaInstanceHandler javaInstance = YR.instance(JavaInstanceClass.of("org.yuanrong.testutils.TestUtils"))
                .setUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                .invoke();
            ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("returnInt", int.class)).invoke(1);
            int res = (int) YR.get(ref1, 10000);
            Assert.assertEquals(res, 1);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java通过函数名调用java 有状态函数
     * @precondition:
     * @step:   1.调用openYuanRong 有状态函数 returnInt
     * @expect: 1.预期有异常抛出
     */
    @Test
    public void test_java_task_invoke_failed() throws Exception {
        TestUtils.initYR(false);
        try {
            boolean isException = false;
            try {
                JavaInstanceHandler javaInstance = YR.instance(JavaInstanceClass.of("TestUtils"))
                    .setUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                    .invoke();
                ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("returnInt", int.class)).invoke(1);
                YR.get(ref1, 10000);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("ClassNotFoundException"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                JavaInstanceHandler javaInstance = YR.instance(
                    JavaInstanceClass.of("org.yuanrong.testutils.TestUtils")).invoke();
                ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("returnInt", int.class)).invoke(1);
                YR.get(ref1, 10000);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("invalid function"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                JavaInstanceHandler javaInstance = YR.instance(
                    JavaInstanceClass.of("org.yuanrong.testutils.TestUtils")).setUrn("abc123").invoke();
                ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("returnInt", int.class)).invoke(1);
                YR.get(ref1, 10000);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("function urn format is wrong"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                JavaInstanceHandler javaInstance = YR.instance(
                        JavaInstanceClass.of("org.yuanrong.testutils.Counter"))
                    .setUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                    .invoke();
                ObjectRef ref1 = javaInstance.function(JavaInstanceMethod.of("returnIntWithException", int.class))
                    .invoke(1);
                YR.get(ref1, 10000);
            } catch (YRException e) {
                Assert.assertTrue(e.getErrorCode().toString().equals("2002"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java通过函数名调用cpp 有状态函数
     * @precondition:
     * @step:   1.调用openYuanRong 有状态函数 AddTwo
     * @expect: 1.预期无异常抛出
     */
    @Test
    public void test_cpp_task_invoke_successfully() throws Exception {
        TestUtils.initYR(false);
        try {
            CppInstanceHandler cppInstance = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
                .setUrn("sn:cn:yrk:default:function:0-yr-stcpp:$latest")
                .invoke(1);
            ObjectRef ref1 = cppInstance.function(CppInstanceMethod.of("AddTwo", int.class)).invoke(2, 3);
            int res = (int) YR.get(ref1, 10000);
            Assert.assertEquals(res, 6);

            cppInstance.terminate();
        } finally {
            YR.Finalize();
        }
    }

    @Ignore("Check whether customextension has been written into the request body by viewing log file.")
    @Test
    public void test_invoke_instance_with_customextension() throws Exception {
        TestUtils.initYR();
        try {
            InvokeOptions createOpts = InvokeOptions.builder()
                .addCustomExtension("endpoint", "CreateInstance1")
                .addCustomExtension("app_name", "CreateInstance2")
                .addCustomExtension("tenant_id", "CreateInstance3")
                .build();
            InstanceHandler ins = YR.instance(Counter::new).options(createOpts).invoke();
            InvokeOptions invokeOpts = InvokeOptions.builder()
                .addCustomExtension("endpoint", "InvokeInstance1")
                .addCustomExtension("app_name", "InvokeInstance2")
                .addCustomExtension("tenant_id", "InvokeInstance3")
                .build();
            ObjectRef ref = ins.function(Counter::addOne).options(invokeOpts).invoke();
            int res = (int) YR.get(ref, 10);
            Assert.assertEquals(res, 1);
        } finally {
            YR.Finalize();
        }
    }

    @Ignore("Check whether customextension has been written into the request body by viewing log file.")
    @Test
    public void test_anti_others_success() throws Exception {
        TestUtils.initYR();
        try {
            LabelOperator op = new LabelOperator(OperatorType.LABEL_EXISTS, "label1");
            List<LabelOperator> operatorList = new ArrayList<>();
            operatorList.add(op);
            Affinity affinity = new Affinity(AffinityKind.RESOURCE, AffinityType.PREFERRED, operatorList);
            InvokeOptions createOpts = InvokeOptions.builder()
                .addScheduleAffinity(affinity)
                .preferredAntiOtherLabels(true)
                .build();
            YR.instance(Counter::new).options(createOpts).invoke(1);
        } finally {
            YR.Finalize();
        }
    }

    @Ignore("Have to manually send SIGTERM signals to kill runtime processe during sleep.")
    @Test
    public void test_shutdown_with_manual_sigterm() throws Exception {
        TestUtils.initYR();
        try {
            YR.kv().del(Counter.SHUTDOWN_KEY);
            YR.instance(Counter::new).invoke();
            Thread.sleep(30000);

            byte[] byteArray = YR.kv().get(Counter.SHUTDOWN_KEY, 30);
            String shutdownValue = new String(byteArray, StandardCharsets.UTF_8);
            Assert.assertEquals(shutdownValue, Counter.SHUTDOWN_VALUE);
        } finally {
            YR.Finalize();
        }
    }

    @Test
    public void test_graceful_shutdown_with_terminate() throws Exception {
        TestUtils.initYR();
        try {
            YR.kv().del(Counter.SHUTDOWN_KEY);
            InstanceHandler ins = YR.instance(Counter::new).invoke();
            ObjectRef ref = ins.function(Counter::addOne).invoke();
            YR.get(ref, 10);
            ins.terminate();

            byte[] byteArray = YR.kv().get(Counter.SHUTDOWN_KEY, 30);
            String shutdownValue = new String(byteArray, StandardCharsets.UTF_8);
            Assert.assertEquals(shutdownValue, Counter.SHUTDOWN_VALUE);
        } finally {
            YR.Finalize();
        }
    }

    @Test
    public void test_save_load_state() throws Exception {
        TestUtils.initYR();
        try {
            InstanceHandler ins = YR.instance(Counter::new).invoke();
            ins.function(Counter::save).invoke();

            ObjectRef ref = ins.function(Counter::addOne).invoke();
            Assert.assertEquals(YR.get(ref, 10), 1);

            ref = ins.function(Counter::load).invoke();
            YR.get(ref, 10);

            ref = ins.function(Counter::get).invoke();
            Assert.assertEquals(YR.get(ref, 10), 0);
        } finally {
            YR.Finalize();
        }
    }

    @Test
    public void test_index_out_of_bounds() throws Exception {
        TestUtils.initYR();
        try {
            InstanceHandler h_myYRapp = YR.instance(Counter::new).invoke();
            ObjectRef res = h_myYRapp.function(Counter::indexOutOfBoundsTest).invoke();
            res.setType(String.class);
            boolean isException = false;
            try {
                YR.get(res, 30);
            } catch (YRException e) {
                Assert.assertTrue(e.getErrorCode().toString().equals("2002"), TestUtils.getStackTraceStr(e));
                Assert.assertTrue(e.getErrorMessage().contains("throwable: java.lang.IndexOutOfBoundsException"));
                isException = true;
            }
            Assert.assertTrue(isException);
        } finally {
            YR.Finalize();
        }
    }

    @Test
    public void test_hello() throws Exception {
        TestUtils.initYR();
        try {
            InstanceHandler h_myYRapp = YR.instance(Counter::new).invoke();
            ObjectRef res = h_myYRapp.function(Counter::hello).invoke();
            res.setType(String.class);
            Assert.assertEquals(YR.get(res, 30), "hello world");
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java支持重拉配置并支持用户自定义recover
     * @precondition:
     * @step:   1.配置invokeoptions中的recoverRetryTimes大于0
     * @step:   2.配置yrRecover
     * @step:   3.创建实例，获取其进程号并kill
     * @step:   4.再次调用实例的成员函数
     * @expect: 1.实例被kill后成功recover，并且调用yrRecover
     */
    @Test
    public void test_recover_successfully() throws Exception {
        TestUtils.initYR();
        try {
            InvokeOptions opts = InvokeOptions.builder().recoverRetryTimes(1).build();
            InstanceHandler ins = YR.instance(Counter::new).options(opts).invoke();
            ObjectRef ref = ins.function(Counter::getPid).invoke();
            long pid = (long) YR.get(ref, 10);
            ProcessBuilder pb = new ProcessBuilder("kill", "-9", String.valueOf(pid));
            pb.inheritIO();
            pb.start().waitFor();
            byte[] byteArray = YR.kv().get(Counter.RECOVER_KEY, 30);
            String recoverValue = new String(byteArray, StandardCharsets.UTF_8);
            Assert.assertEquals(recoverValue, Counter.RECOVER_VALUE);
            YR.kv().del(Counter.RECOVER_KEY);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java支持重拉配置并支持用户自定义recover
     * @precondition:
     * @step:   1.配置invokeoptions中的recoverRetryTimes大于0
     * @step:   2.创建实例，获取其进程号并kill
     * @step:   3.再次调用实例的成员函数
     * @expect: 1.实例被kill后成功recover
     */
    @Ignore("need to repair.")
    @Test
    public void test_recover_cpp_successfully() throws Exception {
        TestUtils.initYR(false);
        try {
            InvokeOptions opts = InvokeOptions.builder().recoverRetryTimes(1).build();
            CppInstanceHandler cppInstance = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
                .setUrn("sn:cn:yrk:default:function:0-yr-stcpp:$latest")
                .options(opts)
                .invoke(1);
            ObjectRef ref1 = cppInstance.function(CppInstanceMethod.of("GetPid", long.class)).invoke();
            long pid = (long) YR.get(ref1, 10000);
            ProcessBuilder pb = new ProcessBuilder("kill", "-9", String.valueOf(pid));
            pb.inheritIO();
            pb.start().waitFor();
            ObjectRef ref2 = cppInstance.function(CppInstanceMethod.of("GetRecoverFlag", int.class)).invoke();
            int res2 = (int) YR.get(ref2, 10000);
            System.out.println("res: " + res2);
            Assert.assertTrue(res2 == 1);
            cppInstance.terminate();
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java重拉配置错误
     * @precondition:
     * @step:   1.配置invokeoptions中的recoverRetryTimes小于0
     * @expect: 1.实例被kill后成功recover，并且调用yrRecover
     */
    @Test
    public void test_recover_failed() throws Exception {
        TestUtils.initYR();
        boolean isException = false;
        try {
            InvokeOptions opts = InvokeOptions.builder().recoverRetryTimes(-1).build();
            InstanceHandler ins = YR.instance(Counter::new).options(opts).invoke();
        } catch (YRException e) {
            Assert.assertTrue(e.getErrorCode().toString().equals("1001"), TestUtils.getStackTraceStr(e));
            Assert.assertTrue(e.getErrorMessage().contains("is invalid, which must be non-nagative"));
            isException = true;
        } finally {
            YR.Finalize();
        }
        Assert.assertTrue(isException);
    }

    /*case
     * @title: java init失败
     * @precondition:
     * @step:   1.配置config非法，并init
     * @expect: 1.init抛出异常
     */
    @Test
    public void test_init_failed() throws Exception {
        boolean isException = false;
        try {
            TestUtils.initYRInvalid();
        } catch (YRException e) {
            Assert.assertTrue(e.getErrorCode().toString().equals("4001"), TestUtils.getStackTraceStr(e));
            Assert.assertTrue(e.getErrorMessage().contains("dataSystemAddress is invalid"));
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    /*case
     * @title: java默认开启保序
     * @precondition:
     * @step:   1.调用addOne方法10次，验证每次获取的结果
     * @expect: 1.获取结果保序
     */
    @Test
    public void test_need_order_successfully() throws Exception {
        TestUtils.initYR();
        int jobNums = 10;
        try {
            InstanceHandler ins = YR.instance(Counter::new).invoke();
            ObjectRef[] objectRefs = new ObjectRef[jobNums];
            for (int i = 0; i < jobNums; i++) {
                objectRefs[i] = ins.function(Counter::addOne).invoke();
            }

            for (int i = 0; i < jobNums; i++) {
                Assert.assertEquals(YR.get(objectRefs[i], 10), i + 1);
            }
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: Concurrency为1时默认开启保序，大于1时默认关闭保序
     * @precondition:
     * @step:   1.设置Concurrency为10和1，验证InvokeOptions的保序配置值
     * @expect: 1.Concurrency为10时，保序默认关，Concurrency为1时，保序默认开
     */
    @Test
    public void test_need_order_default_value() throws Exception {
        TestUtils.initYR();
        try {
            InvokeOptions option = new InvokeOptions();
            option.getCustomExtensions().put("Concurrency", "10");
            InstanceHandler ins = YR.instance(Counter::new).options(option).invoke();
            Assert.assertFalse(ins.isNeedOrder());
            Assert.assertTrue(YR.instance(Counter::new).invoke().isNeedOrder());
        } finally {
            YR.Finalize();
        }
    }
}
