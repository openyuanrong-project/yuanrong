# 函数服务中运行分布式作业

函数服务常用于开发服务类应用，无状态和有状态函数常用于开发作业类应用。实际业务场景中，一个复杂的服务请求可能需要通过执行多个分布式作业来完成，常见的方案是使用函数服务接收和解析外部请求，再调用无状态或有状态函数处理业务逻辑。我们通过蒙特卡洛方法计算 π 的示例介绍这类方案的实现。

## 准备工作

我们使用 Python 语言开发，通过定义一个无状态函数来实现打点任务，一个函数服务来接收外部请求，自定义打点任务的数量。示例部署在 openYuanrong 主机集群。

- 已在[主机上部署 openYuanrong](../../deploy/deploy_processes/index.md)并配置支持运行函数服务。
- 在所有节点创建相同的代码包目录，例如 `/opt/code`，用于存放构建生成的可执行函数代码。

## 实现流程

### 定义无状态函数处理打点任务

在代码包目录新建文件 `monte_carlo.py`，内容如下，定义无状态函数 `compute_pi`，用以处理打点任务，统计落入圆中的点的个数。

```python
import yr
import random

@yr.invoke
def compute_pi(total_points: int) -> int:
    circle_points = 0
    for i in range(total_points):
        rand_x = random.uniform(-1, 1)
        rand_y = random.uniform(-1, 1)

        origin_dist = rand_x**2 + rand_y**2
        if origin_dist <= 1:
            circle_points += 1

    return circle_points
```

使用 curl 工具注册该无状态函数，参数含义详见 [API 说明](../api/function_service/register_function.md)：

```bash
# 替换 /opt/code 为您的代码包目录
META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为 http://{主节点 ip}:31182>
curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -d '{"name":"0-baas-task","runtime":"python3.9","kind":"yrlib","cpu":600,"memory":512,"timeout":60,"storageType":"local","codePath":"/opt/code"}'
```

记录返回格式中 `id` 字段的值用于在 FaaS 函数中使用，这里对应 `sn:cn:yrk:default:function:0-baas-task:$latest`。

```bash
{"code":0,"message":"SUCCESS","function":{"id":"sn:cn:yrk:default:function:0-baas-task:$latest","createTime":"2025-04-28 11:31:20.986 UTC","updateTime":"","functionUrn":"sn:cn:yrk:default:function:0-baas-task","name":"0-baas-task","tenantId":"default","businessId":"yrk","productId":"","reversedConcurrency":0,"description":"","tag":null,"functionVersionUrn":"sn:cn:yrk:default:function:0-baas-task:$latest","revisionId":"20250428113120986","codeSize":0,"codeSha256":"","bucketId":"","objectId":"","handler":"","layers":null,"cpu":600,"memory":512,"runtime":"python3.9","timeout":60,"versionNumber":"$latest","versionDesc":"$latest","environment":{},"customResources":null,"statefulFlag":0,"lastModified":"","Published":"2025-04-28 11:31:20.986 UTC","minInstance":0,"maxInstance":100,"concurrentNum":100,"funcLayer":[],"status":"","instanceNum":0,"device":{},"created":""}}
```

### 定义函数服务接收外部请求

在代码包目录下新建文件 `service_entry.py`，内容如下。在该函数服务的 `handler` 接口中会解析外部请求配置的 `tasksNumber` 参数，设置打点任务个数并调用无状态函数 `compute_pi` 运行打点任务。

```python
import yr
import monte-carlo

def handler(event, context):
    tasks_number = event.get("tasksNumber")

    POINTS_PER_TASK = 10000000
    TOTAL_POINTS = tasks_number * POINTS_PER_TASK

    # 这里动态指定了调用无状态函数 compute_pi 时使用的资源量，如果不指定，将使用注册该函数元数据时配置的资源（cpu 600 毫核，内存 512 MiB）
    opt = yr.InvokeOptions(cpu=1000, memory=1000)
    results = [
        monte_carlo.compute_pi.options(opt).invoke(POINTS_PER_TASK)
        for i in range(tasks_number)
    ]

    # 计算最终结果
    circle_points = sum(yr.get(results))
    pi = (circle_points * 4) / TOTAL_POINTS
    print(f"π is: {pi}")

    return pi

def init(context):
    # 配置无状态函数 compute_pi 的注册 id
    conf = yr.Config(function_id="sn:cn:yrk:default:function:0-baas-task:$latest")
    yr.init(conf)
```

使用 curl 工具注册函数，参数含义详见 [API 说明](../api/function_service/register_function.md)：

```bash
# 替换 /opt/code 为您的代码包目录
META_SERVICE_ENDPOINT=<meta service 组件的终端节点，默认为 http://{主节点 ip}:31182>
curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -d '{"name":"0@faas@demo","runtime":"python3.9","handler":"service_entry.handler","kind":"faas","cpu":600,"memory":512,"timeout":60,"extendedHandler":{"initializer":"service_entry.init"},"extendedTimeout":{"initializer":30},"storageType":"local","codePath":"/opt/code"}'
```

记录返回格式中 `functionVersionUrn` 字段的值用于调用，这里对应 `sn:cn:yrk:default:function:0@faas@demo:latest`。

```bash
{"code":0,"message":"SUCCESS","function":{"id":"sn:cn:yrk:default:function:0@faas@demo:latest","createTime":"2025-04-28 12:02:21.930 UTC","updateTime":"","functionUrn":"sn:cn:yrk:default:function:0@faas@demo","name":"0@faas@demo","tenantId":"default","businessId":"yrk","productId":"","reversedConcurrency":0,"description":"","tag":null,"functionVersionUrn":"sn:cn:yrk:default:function:0@faas@demo:latest","revisionId":"2025042812022193","codeSize":0,"codeSha256":"","bucketId":"","objectId":"","handler":"service_entry.handler","layers":null,"cpu":600,"memory":512,"runtime":"python3.9","timeout":60,"versionNumber":"latest","versionDesc":"latest","environment":{},"customResources":null,"statefulFlag":0,"lastModified":"","Published":"2025-04-28 12:02:21.930 UTC","minInstance":0,"maxInstance":100,"concurrentNum":100,"funcLayer":[],"status":"","instanceNum":0,"device":{},"created":""}}
```

### 测试应用

使用 curl 工具调用函数服务 `0@faas@demo`，参数含义详见 [API 说明](../api/function_service/function_invocation.md)：

```bash
FRONTEND_ENDPOINT=<frontend 组件的终端节点，默认为 http://{主节点 ip}:8888>
FUNCTION_VERSION_URN=<前一步骤记录的 functionVersionUrn>
curl -H "Content-type: application/json" -X POST -i ${FRONTEND_ENDPOINT}/serverless/v1/functions/${FUNCTION_VERSION_URN}/invocations -d '{"tasksNumber":2}'
```

结果输出：

```bash
3.1412376
```
