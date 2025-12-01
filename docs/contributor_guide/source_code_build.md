
# 源码编译 openYuanrong

## 概述

本章节向您介绍如何从源码开始编译 openYuanrong。

## 环境准备

一台配置如下的主机：

* 硬件

    | 项目       | 要求                 |
    |------------|---------------------|
    | CPU        | 4U 以上，推荐 16U    |
    | 内存        | 10GB 以上，推荐 32G  |
    | 硬盘空间    | 50GB 以上            |
    | 操作系统平台 | Linux x64/Linux ARM |

* 操作系统

    | 操作系统             | 版本                             |
    |---------------------|----------------------------------|
    | openEuler (x86_64)  | openEuler 22.03 LTS 及以上版本    |
    | openEuler (aarch64) | openEuler 22.03 LTS 及以上版本    |

在主机上安装以下编译使用的工具并保持版本一致：

| 工具名称           | 工具类型  | 功能        | x86_64 版本    | aarch64 版本   |
|-------------------|-----------|------------|----------------|---------------|
| AdoptOpenJDK      | 开源第三方 | 编译构建    | 8              | 8             |
| Apache Maven      | 开源第三方 | 编译构建    | 3.9.11         | 3.9.11        |
| Golang            | 开源第三方 | 编译构建    | 1.24.1         | 1.24.1        |
| CMake             | 开源第三方 | 编译构建    | 3.22.0         | 3.22.0        |
| Python            | 开源第三方 | 编译构建    | 3.9/3.10/3.11  | 3.9/3.10/3.11 |
| ninja-build/ninja | 开源第三方 | 编译构建    | 1.13.1         | 1.13.1        |
| bazelbuild/bazel  | 开源第三方 | 编译构建    | 6.5.0          | 6.5.0         |
| Node.js           | 开源第三方 | 编译构建    | 20.19.0        | 20.19.0         |

:::{tip}
使用 openeuler_22.03_LTS_sp4 环境执行命令安装编译工具
```bash
wget https://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/build_tools/openeuler_22.03_LTS/check_tools.sh
wget  https://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/build_tools/openeuler_22.03_LTS/install_tools.sh
bash install_tools.sh
source /etc/profile.d/build_tools.sh
```
:::

## 编译

openYuanrong 分为四个代码仓库：

