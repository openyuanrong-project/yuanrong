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

#include "api/cpp/src/runtime_env_parse.h"
#include <gtest/gtest.h>
#include <filesystem>
#include <fstream>
#include "src/dto/invoke_options.h"
#include "common/common.h"

namespace fs = std::filesystem;
namespace YR {
namespace test {
class RuntimeEnvParseTest : public ::testing::Test {
protected:
    void SetUp() override
    {
        // 创建临时测试用的YAML文件
        validYamlPath_ = fs::temp_directory_path() / "valid_env.yaml";
        std::ofstream yamlFile(validYamlPath_);
        yamlFile << "name: test_env\ndependencies:\n  - python=3.8\n  - pip\n  - pip:\n    - numpy\n";
        yamlFile.close();

        // 创建临时测试用的YAML文件
        validNoNameYamlPath_ = fs::temp_directory_path() / "no_name_valid_env.yaml";
        std::ofstream noNameYamlFile(validNoNameYamlPath_);
        noNameYamlFile << "dependencies:\n  - python=3.8\n  - pip\n  - pip:\n    - numpy\n";
        noNameYamlFile.close();

        // 设置必要的环境变量
        setenv("YR_CONDA_HOME", "/fake/conda/path", 1);
    }

    void TearDown() override
    {
        // 清理临时文件
        if (fs::exists(validYamlPath_)) {
            fs::remove(validYamlPath_);
        }

        // 清理临时文件
        if (fs::exists(validNoNameYamlPath_)) {
            fs::remove(validNoNameYamlPath_);
        }
    }

    fs::path validYamlPath_;
    fs::path validNoNameYamlPath_;
};

TEST_F(RuntimeEnvParseTest, ShouldThrowWhenCondaHomeNotSet)
{
    unsetenv("YR_CONDA_HOME");
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<nlohmann::json>("conda", {{"name", "test"}});

    EXPECT_THROW(YR::ParseRuntimeEnv(options, env), YR::Exception);
}

TEST_F(RuntimeEnvParseTest, ShouldProcessPipPackagesCorrectly)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<std::vector<std::string>>("pip", {"numpy", "pandas"});

    YR::ParseRuntimeEnv(options, env);
    EXPECT_EQ(options.createOptions["POST_START_EXEC"], "pip3 install numpy pandas");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectBothPipAndConda)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<std::vector<std::string>>("pip", {"numpy"});
    env.Set<std::string>("conda", "env_name");

    EXPECT_THROW(YR::ParseRuntimeEnv(options, env), YR::Exception);
}

TEST_F(RuntimeEnvParseTest, ShouldHandleWorkingDirCorrectly)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<std::string>("working_dir", "/tmp/test_dir");

    YR::ParseRuntimeEnv(options, env);
    EXPECT_EQ(options.workingDir, "/tmp/test_dir");
}

TEST_F(RuntimeEnvParseTest, ShouldMergeEnvVarsProperly)
{
    YR::Libruntime::InvokeOptions options;
    options.envVars["EXISTING"] = "original";

    YR::RuntimeEnv env;
    env.Set<std::map<std::string, std::string>>("env_vars", {{"NEW", "value"}, {"EXISTING", "new"}});

    YR::ParseRuntimeEnv(options, env);
    EXPECT_EQ(options.envVars["NEW"], "value");
    EXPECT_EQ(options.envVars["EXISTING"], "original");  // 应保留原有值
}

TEST_F(RuntimeEnvParseTest, ShouldProcessCondaYamlFile)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<std::string>("conda", validYamlPath_.string());

    YR::ParseRuntimeEnv(options, env);
    EXPECT_EQ(options.createOptions["CONDA_PREFIX"], "/fake/conda/path");
    EXPECT_EQ(options.createOptions["CONDA_COMMAND"], "conda env create -f env.yaml");
    EXPECT_EQ(options.createOptions["CONDA_CONFIG"], "{\"name\": \"test_env\", \"dependencies\": [\"python=3.8\", \"pip\", {\"pip\": [\"numpy\"]}]}");
}

