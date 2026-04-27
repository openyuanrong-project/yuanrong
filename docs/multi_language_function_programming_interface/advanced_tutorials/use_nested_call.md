# 嵌套调用

嵌套调用有助于子作业的并发。它的工作原理是通过远程自动调用嵌套作业，使嵌套作业并发执行。

嵌套的作业在运行时也会消耗一定的资源，包括增加了主节点对 function agent 的负载管理以及 function agent 对函数实例的运行管理。在作业量和运行性能之间应该有所取舍，过于细粒度的作业，会降低运行的速度。

## 使用示例

```python
# demo.py
import yr

yr.init()

@yr.invoke
def recursive_process(data_list):
    # 递归终止条件
    if len(data_list) <= 10:
        return sum(data_list)

    # 将任务拆分为两个子任务（分治）
    mid = len(data_list) // 2
    left_part = data_list[:mid]
    right_part = data_list[mid:]

    # 动态提交子任务：这就是“嵌套并行”
    left_ref = recursive_process.invoke(left_part)
    right_ref = recursive_process.invoke(right_part)

    # 等待子任务结果并合并
    # 注意：在大型应用中，这里配合使用 yr.wait 可以进一步优化内存
    results = yr.get([left_ref, right_ref])
    return sum(results)

# 启动顶层任务
data = list(range(100))
final_result = yr.get(recursive_process.invoke(data))
print(f"最终计算结果: {final_result}")

yr.finalize()
```
