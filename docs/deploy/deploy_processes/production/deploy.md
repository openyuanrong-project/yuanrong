# 部署 openYuanrong

本节将介绍常见场景下主机部署 openYuanrong 时的配置，配置参数详细说明请参考[部署参数表](../parameters.md)。

## 使用命令行工具 yr 部署 openYuanrong

### 用法

`yr [OPTIONS] COMMAND [ARGS]...`

### 描述

`yr` 是 Python 版实验命令行工具，用于主机部署 openYuanrong，并提供配置渲染、组件编排、健康检查、状态查询与停止能力。

在执行 `yr start` 时，CLI 会：

- 合并配置（默认值 + 用户配置 + 命令行覆盖）并渲染模板。
- 创建部署目录 `/tmp/yr_sessions/<timestamp>/`，更新符号链接 `/tmp/yr_sessions/latest`。
- 按依赖顺序启动组件并执行健康检查。
- 生成会话文件 `/tmp/yr_sessions/latest/session.json`，供 `status` / `health` / `stop` 使用。

在 **master** 模式启动成功后，会打印 worker 节点加入集群的推荐命令和最小配置片段。

### 配置

**配置如何合并**

`yr` 在启动时，会通过合并以下内容来生成最终生效的配置：

优先级从高到低：

1. 命令行覆盖项（`yr start ... -s KEY=VALUE`，可重复指定）
2. 用户配置文件（默认：`/etc/yuanrong/config.toml`，若存在）
3. 内置默认值

#### 基于内置 config.toml.jinja 编写自定义 config.toml

Python `yr` CLI 的最终组件配置由内置模板渲染得到：

- `api/python/yr/cli/values.toml`
- `api/python/yr/cli/config.toml.jinja`

建议流程：

1. 执行 `yr config template` 查看内置模板结构（`values.toml` + `config.toml.jinja`）。
2. 新建自己的 `config.toml`，只写需要覆盖的字段（无需复制整份模板）。
3. 使用 `yr -c /path/to/config.toml start --master`（或 `yr -c ... start`）启动。

示例：在主节点启用 frontend，并覆盖 `function_proxy` 与 `ds_worker` 的启动参数：

```toml
[mode.master]
frontend = true

[values]
log_level = "DEBUG"

[values.function_proxy]
ip = "10.88.0.4"
port = "22772"
grpc_listen_port = "22773"

[function_proxy.args]
enable_metrics = false
litebus_thread_num = 25
runtime_recover_enable = true

[ds_worker.args]
heartbeat_interval_ms = 120000
node_dead_timeout_s = 120
```

执行：

`yr -c /path/to/config.toml start --master`

如需在不改文件的情况下临时覆盖，可叠加 `-s/--set`（例如 `-s 'function_proxy.args.enable_metrics=true'`）。

#### 使用 [mode.*] 控制组件启停

`yr start --master` 启动主节点模式，`yr start` 启动从节点模式。默认组件如下：

- `master`：`etcd`、`ds_master`、`ds_worker`、`function_master`、`function_proxy`、`function_agent`
- `agent`：`ds_worker`、`function_proxy`、`function_agent`

可通过 `[mode.master]` 或 `[mode.agent]` 打开额外组件。示例（在 master 上启用 `frontend`、`dashboard`、`collector`）：

```toml
[mode.master]
frontend = true
dashboard = true
collector = true
```

运行命令：

`yr -c /path/to/config.toml start --master`

#### 使用 [values.*] 覆盖运行时参数

`values` 主要承载运行时参数（IP、端口、TLS 路径、日志路径等），并用于渲染组件配置模板。

示例：开启函数系统和 etcd 的 TLS：

```toml
[values.fs.tls]
enable = true
base_path = "/yuanrong/myssl/cert/yr"
ca_file = "ca.crt"
cert_file = "module.crt"
key_file = "module.key"

[values.etcd]
auth_type = "TLS"

[values.etcd.auth]
base_path = "/yuanrong/myssl/cert/etcd"
ca_file = "ca.crt"
cert_file = "server.crt"
key_file = "server.key"
client_cert_file = "client.crt"
client_key_file = "client.key"
```

运行命令：

`yr -c /path/to/config.toml start --master`

#### 使用 [component.*] 覆盖组件启动参数

除 `mode.*` 和 `values.*` 外，还可直接覆盖组件级配置（例如 `etcd`、`ds_master`、`function_master`、`function_proxy`）。
这些配置会直接影响组件进程启动参数、环境变量和健康检查行为。

常见可覆盖的子表包括：

- `[COMPONENT.args]`：组件命令行参数
- `[COMPONENT.env]`：组件环境变量
- `[COMPONENT.health_check]`：健康检查配置

**环境变量合并行为**

组件环境变量在实现上按以下方式合并：

- 以当前 shell 环境变量为基础。
- 再应用 `[COMPONENT.env]` 中的变量（通常组件配置优先级更高，会覆盖同名 shell 变量）。
- 若 `LD_LIBRARY_PATH` / `LD_PRELOAD` 同时在 shell 和组件配置中存在，则会拼接为：
  `shell_value:config_value`。

