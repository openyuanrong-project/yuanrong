# 函数服务

```{eval-rst}
.. toctree::
   :glob:
   :hidden:

   function-configuration
   version-management
```

openYuanrong 提供了函数服务能力，支持 openYuanrong 函数以 Serverless 服务方式运行，使用 HTTP 请求访问。服务实例随请求并发量全自动弹性伸缩，无请求时缩容到 0。函数服务定义了 Handler 方法作为请求处理入口，其签名如下。

:::::{tab-set}
::::{tab-item} Python

- handler：方法名称，可自定义。
- event：函数服务的请求参数，包含请求头，请求体等数据，格式为 JSON 对象。
- context：由 openYuanrong 运行时提供的上下文信息，接口介绍详见函数服务 [Python SDK](../../api/function_service/zh_cn/Python/python_sdk.rst)。

```python
def handler(event, context)
```

::::

::::{tab-item} C++

- handler：方法名称，可自定义。
- event：函数服务的请求参数，包含请求头，请求体等数据，格式为 string。
- context：由 openYuanrong 运行时提供的上下文信息，接口介绍详见函数服务 [C++ SDK](../../api/function_service/zh_cn/Cpp/cpp_sdk.rst)。

```cpp
std::string HandleRequest(const std::string &event, Function::Context &context) 
```

::::

::::{tab-item} Java

- handler：方法名称，可自定义。
- event：函数服务的请求参数，包含请求头，请求体等数据，格式为 JSON 对象。
- context：由 openYuanrong 运行时提供的上下文信息，接口介绍详见函数服务 [Java SDK](../../api/function_service/zh_cn/Java/java_sdk.md) 。

```java
public String handler(JsonObject event, Context context)
```

::::
:::::

Handler 方法返回值为字符串，以下是一个完整的方法示例：

:::::{tab-set}
::::{tab-item} Python

```python
import datetime

# 服务执行入口，每次请求都会执行
def handler(event, context):
   print("received request,event content:", event)

   response = ""
   try:
      name = event.get("name")
      # 获取配置的环境变量，环境变量在注册和更新函数时设置
      show_date = context.getUserData("show_date")
      if show_date is not None:
            response = "hello " + name + ",today is " + datetime.date.today().strftime('%Y-%m-%d')
      else:
            response = "hello " + name
   except Exception as e:
      print(e)
      response = "please enter your name,for example:{'name':'yuanrong'}"

   return response
```

::::
::::{tab-item} C++

```cpp
#include <string>
#include <ctime>
#include <nlohmann/json.hpp>

#include "Runtime.h"
#include "Function.h"
#include "yr/yr.h"

std::string HandleRequest(const std::string &event, Function::Context &context) {
    std::cout << "received request,event content:" << event << std::endl;
    std::string response = "";
    try {
        nlohmann::json jsonData = nlohmann::json::parse(event);
        // 读取 JSON 数据并输出
        std::string name = jsonData["name"];
        response += "hello ";
        response += name;
            
        std::string showDate = context.GetUserData("show_date");
        if (showDate != "") {
            time_t now = time(0);
            tm *ltm = localtime(&now);
                
            std::stringstream timeStr;
            timeStr << ltm->tm_year + 1900 << "-";
                timeStr << ltm->tm_mon + 1 << "-";
                timeStr << ltm->tm_mday;
            response += ",today is ";
            response += timeStr.str();
        }
    } catch (const std::exception& e) {
        std::cout << "JSON parsing error:" << e.what() << std::endl;
        response = "please enter your name,for example:{'name':'yuanrong'}";
    }
    return response;
}

int main(int argc, char *argv[])
{
    Function::Runtime rt;
    // 同时需要在 main() 函数中注册该方法
    rt.RegisterHandler(HandleRequest);
    rt.Start(argc, argv);
    return 0;
}
```

::::

::::{tab-item} Java

