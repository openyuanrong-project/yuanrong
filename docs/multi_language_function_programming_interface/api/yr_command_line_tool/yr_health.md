# health

检查当前节点组件健康状态。

## 用法

```shell
yr health [OPTIONS]
```

## 参数

* `-f, --file SESSION_FILE`：指定会话文件路径（默认 `/tmp/yr_sessions/latest/session.json`）。

## 说明

* `health` 会读取会话文件中的组件 PID，并检查进程是否仍在运行。
* 若会话不存在或组件未运行，将返回非零退出码。

## 样例

```shell
yr health
```

```shell
yr health -f /tmp/yr_sessions/20260209_020054/session.json
```
