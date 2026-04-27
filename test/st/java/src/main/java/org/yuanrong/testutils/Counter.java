/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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

package org.yuanrong.testutils;

import org.yuanrong.SetParam;
import org.yuanrong.WriteMode;
import org.yuanrong.api.YR;
import org.yuanrong.exception.YRException;

import java.lang.management.ManagementFactory;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;

public class Counter {
    public static final String SHUTDOWN_KEY = "shutdown";
    public static final String SHUTDOWN_VALUE = "shutdownValue";
    public static final String RECOVER_KEY = "recover";
    public static final String RECOVER_VALUE = "recoverValue";
    private int cnt = 0;

    public static Integer returnInt(int num) {
        return num;
    }

    public static Integer returnIntWithException(int num) {
        if (num == 1) {
            throw new RuntimeException("return int function error");
        }
        return num;
    }

    public int addOne() {
        this.cnt += 1;
        return this.cnt;
    }

    public int get() {
        return this.cnt;
    }
    
    public static String hello() {
        return "hello world";
    }

    public void save() throws YRException {
        YR.saveState();
    }

    public boolean load() throws YRException {
        YR.loadState();
        return true;
    }

    public String indexOutOfBoundsTest() throws Exception {
        List<String> list = new ArrayList<>();
        return list.get(0);
    }

    public static long getPid () {
        String processName = ManagementFactory.getRuntimeMXBean().getName();
        long pid = Long.parseLong(processName.split("@")[0]);
        return pid;
    }

    public void yrRecover() throws YRException {
        SetParam setParam = new SetParam.Builder().writeMode(WriteMode.NONE_L2_CACHE_EVICT).build();
        YR.kv().set(RECOVER_KEY, RECOVER_VALUE.getBytes(StandardCharsets.UTF_8), setParam);
        System.err.println("Class Add recover finished");
    }

    public void yrShutdown(int gracePeriodSeconds) throws YRException {
        SetParam setParam = new SetParam.Builder().writeMode(WriteMode.NONE_L2_CACHE_EVICT).build();
        YR.kv().set(SHUTDOWN_KEY, SHUTDOWN_VALUE.getBytes(StandardCharsets.UTF_8), setParam);
        System.err.println("Class Add shutdown finished");
    }
}

