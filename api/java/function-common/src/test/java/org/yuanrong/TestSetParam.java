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

import org.junit.Assert;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.powermock.core.classloader.annotations.PowerMockIgnore;
import org.powermock.modules.junit4.PowerMockRunner;

@RunWith(PowerMockRunner.class)
@PowerMockIgnore("javax.management.*")
public class TestSetParam {
    @Test
    public void testInitSetParam() {
        SetParam setParam = new SetParam.Builder().existence(ExistenceOpt.NX)
            .writeMode(WriteMode.NONE_L2_CACHE)
            .ttlSecond(10)
            .build();
        setParam.getExistence();
        setParam.getTtlSecond();
        setParam.getWriteMode();

        SetParam testSetParam = SetParam.builder().build();
        testSetParam.setExistence(ExistenceOpt.NONE);
        testSetParam.setTtlSecond(15);
        testSetParam.setWriteMode(WriteMode.WRITE_BACK_L2_CACHE);

        Assert.assertNotEquals(setParam.toString(), testSetParam.toString());
        Assert.assertNotEquals(setParam.hashCode(), testSetParam.hashCode());

        setParam.canEqual(testSetParam);
        Assert.assertFalse(setParam.equals(testSetParam));
    }
}
