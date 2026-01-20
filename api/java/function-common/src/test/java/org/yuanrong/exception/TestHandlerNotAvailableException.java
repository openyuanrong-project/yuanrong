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

package org.yuanrong.exception;

import org.junit.Assert;
import org.junit.Test;

public class TestHandlerNotAvailableException {
    @Test
    public void testInitException() {
        HandlerNotAvailableException handlerNotAvailableException = new HandlerNotAvailableException(null);
        HandlerNotAvailableException handlerNotAvailableException1 = new HandlerNotAvailableException(
            new Exception("test-exception", new Throwable("test-cause")));
        handlerNotAvailableException.getMessage();
        handlerNotAvailableException.hashCode();
        HandlerNotAvailableException handlerNotAvailableException2 = new HandlerNotAvailableException(
            new Exception("test2-exception"));

        Assert.assertTrue(handlerNotAvailableException.equals(handlerNotAvailableException));
        Assert.assertFalse(handlerNotAvailableException1.equals(null));
        Assert.assertFalse(handlerNotAvailableException.equals(handlerNotAvailableException2));
    }
}
