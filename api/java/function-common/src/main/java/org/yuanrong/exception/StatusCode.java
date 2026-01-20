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

package org.yuanrong.exception;

import java.util.Set;
import java.util.stream.Collectors;
import java.util.stream.Stream;

/**
 * The type StatusCode of datasystem.
 *
 * @since 2023 /2/9
 */
public class StatusCode {
    /**
     * The constant K_OK.
     */
    public static final int K_OK = 0;

    /**
     * The constant K_DUPLICATED.
     */
    public static final int K_DUPLICATED = 1;

    /**
     * The constant K_INVALID.
     */
    public static final int K_INVALID = 2;

    /**
     * The constant K_NOT_FOUND.
     */
    public static final int K_NOT_FOUND = 3;

    /**
     * The constant K_KVSTORE_ERROR.
     */
    public static final int K_KVSTORE_ERROR = 4;

    /**
     * The constant K_RUNTIME_ERROR.
     */
    public static final int K_RUNTIME_ERROR = 5;

    /**
     * The constant K_OUT_OF_MEMORY.
     */
    public static final int K_OUT_OF_MEMORY = 6;

    /**
     * The constant K_IO_ERROR.
     */
    public static final int K_IO_ERROR = 7;

    /**
     * The constant K_NOT_READY.
     */
    public static final int K_NOT_READY = 8;

    /**
     * The constant K_NOT_AUTHORIZED.
     */
    public static final int K_NOT_AUTHORIZED = 9;

    /**
     * The constant K_UNKNOWN_ERROR.
     */
    public static final int K_UNKNOWN_ERROR = 10;

    /**
     * The constant K_INTERRUPTED.
     */
    public static final int K_INTERRUPTED = 11;

    /**
     * The constant K_OUT_OF_RANGE.
     */
    public static final int K_OUT_OF_RANGE = 12;

    /**
     * The constant K_NO_SPACE.
     */
    public static final int K_NO_SPACE = 13;

    /**
     * The constant K_NOT_LEADER_MASTER.
     */
    public static final int K_NOT_LEADER_MASTER = 14;

    /**
     * The constant K_RECOVERY_ERROR.
     */
    public static final int K_RECOVERY_ERROR = 15;

    /**
     * The constant K_RECOVERY_IN_PROGRESS.
     */
    public static final int K_RECOVERY_IN_PROGRESS = 16;

    /**
     * The constant K_FILE_NAME_TOO_LONG.
     */
    public static final int K_FILE_NAME_TOO_LONG = 17;

    /**
     * The constant K_FILE_LIMIT_REACHED.
     */
    public static final int K_FILE_LIMIT_REACHED = 18;

    /**
     * The constant K_TRY_AGAIN.
     */
    public static final int K_TRY_AGAIN = 19;

    /**
     * The constant K_DATA_INCONSISTENCY.
     */
    public static final int K_DATA_INCONSISTENCY = 20;

    /**
     * The constant K_SHUTTING_DOWN.
     */
    public static final int K_SHUTTING_DOWN = 21;

    /**
     * The constant K_WORKER_ABNORMAL.
     */
    public static final int K_WORKER_ABNORMAL = 22;

    /**
     * The constant K_CLIENT_WORKER_DISCONNECT.
     */
    public static final int K_CLIENT_WORKER_DISCONNECT = 23;

    /**
     * The constant K_WORKER_DEADLOCK.
     */
    public static final int K_WORKER_DEADLOCK = 24;


    /**
     * The constant K_RPC_CANCELLED.
     */
    // rpc error code; range: [1000; 2000)
    public static final int K_RPC_CANCELLED = 1000;

    /**
     * The constant K_RPC_DEADLINE_EXCEEDED.
     */
    public static final int K_RPC_DEADLINE_EXCEEDED = 1001;

    /**
     * The constant K_RPC_UNAVAILABLE.
     */
    public static final int K_RPC_UNAVAILABLE = 1002;

