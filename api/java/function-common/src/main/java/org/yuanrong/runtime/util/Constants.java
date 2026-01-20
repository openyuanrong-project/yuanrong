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

package org.yuanrong.runtime.util;

import java.util.regex.Pattern;

/**
 * Description: Constants
 *
 * @since 2023-09-19
 */
public final class Constants {
    /**
     * The default value of cpu
     */
    public static final int DEFAULT_CPU = 500;

    /**
     * The default value of memory
     */
    public static final int DEFAULT_MEMORY = 500;

    /**
     * The default timeout in seconds of get data from data system (when input times run out)
     */
    public static final int DEFAULT_GET_DATA_TIMEOUT_SEC = 5;

    /**
     * The default timeout in milliseconds of get data from data system (when input times run out)
     */
    public static final int DEFAULT_GET_DATA_TIMEOUT_MS = 5000;

    /**
     * Indicates a space String " ", typically used as a seperator.
     */
    public static final String SPACE = " ";

    /**
     * Indicates a backslash String "/", typically used as a seperator.
     */
    public static final String BACKSLASH = "/";

    /**
     * Indicates a no_timeout String "-1", typically used as a seperator.
     */
    public static final int NO_TIMEOUT = -1;

    /**
     * Indicates a set_to_ms String "1000", typically used as a seperator.
     */
    public static final int SEC_TO_MS = 1000;

    /**
     * The default timeout in seconds of getting the value stored in the data system.
     */
    public static final int DEFAULT_GET_TIMEOUT_SEC = 300;

    /**
     * default put interval time in milliseconds.
     */
    public static final long DEFAULT_PUT_INTERVAL_TIME = 5000L;

    /**
     * default put retryTime
     */
    public static final long DEFAULT_RETRY_TIME = 1L;

    /**
     * Indicates an empty String "".
     */
    public static final String EMPTY_STRING = "";

    /**
     * io thread count for http client
     */
    public static final int DEFAULT_HTTP_IO_THREAD_CNT = 100;

    /**
     * idle time for http client
     */
    public static final int DEFAULT_HTTP_IDLE_TIME = 120;

    /**
     * The key to python packages, which are going to be installed in remote runtime.
     */
    public static final String POST_START_EXEC = "POST_START_EXEC";

    /**
     * The key of authorization ciphertext in environment.
     */
    public static final String AUTHORIZATION = "AUTHORIZATION";

    /**
     * The Instance key.
     */
    public static final String INSTANCE_KEY = "instanceKey";

    /**
     * The Instance id.
     */
    public static final String INSTANCE_ID = "instanceID";

    /**
     * The Instance route.
     */
    public static final String INSTANCE_ROUTE = "instanceRoute";

    /**
     * The function id.
     */
    public static final String FUNCTION_ID = "functionId";

    /**
     * The Function key.
     */
    public static final String FUNCTION_KEY = "functionKey";

    /**
     * The Class name.
     */
    public static final String CLASS_NAME = "className";

    /**
     * The Module name.
     */
    public static final String MODULE_NAME = "moduleName";

    /**
     * The Language key.
     */
    public static final String LANGUAGE_KEY = "targetLanguage";

    /**
     * The Api type.
     */
    public static final String API_TYPE = "apiType";

    /**
     * Need order.
     */
    public static final String NEED_ORDER = "needOrder";

    /**
     * Concurrency.
     */
    public static final String CONCURRENCY = "Concurrency";

    /**
     * PATTERN_FAAS_ENTRY
     */
    public static final Pattern PATTERN_FAAS_ENTRY = Pattern.compile("^[^/]*.[^/]*$");

    /**
     * The constant KEY_USER_INIT_ENTRY.
     */
    public static final String KEY_USER_INIT_ENTRY = "userInitEntry";

    /**
     * The constant KEY_USER_CALL_ENTRY.
     */
    public static final String KEY_USER_CALL_ENTRY = "userCallEntry";

    /**
     * The constant INDEX_FIRST.
     */
    public static final int INDEX_FIRST = 0;

    /**
     * The constant INDEX_SECOND.
     */
    public static final int INDEX_SECOND = 1;

    /**
     * The constant INDEX_THIRD.
     */
    public static final int INDEX_THIRD = 2;

    /**
     * The constant INDEX_FORTH.
     */
    public static final int INDEX_FORTH = 3;

    /**
     * The constant INDEX_FIFTH.
     */
    public static final int INDEX_FIFTH = 4;

    /**
     * The constant INDEX_SIXTH.
     */
    public static final int INDEX_SIXTH = 5;

    /**
     * LogRequestID
     */
    public static final String LOG_REQUEST_ID = "YRLogRequestID";

    /**
     * The log type
     */
    public static final String CFF_LOG_TYPE = "X-Cff-Log-Type";

    /**
     * The request id
     */
    public static final String REQUEST_ID = "requestID";

    /**
     * The logger type
     */
    public static final String LOGGER_TYPE = "logger";

    /**
     * The std type
     */
    public static final String STD_TYPE = "std";
}