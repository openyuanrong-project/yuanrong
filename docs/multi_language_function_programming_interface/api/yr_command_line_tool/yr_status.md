# status

查询集群状态与资源视图。

## 用法

```shell
yr status [OPTIONS]
```

## 参数

* `-f, --file SESSION_FILE`：指定会话文件路径（默认 `/tmp/yr_sessions/latest/session.json`）。

## 说明

* `status` 依赖会话文件中的集群连接信息。
* 只有 `master` 模式会输出完整集群状态；若会话来自 `agent` 模式会返回错误。
* 当 function master 开启 TLS 时，`status` 会自动按会话配置进行 HTTPS 查询。

## 样例

```shell
yr status
```

```shell
yr status -f /tmp/yr_sessions/20260209_020054/session.json
```
