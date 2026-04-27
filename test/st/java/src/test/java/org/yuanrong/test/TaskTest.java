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

package org.yuanrong.test;

import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

import org.junit.Ignore;
import org.junit.Test;
import org.testng.Assert;

import org.yuanrong.Config;
import org.yuanrong.InvokeOptions;
import org.yuanrong.affinity.Affinity;
import org.yuanrong.affinity.AffinityKind;
import org.yuanrong.affinity.AffinityType;
import org.yuanrong.affinity.LabelOperator;
import org.yuanrong.affinity.OperatorType;
import org.yuanrong.api.YR;
import org.yuanrong.CacheType;
import org.yuanrong.ConsistencyType;
import org.yuanrong.CreateParam;
import org.yuanrong.ExistenceOpt;
import org.yuanrong.MSetParam;
import org.yuanrong.GetParam;
import org.yuanrong.GetParams;
import org.yuanrong.SetParam;
import org.yuanrong.WriteMode;
import org.yuanrong.exception.YRException;
import org.yuanrong.function.CppFunction;
import org.yuanrong.function.JavaFunction;
import org.yuanrong.runtime.client.ObjectRef;
import org.yuanrong.testutils.Counter;
import org.yuanrong.testutils.TestUtils;

/**
 * TaskTest
 */
