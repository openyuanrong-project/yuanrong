# 部署 openYuanrong 作业类应用

```{eval-rst}
.. toctree::
   :glob:
   :maxdepth: 1
   :hidden:

   api/index
```

本章节向您介绍如何构建使用单机程序分布式并行化接口开发的作业类应用并在 openYuanrong 集群中部署。

## 环境依赖

您的应用可能包含除运行时外的其他依赖，例如：

- Python 和 Java 程序运行依赖的包，C++ 程序运行依赖的动态库。
- openYuanrong 函数中需要读取的环境变量。
- openYuanrong 函数中需要读取的数据文件，使用的工具等。

openYuanrong 函数可能运行在集群中的任意节点，因此这些依赖需要在每个 openYuanrong 节点存在并保持一致。在生产环境中，建议预先安装相关的依赖。

## 使用 REST API 提交作业

部署集群后（在 VM 或 Kubernetes 上），您就可以运行作业应用程序了！

![](../../images/running_job_ways.png)

提交到作业 API 的内容包括：

- 一个入口命令，例如 python my_script.py

- 一个运行时环境，用于指定应用程序的文件和包依赖项。

作业可以由位于集群外部的远程客户端提交。

提交作业后，它将运行一次直到完成或失败，而与原始提交者的连接无关。重试或使用不同参数进行的不同运行应由提交者处理。作业与集群的生命周期绑定，若集群崩溃，则该集群上所有正在运行的作业都将被终止。

## 本地运行作业类应用

### Python 应用

Python 是解释性语言，无需编译过程。您在 openYuanrong 集群中部署 Python 应用时，需要保证集群中所有节点都已经安装应用所需的依赖。

### C++ 应用

C++ 应用需要编译一个二进制 Driver 程序，以及包含所有 openYuanrong 函数的动态库。部署时，将动态库拷贝到 openYuanrong 集群所有节点的相同路径下，并选择如下方式之一配置该路径。

- 运行 Driver 程序时，使用 `--codePath` 参数指定动态库的绝对路径。
- Driver 程序中调用 `YR::Init(const Config &conf)` 接口时，配置 [Config](../../multi_language_function_programming_interface/api/distributed_programming/zh_cn/Cpp/struct-Config.rst) 的 `loadPaths` 参数。

以使用 CMake 构建为例，一个简单的应用工程目录如下。`src` 目录存放源码，`CMakeLists.txt` 是 CMake 构建系统使用的配置文件，`build` 目录用于存放构建生成的二进制和动态库文件。在 `build` 目录下依次执行命令 `cmake ..` 和 `make`，即可完成构建。

```text
yr-cpp-demo
├── src
│   └── demo.cpp
├── CMakeLists.txt
└── build
```

CMakeLists.txt 文件参考：

```cmake
cmake_minimum_required(VERSION 3.16.1)
# 指定项目名称，例如：yr-cpp-demo
project(yr-cpp-demo LANGUAGES C CXX)
set(CMAKE_CXX_STANDARD 17)

# 指定编译生成文件路径在 build 目录下
set(SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR})
set(BINARY_DIR ${SOURCE_DIR}/build)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${BINARY_DIR})

set(CMAKE_CXX_FLAGS "-pthread")
set(BUILD_SHARED_LIBS ON)

# 替换 YR_INSTALL_PATH 的值为 openYuanrong 实际安装路径
set(YR_INSTALL_PATH "/usr/local/lib/python3.9/site-packages/yr")
link_directories(${YR_INSTALL_PATH}/cpp/lib)
include_directories(
    ${YR_INSTALL_PATH}/cpp/include
)

# 生成可执行文件 cpp-demo，修改 demo.cpp 为您对应的源码文件
add_executable(cpp-demo src/demo.cpp)
target_link_libraries(cpp-demo yr-api)

# 生成动态库文件 cpp-demo-dll，修改 demo.cpp 为您对应的源码文件
add_library(cpp-demo-dll SHARED src/demo.cpp)
target_link_libraries(cpp-demo-dll yr-api)
```

构建成功将生成二进制文件 `cpp-demo` 和 动态库文件 `libcpp-demo-dll.so`。以使用 `/opt/openyuanrong/function/demo` 作为代码路径为例，需要拷贝 `libcpp-demo-dll.so` 文件到集群所有节点该路径下。在 `build` 目录下执行命令 `./cpp-demo --codePath=/opt/openyuanrong/function/demo` 即可运行应用。

### Java 应用

Java 应用构建的 jar 包可根据需要自行选择是否包含依赖。部署时，将 jar 包和依赖拷贝到 openYuanrong 集群所有节点的相同路径下，并选择如下方式之一指定该路径。

- 运行应用时，使用 `-Dyr.codePath` 参数指定 jar 包的绝对路径。
- 应用程序中调用 `init(Config conf)` 接口时，配置 [Config](../../multi_language_function_programming_interface/api/distributed_programming/zh_cn/Java/Config.md) 的 `loadPaths` 参数。

以使用 Maven 构建为例，一个简单的应用工程目录如下。`demo` 目录存放源码，`pom.xml` 是 Maven 构建系统使用的配置文件。在 `yr-java-demo` 目录下执行命令 `mvn clean package`，即可完成构建。

```text
yr-java-demo
├── pom.xml
└── src
    └── main
        └── java
            └── com
                └── demo
                    └── Main.java
```

pom.xml 文件参考：

```xml
<?xml version="1.0" encoding="UTF-8"?>

<project xmlns="http://maven.apache.org/POM/4.0.0"
        xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
        xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>com.demo</groupId>
    <artifactId>main</artifactId>
    <version>1.0.0</version>

    <properties>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <maven.compiler.source>1.8</maven.compiler.source>
        <maven.compiler.target>1.8</maven.compiler.target>
    </properties>

    <dependencies>
        <dependency>
            <!-- 修改版本号为您实际使用版本 -->
            <groupId>org.yuanrong</groupId>
            <artifactId>yr-api-sdk</artifactId>
            <version>1.0.0</version>
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-assembly-plugin</artifactId>
                <executions>
                    <execution>
                        <phase>package</phase>
                        <goals>
                            <goal>single</goal>
                        </goals>
                    </execution>
                </executions>
                <configuration>
                    <archive>
                        <manifest>
                            <mainClass>com.demo.Main</mainClass>
                        </manifest>
                    </archive>
                    <descriptorRefs>
                        <descriptorRef>jar-with-dependencies</descriptorRef>
                    </descriptorRefs>
                    <appendAssemblyId>false</appendAssemblyId>
                </configuration>
            </plugin>
        </plugins>
    </build>
</project>
```

构建成功将在 `yr-java-demo/target` 目录下生成 jar 包 `main-1.0.0.jar`。以使用 `/opt/openyuanrong/function/demo` 作为代码路径为例，需要拷贝 `main-1.0.0.jar` 文件到集群所有节点该路径下。在 `target` 目录下执行命令 `java -Dyr.codePath=/opt/openyuanrong/function/demo -cp ./main-1.0.0.jar com.demo.Main` 即可运行应用。
