# stop

停止当前会话中的 openYuanrong 进程。

## 用法

```shell
yr stop [OPTIONS]
```

## 参数

* `--force`：强制停止，使用 `SIGKILL`。
* `-f, --file SESSION_FILE`：指定会话文件路径（默认 `/tmp/yr_sessions/latest/session.json`）。

## 说明

* 默认行为：向 daemon 发送 `SIGTERM`，由 daemon 执行组件优雅退出流程。
* `--force` 行为：先 `SIGKILL` daemon，再基于会话文件对组件 PID 执行强制停止。
* 如果会话文件不存在或无效，命令会返回失败。

## 样例

```shell
yr stop
```

```shell
yr stop --force
```

```shell
yr stop -f /tmp/yr_sessions/20260209_020054/session.json
```
