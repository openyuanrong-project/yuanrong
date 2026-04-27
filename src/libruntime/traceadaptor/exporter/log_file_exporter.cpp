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

#include "log_file_exporter.h"
#include <opentelemetry/exporters/ostream/common_utils.h>
#include <opentelemetry/sdk/trace/recordable.h>
#include "src/utility/logger/logger.h"

namespace YR {
namespace Libruntime {
constexpr int TRACE_ID_LEN = 32;
constexpr int SPAN_ID_LEN = 16;

std::string TraceIdToString(const opentelemetry::trace::TraceId& traceId)
{
    char traceIdHex[TRACE_ID_LEN];
    traceId.ToLowerBase16(traceIdHex);
    return std::string(traceIdHex, TRACE_ID_LEN);
}

std::string SpanIdToString(const opentelemetry::trace::SpanId& spanId)
{
    char spanIdHex[SPAN_ID_LEN];
    spanId.ToLowerBase16(spanIdHex);
    return std::string(spanIdHex, SPAN_ID_LEN);
}

common_sdk::ExportResult LogFileExporter::Export(
    const nostd::span<std::unique_ptr<trace_sdk::Recordable>> &spans) noexcept
{
    if (isShutDown) {
        YRLOG_ERROR("[YRLOG File Exporter] Exporting {} log(s) failed, exporter is shutdown", spans.size());
        return common_sdk::ExportResult::kFailure;
    }
    for (auto &recordable : spans) {
        auto span = std::unique_ptr<trace_sdk::SpanData>(static_cast<trace_sdk::SpanData *>(recordable.release()));
        if (span == nullptr) {
            continue;
        }
        std::ostringstream oss;
        oss << "span_name: " << span->GetName() << ", "
            << "trace_id: " << TraceIdToString(span->GetTraceId()) << ", "
            << "span_id: " << SpanIdToString(span->GetSpanId()) << ", "
            << "start_time: " << span->GetStartTime().time_since_epoch().count() << " ns" << ", ";
        oss << "duration: "
            << std::chrono::duration_cast<std::chrono::milliseconds>(span->GetDuration()).count()
            << " ms" << ", ";
        auto attributes = span->GetAttributes();
        oss << "attributes: {";
        for (const auto& [key, value] : attributes) {
            oss << " " << key << " = ";
            opentelemetry::exporter::ostream_common::print_value(value, oss);
        }
        oss << "}";
        YRLOG_INFO("trace info: {}", oss.str());
    }
    return common_sdk::ExportResult::kSuccess;
}

std::unique_ptr<trace_sdk::Recordable> LogFileExporter::MakeRecordable() noexcept
{
    return std::make_unique<trace_sdk::SpanData>();
}

bool LogFileExporter::Shutdown(std::chrono::microseconds timeout) noexcept
{
    isShutDown = true;
    return true;
}

bool LogFileExporter::ForceFlush(std::chrono::microseconds timeout) noexcept
{
    return true;
}

}
}

