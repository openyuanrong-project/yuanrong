#!/bin/bash
set -e
# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

BASE_DIR=$(
    cd "$(dirname "$0")"
    pwd
)

SO_NAME="faasmanager"

[ -n "$1" ] && SO_NAME=$1

export DATASYSTEM_ADDR=127.0.0.1:31501
export POSIX_LISTEN_ADDR=127.0.0.1:55555
export YR_FUNCTION_LIB_PATH=$(readlink -f $BASE_DIR/../../functionsystem/build/_output/$SO_NAME)
export INIT_HANDLER=$SO_NAME.InitHandler
export CALL_HANDLER=$SO_NAME.CallHandler
export CHECKPOINT_HANDLER=$SO_NAME.CheckpointHandler
export RECOVER_HANDLER=$SO_NAME.RecoverHandler
export SHUTDOWN_HANDLER=$SO_NAME.ShutdownHandler
export SIGNAL_HANDLER=$SO_NAME.SignalHandler

export LD_LIBRARY_PATH=$BASE_DIR/../build/output/runtime/service/go/bin/
$BASE_DIR/../build/output/runtime/service/go/bin/goruntime \
    -runtimeId=12345678 \
    -instanceId=23456789 \
    -logLevel=DEBUG \
    -grpcAddress=127.0.0.1:55555

