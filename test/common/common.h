/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

#pragma once

#include <gmock/gmock.h>
#include <gtest/gtest.h>

#include "api/cpp/include/yr/api/exception.h"

using namespace testing;

#define EXPECT_THROW_WITH_CODE_AND_MSG(statement, code, msg) \
    EXPECT_THROW({                                           \
        try {                                                \
            (statement);                                     \
        } catch (const Exception &e) {                       \
            EXPECT_EQ(e.Code(), (code));                     \
            EXPECT_THAT(e.Msg(), HasSubstr(msg));            \
            throw;                                           \
        }                                                    \
    }, Exception)
