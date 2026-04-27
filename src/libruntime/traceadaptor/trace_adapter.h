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

#ifndef COMMON_TRACE_TRACE_ADAPTER_H
#define COMMON_TRACE_TRACE_ADAPTER_H

#include <iomanip>
#include <vector>
#include <opentelemetry/nostd/shared_ptr.h>
#include <opentelemetry/sdk/trace/exporter.h>
#include <opentelemetry/sdk/resource/semantic_conventions.h>
#include <opentelemetry/trace/span.h>
#include <opentelemetry/trace/span_startoptions.h>
#include <opentelemetry/sdk/trace/tracer_provider.h>
#include <opentelemetry/trace/noop.h>
#include <opentelemetry/trace/provider.h>
#include <opentelemetry/exporters/otlp/otlp_grpc_exporter_factory.h>
#include <opentelemetry/sdk/trace/batch_span_processor.h>
#include <opentelemetry/sdk/trace/batch_span_processor_factory.h>
#include <opentelemetry/sdk/trace/batch_span_processor_options.h>
#include <opentelemetry/sdk/trace/processor.h>
#include "src/utility/singleton.h"
#include "trace_struct.h"

namespace trace_api = opentelemetry::trace;
namespace trace_sdk = opentelemetry::sdk::trace;

namespace YR {
namespace Libruntime {

using OtelSpan = opentelemetry::nostd::shared_ptr<opentelemetry::trace::Span>;
using OtelAttrVector =
    std::vector<std::pair<const std::string, const opentelemetry::common::AttributeValue>>;

class TraceAdapter : public utility::Singleton<TraceAdapter> {
public:
    TraceAdapter() = default;
    ~TraceAdapter() noexcept override;

    void InitTrace(const std::string &serviceName, const bool &enableTrace, const std::string &traceConfig);

    void SetAttr(const std::string &attr, const std::string &value);

    OtelSpan StartSpan(const std::string &name, const opentelemetry::common::KeyValueIterable &attributes,
                       const opentelemetry::trace::SpanContextKeyValueIterable &links,
                       const opentelemetry::trace::StartSpanOptions &startSpanOptions);

    OtelSpan StartSpan(const std::string &name, const opentelemetry::trace::StartSpanOptions &startSpanOptions = {});

    OtelSpan StartSpan(const std::string &name, OtelAttrVector attrs,
                       const opentelemetry::trace::StartSpanOptions &startSpanOptions = {});

    OtelSpan StartSpan(const std::string &name, const std::string &traceID, const std::string &spanID,
                       OtelAttrVector attrs = {});

    opentelemetry::nostd::shared_ptr<opentelemetry::trace::Tracer> GetTracer(const std::string &name = "yuanrong",
                                                                             const std::string &version = "");

    void ShutDown();

private:
    bool enableTrace_{ false };

    std::map<std::string, std::string> attribute_;

    std::unique_ptr<opentelemetry::sdk::trace::SpanExporter> InitOtlpGrpcExporter(const OtelGrpcExporterConfig &conf);
    std::unique_ptr<opentelemetry::sdk::trace::SpanExporter> InitLogFileExporter();
    opentelemetry::trace::StartSpanOptions BuildOptWithParent(const std::string &traceID, const std::string &spanID);
};
}
}
#endif  // COMMON_TRACE_TRACE_ADAPTER_H
