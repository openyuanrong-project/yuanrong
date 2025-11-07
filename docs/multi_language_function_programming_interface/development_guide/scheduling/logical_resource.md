# 资源

openYuanrong 对物理机器做了抽象，让您能够以资源的形式表达计算需求，系统会根据资源请求进行调度和自动扩缩容。本节将向您介绍如何指定节点和 openYuanrong 函数的资源。

openYuanrong 中的资源使用键值对表示，键表示资源名称，值是浮点数。openYuanrong 原生支持 CPU 和内存资源类型，被称为预定义资源。除此之外，还支持自定义资源如 GPU、NPU 等。

openYuanrong 中的资源为逻辑资源，根据 openYuanrong 函数指定的资源量进行调度（例如禁止 cpu=1000 的任务分配到声明逻辑 CPU 为 0 的节点），但不会限制函数对实际物理资源的使用量。您需要确保函数使用的资源量不超过它指定的值。openYuanrong 不提供 CPU 隔离机制，指定使用 1 核 CPU 资源的函数并不会独占某个物理 CPU。

## 指定节点资源

在部署 openYuanrong 时通过以下[参数](../../../deploy/deploy_processes/parameters.md)指定节点逻辑资源总量：

- `--cpu_num`：节点可用 CPU 总量，单位毫核（ 1/1000 核）。部署主节点时，系统默认占用 1 毫核。您可以在部署主节点时设置为 `--cpu_num=1`，则该节点不用于运行分布式任务。
- `--memory_num`：节点可用内存总量，单位 MB。其中一部分用于存储“数据对象”，一部分用于函数的堆栈内存。
- `--shared_memory_num`：节点可用内存总量中，用于存储“数据对象”的内存量，单位 MB。
- `--gpu_collection_enable`：自动探测节点上的 GPU 资源量。
- `--npu_collection_mode`：自动探测节点上的 NPU 资源。
- `--custom_resources`：除 GPU、NPU 外的其他自定义资源，例如 SSD。

配置示例如下，表示在主节点配置 4000 毫核 CPU（其中 3999 毫核可用于运行分布式任务），8192MB 总内存，4096MB 共享内存，开启 NPU 自动探测并包含 5 个自定义 ssd 资源。

```bash
yr start --master --custom_resources="{\"ssd\":5}" --cpu_num=4000 --memory_num=8192 --shared_memory_num=4096 --npu_collection_mode=all
```

## 指定函数实例资源

在调用无状态函数或创建有状态函数实例时，可以使用 `InvokeOptions` 接口动态指定所需资源。

::::{tab-set}
:::{tab-item} Python

```python
import yr

@yr.invoke
def add(n):
    return n + 1

yr.init()

# 配置运行无状态函数需要 1 核 CPU，1G 内存，1 个 ssd 自定义资源，1 张任意型号的 NPU 卡，1 张任意型号的 GPU 卡
opt = yr.InvokeOptions(cpu=1000, memory=1024)
opt.custom_resources={"ssd":1,"NPU/.+/count":1,"GPU/.+/count":1}
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

    // 配置运行无状态函数需要 1 核 CPU，1G 内存，1 个 ssd 自定义资源，1 张任意型号的 NPU 卡，1 张任意型号的 GPU 卡
    YR::InvokeOptions opt;
    opt.cpu = 1000;
    opt.memory = 1024;
    opt.customResources["ssd"] = 1;
    opt.customResources["NPU/.+/count"] = 1;
    opt.customResources["GPU/.+/count"] = 1;

    auto ref = YR::Function(Add).Options(opt).Invoke(1);
    std::cout << *YR::Get(ref) << std::endl;

    YR::Finalize();
    return 0;
}
```

:::
:::{tab-item} Java

