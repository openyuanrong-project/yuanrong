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

#include <dlfcn.h>
#include <string>
#include "metrics_context.h"
#include "metrics_adaptor.h"


#include "metrics/api/metric_data.h"
#include "metrics/api/null.h"
#include "metrics/api/provider.h"
#include "metrics/sdk/immediately_export_processor.h"
#include "metrics/sdk/meter_provider.h"
#include "src/dto/config.h"
#include "src/utility/logger/logger.h"
#include "src/utility/logger/fileutils.h"

namespace YR {
namespace Libruntime {
const char *const IMMEDIATELY_EXPORT = "immediatelyExport";
const char *const FILE_EXPORTER = "fileExporter";
const char *const PROMETHEUS_PUSH_EXPORTER = "prometheusPushExporter";
const char *const AOM_ALARM_EXPORTER = "aomAlarmExporter";
const char *const YR_SSL_PASSPHRASE_KEY = "YR_SSL_PASSPHRASE";

std::string GetSelfSoPath()
{
    Dl_info info;
    if (dladdr((void*)&GetSelfSoPath, &info) == 0) {
        return "";
    }
    std::string path(info.dli_fname);
    auto pos = path.find_last_of('/');
    if (pos == std::string::npos) {
        return "";
    }

    return path.substr(0, pos);
}

static std::string GetLibraryPath(const std::string &exporterType)
{
    auto path = GetSelfSoPath();
    std::string filePath = "";
    if (exporterType == FILE_EXPORTER) {
        filePath = path + "/libobservability-metrics-file-exporter.so";
    } else if (exporterType == PROMETHEUS_PUSH_EXPORTER) {
        filePath = path + "/libobservability-prometheus-push-exporter.so";
    } else if (exporterType == AOM_ALARM_EXPORTER) {
        filePath = path + "/libobservability-aom-alarm-exporter.so";
    }
    YRLOG_INFO("exporter {} get library path: {}", exporterType, filePath);
    return filePath;
}

std::once_flag MetricsAdaptor::initFlag;
std::shared_ptr<MetricsAdaptor> MetricsAdaptor::instance = nullptr;

MetricsAdaptor::MetricsAdaptor() {}

bool MetricsAdaptor::IsInited() const
{
    return Initialized_;
}

void MetricsAdaptor::Init(const nlohmann::json &json, bool userEnable)
{
    userEnable_ = userEnable;
    YRLOG_DEBUG("start to init metrics adaptor, userEnable {}", userEnable);
    if (json.find("backends") == json.end()) {
        YRLOG_WARN("metrics backends is none");
        return;
    }
    observability::sdk::metrics::LiteBusParams liteBusParam;
    auto mp = std::make_shared<MetricsSdk::MeterProvider>(liteBusParam);
    auto backends = json.at("backends");
    for (auto &[index, backend] : backends.items()) {
        YRLOG_DEBUG("metrics add backend index({})", index);
        for (auto &[key, value] : backend.items()) {
            if (key == IMMEDIATELY_EXPORT) {
                InitImmediatelyExport(mp, value,
                                      [this](std::string backendName) { return GetMetricsFilesName(backendName); });
            } else {
                YRLOG_WARN("unknown backend key: {}", key);
            }
        }
    }
    MetricsApi::Provider::SetMeterProvider(mp);
}

void MetricsAdaptor::SetContextAttr(const std::string &attr, const std::string &value)
{
    metricsContext_.SetAttr(attr, value);
}

std::string MetricsAdaptor::GetContextValue(const std::string &attr) const
{
    return metricsContext_.GetAttr(attr);
}

void MetricsAdaptor::InitImmediatelyExport(const std::shared_ptr<MetricsSdk::MeterProvider> &mp,
                                           const nlohmann::json &backendValue,
                                           const std::function<std::string(std::string)> &getFileName)
{
    YRLOG_DEBUG("metrics add backend {}", IMMEDIATELY_EXPORT);
    if (backendValue.find("enable") == backendValue.end() || !backendValue.at("enable").get<bool>()) {
        YRLOG_DEBUG("metrics backend {} is not enabled", IMMEDIATELY_EXPORT);
        return;
    }
    std::string backendName;
    if (backendValue.find("name") != backendValue.end()) {
        backendName = backendValue.at("name");
        YRLOG_DEBUG("metrics add backend {} of {}", IMMEDIATELY_EXPORT, backendName);
        enabledBackends_.insert(backendName);
    }
    if (backendValue.find("custom") != backendValue.end()) {
        auto custom = backendValue.at("custom");
        if (custom.find("labels") != custom.end()) {
            auto labels = custom.at("labels");
            for (auto &label : labels.items()) {
                YRLOG_DEBUG("metrics backend {} of {} add custom label, key: {}, value: {}", IMMEDIATELY_EXPORT,
                            backendName, label.key(), label.value().dump());
                metricsContext_.SetAttr(label.key(), label.value());
            }
        }
    }
    if (backendValue.find("exporters") != backendValue.end()) {
        for (auto &[index, exporters] : backendValue.at("exporters").items()) {
            YRLOG_DEBUG("metrics add exporter index({}) for backend {}", index, backendName);
            SetImmediatelyExporters(mp, backendName, exporters, getFileName);
        }
    }
}

void MetricsAdaptor::SetImmediatelyExporters(const std::shared_ptr<observability::sdk::metrics::MeterProvider> &mp,
                                             const std::string &backendName, const nlohmann::json &exporters,
                                             const std::function<std::string(std::string)> &getFileName)
{
    if (mp == nullptr) {
        return;
    }
    for (auto &[key, value] : exporters.items()) {
        if (key == FILE_EXPORTER) {
            auto &&exporter = InitFileExporter(IMMEDIATELY_EXPORT, backendName, value, getFileName);
            if (exporter == nullptr) {
                continue;
            }
            Initialized_ = true;
            auto exportConfigs = BuildExportConfigs(value);
            exportConfigs.exporterName = key;
            exportConfigs.exportMode = MetricsSdk::ExportMode::IMMEDIATELY;
            auto processor = std::make_shared<observability::sdk::metrics::ImmediatelyExportProcessor>(
                std::move(exporter), exportConfigs);
            mp->AddMetricProcessor(std::move(processor));
        } else if (key == PROMETHEUS_PUSH_EXPORTER || key == AOM_ALARM_EXPORTER) {
            auto &&exporter = InitHttpExporter(key, IMMEDIATELY_EXPORT, backendName, value);
            if (exporter == nullptr) {
                YRLOG_ERROR("Failed to init exporter {}", key);
                continue;
            }
            Initialized_ = true;
            auto exportConfigs = BuildExportConfigs(value);
            exportConfigs.exporterName = key;
            exportConfigs.exportMode = MetricsSdk::ExportMode::IMMEDIATELY;
            auto processor =
                std::make_shared<MetricsSdk::ImmediatelyExportProcessor>(std::move(exporter), exportConfigs);
            mp->AddMetricProcessor(std::move(processor));
        } else {
            YRLOG_WARN("unknown exporter name: {}", key);
        }
    }
}

std::shared_ptr<MetricsExporters::Exporter> MetricsAdaptor::InitHttpExporter(const std::string &httpExporterType,
                                                                             const std::string &backendKey,
                                                                             const std::string &backendName,
                                                                             const nlohmann::json &exporterValue)
{
    YRLOG_DEBUG("add exporter {} for backend {} of {}", httpExporterType, backendKey, backendName);
    if (exporterValue.find("enable") == exporterValue.end() || !exporterValue.at("enable").get<bool>()) {
        YRLOG_DEBUG("metrics exporter {} for backend {} of {} is not enabled", httpExporterType, backendKey,
                    backendName);
        return nullptr;
    }
    if (exporterValue.find("initConfig") == exporterValue.end()) {
        YRLOG_ERROR("parameter ip is invalid, exporter {} for backend {} of {}", httpExporterType, backendKey,
                    backendName);
        return nullptr;
    }
    std::string initConfig;
    if (exporterValue.find("initConfig") != exporterValue.end()) {
        auto initConfigJson = exporterValue.at("initConfig");
        initConfigJson["jobName"] = "runtime";
        if (initConfigJson.find("ip") != initConfigJson.end() && initConfigJson.find("port") != initConfigJson.end()) {
            initConfigJson["endpoint"] =
                initConfigJson.at("ip").get<std::string>() + ":" + std::to_string(initConfigJson.at("port").get<int>());
        }
        try {
            // print before set ssl config, which can't be printed
            YRLOG_INFO("metrics http exporter for backend {}, initConfig: {}", backendName, initConfigJson.dump());
        } catch (std::exception &e) {
            YRLOG_ERROR("dump initConfigJson failed, error: {}", e.what());
        }
        if (Config::Instance().YR_SSL_ENABLE()) {
            initConfigJson["isSSLEnable"] = true;
            initConfigJson["rootCertFile"] = Config::Instance().YR_SSL_ROOT_FILE();
            initConfigJson["certFile"] = Config::Instance().YR_SSL_CERT_FILE();
            initConfigJson["keyFile"] = Config::Instance().YR_SSL_KEY_FILE();
            auto ret = std::getenv(YR_SSL_PASSPHRASE_KEY);
            if (ret != nullptr) {
                initConfigJson["passphrase"] = ret;
                const int replaceOpt = 1;
                setenv(YR_SSL_PASSPHRASE_KEY, "", replaceOpt);
            } else {
                YRLOG_WARN("can not get metrics passphrase from env.");
            }
        }
        try {
            initConfig = initConfigJson.dump();
        } catch (std::exception &e) {
            YRLOG_ERROR("dump initConfigJson failed, error: {}", e.what());
        }
    }
    std::string error;
    return MetricsPlugin::LoadExporterFromLibrary(GetLibraryPath(httpExporterType), initConfig, error);
}

const MetricsSdk::ExportConfigs MetricsAdaptor::BuildExportConfigs(const nlohmann::json &exporterValue)
{
    try {
        YRLOG_DEBUG("Start to build export config {}", exporterValue.dump());
    } catch (std::exception &e) {
        YRLOG_ERROR("dump exporterValue failed, error: {}", e.what());
    }
    observability::sdk::metrics::ExportConfigs exportConfigs;
    if (exporterValue.contains("batchSize")) {
        exportConfigs.batchSize = exporterValue.at("batchSize");
    }
    if (exporterValue.contains("batchIntervalSec")) {
        exportConfigs.batchIntervalSec = exporterValue.at("batchIntervalSec");
    }
    if (exporterValue.contains("failureQueueMaxSize")) {
        exportConfigs.failureQueueMaxSize = exporterValue.at("failureQueueMaxSize");
    }
    if (exporterValue.contains("failureDataDir")) {
        exportConfigs.failureDataDir = exporterValue.at("failureDataDir");
    }
    if (exporterValue.contains("failureDataFileMaxCapacity")) {
        exportConfigs.failureDataFileMaxCapacity = exporterValue.at("failureDataFileMaxCapacity");
    }
    if (exporterValue.contains("enabledInstruments")) {
        for (auto &it : exporterValue.at("enabledInstruments").items()) {
            YRLOG_INFO("Enabled instrument: {}", it.value().dump());
            exportConfigs.enabledInstruments.insert(it.value().get<std::string>());
        }
    }
    return exportConfigs;
}

void MetricsAdaptor::CleanMetrics() noexcept
{
    std::shared_ptr<MetricsApi::NullMeterProvider> null = nullptr;
    MetricsApi::Provider::SetMeterProvider(null);
}

ErrorInfo MetricsAdaptor::SetUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    if (userEnable_ && Initialized_) {
        return ReportUInt64Counter(data);
    }
    return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                     "not enable metrics");
}

