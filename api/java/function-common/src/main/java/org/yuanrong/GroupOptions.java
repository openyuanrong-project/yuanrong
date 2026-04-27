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

package org.yuanrong;

import lombok.Data;

/**
 * GroupOptions
 *
 * @since 2024/4/25
 */
@Data
public class GroupOptions {
    private int timeout;

    private boolean sameLifecycle = true;

    /**
     * Init GroupOptions
     */
    public GroupOptions() {}

    /**
     * Init GroupOptions with timeout
     *
     * @param timeout the timeout
     */
    public GroupOptions(int timeout) {
        this(timeout, true);
    }

    /**
     * Init GroupOptions with timeout and sameLifecycle
     *
     * @param timeout the timeout
     * @param sameLifecycle the sameLifecycle
     */
    public GroupOptions(int timeout, boolean sameLifecycle) {
        this.timeout = timeout;
        this.sameLifecycle = sameLifecycle;
    }
}
