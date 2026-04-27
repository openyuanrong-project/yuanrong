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

#pragma once
#include <sys/inotify.h>
#include <unistd.h>
#include <atomic>
#include <cerrno>
#include <climits>
#include <cstring>
#include <functional>
#include <iostream>
#include <thread>
#include "src/utility/logger/logger.h"
namespace YR {
namespace utility {
class FileWatcher {
public:
    using Callback = std::function<void(const std::string &)>;

    FileWatcher(const std::string &filename, Callback callback);

    ~FileWatcher();

    void Start();

    void Stop();

private:
    static constexpr size_t EVENT_BUF_LEN = 1024 * (sizeof(struct inotify_event) + NAME_MAX + 1);
    static constexpr uint32_t WATCH_MASK = IN_CLOSE_WRITE;  // monitor write completion

    void Watch();

    std::string filename_;
    Callback callback_;
    std::atomic<bool> running_{false};
    std::thread watcherThread_;
    int fd_{-1};
    int wd_{-1};
};
}  // namespace utility
}  // namespace YR
