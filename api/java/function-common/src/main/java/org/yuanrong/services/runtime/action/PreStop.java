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

package org.yuanrong.services.runtime.action;

import com.google.gson.annotations.SerializedName;

import lombok.Data;

/**
 * The type pre_stop
 *
 * @since 2025.05.28
 */
@Data
public class PreStop {
    @SerializedName("pre_stop_handler")
    private String preStopHandler;

    @SerializedName("pre_stop_timeout")
    private int preStopTimeout;
}