TEST_F(RuntimeEnvParseTest, ShouldProcessCondaNoNameYamlFile)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<std::string>("conda", validNoNameYamlPath_.string());

    YR::ParseRuntimeEnv(options, env);
    EXPECT_EQ(options.createOptions["CONDA_PREFIX"], "/fake/conda/path");
    EXPECT_EQ(options.createOptions["CONDA_COMMAND"], "conda env create -f env.yaml");
    EXPECT_FALSE(options.createOptions["CONDA_DEFAULT_ENV"].empty());
}

TEST_F(RuntimeEnvParseTest, ShouldProcessCondaEnvName)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<std::string>("conda", "existing_env");

    YR::ParseRuntimeEnv(options, env);
    EXPECT_EQ(options.createOptions["CONDA_COMMAND"], "conda activate existing_env");
    EXPECT_EQ(options.createOptions["CONDA_DEFAULT_ENV"], "existing_env");
}

TEST_F(RuntimeEnvParseTest, ShouldProcessCondaJsonConfig)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json config = {{"name", "json_env"}, {"dependencies", {"python=3.8", "numpy"}}};
    env.Set<nlohmann::json>("conda", config);

    YR::ParseRuntimeEnv(options, env);
    EXPECT_EQ(options.createOptions["CONDA_COMMAND"], "conda env create -f env.yaml");
    EXPECT_TRUE(options.createOptions["CONDA_DEFAULT_ENV"].find("json_env") != std::string::npos);
}

TEST_F(RuntimeEnvParseTest, ShouldRejectInvalidYamlFile)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<std::string>("conda", "/nonexistent/path.yaml");

    EXPECT_THROW(YR::ParseRuntimeEnv(options, env), YR::Exception);
}

TEST_F(RuntimeEnvParseTest, ShouldRejectInvalidPipType)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<std::string>("pip", "this_should_be_array");

    EXPECT_THROW(YR::ParseRuntimeEnv(options, env), YR::Exception);
}

TEST_F(RuntimeEnvParseTest, ShouldGenerateRandomNameForEmptyCondaName)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json config = {{"name", ""}, {"dependencies", {"python"}}};
    env.Set<nlohmann::json>("conda", config);

    YR::ParseRuntimeEnv(options, env);
    EXPECT_FALSE(options.createOptions["CONDA_DEFAULT_ENV"].empty());
}

TEST_F(RuntimeEnvParseTest, ShouldAddSharedDirInCreateOpt)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json config = {{"name", "abc"}, {"TTL", 5}};
    env.Set<nlohmann::json>("shared_dir", config);

    YR::ParseRuntimeEnv(options, env);
    EXPECT_EQ(options.createOptions["DELEGATE_SHARED_DIRECTORY"], "abc");
    EXPECT_EQ(options.createOptions["DELEGATE_SHARED_DIRECTORY_TTL"], "5");

    YR::Libruntime::InvokeOptions options2;
    YR::RuntimeEnv env2;
    nlohmann::json config2 = {{"name", "abc"}};
    env2.Set<nlohmann::json>("shared_dir", config2);

    YR::ParseRuntimeEnv(options2, env2);
    EXPECT_EQ(options2.createOptions["DELEGATE_SHARED_DIRECTORY"], "abc");
    EXPECT_EQ(options2.createOptions["DELEGATE_SHARED_DIRECTORY_TTL"], "0");
}

TEST_F(RuntimeEnvParseTest, ShouldThrowExceptionWhenSharedDirConfigInvaild)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json config = {{"name", ""}};
    env.Set<nlohmann::json>("shared_dir", config);
    EXPECT_THROW(YR::ParseRuntimeEnv(options, env), YR::Exception);

    YR::Libruntime::InvokeOptions options2;
    YR::RuntimeEnv env2;
    env2.Set<nlohmann::json>("shared_dir", "str");
    EXPECT_THROW(YR::ParseRuntimeEnv(options2, env2), YR::Exception);
}

TEST_F(RuntimeEnvParseTest, ShouldRejectBothPipAndVenv)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<std::vector<std::string>>("pip", {"numpy"});
    nlohmann::json venvConfig = {{"name", "testVenv"}};
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(
            YR::ParseRuntimeEnv(options, env), 1001,
            "The 'venv' field and 'pip' field of runtime_env cannot both be specified.\n");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectBothCondaAndVenv)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<std::string>("conda", validYamlPath_.string());
    nlohmann::json venvConfig = {{"name", "testVenv"}};
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(
            YR::ParseRuntimeEnv(options, env), 1001,
            "The 'conda' field and 'venv' field of runtime_env cannot both be specified.");
}

