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
- dashboard 组件：包含 `dashboard-run.{timestamp}.log` 和 `{node_id}-dashboard_std.log` 文件，仅在主节点存在。
- function proxy 组件：包含 `{node_id}-function_proxy.log` 和 `{node_id}-function_proxy_std.log` 文件。
- function agent 与 runtime manager 组件：默认共进程部署，包含 `{node_id}-function_agent.log` 文件。
- data worker 组件：`worker.{INFO|WARNING}.log` 和 `ds_worker_std.log` 文件主从节点都包含。`ds_master_std.log` 和 `master/worker.{INFO|WARNING}.log` 文件仅在主节点存在。
- collector 组件：包含 `{node_id}-collector_std.log` 文件。
- etcd 组件：包含 `etcd-run.log` 文件，仅在主节点存在。
