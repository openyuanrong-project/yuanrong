# 容错

openYuanrong 为无状态函数执行提供了自动容错的机制：

- 当函数实例因优先级等原因被驱逐时，调用端将自动发起无状态函数的重调度。
- 当函数实例出现网络故障时，调用端会自动进行重试。

此外，openYuanrong 支持无状态函数实例因其他原因失败如 core dump 时配置重试。

## 配置重试

::::{tab-set}

:::{tab-item} Python

```python
import yr

yr.init()

@yr.invoke
def add(n):
    print("run add")
    raise Exception('my test')

# 设置重试次数为 3
opt = yr.InvokeOptions(retry_times=3)
result = add.options(opt).invoke(1)
print(yr.get(result))

yr.finalize()
```

运行程序，在函数的标准输出[日志](../../../observability/logs.md)文件中，您将看到四行相同的输出 `run add`，表明函数执行一次失败后又重试了 `3` 次。

:::

:::{tab-item} C++

针对 C++ 语言，我们提供了重试次数和重试判断回调函数 `retryChecker` 两个配置，支持用户自定义重试。您需要同时配置 `retryTimes` 参数大于 `0`， 重试判断回调函数才能生效。

回调函数原型如下：

```c++
bool (*retryChecker)(const Exception &e) noexcept = nullptr;
```

```c++
#include <iostream>
#include <yr/yr.h>

int Add(int n)
{
    std::cout << "run Add" << std::endl;
    throw std::runtime_error("connection failed");
}
YR_INVOKE(Add)

bool RetryForConnection(const YR::Exception &e) noexcept
{
    // 返回错误码 2002（代码出现 core dump 或者异常）且错误信息包含 "connection failed" 时触发重试
    if (e.Code() == YR::ErrorCode::ERR_USER_FUNCTION_EXCEPTION) {
        std::string msg = e.what();
        if (msg.find("connection failed") != std::string::npos) {
            return true;
        }
    }
    return false;
}

int main(int argc, char *argv[])
{
    YR::Init(YR::Config{}, argc, argv);

    // 设置重试次数为 3 并配置回调函数
    YR::InvokeOptions opt;
    opt.retryTimes = 3;
    opt.retryChecker = RetryForConnection;

    auto ref = YR::Function(Add).Options(opt).Invoke(1);
    std::cout << *YR::Get(ref) << std::endl;

    YR::Finalize();
    return 0;
}
```

运行程序，在函数的标准输出日志文件中，您将看到四行相同的输出 `run Add`，表明函数执行一次失败后又重试了 `3` 次。

:::

:::{tab-item} Java

暂不支持该特性。

:::
::::
