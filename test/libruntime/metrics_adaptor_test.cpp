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

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <boost/asio/ssl.hpp>
#include <boost/beast/http.hpp>
#include <cstdlib>
#include <string>
#include <unistd.h>
#include <dlfcn.h>
#include <climits>
#include <filesystem>
#include <iostream>

#define private public
#include "metrics/api/provider.h"
#include "src/dto/config.h"
#include "src/libruntime/err_type.h"
#include "src/libruntime/metricsadaptor/metrics_adaptor.h"
#include "src/utility/logger/fileutils.h"
#include "src/utility/logger/logger.h"

using namespace testing;
using namespace YR::utility;
using namespace YR::Libruntime;
namespace MetricsApi = observability::api::metrics;
namespace YR {
namespace test {

nlohmann::json GetValidConfig()
{
    const std::string str = R"(
{
    "backends": [
        {
            "immediatelyExport": {
                "name": "CAAS_Alarm",
                "enable": true,
                "custom": {
                    "labels": {
                        "site": "",
                        "tenant_id": "",
                        "application_id": "",
                        "service_id": ""
                    }
                },
                "exporters": [
                    {
                        "fileExporter": {
                            "enable": true,
                            "fileDir": "/home/sn/metrics/",
                            "rolling": {
                                "enable": true,
                                "maxFiles": 3,
                                "maxSize": 10000
                            },
                            "contentType": "LABELS"
                        }
                    }
                ]
            }
        },
        {
            "immediatelyExport": {
                "name": "LingYun",
                "enable": true,
                "exporters": [
                    {
                        "prometheusPushExporter": {
                            "enable": true,
                            "batchSize": 2,
                            "batchIntervalSec": 10,
                            "failureQueueMaxSize": 2,
                            "failureDataDir": "/home/sn/metrics/failure",
                            "failureDataFileMaxCapacity": 1,
                            "initConfig": {
                                "ip": "127.0.0.1",
                                "port": 31061
                            }
                        }
                    }
                ]
            }
        }
    ]
}
    )";
    return nlohmann::json::parse(str);
}

nlohmann::json GetUnsupportedConfig()
{
    const std::string str = R"(
{
    "backends": [
        {
            "batchExport": {"name": "CAAS_Alarm"}
        }
    ]
}
    )";
    return nlohmann::json::parse(str);
}

nlohmann::json GetInvalidConfig()
{
    const std::string str = R"(
{
    "invalid": []
}
    )";
    return nlohmann::json::parse(str);
}

nlohmann::json GetImmedExportNotEnableConfig()
{
    const std::string str = R"(
{
    "backends": [
        {
            "immediatelyExport": {
                "name": "LingYun",
                "enable": false,
                "exporters": [
                    {
                        "prometheusPushExporter": {
                            "enable": true,
                            "ip": "127.0.0.1",
                            "port": 9091
                        }
                    }
                ]
            }
        }
    ]
}
    )";
    return nlohmann::json::parse(str);
}

nlohmann::json GetPromExportNotEnableConfig()
{
    const std::string str = R"(
{
    "backends": [
        {
            "immediatelyExport": {
                "name": "LingYun",
                "enable": true,
                "exporters": [
                    {
                        "prometheusPushExporter": {
                            "enable": false,
                            "ip": "127.0.0.1",
                            "port": 9091
                        }
                    }
                ]
            }
        }
    ]
}
    )";
    return nlohmann::json::parse(str);
}

nlohmann::json GetPrometheusPushExporterConfig()
{
    std::string jsonStr = R"(
{
    "enable": true,
    "initConfig": {
        "ip": "x",
        "port": 0
    }
}
    )";
    return nlohmann::json::parse(jsonStr);
}

nlohmann::json GetExportConfigs()
{
    std::string jsonStr = R"(
{
    "enable": true,
    "enabledInstruments": ["name"],
    "batchSize": 5,
    "initConfig": {
        "ip": "127.0.0.1",
        "port": 31061
    }
}
    )";
    return nlohmann::json::parse(jsonStr);
}

