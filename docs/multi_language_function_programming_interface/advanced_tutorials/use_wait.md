# 使用 yr.wait 限制并发/待处理任务的数量

如果发送作业的速率大于处理作业的速率，会导致作业积压在作业队列中，甚至出现 OOM。`yr.wait()` 允许反压并且可以限制待处理作业的总数，从而使作业队列不会无限扩展进而避免 OOM。

注意，该方法主要用于限制同一时间内允许执行作业的数量。该方法也可以用于限制作业并发的数量，但这会损失分发作业的性能，所以不建议这样用。openYuanrong 会根据资源的数量和作业需要的资源大小，自动分发和调整并发作业的数量。

## 使用示例

```python
import yr
import time

# 初始化 Ray
yr.init()


@yr.invoke
def heavy_computation_task(i):
    # 模拟耗时操作，例如图像处理或模型推理
    time.sleep(1)
    return f"Result from task {i}"


# --- 配置参数 ---
TOTAL_TASKS = 100
MAX_CONCURRENT_TASKS = 20  # 最大并行/在途任务数，防止 OOM
TIMEOUT = 10
WAIT_NUM = 1

# 存储正在执行的任务句柄 (Object Refs)
pending_refs = []
results = []

print(f"开始提交任务，限制最大在途任务数为: {MAX_CONCURRENT_TASKS}")

for i in range(TOTAL_TASKS):
    # 【核心逻辑】如果当前正在运行的任务达到了上限
    if len(pending_refs) >= MAX_CONCURRENT_TASKS:
        # 使用 yr.wait 阻塞，直到至少有一个任务完成
        # timeout=None 表示无限等待直到有结果返回
        ready_refs, pending_refs = yr.wait(pending_refs, wait_num=WAIT_NUM, timeout=TIMEOUT)

        # 处理已经完成的结果
        for ref in ready_refs:
            result = yr.get(ref)
            results.append(result)
            # print(f"完成并清理内存: {result}")

    # 提交新任务
    task_ref = heavy_computation_task.invoke(i)
    pending_refs.append(task_ref)

    if i % 10 == 0:
        print(f"已提交任务 {i}，当前队列负载: {len(pending_refs)}")

# --- 收尾工作 ---
# 提交完所有任务后，等待最后剩下的任务完成
print("所有任务已提交，正在等待最后剩余的任务...")
final_results = yr.get(pending_refs)
results.extend(final_results)

print(f"全部完成！成功处理了 {len(results)} 个任务。")
yr.finalize()
```
