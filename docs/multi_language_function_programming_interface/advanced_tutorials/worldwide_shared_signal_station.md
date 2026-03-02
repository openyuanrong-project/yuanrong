# 使用有状态函数作为全局信号站

在分布式系统中，由于不同的任务可能运行在不同的物理节点上，Python 原生的 asyncio.Event 无法跨进程工作。
通过 openYuanrong 有状态函数，我们可以创建一个全局共享的信号站，让成百上千个分布式任务同时“监听”某一个事件的发生。

## 核心原理

- 有状态函数作为中心节点：维护一个真正的 asyncio.Event 对象。
- 方法订阅：其他任务调用有状态函数的 wait() 方法，该方法会 await 内部的事件。
- 广播触发：当某个任务调用有状态函数的 set() 时，所有正在 await 的任务都会被同时唤醒。

## 场景

- 同步启动（Barrier Synchronization）：比如在分布式训练中，确保所有节点都加载完模型后，再同时开始训练。
- 配置更新：当配置发生变化时，通过 Event 通知所有运行中的有状态函数重新加载配置。
- 依赖触发：任务 B 必须等任务 A 的某个中间步骤完成后才能继续，但任务 A 并没有结束（无法用 `yr.get()` 结果作为触发点）。

## 使用示例

```python
import yr
import asyncio
import time


# 1. 定义分布式事件中心
@yr.instance
class SharedEvent:
    def __init__(self):
        # 核心：使用 asyncio.Event 驱动异步非阻塞等待
        self.event = asyncio.Event()

    async def wait(self):
        """分布式任务在此挂起等待"""
        print("信号中心：收到一个等待请求...")
        await self.event.wait()
        return "SIGNAL_RECEIVED"

    async def set(self):
        """触发事件，所有等待的任务将同步启动"""
        self.event.set()
        print("信号中心：广播信号已发出！")

    async def clear(self):
        """重置信号，以便进行下一轮同步"""
        self.event.clear()


# 2. 定义分布式 Worker
@yr.invoke
def heavy_worker(worker_id, event_actor):
    print(f"Worker {worker_id}: 正在准备基础环境（如加载模型）...")
    time.sleep(1)  # 模拟准备工作

    print(f"Worker {worker_id}: 准备就绪，阻塞等待全局启动信号...")
    # 这里会通过网络向 Actor 发起异步等待
    yr.get(event_actor.wait.invoke())

    print(f"Worker {worker_id}: 收到信号，开始执行并行任务！")
    return f"Worker {worker_id} Success"


# --- 3. 执行流程 ---
yr.init()

# 创建全局唯一的信号中心
event_center = SharedEvent.invoke()

# 启动 5 个分布在集群各处的 Worker
worker_ids = range(5)
futures = [heavy_worker.invoke(i, event_center) for i in worker_ids]

print("\n--- 主程序：等待 3 秒后统一放行 ---")
time.sleep(3)

# 触发信号，唤醒所有阻塞在 wait() 处的 Worker
event_center.set.invoke()

# 查看执行结果
results = yr.get(futures)
print("\n所有任务执行完毕:", results)

yr.finalize()
```
