# 容错

openYuanrong 支持有状态函数实例故障时重新拉起，通过配置 `recover_retry_times` 参数控制最大重拉次数，超出重试次数后抛出异常，默认不重拉实例。

openYuanrong 通过实例故障重拉机制实现对函数实例调用的 at-least-once 语义，即用户代码中执行对函数实例的调用，openYuanrong 保证该调用至少执行一次。

## 使用限制

存在不可恢复的异常场景时，openYuanrong 可能会提前停止重试并抛出异常。

## 配置重新拉起实例

::::{tab-set}

:::{tab-item} Python

```python
import yr

@yr.instance
class Counter:
    def __init__(self):
        self.count = 0

if __name__ == "__main__":
    yr.init()

    opt =  yr.InvokeOptions()
    # 设置最大重拉次数为 3
    opt.recover_retry_times = 3
    counter = Counter.options(opt).invoke()

    counter.terminate()
    yr.finalize()
```

:::

:::{tab-item} C++

```cpp
#include <iostream>
#include "yr/yr.h"

class Counter {
public:
    Counter() : count(0) {}
    static Counter *FactoryCreate()
    {
        return new Counter();
    }
    YR_STATE(count);
private:
    int count;
};

YR_INVOKE(Counter::FactoryCreate)

int main(int argc, char *argv[])
{
    YR::Init(YR::Config{}, argc, argv);

    YR::InvokeOptions opt;
    // 设置最大重拉次数为 3
    opt.recoverRetryTimes = 3;
    auto counter = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke();

    counter.Terminate();
    YR::Finalize();
    return 0;
}
```

:::

:::{tab-item} Java

```java
// Counter.java
package com.example;

// 定义有状态函数  
public class Counter {
    private int count;

    public Counter() {
       this.count = 0;
    }
}
```

```java
// Main.java
package com.example;

import org.yuanrong.Config;
import org.yuanrong.InvokeOptions;
import org.yuanrong.api.YR;
import org.yuanrong.call.InstanceHandler;
import org.yuanrong.call.InstanceCreator;
import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.client.ObjectRef;


public class Main {
    public static void main(String[] args) throws YRException {
        YR.init();

        // 设置最大重拉次数为 3
        InvokeOptions opt = new InvokeOptions.Builder().recoverRetryTimes(3).build();
        InstanceHandler counter = YR.instance(Counter::new).options(opt).invoke();

        counter.terminate();
        YR.Finalize();
    }
}
```

:::
::::
