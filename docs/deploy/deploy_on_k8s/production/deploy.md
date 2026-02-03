# 部署 openYuanrong

本节介绍部署 openYuanrong 的流程。

(k8s-deploy-download-release-package)=

## 添加 openYuanrong 的 helm 仓库

openYuanrong K8s 安装包依赖于 helm，需要添加 openYuanrong 提供的 helm 仓库地址。

```bash
helm repo add yr http://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/charts/
helm repo update
```

## 配置 openYuanrong 的镜像仓库

openYuanrong 版本镜像存放在私有镜像仓库上，在 K8s 各节点终端执行 `vim /etc/docker/daemon.json` 命令，新增如下内容为 docker 添加 openYuanrong 白名单镜像仓库地址。

```json
{
    "insecure-registries": [
        "swr.cn-southwest-2.myhuaweicloud.com"
    ]
}
```

执行如下命令使配置生效：

```shell
systemctl daemon-reload
systemctl restart docker
```

## 部署

在任意 K8s 节点上使用 helm 命令部署 openYuanrong。

查找 openyuanrong 版本。

```bash  
helm search repo yr/openyuanrong
```

结果输出：

```bash
NAME               CHART VERSION    APP VERSION    DESCRIPTION
yr/openyuanrong    0.2.6            1.16.0         A Helm chart for Kubernetes
```

指定 openyuanrong 版本安装。

```bash
helm pull --untar yr/openyuanrong --version 0.2.6
cd openyuanrong
```

安装最新版本。 

```bash
helm repo update
helm pull --untar yr/openyuanrong 
cd openyuanrong 
```

:::{caution}

如果集群环境中 etcd 非全新安装，请先清理残留数据避免部署失败。

:::

- 只部署 openYuanrong，不创建 Pod 资源池，需要配置 ETCD 的地址和端口信息、MinIO 的 AccessKey 和 SecreteKey，适合自行通过[资源池管理 API](../api/index.md) 创建资源池的场景。其中，ETCD为集群时，多个节点的地址以逗号分隔，例如 `192.168.10.1X:2379,192.168.10.2X:2379,192.168.10.3X:2379`。

  ```shell
  helm install openyuanrong --set global.etcd.etcdAddress=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.systemUpgradeConfig.systemUpgradeWatchAddress=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.etcdManagement.detcd=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.etcdManagement.metcd=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.obsManagement.s3AccessKey=${Your Minio AccessKey} \
    --set global.obsManagement.s3SecretKey=${Your Minio SecretKey}  .
  ```

- 部署 openYuanrong 并创建默认规格为（3 核 cpu，6GB 内存）的 Pod 资源池，建议用于开发有状态函数及无状态函数。

  ```shell
  helm install openyuanrong --set global.etcd.etcdAddress=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.systemUpgradeConfig.systemUpgradeWatchAddress=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.etcdManagement.detcd=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.etcdManagement.metcd=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.obsManagement.s3AccessKey=${Your Minio AccessKey} \
    --set global.obsManagement.s3SecretKey=${Your Minio SecretKey} \
    --set global.pool.poolSize=1 \
    --set global.pool.requestCpu=3000m \
    --set global.pool.requestMemory=6144Mi \
    --set global.pool.limitCpu=3000m \
    --set global.pool.limitMemory=6144Mi .
  ```

- 部署 openYuanrong 并创建默认规格为 600 毫核 cpu，512MB 内存的 Pod 资源池，建议用于开发函数服务。

  ```shell
  helm install openyuanrong --set global.etcd.etcdAddress=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.systemUpgradeConfig.systemUpgradeWatchAddress=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.etcdManagement.detcd=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.etcdManagement.metcd=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.obsManagement.s3AccessKey=${Your Minio AccessKey} \
    --set global.obsManagement.s3SecretKey=${Your Minio SecretKey} \
    --set global.pool.poolSize=1 \
    --set global.pool.requestCpu=600m \
    --set global.pool.requestMemory=512Mi \
    --set global.pool.limitCpu=600m \
    --set global.pool.limitMemory=512Mi .
  ```

- 其他：修改单个组件的镜像版本示例。

  ```shell
  helm install openyuanrong --set global.etcd.etcdAddress=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.systemUpgradeConfig.systemUpgradeWatchAddress=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.etcdManagement.detcd=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.etcdManagement.metcd=${Your ETCD ADDRESS}:${Your ETCD Port} \
    --set global.obsManagement.s3AccessKey=${Your Minio AccessKey} \
    --set global.obsManagement.s3SecretKey=${Your Minio SecretKey} \
    --set global.images.functionAgent=function-agent:${version} .
  ```