ErrorInfo MetricsAdaptor::ResetUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    if (userEnable_ && Initialized_) {
        return DoResetUInt64Counter(data);
    }
    return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                     "not enable metrics");
}

ErrorInfo MetricsAdaptor::IncreaseUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    if (userEnable_ && Initialized_) {
        return DoIncreaseUInt64Counter(data);
    }
    return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                     "not enable metrics");
}

std::pair<ErrorInfo, uint64_t> MetricsAdaptor::GetValueUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    if (userEnable_ && Initialized_) {
        return DoGetValueUInt64Counter(data);
    }
    return std::make_pair(ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                    YR::Libruntime::ModuleCode::RUNTIME, "not enable metrics"),
                          0);
}

ErrorInfo MetricsAdaptor::ReportUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    std::lock_guard<std::mutex> l(uint64_counter_mutex_);
    auto err = InitUInt64Counter(data);
    if (!err.OK()) {
        return err;
    }
    MetricsSdk::PointLabels labels;
    for (auto iter = data.labels.begin(); iter != data.labels.end(); ++iter) {
        labels.emplace_back(*iter);
    }
    uInt64CounterMap_.find(data.name)->second->Set(data.value, labels);
    YRLOG_DEBUG("finished set uint64 counter value {}", data.value);
    return ErrorInfo();
}

