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

#include <gmock/gmock.h>
#include <gtest/gtest.h>
#include <boost/asio/ssl.hpp>
#include <boost/beast/http.hpp>

#define private public
#include "src/libruntime/utils/ak_sk_manager.h"

using namespace YR::Libruntime;
using namespace testing;
using namespace YR::utility;

namespace YR {
namespace test {
class AkSkInfoTest : public ::testing::Test {
    void SetUp() override {}
    void TearDown() override {}
};

TEST_F(AkSkInfoTest, DefaultConstructor)
{
    AkSkInfo info;
    EXPECT_TRUE(info.accessKey.empty());
    EXPECT_TRUE(info.secretKey.Empty());
    EXPECT_EQ(info.expiredTimeStamp, 0);
}

TEST_F(AkSkInfoTest, ToString)
{
    AkSkInfo info;
    info.accessKey = "testAccessKey";
    info.secretKey = std::string("testSecretKey");
    info.expiredTimeStamp = 1234567890;
    std::string result = info.toString();
    EXPECT_EQ(result, "ak: testAccessKey, expiredTimeStamp: 1234567890");
}

class MockClientManager : public ClientManager {
public:
    explicit MockClientManager(const std::shared_ptr<LibruntimeConfig> &librtCfg) : ClientManager(librtCfg) {}
    MOCK_METHOD1(Init, ErrorInfo(const ConnectionParam &param));
    MOCK_METHOD6(SubmitInvokeRequest,
    void(const http::verb &method, const std::string &target,
    const std::unordered_map<std::string, std::string> &headers, const std::string &body,
    const std::shared_ptr<std::string> requestId, const HttpCallbackFunction &receiver));
};

class MockIamClient : public IamClient {
public:
    explicit MockIamClient(std::shared_ptr<ClientsManager> clientsMgr, std::shared_ptr<Security> security,
        std::shared_ptr<LibruntimeConfig> librtConfig) : IamClient(clientsMgr, security, librtConfig) {}
    MOCK_METHOD(ErrorInfo, Init, (), (override));
    MOCK_METHOD(ErrorInfo, DecryptKey, (SensitiveValue &key), (override));
};

class MockInvocationSync : public GwClient {
public:
    explicit MockInvocationSync(const std::string &funcId, FSIntfHandlers handlers,
                                std::shared_ptr<Security> security) : GwClient(funcId, handlers, security) {}
    MOCK_METHOD(void, InvocationSync, ((const std::string &method), (const std::string &path),
                                       (const std::unordered_map<std::string, std::string> &headers),
                                       (const std::string &body), (const InvocationCallback &callback)),
                (override));
};

class IamClientTest : public ::testing::Test {
    void SetUp() override
    {
        librtConfig = std::make_shared<LibruntimeConfig>();
        librtConfig->ak_ = "ak";
        librtConfig->sk_ = "sk";
        librtConfig->dk_ = "dk";
        librtConfig->iamAddress = "127.0.0.1:11111";
        clientMgr = std::make_shared<MockClientManager>(librtConfig);
        clientsMgr = std::make_shared<ClientsManager>();
        clientsMgr->httpClients[librtConfig->iamAddress] = clientMgr;
        security = std::make_shared<Security>();
        security->InitWithDriver(this->librtConfig);
        FSIntfHandlers handlers;
        std::string functionId;
        gwClient = std::make_shared<GwClient>(functionId, handlers, security);
    }

    void TearDown() override
    {
        if (gwClient) {
            gwClient.reset();
        }
        clientsMgr = nullptr;
        security.reset();
    }