TEST_F(RuntimeEnvParseTest, ShouldProcessVenvWithHost)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
            {"name", "123testVenv"},
            {"path", {
                    {"site_package_path", "obs://mini-kernel/site-package.zip"}
            }},
            {"dependencies", {
                    {"pypi", {"matplotlib", "msgpack-python == 1.0.5"}},
                    {"trusted_host", "mirrors.tools.nobody.com"},
                    {"index_url", "http://mirrors.tools.nobody.com/pypi/simple/"},
            }}
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    YR::ParseRuntimeEnv(options, env);
    EXPECT_EQ(options.createOptions["VIRTUALENV_KIND"], "venv");
    EXPECT_EQ(options.createOptions["VIRTUALENV_NAME"], "123testVenv");
    EXPECT_EQ(options.createOptions["VIRTUALENV_COMMAND"], "pip install \"matplotlib\" \"msgpack-python==1.0.5\" "
                                                           "-i http://mirrors.tools.nobody.com/pypi/simple/ --trusted-host mirrors.tools.nobody.com");
    EXPECT_EQ(options.createOptions["VIRTUALENV_PATH"], "{\"site_package_path\":\"obs://mini-kernel/site-package.zip\"}");
}

TEST_F(RuntimeEnvParseTest, ShouldProcessVenvWithoutHost)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
            {"name", "testVenv"},
            {"path", {
                    {"site_package_path", "obs://mini-kernel/site-package.zip"}
            }},
            {"dependencies", {
                    {"pypi", {"matplotlib  ", "msgpack-python==1.0.5"}}
            }}
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    YR::ParseRuntimeEnv(options, env);
    EXPECT_EQ(options.createOptions["VIRTUALENV_KIND"], "venv");
    EXPECT_EQ(options.createOptions["VIRTUALENV_NAME"], "testVenv");
    EXPECT_EQ(options.createOptions["VIRTUALENV_COMMAND"], "pip install \"matplotlib\" \"msgpack-python==1.0.5\"");
    EXPECT_EQ(options.createOptions["VIRTUALENV_PATH"], "{\"site_package_path\":\"obs://mini-kernel/site-package.zip\"}");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvNotJson)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    env.Set<std::string>("venv", "xxx");

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv format must be json");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvKeysInvalid)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {{"key", "value"}};
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv keys can only be name, dependencies, path");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvNameNotString)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {{"name", 123}};
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.name format must be string");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvNameInvalid)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {{"name", "!"}};
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.name is invalid");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvDependenciesNotJson)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {{"dependencies", "xxx"}};
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.dependencies format must be json");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvDependenciesKeysInvalid)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
            {"dependencies", {
                    {"key", "value"}
            }}
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.dependencies keys can only be pypi, trusted_host, index_url");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvDependenciesPypiNotArray)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
            {"dependencies", {
                    {"pypi", "xxx"}
            }}
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.dependencies.pypi format must be array");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvDependenciesTrustedHostNotString) {
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
            {"dependencies", {
                    {"pypi", {"matplotlib", "msgpack-python==1.0.5"}},
                    {"trusted_host", 123}
            }}
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.dependencies.trusted_host format must be string");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvDependenciesIndexURLNotString) {
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
        {"dependencies", {
                        {"pypi", {"matplotlib", "msgpack-python==1.0.5"}},
                        {"index_url", 123}
        }}
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.dependencies.index_url format must be string");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvDependenciesPypiItemNotString)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
            {"dependencies", {
                    {"pypi", {123}},
                    {"trusted_host", "mirrors.tools.nobody.com"}
            }}
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.dependencies.pypi item format must be string");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvDependenciesPypiItemEmpty)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
            {"dependencies", {
                    {"pypi", {"matplotlib", " "}},
                    {"trusted_host", "mirrors.tools.nobody.com"}
            }}
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.dependencies.pypi item is empty");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvPathNotJson)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {{"path", "xxx"}};
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.path format must be json");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvPathKeysInvalid)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
            {"path", {
                    {"key", "value"}
            }}
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.path keys can only be site_package_path");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvPathSitePackagePathNotString)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
            {"path", {
                    {"site_package_path", 123}
            }},
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.path.site_package_path format must be string");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvPypiItemContainInvalidOperator)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
            {"dependencies", {
                    {"pypi", {"requests=2.28.1"}}
            }}
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "venv.dependencies.pypi item contains illegal operator '='");
}