static std::string GetSelfPathForTest()
{
    Dl_info info;
    if (dladdr((void*)&GetSelfPathForTest, &info) == 0) {
        return "";
    }
    std::string path(info.dli_fname);
    auto pos = path.find_last_of('/');
    if (pos == std::string::npos) {
        return "";
    }
    return path.substr(0, pos);
}

static void PrepareExporterSo()
{
    static bool prepared = false;
    if (prepared) {
        return;
    }
    prepared = true;

    auto path = GetSelfPathForTest();

    auto idx = path.find("yuanrong/");
    if (idx == std::string::npos) {
        std::cerr << "cannot find '/yuanrong/' in path: " << path << "\n";
        return;
    }

    std::string srcDir = path.substr(0, idx) + "yuanrong/metrics/lib";
    std::string dstDir = std::filesystem::current_path().string() + "/test";

    std::filesystem::create_directories(dstDir);

    const std::vector<std::string> exporters = {
        "libobservability-metrics-file-exporter.so",
        "libobservability-prometheus-push-exporter.so"
    };

    for (const auto& so : exporters) {
        std::filesystem::path src = srcDir + "/" + so;
        std::filesystem::path dst = dstDir + "/" +so;

        try {
            std::filesystem::copy_file(
                src,
                dst,
                std::filesystem::copy_options::overwrite_existing
            );
        } catch (const std::filesystem::filesystem_error & e) {
            std::cerr << "Failed to copy " << "\n";
        }
    }
}

class MetricsAdaptorTest : public testing::Test {
public:
    MetricsAdaptorTest(){};
    ~MetricsAdaptorTest(){};
    void SetUp() override
    {
        PrepareExporterSo();
        Mkdir("/tmp/log");
        LogParam g_logParam = {
            .logLevel = "DEBUG",
            .logDir = "/tmp/log",
            .nodeName = "test-runtime",
            .modelName = "test",
            .maxSize = 100,
            .maxFiles = 1,
            .logFileWithTime = false,
            .logBufSecs = 30,
            .maxAsyncQueueSize = 1048510,
            .asyncThreadCount = 1,
            .alsoLog2Stderr = true,
        };
        InitLog(g_logParam);
    }
    void TearDown() override {}

    std::string config;

private:
    pthread_t tids[1];
    int port = 22222;
};

TEST_F(MetricsAdaptorTest, InitSuccessullyTest)
{
    setenv("YR_SSL_ENABLE", "true", 1);
    setenv("YR_SSL_PASSPHRASE", "YR_SSL_PASSPHRASE", 1);
    Config::c = Config();
    auto nullMeterProvider = MetricsApi::Provider::GetMeterProvider();
    auto jsonStr = GetValidConfig();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    metricsAdaptor->Init(jsonStr, true);
    EXPECT_NE(MetricsApi::Provider::GetMeterProvider(), nullMeterProvider);
    metricsAdaptor->CleanMetrics();
    EXPECT_EQ(MetricsApi::Provider::GetMeterProvider(), nullptr);
    unsetenv("YR_SSL_ENABLE");
    unsetenv("YR_SSL_PASSPHRASE");
}

TEST_F(MetricsAdaptorTest, UnsupportedInitTest)
{
    Config::c = Config();
    auto nullMeterProvider = MetricsApi::Provider::GetMeterProvider();
    auto jsonStr = GetUnsupportedConfig();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    metricsAdaptor->Init(jsonStr, true);
    EXPECT_NE(MetricsApi::Provider::GetMeterProvider(), nullMeterProvider);
    metricsAdaptor->CleanMetrics();
    EXPECT_EQ(MetricsApi::Provider::GetMeterProvider(), nullptr);
}

TEST_F(MetricsAdaptorTest, InvalidInitTest)
{
    Config::c = Config();
    auto jsonStr = GetInvalidConfig();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    metricsAdaptor->Init(jsonStr, true);
    metricsAdaptor->CleanMetrics();
    EXPECT_EQ(MetricsApi::Provider::GetMeterProvider(), nullptr);
}

