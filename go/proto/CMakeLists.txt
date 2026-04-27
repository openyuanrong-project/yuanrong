# Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

set(GRPC_PROTO_SRCS "${CMAKE_CURRENT_SOURCE_DIR}/posix/common.proto"
        "${CMAKE_CURRENT_SOURCE_DIR}/posix/core_service.proto"
        "${CMAKE_CURRENT_SOURCE_DIR}/posix/runtime_rpc.proto"
        "${CMAKE_CURRENT_SOURCE_DIR}/posix/runtime_service.proto"
        "${CMAKE_CURRENT_SOURCE_DIR}/posix/affinity.proto"
        "${CMAKE_CURRENT_SOURCE_DIR}/posix/inner_service.proto"
        "${CMAKE_CURRENT_SOURCE_DIR}/posix/bus_service.proto"
        "${CMAKE_CURRENT_SOURCE_DIR}/posix/message.proto"
        "${CMAKE_CURRENT_SOURCE_DIR}/posix/resource.proto"
        "${CMAKE_CURRENT_SOURCE_DIR}/posix/bus_adapter.proto")
GENERATE_GRPC_CPP(GRPCPB_SRCS GRPCPB_HDRS ${CMAKE_CURRENT_SOURCE_DIR}/pb/posix
        PROTO_FILES  ${GRPC_PROTO_SRCS}
        SOURCE_ROOT  ${CMAKE_CURRENT_SOURCE_DIR}/posix)
add_library(posix_pb STATIC ${GRPCPB_SRCS})
set_target_properties(posix_pb PROPERTIES UNITY_BUILD ON)
set_target_properties(posix_pb PROPERTIES UNITY_BUILD_BATCH_SIZE 2)
add_dependencies(posix_pb protobuf grpc)
target_link_libraries(
        posix_pb
        PRIVATE
        dl
        ${protobuf_LIB}
        ${grpcpp_LIB}
        ${grpc_LIB}
        ${gpr_LIB})