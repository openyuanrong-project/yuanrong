# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

cmake_minimum_required(VERSION 3.16.1)
# 项目信息
project(yrapicpp CXX)
#c++ 17
set(CMAKE_CXX_STANDARD 17)
message(STATUS "PROJECT_SOURCE_DIR=${PROJECT_SOURCE_DIR}")

message("引入GoogleTest单元测试框架:")

#查找引入外部库
find_package(Threads)
#管理'第三方'依赖，下载编译gtest
include(ExternalProject)
ExternalProject_Add(googletest
        PREFIX ${CMAKE_CURRENT_SOURCE_DIR}/gtest
        BUILD_IN_SOURCE 1
        URL https://gitee.com/mirrors/googletest/repository/archive/release-1.8.1.zip
        LOG_DOWNLOAD ON
        LOG_CONFIGURE ON
        LOG_BUILD ON
        CMAKE_ARGS
        -DCMAKE_INSTALL_PREFIX=${CMAKE_CURRENT_SOURCE_DIR}/gtest/install
        )
ExternalProject_Get_Property(googletest SOURCE_DIR)

message("SDK DIR: ${SDK_DIR}")
message("SOURCE_DIR: ${SOURCE_DIR}")
set(gtest_INCLUDE "${CMAKE_CURRENT_SOURCE_DIR}/gtest/install/include")
ExternalProject_Get_Property(googletest BINARY_DIR)
set(gtest_LIBRARY "${BINARY_DIR}/${CMAKE_CFG_INTDIR}/googlemock/gtest/${CMAKE_STATIC_LIBRARY_PREFIX}gtest${CMAKE_STATIC_LIBRARY_SUFFIX}")

#设定可执行二进制文件的目录
#SET( EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/bin)
message("gtest_LIBRARY: ${gtest_LIBRARY}")

include_directories(
     ${gtest_INCLUDE}
     ${PROJECT_SOURCE_DIR}/src
     ${SDK_DIR}/include
)
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -g3 -pthread -w")
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -g3 -pthread -w")
if("${SANITIZER}" STREQUAL "address")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fsanitize=address -fsanitize=leak -fno-delete-null-pointer-checks")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fsanitize=address -fsanitize=leak -fno-delete-null-pointer-checks")
endif("${SANITIZER}" STREQUAL "address")

if("${SANITIZER}" STREQUAL "thread")
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -fsanitize=thread")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fsanitize=thread")
endif("${SANITIZER}" STREQUAL "thread")

# 链接库搜索目录
link_directories(
     ${CMAKE_CURRENT_SOURCE_DIR}/gtest/install/lib
     ${CMAKE_CURRENT_SOURCE_DIR}/gtest/install/lib64
     ${SDK_DIR}/lib
)
#搜索所有源文件
file(GLOB_RECURSE TEST_SRC src/*.cpp)
# 指定生成目标
add_executable(yrapicpp ${TEST_SRC})
add_dependencies(yrapicpp googletest)
message("TEST_SRC: ${TEST_SRC}, PROJECT_SOURCE_DIR: ${PROJECT_SOURCE_DIR}")
add_library(user_common_func SHARED src/user_common_func.cpp)
target_link_libraries(user_common_func yr-api)

target_link_libraries(yrapicpp yr-api gtest gmock)
