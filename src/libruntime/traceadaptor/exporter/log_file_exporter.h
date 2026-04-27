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

#pragma once

#include <opentelemetry/sdk/trace/exporter.h>
#include <opentelemetry/sdk/trace/simple_processor.h>
#include <opentelemetry/sdk/trace/span_data.h>
#include <opentelemetry/sdk/trace/tracer_provider.h>
#include <opentelemetry/trace/provider.h>
#include <atomic>
#include <chrono>
#include <string>
#include <thread>

namespace nostd = opentelemetry::nostd;
namespace trace_sdk = opentelemetry::sdk::trace;
namespace common_sdk = opentelemetry::sdk::common;

namespace YR {
namespace Libruntime {
std::string TraceIdToString(const opentelemetry::trace::TraceId& traceId);
std::string SpanIdToString(const opentelemetry::trace::SpanId& spanId);

class LogFileExporter final : public trace_sdk::SpanExporter {
public:
    common_sdk::ExportResult Export(
        const nostd::span<std::unique_ptr<trace_sdk::Recordable>> &spans) noexcept override;

    std::unique_ptr<trace_sdk::Recordable> MakeRecordable() noexcept override;

    bool Shutdown(std::chrono::microseconds timeout = std::chrono::microseconds::max()) noexcept override;

    bool ForceFlush(std::chrono::microseconds timeout) noexcept override;

private:
    std::atomic<bool> isShutDown{false};
};
}
}