TEST_F(MetricsAdaptorTest, InitNotEnableTest)
{
    Config::c = Config();
    auto nullMeterProvider = MetricsApi::Provider::GetMeterProvider();
    auto jsonStr = GetImmedExportNotEnableConfig();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    metricsAdaptor->Init(jsonStr, true);
    EXPECT_NE(MetricsApi::Provider::GetMeterProvider(), nullMeterProvider);
    metricsAdaptor->CleanMetrics();
    EXPECT_EQ(MetricsApi::Provider::GetMeterProvider(), nullptr);

    auto jsonStr2 = GetPromExportNotEnableConfig();
    auto metricsAdaptor2 = std::make_shared<MetricsAdaptor>();
    metricsAdaptor2->Init(jsonStr2, true);
    EXPECT_NE(MetricsApi::Provider::GetMeterProvider(), nullMeterProvider);
    metricsAdaptor2->CleanMetrics();
    EXPECT_EQ(MetricsApi::Provider::GetMeterProvider(), nullptr);
}

TEST_F(MetricsAdaptorTest, DoubleGaugeTest)
{
    Config::c = Config();
    auto jsonStr = GetValidConfig();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    metricsAdaptor->Init(jsonStr, true);
    YR::Libruntime::GaugeData gauge;
    gauge.name = "name";
    gauge.description = "desc";
    gauge.unit = "unit";
    gauge.value = 1.11;
    auto err = metricsAdaptor->ReportMetrics(gauge);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    err = metricsAdaptor->ReportGauge(gauge);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    metricsAdaptor->CleanMetrics();
    EXPECT_EQ(MetricsApi::Provider::GetMeterProvider(), nullptr);
}

TEST_F(MetricsAdaptorTest, SetAlarmTest)
{
    Config::c = Config();
    auto jsonStr = GetValidConfig();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    metricsAdaptor->Init(jsonStr, true);
    YR::Libruntime::AlarmInfo alarmInfo;
    alarmInfo.alarmName = "name";
    alarmInfo.locationInfo = "info";
    alarmInfo.cause = "cause";
    auto err = metricsAdaptor->SetAlarm("name", "desc", alarmInfo);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    err = metricsAdaptor->SetAlarm("name", "desc", alarmInfo);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    metricsAdaptor->CleanMetrics();
    EXPECT_EQ(MetricsApi::Provider::GetMeterProvider(), nullptr);
}

TEST_F(MetricsAdaptorTest, DoubleCounterTest)
{
    Config::c = Config();
    auto jsonStr = GetValidConfig();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    metricsAdaptor->Init(jsonStr, true);
    YR::Libruntime::DoubleCounterData data;
    data.name = "name";
    data.description = "desc";
    data.unit = "unit";
    data.value = 1.11;
    auto err = metricsAdaptor->SetDoubleCounter(data);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    auto res = metricsAdaptor->GetValueDoubleCounter(data);
    EXPECT_EQ(res.second, 1.11);
    err = metricsAdaptor->IncreaseDoubleCounter(data);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    res = metricsAdaptor->GetValueDoubleCounter(data);
    EXPECT_EQ(res.second, 2.22);
    err = metricsAdaptor->ResetDoubleCounter(data);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    res = metricsAdaptor->GetValueDoubleCounter(data);
    EXPECT_EQ(res.second, 0);
    metricsAdaptor->CleanMetrics();
    EXPECT_EQ(MetricsApi::Provider::GetMeterProvider(), nullptr);
}

TEST_F(MetricsAdaptorTest, UInt64CounterTest)
{
    Config::c = Config();
    auto jsonStr = GetValidConfig();
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    metricsAdaptor->Init(jsonStr, true);
    YR::Libruntime::UInt64CounterData data;
    data.name = "name";
    data.description = "desc";
    data.unit = "unit";
    data.value = 1;
    auto err = metricsAdaptor->SetUInt64Counter(data);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    auto res = metricsAdaptor->GetValueUInt64Counter(data);
    EXPECT_EQ(res.second, 1);
    err = metricsAdaptor->IncreaseUInt64Counter(data);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    res = metricsAdaptor->GetValueUInt64Counter(data);
    EXPECT_EQ(res.second, 2);
    err = metricsAdaptor->ResetUInt64Counter(data);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_OK);
    res = metricsAdaptor->GetValueUInt64Counter(data);
    EXPECT_EQ(res.second, 0);
    metricsAdaptor->CleanMetrics();
    EXPECT_EQ(MetricsApi::Provider::GetMeterProvider(), nullptr);
}