```java
import com.yuanrong.InvokeOptions;
import com.yuanrong.Config;
import com.yuanrong.api.YR;
import com.yuanrong.runtime.client.ObjectRef;

import java.util.HashMap;

public class Main {
    public static class MyApp {
        public static int add(int n) {
            return n + 1;
        }
    }

    public static void main(String[] args) throws Exception {
        Config conf = new Config();
        YR.init(conf);

        // 配置运行无状态函数需要 1 核 CPU，1G 内存，1 个 ssd 自定义资源，1 张任意型号的 NPU 卡，1 张任意型号的 GPU 卡
        HashMap<String, Float> customResources = new HashMap<>();
        customResources.put("ssd", 1.0f);
        customResources.put("NPU/.+/count", 1.0f);
        customResources.put("GPU/.+/count", 1.0f);

        InvokeOptions opt = new InvokeOptions.Builder().customResources(customResources).cpu(1000).memory(1024).build();
        ObjectRef ref = YR.function(MyApp::add).options(opt).invoke(1);
        System.out.println(YR.get(ref, 3));

        YR.Finalize();
    }
}
```

:::
::::

(development-scheduling-config-npu)=

### NPU 资源的使用规则

指定函数实例时，通过三元组声明 NPU 资源，格式为 `NPU/{chip name}/{resource type}`，配置的资源量必须大于 0。

`chip name` 用于指定 NPU 芯片设备型号（例如：`Ascend310`、`Ascend910`），可以使用正则表达式进行匹配。

- 匹配多个特定型号：

    ```python
    # 匹配设备型号为 Ascend310 或 Ascend910 的所有设备
    npu_resource = {"NPU/Ascend(310|910)/count": 3}
    ```

- 匹配特定前缀的设备型号：

    ```python
    # 匹配所有以 `Ascend` 开头的设备型号（如 Ascend310、Ascend910 等）
    npu_resource = {"NPU/Ascend.*/count": 3}
    ```

- 在设备型号部分使用 `^` 会导致正则表达式从设备型号的开头开始匹配，但因为设备型号前面已经有`NPU/`，所以`^`在这里没有意义，导致匹配不正确。

    ```python
    # 无效配置
    npu_resource = {"NPU/^Ascend910.*/YY": 3}
    ```

- 缺少闭合括号，导致正则表达式解析失败，无法正确匹配。

    ```python
    # 无效配置
    npu_resource = {"NPU/(Ascend910/YY": 3}
    ```

:::{hint}

- 确保正则表达式语法正确，建议先测试验证。
- 谨慎使用过度宽泛匹配，如：`NPU/.+/count`，将匹配所有设备，可能会导致调度到非预期的设备上。请确保集群所有设备型号的 NPU 资源都能兼容您的代码。

:::

`resource type` 可以配置为 `count`、 `HBM`、`latency`、`stream`，使用说明和约束如下。

- `count` 用于指定的卡的数量。该参数与其他参数互斥，即指定 `count` 将无法设置其他参数。
- `HBM`、`latency`、`streams` 用于支持细粒度资源分配，三者需要同时配置。例如：`npu_resource = {"NPU/Ascend910/HBM":30000, "NPU/Ascend910/stream":2, "NPU/Ascend910/latency":50}`。
- `HBM` 指定该模型占用的显存量。
- `latency`：模型执行单次推理的时延（单位：ms）。
- `streams`：模型内部 stream 数量，默认为 1，如果大于 1 则不会与其他进程共卡调度。

openYuanrong 通过自动设置 `ASCEND_RT_VISIBLE_DEVICES` 环境变量实现 NPU 隔离。使用 NPU 资源时，用户需要在函数中通过环境变量 `ASCEND_RT_VISIBLE_DEVICES` 获取为该实例分配的 NPU ID，参考[示例](../../examples/use_NPU_resource.md)。

:::{attention}

- 用户请求申请使用 M 个 NPU，最多只能使用 M 个卡，超过 M 卡昇腾会抛出异常。
- 用户请求不使用 NPU 时，函数中仍可以获取并使用当前节点所有 NPU 卡。用户需保证不使用 NPU 卡，以避免对其他 NPU 实例的干扰。

:::

### GPU 资源的使用规则

指定函数实例资源时，通过三元组声明 GPU 资源，格式为 `GPU/{chip name}/count`，`chip name` 用于指定 GPU 芯片设备型号。配置的资源量必须大于 ``0``。
