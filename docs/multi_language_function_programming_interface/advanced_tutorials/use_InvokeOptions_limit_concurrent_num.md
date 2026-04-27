# 使用资源用量限制任务并发数量

作业（无状态和有状态）默认占用 `cpu` 500 毫核和 `memory` 500 MiB。当一个作业需要的算力超过配置好的资源量时，很可能受其他并发执行作业的影响，导致该作业执行变慢，但它依然可以稳定地执行完毕。

这些作业对内存的占用会同比例增长，而这会导致总的资源用量超出节点的内存阈值或者导致 OOM。为了避免这个问题，可通过配置更大的资源量，降低单节点上的并发作业数量，为作业的执行提供资源冗余量。openYuanrong 的调度机制会保证所有并发执行的作业请求的资源总量不超过节点的阈值。

## 使用示例

示例展示了一种使用 openYuanrong 的分布式调用加载大文件到内存中并处理数据的方法。当这种作业并发量过大，会导致 OOM。使用 openYuanrong 显式地为每个函数设置 `memory` 和 `cpu` 资源大小，以限制函数的并发数。

:::{note}

内存资源是逻辑定义，不是物理定义。openYuanrong 不会强制作业执行时内存用量在 `memory` 显示请求量以内。

:::

```python
import yr
import time

# 初始化 yr
yr.init()

# 1GB = 1024 * 1024 * 1024 bytes
# 模拟一个处理大文件的函数
@yr.invoke
def process_large_file(file_name):
    print(f"正在处理文件: {file_name}，预计占用内存 > 3GB...")

    # 模拟加载 3GB 数据到堆内存中
    # 在实际场景中，这可能是 pd.read_csv() 或 np.load()
    data = "X" * (3 * 1024 * 1024 * 1024)

    time.sleep(5)  # 模拟计算耗时
    result_size = len(data)

    # 释放大对象引用，帮助 GC 及时回收
    del data
    return f"文件 {file_name} 处理完成，大小: {result_size // (1024**3)} GB"

# --- 调度逻辑 ---
file_list = [f"file_{i}.dat" for i in range(10)]
pending_refs = []
max_in_flight = 3

# 我们设置逻辑需求为 4GB，为内存抖动留出空间
opt = yr.InvokeOptions()
opt.memory = 4 * 1024

print("开始分批处理大文件...")

for file in file_list:
    # 如果在途任务太多，先等待完成一个，腾出逻辑资源位
    if len(pending_refs) >= max_in_flight:
        ready, pending_refs = yr.wait(pending_refs, wait_num=1)
        print(f"完成任务: {yr.get(ready[0])}")

    # 提交任务
    # yr 会检查当前节点剩余的 'memory' 资源是否满足 4GB
    # 如果不满足，此任务会保持 Pending 状态直到其他任务释放资源
    ref = process_large_file.options(opt).invoke(file)
    pending_refs.append(ref)

# 等待剩余任务
results = yr.get(pending_refs)
print("所有大文件处理完毕！")
yr.finalize()
```
