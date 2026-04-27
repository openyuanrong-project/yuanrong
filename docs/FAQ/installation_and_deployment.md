# 安装部署 FAQ

## 一、安装问题

### 成功安装 openYuanrong，但运行 Python Driver 程序时报错：“ModuleNotFoundError: No module named 'yr'”

原因：当前环境存在多个 Python 版本。

解决方法：使用您安装 openYuanrong 时对应的 Python 版本运行程序。

## 二、部署问题

### 在主机上部署 openYuanrong，通过 `yr status` 命令查看集群状态，从节点未加入集群

- 原因一：主机开启了防火墙，openYuanrong 组件使用的端口网络不通。

  解决方法：非生产环境可通过如下命令暂时关闭防火墙并重新部署，生产环境需开放 openYuanrong 使用的端口。

  ```shell
  systemctl stop firewalld
  systemctl disable firewalld
  ```

- 原因二：主从节点安装的 openYuanrong 版本不一致，可通过 `pip show yr` 命令查看安装版本。

  解决方法：在每个节点重装相同的 openYuanrong 版本。

### 在主机上部署 openYuanrong，通过 `yr status` 命令查看集群状态时报错：“Connect to etcd server failed.context deadline exceeded.”

原因：当前环境中使用 http_proxy/https_proxy 配置了代理。

解决方法：关闭代理配置。
