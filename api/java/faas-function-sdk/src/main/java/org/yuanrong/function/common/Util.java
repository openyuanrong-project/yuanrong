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

package org.yuanrong.function.common;

import org.yuanrong.function.common.RspErrorCode;
import org.yuanrong.function.runtime.exception.InvokeException;
import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.google.gson.JsonSyntaxException;

import lombok.extern.slf4j.Slf4j;

import org.apache.commons.lang3.StringUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Util API
 *
 * @since 2024-07-10
 */
@Slf4j
public class Util {
    private static Logger LOG = LoggerFactory.getLogger(Util.class);

    private static final String FUNC_NAME_PATTERN_STRING = "^[a-zA-Z]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$";
    private static final Pattern FUNC_NAME_PATTERN = Pattern.compile(FUNC_NAME_PATTERN_STRING);
    private static final int FUNC_NAME_LENGTH_LIMIT = 60;
    private static final String VERSION_PATTERN_STRING =
            "^[a-zA-Z0-9]([a-zA-Z0-9_-]*\\\\.)*[a-zA-Z0-9_-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$";
    private static final Pattern VERSION_PATTERN = Pattern.compile(VERSION_PATTERN_STRING);
    private static final int VERSION_LENGTH_LIMIT = 32;

    private static final String ALIAS_PATTERN_STRING = "^[a-zA-Z]([a-zA-Z0-9_-]*[a-zA-Z0-9])?$";
    private static final Pattern ALIAS_PATTERN = Pattern.compile(ALIAS_PATTERN_STRING);
    private static final int ALIAS_LENGTH_LIMIT = 32;
    private static final String ALIAS_PREFIX = "!";

    private static final Gson GSON = new Gson();

    /**
     * check funcName is valid
     *
     * @param funcName function name
     * @return res[0]: funcNameBase, res[1]: version
     * @throws InvokeException when invalid funcName
     */
    public static String[] checkFuncName(String funcName) {
        if (funcName == null) {
            throw new InvokeException(RspErrorCode.INVALID_PARAMETER.getErrorCode(),
                    "invalid funcName, expect not null");
        }
        String name = "";
        String version = "latest";
        if (StringUtils.contains(funcName, ":")) {
            String[] nameAndVersion = StringUtils.split(funcName, ":");
            if (nameAndVersion.length != 2) {
                throw new InvokeException(RspErrorCode.INVALID_PARAMETER.getErrorCode(),
                        "invalid funcName, not match regular expression");
            }
            name = nameAndVersion[0];
            if (!checkFunctionName(nameAndVersion[0])) {
                throw new InvokeException(RspErrorCode.INVALID_PARAMETER.getErrorCode(),
                        "invalid funcName, not match regular expression");
            }
            version = nameAndVersion[1];
            if (StringUtils.startsWith(version, ALIAS_PREFIX)) {
                if (!checkAlias(version)) {
                    throw new InvokeException(RspErrorCode.INVALID_PARAMETER.getErrorCode(),
                            "invalid funcName, not match regular expression");
                }
            } else {
                if (!checkVersion(version)) {
                    throw new InvokeException(RspErrorCode.INVALID_PARAMETER.getErrorCode(),
                            "invalid funcName, not match regular expression");
                }
            }
        } else {
            name = funcName;
            if (!checkFunctionName(funcName)) {
                throw new InvokeException(RspErrorCode.INVALID_PARAMETER.getErrorCode(),
                        "invalid funcName, not match regular expression");
            }
        }

        return new String[]{name, version};
    }

    /**
     * check payload is valid json string
     *
     * @param payload json string
     * @throws InvokeException when invalid payload
     */
    public static void checkPayload(String payload) {
        if (StringUtils.isBlank(payload)) {
            throw new InvokeException(RspErrorCode.INVALID_PARAMETER.getErrorCode(),
                    "invalid payload, expect not null");
        }
        try {
            GSON.fromJson(payload, JsonObject.class);
        } catch (JsonSyntaxException jsonSyntaxException) {
            log.error("throw JsonSyntaxException", jsonSyntaxException);
            throw new InvokeException(RspErrorCode.INVALID_PARAMETER.getErrorCode(),
                    "invalid payload, invalid json string");
        }
    }

    /**
     * check error and throw
     *
     * @param errorInfo errorInfo
     * @param msg msg
     * @throws InvokeException when not ok errorInfo
     */
    public static void checkErrorAndThrow(ErrorInfo errorInfo, String msg) throws InvokeException {
        if (!ErrorCode.ERR_OK.equals(errorInfo.getErrorCode())) {
            log.error("{} occurs exception {}", msg, errorInfo);
            throw new InvokeException(errorInfo.getErrorCode().getValue(), errorInfo.getErrorMessage());
        }
    }

    /**
     * Check cpu and memory non-negative
     *
     * @param cpu cpu to check
     * @param memory memory to check
     */
    public static void checkDynamicResource(int cpu, int memory) {
        if (cpu < 0 || memory < 0) {
            throw new InvokeException(RspErrorCode.INVALID_PARAMETER.getErrorCode(),
                "Invalid dynamic resource options, not allow negative number, cpu is " + cpu + ".memory is " + memory);
        }
    }

    private static boolean checkFunctionName(String funcName) {
        Matcher matcher = FUNC_NAME_PATTERN.matcher(funcName);
        if (!matcher.matches()) {
            return false;
        }
        return StringUtils.length(funcName) <= FUNC_NAME_LENGTH_LIMIT;
    }

    private static boolean checkVersion(String version) {
        Matcher matcher = VERSION_PATTERN.matcher(version);
        if (!matcher.matches()) {
            return false;
        }
        return StringUtils.length(version) <= VERSION_LENGTH_LIMIT;
    }

    private static boolean checkAlias(String alias) {
        Matcher matcher = ALIAS_PATTERN.matcher(alias);
        if (!matcher.matches()) {
            return false;
        }
        return StringUtils.length(alias) <= ALIAS_LENGTH_LIMIT;
    }
}
