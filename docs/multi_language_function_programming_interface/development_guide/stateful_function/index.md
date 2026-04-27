# 有状态函数

```{eval-rst}
.. toctree::
   :glob:
   :maxdepth: 1
   :hidden:

   named-instance
   parallel-execution
   fault-tolerance
```

一个有状态函数本质上是一个有状态的工作进程（或服务）。当你实例化一个有状态函数时，openYuanrong 会创建一个新的工作进程，并将有状态函数的方法调用调度到该工作进程上，通过这些方法可访问和修改工作进程的状态。

对不同有状态函数实例的方法调用并行执行，在同一个有状态函数实例上的方法调用按顺序串行执行并共享状态。

:::{Note}

openYuanrong 并不会主动回收有状态函数实例，您需要在所有任务结束时主动调用 `terminate` 方法销毁，否则会产生资源泄露。

:::

::::{tab-set}
:::{tab-item} Python

```python
import yr

# 声明有状态函数
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
    # 初始化openYuanrong 运行时环境
    yr.init()

    # 创建有状态函数实例
    counter = Counter.invoke()

    # 异步（不阻塞）调用有状态函数的方法与其交互
    result_add = counter.add.invoke(3)
    result_get = counter.get.invoke()

    # 同步（阻塞）获取结果，输出 3 3
    # 也可以使用 yr.wait() 方法，它只等待调用完成，并不直接获取结果
    print(yr.get(result_add), yr.get(result_get))

    # 销毁函数实例
    counter.terminate()
    # 释放openYuanrong 环境资源
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
// 声明有状态函数
YR_INVOKE(Counter::FactoryCreate, &Counter::Add, &Counter::Get)

int main(int argc, char *argv[])
{
    // 初始化openYuanrong 运行时环境
    YR::Init(YR::Config{}, argc, argv);
    // 创建有状态函数实例
    auto counter = YR::Instance(Counter::FactoryCreate).Invoke();
    // 异步（不阻塞）调用有状态函数的方法与其交互
    auto resultAdd = counter.Function(&Counter::Add).Invoke(3);
    auto resultGet = counter.Function(&Counter::Get).Invoke();

    // 同步（阻塞）获取结果，输出 3:3
    // 也可以使用 YR::Wait() 方法，它只等待调用完成，并不直接获取结果
    std::cout << *YR::Get(resultAdd) << ":" << *YR::Get(resultGet) << std::endl;

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
        // 初始化openYuanrong 运行时环境
        YR.init();
        // 创建有状态函数实例
        InstanceHandler counter = YR.instance(Counter::new).invoke();
        // 异步（不阻塞）调用有状态函数的方法与其交互
        ObjectRef refAdd = counter.function(Counter::add).invoke(3);
        ObjectRef refGet = counter.function(Counter::get).invoke();

        // 同步（阻塞）获取结果，输出 3:3
        // 也可以使用YR.wait()方法，它只等待调用完成，并不直接获取结果
        System.out.println(YR.get(refAdd, 9) + ":" + YR.get(refGet, 9));
        // 销毁函数实例
        counter.terminate();
        // 释放openYuanrong 环境资源
        YR.Finalize();
    }
}
```

:::
::::

## 指定有状态函数需要的资源

实例化有状态函数时，可以动态配置其资源。不配置时，默认资源为 `cpu` 500 毫核，`memory` 500 MiB。其他自定义资源（例如 NPU、GPU 等）可通过字段 `custom_resources` 以键值对的方式配置。有关资源的更多信息，请参考[资源](../scheduling/logical_resource.md)章节。

自定义资源的类型及总量需要在部署 openYuanrong 时指定，除 GPU 和 NPU 外，openYuanrong 并不会检测其他自定义资源。参考以下示例指定节点上的自定义资源，自定义资源量仅用于在调度时做逻辑扣减，并不会限制 openYuanrong 函数对实际物理资源的使用量。

```bash
yr start --master -s 'function_agent.args.custom_resources="{\"ssd\":1}"'
```

以下为配置有状态函数的资源。

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

    # 在 1 核 CPU，1G 内存，1 个自定义 ssd 的资源上运行 Counter 函数
    opt = yr.InvokeOptions()
    opt.cpu = 1000
    opt.memory = 1024
    opt.custom_resources = {"ssd": 1}
    counter = Counter.options(opt).invoke()
    result_add = counter.add.invoke(3)
    print(yr.get(result_add))

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

    // 在 1 核 CPU，1G 内存，1 个自定义 ssd 的资源上运行 Counter 函数
    YR::InvokeOptions opt;
    opt.cpu = 1000;
    opt.memory = 1024;
    opt.customResources["ssd"] = 1;

    auto counter = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke();
    auto resultAdd = counter.Function(&Counter::Add).Invoke(3);
    std::cout << *YR::Get(resultAdd) << std::endl;

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
import org.yuanrong.call.InstanceCreator;
import org.yuanrong.exception.YRException;
import org.yuanrong.runtime.client.ObjectRef;


public class Main {
    public static void main(String[] args) throws YRException {
        YR.init();

        // 在 1 核 CPU，1G 内存，1 个自定义 ssd 的资源上运行 Counter 函数
        InvokeOptions opt = new InvokeOptions.Builder().addCustomResource("ssd", 1.0f).cpu(1000).memory(1024).build();
        InstanceHandler counter = YR.instance(Counter::new).options(opt).invoke();
        ObjectRef refAdd = counter.function(Counter::add).invoke(3);
        System.out.println(YR.get(refAdd, 9));

        counter.terminate();
        YR.Finalize();
    }
}
```

:::
::::

## 传递有状态函数的句柄

有状态函数的句柄可以作为参数传递给其他无状态函数或有状态函数，在函数中可使用该句柄调用有状态函数的方法。

::::{tab-set}
:::{tab-item} Python

```python
import yr

@yr.invoke
def get(counter):
    return yr.get(counter.get.invoke())

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

    counter = Counter.invoke()
    result_add = counter.add.invoke(3)
    print(yr.get(result_add))

    result_get = get.invoke(counter)
    # 输出 3
    print(yr.get(result_get))

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

int Get(YR::NamedInstance<Counter> counter)
{
    return *YR::Get(counter.Function(&Counter::Get).Invoke());
}
YR_INVOKE(Get)

int main(int argc, char *argv[])
{
    YR::Init(YR::Config{}, argc, argv);
    auto counter = YR::Instance(Counter::FactoryCreate).Invoke();
    auto resultAdd = counter.Function(&Counter::Add).Invoke(3);
    std::cout << *YR::Get(resultAdd) << std::endl;

    auto resultGet = YR::Function(Get).Invoke(counter);
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

暂不支持该特性。

:::
::::

## 在有状态函数间通信

有状态函数实例间通信可通过函数中的方法调用完成，数据可以通过[数据对象](../data_object/index.md)共享或数据流传递，以实现协同。

## 调度

openYuanrong 会根据有状态函数指定的资源量和配置的调度策略选择合适节点运行它。详情请参阅[调度](../scheduling/index.md)章节。

## 更多使用方式

- [具名实例](./named-instance.md)
- [并发](./parallel-execution.md)
- [容错](./fault-tolerance.md)
- [自定义优雅退出](../../advanced_tutorials/yr_shutdown.md)
