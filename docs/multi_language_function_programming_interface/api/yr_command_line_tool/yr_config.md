# config

查看或导出 CLI 生效配置。

## 子命令

* `yr config dump`：输出合并后的最终配置（TOML）。
* `yr config template`：输出内置 `values.toml` 和 `config.toml.jinja` 模板。

## 用法

```shell
yr config dump [OPTIONS]
yr config template
```

## 参数

`yr config dump` 支持：

* `-s, --set KEY=VALUE`：命令行覆盖配置，可重复指定；`VALUE` 需为合法 TOML 字面量。

## 说明

`config dump` 的配置合并优先级（高到低）：

1. `-s/--set` 命令行覆盖
2. `-c/--config` 指定或默认路径的用户配置文件
3. 内置默认模板

## 基于内置 config.toml.jinja 自定义配置

`yr` 的最终生效配置是由以下两份内置文件渲染合并得到：

- `api/python/yr/cli/values.toml`
- `api/python/yr/cli/config.toml.jinja`

如果想调整某个组件的启动参数，推荐用自定义 `config.toml` 覆盖对应字段，而不是改内置模板文件。

推荐步骤：

1. 执行 `yr config template` 查看内置模板结构。
2. 新建 `config.toml`，只写你要覆盖的字段。
3. 先执行 `yr -c /path/to/config.toml config dump` 校验渲染结果。
4. 执行 `yr -c /path/to/config.toml start --master`（或 `start`）启动。

示例（覆盖组件启动参数）：

