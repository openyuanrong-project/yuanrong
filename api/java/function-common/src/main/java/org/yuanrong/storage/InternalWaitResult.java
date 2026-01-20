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

package org.yuanrong.storage;

import org.yuanrong.errorcode.ErrorInfo;

import lombok.Data;

import java.util.List;
import java.util.Map;

/**
 * Description: inter wait result
 *
 * @since 2023-11-16
 */
@Data
public class InternalWaitResult {
    private List<String> readyIds;
    private List<String> unreadyIds;
    private Map<String, ErrorInfo> exceptionIds;

    public InternalWaitResult(List<String> readyIds, List<String> unreadyIds, Map<String, ErrorInfo> exceptionIds) {
        this.readyIds = readyIds;
        this.unreadyIds = unreadyIds;
        this.exceptionIds = exceptionIds;
    }
}
