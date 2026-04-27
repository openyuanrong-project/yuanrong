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
#include <stdlib.h>
#include <string>

#include "base/utils.h"
#include "gmock/gmock.h"
#include "gtest/gtest.h"
#include "user_common_func.h"
#include "yr/yr.h"

class DsTest : public testing::Test {
public:
   DsTest(){};
   ~DsTest(){};
   static void SetUpTestCase(){};
   static void TearDownTestCase(){};
   void SetUp()
   {
       YR::Config config;
       config.mode = YR::Config::Mode::CLUSTER_MODE;
       auto info = YR::Init(config);
       std::cout << "job id: " << info.jobId << std::endl;
   };
   void TearDown()
   {
       YR::Finalize();
   };
};

TEST_F(DsTest, KVWithTenantId)
{
   // set with valid tenantId
   YR::SetParam setParam;
   setParam.traceId = "executor-0";
   auto key = "cppkey1";
   auto value = "value1";
   YR::KV().Set(key, value, setParam);
   auto get = YR::KV().Get(key);
   ASSERT_EQ(value, get);

   // del with invalid tenantId
   YR::DelParam delParam;
   delParam.traceId = "1234$5678";
   ASSERT_THROW(YR::KV().Del(key, delParam), YR::Exception);

   // del with valid tenantId
   delParam.traceId = "executor-1";
   ASSERT_NO_THROW(YR::KV().Del(key, delParam));

   // set with invalid tenantId
   setParam.traceId = "1234$5678";
   ASSERT_THROW(YR::KV().Set(key, value, setParam), YR::Exception);
}
