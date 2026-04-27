# 基于 openYuanrong 实现蒙特卡罗方法

蒙特卡罗(Monte Carlo)方法，或称计算机随机模拟方法，是一种基于“随机数”的计算方法，它是高性能计算 HPC 常见的算法之一。一个简单的用法是计算圆周率 π，主要思想是，在边长为 R 的正方形内画一个内切圆，然后在正方形内随机打点，则点落在圆内的概率为圆面积除以正方形面积，即：`π/4` 。据此通过计算概率，就能估算出 π 的值，并且打的点越多，估算的精度越高。

本示例将向您展示如何使用 openYuanrong 开发一个简单的 HPC 服务，实现动态任务并行，它包含以下内容：

- 如何使用无状态函数并行多个任务。
- 如何使用有状态函数获取并保存并行任务的执行状态。

## 方案介绍

我们使用 Python 语言，通过定义一个无状态函数来实现打点任务，一个有状态函数来统计任务的运行状态。代码可以运行在单台主机上，也可直接扩展到集群上，以配置更大的任务量提升计算精度。

## 准备工作

参考[在主机上部署](../../deploy/deploy_processes/index.md)完成 openYuanrong 部署。

## 实现流程

### 定义有状态函数统计打点任务的状态

我们通过 `@yr.instance` 装饰类 TaskSummary 来定义一个有状态函数，用以统计打点任务的运行状态。打点任务通过调用它的 report_status 方法上报任务状态，主程序通过调用它的 get_task_status 方法获取所有任务的状态。该类的实例会在远程进程上运行，其成员变量 `task_started_num` 等在运行中一直保持状态。持有类对象的句柄即可调用它的方法，改变成员变量的状态。

```python
class TaskStatus(Enum):
    PENDING = -1
    STARTED = 0
    COMPLETED = 1

@yr.instance
class TaskSummary:
    def __init__(self, total_task_num: int):
        self.total_task_num = total_task_num
        self.task_started_num = 0
        self.task_completed_num = 0

    def report_status(self, task_status: TaskStatus) -> None:
        if task_status == TaskStatus.STARTED:
            self.task_started_num += 1
        elif task_status == TaskStatus.COMPLETED:
            self.task_completed_num += 1

    def get_task_status(self) -> int:
        return self.task_started_num, self.task_completed_num
```

### 定义无状态函数处理打点任务

我们通过 `@yr.invoke` 装饰普通函数 monte_carlo_pi 来定义一个无状态函数，用以处理打点任务，统计落入圆中的点的个数。函数实例会在远程进程上异步运行，通过有状态函数 `TaskSummary` 的句柄调用 report_status 方法上报任务状态。

```python
@yr.invoke
def monte_carlo_pi(total_points: int, task_summary: yr.decorator.instance_proxy.InstanceProxy) -> int:
    # 上报任务已开始
    task_summary.report_status.invoke(TaskStatus.STARTED)

    circle_points = 0
    for i in range(total_points):
        rand_x = random.uniform(-1, 1)
        rand_y = random.uniform(-1, 1)

        origin_dist = rand_x**2 + rand_y**2
        if origin_dist <= 1:
            circle_points += 1

    # 上报任务已结束
    task_summary.report_status.invoke(TaskStatus.COMPLETED)
    return circle_points
```

### 定义主流程

主流程中通过 `yr.init()` 初始化 openYuanrong 运行上下文，您可以根据自己的集群规模调整任务的数量。

我们通过 `yr.InvokeOptions` 指定了无状态函数任务运行所需资源量，以便更好的观察它们的并行状态。使用 `monte_carlo_pi.options(opt).invoke(POINTS_PER_TASK, task_summary)`  启动任务，在 while 循环中查询任务的运行状态，直到所有任务全部完成。

最后，汇总各任务返回的打点数据计算 π，并调用 yr.finalize() 来清理上下文。

```python
import yr
import time
import random
from enum import Enum

if __name__ == '__main__':

    yr.init()

    # 根据集群规模调整
    TASKS_NUM = 15
    POINTS_PER_TASK = 10000000
    TOTAL_POINTS = TASKS_NUM * POINTS_PER_TASK

    # 创建负责任务统计的有状态函数实例
    task_summary = TaskSummary.invoke(POINTS_PER_TASK)

    # 并行执行所有任务，这里也可以不指定资源需求
    opt = yr.InvokeOptions(cpu=1000, memory=1000)
    results = [
        monte_carlo_pi.options(opt).invoke(POINTS_PER_TASK, task_summary)
        for i in range(TASKS_NUM)
    ]

    # 查询任务运行状态
    while True:
        started_num, completed_num = yr.get(task_summary.get_task_status.invoke())
        print("Total tasks:", TASKS_NUM, "Started:", started_num, "Completed", completed_num)

        if completed_num == TASKS_NUM:
            break

        time.sleep(1)

    # 计算最终结果
    circle_points = sum(yr.get(results))
    pi = (circle_points * 4) / TOTAL_POINTS
    print(f"π is: {pi}")

    yr.finalize()
```