检查部署结果：以下所有 Pod 全部为 Running 状态即部署成功。

```shell
kubectl get pods -owide -w

NAME                                                              READY   STATUS    RESTARTS   AGE     IP              NODE           NOMINATED NODE   READINESS GATE
ds-core-etcd-0                                                    1/1     Running   0          19d     10.x.x.122     dggphis18024   <none>           <none>
ds-worker-llcq8                                                   1/1     Running   0          2m56s   10.x.x.56      dggphis18023   <none>           <none>
ds-worker-wrmtn                                                   1/1     Running   0          2m56s   10.x.x.146     dggphis18024   <none>           <none>
ds-worker-wrztn                                                   1/1     Running   0          2m56s   10.x.x.146     dggphis18024   <none>           <none>
faas-frontent-8d47bf8d5f-k5s2x                                    1/1     Running   0          104s    10.x.x.217     dggphis18023   <none>           <none>
faas-mananger-c886466cd-xkspb                                     1/1     Running   0          91s     10.x.x.218     dggphis18023   <none>           <none>
faas-scheduler-b862b1c6d8eb-346f745bbd-7w6gj                      1/1     Running   0          81s     10.x.x.219     dggphis18023   <none>           <none>
faas-scheduler-e19efb73cb3c-446f745bbd-7w6gj                      1/1     Running   0          81s     10.x.x.220     dggphis18023   <none>           <none>
function-master-777f6bb8c5-csn5m                                  1/1     Running   0          2m56s   10.x.x.215     dggphis18023   <none>           <none>
function-proxy-5z27x                                              1/1     Running   0          2m56s   10.x.x.146     dggphis18024   <none>           <none>
function-proxy-l7d59                                              1/1     Running   0          2m56s   10.x.x.56      dggphis18023   <none>           <none>
function-proxy-z7d59                                              1/1     Running   0          2m56s   10.x.x.56      dggphis18023   <none>           <none>
iam-adaptor-bb5cf566-dhsvm                                        1/1     Running   0          2m56s   10.x.x.214     dggphis18023   <none>           <none>
meta-service-587d5fc6db-p5wh7                                     1/1     Running   0          2m56s   10.x.x.216     dggphis18023   <none>           <none>
minio-884b9bdb6-bc2bj                                             1/1     Running   0          2m56s   10.x.x.216     dggphis18023   <none>           <none>
```

参考[在 K8s 集群中运行函数服务](example-project-function-k8s-service)进一步验证部署结果。

openYuanrong 日志默认开启并挂载到 K8s 节点上。其中 data worker 组件日志路径为 `/home/sn/datasystem/logs/`，其他组件日志路径为 `/var/paas/sys/log/cff/default/componentlogs`，函数实例 runtime 日志路径为 `/var/paas/sys/log/cff/default/servicelogs`，函数实例用户日志路径为 `/var/paas/sys/log/cff/default/processrouters/stdlogs`。如果部署失败，可通过日志分析原因。

## 自定义 Pod 资源池

部署时创建的默认 Pod 资源池通常用于开发或者测试。实际生产场景中，为了更好的匹配业务工作负载，openYuanrong 支持自定义 Pod 资源池。您可以在部署时通过修改 `values.yaml` 文件创建，也可以通过[资源池管理 API](../api/index.md) 动态创建，两种方式效果一致。

多个业务或者开发、测试等多个环境共用一套 K8s 时，您可以创建多个自定义 Pod 资源池并按业务或者环境类型打上标签，使用openYuanrong 的[亲和调度策略](../../../multi_language_function_programming_interface/development_guide/scheduling/affinity.md)将函数实例指派到特定的 Pod 上，实现隔离。此外，[创建 pod 资源池 API](../api/create_pod_pool.md) 提供了 `node_selector` 和 `affinities` 字段，原生支持 [k8s pod 亲和调度](https://kubernetes.io/zh-cn/docs/concepts/scheduling-eviction/assign-pod-node/){target="_blank"}，将创建的 Pod 指派给特定的节点。两者结合可实现更灵活的资源匹配。

## 卸载

执行如下命令卸载：

```shell
helm uninstall openyuanrong
```