TEST_F(RuntimeEnvParseTest, ShouldProcessVenvWithPypiExtras)
{
    // extras包含一个元素
    YR::Libruntime::InvokeOptions options1;
    YR::RuntimeEnv env1;
    nlohmann::json venvConfig1 = {
            {"dependencies", {
                    {"pypi", {"requests[socks]"}}
            }}
    };
    env1.Set<nlohmann::json>("venv", venvConfig1);

    YR::ParseRuntimeEnv(options1, env1);
    EXPECT_EQ(options1.createOptions["VIRTUALENV_KIND"], "venv");
    EXPECT_EQ(options1.createOptions["VIRTUALENV_COMMAND"], "pip install \"requests[socks]\"");

    // extras包含多个元素
    YR::Libruntime::InvokeOptions options2;
    YR::RuntimeEnv env2;
    nlohmann::json venvConfig2 = {
            {"dependencies", {
                    {"pypi", {"django[argon2, bcrypt]"}}
            }}
    };
    env2.Set<nlohmann::json>("venv", venvConfig2);

    YR::ParseRuntimeEnv(options2, env2);
    EXPECT_EQ(options2.createOptions["VIRTUALENV_KIND"], "venv");
    EXPECT_EQ(options2.createOptions["VIRTUALENV_COMMAND"], "pip install \"django[argon2,bcrypt]\"");
}

TEST_F(RuntimeEnvParseTest, ShouldProcessVenvWithPypiVersion)
{
    // 精确版本
    YR::Libruntime::InvokeOptions options1;
    YR::RuntimeEnv env1;
    nlohmann::json venvConfig1 = {
            {"dependencies", {
                    {"pypi", {"requests==2.28.1"}}
            }}
    };
    env1.Set<nlohmann::json>("venv", venvConfig1);

    YR::ParseRuntimeEnv(options1, env1);
    EXPECT_EQ(options1.createOptions["VIRTUALENV_KIND"], "venv");
    EXPECT_EQ(options1.createOptions["VIRTUALENV_COMMAND"], "pip install \"requests==2.28.1\"");

    // 任意版本
    YR::Libruntime::InvokeOptions options2;
    YR::RuntimeEnv env2;
    nlohmann::json venvConfig2 = {
            {"dependencies", {
                    {"pypi", {"package===1.0.0"}}
            }}
    };
    env2.Set<nlohmann::json>("venv", venvConfig2);

    YR::ParseRuntimeEnv(options2, env2);
    EXPECT_EQ(options2.createOptions["VIRTUALENV_KIND"], "venv");
    EXPECT_EQ(options2.createOptions["VIRTUALENV_COMMAND"], "pip install \"package===1.0.0\"");

    // 范围约束
    YR::Libruntime::InvokeOptions options3;
    YR::RuntimeEnv env3;
    nlohmann::json venvConfig3 = {
            {"dependencies", {
                    {"pypi", {"requests>=2.0,<3.0"}}
            }}
    };
    env3.Set<nlohmann::json>("venv", venvConfig3);

    YR::ParseRuntimeEnv(options3, env3);
    EXPECT_EQ(options3.createOptions["VIRTUALENV_KIND"], "venv");
    EXPECT_EQ(options3.createOptions["VIRTUALENV_COMMAND"], "pip install \"requests>=2.0,<3.0\"");

    // extras和版本
    YR::Libruntime::InvokeOptions options4;
    YR::RuntimeEnv env4;
    nlohmann::json venvConfig4 = {
            {"dependencies", {
                    {"pypi", {"requests[socks]==2.28.1"}}
            }}
    };
    env4.Set<nlohmann::json>("venv", venvConfig4);

    YR::ParseRuntimeEnv(options4, env4);
    EXPECT_EQ(options4.createOptions["VIRTUALENV_KIND"], "venv");
    EXPECT_EQ(options4.createOptions["VIRTUALENV_COMMAND"], "pip install \"requests[socks]==2.28.1\"");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvPypiItemPkgNameInvalid)
{
    YR::Libruntime::InvokeOptions options;
    YR::RuntimeEnv env;
    nlohmann::json venvConfig = {
            {"dependencies", {
                    {"pypi", {"aaa123!@#-*1a"}}
            }}
    };
    env.Set<nlohmann::json>("venv", venvConfig);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options, env), 1001, "package name of venv.dependencies.pypi item is invalid");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvPypiItemExtrasInvalid)
{
    // 未包含']'
    YR::Libruntime::InvokeOptions options1;
    YR::RuntimeEnv env1;
    nlohmann::json venvConfig1 = {
            {"dependencies", {
                    {"pypi", {"requests[socks"}}
            }}
    };
    env1.Set<nlohmann::json>("venv", venvConfig1);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options1, env1), 1001, "extras of venv.dependencies.pypi item is invalid, does not contain ']'");

    // extras中的元素非法
    YR::Libruntime::InvokeOptions options2;
    YR::RuntimeEnv env2;
    nlohmann::json venvConfig2 = {
            {"dependencies", {
                    {"pypi", {"requests[socks!]"}}
            }}
    };
    env2.Set<nlohmann::json>("venv", venvConfig2);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options2, env2), 1001, "extras item of venv.dependencies.pypi item is invalid");

    // extras为空
    YR::Libruntime::InvokeOptions options3;
    YR::RuntimeEnv env3;
    nlohmann::json venvConfig3 = {
            {"dependencies", {
                    {"pypi", {"requests[]"}}
            }}
    };
    env3.Set<nlohmann::json>("venv", venvConfig3);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options3, env3), 1001, "extras of venv.dependencies.pypi item is empty");

    // extras中的部分元素为空
    YR::Libruntime::InvokeOptions options4;
    YR::RuntimeEnv env4;
    nlohmann::json venvConfig4 = {
            {"dependencies", {
                    {"pypi", {"requests[socks,,xxx]"}}
            }}
    };
    env4.Set<nlohmann::json>("venv", venvConfig4);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options4, env4), 1001, "extras item of venv.dependencies.pypi item is invalid");
}