```java

import com.services.runtime.Context;
import com.google.gson.JsonObject;
import java.time.LocalDate;

public String handler(JsonObject event, Context context) {
    System.out.println("received request,event content:" + event);
    String response = "";
    try {
    String name = event.get("name").getAsString();
    // 获取配置的环境变量，环境变量在注册和更新函数时设置
    String showDate = context.getUserData("show_date");
        if (showDate != null) {
            response = "hello " + name + ",today is " + LocalDate.now();
        } else {
            response = "hello " + name;
        }
    } catch(Exception e) {
        e.printStackTrace();
        response = "please enter your name,for example:{'name':'yuanrong'}";
    }
    return response;       
}
```

::::
:::::

查看[部署 openYuanrong 服务类应用](../../../deploy/service_app_guide.md)了解如何部署函数服务。

## 函数生命周期回调

在函数实例生命周期事件发生时，可以触发相应的回调方法，包括 Initializer 和 PreStop。回调方法可根据实际业务需要选择是否实现。

### Initializer 回调

初始化（Initializer） 回调方法在函数实例启动之后，请求处理方法（Handler）之前执行。在函数实例生命周期内，成功且只成功执行一次。如果初始化方法执行失败，发送到该函数实例的请求将直接返回失败，实例自动被系统回收。

初始化方法可用于处理后端建链等逻辑，在配置单个函数实例可处理多个并发的场景，请求间可复用链路，避免重复建链，降低处理时延。

Initializer 回调方法的签名如下：

:::::{tab-set}
::::{tab-item} Python

- initializer：方法名称，可自定义。
- context：由 openYuanrong 运行时提供的上下文信息，接口介绍详见函数服务 [Python SDK](../../api/function_service/zh_cn/Python/python_sdk.rst)。

```python
def initializer(context)
```

::::
::::{tab-item} C++

- Initializer：方法名称，可自定义。
- context：由 openYuanrong 运行时提供的上下文信息，接口介绍详见函数服务 [C++ SDK](../../api/function_service/zh_cn/Cpp/cpp_sdk.rst)。

```cpp
void Initializer(Function::Context &context)
```

::::
::::{tab-item} Java

- initializer：方法名称，可自定义。
- context：由 openYuanrong 运行时提供的上下文信息，接口介绍详见函数服务 [Java SDK](../../api/function_service/zh_cn/Java/java_sdk.md)。

```java
public void initializer(Context context) 
```

::::
:::::

Initializer 方法无返回值，一个简单的示例如下：

:::::{tab-set}
::::{tab-item} Python

```python
def initializer(context):
   print("function instance initialization completed")
```

::::
::::{tab-item} C++

```cpp
std::string HandleRequest(const std::string &event, Function::Context &context) {
    return "ok";
}

void Initializer(Function::Context &context) {
    std::cout << "function instance initialization completed" << std::endl;
    return;
}

int main(int argc, char *argv[])
{
    Function::Runtime rt;
    rt.RegisterHandler(HandleRequest);
    // 同时需要在 main() 函数中注册该方法
    rt.RegisterInitializerFunction(Initializer);
    rt.Start(argc, argv);
    return 0;
}

```

::::
::::{tab-item} Java

```java
public void initializer(Context context) {
    System.out.println("function instance initialization completed");
}
```

::::
:::::

### PreStop 回调

预停止（PreStop）回调方法在函数实例退出前执行，可用于断开链路，保存持久化数据等操作。

PreStop 回调方法的签名如下：

:::::{tab-set}
::::{tab-item} Python

- pre_stop：方法名称，可自定义。

```python
def pre_stop()
```

::::
::::{tab-item} C++

- PreStop：方法名称，可自定义。

```cpp
void PreStop(Function::Context &context) 
```

::::
::::{tab-item} Java

- preStop：方法名称，可自定义。

```java
public void preStop(Context context)
```

::::
:::::

PreStop 方法无返回值，一个简单的示例如下：

:::::{tab-set}
::::{tab-item} Python

```python
def pre_stop():
   print("function instance is being destroyed")
```

