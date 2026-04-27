/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.
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
//! [faasCallerHandler]
#include <cstdlib>
#include <string>
#include "Runtime.h"
#include "FunctionError.h"
#include "Function.h"
bool flags = false;
std::string HandleRequest(const std::string &request, Function::Context &context)
{
    Function::FunctionLogger logger = context.GetLogger();

    logger.Info("invoke test function begin");
    std::string result = "";
    try {
        auto func = Function::Function(context, "test:latest");
        Function::InvokeOptions invokeOptions; // use default option
        func.Options(invokeOptions);
        auto ref = func.Invoke("{\"hello\":\"world\"}");
        result = ref.Get();
    } catch (Function::FunctionError e) {
        logger.Error("invoke test function failed, err: %s", e.GetJsonString());
        return e.GetJsonString();
    }
    logger.Info("invoke test function end, result: %s", result);

    return result;
}

void Initializer(Function::Context &context)
{
    flags = true;
}

const std::string DEFAULT_PORT = "31552";
int main(int argc, char *argv[])
{
    Function::Runtime rt;
    rt.RegisterHandler(HandleRequest);
    rt.RegisterInitializerFunction(Initializer);
    rt.Start(argc, argv);
    return 0;
}
//! [faasCallerHandler]