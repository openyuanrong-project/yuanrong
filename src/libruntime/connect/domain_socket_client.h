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

#pragma once

#include <atomic>
#include <condition_variable>
#include <mutex>
#include <queue>
#include <string>
#include <thread>

#include <sys/socket.h>
#include <sys/un.h>

#include "securec.h"

#include "src/libruntime/err_type.h"

namespace YR {
namespace Libruntime {
#define SOCKET_INIT_ONCE()     \
    do {                       \
        InitOnce();            \
        if (!initErr_.OK()) {  \
            return initErr_;   \
        }                      \
    } while (0)

class DomainSocketClient {
public:
    explicit DomainSocketClient(std::string socketPath);
    ~DomainSocketClient();
    ErrorInfo Send(std::string msg);

private:
    void InitOnce();
    ErrorInfo DoInitOnce();
    void Stop();
    void CleanupSocket();
    bool IsEmpty();
    void HandleWrite();
    void PopAndSendBatch();

    int sockfd_ = -1;
    std::string socketPath_;
    std::thread writeThread_;
    mutable std::mutex mu_;
    std::condition_variable cv_;
    std::atomic<bool> running_;
    std::queue<std::string> msgQueue_;
    std::once_flag initFlag_;
    ErrorInfo initErr_;
};
}  // namespace Libruntime
}  // namespace YR
