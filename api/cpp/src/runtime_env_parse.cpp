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

#if __has_include(<filesystem>)
#include <filesystem>
#else
#include <experimental/filesystem>
#endif
#include <iostream>

#include <yaml-cpp/yaml.h>
#include <json.hpp>

#include "yr/api/exception.h"
#include "src/libruntime/err_type.h"
#include "api/cpp/src/utils/utils.h"
#include "yr/api/runtime_env.h"
#include "src/dto/invoke_options.h"
#include "api/cpp/src/runtime_env_parse.h"

#ifdef __cpp_lib_filesystem
namespace filesystem = std::filesystem;
#elif __cpp_lib_experimental_filesystem
namespace filesystem = std::experimental::filesystem;
#endif

namespace YR {
const std::string CONDA = "conda";
const std::string PIP = "pip";
const std::string YR_CONDA_HOME = "YR_CONDA_HOME";
const std::string WORKER_DIR = "working_dir";
const std::string ENV_VARS = "env_vars";
const std::string SHARED_DIR = "shared_dir";

const std::string POST_START_EXEC = "POST_START_EXEC";
const std::string CONDA_PREFIX = "CONDA_PREFIX";
const std::string CONDA_CONFIG = "CONDA_CONFIG";
const std::string CONDA_COMMAND = "CONDA_COMMAND";
const std::string CONDA_DEFAULT_ENV = "CONDA_DEFAULT_ENV";

const std::string VENV = "venv";
const std::string VENV_NAME = "name";
const std::string VENV_DEPENDENCIES = "dependencies";
const std::string VENV_PATH = "path";
const std::string VENV_DEPENDENCIES_PYPI = "pypi";
const std::string VENV_DEPENDENCIES_TRUSTED_HOST = "trusted_host";
const std::string VENV_DEPENDENCIES_INDEX_URL = "index_url";
const std::string VENV_PATH_SITE_PACKAGE_PATH = "site_package_path";
const std::unordered_set<std::string> VENV_ALLOWED_KEYS = {VENV_NAME, VENV_DEPENDENCIES, VENV_PATH};
const std::unordered_set<std::string> VENV_DEPENDENCIES_ALLOWED_KEYS = {
        VENV_DEPENDENCIES_PYPI, VENV_DEPENDENCIES_TRUSTED_HOST, VENV_DEPENDENCIES_INDEX_URL
};
const std::unordered_set<std::string> VENV_PATH_ALLOWED_KEYS = {VENV_PATH_SITE_PACKAGE_PATH};
const re2::RE2 VENV_NAME_REGEX(R"(^[0-9A-Za-z][0-9A-Za-z_-]{0,35}$)");

const std::vector<std::string> VENV_PYPI_VALID_OPERATORS = {"===", "==", "~=", ">=", "<=", ">", "<", "!="};
const re2::RE2 VENV_PYPI_PKG_NAME_EXTRAS_REGEX(R"(^[a-zA-Z0-9](?:[a-zA-Z0-9._-]*[a-zA-Z0-9])?$)");
const re2::RE2 VENV_PYPI_VERSION_OPERATORS_REGEX(R"((?:===|==|~=|>=|<=|!=|>|<)([0-9a-zA-Z._+-]*)(?:$|,))");

const std::string VIRTUALENV_KIND = "VIRTUALENV_KIND";
const std::string VIRTUALENV_NAME = "VIRTUALENV_NAME";
const std::string VIRTUALENV_COMMAND = "VIRTUALENV_COMMAND";
const std::string VIRTUALENV_PATH = "VIRTUALENV_PATH";

const std::vector<std::pair<std::string, std::string>> RUNTIME_ENV_MUTEX_PAIRS = {
        {CONDA, PIP},
        {VENV, PIP},
        {CONDA, VENV},
};

std::string GetCondaBinExecutable()
{
    if (auto envStr = GetEnv(YR_CONDA_HOME); !envStr.empty()) {
        return envStr;
    }
    if (auto envStr = GetEnv(CONDA_PREFIX); !envStr.empty()) {
        return envStr;
    }
    throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                        "please configure YR_CONDA_HOME environment variable which contain a bin subdirectory");
}