ErrorInfo MetricsAdaptor::DoResetUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    std::lock_guard<std::mutex> l(uint64_counter_mutex_);
    auto err = InitUInt64Counter(data);
    if (!err.OK()) {
        return err;
    }
    uInt64CounterMap_.find(data.name)->second->Reset();
    YRLOG_DEBUG("finished reset uint64 counter, name {}", data.name);
    return ErrorInfo();
}

ErrorInfo MetricsAdaptor::DoIncreaseUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    std::lock_guard<std::mutex> l(uint64_counter_mutex_);
    auto err = InitUInt64Counter(data);
    if (!err.OK()) {
        return err;
    }
    auto iter = uInt64CounterMap_.find(data.name);
    if (iter != uInt64CounterMap_.end()) {
        iter->second->Increment(data.value);
        YRLOG_DEBUG("finished increase uint64 counter value {}", data.value);
    }
    return ErrorInfo();
}

std::pair<ErrorInfo, uint64_t> MetricsAdaptor::DoGetValueUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    std::lock_guard<std::mutex> l(uint64_counter_mutex_);
    auto err = InitUInt64Counter(data);
    if (!err.OK()) {
        return std::make_pair(err, 0);
    }
    auto val = uInt64CounterMap_.find(data.name)->second->GetValue();
    YRLOG_DEBUG("finished get value: {} of uint64 counter value", val);
    return std::make_pair(ErrorInfo(), val);
}

