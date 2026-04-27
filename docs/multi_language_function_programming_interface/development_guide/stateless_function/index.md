# 无状态函数

```{eval-rst}
.. toctree::
   :glob:
   :maxdepth: 1
   :hidden:

   concurrency
   fault-tolerance
```

无状态函数是有状态函数的特例，它的执行不依赖状态，仅依赖输入参数。openYuanrong 将在无状态函数执行结束后自动释放其占用的资源。

简单示例如下：

:::::{tab-set}
::::{tab-item} Python

```python
import yr
# 初始化openYuanrong 运行时环境
yr.init()

# 声明 add 函数可在远端执行
@yr.invoke
def add(n):
    return n + 1

# 在远端异步（不阻塞）执行 add 函数，result 是返回结果的引用
result = add.invoke(1)
# 通过 yr.get() 方法同步（阻塞）获取实际返回值，将输出 2
# 也可以使用 yr.wait() 方法，它只等待调用完成，并不直接获取结果
print(yr.get(result))

# 释放openYuanrong 环境资源
yr.finalize()
```

::::
::::{tab-item} C++

```cpp
#include <iostream>
#include <yr/yr.h>

int Add(int n)
{
    return n + 1;
}
// 声明 Add 函数可在远端执行
YR_INVOKE(Add)

int main(int argc, char *argv[])
{
    // 初始化openYuanrong 运行时环境
    YR::Init(YR::Config{}, argc, argv);

    // 在远端异步（不阻塞）执行 Add 函数，ref 是返回结果的引用
    auto ref = YR::Function(Add).Invoke(1);
    // 通过 YR::Get() 方法同步（阻塞）获取实际返回值，将输出 2
    // 也可以使用 YR::Wait() 方法，它只等待调用完成，并不直接获取结果
    std::cout << *YR::Get(ref) << std::endl;

    // 释放openYuanrong 环境资源
    YR::Finalize();
    return 0;
}
```

::::
::::{tab-item} Java

```java
import org.yuanrong.InvokeOptions;
import org.yuanrong.Config;
import org.yuanrong.api.YR;
import org.yuanrong.runtime.client.ObjectRef;

public class Main {
    public static class MyApp {
        // 定义无状态函数 add，必须为静态方法
        public static int add(int n) {
            return n + 1;
        }
    }

    public static void main(String[] args) throws Exception {
        // 初始化openYuanrong 运行时环境
        Config conf = new Config();
        YR.init(conf);

        // 在远端异步（不阻塞）执行 add 函数，ref 是返回结果的引用
        ObjectRef ref = YR.function(MyApp::add).invoke(1);
        // 通过 YR::get() 方法同步（阻塞）获取实际返回值，将输出 2
        // 也可以使用 YR.wait() 方法，它只等待调用完成，并不直接获取结果
        System.out.println(YR.get(ref, 9));

        // 释放openYuanrong 环境
        YR.Finalize();
    }
}
```

::::
:::::

## 指定无状态函数需要的资源

您可以通过指定资源来确保无状态函数在满足特定资源条件的节点上运行。使用 `InvokeOptions` 接口动态指定需要的 CPU（单位：毫核）、内存（单位：MB）或自定义资源。有关资源的更多信息，请参考[资源](../scheduling/logical_resource.md)章节。

自定义资源的类型及总量需要在部署 openYuanrong 时指定，除 GPU 和 NPU 外，openYuanrong 并不会检测其他自定义资源。参考以下示例指定节点上的自定义资源，自定义资源量仅用于在调度时做逻辑扣减，并不会限制 openYuanrong 函数对实际物理资源的使用量。

```bash
yr start --master -s 'function_agent.args.custom_resources="{\"ssd\":1}"'
```

以下为配置无状态函数的资源。

::::{tab-set}
:::{tab-item} Python

```python
import yr

@yr.invoke
def add(n):
    return n + 1

yr.init()

# 在 1 核 CPU，1G 内存，1 个自定义 ssd 的资源上运行 add 函数
opt = yr.InvokeOptions(cpu=1000, memory=1024)
opt.custom_resources={"ssd":1}
result = add.options(opt).invoke(1)
print(yr.get(result))

yr.finalize()
```

