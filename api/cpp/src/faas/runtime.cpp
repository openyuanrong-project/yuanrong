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

#include "Runtime.h"

#include <utility>

#include "api/cpp/src/executor/executor_holder.h"
#include "api/cpp/src/faas/faas_executor.h"
#include "api/cpp/src/faas/register_runtime_handler.h"
#include "yr/yr.h"

namespace Function {
const std::string DEFAULT_RUNTIME_LOG_PATH = "/home/sn/log/";
const std::string DEFAULT_RUNTIME_LOG_NAME = "cppbin-runtime";
Runtime::Runtime()
{
    InitRuntimeLogger();
}

Runtime::~Runtime()
{
    ReleaseRuntimeLogger();
}

void Runtime::InitRuntimeLogger() const
{
    return;
}

void Runtime::ReleaseRuntimeLogger() const
{
    return;
}

void Runtime::RegisterHandler(
    std::function<std::string(const std::string &request, Function::Context &context)> handleRequestFunc)
{
    this->handleRequest = std::move(handleRequestFunc);
}
void Runtime::RegisterInitializerFunction(std::function<void(Function::Context &)> initializerFunc)
{
    this->initializerFunction = std::move(initializerFunc);
}
void Runtime::RegisterPreStopFunction(std::function<void(Function::Context &)> preStopFunc)
{
    this->preStopFunction = std::move(preStopFunc);
}

void Runtime::InitState(std::function<void(const std::string &, Function::Context &)> initStateFunc)
{
    this->initStateFunction = std::move(initStateFunc);
}

void Runtime::BuildRegisterRuntimeHandler() const
{
    auto handlerPtr = std::make_unique<RegisterRuntimeHandler>();
    handlerPtr->RegisterHandler(this->handleRequest);
    handlerPtr->RegisterInitializerFunction(this->initializerFunction);
    handlerPtr->InitState(this->initStateFunction);
    handlerPtr->RegisterPreStopFunction(this->preStopFunction);
    SetRuntimeHandler(std::move(handlerPtr));
}

void Runtime::Start(int argc, char *argv[])
{
    this->BuildRegisterRuntimeHandler();
    YR::internal::ExecutorHolder::Singleton().SetExecutor(std::make_shared<FaasExecutor>());
    YR::Config conf;
    conf.isDriver = false;
    YR::Init(conf, argc, argv);
    YRLOG_INFO("success to init faas binrary runtime.");
    YR::ReceiveRequestLoop();
}
}  // namespace Function