ErrorInfo MetricsAdaptor::InitUInt64Counter(const YR::Libruntime::UInt64CounterData &data)
{
    if (uInt64CounterMap_.find(data.name) != uInt64CounterMap_.end()) {
        return ErrorInfo();
    }
    auto provider = MetricsApi::Provider::GetMeterProvider();
    if (provider == nullptr) {
        YRLOG_ERROR("Metrics provider is null");
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "there is no metrics provider ");
    }
    std::shared_ptr<MetricsApi::Meter> meter = provider->GetMeter("uinit64_counter_meter");
    if (meter == nullptr) {
        YRLOG_ERROR("Metrics meter is null");
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "there is no metrics meter");
    }
    auto uInt64Counter = meter->CreateUInt64Counter(data.name, data.description, data.unit);
    uInt64CounterMap_[data.name] = std::move(uInt64Counter);
    return ErrorInfo();
}

ErrorInfo MetricsAdaptor::SetDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    if (userEnable_ && Initialized_) {
        return ReportDoubleCounter(data);
    }
    return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                     "not enable metrics");
}

ErrorInfo MetricsAdaptor::ResetDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    if (userEnable_ && Initialized_) {
        return DoResetDoubleCounter(data);
    }
    return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                     "not enable metrics");
}

ErrorInfo MetricsAdaptor::IncreaseDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    if (userEnable_ && Initialized_) {
        return DoIncreaseDoubleCounter(data);
    }
    return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                     "not enable metrics");
}