:::
:::{tab-item} C++

```c++
#include <iostream>
#include <yr/yr.h>

int Add(int n)
{
    return n + 1;
}

YR_INVOKE(Add)

int main(int argc, char *argv[])
{
    YR::Init(YR::Config{}, argc, argv);

    // 在 1 核 CPU，1G 内存，1 个自定义 ssd 的资源上运行 add 函数
    YR::InvokeOptions opt;
    opt.cpu = 1000;
    opt.memory = 1024;
    opt.customResources["ssd"] = 1;

    auto ref = YR::Function(Add).Options(opt).Invoke(1);
    std::cout << *YR::Get(ref) << std::endl;

    YR::Finalize();
    return 0;
}
```

:::
:::{tab-item} Java

```java
import org.yuanrong.InvokeOptions;
import org.yuanrong.Config;
import org.yuanrong.api.YR;
import org.yuanrong.runtime.client.ObjectRef;

public class Main {
    public static class MyApp {
        public static int add(int n) {
            return n + 1;
        }
    }

    public static void main(String[] args) throws Exception {
        Config conf = new Config();
        YR.init(conf);

        // 在 1 核 CPU，1G 内存，1 个自定义 ssd 的资源上运行 add 函数
        InvokeOptions opt = new InvokeOptions.Builder().addCustomResource("ssd", 1.0f).cpu(1000).memory(1024).build();
        ObjectRef ref = YR.function(MyApp::add).options(opt).invoke(1);
        System.out.println(YR.get(ref, 9));

        YR.Finalize();
    }
}
```

:::
::::

## 传递数据对象的引用作为无状态函数的参数

无状态函数的参数可以传值或者数据对象的引用。这个数据对象可以是 `yr.put()` 接口的返回，也可以是一个 openYuanrong 函数的返回。当无状态函数执行时，作为参数的数据对象将自动转换为值。

::::{tab-set}
:::{tab-item} Python

```python
import yr

yr.init()

@yr.invoke
def add(n):
    return n + 1

result = add.invoke(1)
# 使用前一次函数调用返回的数据对象作为参数再次调用无状态函数 add
result_next = add.invoke(result)
print(yr.get(result_next))

yr.finalize()
```

:::
:::{tab-item} C++

```c++
#include <iostream>
#include <yr/yr.h>

int Add(int n)
{
    return n + 1;
}

YR_INVOKE(Add)

int main(int argc, char *argv[])
{
    YR::Init(YR::Config{}, argc, argv);

    auto ref = YR::Function(Add).Invoke(1);
    // 使用前一次函数调用返回的数据对象作为参数再次调用无状态函数 Add
    auto ref_next = YR::Function(Add).Invoke(ref);
    std::cout << *YR::Get(ref_next) << std::endl;

    YR::Finalize();
    return 0;
}
```

:::
:::{tab-item} Java

```java
import org.yuanrong.InvokeOptions;
import org.yuanrong.Config;
import org.yuanrong.api.YR;
import org.yuanrong.runtime.client.ObjectRef;

public class Main {
    public static class MyApp {
        public static int add(int n) {
            return n + 1;
        }
    }

    public static void main(String[] args) throws Exception {
        Config conf = new Config();
        YR.init(conf);

        ObjectRef ref = YR.function(MyApp::add).invoke(1);
        // 使用前一次函数调用返回的数据对象作为参数再次调用无状态函数 add
        ObjectRef ref_next = YR.function(MyApp::add).invoke(ref);
        System.out.println(YR.get(ref_next, 9));

        YR.Finalize();
    }
}
```

:::
::::

上述示例中，第二次函数调用任务将会在第一次函数调用执行完成后才真正开始运行，因为它依赖第一次函数调用的返回结果。

两次函数调用都是异步的，并不会阻塞主程序。在调用 `yr.get(ref2)` / `YR::Get(ref2)` / `YR.get(ref2)` 同步等待返回结果前，您可以执行其他任务。

## 等待无状态函数执行结束

