# 部署参数表

本页列出 `yr` Python CLI 的命令行选项以及通过 TOML 配置文件可覆盖的运行时参数。

配置方式参考[部署文档](./production/deploy.md)中的**配置**章节。

## 命令行选项

### 全局选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `-c, --config PATH` | 指定 `config.toml` 路径；若文件不存在，命令报错退出。未指定时尝试读取 `/etc/yuanrong/config.toml`，不存在则使用内置默认值。 | `/etc/yuanrong/config.toml` |
| `-v, --verbose` | 启用 DEBUG 级别日志输出。 | 关闭 |
| `--version` | 打印版本后退出。 | — |
| `-h, --help` | 显示帮助信息后退出。 | — |

### yr start

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `--master` | 以主节点模式启动。未指定时以从节点（agent）模式启动。主节点默认启动：`etcd`、`ds_master`、`ds_worker`、`function_master`、`function_proxy`、`function_agent`；从节点默认启动：`ds_worker`、`function_proxy`、`function_agent`。 | 从节点模式 |
| `-s, --set KEY=VALUE` | 命令行覆盖配置值，可重复指定。`VALUE` 必须是合法 TOML 字面量。 | — |
| `--master_address http(s)://HOST:PORT` | 仅 agent 模式可用。启动前从指定 `function_master` 拉取服务发现信息，自动生成 `-s` 覆盖项。与 `--master` 同时使用时报错退出。使用 `https://` 时还需通过 `--config` 或 `-s` 提供 TLS 证书路径。 | — |

### yr launch

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `COMPONENT` | 要启动的组件名称，如 `etcd`、`ds_master`、`ds_worker`、`function_master`、`function_proxy`、`function_agent`、`frontend`、`dashboard`、`collector` 等。 | 必填 |
| `--inherit-env` | 继承父进程环境变量。 | 关闭 |
| `--env-subst KEY` | 将配置文件中的 `{KEY}` 替换为环境变量 `KEY` 的值。可重复指定或逗号分隔。 | — |

### yr status / yr health

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `-f, --file SESSION_FILE` | 指定会话文件路径。 | `/tmp/yr_sessions/latest/session.json` |

### yr stop

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `--force` | 强制停止：向 daemon 发送 SIGKILL，并对会话中所有组件 PID 执行强制终止。 | 关闭（默认发送 SIGTERM，等待最长 40 秒） |
| `-f, --file SESSION_FILE` | 指定会话文件路径。路径必须存在。 | `/tmp/yr_sessions/latest/session.json` |

### yr config dump

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `-s, --set KEY=VALUE` | 命令行覆盖配置值，可重复指定。 | — |

## TOML 配置字段

以下字段均可在用户 `config.toml` 的 `[values]` 或对应子表下覆盖，也可通过 `-s/--set` 在命令行中临时覆盖。带 `{{ ... }}` 的默认值表示字段由 CLI 在启动时自动填充，无需手动配置。

### [values] 核心字段

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `host_ip` | `{{ ip }}` | 节点 IP，自动获取当前机器 IP。 |
| `deploy_path` | `{{ deploy_path }}` | 部署目录，默认为 `/tmp/yr_sessions/<timestamp>`。 |
| `cpu_num` | `{{ cpu_millicores }}` | 可用 CPU（毫核），默认使用系统全部 CPU。 |
| `memory_num` | `{{ memory_num_mb }}` | 可用内存（MB），默认从系统采集。 |
| `shared_memory_num` | `{{ memory_num_mb // 3 }}` | 数据系统可用共享内存（MB），默认为总内存的 1/3。 |
| `node_id` | `{{ node_id }}` | 节点唯一 ID，格式为 `<hostname>-<pid>`，若手动配置请保证全局唯一。 |
| `log_level` | `INFO` | 全局日志级别，取值：`DEBUG`、`INFO`、`WARN`、`ERROR`。 |

### [values.fs.log] 函数系统日志

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `level` | `{{ values.log_level }}` | 函数系统日志级别。 |
| `path` | `{{ values.deploy_path }}/logs/function_system` | 日志目录。 |
| `rolling_max_size` | `40` | 单文件最大大小（MB）。 |
| `rolling_max_files` | `10` | 滚动保留文件数。 |
| `rolling_retention_days` | `30` | 日志保留天数。 |
| `compress_enable` | `false` | 是否压缩历史日志。 |
| `also_log_to_stderr` | `false` | 是否同时输出到 stderr。 |

### [values.fs.tls] 函数系统 TLS

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `enable` | `false` | 是否开启函数系统 TLS。 |
| `base_path` | `""` | 证书根目录，相对路径的证书文件基于此目录解析。 |
| `ca_file` | `ca.crt` | CA 证书文件名。 |
| `cert_file` | `module.crt` | 证书文件名。 |
| `key_file` | `module.key` | 私钥文件名。 |

### [values.fs.metrics] Metrics 配置

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `metrics_config` | `""` | metrics 配置 JSON 字符串。 |
| `metrics_config_file` | `{{ values.yr_package_path }}/functionsystem/config/metrics/metrics_config.json` | metrics 配置文件路径。 |

