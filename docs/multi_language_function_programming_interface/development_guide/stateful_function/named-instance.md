# 具名函数实例

用户在创建有状态函数实例时，支持指定函数实例的名字。在需要共享函数实例但无法传递实例句柄的场景具有显著的应用价值。

::::{tab-set}
:::{tab-item} Python

```python
import yr

@yr.instance
class Counter:
    def __init__(self):
        self.count = 0

    def add(self, n):
        self.count += n
        return self.count

    def get(self):
        return self.count

if __name__ == "__main__":
    yr.init()

    # 创建具名函数实例，名称为 counter-1，命名空间为 demo
    opt = yr.InvokeOptions()
    opt.name = "counter-1"
    opt.ns = "demo"
    counter = Counter.options(opt).invoke()
    result_add = counter.add.invoke(3)
    print(yr.get(result_add))

    # 可以在其他应用中使用具名函数实例，这里并不会创建一个新的函数实例，将使用上面已创建的
    counter_exist = Counter.options(opt).invoke()
    result_get = counter_exist.get.invoke()
    # 获取之前调用 add 方法写入的状态，输出 3
    print(yr.get(result_get))

    # 也可以使用 yr.get_instance 获取已创建的函数实例
    counter_query = yr.get_instance("counter-1")
    result_recheck = counter_exist.get.invoke()
    # 获取之前调用 add 方法写入的状态，输出 3
    print(yr.get(result_recheck))

    counter.terminate()
    yr.finalize()
```

:::
:::{tab-item} C++

```c++
#include <iostream>
#include "yr/yr.h"

class Counter {
public:
    Counter() : count(0) {}
    static Counter *FactoryCreate()
    {
        return new Counter();
    }

    int Add(int n)
    {
        count += n;
        return count;
    }

    int Get()
    {
        return count;
    }
    YR_STATE(count);
private:
    int count;
};

YR_INVOKE(Counter::FactoryCreate, &Counter::Add, &Counter::Get)

int main(int argc, char *argv[])
{
    YR::Init(YR::Config{}, argc, argv);
    // 创建具名函数实例，名称为 counter-1，命名空间为 demo
    auto counter = YR::Instance(Counter::FactoryCreate, "counter-1", "demo").Invoke();
    auto resultAdd = counter.Function(&Counter::Add).Invoke(3);
    std::cout << *YR::Get(resultAdd) << std::endl;

    // 可以在其他应用中使用具名函数实例，这里并不会创建一个新的函数实例，使用上面已创建的
    auto counter_exist = YR::Instance(Counter::FactoryCreate, "counter-1", "demo").Invoke();
    auto resultGet = counter_exist.Function(&Counter::Get).Invoke();
    // 获取之前调用 Add 方法写入的状态，输出 3
    std::cout << *YR::Get(resultGet) << std::endl;

    // 销毁函数实例
    counter.Terminate();
    // 释放openYuanrong 环境资源
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

    public int add(int n) {
        this.count += n;
        return this.count;
    }

    public int get() {
        return this.count;
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
import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.client.ObjectRef;


public class Main {
    public static void main(String[] args) throws YRException {
        YR.init();
        // 创建具名函数实例，名称为 counter-1，命名空间为 demo
        InstanceHandler counter = YR.instance(Counter::new, "counter-1", "demo").invoke();
        ObjectRef refAdd = counter.function(Counter::add).invoke(3);
        System.out.println(YR.get(refAdd, 9));

        // 可以在其他应用中使用具名函数实例，这里并不会创建一个新的函数实例，使用上面已创建的
        InstanceHandler counter_exist = YR.instance(Counter::new, "counter-1", "demo").invoke();
        ObjectRef refGet = counter_exist.function(Counter::get).invoke();
        // 获取之前调用 add 方法写入的状态，输出 3
        System.out.println(YR.get(refGet, 9));

        counter.terminate();
        YR.Finalize();
    }
}
```

:::
::::
