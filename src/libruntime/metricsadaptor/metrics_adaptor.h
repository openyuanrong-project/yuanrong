/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.
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

#include <json.hpp>
#include <map>
#include <memory>
#include <string>
#include <unordered_set>

#include "metrics_context.h"

#include "metrics/api/alarm.h"
#include "metrics/api/counter.h"
#include "metrics/api/gauge.h"
#include "metrics/exporters/exporter.h"
#include "metrics/plugin/dynamic_library_handle_unix.h"
#include "metrics/sdk/meter_provider.h"
#include "metrics/sdk/metric_processor.h"
#include "src/dto/invoke_options.h"
#include "src/libruntime/err_type.h"
#include "src/utility/singleton.h"

namespace YR {
namespace Libruntime {
namespace MetricsSdk = observability::sdk::metrics;
namespace MetricsApi = observability::api::metrics;
namespace MetricsExporters = observability::exporters::metrics;
namespace MetricsPlugin = observability::plugin::metrics;

class MetricsAdaptor {
public:
    MetricsAdaptor();
    MetricsAdaptor(const MetricsAdaptor&) = delete;
    MetricsAdaptor& operator=(const MetricsAdaptor&) = delete;
    static std::shared_ptr<MetricsAdaptor> GetInstance()
    {
        std::call_once(initFlag, []() { instance = std::make_shared<MetricsAdaptor>(); });
        return instance;
    }
    bool IsInited() const;
    void Init(const nlohmann::json &json, bool userEnable);
    void SetContextAttr(const std::string &attr, const std::string &value);
    std::string GetContextValue(const std::string &attr) const;
    void CleanMetrics() noexcept;

    ErrorInfo SetUInt64Counter(const YR::Libruntime::UInt64CounterData &data);
    ErrorInfo ResetUInt64Counter(const YR::Libruntime::UInt64CounterData &data);
    ErrorInfo IncreaseUInt64Counter(const YR::Libruntime::UInt64CounterData &data);
    std::pair<ErrorInfo, uint64_t> GetValueUInt64Counter(const YR::Libruntime::UInt64CounterData &data);
    ErrorInfo SetDoubleCounter(const YR::Libruntime::DoubleCounterData &data);
    ErrorInfo ResetDoubleCounter(const YR::Libruntime::DoubleCounterData &data);
    ErrorInfo IncreaseDoubleCounter(const YR::Libruntime::DoubleCounterData &data);
    std::pair<ErrorInfo, double> GetValueDoubleCounter(const YR::Libruntime::DoubleCounterData &data);
    ErrorInfo ReportGauge(const YR::Libruntime::GaugeData &gauge);
    ErrorInfo ReportMetrics(const YR::Libruntime::GaugeData &gauge);
    ErrorInfo SetAlarm(const std::string &name, const std::string &description,
                       const YR::Libruntime::AlarmInfo &alarmInfo);

private:
    std::shared_ptr<MetricsExporters::Exporter> InitHttpExporter(const std::string &httpExporterType,
                                                                 const std::string &backendKey,
                                                                 const std::string &backendName,
                                                                 const nlohmann::json &exporterValue);
    void InitImmediatelyExport(const std::shared_ptr<observability::sdk::metrics::MeterProvider> &mp,
                               const nlohmann::json &backendValue,
                               const std::function<std::string(std::string)> &getFileName);
    void SetImmediatelyExporters(const std::shared_ptr<observability::sdk::metrics::MeterProvider> &mp,
                                 const std::string &backendName, const nlohmann::json &exporters,
                                 const std::function<std::string(std::string)> &getFileName);
    std::shared_ptr<MetricsExporters::Exporter> InitFileExporter(
        const std::string &backendKey, const std::string &backendName, const nlohmann::json &exporterValue,
        const std::function<std::string(std::string)> &getFileName);

    void InitExport(const std::shared_ptr<observability::sdk::metrics::MeterProvider> &provider);
    const MetricsSdk::ExportConfigs BuildExportConfigs(const nlohmann::json &exporterValue);

    ErrorInfo InitUInt64Counter(const YR::Libruntime::UInt64CounterData &data);
    ErrorInfo ReportUInt64Counter(const YR::Libruntime::UInt64CounterData &data);
    ErrorInfo DoResetUInt64Counter(const YR::Libruntime::UInt64CounterData &data);
    ErrorInfo DoIncreaseUInt64Counter(const YR::Libruntime::UInt64CounterData &data);
    std::pair<ErrorInfo, uint64_t> DoGetValueUInt64Counter(const YR::Libruntime::UInt64CounterData &data);
    ErrorInfo InitDoubleCounter(const YR::Libruntime::DoubleCounterData &data);
    ErrorInfo ReportDoubleCounter(const YR::Libruntime::DoubleCounterData &data);
    ErrorInfo DoResetDoubleCounter(const YR::Libruntime::DoubleCounterData &data);
    ErrorInfo DoIncreaseDoubleCounter(const YR::Libruntime::DoubleCounterData &data);
    std::pair<ErrorInfo, double> DoGetValueDoubleCounter(const YR::Libruntime::DoubleCounterData &data);
    ErrorInfo InitDoubleGauge(const YR::Libruntime::GaugeData &gauge);
    ErrorInfo ReportDoubleGauge(const YR::Libruntime::GaugeData &gauge, const std::list<std::string> &contextAttrs);
    std::string GetMetricsFilesName(const std::string &backendName);
    ErrorInfo InitAlarm(const std::string &name, const std::string &description);
    ErrorInfo ReportAlarm(const std::string &name, const std::string &description,
                          const YR::Libruntime::AlarmInfo &alarmInfo);

    std::unordered_map<std::string, std::unique_ptr<MetricsApi::Counter<double>>> doubleCounterMap_{};
    std::unordered_map<std::string, std::unique_ptr<MetricsApi::Counter<uint64_t>>> uInt64CounterMap_{};
    std::map<std::string, std::unique_ptr<MetricsApi::Gauge<double>>> doubleGaugeMap_{};
    std::unordered_map<std::string, std::unique_ptr<MetricsApi::Alarm>> alarmMap_{};
    std::unordered_set<std::string> enabledBackends_;
    MetricsContext metricsContext_;
    bool Initialized_ = false;
    bool userEnable_ = false;
    std::mutex gauge_mutex_{};
    std::mutex alarm_mutex_{};
    std::mutex uint64_counter_mutex_{};
    std::mutex double_counter_mutex_{};
    static std::shared_ptr<MetricsAdaptor> instance;
    static std::once_flag initFlag;
};
}  // namespace Libruntime
}  // namespace YR
