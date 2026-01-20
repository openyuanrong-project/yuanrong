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

package org.yuanrong.codemanager;

import org.yuanrong.errorcode.ErrorCode;
import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.ModuleCode;
import org.yuanrong.runtime.util.ExtClasspathLoader;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.util.List;

/**
 * CodeLoader
 *
 * @since 2023/10/23
 */
public class CodeLoader {
    private static final Logger LOG = LoggerFactory.getLogger(CodeLoader.class);

    static ErrorInfo Load(List<String> codePaths) {
        LOG.info("CodeLoader is running");
        try {
            ExtClasspathLoader.loadClasspath(codePaths);
        } catch (InvocationTargetException | IllegalAccessException | IOException e) {
            String errorMsg = "failed to load code in specified path (" + codePaths + ") due to exception (" + e + ")";
            LOG.error(errorMsg);
            return new ErrorInfo(ErrorCode.ERR_PARAM_INVALID, ModuleCode.RUNTIME, errorMsg);
        }
        return new ErrorInfo(ErrorCode.ERR_OK, ModuleCode.RUNTIME, "");
    }
}