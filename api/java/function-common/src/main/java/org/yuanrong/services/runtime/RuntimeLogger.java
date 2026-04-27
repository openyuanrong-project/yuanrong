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

package org.yuanrong.services.runtime;

/**
 * The RuntimeLogger the logger for user to use
 *
 * @since 2024/7/22
 */
public interface RuntimeLogger {
    /**
     * log
     *
     * @param msg log info
     */
    void log(String msg);

    /**
     * info
     *
     * @param msg log info
     */
    void info(String msg);

    /**
     * debug
     *
     * @param msg log info
     */
    void debug(String msg);

    /**
     * warn
     *
     * @param msg log info
     */
    void warn(String msg);

    /**
     * error
     *
     * @param msg log info
     */
    void error(String msg);

    /**
     * Set log level for function stage Logs
     *
     * @param level log level.
     */
    void setLevel(String level);
}
