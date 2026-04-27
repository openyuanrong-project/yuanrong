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

//! [parallel for]
#include "yr/parallel/parallel_for.h"

int main(int argc, char **argv)
{
    YR::Config conf;
    conf.mode = YR::Config::Mode::LOCAL_MODE;
    const int threadPoolSize = 8;
    conf.threadPoolSize = threadPoolSize;
    YR::Init(conf);
    std::vector<uint64_t> results;
    const uint32_t start = 0;
    const uint32_t end = 1000000;
    results.resize(end);
    const uint32_t chunkSize = 1000;
    const int workerNum = 4;

    auto handler = [&results](size_t start, size_t end) {
        for (size_t i = start; i < end; i++) {
            results[i] += i;
        }
    };
    YR::Parallel::ParallelFor<uint32_t>(start, end, handler, chunkSize, workerNum);

    std::vector<std::vector<int>> v(workerNum);
    auto f = [&v](int start, int end, const YR::Parallel::Context &ctx) {
        std::cout << "start: " << start << " , end: " << end << " ctx: " << ctx.id << std::endl;
        for (int i = start; i < end; i++) {
            int result = i;
            if (result) {
                v[ctx.id].emplace_back(result);
            }
        }
    };
    YR::Parallel::ParallelFor<uint32_t>(start, end, f, chunkSize, workerNum);
}
//! [parallel for]