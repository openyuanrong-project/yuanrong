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
project(custom_test CXX)
#c++ 17
set(CMAKE_CXX_STANDARD 17)
message(STATUS "PROJECT_SOURCE_DIR=${PROJECT_SOURCE_DIR}")


#查找引入外部库
find_package(Threads)

message("SDK DIR: ${SDK_DIR}")
message("SOURCE_DIR: ${SOURCE_DIR}")

#设定可执行二进制文件的目录
#SET( EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/bin)

include_directories(
     ${PROJECT_SOURCE_DIR}/src
     ${SDK_DIR}/include
)
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -g3 -pthread -w")
set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -g3 -pthread -w")
# 链接库搜索目录
link_directories(
     ${SDK_DIR}/lib
)
#搜索所有源文件
file(GLOB_RECURSE TEST_SRC src/*.cpp)
# 指定生成目标
add_executable(main ${TEST_SRC})
message("TEST_SRC: ${TEST_SRC}, PROJECT_SOURCE_DIR: ${PROJECT_SOURCE_DIR}")
add_library(user_common_func SHARED src/main.cpp)
target_link_libraries(user_common_func yr-api)

target_link_libraries(main yr-api)
