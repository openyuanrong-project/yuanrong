# 使用有状态函数构造树状作业图

这种模式将“业务逻辑”与“任务编排”解耦，构建高效的分布式系统。

## 核心架构逻辑

在这种模式下，Driver （驱动程序） 就像是公司的 CEO，他只对接几位 Supervisor （主管）。每位主管带一个 Workers （员工） 团队。

- Driver：负责高层决策和初始化 Supervisor。
- Supervisor：负责创建 Worker、分配任务、监控状态、以及在 Worker 崩溃时进行重启（Handle failures）。
- Worker：执行具体的重体力活。

:::{Note}

- 如果 Supervisor 被销毁，那么由于引用计数的在维系有状态函数的生命周期，它管理的 Workers 也将被销毁。
- 状态函数可以嵌套多层，进而构造一个树状结构。

:::

## 以数据训练为例，构建双层结构

第一层：启动多个有状态函数。每个有状态函数负责一组特定的超参数（例如：学习率 \(0.01\)，\(0.001\)）。

第二层：每个有状态函数内部再创建一组 Worker 有状态函数。这些 Worker 共享同一组超参数，但各自处理数据的不同切片（Data
Shards），并通过 all-reduce 或参数服务器同步梯度。

### 资源管理建议

在这个模式中，资源计算变得复杂了。如果你有 32 个 CPU：

- 如果每个 Supervisor 占 1 个 CPU，它创建 3 个各占 1 CPU 的 Worker。
- 那么一个试验总共占用 4 个 CPU。
- 你最多只能同时运行 \(32 / 4 = 8\) 个超参数试验。

:::{Warning}

务必给 Supervisor 设置 num_cpus=1 或更小，否则如果你有大量超参数，所有 CPU 都会被 Supervisor 占满，导致 Worker 无法启动，发生资源死锁。

:::

### 使用示例

```python
import yr
import time

# 创建有状态函数类实例并设置其运行所需资源（1核CPU）
opt = yr.InvokeOptions(cpu=1000)


# --- 2. 底层：执行实际训练的 Worker ---
@yr.instance(invoke_options=opt)
class TrainingWorker:
    def train(self, shard_id, hyperparams):
        # 模拟训练过程
        lr = hyperparams["lr"]
        print(f"Worker {shard_id} 正在使用 lr={lr} 训练数据分片...")
        time.sleep(2)
        return f"Loss: {0.1 / lr}"


# --- 1. 中层：负责编排的 Supervisor ---
@yr.instance(invoke_options=opt)
class TrainingSupervisor:
    def __init__(self, hyperparams):
        self.hyperparams = hyperparams
        self.workers = []

    def start_training(self, num_workers):
        # 动态创建属于自己的 Worker 团队
        self.workers = [TrainingWorker.invoke() for i in range(num_workers)]

        # 调度所有 Worker 并行训练数据分片
        results = yr.get([
            w.train.invoke(i, self.hyperparams)
            for i, w in enumerate(self.workers)
        ])
        return f"超参数 {self.hyperparams} 试验完成，结果: {results}"


# --- 主控端 ---
yr.init()

# 并行启动两个不同超参数的试验（两个 Supervisor）
configs = [{"lr": 0.01}, {"lr": 0.001}]
supervisors = [TrainingSupervisor.invoke(cfg) for cfg in configs]

# 所有的试验并行运行
final_reports = yr.get([s.start_training.invoke(num_workers=2) for s in supervisors])

for report in final_reports:
    print(report)

yr.finalize()
```
