/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

package org.yuanrong.stream;

import org.junit.Assert;
import org.junit.Test;

public class TestProducerConfig {
    @Test
    public void testInitProducerConfig() {
        ProducerConfig.Builder builder = ProducerConfig.builder();
        ProducerConfig.Builder builder1 = new ProducerConfig.Builder();
        builder1.delayFlushTimeMs(1L);
        builder1.pageSizeByte(2L);
        builder1.maxStreamSize(3L);
        builder1.autoCleanup(true);
        builder1.encryptStream(false);
        builder1.reserveSize(4L);
        ProducerConfig build = builder1.build();
        Assert.assertEquals(1L, build.getDelayFlushTimeMs());
    }
}
