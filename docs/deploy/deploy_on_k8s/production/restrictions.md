# 规格与限制

包含对软件、网络、硬件环境的要求及限制， 例如不支持在 windows 环境上的部署。

## 1. ds-worker 启动参数规格限制

| 名称 | 含义 | 默认值 | 约束 |
|--|--|--|--|
| unix_domain_socket_dir | 指定 unix domain socket 文件创建的路径 | ~/.datasystem/unix_domain_socket_dir/ | 有效的路径格式&长度不超过 80 个字符 |

## 2. function-agent K8S POD pids 资源规格限制

| 名称 | 含义 | 默认值 | 约束 |
|--|--|--|--|
| 平均每个 CppRuntime 所需 pids 资源 | 基于 function-agent K8S POD pids 资源规格限制 (`/sys/fs/cgroup/pids/kubepods/{burstable,besteffort,}/pod${uuid}/pids.max`) 可计算单个 POD 所能 拉起的最大 CppRuntime 实例个数 | 51 | function-agent K8S POD 内函数系统组件占用 60 pids 的资源，此外对 CppRuntime (不带数据系统) 每个 runtime 需要预留 27 个以上 pids 的资源，(带数据系统) 每个 runtime 需要预留 51 个以上 pids 的资源。修改 K8S POD pids 资源规格，需修改 kubelet 的 `--pod-max-pids` 启动参数；临时修改已有 POD pids cgroups 资源规格可修改`/sys/fs/cgroup/pids/kubepods/{burstable,besteffort,}/pod${uuid}/pids.max` |

## 3. function-master K8S POD 资源规格限制

| 名称 | 含义 | 默认值 | 约束 |
|--|--|--|--|
| .Values.global.resources.functionMaster.limits.memory | 3 Gi | 当集群实例规模大于 10000 ， 建议该参数设置大于 5 Gi ，否则可能出现 function-master 内存不足 oom 。（ k8s 部署当前支持的最大规格不超过 20000 实例） |

## 4. openYuanrong runtime 及 runtime-manager 资源占用情况说明

openYuanrong资源池 pool pod 内，会运行 runtime 及 runtime-manager 。当 runtime 实例被拉起时，这两个组件本身也会占用内存。因此，能提供给用户代码的内存资源，小于 pool 本身的资源配额。具体来说，不同语言 runtime 进程占用 RSS 的情况如下。数据随实际场景和软件版本更迭存在波动，供参考（单位 MB ）。

| 测试内容 | Java             | Python | Cpp |
| --- | -----------------| --- | --- |
| Sleep 60s | 175 | 107 | 95 |
| 100 次 invoke | 176 –> 181 | 107 | 95 |

注：以上数据来自 MultiLanguageRuntime 版的 runtime 。

runtime-manager 占用的内存为 25 MB - 27 MB.

以一个正在运行的空载 cpp runtime 实例为例， runtime 和 runtime-mananger 占用的内存 RSS 之和约为 110 MB. 对于一个配额为 128 MB 的资源池 pod ，可供用户使用的内存不超过 18 MB 。容易导致 runtime 和 runtime-manager 因为 OOM 被杀。

用户需根据其代码运行占用的内存，和其代码的动态链接库 .so 文件加载占用的内存，自行评估合理的资源池内存大小配置。

## 5. etcd 部署规格说明

### etcd 配置说明

| 参数名称 | 含义                                                             | 默认值        | 参数影响                                                   |
|------|----------------------------------------------------------------|------------|--------------------------------------------------------|
| `--snapshot-count`     | 指收到多少个写请求后就触发生成一次快照，并对 Raft 日志条目进行压缩。                          | 100000     | 执行快照前，数据需要保存在内存中，如果值过大，会消耗较多的内存，如果值过小，可能频繁触发快照，影响写入性能。 |
| `--quota-backend-bytes`     | 后端存储的最大存储字节限制，单位 B                                             | 8589934592 | 当写入数据超过这个限制时，etcd 将停止接受新的写入操作                          |
| `--auto-compaction-mode`     | 自动压缩策略, 配置 periodic 表示定期压缩，配置 revision 表示按照修订版本数量压缩            | revision   | 如果 etcd 非定时写入数据，建议配置 revision 策略                       |
| `--auto-compaction-retention`     | 指定自动压缩的保留值，对于 periodic 策略，表示保留的时间窗口，对于 revision 策略，表示保留的修订版本数量 | 1000       | 自动压缩会清理旧的修订版本，减少 etcd 的存储空间占用，但可能会对性能产生一定影响    |

### 性能基线配置

| Cpu | Memory | quota-backend-bytes         | snapshot-count | auto-compaction-mode | auto-compaction-retention | 集群节点数 | 最大并发实例创建数 | 注册函数数量 | 最大运行中实例数 |
|-----|--------|----------------|----------------|---------------|---------------------------|-------|-----------|--------|----------|
| 4C  | 8G     | 8589934592 (8G) | 20000          | revision      | 5000                      | <100  | <1000     | <1000  | <2000    |
| 8C  | 16G    | 8589934592 (8G) | 50000          | revision      | 10000                     | <500  | <2000     | <5000  | <5000    |
| 16C | 32G    | 8589934592 (8G) | 100000         | revision      | 25000                     | <1000 | <5000     | <10000 | <10000   |

Cpu 和 Memory 参考配置(修改 etcd 部署 StatefulSet):

```yaml
resources:
  limits:
    cpu: "4"
    memory: 8Gi
  requests:
    cpu: "4"
    memory: 8Gi
```

压缩参数和 DB Size 参考配置(修改 etcd 部署 StatefulSet)：

```yaml
# 新增如下环境变量
# 配置 auto-compaction-mode 压缩模式
- name: ETCD_AUTO_COMPACTION_MODE
  value: "revision"
#  配置 auto-compaction-retention 配置压缩参数
- name: ETCD_AUTO_COMPACTION_RETENTION
  value: "10000"
# 配置 quota-backend-bytes DB Size(配置单位为 bytes)
- name: ETCD_QUOTA_BACKEND_BYTES
  value: "8589934592"
# 配置  snapshot-count
- name: ETCD_SNAPSHOT_COUNT
  value: "20000"
```

### 特别说明

1. 最大并发实例创建数增大，请增大 Cpu,Memory, 自动压缩参数。
2. 注册函数数量/最大运行中实例数增大，请增大 Memory,DB Size。
3. 集群节点数增大，请增大 Cpu,Memory。
4. 当内存不足时，请减少保存快照数量，防止 etcd 发生 OOM。
5. etcd 性能依赖磁盘，请尽量使用高性能磁盘，参照官网 [etcd](https://etcd.io/docs/v3.5/op-guide/hardware/)。
