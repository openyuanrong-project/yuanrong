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

package org.yuanrong.jobexecutor;

/**
 * The enum type YRJobStatus.
 *
 * @since 2023 /06/06
 */
public enum YRJobStatus {
    /**
     * The status indicates that attached-runtime is successfully created, and the
     * driver code is being executed.
     */
    RUNNING,

    /**
     * The status indicates that the attached-runtime running the driver code exits
     * with exit code 0.
     */
    SUCCEEDED,

    /**
     * The status indicates that the attached-runtime has been stopped by user while
     * the driver code is NOT completely executed.
     */
    STOPPED,

    /**
     * The status indicates that the attached-runtime running the driver code or the
     * function instance of JobExecutor exits abnormally.
     */
    FAILED
}
