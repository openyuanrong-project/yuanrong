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

package org.yuanrong.services.model;

import lombok.Data;

import java.util.Map;

/**
 * The type CallResponse.
 *
 * @since 2024-07-20
 */
@Data
public class CallRequest {
    private Map<String, String> header;

    private Object body;

    /**
     * The Builder class of CallRequest class
     */
    public static final class Builder {
        private Map<String, String> header;

        private Object body;

        /**
         * With withHead method
         *
         * @param headers head
         * @return the builder
         */
        public Builder withHead(Map<String, String> header) {
            this.header = header;
            return this;
        }

        /**
         * With withBody method
         *
         * @param body body
         * @return the builder
         */
        public Builder withBody(Object body) {
            this.body = body;
            return this;
        }

        /**
         * build method
         *
         * @return CallRequest
         */
        public CallRequest build() {
            CallRequest req = new CallRequest();
            req.setHeader(header);
            req.setBody(body);
            return req;
        }
    }
}
