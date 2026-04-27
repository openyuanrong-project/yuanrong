# 避免过度并发

相比普通函数调用，并发或者分布式调用需要更多的工作负载。并发调用过于细粒度的作业可能会比普通函数调用更低效。为了避免上述问题，需谨慎设计函数并规划并发作业。如果需要并发调用一个细粒度的作业，可以使用批量处理，让每一次调用处理更多有效工作。

## 使用示例

```python
# demo.py
import time
import numpy as np
import yr

yr.init()


def process_ordinary(batch):
    time.sleep(0.00001)
    # 使用 numpy 向量化计算，极快
    return batch * 2


@yr.invoke
def process_batch(batch):
    # 使用 numpy 向量化计算，极快
    return process_ordinary(batch)


# 将数据切分为较大的块（Chunks）
datas = np.arange(1000)

start_time = time.time()
results = [process_ordinary(data) for data in datas]
end_time = time.time()
print(f"Ordinary function call takes {end_time - start_time} seconds")
# Ordinary function call takes 0.07148480415344238 seconds

start_time = time.time()
results = yr.get([process_batch.invoke(data) for data in datas])
end_time = time.time()
print(f"Parallelizing tasks takes {end_time - start_time} seconds")
# Parallelizing tasks takes 6.002047777175903 seconds

chunks = np.array_split(datas, 10)  # 拆分成 10 个大任务
start_time = time.time()
results = yr.get([process_batch.invoke(chunk) for chunk in chunks])  # 只有 10 次调度开销，计算效率极高
end_time = time.time()
print(f"Parallelizing tasks with batching takes {end_time - start_time} seconds")
# Parallelizing tasks with batching takes 0.016678571701049805 seconds

yr.finalize()
```

上述例子展示了并发调度的效率不如普通调用的效率。但是通过批量处理，使并发中的每一次调用处理了更多的作业，从而实现了预期的处理效率。
