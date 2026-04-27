# start

启动 openYuanrong 集群，支持 `master` 和 `agent` 两种模式。

## 用法

```shell
yr start [OPTIONS]
```

## 参数

* `--master`：以主节点模式启动（默认不带该参数时为 agent 模式）。
* `-s, --set KEY=VALUE`：命令行覆盖配置，可重复指定；`VALUE` 必须是合法 TOML 字面量。
* `--master_address http(s)://host:port`：仅 agent 模式可用。启动前从指定 `function_master` 拉取服务发现信息并自动转换为 `-s` 覆盖项。

:::{Note}

`-s/--set` 常见示例：

* 字符串：`-s 'values.fs.log.level="DEBUG"'`
* 数值：`-s 'values.ds_master.port=12123'`
* 数组表：`-s 'values.etcd.address=[{ip="10.88.0.9",peer_port="32380",port="32379"}]'`

`--master_address` 使用限制：

* 仅 `yr start`（agent 模式）支持。
* 与 `--master` 同时使用会报错并退出。
* 服务发现失败会直接退出。
* 参数值必须包含协议头，即 `http://` 或 `https://`。
* 当使用 `https://` 时，需通过 `--config` 或 `-s` 提供 `values.fs.tls.ca_file`、`values.fs.tls.cert_file`、`values.fs.tls.key_file`（可配合 `values.fs.tls.base_path`）。

:::

## 配置合并规则

`yr start` 启动时，最终配置按以下优先级合并（高到低）：

1. 命令行覆盖：`-s/--set`
2. 用户配置文件：默认 `/etc/yuanrong/config.toml`（可通过 `-c` 指定）
3. 内置默认值：`api/python/yr/cli/values.toml` + `api/python/yr/cli/config.toml.jinja`

## 启动行为

执行 `yr start` 后，CLI 会：

1. 创建部署目录 `/tmp/yr_sessions/<timestamp>/`
2. 更新符号链接 `/tmp/yr_sessions/latest -> /tmp/yr_sessions/<timestamp>/`
3. 启动模式对应组件并执行健康检查
4. 写入会话文件 `/tmp/yr_sessions/latest/session.json`

其中：

* `master` 默认启动：`etcd`、`ds_master`、`ds_worker`、`function_master`、`function_proxy`、`function_agent`
* `agent` 默认启动：`ds_worker`、`function_proxy`、`function_agent`

## 样例

启动主节点：

```shell
yr start --master
```

启动成功后，终端会打印用于 worker 节点加入集群的推荐命令（包含 `-s 'values.*=...'` 覆盖项）。

在 worker 节点可直接执行该命令加入集群，例如：

```shell
yr start -s 'values.etcd.address=[{ip="10.88.0.4",peer_port="32380",port="32379"}]' \
  -s 'values.ds_master.ip="10.88.0.4"' \
  -s 'values.ds_master.port="12123"' \
  -s 'values.function_master.ip="10.88.0.4"' \
  -s 'values.function_master.global_scheduler_port="22770"'
```

或使用自动发现模式（agent）：

```shell
yr start --master_address http://10.88.0.4:22770
```