    std::shared_ptr<MockClientManager> clientMgr;
    std::shared_ptr<ClientsManager> clientsMgr;
    std::shared_ptr<GwClient> gwClient;
    std::shared_ptr<Security> security;
    std::shared_ptr<LibruntimeConfig> librtConfig;
};

TEST_F(IamClientTest, init)
{
    auto iamClient = std::make_shared<IamClient>(this->clientsMgr, this->security, this->librtConfig);
    ASSERT_TRUE(iamClient->Init().OK());
}

TEST_F(IamClientTest, DecryptKey)
{
    auto security = std::make_shared<Security>();
    auto librtConfig = std::make_shared<LibruntimeConfig>();
    SensitiveValue encryptedKey = std::string("00123");

    // 1、系统级DK为空
    auto iamClient1 = std::make_shared<IamClient>(this->clientsMgr, security, librtConfig);
    auto err = iamClient1->DecryptKey(encryptedKey);
    ASSERT_FALSE(err.OK());

    // 2、解密失败
    auto iamClient2 = std::make_shared<IamClient>(this->clientsMgr, this->security, this->librtConfig);
    ASSERT_TRUE(iamClient2->Init().OK());
    err = iamClient2->DecryptKey(encryptedKey);
    ASSERT_FALSE(err.OK());
}

TEST_F(IamClientTest, GetTenantAkSkInfo)
{
    auto mockIamClient = std::make_shared<MockIamClient>(this->clientsMgr, this->security, this->librtConfig);
    FSIntfHandlers handlers;
    std::string functionId;
    auto mockInvocationSync = std::make_shared<MockInvocationSync>(functionId, handlers, this->security);

    // 1、租户为空，结果为空
    {
        auto result = mockIamClient->GetTenantAkSkInfo("");
        EXPECT_EQ(result, nullptr);
    }

    // 2、gwClient_为空，init后也为空，结果为空
    {
        EXPECT_CALL(*mockIamClient, Init()).WillOnce(testing::Return(ErrorInfo()));
        auto result = mockIamClient->GetTenantAkSkInfo("");
        EXPECT_EQ(result, nullptr);
    }
}

class MockIamClient2 : public IamClient {
public:
    explicit MockIamClient2(std::shared_ptr<ClientsManager> clientsMgr, std::shared_ptr<Security> security,
                            std::shared_ptr<LibruntimeConfig> librtConfig)
        : IamClient(clientsMgr, security, librtConfig)
    {}
    MOCK_METHOD(std::shared_ptr<AkSkInfo>, GetTenantAkSkInfo, (const std::string& tenantId), (override));
};

class TenantAKSKManagerTest : public ::testing::Test {
public:
    void SetUp() override {}
    void TearDown() override {}
};

TEST_F(TenantAKSKManagerTest, GetInstance)
{
    // 获取单例实例
    TenantAKSKManager& instance1 = TenantAKSKManager::GetInstance();
    TenantAKSKManager& instance2 = TenantAKSKManager::GetInstance();

    // 验证两个实例是否为同一个对象
    EXPECT_EQ(&instance1, &instance2);
}

TEST_F(TenantAKSKManagerTest, Initialize)
{
    auto clientsManager = std::make_shared<ClientsManager>();
    auto security = std::make_shared<Security>();
    auto librtConfig = std::make_shared<LibruntimeConfig>();

    // 首次初始化
    TenantAKSKManager::GetInstance().Initialize(clientsManager, security, librtConfig);

    // 验证iamClient_是否被正确初始化
    EXPECT_NE(TenantAKSKManager::GetInstance().iamClient_, nullptr);

    // 验证initFlag_是否被设置为true
    EXPECT_TRUE(TenantAKSKManager::GetInstance().initFlag_);
}

TEST_F(TenantAKSKManagerTest, GetValue)
{
    auto clientsManager = std::make_shared<ClientsManager>();
    auto security = std::make_shared<Security>();
    auto librtConfig = std::make_shared<LibruntimeConfig>();
    auto mockIamClient = std::make_shared<MockIamClient2>(clientsManager, security, librtConfig);
    TenantAKSKManager::GetInstance().iamClient_ = mockIamClient;

    // 测试1：tenantId为空，返回空指针
    {
        auto result = TenantAKSKManager::GetInstance().GetValue("");
        EXPECT_EQ(result, nullptr);
    }

    // 测试2：缓存中不存在AkSkInfo，从IamClient获取新的AkSkInfo
    {
        // 设置期望调用
        TenantAKSKManager::GetInstance().cacheMap.clear();
        auto newAkSkInfo = std::make_shared<AkSkInfo>();
        newAkSkInfo->accessKey = "ak1";
        newAkSkInfo->secretKey = SensitiveValue("sk");
        newAkSkInfo->expiredTimeStamp = 0;
        std::string tenantId = "tenant1";
        EXPECT_CALL(*mockIamClient, GetTenantAkSkInfo(_)).WillOnce(testing::Return(newAkSkInfo));

        auto result = TenantAKSKManager::GetInstance().GetValue(tenantId);

        // 验证结果
        EXPECT_NE(result, nullptr);
        EXPECT_EQ(result->accessKey, "ak1");
    }

    // 测试3：GetTenantAkSkInfo返回空指针或无效的AkSkInfo， 返回空指针
    {
        // 设置期望调用
        TenantAKSKManager::GetInstance().cacheMap.clear();
        std::string tenantId = "tenant2";
        EXPECT_CALL(*mockIamClient, GetTenantAkSkInfo(_)).WillOnce(testing::Return(nullptr));

        auto result = TenantAKSKManager::GetInstance().GetValue(tenantId);

        // 验证结果
        EXPECT_EQ(result, nullptr);
    }

    // 测试4：缓存中存在未过期的AkSkInfo，直接返回
    {
        // 设置期望调用，将AkSkInfo添加到缓存
        std::string tenantId = "tenant3";
        auto akskInfo = std::make_shared<AkSkInfo>();
        akskInfo->accessKey = "ak3";
        akskInfo->secretKey = std::string("sk3");
        akskInfo->expiredTimeStamp = 0;
        TenantAKSKManager::GetInstance().cacheMap[tenantId] = akskInfo;
        EXPECT_CALL(*mockIamClient, GetTenantAkSkInfo(_)).Times(0);

        // 调用被测方法
        auto result = TenantAKSKManager::GetInstance().GetValue(tenantId);

        // 验证结果
        EXPECT_NE(result, nullptr);
        EXPECT_EQ(result->accessKey, "ak3");
    }

    // 测试5：缓存中存在已过期的AkSkInfo，从IamClient获取新的AkSkInfo
    {
        std::string tenantId = "tenant4";
        // 将过期的AkSkInfo添加到缓存
        auto oldAkSkInfo = std::make_shared<AkSkInfo>();
        oldAkSkInfo->accessKey = "oldAccessKey";
        oldAkSkInfo->secretKey = std::string("oldSecretKey");
        oldAkSkInfo->expiredTimeStamp = static_cast<uint64_t>(std::time(nullptr)) - 3600;  // 1小时前过期
        TenantAKSKManager::GetInstance().cacheMap[tenantId] = oldAkSkInfo;

        // 设置期望调用
        auto newAkSkInfo = std::make_shared<AkSkInfo>();
        newAkSkInfo->accessKey = "ak4";
        newAkSkInfo->secretKey = std::string("sk4");
        newAkSkInfo->expiredTimeStamp = 0;
        EXPECT_CALL(*mockIamClient, GetTenantAkSkInfo(_)).WillOnce(testing::Return(newAkSkInfo));

        // 调用被测方法
        auto result = TenantAKSKManager::GetInstance().GetValue(tenantId);

        // 验证结果
        EXPECT_NE(result, nullptr);
        EXPECT_EQ(result->accessKey, "ak4");
    }
}
}  // namespace test
}  // namespace YR