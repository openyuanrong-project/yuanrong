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

package org.yuanrong.api;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyList;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

import org.yuanrong.FunctionWrapper;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.exception.YRException;
import org.yuanrong.exception.handler.traceback.StackTraceUtils;
import org.yuanrong.jni.LibRuntime;
import org.yuanrong.jobexecutor.OBSoptions;
import org.yuanrong.jobexecutor.RuntimeEnv;
import org.yuanrong.jobexecutor.YRJobInfo;
import org.yuanrong.jobexecutor.YRJobParam;
import org.yuanrong.runtime.ClusterModeRuntime;
import org.yuanrong.storage.InternalWaitResult;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.core.classloader.annotations.SuppressStaticInitializationFor;
import org.powermock.modules.junit4.PowerMockRunner;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.Optional;

@RunWith(PowerMockRunner.class)
@PrepareForTest( {LibRuntime.class, YR.class, StackTraceUtils.class})
@SuppressStaticInitializationFor( {"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestJobExecutorCaller {
    @Test
    public void testSubmitJob() throws Exception {
        String mockUserJobID = "test-userJobID";
        ArrayList<String> readyIds = new ArrayList<>();
        readyIds.add("test-readId");
        ArrayList<String> unreadyIds = new ArrayList<>();
        unreadyIds.add("test-unreadyId");
        HashMap<String, ErrorInfo> exceptionIds = new HashMap<>();
        exceptionIds.put("test-error",
            new ErrorInfo(ErrorCode.ERR_ETCD_OPERATION_ERROR, ModuleCode.RUNTIME_INVOKE, "msg"));
        InternalWaitResult result = new InternalWaitResult(readyIds, unreadyIds, exceptionIds);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
        Pair<ErrorInfo, String> errorInfoStringPair = new Pair<ErrorInfo, String>(errorInfo, "ok");

        ClusterModeRuntime runtime = PowerMockito.mock(ClusterModeRuntime.class);
        PowerMockito.whenNew(ClusterModeRuntime.class).withAnyArguments().thenReturn(runtime);

        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.CreateInstance(any(), anyList(), any())).thenReturn(errorInfoStringPair);
        when(LibRuntime.GetRealInstanceId(anyString())).thenReturn(mockUserJobID);
        when(LibRuntime.Wait(anyList(), anyInt(), anyInt())).thenReturn(result);

        PowerMockito.mockStatic(StackTraceUtils.class);
        PowerMockito.doNothing()
            .when(StackTraceUtils.class, "checkErrorAndThrowForInvokeException", any(), anyString());

        ArrayList<String> testEntryPoint = new ArrayList<>();
        testEntryPoint.add("test-entryPoint1");
        testEntryPoint.add("test-entryPoint2");
        testEntryPoint.add("test-entryPoint3");
        testEntryPoint.add("test-entryPoint4");

        boolean isException = false;
        YRJobParam yrJobParam = null;
        try {
            yrJobParam = new YRJobParam.Builder().cpu(500)
                .memory(500)
                .entryPoint(testEntryPoint)
                .localCodePath("/tmp")
                .jobName("test-job")
                .addScheduleAffinity(null)
                .obsOptions(new OBSoptions())
                .runtimeEnv(new RuntimeEnv())
                .preferredPriority(true)
                .scheduleAffinity(new ArrayList<>())
                .build();
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            JobExecutorCaller jobExecutorCaller = new JobExecutorCaller();
            JobExecutorCaller.submitJob(yrJobParam);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);
    }

    @Test
    public void testStopJob() throws Exception {
        ClusterModeRuntime runtime = PowerMockito.mock(ClusterModeRuntime.class);
        PowerMockito.whenNew(ClusterModeRuntime.class).withAnyArguments().thenReturn(runtime);

        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.IsInitialized()).thenReturn(true);

        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
        Pair<ErrorInfo, String> errorInfoStringPair = new Pair<ErrorInfo, String>(errorInfo, "ok");
        when(LibRuntime.InvokeInstance(any(), anyString(), anyList(), any())).thenReturn(errorInfoStringPair);

        boolean isException = false;

        try {
            JobExecutorCaller.stopJob("test-ID");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testGetYrJobInfo() throws Exception {
        FunctionWrapper function = PowerMockito.mock(FunctionWrapper.class);
        when(function.getReturnType()).thenReturn(Optional.empty());

        ClusterModeRuntime runtime = PowerMockito.mock(ClusterModeRuntime.class);
        PowerMockito.when(runtime.get(any(), anyInt())).thenReturn(new YRJobInfo());
        PowerMockito.when(runtime.getJavaFunction(any())).thenReturn(function);
        PowerMockito.whenNew(ClusterModeRuntime.class).withAnyArguments().thenReturn(runtime);

        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.IsInitialized()).thenReturn(true);
        ErrorInfo errorInfo = new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
        Pair<ErrorInfo, String> errorInfoStringPair = new Pair<ErrorInfo, String>(errorInfo, "ok");
        when(LibRuntime.InvokeInstance(any(), anyString(), anyList(), any())).thenReturn(errorInfoStringPair);

        boolean isException = false;
        try {
            JobExecutorCaller.getYrJobInfo("test-ID");
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            JobExecutorCaller.getJobStatus("test-jobStatus");
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            JobExecutorCaller.listJobs("test-listJob");
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            JobExecutorCaller.listJobs();
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }

    @Test
    public void testDeleteJob() throws Exception {
        boolean isException = false;

        ClusterModeRuntime runtime = PowerMockito.mock(ClusterModeRuntime.class);
        PowerMockito.whenNew(ClusterModeRuntime.class).withAnyArguments().thenReturn(runtime);
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.IsInitialized()).thenReturn(true);

        try {
            JobExecutorCaller.deleteJob("");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertTrue(isException);

        isException = false;
        try {
            JobExecutorCaller.deleteJob("test-JobExecutorCaller");
        } catch (YRException e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }
}
