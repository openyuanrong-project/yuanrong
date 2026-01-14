# 日志

本节向您介绍 openYuanrong 的日志系统。

## 主机部署环境上的日志

主节点及从节点默认日志路径及文件如下，每次部署会使用时间戳生成新的日志目录例如 `20250520091445`，其中 log 目录存放组件及函数日志。您也可以参考[部署参数表](../deploy/deploy_processes/parameters.md)自定义日志路径和日志级别。

主节点：

```bash
/tmp/yr_sessions
├── 20250520091445
├── 20250520091612
└── latest
    ├── deploy_std.log
    └── log
```

从节点：

```bash
/tmp/yr_sessions
├── 20250520091445
├── 20250520091612
└── latest
    ├── deploy_std.log
    └── {node_id}
        └── log
```

### 部署日志

部署日志文件默认路径为 `/tmp/yr_sessions/latest/deploy_std.log`，您可通过该日志看到部署参数及部署过程，定位部署问题。

### 函数日志

函数日志在 log 目录下，函数标准输出默认合并为一个文件 `{node_id}-user_func_std.log`。当您在部署时配置了 `--enable_separated_redirect_runtime_std=true` 选项，每次函数执行的标准输出将分别生成一个 `runtime-{runtime_id}.out` 文件。

函数进程中也包含了部分平台逻辑比如加载代码，输出的日志为 `job-{job_id}-runtime-{runtime_id}.log` 和 `runtime-{runtime_id}.log`。在函数实例未正常拉起时，可查看该日志定位是否存在函数包路径等配置错误问题。

### 组件日志

组件日志在 log 目录下：

- function master 组件：包含 `{node_id}-function_master.log` 和 `{node_id}-function_master_std.log` 文件，仅在主节点存在。
- function scheduler 组件：包含 `{node_id}-scheduler_libruntime.log`、`faasscheduler.so-run.{timestamp}.log` 和 `{node_id}-scheduler_std.log` 文件，仅在主节点存在。
- frontend 组件：包含 `{node_id}-faas_frontend_libruntime.log`、`faasfrontend.so-run.{timestamp}.log` 和 `{node_id}-faas_frontend_std.log` 文件，仅在主节点存在。
- meta service 组件：包含 `meta-service-run.{timestamp}.log` 和 `{node_id}-metaservice_std.log` 文件，仅在主节点存在。
- dashboard 组件：包含 `dashboard-run.{timestamp}.log` 和 `{node_id}-dashboard_std.log` 文件，仅在主节点存在。
- function proxy 组件：包含 `{node_id}-function_proxy.log` 和 `{node_id}-function_proxy_std.log` 文件。
- function agent 与 runtime manager 组件：默认共进程部署，包含 `{node_id}-function_agent.log` 文件。
- data worker 组件：`worker.{INFO|WARNING}.log` 和 `ds_worker_std.log` 文件主从节点都包含。`ds_master_std.log` 和 `master/worker.{INFO|WARNING}.log` 文件仅在主节点存在。
- collector 组件：包含 `{node_id}-collector_std.log` 文件。
- etcd 组件：包含 `etcd-run.log` 文件，仅在主节点存在。

## K8s 部署环境上的日志

openYuanrong 组件及函数日志默认挂载到宿主机，其中[函数系统](glossary-functionsystem)挂载路径为 `/var/paas/sys/log/cff/default`，[数据系统](glossary-datasystem)挂载路径为 `/home/sn/datasystem/logs`。

函数系统日志：

```bash
/var/paas/sys/log/cff/default
├── componentlogs
├── processrouters
│   └── stdlogs
└── servicelogs
```

数据系统日志：

```bash
/home/sn/datasystem
└── logs
```

### 挂载的函数日志

函数日志默认路径为 `/var/paas/sys/log/cff/default/processrouters/stdlogs`。函数标准输出生成在资源池 Pod 名称对应目录下的 `function-agent-{pools_id}-xxx/{node_id}-user-func_std.log` 文件中。相同 Pod 中运行的函数，日志合并为一个文件。

函数进程中也包含了部分平台逻辑比如加载代码，输出的日志默认路径为 `/var/paas/sys/log/cff/default/servicelogs`。在相同资源池 Pod 名称目录下，每次函数运行分别生成名为 `job-{job_id}-runtime-{runtime_id}.log` 和 `runtime-{runtime_id}.log` 的文件。在函数实例未正常拉起时，可查看该日志定位是否存在函数包路径等配置错误问题。

### 挂载的组件日志