std::string YamlToJson(YAML::Node &node)
{
    YAML::Emitter emitter;
    emitter << YAML::DoubleQuoted << YAML::Flow << YAML::BeginSeq << node;

    return std::string(emitter.c_str() + 1);
}
void HandleCondaConfig(YR::Libruntime::InvokeOptions& invokeOptions, const nlohmann::json& condaConfig);

void HandleVenvConfig(YR::Libruntime::InvokeOptions& invokeOptions, const nlohmann::json& venvConfig);

void HandleSharedDirConfig(YR::Libruntime::InvokeOptions &invokeOptions, const nlohmann::json &sharedDirConfig)
{
    if (sharedDirConfig.is_object()) {
        // 处理JSON对象类型的conda配置
        std::string name = sharedDirConfig.value("name", "");
        if (name.empty()) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "shared dir name must be string");
        }
        int ttl = sharedDirConfig.value("TTL", 0);
        invokeOptions.createOptions["DELEGATE_SHARED_DIRECTORY"] = name;
        invokeOptions.createOptions["DELEGATE_SHARED_DIRECTORY_TTL"] = std::to_string(ttl);
    } else {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID, "shared dir format must be json");
    }
}

void CheckRuntimeEnvMutexPairs(const YR::RuntimeEnv& runtimeEnv)
{
    for (const auto& [field1, field2] : RUNTIME_ENV_MUTEX_PAIRS) {
        if (runtimeEnv.Contains(field1) && runtimeEnv.Contains(field2)) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                "The '" + field1 + "' field and '" + field2 + "' field of runtime_env "
                                                                              "cannot both be specified.\n");
        }
    }
}

void ParseRuntimeEnv(YR::Libruntime::InvokeOptions& invokeOptions, const YR::RuntimeEnv& runtimeEnv)
{
    if (runtimeEnv.Empty()) {
        return;
    }

    CheckRuntimeEnvMutexPairs(runtimeEnv);

    if (runtimeEnv.Contains(PIP)) {
        try {
            const auto &pipPackages = runtimeEnv.Get<std::vector<std::string>>(PIP);
            std::ostringstream  pipCommand;
            pipCommand << "pip3 install";
            for (size_t i = 0; i < pipPackages.size(); i++) {
                pipCommand << " " << pipPackages[i];
            }
            invokeOptions.createOptions[POST_START_EXEC] = pipCommand.str();
        }  catch (std::exception &e) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                std::string("Failed to parse pip field of RuntimeEnv: ") + e.what());
        }
    }
    if (runtimeEnv.Contains(WORKER_DIR)) {
        try {
            const std::string &workingDir = runtimeEnv.Get<std::string>(WORKER_DIR);
            invokeOptions.workingDir = workingDir;
        }  catch (std::exception &e) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                std::string("`working_dir` must be a string: ") + e.what());
        }
    }
    if (runtimeEnv.Contains(ENV_VARS)) {
        try {
            const auto& envVars = runtimeEnv.Get<std::map<std::string, std::string>>(ENV_VARS);
            for (const auto& pair : envVars) {
                if (invokeOptions.envVars.find(pair.first) == invokeOptions.envVars.end()) {
                    invokeOptions.envVars.insert(pair);
                }
            }
        }  catch (std::exception &e) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                std::string("`envs` must be a map[string: string]: ") + e.what());
        }
    }
    if (runtimeEnv.Contains(CONDA)) {
        nlohmann::json condaJson;
        try {
        condaJson = runtimeEnv.Get<nlohmann::json>(CONDA);
        } catch(std::exception &e) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                std::string("`get conda to nlohmann:json format failed: ") + e.what());
        }
        HandleCondaConfig(invokeOptions, condaJson);
    }
    if (runtimeEnv.Contains(SHARED_DIR)) {
        nlohmann::json sharedDirJson;
        try {
            sharedDirJson = runtimeEnv.Get<nlohmann::json>(SHARED_DIR);
            HandleSharedDirConfig(invokeOptions, sharedDirJson);
        } catch (std::exception &e) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                std::string("`get shared dir to nlohmann:json format failed: ") + e.what());
        }
    }
    if (runtimeEnv.Contains(VENV)) {
        nlohmann::json venvJson;
        try {
            venvJson = runtimeEnv.Get<nlohmann::json>(VENV);
        } catch(std::exception &e) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                std::string("get venv to nlohmann:json format failed: ") + e.what());
        }
        HandleVenvConfig(invokeOptions, venvJson);
    }
}

