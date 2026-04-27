/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
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

//! [runtime env demo]
#include <iostream>
#include "yr/api/runtime_env.h"
#include "yr/yr.h"
int main(int argc, char **argv) {
    std::string pyFunctionUrn = "sn:cn:yrk:default:function:0-yr-mypython:$latest";

    YR::Config conf;
    YR::Init(conf, argc, argv);
    YR::InvokeOptions opts;
    YR::RuntimeEnv runtimeEnv;
    runtimeEnv.Set<std::string>("conda", "pytorch_p39");
    runtimeEnv.Set<std::map<std::string, std::string>>("env_vars", {{"OMP_NUM_THREADS", "32"}, {"TF_WARNINGS", "none"}, {"YR_CONDA_HOME", "/home/snuser/.conda"}});
    opts.runtimeEnv = runtimeEnv;
    auto resFutureSquare = YR::PyFunction<int>("calculator", "square").SetUrn(pyFunctionUrn).Options(opts).Invoke(2);
    auto resSquare = *YR::Get(resFutureSquare);
    std::cout << resSquare << std::endl;
    return 0;
}
//! [runtime env demo]