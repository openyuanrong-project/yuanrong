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

package org.yuanrong.services.logger;

import org.junit.Test;

public class TestUserFunctionLogger {
    private UserFunctionLogger logger = new UserFunctionLogger();

    @Test
    public void testLog() {
        logger.log("no args");
    }

    @Test
    public void testInfo() {
        logger.info("no args");
    }

    @Test
    public void testDebug() {
        logger.debug("no args");
    }

    @Test
    public void testWarn() {
        logger.warn("no args");
    }

    @Test
    public void testError() {
        logger.error("no args");
    }
}
