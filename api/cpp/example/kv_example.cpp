/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
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

#include <iostream>

#include "yr/yr.h"

int main()
{
    YR::Config conf;
    conf.inCluster = true;
    conf.functionUrn = "sn:cn:yrk:default:function:0-opc-opc:$latest";
    conf.serverAddr = "";      // bus proxy IP:port
    conf.dataSystemAddr = "";  // datasystem worker IP:port
    std::vector<std::string> paths;
    paths.emplace_back("");  // add your library path
    conf.loadPaths = std::move(paths);
    YR::Init(conf);

    std::string key = "kv-id-888";
    std::string value = "kv-value-888";
    YR::KV().Write(key, value);

    std::shared_ptr<std::string> result = YR::KV().Read<std::string>(key);
    std::cout << *result << std::endl;

    YR::KV().Del(key);

    // Legacy API
    YR::KV().Set(key, value);

    std::string result2 = YR::KV().Get(key);
    std::cout << result2 << std::endl;

    YR::KV().Del(key);

    // Set with SetParam
    YR::ExistenceOpt existence = YR::ExistenceOpt::NONE;
    YR::WriteMode writeMode = YR::WriteMode::NONE_L2_CACHE;
    YR::SetParam setParam = {
        .writeMode = writeMode,
        .ttlSecond = 5,
        .existence = existence,
    };
    YR::KV().Set(key, value, setParam);

    std::string result3 = YR::KV().Get(key);
    std::cout << result3 << std::endl;
    sleep(8);  // sleep 8s
    try {
        std::string result4 = YR::KV().Get(key, 10);
        std::cout << result4 << std::endl;
    } catch (YR::Exception &e) {
        std::cout << e.what() << std::endl;
    }

    YR::Finalize();
    return 0;
}