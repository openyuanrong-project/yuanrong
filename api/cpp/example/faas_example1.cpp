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

#include <cstdlib>
#include <string>
#include "Runtime.h"
#include "Constant.h"
#include "FunctionError.h"
#include "Function.h"
bool flags = false;
std::string HandleRequest(const std::string &request, Function::Context &context)
{
    {
        //! [logger_usage]
        Function::FunctionLogger logger = context.GetLogger();
        logger.setLevel("INFO");
        logger.Info("hello cpp %s ", "user info log");
        logger.Error("hello cpp %s ", "user error log");
        logger.Warn("hello cpp %s ", "user warn log");
        logger.Debug("hello cpp %s ", "user debug log");
        logger.Error("hello cpp %s ", context.GetFunctionName().c_str());
        logger.Error("hello cpp %s ", context.GetUserData("b").c_str());
        logger.Error("hello cpp %s ", context.GetUserData("key1").c_str());
        //! [logger_usage]
    }

    {
        //! [context_usage]
        Function::FunctionLogger logger = context.GetLogger();
        logger.setLevel("INFO");
        logger.Info("hello cpp %s ", "user info log");
        std::string envValue = context.GetUserData("a");
        std::string funcName = context.GetFunctionName();
        logger.Info("env: a's value is %s, function name is %s", envValue, funcName);
        //! [context_usage]
    }

    {
        //! [functionerror_usage]
        // need add #include "FunctionError.h"
        Function::FunctionLogger logger = context.GetLogger();
        try {
            auto func = Function::Function(context, "hello");
            auto ref = func.Invoke("{}");
            string result = ref.Get();
        } catch (Function::FunctionError e) {
            if (e.GetErrorCode() == Function::ErrorCode::FUNCTION_EXCEPTION) {  // need add #include "Constant.h"
                logger.Error("err is FUNCTION_EXCEPTION, errCode is %d, errMsg is %s",
                    e.GetErrorCode(), e.GetMessage());
            } else {
                logger.Error("errCode is %d, errMsg is %s", e.GetErrorCode(), e.GetMessage());
            }
        }
        //! [functionerror_usage]
    }

    return request;
}

void InitState(const std::string &request, Function::Context &context)
{
    context.SetState(request);
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
