/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
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

#include "yr/yr.h"

int main(void)
{
    {
        //! [Init localMode]
        // local mode
        YR::Config conf;
        conf.mode = YR::Config::Mode::LOCAL_MODE;
        config.threadPoolSize = 10;
        YR::Init(conf);
        //! [Init localMode]
    }
    {
        //! [Init clusterMode]
        // cluster mode
        YR::Config conf;
        conf.mode = YR::Config::Mode::CLUSTER_MODE;
        conf.functionID = "sn:cn:yrk:xxxxx:function:0-mysvc-myfunc";
        conf.serverAddr = "127.0.0.1:8888";
        conf.dataSystemAddr = "127.0.0.1:20001";
        YR::Init(conf);
        //! [Init clusterMode]
    }
    {
        //! [Init and Finalize]
        YR::Config conf;
        conf.mode = YR::Config::Mode::CLUSTER_MODE;
        YR::Init(conf);
        YR::Finalize();
        //! [Init and Finalize]
    }
    return 0;
}