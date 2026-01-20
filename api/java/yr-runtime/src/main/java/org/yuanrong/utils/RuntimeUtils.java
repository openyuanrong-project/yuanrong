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

package org.yuanrong.utils;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;

import java.nio.ByteBuffer;
import java.nio.charset.Charset;
import java.util.ArrayList;
import java.util.List;

/**
 * Runtime utils.
 *
 * @since 2024 /7/21
 */
public class RuntimeUtils {
    private static final Logger LOGGER = LogManager.getLogger(RuntimeUtils.class);

    /**
     * Helper function to convert arglist to string List.
     *
     * @param argList Arg list
     * @return a list of string
     */
    public static List<String> convertArgListToStringList(List<ByteBuffer> argList) {
        List<String> args = new ArrayList<>();
        for (ByteBuffer arg : argList) {
            Charset charset = Charset.forName("UTF-8");
            args.add(charset.decode(arg).toString());
        }
        return args;
    }
}