public class TaskTest {
    /* case
     * @title: 多线程调用无状态函数
     * @precondition:
     * @step: 1. 创建两个线程，每个线程 invoke 无状态函数 10次
     * @expect:  1.预期无异常抛出
     */
    @Test
    public void test_multithread_invoke() throws Exception {
        TestUtils.initYR();
        try {
            int theadNums = 2;
            int jobNums = 10;
            ExecutorService threadpool = Executors.newFixedThreadPool(theadNums);
            // 1. define stateless function
            Runnable invoke = () -> {
                ObjectRef obj;
                try {
                    obj = YR.function(TestUtils::returnInt).invoke(1);
                    YR.get(obj, 30);
                } catch (YRException e) {
                    e.printStackTrace();
                }
            };

            Callable<String> task = () -> {
                for (int i = 0; i < jobNums; i++) {
                    invoke.run();
                }
                return "ok";
            };
            // 2. submit tasks
            List<Future<String>> futures = new ArrayList<>();
            for (int i = 0; i < theadNums; i++) {
                futures.add(threadpool.submit(task));
            }

            threadpool.shutdown();
            // 3. assert returned value
            for (Future<String> future : futures) {
                Assert.assertEquals(future.get(), "ok");
            }
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java通过函数名调用java 无状态函数
     * @precondition:
     * @step:   1.调用openYuanRong 无状态函数 returnInt
     * @expect: 1.预期无异常抛出
     */
    @Test
    public void test_java_task_invoke_successfully() throws Exception {
        TestUtils.initYR(false);
        try {
            ObjectRef ref1 = YR.function(JavaFunction.of("org.yuanrong.testutils.TestUtils", "returnInt", int.class))
                .setUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                .invoke(1);
            int res = (int) YR.get(ref1, 10000);
            System.out.println("get val: " + res);
            Assert.assertEquals(res, 1);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java通过函数名调用java 无状态函数
     * @precondition:
     * @step:   1.调用openYuanRong 无状态函数 returnInt
     * @expect: 1.预期有异常抛出
     */
    @Test
    public void test_java_task_invoke_failed() throws Exception {
        TestUtils.initYR(false);
        try {
            boolean isException = false;
            try {
                ObjectRef ref1 = YR
                    .function(JavaFunction.of("org.yuanrong.testutils.TestUtils", "returnInt", int.class))
                    .invoke(1);
                YR.get(ref1, 15);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("invalid function"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                YR.function(JavaFunction.of("org.yuanrong.testutils.TestUtils", "returnInt", int.class))
                    .setUrn("abc123")
                    .invoke(1);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("function urn format is wrong"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                ObjectRef ref1 = YR.function(JavaFunction.of("TestUtils", "returnInt", int.class))
                    .setUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                    .invoke(1);
                YR.get(ref1, 10000);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("ClassNotFoundException"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                ObjectRef ref1 = YR
                    .function(JavaFunction.of("org.yuanrong.testutils.TestUtils", "addOne", int.class))
                    .setUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                    .invoke(1);
                YR.get(ref1, 10000);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("Failed to find user-definded method"),
                    TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                ObjectRef ref1 = YR
                    .function(JavaFunction.of("org.yuanrong.testutils.TestUtils", "returnInt", ArrayList.class))
                    .setUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                    .invoke(1);
                YR.get(ref1, 10000);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("Cannot deserialize value of type"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                ObjectRef ref1 = YR
                    .function(JavaFunction.of("org.yuanrong.testutils.TestUtils", "returnInt", int.class))
                    .setUrn("sn:cn:yrk:default:function:0-yr-stjava:$latest")
                    .invoke("test");
                YR.get(ref1, 10000);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("Cannot deserialize value of type"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            YR.Finalize();
            TestUtils.initYR();
            isException = false;
            try {
                YR.function(JavaFunction.of("org.yuanrong.testutils.TestUtils", "returnInt", int.class))
                    .setUrn("abc123")
                    .invoke(1);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("function urn format is wrong"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java通过函数名调用cpp 无状态函数
     * @precondition:
     * @step:   1.调用openYuanRong 无状态函数 AddTwo
     * @expect: 1.预期无异常抛出
     */
    @Test
    public void test_cpp_task_invoke_successfully() throws Exception {
        TestUtils.initYR(false);
        try {
            ObjectRef ref1 = YR.function(CppFunction.of("AddTwo", int.class))
                .setUrn("sn:cn:yrk:default:function:0-yr-stcpp:$latest")
                .invoke(1);
            int res = (int) YR.get(ref1, 10000);
            System.out.println("get val: " + res);
            Assert.assertEquals(res, 3);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java通过函数名调用cpp 无状态函数
     * @precondition:
     * @step:   1.调用openYuanRong 无状态函数 AddTwo
     * @expect: 1.预期有异常抛出
     */
    @Test
    public void test_cpp_task_invoke_failed() throws Exception {
        TestUtils.initYR(false);
        try {
            boolean isException = false;
            try {
                ObjectRef ref1 = YR.function(CppFunction.of("AddTwo", int.class)).invoke(1);
                YR.get(ref1, 15);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("invalid function"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                YR.function(CppFunction.of("AddTwo", int.class)).setUrn("abc123").invoke(1);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("function urn format is wrong"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                ObjectRef ref1 = YR.function(CppFunction.of("PlusTwo", int.class))
                    .setUrn("sn:cn:yrk:default:function:0-yr-stcpp:$latest")
                    .invoke(1);
                YR.get(ref1, 10000);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("PlusTwo is not found in FunctionHelper"),
                    TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                ObjectRef ref1 = YR.function(CppFunction.of("AddTwo", int.class))
                    .setUrn("sn:cn:yrk:default:function:0-yr-stcpp:$latest")
                    .invoke("test");
                YR.get(ref1, 10000);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("std::bad_cast"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            YR.Finalize();
            TestUtils.initYR();
            isException = false;
            try {
                YR.function(CppFunction.of("AddTwo", int.class))
                    .setUrn("abc123")
                    .invoke(1);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("function urn format is wrong"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);
        } finally {
            YR.Finalize();
        }
    }

    @Ignore("Check whether customextension has been written into the request body by viewing log file.")
    @Test
    public void test_invoke_function_with_customextension() throws Exception {
        TestUtils.initYR();
        try {
            InvokeOptions opts = InvokeOptions.builder()
                .addCustomExtension("endpoint", "InvokeFunction1")
                .addCustomExtension("app_name", "InvokeFunction2")
                .addCustomExtension("tenant_id", "InvokeFunction3")
                .build();
            ObjectRef ref = YR.function(Counter::returnInt).options(opts).invoke(0);
            int res = (int) YR.get(ref, 10);
            Assert.assertEquals(res, 0);
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
            InvokeOptions invokeOpts = InvokeOptions.builder()
                .addScheduleAffinity(affinity)
                .preferredAntiOtherLabels(true)
                .build();
            YR.instance(Counter::returnInt).options(invokeOpts).invoke(1);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: 开启租户上下文，java通过函数名调用java 无状态函数
     * @precondition:
     * @step:   1.调用openYuanRong 无状态函数 returnInt
     * @expect: 1.预期无异常抛出
     * @expect: 2.租户上下文切换成功
     */
    @Test
    public void test_tenant_ctx_task_invoke_successfully() throws Exception {
        String ctx1 = TestUtils.initYRWithTenantCtx(true, false);
        try {
            ObjectRef ref1 = YR.function(TestUtils::returnInt).invoke(1);
            int res = (int) YR.get(ref1, 30);
            System.out.println("get val: " + res);
            Assert.assertEquals(res, 1);

            String ctx2 = TestUtils.initYRWithTenantCtx(true, false);
            try {
                ObjectRef ref2 = YR.function(TestUtils::returnInt).invoke(1);
                int res2 = (int) YR.get(ref2, 30);
                System.out.println("get val: " + res2);
                Assert.assertEquals(res2, 1);

                YR.setContext(ctx1);
                ObjectRef ref3 = YR.function(TestUtils::returnInt).invoke(1);
                int res3 = (int) YR.get(ref3, 30);
                System.out.println("get val: " + res3);
                Assert.assertEquals(res3, 1);
            } finally {
                YR.finalize(ctx2);
            }
        } finally {
            YR.finalize(ctx1);
        }
    }

    /*case
     * @title: 开启租户上下文，各种失败场景
     * @precondition:
     * @step:   1.调用YR.init YR.setContext(ctx) YR.finalize(ctx)
     * @expect: 1.预期异常抛出
     */
    @Test
    public void test_tenant_ctx_task_invoke_failed() throws Exception {
        boolean isException = false;
        try {
            String ctx1 = TestUtils.initYRWithTenantCtx(true, true);
        } catch (YRException e) {
            Assert.assertTrue(e.toString().contains("isThreadLocal and enableSetContext cannot be set to true at the same time"),
                TestUtils.getStackTraceStr(e));
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            YR.setContext("ctx1");
        } catch (YRException e) {
            Assert.assertTrue(e.toString().contains("please init YR first"), TestUtils.getStackTraceStr(e));
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            YR.finalize("ctx1");
        } catch (YRException e) {
            Assert.assertTrue(e.toString().contains("please init YR first"), TestUtils.getStackTraceStr(e));
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    /*case
     * @title: java调用kv接口
     * @precondition:
     * @step:   1.调用KV set接口
     * @step:   2.调用KV get接口
     * @step:   3.调用KV del接口
     * @expect: 1.get成功
     */
    @Test
    public void test_kv_set_and_get() throws Exception {
        TestUtils.initYR(false);
        try {
            String key = "kv-key";
            String data = "kv-value";
            SetParam setParam = new SetParam.Builder().writeMode(WriteMode.NONE_L2_CACHE_EVICT).build();
            YR.kv().set(key, data.getBytes(StandardCharsets.UTF_8), setParam);
            String result = new String(YR.kv().get(key));
            Assert.assertEquals(result, data);
            GetParam param = new GetParam.Builder().offset(0).size(2).build();
            List<GetParam> paramList = new ArrayList<GetParam>(){{add(param);}};
            GetParams params = new GetParams.Builder().getParams(paramList).build();
            result = new String(YR.kv().getWithParam(Arrays.asList("kv-key"), params).get(0));
            Assert.assertEquals(result, "kv");
            YR.kv().del(key);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java调用kv接口
     * @precondition:
     * @step:   1.调用KV mset tx接口
     * @step:   2.调用KV get接口
     * @step:   3,调用KV del接口
     * @expect: 1.get成功
     */
    @Ignore("Check whether shared disk is enabled.")
    @Test
    public void test_kv_mset_tx_and_get() throws Exception {
        TestUtils.initYR(false);
        try {
            MSetParam msetParam = new MSetParam();
            msetParam.setExistence(ExistenceOpt.NX);
            msetParam.setTtlSecond(10);
            msetParam.setCacheType(CacheType.DISK);
            List<String> keys = new ArrayList<String>(){{
                add("synchronous-key1");
                add("synchronous-key2");
                add("synchronous-key3");
            }};
            List<byte[]> vals = new ArrayList<byte[]>(){{
                add("synchronous-value1".getBytes(StandardCharsets.UTF_8));
                add("synchronous-value2".getBytes(StandardCharsets.UTF_8));
                add("synchronous-value3".getBytes(StandardCharsets.UTF_8));
            }};
            YR.kv().mSetTx(keys, vals, msetParam);
            List<byte[]> res = YR.kv().get(keys);
            for (int i = 0; i < res.size(); i++) {
                Assert.assertEquals(new String(res.get(i)), new String(vals.get(i)));
            }
            YR.kv().del(keys);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java调用put get接口
     * @precondition:
     * @step:   1.调用put接口
     * @step:   2.调用get接口
     * @expect: 1.get成功
     */
    @Test
    public void test_put_and_get() throws Exception {
        TestUtils.initYR(false);
        try {
            CreateParam createParam = new CreateParam();
            createParam.setConsistencyType(ConsistencyType.PRAM);
            // Check whether shared disk is enabled.
            ObjectRef ref = YR.put(10, createParam);
            Assert.assertEquals(YR.get(ref, 3000), 10);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: 调用kv set or mSetTx，各种失败场景
     * @precondition:
     * @step:   1.调用kv set or mSetTx，各种失败场景
     * @expect: 1.预期异常抛出
     */
    @Test
    public void test_kv_set_or_msettx_failed() throws Exception {
        TestUtils.initYR(false);
        try {
            boolean isException = false;
            try {
                MSetParam msetParam = new MSetParam();
                msetParam.setExistence(ExistenceOpt.NONE);
                List<String> keys = new ArrayList<String>(){{
                    add("synchronous-key1");
                }};
                List<byte[]> vals = new ArrayList<byte[]>(){{
                    add("synchronous-value1".getBytes(StandardCharsets.UTF_8));
                }};
                YR.kv().mSetTx(keys, vals, msetParam);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("MSetParam's existence should be NX"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                MSetParam msetParam = new MSetParam();
                List<String> keys = new ArrayList<String>(){{
                    add("synchronous-key1");
                    add("synchronous-key2");
                }};
                List<byte[]> vals = new ArrayList<byte[]>(){{
                    add("synchronous-value1".getBytes(StandardCharsets.UTF_8));
                }};
                YR.kv().mSetTx(keys, vals, msetParam);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("Arguments vector size not equal"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);

            isException = false;
            try {
                MSetParam msetParam = new MSetParam();
                List<String> keys = new ArrayList<String>();
                List<byte[]> vals = new ArrayList<byte[]>();
                YR.kv().mSetTx(keys, vals, msetParam);
            } catch (YRException e) {
                Assert.assertTrue(e.toString().contains("The keys should not be empty"), TestUtils.getStackTraceStr(e));
                isException = true;
            }
            Assert.assertTrue(isException);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: java调用kv接口
     * @precondition:
     * @step:   1.调用KV set接口
     * @step:   2.调用KV get接口
     * @expect: 1.get成功
     */
    @Test
    public void test_get_with_pram_partial() throws Exception {
        TestUtils.initYR(false);
        try {
            String key1 = "kv-java-id-testGetWithParamPartial0";
            String key2 = "kv-java-id-testGetWithParamPartial1";
            String value = "kv-value123456";
            GetParam param = new GetParam.Builder().offset(0).size(0).build();
            List<byte[]> res = new ArrayList<>();
            List<String> result = new ArrayList<>();
            List<String> keys = new ArrayList<String>(){{
                add(key1);
                add(key2);
            }};
            List<GetParam> paramList = new ArrayList<GetParam>(){{
                add(param);
                add(param);
            }};
            GetParams params = new GetParams.Builder().getParams(paramList).build();
            SetParam setParam = new SetParam.Builder().writeMode(WriteMode.NONE_L2_CACHE_EVICT).build();
            YR.kv().set(key1, value.getBytes(StandardCharsets.UTF_8), setParam);
            res = YR.kv().getWithParam(keys, params, 4);
            for (int i = 0; i < res.size(); ++i) {
                if (res.get(i) != null) {
                    Assert.assertEquals(new String(res.get(i)), value);
                    result.add(new String(res.get(i)));
                }
            }
            Assert.assertEquals(result.size(), 1);
        } finally {
            YR.Finalize();
        }
    }

    /*case
     * @title: 测试Config中新增的customEnvs参数是否生效
     * @precondition:
     * @step:  1.Config中设置customEnvs
     * @step:  2.发起函数Invoke
     * @expect:  1.customEnvs参数生效
     */
    @Test
    public void test_custom_envs_config() throws Exception {
        Config conf = TestUtils.createConfig(true);
        String key = "LD_LIBRARY_PATH";
        String value = "${LD_LIBRARY_PATH}:${YR_FUNCTION_LIB_PATH}/depend";
        Map<String, String> customEnvs = conf.getCustomEnvs();
        customEnvs.put(key, value);
        YR.init(conf);
        try {
            ObjectRef ref = YR.function(TestUtils::returnCustomEnvs).invoke(key);
            String res = (String) YR.get(ref, 30);
            System.out.println(res);
            Assert.assertTrue(res.contains("depend"));
            Assert.assertFalse(res.contains("YR_FUNCTION_LIB_PATH"));
            Assert.assertFalse(res.contains("LD_LIBRARY_PATH"));
        } finally {
            YR.Finalize();
        }
    }

    @Test
    public void test_env_vars() throws Exception {
        String key = "A";
        String value = "A_VARS";
        InvokeOptions opts = new InvokeOptions();
        HashMap<String, String> envmap = new HashMap<>();
        envmap.put(key, value);
        opts.setEnvVars(envmap);
        TestUtils.initYR();
        try {
            ObjectRef ref = YR.function(TestUtils::returnCustomEnvs).options(opts).invoke(key);
            String res = (String) YR.get(ref, 30);
            Assert.assertEquals(res, value);
        } finally {
            YR.Finalize();
        }
    }
}
