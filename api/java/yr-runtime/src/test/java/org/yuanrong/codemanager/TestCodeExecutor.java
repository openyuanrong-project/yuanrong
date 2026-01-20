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

package org.yuanrong.codemanager;

import static org.mockito.ArgumentMatchers.any;

import org.yuanrong.libruntime.generated.Libruntime;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.api.mockito.PowerMockito;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.core.classloader.annotations.PrepareForTest;
import org.powermock.modules.junit4.PowerMockRunner;

import java.io.IOException;
import java.util.ArrayList;

@RunWith(PowerMockRunner.class)
@PowerMockIgnore("javax.management.*")
public class TestCodeExecutor {
    @Test
    public void testInitCodeExecutor() {

        CodeExecutor codeExecutor = new CodeExecutor();
        boolean isException = false;
        CodeExecutor.loadHandler();

        FunctionMeta meta = FunctionMeta.newBuilder()
            .setClassName("MockClassName")
            .setSignature("MockSignature")
            .setFunctionName("MockMethodName")
            .build();
        try {
            CodeExecutor.execute(meta, Libruntime.InvokeType.CreateInstanceStateless, null);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        CodeExecutor.shutdown(10);

        isException = false;
        try {
            CodeExecutor.dumpInstance("test");
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);

        isException = false;
        try {
            CodeExecutor.loadInstance(null, null);
        } catch (Exception e) {
            isException = true;
        }
        Assert.assertFalse(isException);
    }
}
