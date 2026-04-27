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

package org.yuanrong.executor;

import org.yuanrong.errorcode.ErrorInfo;
import org.yuanrong.errorcode.Pair;
import org.yuanrong.libruntime.generated.Libruntime;
import org.yuanrong.libruntime.generated.Libruntime.FunctionMeta;

import com.fasterxml.jackson.core.JsonProcessingException;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.util.List;

/**
 * The interface HandlerIntf.
 *
 * @since 2024 /7/1
 */
public interface HandlerIntf {
    /**
     * Execute function
     *
     * @param meta function meta
     * @param type invoke type
     * @param args args
     * @return the return value, in ByteBuffer type, may need to release bytebuffer
     *         `buffer.clear()`
     * @throws Exception the Exception
     */
    ReturnType execute(FunctionMeta meta, Libruntime.InvokeType type, List<ByteBuffer> args) throws Exception;

    /**
     * Shutdown the instance gracefully.
     *
     * @param gracePeriodSeconds the time to wait for the instance to shutdown gracefully.
     * @return ErrorInfo, the ErrorInfo of the execution of shutdown function.
     */
    ErrorInfo shutdown(int gracePeriodSeconds);

    /**
     * Serializes the instance of the CodeExecutor class and returns a Pair
     * containing the serialized byte arrays of the instance and the class name.
     *
     * @param instanceID the ID of the instance to be dumped
     * @return a Pair containing the serialized byte arrays of the instance and the
     *         class name
     * @throws JsonProcessingException if there is an error during serialization
     */
    Pair<byte[], byte[]> dumpInstance(String instanceID) throws JsonProcessingException;

    /**
     * Loads an instance of a class from serialized byte arrays.
     *
     * @param  instanceBytes  the serialized byte array representing the instance
     * @param  clzNameBytes   the serialized byte array representing the class name
     * @throws IOException     if there is an error reading the byte arrays
     * @throws ClassNotFoundException if the class specified by the class name is not found
     */
    void loadInstance(byte[] instanceBytes, byte[] clzNameBytes) throws IOException, ClassNotFoundException;

    /**
     * Recover the instance.
     *
     * @return ErrorInfo, the ErrorInfo of the execution of recover function.
     */
    ErrorInfo recover();
}
