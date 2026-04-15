# CLI - the yr command

`yr` 是 openYuanrong Python 部署与运维 CLI，用于主机模式下的组件启动、停止、状态检查与配置渲染。

## 生命周期命令

| 命令 | 说明 |
| ---- | ---- |
| [`start`](./yr_start.md) | 启动集群（`master` 或 `agent` 模式）。 |
| [`launch`](./yr_launch.md) | 仅启动单个组件，通常用于容器 entrypoint 场景。 |
| [`health`](./yr_health.md) | 读取会话文件，查看当前节点组件运行健康状态。 |
| [`status`](./yr_status.md) | 查看集群状态（仅 `master` 模式可查询集群资源视图）。 |
| [`stop`](./yr_stop.md) | 停止会话中记录的 daemon/组件进程。 |

## 配置命令

| 命令 | 说明 |
| ---- | ---- |
| [`config`](./yr_config.md) | 输出合并后的配置或内置配置模板。 |

## 全局选项

| 选项 | 说明 |
| ---- | ---- |
| `-c, --config PATH` | 指定 `config.toml` 路径；未指定时默认读取 `/etc/yuanrong/config.toml`。 |
| `-v, --verbose` | 开启 DEBUG 日志。 |
| `--version` | 输出 `yr` 版本并退出。 |
| `-h, --help` | 查看帮助信息。 |

```{eval-rst}
.. toctree::
  :hidden:

  yr_start
  yr_launch
  yr_health
  yr_status
  yr_stop
  yr_config
  yr_version
```



## 获取安装路径

```shell
python3 -c "import yr;print(yr.__path__[0])"
# /usr/local/lib/python3.9/site-packages/yr
```