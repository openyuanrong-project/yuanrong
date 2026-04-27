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

#ifndef COMMON_TRACE_TRACE_STRUCT_H
#define COMMON_TRACE_TRACE_STRUCT_H
#include <string>
#include <memory>
#include <unordered_map>
#include <list>
#include <variant>

namespace YR {
namespace Libruntime {

const std::string OTLP_GRPC_EXPORTER = "otlpGrpcExporter";
const std::string LOG_FILE_EXPORTER = "logFileExporter";

enum class TraceExporterType : int {
    LOG_FILE, OTEL_GRPC, OTEL_HTTP
};

struct OtelGrpcExporterConfig {
    std::string endpoint;
};
}
}
#endif  // COMMON_TRACE_TRACE_STRUCT_H