void HandleCondaConfig(YR::Libruntime::InvokeOptions& invokeOptions, const nlohmann::json& condaConfig)
{
    invokeOptions.createOptions[CONDA_PREFIX] = GetCondaBinExecutable();

    if (condaConfig.is_string()) {
        // 处理字符串类型的conda配置（YAML文件路径或环境名称）
        const std::string& condaStr = condaConfig.get<std::string>();
        const filesystem::path condaPath(condaStr);

        if (condaPath.extension() == ".yaml" || condaPath.extension() == ".yml") {
            if (!filesystem::exists(condaPath)) {
                throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                    std::string("Can't find conda YAML file ") + condaStr);
            }

            YAML::Node yamlNode = YAML::LoadFile(condaStr);
            std::string envName;
            try {
                envName = yamlNode["name"].as<std::string>();
            } catch (std::exception& e) {
                // 支持传空，以及类型异常处理，yaml-cpp高低版本之间判断是否是string类型的函数不一致，使用try catch保持一致
            }
            if (envName.empty()) {
                envName = "virtual_env-" + YR::utility::IDGenerator::GenRequestId();
            }

            invokeOptions.createOptions[CONDA_CONFIG] = YamlToJson(yamlNode);
            invokeOptions.createOptions[CONDA_COMMAND] = "conda env create -f env.yaml";
            invokeOptions.createOptions[CONDA_DEFAULT_ENV] = envName;
        } else {
            // 直接使用环境名称
            invokeOptions.createOptions[CONDA_COMMAND] = "conda activate " + condaStr;
            invokeOptions.createOptions[CONDA_DEFAULT_ENV] = condaStr;
        }
    } else if (condaConfig.is_object()) {
        // 处理JSON对象类型的conda配置
        std::string envName = condaConfig.value("name", "");
        if (envName.empty()) {
            envName = "virtual_env-" + YR::utility::IDGenerator::GenRequestId();
        }

        invokeOptions.createOptions[CONDA_CONFIG] = condaConfig.dump();
        invokeOptions.createOptions[CONDA_COMMAND] = "conda env create -f env.yaml";
        invokeOptions.createOptions[CONDA_DEFAULT_ENV] = envName;
    } else {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "conda format must be string or json");
    }
}

bool IsKeysValid(const nlohmann::json& obj, const std::unordered_set<std::string>& allowedKeys)
{
    for (const auto& item : obj.items()) {
        if (allowedKeys.find(item.key()) == allowedKeys.end()) {
            return false;
        }
    }
    return true;
}

std::string ParseVenvName(const nlohmann::json& nameObj)
{
    // name必须是字符串
    if (!nameObj.is_string()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "venv.name format must be string");
    }

    std::string name = nameObj.get<std::string>();
    if (!name.empty()) {
        // 校验name格式
        if (!RE2::FullMatch(name, VENV_NAME_REGEX)) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                "venv.name is invalid");
        }
    }

    return name;
}