### 运行程序

:::
:::{dropdown} 完整代码
:chevron: down-up
:icon: chevron-down

```python
import yr
import time
import random
from enum import Enum

class TaskStatus(Enum):
    PENDING = -1
    STARTED = 0
    COMPLETED = 1

@yr.invoke
def monte_carlo_pi(total_points: int, task_summary: yr.decorator.instance_proxy.InstanceProxy) -> int:
    # 上报任务已开始
    task_summary.report_status.invoke(TaskStatus.STARTED)

    circle_points = 0
    for i in range(total_points):
        rand_x = random.uniform(-1, 1)
        rand_y = random.uniform(-1, 1)

        origin_dist = rand_x**2 + rand_y**2
        if origin_dist <= 1:
            circle_points += 1

    # 上报任务已结束
    task_summary.report_status.invoke(TaskStatus.COMPLETED)
    return circle_points


@yr.instance
class TaskSummary:
    def __init__(self, total_task_num: int):
        self.total_task_num = total_task_num
        self.task_started_num = 0
        self.task_completed_num = 0

    def report_status(self, task_status: TaskStatus) -> None:
        if task_status == TaskStatus.STARTED:
            self.task_started_num += 1
        elif task_status == TaskStatus.COMPLETED:
            self.task_completed_num += 1

    def get_task_status(self) -> tuple[int, int]:
        return self.task_started_num, self.task_completed_num

if __name__ == '__main__':

    yr.init()

    # 根据集群规模调整
    TASKS_NUM = 15
    POINTS_PER_TASK = 10000000
    TOTAL_POINTS = TASKS_NUM * POINTS_PER_TASK

    # 创建负责任务统计的有状态函数实例
    task_summary = TaskSummary.invoke(TASKS_NUM)

    # 并行执行所有任务，这里也可以不指定资源需求
    opt = yr.InvokeOptions(cpu=1000, memory=1000)
    results = [
        monte_carlo_pi.options(opt).invoke(POINTS_PER_TASK, task_summary)
        for i in range(TASKS_NUM)
    ]

    # 查询任务运行状态
    while True:
        started_num, completed_num = yr.get(task_summary.get_task_status.invoke())
        print("Total tasks:", TASKS_NUM, "Started:", started_num, "Completed", completed_num)

        if completed_num == TASKS_NUM:
            break

        time.sleep(1)

    # 计算最终结果
    circle_points = sum(yr.get(results))
    pi = (circle_points * 4) / TOTAL_POINTS
    print(f"π is: {pi}")

    yr.finalize()
```

:::

在一个单节点 CPU 资源 9000（单位：1/1000 核）的 openYuanrong 环境上运行，输出如下：

```bash
Total tasks: 15 Started: 8 Completed 0
Total tasks: 15 Started: 9 Completed 0
Total tasks: 15 Started: 9 Completed 0
Total tasks: 15 Started: 9 Completed 0
Total tasks: 15 Started: 9 Completed 0
Total tasks: 15 Started: 9 Completed 0
Total tasks: 15 Started: 9 Completed 0
Total tasks: 15 Started: 9 Completed 0
Total tasks: 15 Started: 9 Completed 0
Total tasks: 15 Started: 9 Completed 0
Total tasks: 15 Started: 9 Completed 0
Total tasks: 15 Started: 9 Completed 0
Total tasks: 15 Started: 10 Completed 1
Total tasks: 15 Started: 13 Completed 4
Total tasks: 15 Started: 14 Completed 5
Total tasks: 15 Started: 15 Completed 7
Total tasks: 15 Started: 15 Completed 9
Total tasks: 15 Started: 15 Completed 9
Total tasks: 15 Started: 15 Completed 9
Total tasks: 15 Started: 15 Completed 9
Total tasks: 15 Started: 15 Completed 9
Total tasks: 15 Started: 15 Completed 9
Total tasks: 15 Started: 15 Completed 9
Total tasks: 15 Started: 15 Completed 9
Total tasks: 15 Started: 15 Completed 11
Total tasks: 15 Started: 15 Completed 13
Total tasks: 15 Started: 15 Completed 13
Total tasks: 15 Started: 15 Completed 14
Total tasks: 15 Started: 15 Completed 14
Total tasks: 15 Started: 15 Completed 15
π is: 3.14155216
```
