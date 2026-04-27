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
 * The type MSetParam.
 *
 * @since 2025-1-24
 */
@Data
public class MSetParam {
    private ExistenceOpt existence = ExistenceOpt.NX;

    private WriteMode writeMode = WriteMode.NONE_L2_CACHE;

    private int ttlSecond = 0;

    private CacheType cacheType = CacheType.MEMORY;

    /**
     * Returns the Builder object of MSetParam class.
     *
     * @return MSetParam Builder class object.
     */
    public static Builder builder() {
        return new Builder();
    }

    /**
     * The Builder class of MSetParam class
     */
    public static class Builder {
        private MSetParam mSetParam;

        /**
         * Instantiates new MSetParam.
         */
        public Builder() {
            mSetParam = new MSetParam();
        }

        /**
         * set the existence
         *
         * @param existence the existence
         * @return MSetParam Builder class object.
         */
        public Builder existence(ExistenceOpt existence) {
            mSetParam.existence = existence;
            return this;
        }

        /**
         * set the writeMode
         *
         * @param writeMode the writeMode
         * @return MSetParam Builder class object.
         */
        public Builder writeMode(WriteMode writeMode) {
            mSetParam.writeMode = writeMode;
            return this;
        }

        /**
         * set the ttlSecond
         *
         * @param ttlSecond the ttlSecond
         * @return MSetParam Builder class object.
         */
        public Builder ttlSecond(int ttlSecond) {
            mSetParam.ttlSecond = ttlSecond;
            return this;
        }

        /**
         * set the cacheType
         *
         * @param cacheType the cacheType
         * @return MSetParam Builder class object.
         */
        public Builder cacheType(CacheType cacheType) {
            mSetParam.cacheType = cacheType;
            return this;
        }

        /**
         * Builds the MSetParam object.
         *
         * @return MSetParam class object.
         */
        public MSetParam build() {
            return mSetParam;
        }
    }
}
