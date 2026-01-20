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

package org.yuanrong.function;

import org.yuanrong.InvokeOptions;

import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.util.Map;
import java.util.HashMap;

/**
 * Create options for new function instance
 *
 * @since 2024-09-06
 */
@Setter
@Getter
@NoArgsConstructor
public class CreateOptions {
    private int cpu = 0;

    private int memory = 0;
    private Map<String, String> aliasParams = new HashMap<>();

    /**
     * construct method
     *
     * @param memory function instance's memory
     */
    public CreateOptions(int memory) {
        this(0, memory, new HashMap<>());
    }

    /**
     * construct method
     *
     * @param aliasParams alias params
     */
    public CreateOptions(Map<String, String> aliasParams) {
        this(0, 0, aliasParams);
    }

    /**
     * construct method
     *
     * @param cpu function instance's cpu
     * @param memory function instance's memory
     */
    public CreateOptions(int cpu, int memory) {
        this(cpu, memory, new HashMap<>());
    }

    /**
     * construct method
     *
     * @param cpu function instance's cpu
     * @param memory function instance's memory
     * @param aliasParams function alias params
     */
    public CreateOptions(int cpu, int memory, Map<String, String> aliasParams) {
        this.cpu = cpu;
        this.memory = memory;
        this.aliasParams = aliasParams;
    }

    /**
     * convertInvokeOptions convert to invoke options
     *
     * @return InvokeOptions
     */
    public InvokeOptions convertInvokeOptions() {
        InvokeOptions yrOptions = new InvokeOptions();
        yrOptions.setCpu(this.cpu);
        yrOptions.setMemory(this.memory);
        yrOptions.setAliasParams(this.aliasParams);
        return yrOptions;
    }
}