示例（为 `function_master` 追加库路径）：

```sh
export LD_LIBRARY_PATH=/opt/custom/lib
yr -c /path/to/config.toml start --master
```

假设配置文件中包含：

```toml
[function_master.env]
LD_LIBRARY_PATH = "/opt/yr/lib"
```

示例：覆盖部分 `function_master` 和 `function_proxy` 的启动参数：

```toml
[function_master.args]
runtime_recover_enable = true
litebus_thread_num = 25
system_timeout = 2000000

[function_proxy.args]
enable_metrics = false
litebus_thread_num = 25
runtime_recover_enable = true
```

运行命令：

`yr -c /path/to/config.toml start --master`

### 使用 -s 的命令行覆盖（TOML 片段）

启动时可通过 `-s/--set` 临时覆盖配置，避免修改文件。

规则：

- `-s` 可重复指定。
- 每条必须是合法 TOML 赋值语句（`KEY=VALUE`）。

示例：

- 覆盖 master 地址与端口：

    `yr start --master -s 'values.ds_master.ip="10.88.0.9"' -s 'values.ds_master.port="12123"'`

- 覆盖数组表（etcd 地址列表）：

    `yr start --master -s 'values.etcd.address=[{ip="10.88.0.9",peer_port="32380",port="32379"}]'`

`-c/--config` 指定的配置文件路径必须存在；若不指定 `-c`，CLI 会尝试读取默认路径 `/etc/yuanrong/config.toml`，不存在时回退到内置默认值。

### 全局选项

`-c`, `--config` PATH

:   `config.toml` 的路径。若提供，该路径必须存在。

`-v`, `--verbose`

:   启用更详细的日志输出。

`--version`

:   打印版本后退出。

`-h`, `--help`

:   显示帮助。

### 命令

#### start

以 master 或 agent 模式启动 YuanRong 系统。

用法

