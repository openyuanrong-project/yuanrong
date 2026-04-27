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

import java.util.List;
import java.util.ArrayList;

/**
 * The type GetParams.
 *
 * @since 2025-02-22
 */
@Data
public class GetParams {
    private List<GetParam> getParams = new ArrayList<>();

    private String traceId;

    /**
     * Returns the Builder object of GetParams class.
     *
     * @return GetParams Builder class object.
     */
    public static Builder builder() {
        return new Builder();
    }

    /**
     * The Builder class of GetParams class
     */
    public static class Builder {
        private GetParams params;

        /**
         * Instantiates new GetParams.
         */
        public Builder() {
            params = new GetParams();
        }

        /**
         * set the getParams
         *
         * @param getParams the getParams
         * @return GetParams Builder class object.
         */
        public Builder getParams(List<GetParam> getParams) {
            params.getParams = getParams;
            return this;
        }

        /**
         * set the traceId
         *
         * @param traceId the traceId
         * @return GetParams Builder class object.
         */
        public Builder traceId(String traceId) {
            params.traceId = traceId;
            return this;
        }

        /**
         * Builds the GetParams object.
         *
         * @return GetParams class object.
         */
        public GetParams build() {
            return params;
        }
    }
}