std::pair<ErrorInfo, double> MetricsAdaptor::GetValueDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    if (userEnable_ && Initialized_) {
        return DoGetValueDoubleCounter(data);
    }
    return std::make_pair(ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR,
                                    YR::Libruntime::ModuleCode::RUNTIME, "not enable metrics"),
                          0);
}

ErrorInfo MetricsAdaptor::ReportDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    std::lock_guard<std::mutex> l(double_counter_mutex_);
    auto err = InitDoubleCounter(data);
    if (!err.OK()) {
        return err;
    }
    MetricsSdk::PointLabels labels;
    for (auto iter = data.labels.begin(); iter != data.labels.end(); ++iter) {
        labels.emplace_back(*iter);
    }
    auto it = doubleCounterMap_.find(data.name);
    if (it != doubleCounterMap_.end()) {
        it->second->Set(data.value, labels);
    }
    YRLOG_DEBUG("finished set double counter value {}", data.value);
    return ErrorInfo();
}

ErrorInfo MetricsAdaptor::DoResetDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    std::lock_guard<std::mutex> l(double_counter_mutex_);
    auto err = InitDoubleCounter(data);
    if (!err.OK()) {
        return err;
    }
    auto it = doubleCounterMap_.find(data.name);
    if (it != doubleCounterMap_.end()) {
        it->second->Reset();
    }
    YRLOG_DEBUG("finished reset double counter, name: {}", data.name);
    return ErrorInfo();
}

ErrorInfo MetricsAdaptor::DoIncreaseDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    std::lock_guard<std::mutex> l(double_counter_mutex_);
    auto err = InitDoubleCounter(data);
    if (!err.OK()) {
        return err;
    }
    auto it = doubleCounterMap_.find(data.name);
    if (it != doubleCounterMap_.end()) {
        it->second->Increment(data.value);
    }
    YRLOG_DEBUG("finished increase double counter value {}", data.value);
    return ErrorInfo();
}

std::pair<ErrorInfo, double> MetricsAdaptor::DoGetValueDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    std::lock_guard<std::mutex> l(double_counter_mutex_);
    auto err = InitDoubleCounter(data);
    if (!err.OK()) {
        return std::make_pair(err, 0);
    }
    auto it = doubleCounterMap_.find(data.name);
    if (it != doubleCounterMap_.end()) {
        auto val = it->second->GetValue();
        YRLOG_DEBUG("finished get value: {} of double counter", val);
        return std::make_pair(ErrorInfo(), val);
    }
    return std::make_pair(ErrorInfo(), 0);
}

ErrorInfo MetricsAdaptor::InitDoubleCounter(const YR::Libruntime::DoubleCounterData &data)
{
    if (doubleCounterMap_.find(data.name) != doubleCounterMap_.end()) {
        return ErrorInfo();
    }
    auto provider = MetricsApi::Provider::GetMeterProvider();
    if (provider == nullptr) {
        YRLOG_ERROR("Metrics provider is null");
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "there is no metrics provider ");
    }
    std::shared_ptr<MetricsApi::Meter> meter = provider->GetMeter("double_counter_meter");
    if (meter == nullptr) {
        YRLOG_ERROR("Metrics meter is null");
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "there is no metrics meter");
    }
    auto doubleCounter = meter->CreateDoubleCounter(data.name, data.description, data.unit);
    doubleCounterMap_[data.name] = std::move(doubleCounter);
    return ErrorInfo();
}

ErrorInfo MetricsAdaptor::ReportMetrics(const YR::Libruntime::GaugeData &gauge)
{
    if (Initialized_) {
        std::list<std::string> contextAttrs = {"node_id", "ip"};
        return ReportDoubleGauge(gauge, contextAttrs);
    }
    return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                     "not enable metrics");
}

