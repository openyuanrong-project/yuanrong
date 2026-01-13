# 环境准备

本节介绍openYuanrong对部署环境的要求

## 操作系统及硬件

- 主机操作系统平台为 Linux X86_64 或 ARM_64
- 单台主机至少有 16 个 CPU，32G 以上内存，总集群至少有 32 个 CPU，64G 以上内存。
- 主机磁盘可用空间大于 40G，用于下载openYuanrong组件镜像
- 主机之间通信正常，开发测试环境可以关闭防火墙规避

    ```bash
    systemctl stop firewalld
    systemctl disable firewalld
    ```

## 依赖服务

- k8s 集群 及 kubectl 工具 1.19.4 及以上版本
- MinIO 不限制版本
- ETCD v3 版本
- helm 工具 3.2 及以上版本，用于部署openYuanrong。参考：[环境安装](https://helm.sh/zh/docs/intro/install/){target="_blank"}
 