(overview-getting-started)=

# 入门

openYuanrong 是一个 Serverless 分布式计算引擎，致力于以一套统一 Serverless 架构支持 AI、大数据、微服务等各类分布式应用。它提供多语言函数编程接口，以单机编程体验简化分布式应用开发；提供分布式动态调度和数据共享等能力，实现分布式应用的高性能运行和集群的高效资源利用。

![](./images/introduction.png)

openYuanrong 由多语言函数运行时、函数系统和数据系统组成，支持按需灵活单独或组合使用。

- **多语言函数运行时**：提供函数分布式编程，支持 Python、Java、C++ 语言，实现类单机编程高性能分布式运行。
- **函数系统**：提供大规模分布式动态调度，支持函数实例极速弹性扩缩和跨节点迁移，实现集群资源高效利用。
- **数据系统**：提供异构分布式多级缓存，支持 Object、Stream 语义，实现函数实例间高性能数据共享及传递。

**函数**是 openYuanrong 的核心概念抽象，是 openYuanrong 分布式调度运行的基本单位。相比传统 Serverless 函数概念，openYuanrong 函数更加通用，支持运行中动态创建、长时运行、相互间异步调用、有状态等等，可以表达任意分布式应用的运行实例，起到类似单机 OS 中进程的作用。[了解 openYuanrong 的更多概念抽象](./multi_language_function_programming_interface/key_concept.md)。

## 选择您的入门指南

按需选择入门指南，开始使用 openYuanrong。

- 开发分布式应用：[多语言函数编程接口快速入门](quickstart-multilingual-functional-programming-interface)
- 安装部署：[安装部署快速入门](quickstart-deployment)
- 调试监控应用：[监控与调试快速入门](quickstart-observability)

(quickstart-multilingual-functional-programming-interface)=

## 多语言函数编程接口快速入门

使用多语言函数编程接口，仅需几行代码即可将常规 Python、C++、Java 函数和类转换为 openYuanrong [无状态函数](key-concept-stateless-function)和[有状态函数](key-concept-statefull-function)，轻松构建分布式应用。

通过以下示例向您展示：

1. 将 Python、C++、Java 函数转换为 openYuanrong 无状态函数并行执行。
2. 将 Python、C++、Java 类转换为 openYuanrong 有状态函数进行分布式有状态计算。

::::::{dropdown} 使用 openYuanrong 无状态函数并行化 Python、C++、Java 函数
:color: success
:chevron: down-up
:icon: chevron-down

:::{Note}

运行以下示例，需先安装 openYuanrong，环境要求如下：

- 安装 openYuanrong 并运行 Python 示例：`python<=3.11,>=3.9` 。
- 运行 Java 示例：`java 8/17/21` 。
- 运行 C++ 示例：`gcc>=10.3.0 且 stdc++>=14` 。

```bash
pip install https://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/release/0.6.0/linux/x86_64/openyuanrong-0.6.0-cp39-cp39-manylinux_2_34_x86_64.whl
```

:::

:::::{tab-set}
::::{tab-item} Python

引入 openYuanrong SDK 并调用 `yr.init()` 初始化。使用 `yr.invoke` 装饰一个函数，声明该函数可以在远端运行。调用时使用 `.invoke()` 触发远端运行，返回结果是数据对象的一个引用，需要通过 `yr.get()` 获取值。程序结束时，使用 `yr.finalize()` 清理运行上下文。

```python
# example.py
import yr

# 定义无状态函数
@yr.invoke
def say_hello(name):
    return 'hello, ' + name


# Init（初始化）一次
yr.init()

# 并行异步调用无状态函数
results_ref = [say_hello.invoke('yuanrong') for i in range(3)]
print(yr.get(results_ref))

# 释放环境资源
yr.finalize()
```

运行程序。

```bash
python example.py
# ['hello, yuanrong', 'hello, yuanrong', 'hello, yuanrong']
```

::::
::::{tab-item} C++

引入头文件 `yr/yr.h` 并调用 `YR::Init()` 初始化。使用 `YR_INVOKE()` 宏声明该函数可以在远端运行。调用时使用 `YR::Function().Invoke()` 触发远端运行，返回结果是数据对象的一个引用，需要通过 `YR::Get()` 获取值。程序结束时，使用 `YR::Finalize()` 清理运行上下文。