bool ContainIllegalOperator(const std::string& item)
{
    std::string tempItem = item;

    // 替换所有合法操作符为空格（避免误判）
    for (const auto& op : VENV_PYPI_VALID_OPERATORS) {
        size_t pos = 0;
        while ((pos = tempItem.find(op, pos)) != std::string::npos) {
            tempItem.replace(pos, op.length(), op.length(), ' ');
            pos += op.length();
        }
    }

    // 检查是否还有剩余的=
    return tempItem.find('=') != std::string::npos;
}

void CheckPypiPkgName(const std::string& item)
{
    std::string pkgName = item;

    // 移除extras
    size_t bracketPos = pkgName.find('[');
    if (bracketPos != std::string::npos) {
        pkgName = pkgName.substr(0, bracketPos);
    } else {
        // 移除版本和运算符
        size_t opPos = std::string::npos;
        for (const auto& op : VENV_PYPI_VALID_OPERATORS) {
            size_t pos = pkgName.find(op);
            if (pos != std::string::npos && (opPos == std::string::npos || pos < opPos)) {
                opPos = pos;
            }
        }
        if (opPos != std::string::npos) {
            pkgName = pkgName.substr(0, opPos);
        }
    }

    if (!RE2::FullMatch(pkgName, VENV_PYPI_PKG_NAME_EXTRAS_REGEX)) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "package name of venv.dependencies.pypi item is invalid");
    }
}

void CheckPypiExtras(const std::string& item)
{
    std::string extras = item;

    size_t bracketStartPos = extras.find("[");
    if (bracketStartPos == std::string::npos) {
        return;
    }

    size_t bracketEndPos = extras.find("]");
    // extras格式错误，未包含']'
    if (bracketEndPos == std::string::npos) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "extras of venv.dependencies.pypi item is invalid, does not contain ']'");
    }

    extras = extras.substr(bracketStartPos + 1, bracketEndPos - bracketStartPos - 1);
    if (extras.empty()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "extras of venv.dependencies.pypi item is empty");
    }

    // 校验extras中的元素
    std::stringstream ss(extras);
    std::string extra;
    while (std::getline(ss, extra, ',')) {
        if (!RE2::FullMatch(extra, VENV_PYPI_PKG_NAME_EXTRAS_REGEX)) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                "extras item of venv.dependencies.pypi item is invalid");
        }
    }
}

void CheckPypiVersion(const std::string& item)
{
    re2::StringPiece input(item);
    std::string version;

    while (RE2::FindAndConsume(&input, VENV_PYPI_VERSION_OPERATORS_REGEX, &version)) {
        if (version.empty()) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                "version of venv.dependencies.pypi item is empty");
        }
    }
}

std::string ParsePypiItem(const nlohmann::json& itemObj)
{
    // dependencies.pypi中的元素必须是字符串
    if (!itemObj.is_string()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "venv.dependencies.pypi item format must be string");
    }

    std::string item = itemObj.get<std::string>();

    // 删除item中的空格
    item.erase(std::remove(item.begin(), item.end(), ' '), item.end());
    if (item.empty()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "venv.dependencies.pypi item is empty");
    }

    // item不能包含非法运算符'='
    if (ContainIllegalOperator(item)) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "venv.dependencies.pypi item contains illegal operator '='");
    }

    // 校验item中的包名
    CheckPypiPkgName(item);

    // 校验item中的extras
    CheckPypiExtras(item);

    // 校验item中的version
    CheckPypiVersion(item);

    return item;
}

