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

package org.yuanrong.executor;

/**
 * The type Request event test.
 *
 * @since 2024.07.01
 */
public class TestRequestEvent {
    /**
     * The Name.
     */
    private String name;

    /**
     * The Level.
     */
    private int level;

    /**
     * Instantiates a new Request event test.
     *
     * @param name  the name
     * @param level the level
     */
    public TestRequestEvent(String name, int level) {
        this.name = name;
        this.level = level;
    }

    public int getLevel() {
        return this.level;
    }
}
