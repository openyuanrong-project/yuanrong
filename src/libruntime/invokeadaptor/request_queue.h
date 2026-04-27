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

#pragma once

#include <memory>
#include <queue>

#include "absl/synchronization/mutex.h"

#include "src/libruntime/invoke_spec.h"

namespace YR {
namespace Libruntime {

class BaseQueue {
public:
    virtual ~BaseQueue() = default;
    virtual void Pop() = 0;
    virtual std::shared_ptr<InvokeSpec> Top() = 0;
    virtual void Push(std::shared_ptr<InvokeSpec> spec) = 0;
    virtual size_t Size() const = 0;
    virtual bool Empty() const = 0;
public:
    mutable std::mutex atomicMtx;
};

struct Cmp {
    bool operator()(std::shared_ptr<InvokeSpec> left, std::shared_ptr<InvokeSpec> right)
    {
        return left->opts.priority <= right->opts.priority;
    }
};

class PriorityQueue : public BaseQueue {
public:
    void Pop() override;
    std::shared_ptr<InvokeSpec> Top() override;
    void Push(std::shared_ptr<InvokeSpec> spec) override;
    size_t Size() const override;
    bool Empty() const override;

private:
    std::priority_queue<std::shared_ptr<InvokeSpec>, std::vector<std::shared_ptr<InvokeSpec>>, Cmp> queue;
    mutable absl::Mutex mutex;
};

class Queue : public BaseQueue {
public:
    void Pop() override;
    std::shared_ptr<InvokeSpec> Top() override;
    void Push(std::shared_ptr<InvokeSpec> spec) override;
    size_t Size() const override;
    bool Empty() const override;

private:
    std::queue<std::shared_ptr<InvokeSpec>> queue;
    mutable absl::Mutex mutex;
};

}  // namespace Libruntime
}  // namespace YR
