/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2024. All rights reserved.
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

package org.yuanrong.services.logger;

import org.yuanrong.services.runtime.RuntimeLogger;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * UserFunctionLogger
 *
 * @since 2024/7/22
 */
public class UserFunctionLogger implements RuntimeLogger {
    /**
     * userLogger name at log4j2.xml
     */
    private static final Logger logger = LoggerFactory.getLogger(UserFunctionLogger.class);

    private static String logLevel = "INFO";

    /**
     * log with msg
     *
     * @param msg log info
     */
    @Override
    public void log(String msg) {
        switch (this.logLevel) {
            case "DEBUG":
                this.debug(msg);
                break;
            case "INFO":
                this.info(msg);
                break;
            case "WARN":
                this.warn(msg);
                break;
            case "ERROR":
                this.error(msg);
                break;
            default:
                this.info(msg);
        }
    }

    /**
     * info with msg
     *
     * @param msg log info
     */
    @Override
    public void info(String msg) {
        logger.info(msg);
    }

    /**
     * debug with msg
     *
     * @param msg log info
     */
    @Override
    public void debug(String msg) {
        logger.debug(msg);
    }

    /**
     * warm with msg
     *
     * @param msg log info
     */
    @Override
    public void warn(String msg) {
        logger.warn(msg);
    }

    /**
     * error with msg
     *
     * @param msg log info
     */
    @Override
    public void error(String msg) {
        logger.error(msg);
    }

    /**
     * set level
     *
     * @param level log level
     */
    @Override
    public void setLevel(String level) {
        if ("DEBUG".equals(level) || "INFO".equals(level) || "WARN".equals(level) || "ERROR".equals(level)) {
            this.logLevel = level;
        }
    }
}
