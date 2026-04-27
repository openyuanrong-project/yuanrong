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

/**
 * An enumeration representing options for write mode.
 * The options are NONE_L2_CACHE, WRITE_THROUGH_L2_CACHE and WRITE_BACK_L2_CACHE.
 *
 * @since 2023 /11/16
 */
public enum WriteMode {
    /**
     * NONE_L2_CACHE indicates that no write to L2 cache.
     */
    NONE_L2_CACHE,

    /**
     * WRITE_THROUGH_L2_CACHE indicates that sync write to L2 cache.
     */
    WRITE_THROUGH_L2_CACHE,

    /**
     * WRITE_BACK_L2_CACHE indicates that async write to L2 cache.
     */
    WRITE_BACK_L2_CACHE,

    /**
     * NONE_L2_CACHE_EVICT indicates evictable objects.
     */
    NONE_L2_CACHE_EVICT,
}