`yr start` [`--master`] [`--master_address` http(s)://HOST:PORT] [`-s` KEY=VALUE]...

关键选项

`--master`

:   切换为 master 模式（默认是 agent 模式）。

`-s`, `--set` KEY=VALUE

:   从命令行覆盖配置值。可指定多次。

    每个 `KEY=VALUE` 必须是合法的 TOML 赋值语句，并且通常使用 `values.*` 下的点路径。

    示例：

    - `-s 'values.ds_master.ip="10.88.0.9"'`
    - `-s 'values.ds_master.port="12123"'`
    - `-s 'values.etcd.address=[{ip="10.88.0.9",peer_port="32380",port="32379"}]'`

`--master_address` http(s)://HOST:PORT

:   仅 agent 模式可用。启动前从指定 `function_master` 获取服务发现信息，并自动转换为配置覆盖项。若与 `--master` 同时使用，命令会报错退出。

    参数值必须包含协议头，即 `http://` 或 `https://`。当使用 `https://` 时，还需要通过 `--config` 或 `-s` 提供 `values.fs.tls.ca_file`、`values.fs.tls.cert_file`、`values.fs.tls.key_file`（可配合 `values.fs.tls.base_path`）。

说明

- 成功后，会话文件会写入 `/tmp/yr_sessions/latest/session.json`。
- 日志会写入部署目录下（见 [文件](#文件)）。
- 启动 `--master` 时，`yr` 会打印推荐的 `yr start ...` 加入命令，以及最小配置示例。

#### launch

基于 `config.toml` 启动单个组件。通常用于作为容器的 entrypoint。

用法

`yr launch` [`--inherit-env`] [`--env-subst` KEY]... COMPONENT

描述

- 仅启动单个组件，不处理依赖拓扑。
- 常用于容器 entrypoint。
- `--env-subst KEY` 可将配置文件中的 `{KEY}` 替换为环境变量 `KEY` 的值。

#### status

显示系统状态。

用法

`yr status` [`-f` SESSION_FILE]

选项

`-f`, `--file` SESSION_FILE

:   会话 JSON 文件路径（默认：`/tmp/yr_sessions/latest/session.json`）。

描述

`yr status` 读取会话后访问 `function_master` 的 global-scheduler 接口，输出 Ready agent 数量与资源概览。该命令仅支持 `master` 模式会话。

#### stop

停止系统组件。

用法

`yr stop` [`--force`] [`-f` SESSION_FILE]

选项

`--force`

:   强制停止组件（使用 SIGKILL 代替 SIGTERM）。

`-f`, `--file` SESSION_FILE

:   会话 JSON 文件路径。若提供，该路径必须存在。

描述

默认情况下，`yr stop` 会对会话文件中记录的 daemon 进程发送 SIGTERM 尝试停止。

使用 `--force` 时，会对 daemon 发送 SIGKILL，并额外尝试停止会话中记录的所有组件 PID。

#### health

显示当前节点已拉起的所有组件的健康状态与运行情况。

用法

`yr health` [`-f` SESSION_FILE]

选项

`-f`, `--file` SESSION_FILE

:   会话 JSON 文件路径（默认：`/tmp/yr_sessions/latest/session.json`）。

描述

`yr health` 读取会话文件并检查组件 PID 是否在运行。

#### config

与配置相关的命令。

##### config dump

输出合并后的配置。

用法

`yr config dump` [`-s` KEY=VALUE]...

描述

渲染生效配置（默认值 + 配置文件 + 覆盖项），并以 TOML 格式打印。

##### config template

打印内置模板。

用法

`yr config template`

描述

打印内置 `values.toml` 与 `config.toml.jinja` 模板内容。

### 示例

#### 启动 master 集群

```bash
yr start --master
```

#### 启动 agent 并加入 master

执行 `yr start --master` 后，终端会打印推荐的 worker 加入命令，可直接在 worker 节点执行。示例：

`yr start -s 'values.etcd.address=[{ip="10.88.0.4",peer_port="32380",port="32379"}]' -s 'values.ds_master.ip="10.88.0.4"' -s 'values.ds_master.port="12123"' -s 'values.function_master.ip="10.88.0.4"' -s 'values.function_master.global_scheduler_port="22770"'`

也可以在 agent 节点通过自动发现方式加入（等效于自动生成 `-s` 覆盖项）：

`yr start --master_address http://10.88.0.4:22770`

#### 使用默认会话路径查看状态

`yr status`

#### 优雅停止集群

`yr stop`

#### 强制停止（SIGKILL）

`yr stop --force`

#### 查看渲染后的生效配置

`yr config dump`

#### 查看内置模板

`yr config template`

#### 配置安全通信

请先按照[安全通信](./security.md)章节准备组件证书

```toml
[mode.master]
dashboard = true
collector = true

[values]
log_level = "DEBUG"

[values.fs.tls]
enable = true
base_path = "/opt/ssl/cert/yr"
ca_file = "ca.crt"
cert_file = "module.crt"
key_file = "module.key"

[values.ds.curve]
enable = true
base_path = "/opt/ssl/cert/curve"
cache_storage_auth_type = "ZMQ"
cache_storage_auth_enable = true

[values.etcd]
auth_type = "TLS"

[values.etcd.auth]
base_path = "/opt/ssl/cert/etcd"
ca_file = "ca.crt"
cert_file = "server.crt"
key_file = "server.key"
client_cert_file = "client.crt"
client_key_file = "client.key"

[values.dashboard.auth]
enable = true
cert_file = "/opt/ssl/cert/dashboard/server.crt"
key_file = "/opt/ssl/cert/dashboard/server.key"

[values.dashboard.prometheus]
address = "10.88.0.3:9090"

[values.dashboard.prometheus.auth]
enable = true
base_path = "/opt/ssl/cert/prometheus"
ca_file = "ca.crt"
cert_file = "client.crt"
key_file = "client.key"
```

执行以下命令：
`yr -c <your-config-path> start --master`

#### 配置多master

```toml
[mode.master]
frontend = true
function_scheduler = true
dashboard = true
collector = true

[values]
log_level = "DEBUG"

[values.etcd]
enable_multi_master = true

[[values.etcd.address]]
ip = "10.88.0.2"
peer_port = "32380"
port = "32379"

[[values.etcd.address]]
ip = "10.88.0.3"
peer_port = "32380"
port = "32379"
```

执行以下命令：
`yr -c <your-config-path> start --master`

### 文件

`/etc/yuanrong/config.toml`

:   默认配置文件路径（若存在）。

`/tmp/yr_sessions/<timestamp>/`

:   启动时创建的单次运行部署目录。

`/tmp/yr_sessions/latest`

:   指向当前部署目录的符号链接。

`/tmp/yr_sessions/latest/session.json`

:   供 `yr status`、`yr health` 与 `yr stop` 使用的会话文件。

`/tmp/yr_sessions/latest/logs/`

:   当前部署的日志目录。

`/tmp/yr_sessions/latest/logs/yr_daemon.log`

:   daemon 的 stdout/stderr 重定向目标文件。

`/tmp/yr_sessions/latest/logs/<component>_stdout.log`

:   组件标准输出/错误输出日志文件。

`/tmp/yr_sessions/yr_current_master_info`

:   主节点信息文件供yr SDK解析。

### 退出状态

0

:   成功。

1

:   失败（启动失败、状态不健康、会话无效、模式不匹配等）。

## 验证部署状态

部署完成后，可执行 `yr status` 查看集群状态。正常情况下，`ReadyAgentsCount` 与实际可用 worker 数量一致。
若部署失败，可优先检查：

- `/tmp/yr_sessions/latest/logs/yr_daemon.log`
- `/tmp/yr_sessions/latest/logs/<component>_stdout.log`

```bash
Cluster Status:
  ReadyAgentsCount: 2
```

可运行[简单示例](../../../multi_language_function_programming_interface/examples/simple-function-template.md)进一步验证部署结果。

## 删除 openYuanrong 集群

在所有节点执行如下命令即完成集群删除。

```shell
yr stop
```

:::{Note}

`yr stop` 命令并不会删除部署目录下的文件，如果您修改了配置重新部署，请清空该目录下的文件或指定一个新的目录。

:::