std::string ParseVenvDependencies(const nlohmann::json& dependenciesObj)
{
    // dependencies必须是json
    if (!dependenciesObj.is_object()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "venv.dependencies format must be json");
    }

    // dependencies的key只能是pypi, host
    if (!IsKeysValid(dependenciesObj, VENV_DEPENDENCIES_ALLOWED_KEYS)) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "venv.dependencies keys can only be pypi, trusted_host, index_url");
    }

    if (dependenciesObj.contains(VENV_DEPENDENCIES_PYPI)) {
        // dependencies.pypi必须是数组
        const auto& pypi = dependenciesObj[VENV_DEPENDENCIES_PYPI];
        if (!pypi.is_array()) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                "venv.dependencies.pypi format must be array");
        }

        if (pypi.empty()) {
            return "";
        }

        std::string trustedHost = "";
        if (dependenciesObj.contains(VENV_DEPENDENCIES_TRUSTED_HOST)) {
            // dependencies.trusted_host必须是字符串
            const auto& hostObj = dependenciesObj[VENV_DEPENDENCIES_TRUSTED_HOST];
            if (!hostObj.is_string()) {
                throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                    "venv.dependencies.trusted_host format must be string");
            }

            trustedHost = hostObj.get<std::string>();
        }

        std::string indexURL = "";
        if (dependenciesObj.contains(VENV_DEPENDENCIES_INDEX_URL)) {
            // dependencies.index_url必须是字符串
            const auto& indexURLObj = dependenciesObj[VENV_DEPENDENCIES_INDEX_URL];
            if (!indexURLObj.is_string()) {
                throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                    "venv.dependencies.index_url format must be string");
            }
            indexURL = indexURLObj.get<std::string>();
        }

        std::string command = "pip install";
        for (const auto& itemObj : pypi) {
            // 把dependencies.pypi中的元素拼接为pip install命令
            command += " \"" + ParsePypiItem(itemObj) + "\"";
        }

        if (!indexURL.empty()) {
            // 拼接依赖下载的镜像源
            command += " -i " + indexURL;
        }

        if (!trustedHost.empty()) {
            // 拼接依赖下载的镜像源
            command += " --trusted-host " + trustedHost;
        }

        return command;
    }

    return "";
}

std::string ParseVenvPath(const nlohmann::json& pathObj)
{
    // path必须是json
    if (!pathObj.is_object()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "venv.path format must be json");
    }

    // path的key只能是site_package_path
    if (!IsKeysValid(pathObj, VENV_PATH_ALLOWED_KEYS)) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "venv.path keys can only be site_package_path");
    }

    if (pathObj.contains(VENV_PATH_SITE_PACKAGE_PATH)) {
        // path.site_package_path必须是字符串
        const auto& sitePackagePathObj = pathObj[VENV_PATH_SITE_PACKAGE_PATH];
        if (!sitePackagePathObj.is_string()) {
            throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                                "venv.path.site_package_path format must be string");
        }

        std::string sitePackagePath = sitePackagePathObj.get<std::string>();
        if (sitePackagePath.empty()) {
            return "";
        }

        return pathObj.dump();
    }

    return "";
}

void HandleVenvConfig(YR::Libruntime::InvokeOptions& invokeOptions, const nlohmann::json& venvConfig)
{
    invokeOptions.createOptions[VIRTUALENV_KIND] = VENV;

    // venv必须是json对象
    if (!venvConfig.is_object()) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "venv format must be json");
    }

    // venv的key只能是name, dependencies, path
    if (!IsKeysValid(venvConfig, VENV_ALLOWED_KEYS)) {
        throw YR::Exception(YR::Libruntime::ErrorCode::ERR_PARAM_INVALID,
                            "venv keys can only be name, dependencies, path");
    }

    if (venvConfig.contains(VENV_NAME)) {
        std::string name = ParseVenvName(venvConfig[VENV_NAME]);
        if (!name.empty()) {
            invokeOptions.createOptions[VIRTUALENV_NAME] = name;
        }
    }
    if (venvConfig.contains(VENV_DEPENDENCIES)) {
        std::string command = ParseVenvDependencies(venvConfig[VENV_DEPENDENCIES]);
        if (!command.empty()) {
            invokeOptions.createOptions[VIRTUALENV_COMMAND] = command;
        }
    }
    if (venvConfig.contains(VENV_PATH)) {
        std::string path = ParseVenvPath(venvConfig[VENV_PATH]);
        if (!path.empty()) {
            invokeOptions.createOptions[VIRTUALENV_PATH] = path;
        }
    }
}
}
