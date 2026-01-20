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
 * The type GetParam.
 *
 * @since 2025-02-22
 */
@Data
public class GetParam {
    private long offset = 0L;

    private long size = 0L;

    /**
     * Returns the Builder object of GetParam class.
     *
     * @return GetParam Builder class object.
     */
    public static Builder builder() {
        return new Builder();
    }

    /**
     * The Builder class of GetParam class
     */
    public static class Builder {
        private GetParam param;

        /**
         * Instantiates new GetParam.
         */
        public Builder() {
            param = new GetParam();
        }

        /**
         * set the offset
         *
         * @param offset the offset
         * @return GetParam Builder class object.
         */
        public Builder offset(long offset) {
            param.offset = offset;
            return this;
        }

        /**
         * set the size
         *
         * @param size the size
         * @return GetParam Builder class object.
         */
        public Builder size(long size) {
            param.size = size;
            return this;
        }

        /**
         * Builds the GetParam object.
         *
         * @return GetParam class object.
         */
        public GetParam build() {
            return param;
        }
    }
}
