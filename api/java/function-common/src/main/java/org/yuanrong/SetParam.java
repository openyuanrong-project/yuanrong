/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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
 * The type SetParam.
 *
 * @since 2023-12-04
 */
@Data
public class SetParam {
    private ExistenceOpt existence = ExistenceOpt.NONE;

    private WriteMode writeMode = WriteMode.NONE_L2_CACHE;

    private int ttlSecond = 0;

    private CacheType cacheType = CacheType.MEMORY;

    /**
     * Returns the Builder object of SetParam class.
     *
     * @return SetParam Builder class object.
     */
    public static Builder builder() {
        return new Builder();
    }

    /**
     * The Builder class of SetParam class
     */
    public static class Builder {
        private SetParam setParam;

        /**
         * Instantiates new SetParam.
         */
        public Builder() {
            setParam = new SetParam();
        }

        /**
         * set the existence
         *
         * @param existence the existence
         * @return SetParam Builder class object.
         */
        public Builder existence(ExistenceOpt existence) {
            setParam.existence = existence;
            return this;
        }

        /**
         * set the writeMode
         *
         * @param writeMode the writeMode
         * @return SetParam Builder class object.
         */
        public Builder writeMode(WriteMode writeMode) {
            setParam.writeMode = writeMode;
            return this;
        }

        /**
         * set the ttlSecond
         *
         * @param ttlSecond the ttlSecond
         * @return SetParam Builder class object.
         */
        public Builder ttlSecond(int ttlSecond) {
            setParam.ttlSecond = ttlSecond;
            return this;
        }

        /**
         * set the cacheType
         *
         * @param cacheType the cacheType
         * @return SetParam Builder class object.
         */
        public Builder cacheType(CacheType cacheType) {
            setParam.cacheType = cacheType;
            return this;
        }

        /**
         * Builds the SetParam object.
         *
         * @return SetParam class object.
         */
        public SetParam build() {
            return setParam;
        }
    }
}
