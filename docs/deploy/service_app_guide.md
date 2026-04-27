# 部署 openYuanrong 服务类应用

本章节向您介绍如何构建并部署使用 openYuanrong 函数服务接口开发的服务类应用，应用通过 REST API 对外提供服务。这里以单个 openYuanrong 函数开发的应用为例展开介绍，多函数应用每个函数都需要单独部署。

## 环境依赖

您的应用可能包含除运行时外的其他依赖，例如：

- Python 和 Java 程序运行依赖的包，C++ 程序运行依赖的动态库。
- openYuanrong 函数中需要读取的环境变量。
- openYuanrong 函数中需要读取的数据文件，使用的工具等。

openYuanrong 函数可能运行在集群中的任意节点，因此这些依赖需要在每个 openYuanrong 节点存在并保持一致。函数需要的环境变量可在注册函数时设置，通过上下文获取。在主机集群上部署时，建议预先在主机节点上安装其他依赖；在 K8s 集群上部署时，相关依赖可以和函数代码打包在一起或通过网络共享。

## 构建部署包

函数部署在 K8s 集群上时，需要将编译好的代码打成 zip 包，您也可以选择将依赖一起打包。

Python 是解释性语言，无需编译过程，打包时确保函数的 Handler 等方法实现所在源码文件存放在包的根目录。例如源码文件名为 `handler.py`，使用如下打包命令将生成部署包 `demo.zip`。

```bash
zip -j demo.zip handler.py
```

C++ 函数需要先编译为二进制，编译依赖 openYuanrong [C++ SDK](install-yuanrong-cpp-sdk)。打包时确保函数编译生成的二进制文件存放在包的根目录。例如函数二进制文件为 `handler`，参考如下打包命令可生成应用代码包 `demo.zip`。

```bash
ldd ./handler | awk '/=>/ {print $3}' | grep -v '^$' | tr '\n' '\0' | xargs zip -j demo.zip ./handler
```

构建 Java 函数的 jar 包依赖 openYuanrong [Java SDK](install-yuanrong-java-sdk)。函数及依赖都在一个 jar 包中时，无需打包；分开为多个文件时，函数 jar 包放在根目录，全部文件打成一个 zip 包。

## 注册函数

函数需要先注册到 openYuanrong 集群才能被调用。使用[注册函数](../multi_language_function_programming_interface/api/function_service/register_function.md) API 完成函数注册，API 涉及的基础配置项如下：

- name：函数名称，参考 API 中的 FaaS 函数命名规则填写。
- runtime：函数使用的运行时类型，枚举值如 Python3.9，和您开发环境的语言版本保持一致。
- kind：函数类型，配置为 `faas`。
- cpu：函数运行需要的 CPU 资源量。
- memory：服务运行需要的内存资源量。
- timeout：函数处理请求的超时时间。
- storageType：函数代码包存储类型。在 openYuanrong 主机集群上部署时，建议配置为 `local`，即存放在主机磁盘上；在 openYuanrong K8s 集群上部署时，建议配置为 `s3`，即存放在支持 S3 协议的对象存储服务如 MinIO 中。
- handler：函数中 Handler 方法的定义。
  - Python 语言为 `{源码文件名}.{Handler 方法名}`，例如 python_func.handler。
  - C++ 语言为 `{函数二进制文件名}`，例如 cpp_func。
  - Java 语言为 `{包名}.{类名}.{Handler 方法名}`，例如 com.xxx.MyApp.myHandler。
- extendedHandler: 函数中 Initializer 和 PreStop 方法的定义，配置方式与 Handler 方法相同。
- extendedTimeout: 函数中 Initializer 和 PreStop 方法的执行超时时间。

通过[函数服务示例工程](example-project-function-service)查看完整示例。

## 调用函数

函数支持通过 REST API 同步调用，流式（Server-Send Events）或非流式应答。流式应答数据将被分块，逐步生成并返回。非流式应答一次性返回全部结果。

使用[调用函数服务] API 执行函数调用，API 涉及的主要配置项如下：

- Header 参数 `Accept`：在使用流式应答时设置。
- Header 参数 `X-Pool-Label`：在 K8s 集群上部署时，函数实例亲和的资源池标签。
- Header 参数 `X-Instance-Label`：在具有该标签的函数实例上处理请求。您可以在创建函数预留实例时配置标签。

通过[函数服务示例工程](example-project-function-service)查看完整示例。
