/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
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

package org.yuanrong.storage;

import org.yuanrong.runtime.client.ObjectRef;

import java.util.List;

/**
 * The type Wait result.
 *
 * @since 2022 /08/30
 */
public class WaitResult {
    private final List<ObjectRef> ready;
    private final List<ObjectRef> unready;

    /**
     * Instantiates a new Wait result.
     *
     * @param ready the ready
     * @param unready the unready
     */
    public WaitResult(List<ObjectRef> ready, List<ObjectRef> unready) {
        this.ready = ready;
        this.unready = unready;
    }

    /**
     * Gets ready.
     *
     * @return the ready
     */
    public List<ObjectRef> getReady() {
        return ready;
    }

    /**
     * Gets unready.
     *
     * @return the unready
     */
    public List<ObjectRef> getUnready() {
        return unready;
    }
}
