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

package org.yuanrong.runtime.client;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyBoolean;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.ArgumentMatchers.anyString;

import org.yuanrong.Config;
import org.yuanrong.ExistenceOpt;
import org.yuanrong.GetParam;
import org.yuanrong.GetParams;
import org.yuanrong.MSetParam;
import org.yuanrong.SetParam;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.YRException;
import org.yuanrong.jni.LibRuntime;

import org.junit.Assert;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.core.classloader.annotations.SuppressStaticInitializationFor;
import org.powermock.modules.junit4.PowerMockRunner;

import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

@RunWith(PowerMockRunner.class)
@PrepareForTest({LibRuntime.class})
@SuppressStaticInitializationFor({"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestKVManager {
    Config conf = new Config(
        "sn:cn:yrk:default:function:0-crossyrlib-helloworld:$latest",
        "127.0.0.0",
        "127.0.0.0",
        "sn:cn:yrk:default:function:0-test-hello:$latest",
        true);

    @Before
    public void init() throws Exception {
    }

    @Test
    public void testSet() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.when(LibRuntime.class, "IsInitialized").thenReturn(true);
        PowerMockito.when(LibRuntime.class, "Init", any()).thenReturn(new ErrorInfo());
        KVManager kvManager = new KVManager();
        PowerMockito.when(LibRuntime.class, "KVWrite", anyString(), any(), any()).thenReturn(new ErrorInfo());
        kvManager.set("testKey", "test".getBytes(StandardCharsets.UTF_8));
        kvManager.set("testKey", "test".getBytes(StandardCharsets.UTF_8), ExistenceOpt.NONE);
        kvManager.set("test", "testKey".getBytes(StandardCharsets.UTF_8), 100, ExistenceOpt.NONE);
        kvManager.set("test", "test".getBytes(StandardCharsets.UTF_8), new SetParam());
        kvManager.set("test", "test".getBytes(StandardCharsets.UTF_8), 100, new SetParam());
        kvManager.set("testKey", "test".getBytes(StandardCharsets.UTF_8), new KVCallback() {
            @Override
            public void onComplete() {
            }
        });
    }

    @Test
    public void testSetException() {
        KVManager kvManager = new KVManager();
        boolean isException = false;
        SetParam setParam = new SetParam.Builder().ttlSecond(-1).build();
        try {
            kvManager.set("test", "test".getBytes(StandardCharsets.UTF_8), setParam);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.set("test", "t".getBytes(StandardCharsets.UTF_8), 100000, ExistenceOpt.NONE);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.set("test", null, 10, ExistenceOpt.NONE);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.set("test", null, 10, new SetParam());
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.set("test", "t".getBytes(StandardCharsets.UTF_8), 100000, new SetParam());
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.set("test", "test".getBytes(StandardCharsets.UTF_8), 10, setParam);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.set("test", "test".getBytes(StandardCharsets.UTF_8));
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testMSetTx() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.when(LibRuntime.class, "IsInitialized").thenReturn(true);
        PowerMockito.when(LibRuntime.class, "Init", any()).thenReturn(new ErrorInfo());
        KVManager kvManager = new KVManager();
        PowerMockito.when(LibRuntime.class, "KVMSetTx", any(), any(), any()).thenReturn(new ErrorInfo());
        MSetParam mSetParam = new MSetParam();
        List<String> keys = new ArrayList<String>(){{
            add("test1");
            add("test2");
        }};
        List<byte[]> vals = new ArrayList<byte[]>(){{
            add("test".getBytes(StandardCharsets.UTF_8));
            add("test".getBytes(StandardCharsets.UTF_8));
        }};
        List<Integer> lengths = Arrays.asList(5, 5);
        boolean isException = false;
        try {
            kvManager.mSetTx(keys, vals, mSetParam);
            kvManager.mSetTx(keys, vals, lengths, mSetParam);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testMSetTxException() {
        KVManager kvManager = new KVManager();
        List<String> keys1 = new ArrayList<String>(){{
            add("test1");
            add("test2");
        }};
        List<String> keys2 = new ArrayList<String>(){{
            add("test1");
        }};
        List<byte[]> vals = new ArrayList<byte[]>(){{
            add("test".getBytes(StandardCharsets.UTF_8));
            add("test".getBytes(StandardCharsets.UTF_8));
        }};
        List<byte[]> vals2 = new ArrayList<byte[]>(){{
            add(null);
        }};
        List<Integer> lengths = Arrays.asList(5, 5);
        boolean isException = false;
        MSetParam abnormalParam = new MSetParam.Builder().ttlSecond(-1).build();
        MSetParam abnormalParam2 = new MSetParam.Builder().existence(ExistenceOpt.NONE).build();
        MSetParam okParam = new MSetParam();
        try {
            kvManager.mSetTx(keys1, vals, abnormalParam);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.mSetTx(keys1, null, okParam);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.mSetTx(null, vals, okParam);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.mSetTx(null, null, okParam);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.mSetTx(keys2, vals, okParam);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.mSetTx(keys1, vals, abnormalParam2);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        List<Integer> abnormalLengths = Arrays.asList(100000, 100000);
        try {
            kvManager.mSetTx(keys1, vals, abnormalLengths, okParam);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            kvManager.mSetTx(keys2, vals2, okParam);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testWrite() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.when(LibRuntime.class, "IsInitialized").thenReturn(true);
        PowerMockito.when(LibRuntime.class, "Init", any()).thenReturn(new ErrorInfo());
        PowerMockito.when(LibRuntime.class, "KVWrite", anyString(), any(), any()).thenReturn(new ErrorInfo());
        KVManager kvManager = new KVManager();
        boolean isException = false;
        try {
            kvManager.write("test", "test", ExistenceOpt.NONE);
            kvManager.write("test", "test", new SetParam());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testMWriteTx() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.when(LibRuntime.class, "IsInitialized").thenReturn(true);
        PowerMockito.when(LibRuntime.class, "Init", any()).thenReturn(new ErrorInfo());
        PowerMockito.when(LibRuntime.class, "KVMSetTx", any(), any(), any()).thenReturn(new ErrorInfo());
        KVManager kvManager = new KVManager();
        List<String> keys = new ArrayList<String>(){{
            add("test1");
            add("test2");
        }};
        List<Object> vals = new ArrayList<Object>(){{
            add("test1");
            add("test2");
        }};
        boolean isException = false;
        try {
            kvManager.mWriteTx(keys, vals, new MSetParam());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testGet() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.when(LibRuntime.class, "IsInitialized").thenReturn(true);
        PowerMockito.when(LibRuntime.class, "Init", any()).thenReturn(new ErrorInfo());
        PowerMockito.when(LibRuntime.class, "KVRead", anyString(), anyInt())
            .thenReturn(new Pair<>("test".getBytes(StandardCharsets.UTF_8), new ErrorInfo()));
        PowerMockito.when(LibRuntime.class, "KVRead", anyList(), anyInt(), anyBoolean())
            .thenReturn(new Pair<>(new ArrayList<>(), new ErrorInfo()));
        KVManager kvManager = new KVManager();
        kvManager.get("test");
        kvManager.get("test", 10);
        ArrayList<String> strings = new ArrayList<>();
        strings.add("test");
        kvManager.get(strings, 10, false);
        kvManager.get(strings, false);
        kvManager.get(strings);
    }

    @Test
    public void testGetException() {
        KVManager kvManager = new KVManager();
        boolean isException = false;
        try {
            kvManager.get(new ArrayList<>());
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testRead() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.when(LibRuntime.class, "IsInitialized").thenReturn(true);
        PowerMockito.when(LibRuntime.class, "Init", any()).thenReturn(new ErrorInfo());
        KVManager kvManager = new KVManager();
        PowerMockito.when(LibRuntime.class, "KVRead", anyString(), anyInt())
            .thenReturn(new Pair<>("test".getBytes(StandardCharsets.UTF_8), new ErrorInfo()));
        ArrayList<byte[]> list = new ArrayList<>();
        list.add("test".getBytes(StandardCharsets.UTF_8));
        PowerMockito.when(LibRuntime.class, "KVRead", anyList(), anyInt(), anyBoolean())
            .thenReturn(new Pair<>(list, new ErrorInfo()));
        kvManager.read("test", 10, String.class);
        ArrayList<String> arrayList = new ArrayList<>();
        arrayList.add("test");
        ArrayList<Class<?>> classes = new ArrayList<>();
        classes.add(String.class);
        kvManager.read(arrayList, 10, classes, false);
    }

    @Test
    public void testGetWithParam() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.when(LibRuntime.class, "IsInitialized").thenReturn(true);
        PowerMockito.when(LibRuntime.class, "Init", any()).thenReturn(new ErrorInfo());
        KVManager kvManager = new KVManager();
        ArrayList<byte[]> list = new ArrayList<>();
        list.add("test".getBytes(StandardCharsets.UTF_8));
        PowerMockito.when(LibRuntime.class, "KVGetWithParam", anyList(), any(), anyInt())
            .thenReturn(new Pair<>(list, new ErrorInfo()));
        List<String> keys = new ArrayList<String>(){{add("key1");}};
        GetParam param = new GetParam.Builder().offset(0).size(2).build();
        List<GetParam> paramList = new ArrayList<GetParam>(){{add(param);}};
        GetParams params = new GetParams.Builder().getParams(paramList).build();
        List<GetParam> paramList2 = new ArrayList<GetParam>(){{
            add(param);
            add(param);
        }};
        GetParams params2 = new GetParams.Builder().getParams(paramList2).build();
        GetParams errParams = new GetParams.Builder().build();
        boolean isException = false;
        try {
            kvManager.getWithParam(keys, errParams);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        isException = false;
        try {
            kvManager.getWithParam(keys, params2);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);
        isException = false;
        try {
            kvManager.getWithParam(keys, params);
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testDel() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        PowerMockito.when(LibRuntime.class, "IsInitialized").thenReturn(true);
        PowerMockito.when(LibRuntime.class, "Init", any()).thenReturn(new ErrorInfo());
        KVManager kvManager = new KVManager();
        PowerMockito.when(LibRuntime.class, "KVDel", anyString()).thenReturn(new ErrorInfo());
        ArrayList<String> arrayList = new ArrayList<>();
        arrayList.add("test");
        PowerMockito.when(LibRuntime.class, "KVDel", anyList()).thenReturn(new Pair<>(arrayList, new ErrorInfo()));
        kvManager.del("test");
        kvManager.del("test", new KVCallback() {
            @Override
            public void onComplete() {
            }
        });
        kvManager.del(arrayList);
        kvManager.del(arrayList, new KVCallback() {
            @Override
            public void onComplete() {
            }
        });
    }
}
