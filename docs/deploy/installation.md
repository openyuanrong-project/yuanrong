(overview-installation)=

# 安装指南

openYuanrong 目前支持在 Linux x86_64 及 aarch64 (ARM) 上安装。不同的开发环境安装依赖如下：

- 安装 openYuanrong 并开发 Python 应用：`python<=3.11,>=3.9` 。
- 开发 Java 应用：`java 8/17/21` 。
- 开发 C++ 应用：`gcc>=10.3.0 且 stdc++>=14` 。

(install-yuanrong-with-pip)=

## 安装 openYuanrong

使用 pip 从 PyPI 上安装 openYuanrong 最新官方版本，内含 openYuanrong SDK 及命令行工具 yr。

```bash
# 支持以下版本
# openyuanrong-0.6.0-cp310-cp310-manylinux_2_34_x86_64.whl
# openyuanrong-0.6.0-cp311-cp311-manylinux_2_34_x86_64.whl
# openyuanrong-0.6.0-cp39-cp39-manylinux_2_34_x86_64.whl
# openyuanrong-0.6.0-cp310-cp310-manylinux_2_34_aarch64.whl
# openyuanrong-0.6.0-cp311-cp311-manylinux_2_34_aarch64.whl
# openyuanrong-0.6.0-cp39-cp39-manylinux_2_34_aarch64.whl
pip install https://openyuanrong.obs.cn-southwest-2.myhuaweicloud.com/release/0.6.0/linux/x86_64/openyuanrong-0.6.0-cp39-cp39-manylinux_2_34_x86_64.whl
```

您可能希望从源码编译 openYuanrong 版本，以满足更多自定义场景，请参考章节：[源码编译 openYuanrong](../contributor_guide/source_code_build.md)。

## 使用 openYuanrong SDK

:::{admonition} SDK 会自动初始化环境
:class: note

[Driver](#glossary-driver) 中调用 yr.init() 接口（不配置 openYuanrong 集群的地址），运行在非 openYuanrong 节点上时，SDK 会尝试拉起一个 openYuanrong 临时环境，该环境在进程退出时**自动销毁**。该机制在调试时可能很有用，但会导致程序运行时间变长。

:::

### Python SDK

[安装 openYuanrong](install-yuanrong-with-pip)后，即可在当前环境中使用，应用开发请参考[开发指南](../multi_language_function_programming_interface/development_guide/index.md)。

### C++ SDK

[安装 openYuanrong](install-yuanrong-with-pip)后，参考如下命令查看 C++ SDK 路径，应用开发请参考[开发指南](../multi_language_function_programming_interface/development_guide/index.md)。

```console
[xxx]# yr version
CLI version: DEV.
Using yuanrong at: /home/.miniconda3/envs/yr/lib/python3.9/site-packages/yr/inner

[xxx]# ls /home/.miniconda3/envs/yr/lib/python3.9/site-packages/yr/inner/runtime/sdk/cpp/
bin  include  lib  VERSION
```

(install-yuanrong-java-sdk)=

### Java SDK

[安装 openYuanrong](install-yuanrong-with-pip)后，参考如下命令查看 Java SDK 路径。应用开发请参考[开发指南](../multi_language_function_programming_interface/development_guide/index.md)。

```console
[xxx]# yr version
CLI version: DEV.
Using yuanrong at: /home/.miniconda3/envs/yr/lib/python3.9/site-packages/yr/inner

[xxx]# ls /home/.miniconda3/envs/yr/lib/python3.9/site-packages/yr/inner/runtime/sdk/java/
yr-api-sdk-1.0.0.jar  pom.xml
```

如果您通过 Maven 管理 Java 项目，可参考如下命令安装 openYuanrong  Java SDK，并在项目 `pom.xml` 文件中引入依赖。

```shell
# 替换 yr-api-sdk-1.0.0.jar 为您实际的 SDK 版本包
mvn install:install-file -Dfile=./yr-api-sdk-1.0.0.jar -DartifactId=yr-api-sdk \
    -DgroupId=com.yuanrong -Dversion=1.0.0 -Dpackaging=jar -DpomFile=./pom.xml
```

修改项目 `pom.xml` 文件，引入 openYuanrong 依赖。

```xml
<dependencies>
    <dependency>
        <groupId>com.yuanrong</groupId>
        <artifactId>yr-api-sdk</artifactId>
        <version>1.0.0</version>
    </dependency>
</dependencies>
```

## 使用 openYuanrong 命令行工具

详见[openYuanrong 命令行工具](../multi_language_function_programming_interface/api/yr_command_line_tool/index.md)。
