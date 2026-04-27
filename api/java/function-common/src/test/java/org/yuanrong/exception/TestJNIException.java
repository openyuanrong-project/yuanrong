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

public class TestJNIException {
    @Test
    public void testInitException() {
        JNIException jniException = new JNIException("test-JNIException");
        JNIException jniException1 = new JNIException(FuncErrorCode.ERR_EXECUTE_REQUEST_ERROR);
        JNIException jniException2 = new JNIException(jniException);

        jniException.setErrorCode(FuncErrorCode.ERR_DATA_SYSTEM_EXCEPTION);
        jniException.getMessage();
        jniException.getErrorCode().getValue();
        jniException1.hashCode();

        Assert.assertTrue(jniException2.equals(jniException2));
        Assert.assertFalse(jniException2.equals(null));
        Assert.assertFalse(jniException2.equals(jniException1));
    }
}