::::
::::{tab-item} C++

```cpp
std::string HandleRequest(const std::string &event, Function::Context &context) {
    return "ok";
}

void PreStop(Function::Context &context) {
    std::cout << "function instance is being destroyed" << std::endl;
}

int main(int argc, char *argv[])
{
    Function::Runtime rt;
    rt.RegisterHandler(HandleRequest);
    // 同时需要在 main() 函数中注册该方法
    rt.RegisterPreStopFunction(PreStop);
    rt.Start(argc, argv);
    return 0;
}
```

::::
::::{tab-item} Java

```java
public void preStop(Context context) {
    System.out.println("function instance is being destroyed");
}
```

::::
:::::

## 函数日志

函数的各方法往标准输出 stdout 打印的日志会被 openYuanrong 收集存储，你可以使用以下方式打印日志。

### 使用 openYuanRong 的日志记录器

您可通过上下文方法 `getLogger()` 打印日志，将获得和 openYuanrong 组件一样的输出格式。每条日志中都包含时间、请求 ID 和日志级别等信息。一个简单的示例如下。

:::::{tab-set}
::::{tab-item} Python

```python
def handler(event, context):
    context.getLogger().info("hello world")
    return 'ok'
```

::::
::::{tab-item} C++

```cpp
std::string HandlerRequest(const std::string &event, Function::Context &context) {
    Function::FunctionLogger logger = context.GetLogger();
    logger.setLevel("INFO");
    logger.Info("hello world");
    return "ok";
}

int main(int argc, char *argv[])
{
    Function::Runtime rt;
    // 同时需要在 main() 函数中注册该方法
    rt.RegisterHandler(HandleRequest);
    rt.Start(argc, argv);
    return 0;
}
```

::::
::::{tab-item} Java

```java
public String handler(JsonObject event, Context context) {
    context.getLogger().info("hello world");
    return "ok";
}
```

::::
:::::

运行函数，预期输出的日志内容如下。

```bash
2025-xx-xx xx:xx:xx xxxxxxxx-xxxx-xxxx-xxxx-xxxxx****xx [INFO] hello world
```

### 使用编程语言的日志输出函数

使用编程语言日志输出函数打印的日志，内容将原样输出到日志文件中。一个简单的示例如下。

:::::{tab-set}
::::{tab-item} Python

```python
def handler(event, context):
    print('hello world')
    return 'ok'
```

::::
::::{tab-item} C++

```cpp
std::string HandlerRequest() {
    std::cout << "hello world" << std::endl;
    return "ok";
}

int main(int argc, char *argv[])
{
    Function::Runtime rt;
    // 同时需要在 main() 函数中注册该方法
    rt.RegisterHandler(HandleRequest);
    rt.Start(argc, argv);
    return 0;
}
```

::::
::::{tab-item} Java

```java
public String handler() {
    System.out.println("hello world");
    return "ok";
}
```

::::
:::::

运行服务，预期输出的日志内容如下。

```bash
hello world
```

## 实例弹性策略

openYuanrong 支持基于并发度的弹性策略，当并发度达到配置的单实例并发度时，触发函数实例扩容。函数实例 1 分钟内无请求处理时，触发缩容，函数实例可缩容到 0。

## 实例调度

openYuanrong 会根据函数指定的资源量和配置的调度策略选择合适节点运行它。详情请参阅[调度](../scheduling/index.md)章节。

## 请求调度

openYuanrong 支持为函数配置不同的请求调度策略，包括 concurrency、 round-robin 和 microservice 。

- concurrency：调度策略即为根据当前实例的并发度情况完成调度，请求会被优先调度到并发度较低的实例上。
- round-robin：调度策略即轮询调度，请求会被轮流分配给不同的实例。
- microservice：调度策略代表您没有为单个函数指定调度策略，具体使用什么调度策略将在 openYuanrong 集群部署时指定。

## 更多使用方式

- [配置函数](./function-configuration.md)
- [版本管理](./version-management.md)
