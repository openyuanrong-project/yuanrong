# 部署 openYuanrong

本节将介绍常见场景下主机部署 openYuanrong 时的配置，配置参数详细说明请参考[部署参数表](../parameters.md)。

## 使用命令行工具 yr 部署 openYuanrong

首先参考[安装指南](../../installation.md)在所有部署主机上安装 openYuanrong 命令行工具 yr，我们将使用它部署。

命令行工具 `yr` 部署 openYuanrong 时提供了如下配置项的默认值。需要时，您可以通过直接设定配置项覆盖该默认值。

|  配置项  | 说明    |  默认值 |
| ------- | ------- | ------- |
| `--ip_address` | 节点 IP | 主机 IP |
| `--deploy_path` | 部署路径 | 创建 `/tmp/yr_sessions` 目录，并以时间戳区分多次启动，例如：/tmp/yr_sessions/20250421200740 |
| `--cpu_num` | 节点可用 CPU 总量 | 从 `/proc/cpuinfo` 读取 |
| `--memory_num` | 节点可用内存总量 | 从 `/proc/meminfo` 读取  |
| `--shared_memory_num` | 节点可用于存储“数据对象”的内存量 | `--memory_num` 配置项值的三分之一 |
| `--master_info_output` | 主节点启动信息保存路径 | 使用 `--deploy_path` 配置的目录，例如：/tmp/yr_sessions/20250421200740/master.info |
| `--services_path` | 内置函数元数据描述文件 `services.yaml` 路径及文件名 | openYuanrong 安装路径（可通过 `yr version` 命令查看）下的 deploy/process 目录 |
| `--enable_inherit_env` | 创建函数进程时继承 runtime-manager 的环境变量 | true |

### 使用默认配置部署

部署主节点：

```bash
yr start --master
```

部署成功会打印如下主节点信息。

```bash
Cluster master info:
    local_ip:x.x.x.x,master_ip:x.x.x.x,etcd_ip:x.x.x.x,etcd_port:19271,global_scheduler_port:12117,ds_master_port:17611,cluster_deployer_port:22775,etcd_peer_port:19518,bus-proxy:22487,bus:32158,ds-worker:38877,
```

部署从节点：

```bash
# 使用前一步骤打印的主节点信息替换引号中的内容。
yr start --master_info "local_ip:x.x.x.x,master_ip:x.x.x.x,etcd_ip:x.x.x.x,etcd_port:19271,global_scheduler_port:12117,ds_master_port:17611,cluster_deployer_port:22775,etcd_peer_port:19518,bus-proxy:22487,bus:32158,ds-worker:38877,"
```

### 部署时配置资源

您可在 openYuanrong 主从节点上分别配置 CPU 和内存总量，内存中部分用于函数堆栈，部分用于存储数据对象。

- `--cpu_num`：CPU 总量配置参数。主节点 openYuanrong 组件默认占用 1 毫核（单位：1/1000 核），如果您希望主节点不运行分布式任务，只用于管理和调度，可配置 `--cpu_num=1`。
- `--memory_num`：内存总量配置参数。
- `--shared_memory_num`：用于存储数据对象的内存量。如果应用场景中有较多的数据对象存储，可适当调大。

openYuanrong 也支持异构计算资源，通过在主从节点分别配置 `--npu_collection_mode` 及 `--gpu_collection_enable` 参数，openYuanrong 会自动采集节点上的 NPU 及 GPU 资源。此外您也可以在主从节点上分别配置 `--custom_resources` 定义自定义资源。

```bash
# 主节点不用于运行分布式任务
yr start --master -c 1
```

部署成功会打印如下主节点信息。

```bash
Cluster master info:
    local_ip:x.x.x.x,master_ip:x.x.x.x,etcd_ip:x.x.x.x,etcd_port:18107,global_scheduler_port:13611,ds_master_port:11647,cluster_deployer_port:22775,etcd_peer_port:17406,bus-proxy:28169,bus:38575,ds-worker:22903,
```

部署从节点：

```bash
# 从节点支持采集 GPU 资源
# 使用前一步骤打印的主节点信息替换引号中的内容。
yr start --gpu_collection_enable true --master_info "local_ip:x.x.x.x,master_ip:x.x.x.x,etcd_ip:x.x.x.x,etcd_port:18107,global_scheduler_port:13611,ds_master_port:11647,cluster_deployer_port:22775,etcd_peer_port:17406,bus-proxy:28169,bus:38575,ds-worker:22903,"
```

### 部署多个主节点

主节点默认只部署一个，在高可靠场景下，也可按一主多备方式部署多个。这里以常见的一主两备部署方式，使用内置 etcd 为例。

在三台主机上分别执行如下命令，部署主节点：

```bash
# master_ip 替换为每台主机的 ip，并指定每台主机使用的 etcd 端口，请确保端口不冲突
# 例如：yr start --master --etcd_addr_list 192.168.0.1:23279:23280,192.168.0.2:23279:23280,192.168.0.0.3:23279:23280
yr start --master --etcd_addr_list={master-1_ip:etcd-1_port:etcd-1_peer_port,master-2_ip:etcd-2_port:etcd-2_peer_port,master-3_ip:etcd-3_port:etcd-3_peer_port}
```

部署备节点：

```bash
# etcd_addr_list 的配置和主节点保持一致
# 例如：yr start --etcd_addr_list 192.168.0.1:23279:23280,192.168.0.2:23279:23280,192.168.0.0.3:23279:23280
yr start --etcd_addr_list={master-1_ip:etcd-1_port:etcd-1_peer_port,master-2_ip:etcd-2_port:etcd-2_peer_port,master-3_ip:etcd-3_port:etcd-3_peer_port}
```

:::{Note}

使用内置 etcd 时，对于一个有 N 个主节点的 openYuanrong 集群，最多能容忍 (N-1)/2 个主节点故障，当故障主节点数量超过该数值时，集群无法正常提供服务。

:::

### 部署时配置安全通信

openYuanrong 支持内部组件间及内部组件同三方组件 ETCD 间的加密通信。当前只支持配置明文证书密钥，因此存在证书密钥泄露风险。如果您有高安全的密钥管理需求，可基于 openYuanrong 开源代码自行实现密钥解密算法，同时配置加密的证书密钥，其他密钥配置也可参考该方案。

openYuanrong 默认未开启安全通信选项，如需开启请参考[安全通信](./security.md)章节生成相关证书密钥。

## 验证部署状态

部署完成后，在 openYuanrong 集群任一主机上执行 `yr status` 命令可查看集群状态。正常情况下，`current running agents` 的数量和实际部署的节点数量一致。如果部署失败，可在部署路径 `/tmp/yr_sessions/latest` 下查看完整日志文件 `deploy_std.log` 分析原因。

```bash
YuanRong cluster addresses:
              functionsystem: x.x.x.x:30839
                  datasystem: x.x.x.x:39133

YuanRong cluster status:
  current running agents: 2
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
