# 环境准备

本节介绍 openYuanrong 对部署环境的要求。

## 操作系统及硬件

- 主机操作系统平台为 Linux X86_64 或 ARM_64。
- 单台主机至少有 16 个 CPU，32G 以上内存，总集群至少有 32 个 CPU，64G 以上内存。
- 主机磁盘可用空间大于 40G，用于下载 openYuanrong 组件镜像。

## 依赖服务

部署 openYuanrong 前，需要先安装部署以下工具和服务。

- K8s 集群 及 kubectl 工具 1.19.4 及以上版本。学习和开发可通过 [kind](https://kind.sigs.k8s.io/docs/user/quick-start/){target="_blank"} 部署，生产环境建议使用 [Kubeadm](https://kubernetes.io/docs/reference/setup-tools/kubeadm/){target="_blank"} 部署。
- MinIO 不限制版本，部署可参考 [MinIO Kubernetes](https://minio.org.cn/docs/minio/kubernetes/upstream/){target="_blank"}。
- ETCD v3 版本，部署可参考 [Run etcd clusters as a Kubernetes StatefulSet](https://etcd.io/docs/v3.5/op-guide/kubernetes/){target="_blank"}。
- helm 工具 3.2 及以上版本，用于部署 openYuanrong，安装可参考 [Installing Helm](https://helm.sh/zh/docs/intro/install/){target="_blank"}。
 