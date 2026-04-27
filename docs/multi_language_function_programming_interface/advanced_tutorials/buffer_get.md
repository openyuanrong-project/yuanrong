# 接口免序列化与反序列化

openYuanrong 支持在使用 `Put` 及 `invoke` 接口时直接传入 `YR::Buffer` 类型的参数，此时可免序列化。
在使用 `Get` 接口时支持直接返回 `YR::Buffer` 类型，此时可免反序列化。这种方式可以有效提高接口性能。

:::{Note}

在 `Put` 接口中传入的 `YR::Buffer` 参数不能为空指针或 size 为 `0`，否则会抛出异常 1001。

:::

## 使用场景

- `Put` 接口传入 `YR::Buffer` 类型的参数，使用 `Get` 接口获取 `YR::Buffer` 类型的返回值。
- Cpp 程序跨语言调用 Python 函数时，在 `invoke` 接口中传入 `YR::Buffer` 类型的参数，调用 `Get` 接口获取 `YR::Buffer` 类型的返回值。

## 使用示例

### 使用 Put/Get 接口

示例使用 `Put` 接口传入 `YR::Buffer` 类型的参数，然后调用 `Get` 接口获取 `YR::Buffer` 类型的返回值。

示例代码:

```c++
#include <string>

#include "yr/yr.h"

int main(int argc, char *argv[]) {
    YR::Init();

    YR::CreateParam param;
    param.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
    param.consistencyType = YR::ConsistencyType::PRAM;

    std::string str = "success";
    YR::Buffer yrBuf(str.data(), str.length());

    auto resRef = YR::Put(yrBuf, param);
    auto value = YR::Get(resRef);
    std::string result = std::string(static_cast<const char*>(value->ImmutableData()), value->GetSize());
    std::cout << result << std::endl;

    YR::Finalize();
    return 0;
}
```

### Cpp 程序中调用 Python 函数

这是一个 Cpp 程序跨语言调用 Python 函数的示例，在 `invoke` 接口中传入 `YR::Buffer` 类型参数，然后使用 `Get` 接口获取 `YR::Buffer` 类型的返回值。

示例代码：

```python
import yr

@yr.invoke
def echo(str):
    return str
```

```c++
#include <string>

#include "yr/yr.h"

int main(int argc, char *argv[]) {
    YR::Init();

    YR::CreateParam param;
    param.writeMode = YR::WriteMode::NONE_L2_CACHE_EVICT;
    param.consistencyType = YR::ConsistencyType::PRAM;

    std::string str = "success";
    YR::Buffer yrBuf(str.data(), str.length());

    auto ret = YR::PyFunction<YR::Buffer>("common", "echo")
                   .SetUrn("sn:cn:yrk:default:function:0-yr-stpython:$latest")
                   .Invoke(yrBuf);
    auto value = YR::Get(ret);
    std::string result = std::string(static_cast<const char*>(value->ImmutableData()), value->GetSize());
    std::cout << result << std::endl;

    YR::Finalize();
    return 0;
}
```
