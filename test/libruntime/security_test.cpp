/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

#include "src/libruntime/utils/security.h"
#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <stdio.h>
#include <future>
#include <thread>
#include "src/dto/config.h"
#include "src/libruntime/fsclient/protobuf/common.grpc.pb.h"
#include "src/utility/logger/logger.h"
#include "src/utility/notification_utility.h"

using namespace testing;
using namespace YR::Libruntime;

namespace YR {
namespace test {

using TLSConfig = ::common::TLSConfig;
using namespace YR::utility;
struct TestParam {
    bool dsEnable;
    std::string dscPubKey;
    std::string dscPriKey;
    std::string dssPubKey;
    bool fsEnable;
    std::string rootCACert;
    std::string token;
    bool fsServerMode;
    std::string serverNameOverride;
};
TestParam BuildOneCommonTestParam();
std::string BuildOneCommonTlsConfigStr(TestParam tp);
TestParam BuildOneCommonTestParam()
{
    return TestParam{
        .dsEnable = true,
        .dscPubKey = "ds-cli-pub-key",
        .dscPriKey = "ds-cli-pri-key",
        .dssPubKey = "ds-ser-pub-key",
        .fsEnable = true,
        .rootCACert = "root-ca-cert",
        .token = "token0",
        .fsServerMode = true,
        .serverNameOverride = "server-name-override",
    };
}

std::string BuildOneCommonTlsConfigStr(TestParam tp)
{
    TLSConfig tlsConfig;
    tlsConfig.set_dsauthenable(tp.dsEnable);
    tlsConfig.set_dsclientpublickey(tp.dscPubKey);
    tlsConfig.set_dsclientprivatekey(tp.dscPriKey);
    tlsConfig.set_dsserverpublickey(tp.dssPubKey);
    tlsConfig.set_serverauthenable(tp.fsEnable);
    tlsConfig.set_rootcertdata(tp.rootCACert);
    tlsConfig.set_token(tp.token);
    tlsConfig.set_enableservermode(tp.fsServerMode);
    tlsConfig.set_servernameoverride(tp.serverNameOverride);
    auto m = tlsConfig.SerializeAsString();
    return m;
}

class SecurityTest : public ::testing::Test {
public:
    SecurityTest() {}
    void SetUp() override
    {
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
    ~SecurityTest() {}
};

TEST_F(SecurityTest, ParseEmptyConfigTest)
{
    auto s = Security();
    auto err = s.Init();
    ASSERT_TRUE(err.OK());
}

TEST_F(SecurityTest, ParseNormalConfigTest)
{
    struct TestParam {
        bool dsEnable;
        std::string dscPubKey;
        std::string dscPriKey;
        std::string dssPubKey;
        bool fsEnable;
        std::string rootCACert;
        std::string token;
        bool fsServerMode;
        std::string serverNameOverride;
    } tps[] = {
        {
            .dsEnable = true,
            .fsEnable = false,
            .fsServerMode = false,
        },
        {
            .dsEnable = true,
            .dscPubKey = "ds-cli-pub-key",
            .dscPriKey = "ds-cli-pri-key",
            .dssPubKey = "ds-ser-pub-key",
            .fsEnable = true,
            .rootCACert = "root-ca-cert",
            .token = "token0",
            .fsServerMode = true,
            .serverNameOverride = "server-name-override",
        },
        {
            .dsEnable = true,
            .dscPubKey = "ds-cli-pub-keyx",
            .dscPriKey = "ds-cli-pri-keyx",
            .dssPubKey = "ds-ser-pub-keyx",
            .fsEnable = false,
            .rootCACert = "root-ca-certx",
            .token = "tokenx",
            .fsServerMode = true,
            .serverNameOverride = "server-name-overridex",
        },
    };

    int fildes[2];
    int status = pipe(fildes);
    ASSERT_NE(status, -1);
    status = dup2(fildes[0], STDIN_FILENO);
    ASSERT_NE(status, -1);

    for (auto &tp : tps) {
        TLSConfig tlsConfig;
        tlsConfig.set_dsauthenable(tp.dsEnable);
        tlsConfig.set_dsclientpublickey(tp.dscPubKey);
        tlsConfig.set_dsclientprivatekey(tp.dscPriKey);
        tlsConfig.set_dsserverpublickey(tp.dssPubKey);
        tlsConfig.set_serverauthenable(tp.fsEnable);
        tlsConfig.set_rootcertdata(tp.rootCACert);
        tlsConfig.set_token(tp.token);
        tlsConfig.set_enableservermode(tp.fsServerMode);
        tlsConfig.set_servernameoverride(tp.serverNameOverride);
        auto m = tlsConfig.SerializeAsString();
        YRLOG_ERROR("xxx message size: {}", m.size());
        if (m.size()) {
            ssize_t nbytes = write(fildes[1], m.data(), m.size());
            ASSERT_EQ(nbytes, m.size());
        }
        auto s = Security();
        auto err = s.Init();
        ASSERT_TRUE(err.OK());
        std::string dscPubKey, dssPubKey;
        SensitiveValue dscPriKey;
        auto [dsEnable, encryptEnable] = s.GetDataSystemConfig(dscPubKey, dscPriKey, dssPubKey);
        std::string rootCACert;
        std::string certChainData;
        std::string privateKey;
        bool fsEnable = s.GetFunctionSystemConfig(rootCACert, certChainData, privateKey);
        SensitiveValue token;
        s.GetToken(token);
        std::string sernameOverride;
        bool connMode = s.GetFunctionSystemConnectionMode(sernameOverride);
        ASSERT_EQ(dsEnable, tp.dsEnable);
        ASSERT_EQ(dscPubKey, tp.dscPubKey);
        ASSERT_EQ(dscPriKey, tp.dscPriKey);
        ASSERT_EQ(dssPubKey, tp.dssPubKey);
        ASSERT_EQ(fsEnable, tp.fsEnable);
        ASSERT_EQ(rootCACert, tp.rootCACert);
        ASSERT_EQ(token, tp.token);
        ASSERT_EQ(connMode, tp.fsServerMode);
        ASSERT_EQ(sernameOverride, tp.serverNameOverride);
    }
    close(fildes[0]);
    close(fildes[1]);
}

TEST_F(SecurityTest, UpdateHandlerSizeTest)
{
    int fildes[2];
    int status = pipe(fildes);
    ASSERT_NE(status, -1);
    status = dup2(fildes[0], STDIN_FILENO);
    ASSERT_NE(status, -1);
    TLSConfig tlsConfig;
    tlsConfig.set_accesskey("accesskey");
    tlsConfig.set_securitykey("securitykey"); 
    auto m = tlsConfig.SerializeAsString();
    YRLOG_ERROR("3 message size: {}", m.size());
    if (m.size()) {
        ssize_t nbytes = write(fildes[1], m.data(), m.size());
        ASSERT_EQ(nbytes, m.size());
    }
    auto s = Security();
    auto err = s.Init();
    s.WhenTokenUpdated([](const SensitiveValue &token) {});
    ASSERT_EQ(s.GetUpdateHandersSize(), 0);
    close(fildes[0]);
    close(fildes[1]);
}

TEST_F(SecurityTest, UpdateTokenTest)
{
    struct TestParam {
        bool dsEnable;
        std::string dscPubKey;
        std::string dscPriKey;
        std::string dssPubKey;
        bool fsEnable;
        std::string rootCACert;
        std::string token;
        bool fsServerMode;
        std::string serverNameOverride;
    } tps[] = {
        {
            .dsEnable = true,
            .dscPubKey = "ds-cli-pub-key2",
            .dscPriKey = "ds-cli-pri-key2",
            .dssPubKey = "ds-ser-pub-key2",
            .fsEnable = true,
            .rootCACert = "root-ca-cert2",
            .token = "token2",
            .fsServerMode = true,
            .serverNameOverride = "server-name-override2",
        },
        {
            .dsEnable = true,
            .dscPubKey = "ds-cli-pub-key3",
            .dscPriKey = "ds-cli-pri-key3",
            .dssPubKey = "ds-ser-pub-key3",
            .fsEnable = true,
            .rootCACert = "root-ca-cert3",
            .token = "token3",
            .fsServerMode = true,
            .serverNameOverride = "server-name-override3",
        },
    };

    int fildes[2];
    int status = pipe(fildes);
    ASSERT_NE(status, -1);
    status = dup2(fildes[0], STDIN_FILENO);
    ASSERT_NE(status, -1);

    auto s = Security();

    auto &tp = tps[0];

    TLSConfig tlsConfig;
    tlsConfig.set_dsauthenable(tp.dsEnable);
    tlsConfig.set_dsclientpublickey(tp.dscPubKey);
    tlsConfig.set_dsclientprivatekey(tp.dscPriKey);
    tlsConfig.set_dsserverpublickey(tp.dssPubKey);
    tlsConfig.set_serverauthenable(tp.fsEnable);
    tlsConfig.set_rootcertdata(tp.rootCACert);
    tlsConfig.set_token(tp.token);
    tlsConfig.set_enableservermode(tp.fsServerMode);
    tlsConfig.set_servernameoverride(tp.serverNameOverride);
    auto m = tlsConfig.SerializeAsString();
    YRLOG_ERROR("1 message size: {}", m.size());
    if (m.size()) {
        ssize_t nbytes = write(fildes[1], m.data(), m.size());
        ASSERT_EQ(nbytes, m.size());
    }
    auto err = s.Init();
    ASSERT_TRUE(err.OK());
    std::string dscPubKey, dssPubKey;
    SensitiveValue dscPriKey;
    auto [dsEnable, encryptEnable] = s.GetDataSystemConfig(dscPubKey, dscPriKey, dssPubKey);
    std::string rootCACert;
    std::string certChainData;
    std::string privateKey;
    bool fsEnable = s.GetFunctionSystemConfig(rootCACert, certChainData, privateKey);
    SensitiveValue token;
    s.GetToken(token);
    std::string sernameOverride;
    bool connMode = s.GetFunctionSystemConnectionMode(sernameOverride);
    ASSERT_EQ(dsEnable, tp.dsEnable);
    ASSERT_EQ(dscPubKey, tp.dscPubKey);
    ASSERT_EQ(dscPriKey, tp.dscPriKey);
    ASSERT_EQ(dssPubKey, tp.dssPubKey);
    ASSERT_EQ(fsEnable, tp.fsEnable);
    ASSERT_EQ(rootCACert, tp.rootCACert);
    ASSERT_EQ(token, tp.token);
    ASSERT_EQ(connMode, tp.fsServerMode);
    ASSERT_EQ(sernameOverride, tp.serverNameOverride);

    SensitiveValue tokenTest;
    YR::utility::NotificationUtility n;
    s.WhenTokenUpdated([&tokenTest, &n](const SensitiveValue &token) {
        tokenTest = token;
        n.Notify();
    });

    tp = tps[1];

    tlsConfig.set_dsauthenable(tp.dsEnable);
    tlsConfig.set_dsclientpublickey(tp.dscPubKey);
    tlsConfig.set_dsclientprivatekey(tp.dscPriKey);
    tlsConfig.set_dsserverpublickey(tp.dssPubKey);
    tlsConfig.set_serverauthenable(tp.fsEnable);
    tlsConfig.set_rootcertdata(tp.rootCACert);
    tlsConfig.set_token(tp.token);
    tlsConfig.set_enableservermode(tp.fsServerMode);
    tlsConfig.set_servernameoverride(tp.serverNameOverride);
    m = tlsConfig.SerializeAsString();
    YRLOG_ERROR("2 message size: {}", m.size());
    if (m.size()) {
        ssize_t nbytes = write(fildes[1], m.data(), m.size());
        ASSERT_EQ(nbytes, m.size());
    }
    n.WaitForNotification();
    s.GetToken(token);
    ASSERT_EQ(token, tp.token);
    close(fildes[0]);
    close(fildes[1]);
}

// DTS2024071820905
// ENABLE_DS_AUTH stdin Delay Write
// there may be a problem if the file descriptor (fd) hasn't been created after start runtime.
TEST_F(SecurityTest, DelayStdinShouldSuccessTest)
{
    setenv("ENABLE_DS_AUTH", "true", 1);
    Libruntime::Config::Instance() = Libruntime::Config();
    auto tp = BuildOneCommonTestParam();
    auto m = BuildOneCommonTlsConfigStr(tp);
    auto promise = std::make_shared<std::promise<bool>>();
    auto future = std::make_shared<std::future<bool>>(promise->get_future());
    int fds[2];
    int status = pipe(fds);
    ASSERT_NE(status, -1);
    status = dup2(fds[0], STDIN_FILENO);
    ASSERT_NE(status, -1);
    int writeFd = fds[1];
    std::thread t([future, m, writeFd]() {
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
        ssize_t nBytes = write(writeFd, m.data(), m.size());
        std::cout << "nBytes: " << nBytes << std::endl;
        future->get();
    });
    auto s = Security(STDIN_FILENO, 1000);
    auto err = s.Init();
    promise->set_value(true);
    t.join();
    ASSERT_TRUE(err.OK());
    SensitiveValue token;
    s.GetToken(token);
    ASSERT_EQ(token, tp.token);
    unsetenv("ENABLE_DS_AUTH");
    Libruntime::Config::Instance() = Libruntime::Config();
    close(fds[0]);
    close(fds[1]);
}

// ENABLE_DS_AUTH no stdin Write
TEST_F(SecurityTest, NoStdinShouldTimeoutFailedTest)
{
    setenv("ENABLE_DS_AUTH", "true", 1);
    Libruntime::Config::Instance() = Libruntime::Config();
    auto tp = BuildOneCommonTestParam();
    auto m = BuildOneCommonTlsConfigStr(tp);
    auto s = Security(STDIN_FILENO, 1000);
    auto err = s.Init();
    ASSERT_FALSE(err.OK());
    unsetenv("ENABLE_DS_AUTH");
    Libruntime::Config::Instance() = Libruntime::Config();
}

TEST_F(SecurityTest, GetAKSK)
{
    auto s = std::make_shared<Security>();
    std::string ak;
    SensitiveValue sk;
    s->GetAKSK(ak, sk);
    EXPECT_TRUE(ak.empty());
    EXPECT_TRUE(sk.Empty());
}

TEST_F(SecurityTest, GetAKSKDK)
{
    auto s = std::make_shared<Security>();
    std::string ak;
    SensitiveValue sk;
    SensitiveValue dk;
    s->GetAKSKDK(ak, sk, dk);
    EXPECT_TRUE(ak.empty());
    EXPECT_TRUE(sk.Empty());
    EXPECT_TRUE(dk.Empty());
}
}  // namespace test
}  // namespace YR