    /**
     * The constant K_RPC_STREAM_END.
     */
    public static final int K_RPC_STREAM_END = 1003;

    /**
     * The constant K_OC_ALREADY_SEALED.
     */
    // object error code; range: [2000; 3000)
    public static final int K_OC_ALREADY_SEALED = 2000;

    /**
     * The constant K_OC_OBJECT_NOT_IN_USED.
     */
    public static final int K_OC_OBJECT_NOT_IN_USED = 2001;

    /**
     * The constant K_OC_REMOTE_GET_NOT_ENOUGH.
     */
    public static final int K_OC_REMOTE_GET_NOT_ENOUGH = 2002;

    /**
     * The constant K_SC_STREAM_NOT_FOUND.
     */
    // stream error code; range: [3000; 4000)
    public static final int K_SC_STREAM_NOT_FOUND = 3000;

    /**
     * The constant K_SC_PRODUCER_NOT_FOUND.
     */
    public static final int K_SC_PRODUCER_NOT_FOUND = 3001;

    /**
     * The constant K_SC_CONSUMER_NOT_FOUND.
     */
    public static final int K_SC_CONSUMER_NOT_FOUND = 3002;

    /**
     * The constant K_FC_BUSY.
     */
// file error code; range: [4000; 5000)
    // Delete file error code when open source
    public static final int K_FC_BUSY = 4000;

    /**
     * The constant K_FC_FRAGMENT_ERROR.
     */
    public static final int K_FC_FRAGMENT_ERROR = 4001;

    /**
     * The constant K_FC_NOT_FLUSHED.
     */
    public static final int K_FC_NOT_FLUSHED = 4002;

    /**
     * The constant K_FC_SEVERE_ERROR.
     */
    public static final int K_FC_SEVERE_ERROR = 4003;

    /**
     * The constant K_FC_HARD_LIMIT.
     */
    public static final int K_FC_HARD_LIMIT = 4004;

    /**
     * The constant K_FC_SOFT_LIMIT.
     */
    public static final int K_FC_SOFT_LIMIT = 4005;

    /**
     * The constant K_FC_UPDATE_NEEDED.
     */
    public static final int K_FC_UPDATE_NEEDED = 4006;

    /**
     * The constant K_FC_FILE_CLOSED.
     */
    public static final int K_FC_FILE_CLOSED = 4007;

    /**
     * The constant K_FC_DIRECTORY_NOT_EMPTY.
     */
    public static final int K_FC_DIRECTORY_NOT_EMPTY = 4008;

    /**
     * The constant K_FC_FILE_ALREADY_WRITING.
     */
    public static final int K_FC_FILE_ALREADY_WRITING = 4009;

    /**
     * The constant K_FC_FAIL_QUORUM_WRITE.
     */
    public static final int K_FC_FAIL_QUORUM_WRITE = 4010;

    /**
     * The constant K_FC_FAIL_QUORUM_READ.
     */
    public static final int K_FC_FAIL_QUORUM_READ = 4011;

    /**
     * The constant K_FC_RETRY_LAST_COMMIT.
     */
    public static final int K_FC_RETRY_LAST_COMMIT = 4012;

    /**
     * The constant K_FC_CATCHUP_PENDING.
     */
    public static final int K_FC_CATCHUP_PENDING = 4013;

    /**
     * The constant K_FC_FLUSH_PENDING.
     */
    public static final int K_FC_FLUSH_PENDING = 4014;

    /**
     * The constant RETRY_ERR_CDOE_SET.
     */
    public static final Set<Integer> RETRY_ERR_CDOE_SET = Stream.of(K_OUT_OF_MEMORY, K_NOT_FOUND, K_TRY_AGAIN,
        K_RPC_CANCELLED, K_RPC_DEADLINE_EXCEEDED, K_RPC_UNAVAILABLE).collect(Collectors.toSet());
}
