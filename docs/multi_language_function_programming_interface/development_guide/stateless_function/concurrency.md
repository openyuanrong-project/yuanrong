# 并发

默认情况下，一个无状态函数实例只处理一个并发请求。openYuanrong 支持您配置实例并发度，单个无状态函数实例可处理多个并发请求。单实例多并发功能适用于函数中有较多时间在等待下游服务响应的场景，等待响应一般不消耗资源。

## 使用限制

- 实例并发度配置范围为 `[1, 1000]`。
- 当配置无状态函数实例多并发时，需要确保函数代码逻辑是多线程安全的。

## 配置单实例多并发

通过 `InvokeOptions` 接口配置单实例多并发，如下示例设置函数实例的并发度是 `4`。

::::{tab-set}

:::{tab-item} Python

```python
import yr

yr.init()

@yr.invoke
def add(n):
    return n + 1

# 配置函数实例使用资源用量为 1 核 CPU，1G 内存，并发度为 4
opt = yr.InvokeOptions(cpu=1000, memory=1024)
opt.concurrency = 4

# 并发调用 8 次 add 函数
results = [add.options(opt).invoke(i) for i in range(8)]
wait_results = yr.wait(results, len(results))
print(yr.get(wait_results[0]))

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

    // 配置函数实例使用资源用量为 1 核 CPU，1G 内存，并发度为 4
    YR::InvokeOptions opt;
    opt.cpu = 1000;
    opt.memory = 1024;
    opt.customExtensions.insert({"Concurrency","4"});

    // 并发调用 8 次 add 函数
    std::vector<YR::ObjectRef<int>> results;
    for (int i = 0; i < 8; i++) {
        auto ref = YR::Function(Add).Options(opt).Invoke(i);
        results.emplace_back(ref);
    }

    auto waitResults = YR::Wait(results, results.size());
    for (auto ref : waitResults.first) {
        std::cout << *YR::Get(ref) << std::endl;
    }

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
import org.yuanrong.storage.WaitResult;

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

        // 配置函数实例使用资源用量为 1 核 CPU，1G 内存，并发度为 4
        HashMap<String, String> customExtensions = new HashMap<>();
        customExtensions.put("Concurrency", "4");
        InvokeOptions opt = new InvokeOptions();
        opt.setCustomExtensions(customExtensions);
        opt.setCpu(1000);
        opt.setMemory(1024);

        // 并发调用 8 次 add 函数
        List<ObjectRef> objectRefs = new ArrayList<>();
        for (int i = 0; i < 8; i++) {
            ObjectRef ref = YR.function(MyApp::add).options(opt).invoke(1);
            objectRefs.add(ref);
        }

        WaitResult waitResult = YR.wait(objectRefs, objectRefs.size(), -1);
        for (ObjectRef ref : waitResult.getReady()) {
            System.out.println(YR.get(ref, 9));
        }

        YR.Finalize();
    }
}
```

::::

上述示例中，函数实例的并发度为 `4`，即一个函数实例可以并行处理 `4` 个任务。代码中触发了 `8` 次函数执行，当集群中资源充足时，会启动两个函数实例，每个函数实例并行处理 `4` 个请求。

函数实例在处理完任务后并不会立即销毁，它会存活一段时间。存活期间如果有新的任务，请求仍然会调度到该实例，而不是创建出新的实例。您应该在所有任务结束后调用 `finalize` 接口立即释放其占用的资源。
