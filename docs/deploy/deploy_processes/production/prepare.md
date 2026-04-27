# 环境准备

本节介绍 openYuanrong 对部署环境的要求。

## 操作系统及硬件

- 主机为 X86_64 或 ARM_64 平台的 Linux 系统。我们对 openEuler 22.03 LTS/24.03 LTS、Ubuntu 22.04 LTS 及 Debian 12/13 做了充分测试，如果您在其他 Linux 系统上遇到部署问题，请通过提 Issue 的方式联系我们。
- 单台主机至少有 2 个 CPU，4G 以上内存。
- 集群中所有主机网络互通，并参考[部署参数](../parameters.md)中端口相关配置，放通 openYuanrong 组件、开源 etcd 等通信所需端口。

    您可以使用 `ping` 命令检测两台主机是否网络互通，使用 [netcat](https://netcat.sourceforge.net/){target="_blank"} 之类的工具来检查端口是否开放，例如：

    ```shell
    nc -vz 127.0.0.1 8888
    ```

    测试环境可运行以下命令停止并禁用防火墙，保证端口可用。

    ```shell
    systemctl stop firewalld
    systemctl disable firewalld

    # 验证操作结果，如果 Active 属性显示为 inactive 说明已经关闭。
    systemctl status firewalld
    ```

## 依赖服务和工具

部署使用 openYuanrong 命令行工具 `yr`。它依赖 [net-tools](https://net-tools.sourceforge.io/){target="_blank"} 自动获取主机 IP，依赖 [netcat](https://netcat.sourceforge.net/){target="_blank"} 在随机分配组件端口时监测是否存在端口冲突，您可通过如下命令安装：

```shell
yum install -y net-tools
# 验证是否安装成功。
ifconfig

yum install -y nc
# 验证是否安装成功。
nc -vz 127.0.0.1 8888
```

:::{Note}

net-tools 和 netcat 均为开源工具，存在网络端口嗅探风险，需谨慎使用。

:::

## 开发依赖及限制

openYuanrong 支持使用 C++（版本：14 及以上）、Java（版本：8、17、21）、Python（版本：3.9、3.10、3.11）开发应用，您需要在部署主机上安装对应语言的运行环境。特别地，在安装多个 Java 及 Python 语言版本时，需要在 /usr/bin 建立相应的软链接，参考命令如下。

```shell
# java
ln -s ${JAVA8_INSTALL_HOME}/jre/bin/java /usr/bin/java1.8
ln -s ${JAVA17_INSTALL_HOME}/jre/bin/java /usr/bin/java17
ln -s ${JAVA21_INSTALL_HOME}/jre/bin/java /usr/bin/java21
# python
ln -sf $PYTHON_PATH_3_9/bin/python3.9 /usr/bin/python3.9
ln -sf $PYTHON_PATH_3_10/bin/python3.10 /usr/bin/python3.10
ln -sf $PYTHON_PATH_3_11/bin/python3.11 /usr/bin/python3.11
```

其他约束：

- 单个集群最大部署 1000 个从节点。
- 单个集群最大运行 10000 个函数实例。
- 单个数据对象大小不能超过配置的节点共享内存。