ErrorInfo MetricsAdaptor::ReportGauge(const YR::Libruntime::GaugeData &gauge)
{
    if (userEnable_ && Initialized_) {
        std::list<std::string> contextAttrs = {"node_id", "ip"};
        return ReportDoubleGauge(gauge, contextAttrs);
    }
    return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                     "not enable metrics");
}

ErrorInfo MetricsAdaptor::ReportDoubleGauge(const YR::Libruntime::GaugeData &gauge,
                                            const std::list<std::string> &contextAttrs)
{
    std::unique_lock<std::mutex> l(gauge_mutex_);
    auto err = InitDoubleGauge(gauge);
    if (!err.OK()) {
        l.unlock();
        return err;
    }
    if (doubleGaugeMap_.find(gauge.name) == doubleGaugeMap_.end()) {
        l.unlock();
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "can not find gauge name");
    }
    auto it = doubleGaugeMap_.find(gauge.name);
    l.unlock();
    MetricsSdk::PointLabels labels;
    for (auto iter = gauge.labels.begin(); iter != gauge.labels.end(); ++iter) {
        labels.emplace_back(*iter);
    }
    it->second->Set(gauge.value, labels);
    YRLOG_DEBUG("finished set gauge value {}", gauge.value);
    return ErrorInfo();
}

ErrorInfo MetricsAdaptor::InitDoubleGauge(const YR::Libruntime::GaugeData &gauge)
{
    if (doubleGaugeMap_.find(gauge.name) != doubleGaugeMap_.end()) {
        return ErrorInfo();
    }
    auto provider = MetricsApi::Provider::GetMeterProvider();
    if (provider == nullptr) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "there is no metrics provider ");
    }
    std::shared_ptr<MetricsApi::Meter> meter = provider->GetMeter("gauge_meter");
    if (meter == nullptr) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "there is no metrics meter");
    }
    auto gaugeData = meter->CreateDoubleGauge(gauge.name, gauge.description, gauge.unit);
    doubleGaugeMap_[gauge.name] = std::move(gaugeData);
    return ErrorInfo();
}

ErrorInfo MetricsAdaptor::SetAlarm(const std::string &name, const std::string &description,
                                   const YR::Libruntime::AlarmInfo &alarmInfo)
{
    if (userEnable_ && Initialized_) {
        return ReportAlarm(name, description, alarmInfo);
    }
    YRLOG_ERROR("failed to set alarm, userEnable: {}, initialized: {}", userEnable_, Initialized_);
    return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                     "not enable metrics");
}

ErrorInfo MetricsAdaptor::ReportAlarm(const std::string &name, const std::string &description,
                                      const YR::Libruntime::AlarmInfo &alarmInfo)
{
    std::lock_guard<std::mutex> l(alarm_mutex_);
    auto err = InitAlarm(name, description);
    if (!err.OK()) {
        return err;
    }
    if (alarmMap_.find(name) == alarmMap_.end()) {
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "can not find alarm name");
    }
    MetricsApi::AlarmInfo metricsAlarmInfo;
    metricsAlarmInfo.id = alarmInfo.id;
    metricsAlarmInfo.alarmName = alarmInfo.alarmName;
    metricsAlarmInfo.alarmSeverity = static_cast<MetricsApi::AlarmSeverity>(alarmInfo.alarmSeverity);
    metricsAlarmInfo.locationInfo = alarmInfo.locationInfo;
    metricsAlarmInfo.cause = alarmInfo.cause;
    metricsAlarmInfo.startsAt =
        alarmInfo.startsAt == DEFAULT_ALARM_TIMESTAMP
            ? std::chrono::duration_cast<std::chrono::seconds>(std::chrono::system_clock::now().time_since_epoch())
                  .count()
            : alarmInfo.startsAt;
    metricsAlarmInfo.endsAt = alarmInfo.endsAt;
    metricsAlarmInfo.timeout = alarmInfo.timeout;
    for (auto &it : alarmInfo.customOptions) {
        metricsAlarmInfo.customOptions[it.first] = it.second;
    }
    alarmMap_.find(name)->second->Set(metricsAlarmInfo);
    YRLOG_DEBUG("finished set alarm name {}, location info: {}, cause: {}", metricsAlarmInfo.alarmName,
                metricsAlarmInfo.locationInfo, metricsAlarmInfo.cause);
    return ErrorInfo();
}