TEST_F(RuntimeEnvParseTest, ShouldRejectVenvPypiItemVersionEmpty)
{
    // 第1个版本为空
    YR::Libruntime::InvokeOptions options1;
    YR::RuntimeEnv env1;
    nlohmann::json venvConfig1 = {
            {"dependencies", {
                    {"pypi", {"requests>=,<3.0"}}
            }}
    };
    env1.Set<nlohmann::json>("venv", venvConfig1);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options1, env1), 1001, "version of venv.dependencies.pypi item is empty");

    // 第2个版本为空
    YR::Libruntime::InvokeOptions options2;
    YR::RuntimeEnv env2;
    nlohmann::json venvConfig2 = {
            {"dependencies", {
                    {"pypi", {"requests>=2.0,<"}}
            }}
    };
    env2.Set<nlohmann::json>("venv", venvConfig2);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options2, env2), 1001, "version of venv.dependencies.pypi item is empty");

    // 版本为空
    YR::Libruntime::InvokeOptions options3;
    YR::RuntimeEnv env3;
    nlohmann::json venvConfig3 = {
            {"dependencies", {
                    {"pypi", {"requests=="}}
            }}
    };
    env3.Set<nlohmann::json>("venv", venvConfig3);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options3, env3), 1001, "version of venv.dependencies.pypi item is empty");

    // 带extras，版本为空
    YR::Libruntime::InvokeOptions options4;
    YR::RuntimeEnv env4;
    nlohmann::json venvConfig4 = {
            {"dependencies", {
                    {"pypi", {"requests[socks]!="}}
            }}
    };
    env4.Set<nlohmann::json>("venv", venvConfig4);

    EXPECT_THROW_WITH_CODE_AND_MSG(YR::ParseRuntimeEnv(options4, env4), 1001, "version of venv.dependencies.pypi item is empty");
}
}
}