```cpp
#include <iostream>
#include "yr/yr.h"

// 定义无状态函数
std::string SayHello(std::string name)
{
    return "hello, " + name;
}
YR_INVOKE(SayHello)


int main(int argc, char *argv[])
{
    // Init （初始化） 一次
    YR::Init(YR::Config{}, argc, argv);

    // 并行异步调用无状态函数
    std::vector<YR::ObjectRef<std::string>> results_ref;
    for (int i = 0; i < 3; i++) {
        auto result_ref = YR::Function(SayHello).Invoke(std::string("yuanrong"));
        results_ref.emplace_back(result_ref);
    }

    for (auto result : YR::Get(results_ref)) {
        std::cout << *result << std::endl;
    }

    // 释放环境资源
    YR::Finalize();
    return 0;
}
```

参考[无状态函数示例工程](example-project-stateless-function)运行该程序。

::::
::::{tab-item} java

调用 `YR.init()` 初始化 openYuanrong。使用 `YR.function().invoke()` 触发 Java 静态方法在远端运行，返回结果是数据对象的一个引用，需要通过 `YR.get()` 获取值。程序结束时，使用 `YR::Finalize()` 清理运行上下文。

```java
// Greeter.java
package com.yuanrong.example;

public class Greeter {
    // 定义无状态函数
    public static String sayHello(String name) {
        return "hello, " + name;
    }
}
```

```java
// Main.java
package com.yuanrong.example;

import com.yuanrong.Config;
import com.yuanrong.api.YR;
import com.yuanrong.runtime.client.ObjectRef;
import com.yuanrong.exception.YRException;

import java.util.HashMap;
import java.util.ArrayList;
import java.util.List;

public class Main {
    public static void main(String[] args) throws YRException {
        // Init （初始化） 一次
        YR.init(new Config());

        // 并行异步调用无状态函数
        List<ObjectRef> resultsRef = new ArrayList<>();
        for (int i = 0; i < 3; i++) {
            ObjectRef resultRef = YR.function(Greeter::sayHello).invoke("yuanrong");
            resultsRef.add(resultRef);
        }

        for (ObjectRef resultRef : resultsRef) {
            System.out.println(YR.get(resultRef, 30));
        }

        // 并行异步调用无状态函数
        YR.Finalize();
    }
}
```

参考[无状态函数示例工程](example-project-stateless-function)运行该程序。

::::
:::::

无状态函数适合处理无需维护状态的应用，对于需要维护状态的应用，可使用 openYuanrong 有状态函数。

::::::

::::::{dropdown} 使用 openYuanrong 有状态函数并行化 Python、C++、Java 类
:color: success
:chevron: down-up
:icon: chevron-down

openYuanrong 提供了有状态函数，让您可以并行化类的实例。当您实例化了一个有状态函数时，openYuanrong 会在集群中启动该类的远程实例。然后该有状态函数可以执行远程方法调用并维护其自身的内部状态。

:::{Note}

运行以下示例，需先安装 openYuanrong，环境要求如下：

- 安装 openYuanrong 并运行 Python 示例：`python<=3.11,>=3.9` 。
- 运行 Java 示例：`java 8/17/21` 。
- 运行 C++ 示例：`gcc>=10.3.0 且 stdc++>=14` 。

```bash
pip install https://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/release/0.6.0/linux/x86_64/openyuanrong-0.6.0-cp39-cp39-manylinux_2_34_x86_64.whl
```

:::

:::::{tab-set}
::::{tab-item} Python

```python
# example.py
import yr

# 定义有状态函数
@yr.instance
class Object:
    def __init__(self):
        self.value = 0

    def save(self, value):
        self.value = value

    def get(self):
        return self.value


# Init（初始化）一次
yr.init()

# 创建三个有状态函数实例
objs = [Object.invoke() for i in range(3)]

# 并行异步调用有状态函数
[obj.save.invoke(9) for obj in objs]
results_ref = [obj.get.invoke() for obj in objs]
print(yr.get(results_ref))

# 销毁有状态函数实例
[obj.terminate() for obj in objs]

# 创建三个有状态函数实例
yr.finalize()
```

