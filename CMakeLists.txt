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
# only for ide

cmake_minimum_required(VERSION 3.16)
project(runtime)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_VERBOSE_MAKEFILE ON)

# cpp
file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/cpp/src/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/cpp/src/main/*.cpp")
list(APPEND SOURCES ${SOURCE})

# src
file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/src/dto/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/src/libruntime/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/src/libruntime/fsclient/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/src/libruntime/gwclient/http/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/src/libruntime/invokeadaptor/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/src/libruntime/objectstore/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/src/libruntime/statestore/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/src/libruntime/streamstore/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/src/libruntime/utils/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/src/libruntime/utility/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/src/libruntime/utility/logger/*.cpp")
list(APPEND SOURCES ${SOURCE})

# test
file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/test/api/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/test/libruntime/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/test/libruntime/mock/*.cpp")
list(APPEND SOURCES ${SOURCE})

file(GLOB SOURCE "${CMAKE_CURRENT_LIST_DIR}/test/utility/*.cpp")
list(APPEND SOURCES ${SOURCE})

add_library(runtime SHARED ${SOURCES}
        api/java/function-common/src/main/cpp/jni_types.cpp)

# cpp
list(APPEND INCLUDE_PATH ${CMAKE_CURRENT_LIST_DIR}/cpp/include)

# src
list(APPEND INCLUDE_PATH ${CMAKE_CURRENT_LIST_DIR})

target_include_directories(runtime PRIVATE ${INCLUDE_PATH})