已经开始执行的无状态函数，可以通过 `yr.get()` 接口阻塞等待返回结果。当执行多个无状态函数时，您可能希望知道哪些已经结束，这时可以通过调用 `yr.wait()` 接口实现。

简单示例如下：

:::::{tab-set}
::::{tab-item} Python

```python
import yr

yr.init()

@yr.invoke
def add(n):
    return n + 1

results_ref = [add.invoke(i) for i in range(1)]

# 一直等待，直到有 1 个执行结束或 9 秒超时
finished_ref, unfinished_ref = yr.wait(results_ref, 1, 9)
for i in range(len(finished_ref)):
    print(yr.get(finished_ref[0]))

yr.finalize()
```

::::
::::{tab-item} C++

```cpp
#include <iostream>
#include <yr/yr.h>

int Add(int n)
{
    return n + 1;
}

YR_INVOKE(Add)

int main(int argc, char *argv[])
{
    YR::Init(YR::Config{}, argc, argv);

    std::vector<YR::ObjectRef<int>> results;
    for (int i = 0; i < 2; i++) {
        auto ref = YR::Function(Add).Invoke(i);
        results.emplace_back(ref);
    }

    // 一直等待，直到有 1 个执行结束或 9 秒超时
    auto waitResults = YR::Wait(results, 1, 9);
    for (auto ref : waitResults.first) {
        std::cout << *YR::Get(ref) << std::endl;
    }

    YR::Finalize();
    return 0;
}
```

::::
::::{tab-item} Java

```java
import org.yuanrong.Config;
import org.yuanrong.api.YR;
import org.yuanrong.runtime.client.ObjectRef;

import java.util.HashMap;
import java.util.ArrayList;
import java.util.List;

public class Main {
    public static class MyApp {
        public static int add(int n) {
            return n + 1;
        }
    }

    public static void main(String[] args) throws Exception {
        Config conf = new Config();
        YR.init(conf);

        List<ObjectRef> objectRefs = new ArrayList<>();
        for (int i = 0; i < 2; i++) {
            ObjectRef ref = YR.function(MyApp::add).invoke(i);
            objectRefs.add(ref);
        }

        // 一直等待，直到有 1 个执行结束或 9 秒超时
        WaitResult waitResult = YR.wait(objectRefs, 1, 9);
        for (ObjectRef ref : waitResult.getReady()) {
            System.out.println(YR.get(ref, 9));
        }

        YR.Finalize();
    }
}
```

::::
:::::

## 取消无状态函数的执行

您可以使用接口 `cancel()` 取消无状态函数的执行。

:::::{tab-set}
::::{tab-item} Python

```python
import time
import yr

yr.init()

@yr.invoke
def delay():
    time.sleep(10)
    return 0

result_ref = delay.invoke()
# 取消无状态函数的执行
yr.cancel(result_ref)

# 取消成功将在获取对象引用的值时抛出 RuntimeError 异常
try:
    yr.get(result_ref)
except RuntimeError as e:
    print(e)

yr.finalize()
```

::::
::::{tab-item} C++

```cpp
#include <iostream>
#include <yr/yr.h>

int Delay()
{
    std::this_thread::sleep_for(std::chrono::seconds(10));
    return 0;
}

YR_INVOKE(Delay)

int main(int argc, char *argv[])
{
    YR::Init(YR::Config{}, argc, argv);

    auto ref = YR::Function(Delay).Invoke();
    // 取消无状态函数的执行
    YR::Cancel(ref);

    // 取消成功将在获取对象引用的值时抛出 Exception 异常
    try {
        int result = *YR::Get(ref);
    } catch (YR::Exception &e) {
        std::cout << e.what() << std::endl;
    }

    YR::Finalize();
    return 0;
}
```

::::
::::{tab-item} Java

暂不支持。

::::
:::::

## 调度

openYuanrong 会根据无状态函数指定的资源量和配置的调度策略选择合适节点运行它。详情请参阅[调度](../scheduling/index.md)章节。

## 更多使用方式

- [并发](./concurrency.md)
- [容错](./fault-tolerance.md)