* [yuanrong-runtime](https://gitee.com/openeuler/yuanrong-runtime){target="_blank"}：运行时仓库，编译依赖 yuanrong-functionsystem 和 yuanrong-datasystem 仓库的发布包。
* [yuanrong-functionsystem](https://gitee.com/openeuler/yuanrong-functionsystem){target="_blank"}：函数系统仓库，编译依赖 yuanrong-datasystem 仓库的发布包。
* [yuanrong-datasystem](https://gitee.com/openeuler/yuanrong-datasystem){target="_blank"} ：数据系统仓库，可独立编译。
* [ray-adapter](https://gitee.com/openeuler/ray-adapter){target="_blank"}：ray adapter 仓库，可独立编译。

使用多语言函数编程接口，需要编译 yuanrong-runtime 仓库；单独使用数据系统接口，需要编译 yuanrong-datasystem 仓库；使用 ray adapter 接口，需要编译 ray-adapter 仓库。

### 编译 yuanrong-datasystem

首先下载源码。创建一个代码目录，例如 `mkdir -p /opt/openyuanrong/`，在目录下执行如下命令。

```bash
# 这里下载的是 yuanrong-datasystem 仓的 master 分支，按需替换为您 fork 的个人仓及分支。
git clone -b master https://gitee.com/openeuler/yuanrong-datasystem.git
```

执行如下脚本编译。

```bash
# 通过 bash build.sh -h 了解更多编译选项
cd /opt/openyuanrong/yuanrong-datasystem
bash build.sh -X off
```

:::{Note}

可通过编译选项 `-j` 设置编译并发度。在保证编译主机空闲 CPU 和内存 1:2 配置（例如 16U32G）的情况下，建议设置为：CPU 核数 + 1，以避免链接阶段内存不足造成编译失败。

:::

编译产物生成在 `yuanrong-datasystem/output` 目录下，包含如下两个文件：

* `yr-datasystem-vx.x.x.tar.gz`：包含 SDK 及 二进制文件。
* `yr_datasystem-x.x.x.whl`：whl 安装包，包含命令行工具和 SDK。

### 编译 yuanrong-functionsystem

首先下载源码。创建一个代码目录，例如 `mkdir -p /opt/openyuanrong/`，在目录下执行如下命令。

```bash
# 这里下载的是 yuanrong-functionsystem 仓的 master 分支，按需替换为您 fork 的个人仓及分支。
git clone -b master https://gitee.com/openeuler/yuanrong-functionsystem.git
```

函数系统的编译依赖数据系统发布包：

* 拷贝已编译好的发布包 `yr-datasystem-vx.x.x.tar.gz` 到function-system/vendor/src目录下面。

```bash
mkdir -p functionsystem/vendor/src
cp /opt/openyuanrong/yunarong-datasystem/output/yr-datasystem-vx.x.x.tar.gz yuanrong-functionsystem/vendor/src/
```


执行如下脚本编译。

```bash
# 通过 bash run.sh -h 了解更多编译选项
cd /opt/openyuanrong/yuanrong-functionsystem
bash run.sh build
```

:::{Note}

可通过编译选项 `-j` 设置编译并发度。在保证编译主机空闲 CPU 和内存 1:2 配置（例如 16U32G）的情况下，建议设置为：CPU 核数 + 1，以避免链接阶段内存不足造成编译失败。
可通过编译选项 `--version` 指定版本信息
 
:::

编译产物生成在 `yuanrong-functionsystem/output` 目录下，包含一个压缩包文件：

* `yr-functionsystem-vx.x.x.tar.gz`：包含二进制文件、lib、sym(二进制符合信息)、metric、etcd 及 cli 压缩包。

### 编译 yuanrong-runtime

首先下载源码。创建一个代码目录，例如 `mkdir -p /opt/openyuanrong/`，在目录下执行如下命令。

```bash
# 这里下载的是 yuanrong-runtime 仓 master 分支，按需替换为您 fork 的个人仓。
git clone -b master https://gitee.com/openeuler/yuanrong-runtime.git
```

运行时的编译依赖数据系统和函数系统发布包：

* 新建目录 `/opt/openyuanrong/datasystem/output`，拷贝已编译好的数据系统发布包 `yr-datasystem-vx.x.x.tar.gz` 到该目录并解压。解压已编译好的函数系统发布包 `yr-functionsystem-vx.x.x.tar.gz`，拷贝包中 `metrics` 文件夹到 `/opt/openyuanrong` 目录下。

执行如下脚本编译。

```bash
# 通过 bash build.sh -h 了解更多编译选项
cd /opt/openyuanrong/yuanrong-runtime
bash build.sh -P
```

:::{Note}

可通过编译选项 `-j` 设置编译并发度。在保证编译主机空闲 CPU 和内存 1:2 配置（例如 16U32G）的情况下，建议设置为：CPU 核数 + 1，以避免链接阶段内存不足造成编译失败。

:::

编译产物生成在 `yuanrong-runtime/output` 目录下，包含如下三个文件：

* `openyuanrong-vx.x.x.tar.gz`：openYuanrong 全量包。
* `openyuanrong-x.x.x-cp39-cp39-manylinux_2_17_x86_64.whl`：whl 安装包，包括命令行工具及 Python SDK。
* `yr-runtime-vx.x.x.tar.gz`：运行时发布包。

### 编译 ray-adapter

首先下载源码。创建一个代码目录，例如 `mkdir -p /opt/openyuanrong/`，在目录下执行如下命令。

```bash
# 这里下载的是 ray-adapter 仓 master 分支，按需替换为您 fork 的个人仓。
git clone -b master https://gitee.com/openeuler/ray-adapter.git
```

执行如下脚本编译。

```bash
# 通过 bash build.sh -h 了解更多编译选项
cd /opt/openyuanrong/ray-adapter
bash build.sh
```

编译产物生成在 `ray-adapter/output` 目录下，包含如下文件：

* `ray_adapter-x.x.x.whl`：whl 安装包。
