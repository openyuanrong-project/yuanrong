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

package org.yuanrong;

import lombok.Data;

/**
 * The type CreateParam.
 *
 * @since 2025-2-6
 */
@Data
public class CreateParam {
    private ConsistencyType consistencyType = ConsistencyType.PRAM;

    private CacheType cacheType = CacheType.MEMORY;

    /**
     * Returns the Builder object of CreateParam class.
     *
     * @return CreateParam Builder class object.
     */
    public static Builder builder() {
        return new Builder();
    }

    /**
     * The Builder class of CreateParam class
     */
    public static class Builder {
        private CreateParam createParam;

        /**
         * Instantiates new CreateParam.
         */
        public Builder() {
            createParam = new CreateParam();
        }

        /**
         * set the consistencyType
         *
         * @param consistencyType the consistencyType
         * @return CreateParam Builder class object.
         */
        public Builder consistencyType(ConsistencyType consistencyType) {
            createParam.consistencyType = consistencyType;
            return this;
        }

        /**
         * set the cacheType
         *
         * @param cacheType the cacheType
         * @return CreateParam Builder class object.
         */
        public Builder cacheType(CacheType cacheType) {
            createParam.cacheType = cacheType;
            return this;
        }

        /**
         * Builds the CreateParam object.
         *
         * @return CreateParam class object.
         */
        public CreateParam build() {
            return createParam;
        }
    }
}
