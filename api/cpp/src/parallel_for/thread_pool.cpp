/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2020. All rights reserved.
 * Description: memory pool implementation
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

#include "thread_pool.h"
#include <pthread.h>
#include <string>
#include <vector>
namespace YR {
namespace Parallel {
thread_local int g_threadid = 0;
const std::string threadNamePrefix = "yr.parallel";
void ThreadPool::ThreadTask(int workerId)
{
    // master is thread 0
    g_threadid = workerId + 1;
    while (1) {
        intptr_t task;
        int status;
        while (1) {
            if (stopped.load(std::memory_order_relaxed)) {
                return;
            }
            status = tasks.DequeueParallel(&task);
            if (status == H_SUCCESS) {
                break;
            }
            WorkerIdle(workerId);
        }
        Task *tmp = reinterpret_cast<Task *>(task);
        tmp->func();
        tmp->func = nullptr;
        memPool.FreeObjToMemPool(reinterpret_cast<void *>(tmp));
    }
}

void ThreadPool::Init(uint32_t threads)
{
    if (!stopped.load(std::memory_order_relaxed)) {
        return;
    }
    // setup stopped flag before start threads, otherwise threads will exit
    stopped = false;
    sleepWorkerBitmap = 0;
    workerNum = threads;
    // a maximum of 64 threads can be enabled currently。
    if (workerNum > MAX_WORKER_NUM) {
        workerNum = MAX_WORKER_NUM;
    }
    (void)memPool.InitMemPool(sizeof(Task), MAX_MEM_POOL_NUM);
    tasks = TaskQueue(MAX_TASK_QUEUE_NUM);
    workers.resize(workerNum);
    ThreadInit();
}

void ThreadPool::ThreadInit()
{
    for (uint32_t i = 0; i < workerNum; i++) {
        workers[i] = std::thread(&ThreadPool::ThreadTask, this, i);
        std::string threadName = threadNamePrefix + "." + std::to_string(i);
        pthread_setname_np(workers[i].native_handle(), threadName.data());
    }
}

ThreadPool::~ThreadPool()
{
    if (!stopped.load(std::memory_order_relaxed)) {
        Stop();
    }
}

ThreadPool &ThreadPool::GetInstance()
{
    static ThreadPool threadPool;
    return threadPool;
}

int ThreadPool::SubmitTaskToPool(std::function<void()> &&task)
{
    int num;
    void *objTable = memPool.AllocObjFromMemPool();
    if (objTable == nullptr) {
        return H_FAIL;
    }
    Task *tmpTask = (Task *)objTable;
    tmpTask->func = std::move(task);

    num = tasks.EnqueueParallel(reinterpret_cast<intptr_t>(objTable));
    if (num != 0) {
        memPool.FreeObjToMemPool(objTable);
        return H_FAIL;
    }
    WorkerWakeup(1);

    return H_SUCCESS;
}

void ThreadPool::WorkerJoin(int workerId)
{
    sleepWorkers[workerId].event.EventWakeUp();
}

void ThreadPool::Stop()
{
    {
        if (stopped) {
            return;
        }
        stopped = true;
        for (size_t i = 0; i < workers.size(); i++) {
            WorkerJoin(static_cast<int>(i));
        }
    }
    for (auto &worker : workers) {
        worker.join();
    }
    workers.clear();
    memPool.DestroyMemPool();
}

bool ThreadPool::IsStop()
{
    return stopped;
}

}  // namespace Parallel
}  // namespace YR