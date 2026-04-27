/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <boost/asio/ssl.hpp>
#include <boost/beast/http.hpp>
#include <string>
#include "src/utility/logger/fileutils.h"
#include "src/utility/logger/logger.h"
#define private public

#include "src/libruntime/traceadaptor/trace_adapter.h"

#include "src/libruntime/traceadaptor/exporter/log_file_exporter_factory.h"

#include <opentelemetry/sdk/trace/span_data.h>

using namespace testing;
using namespace YR::utility;
using namespace YR::Libruntime;
namespace trace_api = opentelemetry::trace;
namespace trace_sdk = opentelemetry::sdk::trace;
namespace nostd = opentelemetry::nostd;
namespace common_sdk = opentelemetry::sdk::common;

namespace YR {
namespace test {

class TraceAdapterTest : public ::testing::Test {
protected:
    static void SetUpTestCase(){}

    static void TearDownTestCase(){}

    void SetUp(){}

    void TearDown(){}

};

TEST_F(TraceAdapterTest, InitTrace)
{
    const std::string configStr = "{\"otlpGrpcExporter\":{\"enable\":true,\"endpoint\":\"127.0.0.1:4317\"},\"logFileExporter\":{\"enable\":true}}";
    const std::string traceServiceName = "testService";
    // enable: false
    TraceAdapter::GetInstance().InitTrace(traceServiceName, false, configStr);
    ASSERT_FALSE(TraceAdapter::GetInstance().enableTrace_);
    // empty config
    TraceAdapter::GetInstance().InitTrace(traceServiceName, true, "");
    ASSERT_FALSE(TraceAdapter::GetInstance().enableTrace_);
    // invalid json string
    const std::string invalidConfigStr1 = "\"otlpGrpcExporter\":{\"enable\":false,\"endpoint\":\"\"},\"logFileExporter\":{\"enable\":false}}";
    TraceAdapter::GetInstance().InitTrace(traceServiceName, true, invalidConfigStr1);
    ASSERT_FALSE(TraceAdapter::GetInstance().enableTrace_);
    // invalid exporter config
    const std::string invalidConfigStr2 = "{\"otlpGrpcExporter\":{\"enable\":true,\"endpoint\":\"\"},\"logFileExporter\":{\"enable\":false}}";
    TraceAdapter::GetInstance().InitTrace(traceServiceName, true, invalidConfigStr2);
    ASSERT_FALSE(TraceAdapter::GetInstance().enableTrace_);
    // valid exporter config
    TraceAdapter::GetInstance().InitTrace(traceServiceName, true, configStr);
    ASSERT_TRUE(TraceAdapter::GetInstance().enableTrace_);

    // set attribute
    TraceAdapter::GetInstance().SetAttr("component", "proxy");
    ASSERT_TRUE(TraceAdapter::GetInstance().attribute_.find("component") != TraceAdapter::GetInstance().attribute_.end());
    ASSERT_EQ(TraceAdapter::GetInstance().attribute_.find("component")->second, "proxy");
}

TEST_F(TraceAdapterTest, StartSpan)
{
    const std::string configStr = "{\"otlpGrpcExporter\":{\"enable\":true,\"endpoint\":\"127.0.0.1:4317\"}}";
    const std::string traceServiceName = "testService";
    TraceAdapter::GetInstance().InitTrace(traceServiceName, false, configStr);
    EXPECT_FALSE(TraceAdapter::GetInstance().enableTrace_);
    auto disableSpan = TraceAdapter::GetInstance().StartSpan("span");
    EXPECT_FALSE(disableSpan->GetContext().trace_id().IsValid());

    TraceAdapter::GetInstance().InitTrace(traceServiceName, true, configStr);
    EXPECT_TRUE(TraceAdapter::GetInstance().enableTrace_);

    auto span1 = TraceAdapter::GetInstance().StartSpan("span1");
    EXPECT_TRUE(span1->GetContext().trace_id().IsValid());

    auto span2 = TraceAdapter::GetInstance().StartSpan("span2",{{"attr1",123},{"attr2", "value2"}});
    EXPECT_TRUE(span2->GetContext().trace_id().IsValid());

}

TEST_F(TraceAdapterTest, TestLogFileExporter)
{
    auto logFileExporter = std::move(LogFileExporterFactory::Create());
    auto record = logFileExporter->MakeRecordable();
    static_cast<trace_sdk::SpanData *>(record.get())->SetAttribute("requestID", "abc");
    static_cast<trace_sdk::SpanData *>(record.get())->SetName("Create");
    ASSERT_EQ(logFileExporter->Export(nostd::span<std::unique_ptr<trace_sdk::Recordable>>(&record, 1)), common_sdk::ExportResult::kSuccess);

    EXPECT_TRUE(logFileExporter->Shutdown());
    ASSERT_EQ(logFileExporter->Export(nostd::span<std::unique_ptr<trace_sdk::Recordable>>(&record, 1)), common_sdk::ExportResult::kFailure);
}

}}