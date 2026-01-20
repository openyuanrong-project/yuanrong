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

package org.yuanrong.runtime.referencecount;

/**
 * The type ReferenceCount is created to adapt to the original runtime-java test
 * case and has no actual function.
 *
 * @since 2023 /09/12
 */
public class ReferenceCount {
    /**
     * Instantiates a new Reference count.
     */
    private ReferenceCount() {}

    /**
     * The type Reference count getter.
     */
    public static class ReferenceCountInitiator {
        /**
         * The constant INSTANCE.
         */
        private static final ThreadLocal<ReferenceCount> INSTANCE = ThreadLocal.withInitial(ReferenceCount::new);

        /**
         * Get instance of singleton ReferenceCount.
         *
         * @return instance ReferenceCount instance.
         */
        public static ReferenceCount getInstance() {
            return INSTANCE.get();
        }
    }

    /**
     * Query global ref num by object client int.
     *
     * @param objectId the object id
     * @return the int
     */
    public int queryGlobalRefNumByObjectClient(String objectId) {
            return 1;
        }
}