- function scheduler 组件：包括 `componentlogs` 及 `servicelogs` 下 Pod 名称为 `function-agent-xxx-faasscheduler-xxx` 子目录中的文件。
- function manager 组件：包括 `componentlogs` 及 `servicelogs` 下 Pod 名称为 `function-agent-xxx-faasmanager-xxx` 子目录中的文件。
- frontend 组件：包括 `componentlogs` 及 `servicelogs` 下 Pod 名称为 `function-agent-xxx-faasfrontend-xxx` 子目录中的文件。
- function master 组件：包含 `componentlogs` 下 Pod 名称为 `function-master-xxx` 子目录中的文件。
- meta service 组件：包含 `componentlogs` 下 Pod 名称为 `meta-service-xxx` 子目录中的文件。
- IAM adaptor 组件：包含 `componentlogs` 下 Pod 名称为 `iam-adaptor-xxx` 子目录中的文件。
- function proxy 组件：包含 `componentlogs` 下 Pod 名称为 `function-proxy-xxx` 子目录中的文件。
- function agent 与 runtime manager 组件：共 Pod 部署。在 `componentlogs` 下的每个资源池 Pod 名称对应子目录 `function-agent-{pools_id}-xxx` 中，包含 `{node_id}-function_agent.log` 和 `{node_id}-runtime_manager.log` 两个文件。
- data worker 组件：数据系统组件，包含默认路径 `/home/sn/datasystem/logs` 下 `worker` 子目录中的文件。

## 转发无状态和有状态函数的标准输出到 Driver 端

openYuanrong 支持本地程序直接调用无状态和有状态函数时，流式输出函数的 stdout 和 stderr 日志到终端。要开启此功能，可[参考示例](deploy-processes-config-log-to-driver)在部署时设置以下参数：

- `enable_dashboard`：主节点需要开启此选项。
- `enable_collector`：所有节点均需要开启此选项。
- `enable_separated_redirect_runtime_std`：所有节点均需要开启此选项。

同时，在程序初始化 openYuanrong 时开启 `log_to_driver` 配置，相关 API 及参考示例的效果如下。

- [Python API 开启日志去重](yr.Config.dedup_logs)
- [Python API 开启日志转发](yr.Config.log_to_driver)

```python
# demo.py
import yr
yr.init(yr.Config(log_to_driver=True))

@yr.instance
class A:
    def __init__(self, p):
        self.p = p

    def add(self):
        print(f'this is inside the actor')
        self.p += 1
        return self.p

insts = [A.invoke(i) for i in range(10)]
print(yr.get([i.add.invoke() for i in insts]))
yr.finalize()
```

执行 `python demo.py` 运行示例，可以看到有状态函数中的标准输出日志会在当前窗口输出，且重复日志不会反复打印，只显示重复的日志条数：

```bash
(runtime-b6110000-0000-4000-b681-68ec2a2afb86-000000008290) this is inside the actor
[1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
[repeated 9x across cluster] (runtime-bf110000-0000-4000-b36a-375ad5034385-00446c2a2d82) this is inside the actor
```

:::{note}

当前仅支持主机部署且本地程序使用 Python 开发的场景。

:::

## 使用 openYuanrong 的日志记录器

函数服务支持使用 openYuanrong 的日志记录器，您可以通过函数的上下文方法 `context.getLogger()` 打印日志，将获得和 openYuanrong 组件日志一样的输出格式，以 Java 函数服务为例，代码及日志输出如下。

```java
package com.yuanrong.demo;

import com.services.runtime.Context;
import com.services.runtime.RuntimeLogger;
import com.google.gson.JsonObject;

public class Demo {
    // 函数执行入口，每次请求都会执行，其中intput参数及函数返回类型可自定义
    public String handler(JsonObject event, Context context) {
        RuntimeLogger log = context.getLogger();
        log.info("received request,event content:" + event);

        return event.toString();
    }

    // 函数初始化入口，函数实例启动时执行一次
    public void initializer(Context context) {
        RuntimeLogger log = context.getLogger();
        log.info("function instance initialization completed");
    }
}
```

假如请求 body 设置为 {"name":"yuanrong"}，主机部署时在默认日志文件 `/tmp/yr_sessions/latest/log/runtime-{runtime_id}/user-log.log` 中可见如下日志内容。

```bash
[16:41:35:820] | [INFO] | com.services.logger.UserFunctionLogger.info(UserFunctionLogger.java:68) | Thread-1 | userLog | function instance initialization completed
[16:41:35:820] | [INFO] | com.yuanrong.executor.FaaSHandler.faasInitHandler(FaaSHandler.java:288) | Thread-1 | userLog | faas init handler complete.
[16:41:35:827] | [INFO] | com.yuanrong.executor.FaaSHandler.execute(FaaSHandler.java:174) | Thread-2 | userLog | executing udf methods, current type: InvokeFunctionStateless
[16:41:35:828] | [INFO] | com.yuanrong.executor.FaaSHandler.faasCallHandler(FaaSHandler.java:300) | Thread-2 | userLog | faas call handler called.
[16:41:35:832] | [INFO] | com.services.logger.UserFunctionLogger.info(UserFunctionLogger.java:68) | Thread-2 | userLog | received request,event content:{"name":"yuanrong"}
```

