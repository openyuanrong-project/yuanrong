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

package org.yuanrong;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.exception.LibRuntimeException;
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

@RunWith(PowerMockRunner.class)
@PrepareForTest( {LibRuntime.class})
@SuppressStaticInitializationFor( {"org.yuanrong.jni.LibRuntime"})
@PowerMockIgnore("javax.management.*")
public class TestGroup {
    @Before
    public void init() throws Exception {
        PowerMockito.mockStatic(LibRuntime.class);
        when(LibRuntime.IsInitialized()).thenReturn(true);
        when(LibRuntime.Init(any())).thenReturn(new ErrorInfo());
    }

    @Test
    public void testInitGroup() throws LibRuntimeException {
        when(LibRuntime.GroupCreate(anyString(), any())).thenReturn(new ErrorInfo());
        when(LibRuntime.GroupTerminate(anyString())).thenReturn(new ErrorInfo());
        when(LibRuntime.GroupWait(anyString())).thenReturn(new ErrorInfo());

        Group group = new Group();
        group.setGroupName("group");
        group.setGroupOpts(new GroupOptions());
        group.getGroupName();
        group.getGroupOpts();
        group.toString();

        GroupOptions groupOptions = new GroupOptions(10);
        Group testGroup = new Group("test-group", groupOptions);
        testGroup.canEqual(group);
        testGroup.hashCode();
        Assert.assertFalse(testGroup.equals(group));

        boolean isException = false;

        try {
            group.invoke();
        } catch (Throwable e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            group.terminate();
        } catch (Throwable e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            group.Wait();
        } catch (Throwable e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }
}
