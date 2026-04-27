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
#include "file_watcher.h"
namespace YR {
namespace utility {

FileWatcher::FileWatcher(const std::string &filename, Callback callback)
    : filename_(filename), callback_(std::move(callback)), running_(false), fd_(-1), wd_(-1)
{
}

FileWatcher::~FileWatcher()
{
    Stop();
}

void FileWatcher::Start()
{
    if (running_) {
        return;
    }

    running_ = true;
    watcherThread_ = std::thread([this]() {
        while (true) {
            if (!running_) {
                break;
            }
            Watch();
            std::this_thread::sleep_for(std::chrono::seconds(1));
        }
    });
}

void FileWatcher::Stop()
{
    if (!running_) {
        return;
    }
    running_ = false;
    if (wd_ >= 0) {
        inotify_rm_watch(fd_, wd_);
        wd_ = -1;
    }
    if (fd_ >= 0) {
        close(fd_);
        fd_ = -1;
    }
    if (watcherThread_.joinable()) {
        watcherThread_.join();
    }
}

void FileWatcher::Watch()
{
    fd_ = inotify_init1(IN_NONBLOCK);
    if (fd_ < 0) {
        YRLOG_DEBUG("inotify_init1 failed {}", strerror(errno));
        return;
    }

    wd_ = inotify_add_watch(fd_, filename_.c_str(), WATCH_MASK);
    if (wd_ < 0) {
        YRLOG_DEBUG("inotify_add_watch failed {}", strerror(errno));
        close(fd_);
        fd_ = -1;
        return;
    }
    callback_(filename_);
    char buffer[EVENT_BUF_LEN];
    while (running_) {
        int length = read(fd_, buffer, EVENT_BUF_LEN);
        if (length < 0) {
            if (errno == EINTR) {  // Signal interrupted
                continue;
            }
            if (errno == EAGAIN) {  // No data available
                std::this_thread::sleep_for(std::chrono::seconds(1));
                continue;
            }
            YRLOG_WARN("read error: {}", strerror(errno));
            break;
        }

        int i = 0;
        while (i < length) {
            struct inotify_event *event = static_cast<struct inotify_event *>(static_cast<void *>(&buffer[i]));
            if (event->mask & IN_CLOSE_WRITE) {
                callback_(filename_);
            }
            i += sizeof(struct inotify_event) + event->len;
        }
    }
}
}  // namespace utility
}  // namespace YR
