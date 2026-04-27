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

#include <iostream>
#include "yr/api/runtime_env.h"
#include "yr/yr.h"
int main(int argc, char **argv)
{
    std::string pyFunctionUrn = "sn:cn:yrk:default:function:0-yr-mypython:$latest";

    YR::Config conf;
    YR::Init(conf, argc, argv);
    YR::InvokeOptions opts;
    YR::RuntimeEnv runtimeEnv;
    runtimeEnv.Set<std::string>("conda", "pytorch_p39");
    {
        //! [set env_vars]
        YR::RuntimeEnv runtimeEnv;
        runtimeEnv.Set<std::map<std::string, std::string>>("env_vars", {{"OMP_NUM_THREADS", "32"}, {"TF_WARNINGS", "none"}});
        //! [set env_vars]
    }
    {
        //! [set pip]
        YR::RuntimeEnv env;
        env.Set<std::vector<std::string>>("pip", {"numpy=2.3.0", "pandas"});
        //! [set pip]
    }
    {
        //! [set working_dir]
        YR::RuntimeEnv env;
        runtimeEnv.Set<std::string>("working_dir", "file:/opt/mycode/cpp-invoke-python/calculator.zip");
        //! [set working_dir]
    }
    {
        //! [set existed conda environ]
        YR::RuntimeEnv env;
        runtimeEnv.Set<std::string>("conda", "pytorch_p39");
        //! [set existed conda environ]
    }
    {
        //! [set conda environ with dependency]
        // If name is not specified, the name will start with 'virtual_env-' followed by a randomly generated suffix.
        runtimeEnv.Set<nlohmann::json>("conda", {{"name", "pytorch_p39"},{"channels", {"conda-forge"}}, {"dependencies", {"python=3.9", "matplotlib", "msgpack-python=1.0.5", "protobuf", "libgcc-ng", "numpy", "pandas", "cloudpickle=2.0.0", "cython=3.0.10", "pyyaml=6.0.2"}}});
        runtimeEnv.Set<std::map<std::string, std::string>>("env_vars", {{"OMP_NUM_THREADS", "32"}, {"TF_WARNINGS", "none"}, {"YR_CONDA_HOME", "/home/snuser/.conda"}});
        //! [set conda environ with dependency]
    }
    {
        //! [set conda environ with yaml file]
        YR::RuntimeEnv runtimeEnv;
        /* yaml file demo
         * name: myenv3
         * channels:
         *  - conda-forge
         *  - defaults
         * dependencies:
         *  - python=3.9
         *  - numpy
         *  - pandas
         *  - cloudpickle=2.2.1
         *  - msgpack-python=1.0.5
         *  - protobuf
         *  - cython=3.0.10
         *  - pyyaml=6.0.2
         */
        runtimeEnv.Set<std::string>("conda", "/opt/conda/env-xpf.yaml");
        //! [set conda environ with yaml file]
    }
    runtimeEnv.Set<std::map<std::string, std::string>>("env_vars", {{"OMP_NUM_THREADS", "32"}, {"TF_WARNINGS", "none"}, {"YR_CONDA_HOME", "/home/snuser/.conda"}});
    opts.runtimeEnv = runtimeEnv;
    auto resFutureSquare = YR::PyFunction<int>("calculator", "square").SetUrn(pyFunctionUrn).Options(opts).Invoke(2);
    auto resSquare = *YR::Get(resFutureSquare);
    std::cout << resSquare << std::endl;
    return 0;
}
