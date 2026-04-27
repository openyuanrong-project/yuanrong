# 并发

有状态函数支持配置并发度，单个函数实例可处理多个并发。不配置并发度的有状态函数实例默认是单线程的，对于同一个提交者的函数调用，将按顺序处理，但不同提交者间并不能保证执行顺序。您可以使用 `InvokeOptions` 接口配置有状态函数的并行度，并行度大于 1 时，不保证函数调用的处理顺序。

## 单实例单线程

默认有状态函数单实例单线程，同一个提交者的函数调用顺序执行，不同提交者间不保证顺序。

:::::{tab-set}
::::{tab-item} Python

```python
import time
import yr

@yr.instance
class Counter:
    def __init__(self):
        self.count = 0

    def add(self, n):
        self.count += n
        return self.count

# 模拟一个调用者
@yr.invoke
def caller(instance, n):
    return yr.get(instance.add.invoke(n))

@yr.invoke
def delayed_output(n):
    time.sleep(1)
    return n

yr.init()
instance = Counter.invoke()

# 在远程实例中调用
result0 = caller.invoke(instance, delayed_output.invoke(1))

# 在主程序中调用
result1 = instance.add.invoke(2)
result2 = instance.add.invoke(3)

# 输出 [6, 2, 5]
# result1 和 result2 来自同一个调用者，顺序执行，结果为 2(0+2) 和 5(2+3)
# result0 来自另一个提交者，虽然第一个提交，但因为有延迟，最后一个执行，结果为 6(5+1)
print(yr.get([result0, result1, result2]))

instance.terminate()
yr.finalize()
```

::::
::::{tab-item} C++

```cpp
#include <iostream>
#include <unistd.h>
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
    YR_STATE(count);
private:
    int count;
};
YR_INVOKE(Counter::FactoryCreate, &Counter::Add)


int Caller(YR::NamedInstance<Counter> counter, int n)
{
    return *YR::Get(counter.Function(&Counter::Add).Invoke(n));
}
YR_INVOKE(Caller)


int DelayedOutput(int n)
{
    sleep(1);
    return n;
}
YR_INVOKE(DelayedOutput)

int main(int argc, char *argv[])
{
    YR::Init(YR::Config{}, argc, argv);
    auto instance = YR::Instance(Counter::FactoryCreate).Invoke();

    // 在远程实例中调用
    auto n = YR::Function(DelayedOutput).Invoke(1);
    auto result0 = YR::Function(Caller).Invoke(instance, n);

    // 在主程序中调用
    auto result1 = instance.Function(&Counter::Add).Invoke(2);
    auto result2 = instance.Function(&Counter::Add).Invoke(3);

    // 输出 6:2:5
    // result1 和 result2 来自主程序中的调用，顺序执行，结果为 2(0+2) 和 5(2+3)
    // result0 来自另一个提交者，虽然第一个调用，但因为有延迟，最后一个执行，结果为 6(5+1)
    std::cout << *YR::Get(result0) << ":" << *YR::Get(result1) << ":" << *YR::Get(result2) << std::endl;

    instance.Terminate();
    YR::Finalize();
    return 0;
}
```

::::
::::{tab-item} Java

```java
// Counter.java
package com.example;

public class Counter {
    private int count;

    public Counter() {
       this.count = 0;
    }

    public int add(int n) {
        this.count += n;
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
    public static class Caller {
        public static int delayedOutput(int n) {
            try {
                Thread.sleep(1000);
            } catch (InterruptedException e) {
                e.printStackTrace();
            }
            return n;
        }

        public static int caller(int n) {
            // 复用主程序中的具名函数实例
            try {
                InstanceHandler counter_exist = YR.instance(Counter::new, "counter-1", "demo").invoke();
                ObjectRef refGet = counter_exist.function(Counter::add).invoke(n);
                return (int)YR.get(refGet, 9);
            } catch (YRException e) {
                e.printStackTrace();
            }
            return n;
        }
    }

    public static void main(String[] args) throws YRException {
        YR.init();
        // 创建具名函数实例，名称为 counter-1，命名空间为 demo
        InstanceHandler counter = YR.instance(Counter::new, "counter-1", "demo").invoke();
        // 在远程实例中调用
        ObjectRef n = YR.function(Caller::delayedOutput).invoke(1);
        ObjectRef result0 = YR.function(Caller::caller).invoke(n);
        // 在主程序中调用
        ObjectRef result1 = counter.function(Counter::add).invoke(2);
        ObjectRef result2 = counter.function(Counter::add).invoke(3);

        System.out.println(YR.get(result0, 9) + ":" + YR.get(result1, 9) + ":" + YR.get(result2, 9));

        counter.terminate();
        YR.Finalize();
    }
}
```

::::
:::::

## 配置单实例多线程并发

配置单实例多线程支持多并发，此时，不论是同一个提交者还是不同提交者，都不保证执行顺序。

:::::{tab-set}
::::{tab-item} Python

```python
import yr

@yr.instance
class Counter:
    def __init__(self):
        self.count = 0

    def add(self, n):
        self.count += n
        return self.count

if __name__ == "__main__":
    yr.init()

    # 配置实例并发度为 2
    opt = yr.InvokeOptions()
    opt.concurrency = 2
    counter = Counter.options(opt).invoke()

    result0 = counter.add.invoke(1)
    result1 = counter.add.invoke(2)

    # 输出可能是 1 3，也可能是 3 2
    print(yr.get(result0), yr.get(result1))

    counter.terminate()
    yr.finalize()
```

::::
::::{tab-item} C++

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
    YR_STATE(count);
private:
    int count;
};

YR_INVOKE(Counter::FactoryCreate, &Counter::Add)

int main(int argc, char *argv[])
{

    YR::Init(YR::Config{}, argc, argv);
    // 配置实例并发度为 2
    YR::InvokeOptions opt;
    opt.customExtensions.insert({"Concurrency", "2"});
    auto counter = YR::Instance(Counter::FactoryCreate).Options(opt).Invoke();

    auto result0 = counter.Function(&Counter::Add).Invoke(1);
    auto result1 = counter.Function(&Counter::Add).Invoke(2);
    // 输出可能是 1:3，也可能是 3:2
    std::cout << *YR::Get(result0) << ":" << *YR::Get(result1) << std::endl;

    counter.Terminate();
    YR::Finalize();
    return 0;
}
```

::::
::::{tab-item} Java

```java
// Counter.java
package com.example;

public class Counter {
    private int count;

    public Counter() {
       this.count = 0;
    }

    public int add(int n) {
        this.count += n;
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

import java.util.HashMap;


public class Main {
    public static void main(String[] args) throws YRException {
        YR.init();
        // 配置实例并发度为 2
        HashMap<String, String> customExtensions = new HashMap<>();
        customExtensions.put("Concurrency", "2");
        InvokeOptions opt = new InvokeOptions();
        opt.setCustomExtensions(customExtensions);
        InstanceHandler counter = YR.instance(Counter::new).options(opt).invoke();

        ObjectRef result0 = counter.function(Counter::add).invoke(1);
        ObjectRef result1 = counter.function(Counter::add).invoke(2);
        // 输出可能是 1:3，也可能是 3:2
        System.out.println(YR.get(result0, 9) + ":" + YR.get(result1, 9));

        counter.terminate();
        YR.Finalize();
    }
}
```

::::
:::::
