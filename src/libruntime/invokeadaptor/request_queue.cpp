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

#include "request_queue.h"

namespace YR {
namespace Libruntime {

void PriorityQueue::Pop()
{
    absl::WriterMutexLock lock(&this->mutex);
    this->queue.pop();
}

std::shared_ptr<InvokeSpec> PriorityQueue::Top()
{
    // 读写锁
    absl::ReaderMutexLock lock(&this->mutex);
    return this->queue.top();
}

void PriorityQueue::Push(std::shared_ptr<InvokeSpec> spec)
{
    absl::WriterMutexLock lock(&this->mutex);
    this->queue.push(spec);
}

size_t PriorityQueue::Size() const
{
    absl::ReaderMutexLock lock(&this->mutex);
    return this->queue.size();
}

bool PriorityQueue::Empty() const
{
    absl::ReaderMutexLock lock(&this->mutex);
    return this->queue.empty();
}

void Queue::Pop()
{
    absl::WriterMutexLock lock(&this->mutex);
    this->queue.pop();
}

std::shared_ptr<InvokeSpec> Queue::Top()
{
    // 读写锁
    absl::ReaderMutexLock lock(&this->mutex);
    return this->queue.front();
}

void Queue::Push(std::shared_ptr<InvokeSpec> spec)
{
    absl::WriterMutexLock lock(&this->mutex);
    this->queue.push(spec);
}

size_t Queue::Size() const
{
    absl::ReaderMutexLock lock(&this->mutex);
    return this->queue.size();
}

bool Queue::Empty() const
{
    absl::ReaderMutexLock lock(&this->mutex);
    return this->queue.empty();
}

}  // namespace Libruntime
}  // namespace YR