运行程序，输出如下。

```bash
python example.py
# [9, 9, 9]
```

::::
::::{tab-item} C++

```cpp
#include <iostream>
#include "yr/yr.h"

// 定义有状态函数
class Object {
public:
    Object() {}
    Object(int value) { this->value = value; }

    static Object *FactoryCreate(int value) {
        return new Object(value);
    }

    void Save(int value) {
        this->value = value;
    }

    int Get() {
        return value;
    }

    YR_STATE(value);

private:
    int value;
};
YR_INVOKE(Object::FactoryCreate, &Object::Save, &Object::Get);


int main(int argc, char *argv[])
{
    // Init（初始化）一次
    YR::Init(YR::Config{}, argc, argv);

    // 并行异步调用有状态函数
    std::vector<YR::NamedInstance<Object>> objects;
    for (int i = 0; i < 3; i++) {
        auto obj = YR::Instance(Object::FactoryCreate).Invoke(0);
        objects.emplace_back(obj);
    }

    // 并行异步调用有状态函数
    std::vector<YR::ObjectRef<int>> results_ref;
    for (auto obj : objects) {
        obj.Function(&Object::Save).Invoke(9);
        auto result_ref = obj.Function(&Object::Get).Invoke();
        results_ref.emplace_back(result_ref);
    }

    for (auto result : YR::Get(results_ref)) {
        std::cout << *result << std::endl;
    }

    // 销毁有状态函数实例
    for (auto obj : objects) {
        obj.Terminate();
    }

    // 释放环境资源
    YR::Finalize();
    return 0;
}
```

参考[有状态函数示例工程](example-project-stateful-function)运行该程序。

::::
::::{tab-item} java

```java
// Object.java
package com.yuanrong.example;

// 定义有状态函数
public class Object {
    private int value = 0;

    public Object(int value) {
        this.value = value;
    }

    public void save(int value) {
        this.value = value;
    }

    public int get() {
        return this.value;
    }
}
```

```java
// Main.java
package com.yuanrong.example;

import com.yuanrong.Config;
import com.yuanrong.api.YR;
import com.yuanrong.runtime.client.ObjectRef;
import com.yuanrong.call.InstanceHandler;
import com.yuanrong.exception.YRException;

import java.util.HashMap;
import java.util.ArrayList;
import java.util.List;

public class Main {
    public static void main(String[] args) throws YRException {
        // Init（初始化）一次
        YR.init(new Config());

        // 创建三个有状态函数实例
        List<InstanceHandler> objects = new ArrayList<>();
        for (int i = 0; i < 3; i++) {
            InstanceHandler obj = YR.instance(Object::new).invoke(0);
            objects.add(obj);
        }

        // 并行异步调用有状态函数
        List<ObjectRef> resultsRef = new ArrayList<>();
        for (InstanceHandler obj : objects) {
            obj.function(Object::save).invoke(9);
            ObjectRef resultRef = obj.function(Object::get).invoke();
            resultsRef.add(resultRef);
        }

        for (ObjectRef resultRef : resultsRef) {
            System.out.println(YR.get(resultRef, 30));
        }

        // 销毁有状态函数实例
        for (InstanceHandler obj : objects) {
            obj.terminate();
        }

        // 释放环境资源
        YR.Finalize();
    }
}
```

参考[有状态函数示例工程](example-project-stateful-function)运行该程序。

::::
:::::
::::::

[了解多语言函数编程接口的更多信息](./multi_language_function_programming_interface/index.md)。

(quickstart-deployment)=

## 安装部署快速入门

- openYuanrong 支持在单台主机上部署和运行，用于学习和开发，生产时可无缝扩展到大型集群：[了解更多](./deploy/index.md)。

(quickstart-observability)=

## 监控与调试快速入门

- openYuanrong 提供了可观测性工具来监控集群和调试应用。这些工具可帮助您了解应用程序的性能并识别瓶颈：[了解更多](./observability/index.md)。
