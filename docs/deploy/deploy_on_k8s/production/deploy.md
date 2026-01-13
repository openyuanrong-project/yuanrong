# 部署openYuanrong

本节介绍部署openYuanrong的流程。

(k8s-deploy-download-release-package)=

## 获取发布包

openYuanrong发布包在 [每日构建上获取](https://build-logs.openeuler.openatom.cn:38080/temp-archived/openeuler/openYuanrong/yr_daily/){target="_blank"}。根据不同的使用平台选择软件包并下载其中的 `openyuanrong-*.tar.gz` 文件即可。

## 配置镜像仓

在终端执行 `vim /etc/docker/daemon.json` 命令，增加如下文件内容，为 docker 添加openYuanrong白名单镜像仓库地址，openYuanrong组件及基础镜像存放在这里。

```json
{
    "insecure-registries": [
        "openyuanrong.obs.cn-southwest-2.myhuaweicloud.com"
    ]
}
```

执行如下命令使配置生效：

```shell
systemctl daemon-reload
systemctl restart docker
```

## 部署

openYuanrong部署基于 helm 工具，按需设置好配置文件即可一键安装。

:::{caution}

如果 集群环境中 ETCD 非全新安装，请先清理残留数据。

:::

执行如下命令解压发布包：

```bash
tar -zxf openyuanrong-*.tar.gz -C .
cd openyuanrong/deploy/k8s/charts/
```

在 `openyuanrong/deploy/k8s/charts/` 目录下参考以下命令全新部署openYuanrong。如果您你已有openYuanrong集群，可直接执行 `helm upgrade yr .` 升级。

- 只部署openYuanrong，不创建 Pod 资源池，需要配置 MinIO 的密码，适合自行通过[资源池管理 API](../api/meta_service/index.md) 创建资源池的场景。

  ```shell
  helm install openyuanrong --set global.obsManagement.s3SecretKey=${Your Minio Password} .
  ```

- 部署openYuanrong并创建默认规格为（3 核 cpu，6GB 内存）的 Pod 资源池，建议用于开发有状态函数。

  ```shell
  helm install openyuanrong --set global.obsManagement.s3SecretKey=${Your Minio Password},global.pools[0].id=default,global.pools[0].size=1,\
  global.pools[0].resources.requests.cpu=3000m,global.pools[0].resources.requests.memory=6144Mi,\
  global.pools[0].resources.limits.cpu=3000m,global.pools[0].resources.limits.memory=6144Mi, .
  ```

  您也可以在部署完成后按需拉起该默认 Pod 资源池。

  ```shell
  kubectl scale deployment/function-agent-default --replicas=1
  ```

- 部署openYuanrong并创建默认规格为 600 毫核 cpu，512MB 内存的 Pod 资源池，建议用于开发 FaaS 函数。

  ```shell
  helm install openyuanrong --set global.obsManagement.s3SecretKey=${Your Minio Password},global.pools[0].id=pool-600-512-fusion,global.pools[0].size=1,\
  global.pools[0].resources.requests.cpu=600m,global.pools[0].resources.requests.memory=512Mi,\
  global.pools[0].resources.limits.cpu=600m,global.pools[0].resources.limits.memory=512Mi, .
  ```

  您也可以在部署完成后按需拉起该默认 Pod 资源池。

  ```shell
  kubectl scale deployment/function-agent-pool-600-512-fusion --replicas=1
  ```

- 其他：修改单个组件的镜像版本示例。

  ```shell
  helm install openyuanrong --set global.images.functionProxy=function-proxy:${version} .
  ```

检查部署结果：以下所有 Pod 全部为 Running 状态即部署成功。

```shell
kubectl get pods -owide -w

NAME                                                              READY   STATUS    RESTARTS   AGE     IP              NODE           NOMINATED NODE   READINESS GATE
ds-core-etcd-0                                                    1/1     Running   0          19d     10.42.0.122     dggphis18024   <none>           <none>
ds-worker-llcq8                                                   1/1     Running   0          2m56s   10.244.177.56   dggphis18023   <none>           <none>
ds-worker-wrmtn                                                   1/1     Running   0          2m56s   10.247.23.146   dggphis18024   <none>           <none>
ds-worker-wrztn                                                   1/1     Running   0          2m56s   10.247.23.146   dggphis18024   <none>           <none>
faas-frontent-8d47bf8d5f-k5s2x                                    1/1     Running   0          104s    10.42.1.217     dggphis18023   <none>           <none>
faas-mananger-c886466cd-xkspb                                     1/1     Running   0          91s     10.42.1.218     dggphis18023   <none>           <none>
faas-scheduler-b862b1c6d8eb-346f745bbd-7w6gj                      1/1     Running   0          81s     10.42.1.219     dggphis18023   <none>           <none>
faas-scheduler-e19efb73cb3c-446f745bbd-7w6gj                      1/1     Running   0          81s     10.42.1.220     dggphis18023   <none>           <none>
function-master-777f6bb8c5-csn5m                                  1/1     Running   0          2m56s   10.42.1.215     dggphis18023   <none>           <none>
function-proxy-5z27x                                              1/1     Running   0          2m56s   10.247.23.146   dggphis18024   <none>           <none>
function-proxy-l7d59                                              1/1     Running   0          2m56s   10.244.177.56   dggphis18023   <none>           <none>
function-proxy-z7d59                                              1/1     Running   0          2m56s   10.244.177.56   dggphis18023   <none>           <none>
iam-adaptor-bb5cf566-dhsvm                                        1/1     Running   0          2m56s   10.42.1.214     dggphis18023   <none>           <none>
meta-service-587d5fc6db-p5wh7                                     1/1     Running   0          2m56s   10.42.1.216     dggphis18023   <none>           <none>
```

openYuanrong日志默认开启并挂载到 K8s 节点上。其中 data worker 组件日志路径为 `/home/sn/datasystem/logs/`，其他组件日志路径为 `/var/paas/sys/log/cff/default/componentlogs`，函数实例 runtime 日志路径为 `/var/paas/sys/log/cff/default/servicelogs`，函数实例用户日志路径为 `/var/paas/sys/log/cff/default/processrouters/stdlogs`。如果部署失败，可通过日志分析原因。

## 自定义 Pod 资源池

部署时创建的默认 Pod 资源池通常用于开发或者测试。实际生产场景中，为了更好的匹配业务工作负载，openYuanrong支持自定义 Pod 资源池。您可以通过修改 values.yaml 文件在部署时自定义 Pod 资源池，也可以通过[资源池管理 API](../api/meta_service/index.md) 动态创建，两种方式效果一致。

多个业务或者开发、测试等多个环境共用一套 K8s 时，您可以创建多个自定义 Pod 资源池并按业务或者环境类型打上标签，使用openYuanrong的[亲和调度策略](../../../multi_language_function_programming_interface/development_guide/scheduling/affinity.md)将函数实例指派到特定的 Pod 上，实现隔离。另外，[创建 pod 资源池 API](../api/meta_service/create_pod_pool.md) 提供了 node_selector 和 affinities 字段，原生支持 [k8s pod 亲和调度](https://kubernetes.io/zh-cn/docs/concepts/scheduling-eviction/assign-pod-node/){target="_blank"}，将创建的 Pod 指派给特定的节点。两者结合可实现更灵活的资源匹配。

## 卸载

执行如下命令卸载：

```shell
helm uninstall openyuanrong
```