### [values.ds.curve] 数据系统安全

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `enable` | `false` | 是否开启 Curve/ZMQ 安全能力。 |
| `base_path` | `""` | Curve 密钥目录。 |
| `cache_storage_auth_type` | `Noauth` | 缓存认证方式，取值：`Noauth`、`ZMQ`、`AK/SK`。 |
| `cache_storage_auth_enable` | `false` | 是否启用缓存认证。 |

### [values.etcd] etcd 配置

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `enable_multi_master` | `false` | 是否开启多 master 模式（多 etcd 节点）。 |
| `auth_type` | `Noauth` | etcd 认证类型，取值：`Noauth`、`TLS`。 |
| `table_prefix` | `""` | etcd key 前缀，用于多套集群共用同一 etcd 时隔离数据，配置值不能包含空格。 |

#### [[values.etcd.address]] etcd 节点地址

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `ip` | `{{ values.host_ip }}` | etcd 节点 IP。 |
| `port` | `{{ 32379\|check_port() }}` | etcd 客户端端口。 |
| `peer_port` | `{{ 32380\|check_port() }}` | etcd peer 端口。 |

#### [values.etcd.auth] etcd TLS 证书

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `base_path` | `""` | 证书根目录。 |
| `ca_file` | `ca.crt` | CA 文件名。 |
| `cert_file` | `server.crt` | 服务端证书文件名。 |
| `key_file` | `server.key` | 服务端私钥文件名。 |
| `client_cert_file` | `client.crt` | 客户端证书文件名。 |
| `client_key_file` | `client.key` | 客户端私钥文件名。 |

### 组件地址与端口

| 配置段 | 字段 | 默认值 | 说明 |
|--------|------|--------|------|
| `[values.ds_master]` | `ip` | `{{ values.host_ip }}` | ds_master 地址。 |
| `[values.ds_master]` | `port` | `{{ 12123\|check_port() }}` | ds_master 端口。 |
| `[values.ds_worker]` | `ip` | `{{ values.host_ip }}` | ds_worker 地址。 |
| `[values.ds_worker]` | `port` | `{{ 31501\|check_port() }}` | ds_worker 端口。 |
| `[values.function_master]` | `ip` | `{{ values.host_ip }}` | function_master 地址。 |
| `[values.function_master]` | `global_scheduler_port` | `{{ 22770\|check_port() }}` | global scheduler 端口，`yr status` 和 `--master_address` 使用此端口。 |
| `[values.function_proxy]` | `ip` | `{{ values.host_ip }}` | function_proxy 地址。 |
| `[values.function_proxy]` | `port` | `{{ 22772\|check_port() }}` | function_proxy HTTP 端口。 |
| `[values.function_proxy]` | `grpc_listen_port` | `{{ 22773\|check_port() }}` | function_proxy gRPC 端口。 |
| `[values.function_agent]` | `ip` | `{{ values.host_ip }}` | function_agent 地址。 |
| `[values.function_agent]` | `port` | `{{ 58866\|check_port() }}` | function_agent 端口。 |

### [values.dashboard] Dashboard

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `ip` | `{{ values.host_ip }}` | Dashboard 地址。 |
| `port` | `{{ 9080\|check_port() }}` | Dashboard HTTP 端口。 |
| `grpc_port` | `{{ 9081\|check_port() }}` | Dashboard gRPC 端口。 |
| `[values.dashboard.auth] enable` | `false` | 是否开启 Dashboard TLS。 |
| `[values.dashboard.auth] cert_file` | `""` | Dashboard TLS 证书路径（绝对路径）。 |
| `[values.dashboard.auth] key_file` | `""` | Dashboard TLS 私钥路径（绝对路径）。 |
| `[values.dashboard.prometheus] address` | `""` | Prometheus 地址，格式 `ip:port`。 |

### [mode.*] 组件启停控制

通过 `[mode.master]` 或 `[mode.agent]` 可在默认组件基础上开启额外组件。

| 字段 | 类型 | 说明 |
|------|------|------|
| `frontend` | `bool` | 是否启动 frontend 组件。 |
| `dashboard` | `bool` | 是否启动 dashboard 组件。 |
| `collector` | `bool` | 是否启动 collector 组件。 |
| `function_scheduler` | `bool` | 是否启动 function_scheduler 组件（多 master 场景）。 |

### [COMPONENT.args] 组件启动参数

通过 `[COMPONENT.args]` 可直接覆盖各组件的命令行参数，支持的组件名称同 `[mode.*]` 中列出的组件名称及 `etcd`、`ds_master`、`ds_worker`、`function_master`、`function_proxy`、`function_agent`。

常见覆盖示例：

```toml
[function_master.args]
runtime_recover_enable = true
litebus_thread_num = 25
system_timeout = 2000000

[function_proxy.args]
enable_metrics = false
litebus_thread_num = 25
runtime_recover_enable = true

[ds_worker.args]
heartbeat_interval_ms = 120000
node_dead_timeout_s = 120
```

### [COMPONENT.env] 组件环境变量

通过 `[COMPONENT.env]` 可为组件进程注入环境变量。合并规则：

- 以当前 shell 环境变量为基础。
- 再应用 `[COMPONENT.env]` 中的变量（同名时组件配置优先覆盖）。
- `LD_LIBRARY_PATH` / `LD_PRELOAD` 若同时存在于 shell 和组件配置，则拼接为 `shell_value:config_value`。
