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

#include "absl/synchronization/notification.h"

namespace YR {
namespace utility {
class NotificationUtility {
public:
    NotificationUtility() = default;
    ~NotificationUtility() = default;

    void Notify(void)
    {
        if (!notification_.HasBeenNotified()) {
            notification_.Notify();
        }
    }

    void Notify(const YR::Libruntime::ErrorInfo &errInfo)
    {
        errInfo_ = errInfo;
        if (!notification_.HasBeenNotified()) {
            notification_.Notify();
        }
    }

    YR::Libruntime::ErrorInfo WaitForNotification(void)
    {
        notification_.WaitForNotification();
        return errInfo_;
    }

    YR::Libruntime::ErrorInfo WaitForNotificationWithTimeout(absl::Duration timeout,
                                                             const YR::Libruntime::ErrorInfo &errorInfo)
    {
        return notification_.WaitForNotificationWithTimeout(timeout) ? errInfo_ : errorInfo;
    }

private:
    absl::Notification notification_;
    YR::Libruntime::ErrorInfo errInfo_;
};
}  // namespace utility
}  // namespace YR