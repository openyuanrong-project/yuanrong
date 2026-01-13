# 入门

快速部署使用默认配置参数在单节点或多节点 Kubernetes 集群上部署openYuanrong，建议用于学习和开发。

:::{admonition} 部署环境要求
:class: note

- 已部署好 k8s 集群（1.19.4 及以上版本），集群中每个节点（主机）至少有 16 个 CPU，32G 以上内存，总集群至少有 32 个 CPU，64G 以上内存。参考：[环境部署 k8s 集群](https://kubernetes.io/docs/reference/setup-tools/kubeadm/){target="_blank"}

- 配套 k8s 版本的 kubectl 工具。参考：[环境安装](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/){target="_blank"}

- helm 工具（3.2 及以上版本）用于部署openYuanrong。参考：[环境安装](https://helm.sh/zh/docs/intro/install/){target="_blank"}

:::

## 添加openYuanrong的 helm 仓库

openYuanrong k8s 安装包依赖于 helm，需要添加openYuanrong提供的 helm 仓库地址。

```bash
helm repo add yr http://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/charts/
helm repo update
```

## 配置openYuanrong的版本镜像仓库

openYuanrong版本镜像存放在私有镜像仓库上，在终端执行 `vim /etc/docker/daemon.json` 命令，补充如下文件内容，为 docker 添加openYuanrong黄区环境白名单镜像仓库地址。

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

## 部署openYuanrong

openYuanrong依赖三方组件 etcd 和 minio 提供完整服务，以下部署流程会自动为用户安装，如果您 k8s 环境中存在已安装的 minio 及 etcd，需先手动删除。

- etcd: 用于在openYuanrong组件之间共享和交换元信息
- minio: 用于存储用户发布的函数代码包

根据 k8s 硬件平台架构（X86/ARM）不同选择安装包，在任意 k8s 节点上使用 helm 命令部署openYuanrong。

```bash
helm pull --untar yr/openyuanrong
cd openyuanrong
```

查找k8s环境中安装的etcd, 替换values.yaml文件中的base-etcd换成实际k8s环境中的etcd service名字

```bash
helm install openyuanrong .
```

使用 `kubectl get pods` 命令检查openYuanrong组件运行状态：以下 pod 个数正确且运行状态全部为 Running 才表示openYuanrong成功拉起。

```bash
kubectl get pods
# NAME                                                              READY   STATUS    RESTARTS   AGE
# ds-core-etcd-0                                                    1/1     Running   0          2m15s
# ds-worker-7zv94                                                   1/1     Running   0          2m15s
# ds-worker-rd2qv                                                   1/1     Running   0          2m15s
# ds-worker-xkmmw                                                   1/1     Running   0          2m15s
# faas-frontend-7d47bf8d5f-k5s2x                                    1/1     Running   0          2m15s
# faas-manager-c886466cd-xkspb                                      1/1     Running   0          2m15s
# faas-scheduler-546f745bbd-7w6gj                                   1/1     Running   0          2m15s
# faas-scheduler-546f745bbd-7w6gj                                   1/1     Running   0          2m15s
# function-master-6f9b957f9d-bc54t                                  1/1     Running   1          2m15s
# function-proxy-lxxsk                                              1/1     Running   0          2m15s
# function-proxy-62npv                                              1/1     Running   0          2m15s
# function-proxy-nh4xp                                              1/1     Running   0          2m15s
# iam-adaptor-78ff4cbf69-mlwxf                                      1/1     Running   0          2m15s
# meta-service-754df6c44f-mlv6n                                     1/1     Running   0          2m15s
```
 
## 卸载openYuanrong

```bash
helm uninstall openyuanrong
```