```toml
[mode.master]
frontend = true
dashboard = true

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

验证与启动：

```shell
yr -c /path/to/config.toml config dump
yr -c /path/to/config.toml start --master
```

如果只想临时改一个参数，也可以不改文件直接使用 `-s/--set`，例如：

```shell
yr start --master -s 'function_proxy.args.enable_metrics=true'
```

## values.toml 字段说明

内置默认值文件路径：`api/python/yr/cli/values.toml`。

:::{Note}

`values.toml` 中存在 `{{ ... }}` 模板变量，**这些变量会在启动时由 CLI 自动填充，无需用户手动设置**。CLI 会根据运行环境自动获取和计算这些值（如 IP、时间戳、CPU、内存、端口等）。

端口字段常见写法 `{{ 32379|check_port() }}` 表示优先使用给定端口；若被占用则自动选择可用端口。

如需覆盖自动计算的值，可通过 `-s/--set` 参数指定，例如：`-s 'values.host_ip="10.0.0.1"'`。

:::

### [values] 核心字段

:::{tip}
下表中带 `{{ ... }}` 的默认值表示该字段会在运行时自动填充，无需手动配置。
:::

| 字段 | 默认值 | 说明 |
| ---- | ---- | ---- |
| `host_ip` | `{{ ip }}` | 节点 IP。 |
| `deploy_path` | `{{ deploy_path }}` | 本次启动部署目录（通常在 `/tmp/yr_sessions/<timestamp>`）。 |
| `yr_package_path` | `{{ yr_package_path }}` | `openyuanrong` Python 包路径。 |
| `cpu_num` | `{{ cpu_millicores }}` | 可用 CPU（毫核）。 |
| `memory_num` | `{{ memory_num_mb }}` | 可用内存（MB）。 |
| `shared_memory_num` | `{{ memory_num_mb // 3 }}` | 共享内存（MB）。 |
| `node_id` | `{{ node_id }}` | 节点唯一 ID（默认基于 hostname + pid）。 |
| `log_level` | `INFO` | 全局日志级别。 |
| `pid` | `{{ pid }}` | 当前 `yr` 进程 PID。 |
| `ld_library_path` | `{{ ld_library_path }}` | 启动时继承到模板渲染上下文中的 `LD_LIBRARY_PATH`。 |
| `python_path` | `{{ python_path }}` | 启动时继承到模板渲染上下文中的 `PYTHONPATH`。 |

### [values.meta_store]

| 字段 | 默认值 | 说明 |
| ---- | ---- | ---- |
| `enable` | `false` | 是否启用 meta store。 |
| `address` | `""` | meta store 地址。 |
| `mode` | `local` | meta store 模式。 |

### [values.fs] 与函数系统相关字段

| 字段 | 默认值 | 说明 |
| ---- | ---- | ---- |
| `schedule_plugins` | `["Label","ResourceSelector","Default","Heterogeneous","NUMA"]` | 调度插件链。 |

#### [values.fs.log]

| 字段 | 默认值 | 说明 |
| ---- | ---- | ---- |
| `also_log_to_stderr` | `false` | 是否同时输出到 stderr。 |
| `async_log_buf_secs` | `30` | 日志缓冲秒数。 |
| `async_log_max_queue_size` | `51200` | 异步日志队列上限。 |
| `async_log_thread_count` | `1` | 异步日志线程数。 |
| `level` | `{{ values.log_level }}` | 函数系统日志级别。 |
| `path` | `{{ values.deploy_path }}/logs/function_system` | 函数系统日志目录。 |
| `rolling_max_files` | `10` | 日志滚动保留文件数。 |
| `rolling_max_size` | `40` | 单文件最大大小（MB）。 |
| `rolling_retention_days` | `30` | 日志保留天数。 |
| `compress_enable` | `false` | 是否压缩日志。 |

#### [values.fs.tls]

| 字段 | 默认值 | 说明 |
| ---- | ---- | ---- |
| `enable` | `false` | 是否开启函数系统 TLS。 |
| `base_path` | `""` | 证书根目录。 |
| `ca_file` | `ca.crt` | CA 文件名。 |
| `cert_file` | `module.crt` | 证书文件名。 |
| `key_file` | `module.key` | 私钥文件名。 |

#### [values.fs.metrics]

| 字段 | 默认值 | 说明 |
| ---- | ---- | ---- |
| `metrics_config` | `""` | metrics 配置字符串。 |
| `metrics_config_file` | `{{ values.yr_package_path }}/functionsystem/config/metrics/metrics_config.json` | metrics 配置文件路径。 |

### [values.ds.curve]

| 字段 | 默认值 | 说明 |
| ---- | ---- | ---- |
| `enable` | `false` | 是否开启 Curve/ZMQ 安全能力。 |
| `base_path` | `""` | Curve 密钥目录。 |
| `cache_storage_auth_type` | `Noauth` | 缓存认证方式（`Noauth`/`ZMQ`/`AK/SK`）。 |
| `cache_storage_auth_enable` | `false` | 是否启用缓存认证。 |

### [values.etcd] 与 etcd 安全配置

| 字段 | 默认值 | 说明 |
| ---- | ---- | ---- |
| `enable_multi_master` | `false` | 是否开启多 master（多 etcd 节点）模式。 |
| `auth_type` | `Noauth` | etcd 认证类型（如 `Noauth`、`TLS`）。 |
| `table_prefix` | `""` | etcd 表前缀。 |

#### [values.etcd.auth]

| 字段 | 默认值 | 说明 |
| ---- | ---- | ---- |
| `base_path` | `""` | etcd 证书目录。 |
| `ca_file` | `ca.crt` | etcd CA 文件名。 |
| `cert_file` | `server.crt` | etcd 服务端证书文件名。 |
| `key_file` | `server.key` | etcd 服务端私钥文件名。 |
| `client_cert_file` | `client.crt` | etcd 客户端证书文件名。 |
| `client_key_file` | `client.key` | etcd 客户端私钥文件名。 |

#### [[values.etcd.address]]

| 字段 | 默认值 | 说明 |
| ---- | ---- | ---- |
| `ip` | `{{ values.host_ip }}` | etcd 节点 IP。 |
| `peer_port` | `{{ 32380\|check_port() }}` | etcd peer 端口。 |
| `port` | `{{ 32379\|check_port() }}` | etcd client 端口。 |

### 组件地址与端口字段

| 配置段 | 字段 | 默认值 | 说明 |
| ---- | ---- | ---- | ---- |
| `[values.ds_master]` | `ip` | `{{ values.host_ip }}` | ds_master 地址。 |
| `[values.ds_master]` | `port` | `{{ 12123\|check_port() }}` | ds_master 端口。 |
| `[values.ds_worker]` | `ip` | `{{ values.host_ip }}` | ds_worker 地址。 |
| `[values.ds_worker]` | `port` | `{{ 31501\|check_port() }}` | ds_worker 端口。 |
| `[values.function_master]` | `ip` | `{{ values.host_ip }}` | function_master 地址。 |
| `[values.function_master]` | `global_scheduler_port` | `{{ 22770\|check_port() }}` | global scheduler 端口。 |
| `[values.function_proxy]` | `ip` | `{{ values.host_ip }}` | function_proxy 地址。 |
| `[values.function_proxy]` | `port` | `{{ 22772\|check_port() }}` | function_proxy HTTP 端口。 |
| `[values.function_proxy]` | `grpc_listen_port` | `{{ 22773\|check_port() }}` | function_proxy gRPC 端口。 |
| `[values.function_agent]` | `ip` | `{{ values.host_ip }}` | function_agent 地址。 |
| `[values.function_agent]` | `port` | `{{ 58866\|check_port() }}` | function_agent 端口。 |

### Dashboard 相关字段

| 配置段 | 字段 | 默认值 | 说明 |
| ---- | ---- | ---- | ---- |
| `[values.dashboard]` | `ip` | `{{ values.host_ip }}` | dashboard 地址。 |
| `[values.dashboard]` | `port` | `{{ 9080\|check_port() }}` | dashboard HTTP 端口。 |
| `[values.dashboard]` | `grpc_port` | `{{ 9081\|check_port() }}` | dashboard gRPC 端口。 |
| `[values.dashboard.auth]` | `enable` | `false` | dashboard TLS 开关。 |
| `[values.dashboard.auth]` | `cert_file` | `""` | dashboard 证书路径。 |
| `[values.dashboard.auth]` | `key_file` | `""` | dashboard 私钥路径。 |
| `[values.dashboard.prometheus]` | `address` | `""` | prometheus 地址。 |
| `[values.dashboard.prometheus.auth]` | `enable` | `false` | prometheus TLS 开关。 |
| `[values.dashboard.prometheus.auth]` | `base_path` | `""` | prometheus 证书目录。 |
| `[values.dashboard.prometheus.auth]` | `ca_file` | `""` | prometheus CA 文件。 |
| `[values.dashboard.prometheus.auth]` | `cert_file` | `""` | prometheus 客户端证书。 |
| `[values.dashboard.prometheus.auth]` | `key_file` | `""` | prometheus 客户端私钥。 |

## 样例

```shell
yr config dump
```

```shell
yr config dump -s 'values.ds_master.port=12123' -s 'values.fs.log.level="DEBUG"'
```

```shell
yr config template
```
