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

import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.executor.FunctionHandler;
import org.yuanrong.executor.FaaSHandler;
import org.yuanrong.executor.HandlerIntf;
import org.yuanrong.executor.PosixHandler;
import org.yuanrong.executor.ReturnType;
import org.yuanrong.libruntime.generated.Libruntime;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;

import com.fasterxml.jackson.core.JsonProcessingException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.util.List;

/**
 * CodeExecutor
 *
 * @since 2023/10/23
 */
public class CodeExecutor {
    private static final String INIT_HANDLER = "INIT_HANDLER";

    private static final String FUNCTION_HANDLER
            = "org.yuanrong.handler.InitHandler";

    private static final String FAAS_HANDLER = "org.yuanrong.services.handler.FaaSExecutor.faasInitHandler";

    private static final Logger LOG = LoggerFactory.getLogger(CodeExecutor.class);

    private static HandlerIntf handlerIntf;

    /**
     * Determine the execution type based on the INIT_HANDLER environment variable.
     */
    public static void loadHandler() {
        String handler = System.getenv(INIT_HANDLER);
        if (FUNCTION_HANDLER.equals(handler)) {
            handlerIntf = new FunctionHandler();
        } else if (FAAS_HANDLER.equals(handler)) {
            handlerIntf = new FaaSHandler();
        } else {
            handlerIntf = new PosixHandler();
        }
    }

    /**
     * Execute function
     *
     * @param meta function meta
     * @param type invoke type
     * @param args args
     * @return the return value, in ByteBuffer type, may need to release bytebuffer
     * `buffer.clear()`
     * @throws Exception the Exception
     */
    public static ReturnType execute(FunctionMeta meta, Libruntime.InvokeType type, List<ByteBuffer> args)
        throws Exception {
        return handlerIntf.execute(meta, type, args);
    }

    /**
     * Shutdown the instance gracefully.
     *
     * @param gracePeriodSeconds the time to wait for the instance to shutdown gracefully.
     * @return ErrorInfo, the ErrorInfo of the execution of shutdown function.
     */
    public static ErrorInfo shutdown(int gracePeriodSeconds) {
        return handlerIntf.shutdown(gracePeriodSeconds);
    }

    /**
     * Serializes the instance of the CodeExecutor class and returns a Pair
     * containing the serialized byte arrays of the instance and the class name.
     *
     * @param instanceID the ID of the instance to be dumped
     * @return a Pair containing the serialized byte arrays of the instance and the
     * class name
     * @throws JsonProcessingException if there is an error during serialization
     */
    public static Pair<byte[], byte[]> dumpInstance(String instanceID) throws JsonProcessingException {
        return handlerIntf.dumpInstance(instanceID);
    }

    /**
     * Loads an instance of a class from serialized byte arrays.
     *
     * @param instanceBytes the serialized byte array representing the instance
     * @param clzNameBytes the serialized byte array representing the class name
     * @throws IOException if there is an error reading the byte arrays
     * @throws ClassNotFoundException if the class specified by the class name is not found
     */
    public static void loadInstance(byte[] instanceBytes, byte[] clzNameBytes)
        throws IOException, ClassNotFoundException {
        handlerIntf.loadInstance(instanceBytes, clzNameBytes);
    }

    /**
     * Recover the instance.
     *
     * @return ErrorInfo, the ErrorInfo of the execution of recover function.
     */
    public static ErrorInfo recover() {
        return handlerIntf.recover();
    }
}