TEST_F(MetricsAdaptorTest, MetricsFailedTest)
{
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    YR::Libruntime::UInt64CounterData data;
    auto err = metricsAdaptor->SetUInt64Counter(data);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    err = metricsAdaptor->IncreaseUInt64Counter(data);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    err = metricsAdaptor->ResetUInt64Counter(data);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    auto res = metricsAdaptor->GetValueUInt64Counter(data);
    EXPECT_EQ(res.first.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);

    YR::Libruntime::DoubleCounterData data2;
    err = metricsAdaptor->SetDoubleCounter(data2);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    err = metricsAdaptor->IncreaseDoubleCounter(data2);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    err = metricsAdaptor->ResetDoubleCounter(data2);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    auto res2 = metricsAdaptor->GetValueDoubleCounter(data2);
    EXPECT_EQ(res2.first.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);

    YR::Libruntime::AlarmInfo alarmInfo;
    YR::Libruntime::GaugeData gauge;
    err = metricsAdaptor->SetAlarm("name", "desc", alarmInfo);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    err = metricsAdaptor->ReportMetrics(gauge);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
    err = metricsAdaptor->ReportGauge(gauge);
    EXPECT_EQ(err.Code(), YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR);
}

TEST_F(MetricsAdaptorTest, contextTest)
{
    MetricsContext metricsContext;
    std::string attr = "test_attr";
    std::string value = "test_value";
    metricsContext.SetAttr(attr, value);
    std::string result = metricsContext.GetAttr(attr);
    ASSERT_EQ(result, value);

    attr = "test_attr_key";
    result = metricsContext.GetAttr(attr);
    ASSERT_EQ(result, "");
}

TEST_F(MetricsAdaptorTest, InitHttpExporterWithTLS)
{
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();

    setenv("YR_SSL_ENABLE", "true", 1);
    setenv("YR_SSL_ROOT_FILE", "root", 1);
    setenv("YR_SSL_CERT_FILE", "cert", 1);
    setenv("YR_SSL_KEY_FILE", "key", 1);
    setenv("YR_SSL_PASSPHRASE", "123", 1);

    YR::Libruntime::Config::c = YR::Libruntime::Config();

    auto config = GetPrometheusPushExporterConfig();
    // Exporter may be null if SSL cert files are invalid; the key behavior under test is passphrase erasure.
    metricsAdaptor->InitHttpExporter("prometheusPushExporter", "key", "name", config);
    // Cleared via setenv(..., ""); on some platforms (e.g. Darwin) that removes the var, so getenv is nullptr.
    auto value = std::getenv("YR_SSL_PASSPHRASE");
    EXPECT_TRUE(value == nullptr || value[0] == '\0');

    unsetenv("YR_SSL_ENABLE");
    unsetenv("YR_SSL_ROOT_FILE");
    unsetenv("YR_SSL_CERT_FILE");
    unsetenv("YR_SSL_KEY_FILE");
    unsetenv("YR_SSL_PASSPHRASE");
    YR::Libruntime::Config::c = YR::Libruntime::Config();
}

TEST_F(MetricsAdaptorTest, BuildExportConfigsTest)
{
    auto metricsAdaptor = std::make_shared<MetricsAdaptor>();
    auto js = GetExportConfigs();
    ASSERT_TRUE(js.contains("enabledInstruments"));
    auto config = metricsAdaptor->BuildExportConfigs(GetExportConfigs());
    ASSERT_TRUE(config.enabledInstruments.count("name") > 0);
}
}  // namespace test
}  // namespace YR
