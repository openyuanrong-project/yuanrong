
# 源码编译 openYuanrong

## 概述

本章节向您介绍如何从源码开始编译 openYuanrong。

## 环境准备

一台配置如下的主机：

* 硬件

    | 项目       | 要求                  |
    |------------|---------------------|
    | CPU        | 4核 以上，推荐 16核        |
    | 内存        | 10GB 以上，推荐 32GB     |
    | 硬盘空间    | 50GB 以上             |
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
我们提供了在 openeuler_22.03_LTS_sp4 环境上使用的脚本，方便快速安装编译工具，参考以下命令下载并使用。

```bash
yum install -y wget
wget https://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/build_tools/openeuler_22.03_LTS/check_tools.sh
wget https://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/build_tools/openeuler_22.03_LTS/install_tools.sh
bash install_tools.sh
source /etc/profile.d/buildtools.sh
```

:::

## 编译

openYuanrong 分为四个代码仓库：

* [yuanrong](https://gitee.com/openeuler/yuanrong){target="_blank"}：运行时仓库，编译依赖 yuanrong-functionsystem 和 yuanrong-datasystem 仓库的发布包。
* [yuanrong-functionsystem](https://gitee.com/openeuler/yuanrong-functionsystem){target="_blank"}：函数系统仓库，编译依赖 yuanrong-datasystem 仓库的发布包。
* [yuanrong-datasystem](https://gitee.com/openeuler/yuanrong-datasystem){target="_blank"} ：数据系统仓库，可独立编译。
* [ray-adapter](https://gitee.com/openeuler/ray-adapter){target="_blank"}：ray adapter 仓库，可独立编译。

使用多语言函数编程接口，需要编译 yuanrong 仓库；单独使用数据系统接口，需要编译 yuanrong-datasystem 仓库；使用 ray adapter 接口，需要编译 ray-adapter 仓库。

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

**Step1: 源码下载**

在编译目录中通过Git或其他方式下载函数系统的源代码。以下示例将在`/opt/openyuanrong`文件夹中下载函数系统源码为`yuanrong-functionsystem`文件夹。

```shell
cd /opt/openyuanrong
git clone -b master https://gitee.com/openeuler/yuanrong-functionsystem.git
```

**Step2: 获取依赖**

根据架构设计，函数系统的编译过程会依赖数据系统的SDK，因此需要将数据系统的编译产物或发布产物拷贝至函数系统的构建依赖中。函数系统的构建脚本将会在启动时根据[`vendor/VendorList.csv`文件](https://gitee.com/openeuler/yuanrong-functionsystem/blob/master/vendor/VendorList.csv)的配置自动下载或引入依赖，对于数据系统的编译依赖将默认通过本地文件的方式引入。即使用路径描述符`file://localhost/vendor/src/yr-datasystem.tar.gz`描述数据系统来源且不会校验依赖Hash值。

假设数据系统的编译产物路径为`/opt/openyuanrong/yuanrong-datasystem/output/yr-datasystem-v0.0.0.tar.gz`，您可以通过以下命令将数据系统编译产物拷贝至`VendorList.csv`中配置的默认路径。

```shell 
cd /opt/openyuanrong
mkdir -p yuanrong-functionsystem/vendor/src
cp -a yuanrong-datasystem/output/yr-datasystem-v0.0.0.tar.gz yuanrong-functionsystem/vendor/src/yr-datasystem.tar.gz
```

**Step3: 源码编译**

在函数系统的源码根目录中，我们可以使用内置的编译脚本的`build`指令实现源码的编译。编译过程将分为三个环节：三方件下载编译，二方件编译，系统源码编译。整个流程将自动完成，对于已经完成编译的组件将自动跳过。

```shell
cd /opt/openyuanrong/yuanrong-functionsystem
chmod +x run.sh
./run.sh build
```

`build`指令关键参数：

* `JOB_NUM`，`-j`：编译并发度，设置程序编译时使用的CPU核心数量。默认值为：OS核心数
* `VERSION`，`-v`：编译版本号，设置程序编译是程序中记录的版本号。默认值为：0.0.0

更多指令请参考：`./run.sh --help`，`./run.sh build --help`

:::{Note}

建议通过编译选项 `-j` 设置编译并发度，以避免编译链接阶段内存不足造成编译失败。内存峰值占用(GB)约为并发度的1.5~1.8倍，即`MEM(GB) ≈ JOB(CPU) * 1.8(GB/CPU)`，建议设置安全并发度：CPU 核数 /2

:::

**Step4: 产物打包**

函数系统的主要编译产物为多个二进制程序与公共的动态库，同时将搭配一些依赖和运行配置。因此您可以使用`pack`命令完成函数系统的构建产物打包。打包产物和打包内容将统一存放在`output`文件夹中。

```
cd /opt/openyuanrong
chmod +x run.sh
./run.sh pack
```

`pack`指令关键参数：

* `VERSION`，`-v`：编译版本号，设置程序编译是程序中记录的版本号。默认值为：0.0.0

更多指令请参考：`./run.sh --help`，`./run.sh pack --help`

产物目录与关键产物如下所示：

```
tree output/
├── functionsystem
│   ├── bin          # 函数系统编译产物
│   ├── config       # 函数系统运行配置
│   ├── deploy       # 函数系统部署配置
│   ├── lib          # 函数系统动态库
│   └── sym          # 函数系统符号表
├── metrics
│   ├── config       # 可观测配置
│   ├── include      # 可观测接口
│   └── lib          # 可观测动态库
├── metrics.tar.gz   # 可观测打包产物
└── yr-functionsystem-v0.0.0.tar.gz # 函数系统打包产物
```

### 编译 yuanrong

首先下载源码。创建一个代码目录，例如 `mkdir -p /opt/openyuanrong/`，在目录下执行如下命令。

```bash
# 这里下载的是 yuanrong 仓 master 分支，按需替换为您 fork 的个人仓。
git clone -b master https://gitee.com/openeuler/yuanrong.git
```

运行时的编译依赖数据系统和函数系统发布包：

* 新建目录 `/opt/openyuanrong/yuanrong/datasystem/output`，拷贝已编译好的数据系统发布包 `yr-datasystem-vx.x.x.tar.gz` 到该目录并通过 `tar -zxf yr-datasystem-vx.x.x.tar.gz --strip-components=1` 命令解压。解压已编译好的函数系统发布包 `yr-functionsystem-vx.x.x.tar.gz`，拷贝包中 `metrics` 文件夹到 `/opt/openyuanrong/yuanrong/` 目录下。

执行如下脚本编译。

```bash
# 通过 bash build.sh -h 了解更多编译选项
cd /opt/openyuanrong/yuanrong
bash build.sh -P
```
如果出现依赖下载失败报错，可以尝试删除 /opt/openyuanrong/yuanrong/thirdparty/runtime_deps 目录后重新编译

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
