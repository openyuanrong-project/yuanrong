# 自定义优雅退出

openYuanrong 支持有状态函数实例优雅退出，即在有状态函数实例退出前，先执行一个用户自定义的函数用于清理函数状态。它在以下场景中触发：

- openYuanrong 收到对函数实例发出的 `shutdown` 请求，例如在主程序中主动调用 `terminate` 接口。
- openYuanrong 捕获到 SIGTERM 信号，例如集群执行缩容策略，会发送 SIGTERM 信号给函数实例进程。

## 使用场景

您可在优雅退出接口中执行清理、数据持久化、连接关闭等必要操作，保证服务有序退出。

## 使用示例

示例使用 Python 语言，其他开发语言参考 [C++ API](../api/distributed_programming/zh_cn/Cpp/YR_SHUTDOWN.rst) 和 [Java API](../api/distributed_programming/zh_cn/Java/yrShutdown.md)。

接口原型如下，该接口无返回值，不抛出异常。参数 `gracePeriodSecond` 表示优雅退出的超时时间（单位：s），如果超时后接口仍未返回，openYuanrong 将强制退出函数实例。超时时间默认 30s，您可以通过 `InvokeOptions` 接口自定义。

```python
def __yr_shutdown__(self, gracePeriodSecond: int):
```

:::{note}

您的自定义优雅退出函数，签名必须与接口原型一致，否则将无法触发调用。

:::

示例代码:

```python
import yr

@yr.instance
class StatefulFunc:
    def __init__(self, key):
        self.key = key

    def get_key(self):
        return self.key

    # 函数签名必须和接口原型一致
    def __yr_shutdown__(self, gracePeriodSecond):
        # 这里实现您的退出逻辑
        # 您将在函数日志中看到如下标准输出
        print("graceful exit, timeout is %ds" % gracePeriodSecond)

yr.init()

# 配置自定义超时时间为 10s
opt = yr.InvokeOptions()
opt.custom_extensions["GRACEFUL_SHUTDOWN_TIME"] = "10"
instance = StatefulFunc.options(opt).invoke("phase")
# 如下调用方式将使用默认超时时间 30s
# instance = StatefulFunc.invoke("phase")
result = instance.get_key.invoke()
print(yr.get(result))

# 将触发优雅退出
instance.terminate()
yr.finalize()
```
