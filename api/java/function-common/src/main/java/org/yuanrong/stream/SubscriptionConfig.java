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

package org.yuanrong.stream;

import lombok.Data;
import lombok.experimental.Accessors;

import java.util.HashMap;
import java.util.Map;

/**
 * Consumer subscription configuration class
 *
 * @since 2024/09/04
 */
@Data
@Accessors(chain = true)
public class SubscriptionConfig {
    /**
     * Subscription name
     */
    private String subscriptionName = "";

    /**
     * Subscription types include ``STREAM``, ``ROUND_ROBIN``, and ``KEY_PARTITIONS``. Currently, only the ``STREAM``
     * type is supported. Other types are not supported for the time being. The default subscription type is ``STREAM``.
     */
    private SubscriptionType subscriptionType = SubscriptionType.STREAM;

    /**
     * Indicates extended configuration, reserved field.
     */
    private Map<String, String> extendConfig = new HashMap<>();

    /**
     * The SubscriptionConfig class Builder.
     *
     * @return Builder object of SubscriptionConfig class.
     */
    public static Builder builder() {
        return new Builder();
    }

    /**
     * The Builder class of SubscriptionConfig class.
     */
    public static class Builder {
        private SubscriptionConfig subscriptionConfig;

        /**
         * The SubscriptionConfig Builder.
         */
        public Builder() {
            subscriptionConfig = new SubscriptionConfig();
        }

        /**
         * Sets the subscriptionName of SubscriptionConfig class.
         *
         * @param subscriptionName the subscriptionName.
         * @return SubscriptionConfig Builder class object.
         */
        public Builder subscriptionName(String subscriptionName) {
            subscriptionConfig.subscriptionName = subscriptionName;
            return this;
        }

        /**
         * Sets the subscriptionName of SubscriptionConfig class.
         *
         * @param subscriptionType the subscriptionType.
         * @return SubscriptionConfig Builder class object.
         */
        public Builder subscriptionType(SubscriptionType subscriptionType) {
            subscriptionConfig.subscriptionType = subscriptionType;
            return this;
        }

        /**
         * Add the extendConfig of SubscriptionConfig class.
         *
         * @param key the extendConfig key.
         * @param val the extendConfig value.
         * @return SubscriptionConfig Builder class object.
         */
        public Builder addExtendConfig(String key, String val) {
            subscriptionConfig.extendConfig.put(key, val);
            return this;
        }

        /**
         * Sets the extendConfig of SubscriptionConfig class.
         *
         * @param extendConfig the extendConfig map.
         * @return SubscriptionConfig Builder class object.
         */
        public Builder extendConfig(Map<String, String> extendConfig) {
            subscriptionConfig.extendConfig.putAll(extendConfig);
            return this;
        }

        /**
         * Builds the SubscriptionConfig object.
         *
         * @return SubscriptionConfig class object.
         */
        public SubscriptionConfig build() {
            return subscriptionConfig;
        }
    }
}
