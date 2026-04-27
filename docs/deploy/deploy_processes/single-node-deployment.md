# 入门

本节将演示使用默认配置参数在一台或者多台 Linux 主机上部署 openYuanrong ，建议用于学习和开发，生产部署请参考[用户指南](./production/index.md)。

## 部署 openYuanrong

首先参考[安装指南](../installation.md)在所有部署主机上安装 openYuanrong 命令行工具 yr，我们将使用它部署 openYuanrong。

任选一台主机，使用如下命令部署[主节点](glossary-master-node)。

```bash
yr start --master
```

部署成功后，终端会打印 worker 节点加入集群的推荐命令，格式如下：

```text
To join an existing cluster, execute the following commands in your shell on worker nodes:

yr start -s 'values.etcd.address=[{ip="x.x.x.x",peer_port="xxxxx",port="xxxxx"}]' -s 'values.ds_master.ip="x.x.x.x"' -s 'values.ds_master.port="xxxxx"' -s 'values.function_master.ip="x.x.x.x"' -s 'values.function_master.global_scheduler_port="xxxxx"'

OR

mkdir -p /etc/yuanrong/ && cat << EOF > /etc/yuanrong/config.toml && yr start
[values.etcd]
...
EOF

OR

yr start --master_address http://x.x.x.x:xxxxx
```

此时，openYuanrong 服务已经可以使用。需要多节点集群部署时，在其余主机上直接执行主节点打印的推荐命令，以部署[从节点](glossary-agent-node)。以下为两种等价方式：

**方式一：直接使用打印的 `-s` 覆盖命令**

```bash
# 将 x.x.x.x 和端口替换为主节点打印的实际值
yr start \
  -s 'values.etcd.address=[{ip="x.x.x.x",peer_port="xxxxx",port="xxxxx"}]' \
  -s 'values.ds_master.ip="x.x.x.x"' \
  -s 'values.ds_master.port="xxxxx"' \
  -s 'values.function_master.ip="x.x.x.x"' \
  -s 'values.function_master.global_scheduler_port="xxxxx"'
```

**方式二：使用自动发现（`--master_address`）**

```bash
# 将地址替换为主节点的 function_master global_scheduler 地址
yr start --master_address http://x.x.x.x:xxxxx
```

在主节点上执行 `yr status` 命令可查看集群状态。正常情况下，`ReadyAgentsCount` 与实际部署节点数量一致。

```bash
yr status
```

```text
Cluster Status:
  ReadyAgentsCount: 2
```

可运行[简单示例](../../multi_language_function_programming_interface/examples/simple-function-template.md)进一步验证部署结果。

## 删除 openYuanrong 集群

使用命令行工具 yr 在**所有部署节点**上执行如下命令：

```bash
yr stop
```

:::{note}
`yr stop` 命令会向 daemon 进程发送 SIGTERM，等待其优雅退出（最长 40 秒）。若超时，可使用 `yr stop --force` 强制终止。
:::