ErrorInfo MetricsAdaptor::InitAlarm(const std::string &name, const std::string &description)
{
    if (alarmMap_.find(name) != alarmMap_.end()) {
        return ErrorInfo();
    }
    auto provider = MetricsApi::Provider::GetMeterProvider();
    if (provider == nullptr) {
        YRLOG_ERROR("Metrics provider is null");
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "there is no metrics provider ");
    }
    std::shared_ptr<MetricsApi::Meter> meter = provider->GetMeter("alarm_meter");
    if (meter == nullptr) {
        YRLOG_ERROR("Metrics meter is null");
        return ErrorInfo(YR::Libruntime::ErrorCode::ERR_INNER_SYSTEM_ERROR, YR::Libruntime::ModuleCode::RUNTIME,
                         "there is no metrics meter");
    }
    auto alarm = meter->CreateAlarm(name, description);
    alarmMap_[name] = std::move(alarm);
    return ErrorInfo();
}

std::shared_ptr<MetricsExporters::Exporter> MetricsAdaptor::InitFileExporter(
    const std::string &backendKey, const std::string &backendName, const nlohmann::json &exporterValue,
    const std::function<std::string(std::string)> &getFileName)
{
    YRLOG_DEBUG("add exporter {} for backend {} of {}", FILE_EXPORTER, backendKey, backendName);
    if (exporterValue.find("enable") == exporterValue.end() || !exporterValue.at("enable").get<bool>()) {
        YRLOG_DEBUG("metrics exporter {} for backend {} of {} is not enabled", FILE_EXPORTER, backendKey, backendName);
        return nullptr;
    }
    std::string initConfig;
    if (exporterValue.find("initConfig") != exporterValue.end()) {
        auto initConfigJson = exporterValue.at("initConfig");
        if (initConfigJson.find("fileDir") == initConfigJson.end() ||
            initConfigJson.at("fileDir").get<std::string>().empty()) {
            YRLOG_DEBUG("not find the metrics exporter file path, use the log path: {}", GetContextValue("log_dir"));
            initConfigJson["fileDir"] = GetContextValue("log_dir");
        }
        if (!YR::utility::ExistPath(initConfigJson.at("fileDir")) &&
            !YR::utility::Mkdir(initConfigJson.at("fileDir"))) {
            YRLOG_ERROR("failed to mkdir{} for exporter {} for backend {} of {}", initConfigJson.at("fileDir").dump(),
                        FILE_EXPORTER, backendKey, backendName);
            return nullptr;
        }
        initConfigJson["fileName"] = getFileName(backendName);
        try {
            initConfig = initConfigJson.dump();
        } catch (std::exception &e) {
            YRLOG_ERROR("dump initConfigJson failed, error: {}", e.what());
            return nullptr;
        }
    }
    YRLOG_INFO("metrics exporter {} for backend {} of {}, init config: {}", FILE_EXPORTER, backendKey, backendName,
               initConfig);
    std::string error;
    return MetricsPlugin::LoadExporterFromLibrary(GetLibraryPath(FILE_EXPORTER), initConfig, error);
}

std::string MetricsAdaptor::GetMetricsFilesName(const std::string &backendName)
{
    // file reporting is not supported currently,this function is not implemented.
    if (backendName == "ds_alarm") {
        return backendName + ".alarm.dat";
    }
    return backendName + "-metrics.data";
}
}  // namespace Libruntime
}  // namespace YR
