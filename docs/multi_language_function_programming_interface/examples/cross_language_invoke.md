# 跨语言调用无状态和有状态函数

 openYuanrong 支持跨语言调用无状态和有状态函数，本节将通过示例工程介绍如何开发。

- [Python 程序中调用 C++ 无状态和有状态函数](cross-language-python-invoke-cpp)
- [Python 程序中调用 Java 无状态和有状态函数](cross-language-python-invoke-java)
- [C++ 程序中调用 Python 无状态和有状态函数](cross-language-cpp-invoke-python)
- [C++ 程序中调用 Java 无状态和有状态函数](cross-language-cpp-invoke-java)
- [Java 程序中调用 C++ 无状态和有状态函数](cross-language-java-invoke-cpp)

## 准备工作

参考[在主机上部署](../../deploy/deploy_processes/index.md)完成 openYuanrong 部署。

:::{Note}

部署时确保配置 `--enable_meta_service=true` 参数，启动 meta service 组件用于注册函数。

:::

(cross-language-python-invoke-cpp)=

## Python 程序中调用 C++ 无状态和有状态函数

我们将 C++ 无状态与有状态函数编译成动态库，在 Python 主程序中使用 [yr.cpp_function](../api/distributed_programming/zh_cn/Python/yr.cpp_function.rst) 和 [yr.cpp_instance_class](../api/distributed_programming/zh_cn/Python/yr.cpp_instance_class.rst) API 调用。

1. 准备示例工程

    在 openYuanrong 集群所有节点新建目录例如：`/opt/mycode/python-invoke-cpp`，用于存放 C++ 无状态函数和有状态函数代码包。

    新建工程目录 `python-invoke-cpp` 及文件如下：
   - calculator.cpp：定义 C++ 无状态函数 `Square` 和有状态函数 `Counter`。
   - CMakeLists.txt：CMake 的配置文件，用于编译构建 C++ 工程。
   - build：空目录，用于存放 C++ 编译构建中生成的文件。
   - main.py：python 主程序，调用 C++ 无状态函数 `Square` 和有状态函数 `Counter`。

    ```bash
    python-invoke-cpp
    ├── calculator.cpp
    ├── CMakeLists.txt
    ├── build
    └── main.py
    ```

    :::{dropdown} calculator.cpp 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```cpp
    #include "yr/yr.h"
    int Square(int x)
    {
        return x * x;
    }

    // 定义无状态函数 Square
    YR_INVOKE(Square)

    class Counter {
    public:
        Counter() {}
        Counter(int init) { count = init; }

        static Counter *FactoryCreate(int init) {
            return new Counter(init);
        }

        void Add() {
            count += 1;
        }

        int Get() {
            return count;
        }

        YR_STATE(count);

    public:
        int count;
    };

    // 定义有状态函数 Counter
    YR_INVOKE(Counter::FactoryCreate, &Counter::Add, &Counter::Get);
    ```

    :::
    :::{dropdown} CMakeLists.txt 文件内容，**须对应修改 YR_INSTALL_PATH 为您的 openYuanrong 安装路径**
    :chevron: down-up
    :icon: chevron-down

    ```cmake
    cmake_minimum_required(VERSION 3.16.1)
    project(calculator CXX)

    set(CMAKE_CXX_STANDARD 17)
    set(SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR})
    set(BINARY_DIR ${SOURCE_DIR}/build)
    set(BUILD_SHARED_LIBS ON)

    # 替换 YR_INSTALL_PATH 的值为 openYuanrong 安装路径，可通过 yr version 命令查看
    set(YR_INSTALL_PATH "/usr/local/lib/python3.9/site-packages/yr/inner")
    link_directories(${YR_INSTALL_PATH}/runtime/sdk/cpp/lib)
    include_directories(
        ${YR_INSTALL_PATH}/runtime/sdk/cpp/include
    )

    add_library(calculator SHARED calculator.cpp)
    set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${BINARY_DIR})
    ```

    :::
    :::{dropdown} main.py 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```python
    import yr

    yr.init()

    # 我们将在后续步骤中注册C++函数，生成如下URN
    cpp_function_urn = "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-mycpp:$latest"

    # 调用C++无状态函数Square
    cpp_function = yr.cpp_function("Square", cpp_function_urn)
    res = cpp_function.invoke(2)
    print(yr.get(res))

    # 调用C++有状态函数Counter的Add和Get方法
    cpp_function_instance = yr.cpp_instance_class("Counter", "Counter::FactoryCreate", cpp_function_urn).invoke(0)
    res = cpp_function_instance.Add.invoke()
    yr.get(res)
    res = cpp_function_instance.Get.invoke()
    print(yr.get(res))

    # 销毁函数实例
    cpp_function_instance.terminate()
    yr.finalize()
    ```

    :::

2. 编译构建 C++ 函数为动态库

    在 `python-invoke-cpp/build` 目录下，执行如下命令构建：

    ```bash
    cmake ..
    make
    ```

    成功构建将在该目录下生成动态库文件 `libcalculator.so` ，拷贝文件到所有节点您创建的代码包目录例如：`/opt/mycode/python-invoke-cpp`。

3. 注册 C++ 函数为 openYuanrong 函数

    注册名为 `0-yr-mycpp` 的 openYuanrong 函数，对应函数 URN 为 `sn:cn:yrk:12345678901234561234567890123456:function:0-yr-mycpp:$latest`。
    
    使用 curl 工具调用 [注册函数 API](../api/function_service/register_function.md) 动态完成注册。
    
    ```bash
    # 替换 /opt/mycode/python-invoke-cpp 为您的编译后动态库所在目录
    META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为 http://{主节点 ip}:31182>
    curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -d '{"name":"0-yr-mycpp","runtime":"cpp11","kind":"yrlib","cpu":500,"memory":500,"timeout":60,"storageType":"local","codePath":"/opt/mycode/python-invoke-cpp"}'
    ```

4. 运行 Python 主程序测试

    在 `python-invoke-cpp` 目录下运行主程序

    ```bash
    python main.py
    # 输出 4 和 1
    ```

(cross-language-python-invoke-java)=

## Python 程序中调用 Java 无状态和有状态函数

我们在 Python 主程序中使用 [yr.cpp_function](../api/distributed_programming/zh_cn/Python/yr.cpp_function.rst) 和 [yr.cpp_instance_class](../api/distributed_programming/zh_cn/Python/yr.cpp_instance_class.rst) API 调用 Java 无状态和有状态函数。

1. 准备示例工程

    在 openYuanrong 集群所有节点新建目录例如：`/opt/mycode/python-invoke-java`，用于存放 Java 无状态函数和有状态函数代码包。

    新建工程目录 `python-invoke-java` 及文件如下：
   - main.py：python 主程序，调用 Java 无状态函数 `square` 和有状态函数 `Counter`。
   - pom.xml：maven 配置文件，用于编译构建 Java 工程。
   - Square.java：定义无状态函数 `square`。
   - Counter.java：定义有状态函数 `Counter`。

    ```bash
    python-invoke-java
    ├── main.py
    ├── pom.xml
    └── src
        └── main
            └── java
                └── com
                    └── yuanrong
                        └── demo
                            ├── Counter.java
                            └── Square.java
    ```

    :::{dropdown} main.py 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```python
    import yr

    yr.init()

    # 我们将在后续步骤中注册java函数，生成如下URN
    java_function_urn = "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-myjava:$latest"

    # 调用java无状态函数square
    java_function = yr.java_function("com.yuanrong.demo.Square", "square", java_function_urn)
    res = java_function.invoke(2)
    print(yr.get(res))


    # 调用java有状态函数Counter的Add和Get方法
    java_instance = yr.java_instance_class("com.yuanrong.demo.Counter", java_function_urn).invoke(1)
    res = java_instance.add.invoke()
    yr.get(res)
    res = java_instance.get.invoke()
    print(yr.get(res))

    # 销毁函数实例
    java_instance.terminate()
    yr.finalize()
    ```

    :::
    :::{dropdown} pom.xml 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```xml
    <?xml version="1.0" encoding="UTF-8"?>

    <project xmlns="http://maven.apache.org/POM/4.0.0"
            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
        <modelVersion>4.0.0</modelVersion>

        <groupId>com.yuanrong.demo</groupId>
        <artifactId>calculator</artifactId>
        <version>1.0.0</version>

        <properties>
            <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
            <maven.compiler.source>1.8</maven.compiler.source>
            <maven.compiler.target>1.8</maven.compiler.target>
        </properties>

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

    :::
    :::{dropdown} Square.java 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```java
    package com.yuanrong.demo;

    public class Square {
        public static int square(int x) {
            return x * x;
        }
    }
    ```

    :::
    :::{dropdown} Counter.java 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```java
    package com.yuanrong.demo;

    public class Counter {
        private int count = 0;

        public Counter(int count) {
            this.count = count;
        }

        public void add() {
            this.count += 1;
        }

        public int get() {
            return this.count;
        }
    }
    ```

    :::

2. 编译构建 Java 工程

    在 `python-invoke-java` 目录下，执行如下命令构建：

    ```bash
    mvn clean package
    ```

    成功构建将在 `python-invoke-java/target` 目录下生成 `calculator-1.0.0.jar` 文件，拷贝文件到所有节点您创建的代码包目录例如：`/opt/mycode/python-invoke-java`。

3. 注册 Java 函数为 openYuanrong 函数

    注册名为 `0-yr-myjava` 的 openYuanrong 函数，对应函数 URN 为 `sn:cn:yrk:12345678901234561234567890123456:function:0-yr-myjava:$latest`。
    
    使用 curl 工具调用 [注册函数 API](../api/function_service/register_function.md) 动态完成注册。

     ```bash
    # 替换 /opt/mycode/python-invoke-java 为您的代码包目录
    META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为 http://{主节点 ip}:31182>
    curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -d '{"name":"0-yr-myjava","runtime":"java1.8","kind":"yrlib","cpu":500,"memory":500,"timeout":60,"storageType":"local","codePath":"/opt/mycode/python-invoke-java"}'
    ```

4. 运行 Python 主程序测试

    在 `python-invoke-java` 目录下运行主程序

    ```bash
    python main.py
    # 输出 4 和 2
    ```

(cross-language-cpp-invoke-python)=

## C++ 程序中调用 Python 无状态和有状态函数

我们在 C++ 主程序中使用 [PyFunction](../api/distributed_programming/zh_cn/Cpp/PyFunction.rst) 和 [PyInstanceClass::FactoryCreate](../api/distributed_programming/zh_cn/Cpp/PyInstanceClass-FactoryCreate.rst) API 调用 Python 无状态和有状态函数。

1. 准备示例工程

    在 openYuanrong 集群所有节点新建目录例如：`/opt/mycode/cpp-invoke-python`，用于存放 python 无状态函数和有状态函数代码包。

    新建工程目录 `cpp-invoke-python` 及文件如下：
   - calculator.py：定义无状态函数 `square` 和有状态函数 `Counter`。
   - main.cpp：C++ 主程序，调用 python 无状态函数 `square` 和有状态函数 `Counter`。
   - CMakeLists.txt：CMake 的配置文件，用于编译构建 C++ 工程。
   - build：空目录，用于存放 C++ 编译构建中生成的文件。

    ```bash
    cpp-invoke-python
    ├── calculator.py
    ├── main.cpp
    ├── CMakeLists.txt
    └── build
    ```

    :::{dropdown} calculator.py 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```python
    def square(x):
        return x * x

    class Counter:
        def __init__(self, count):
            self.count = count

        def add(self):
            self.count += 1

        def get(self):
            return self.count
    ```

    :::
    :::{dropdown} main.cpp 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```cpp
    #include <iostream>
    #include "yr/yr.h"

    int main(int argc, char **argv)
    {
        // 我们将在后续步骤中注册python函数，生成如下URN
        std::string pyFunctionUrn = "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-mypython:$latest";

        YR::Config conf;
        YR::Init(conf, argc, argv);

        // 调用python无状态函数square
        auto resFutureSquare = YR::PyFunction<int>("calculator", "square").SetUrn(pyFunctionUrn).Invoke(2);
        auto resSquare = *YR::Get(resFutureSquare);
        std::cout << resSquare << std::endl;

        // 调用python有状态函数Counter的add和get方法
        auto instanceHandler = YR::PyInstanceClass::FactoryCreate("calculator", "Counter");
        auto instance = YR::Instance(instanceHandler).SetUrn(pyFunctionUrn).Invoke(0);
        auto resFutureAdd = instance.PyFunction<void>("add").Invoke();
        YR::Wait(resFutureAdd);

        auto resFutureGet = instance.PyFunction<int>("get").Invoke();
        auto resGet = *YR::Get(resFutureGet);
        std::cout << resGet << std::endl;

        // 销毁函数实例
        instance.Terminate();
        YR::Finalize();
        return 0;
    }
    ```

    :::
    :::{dropdown} CMakeLists.txt 文件内容，**须对应修改 YR_INSTALL_PATH 为您的 openYuanrong 安装路径**
    :chevron: down-up
    :icon: chevron-down

    ```cmake
    cmake_minimum_required(VERSION 3.16.1)
    project(main CXX)

    set(CMAKE_CXX_STANDARD 17)
    set(SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR})
    set(BINARY_DIR ${SOURCE_DIR}/build)

    # 替换 YR_INSTALL_PATH 的值为 openYuanrong 安装路径，可通过 yr version 命令查看
    set(YR_INSTALL_PATH "/usr/local/lib/python3.9/site-packages/yr/inner")
    link_directories(${YR_INSTALL_PATH}/runtime/sdk/cpp/lib)
    include_directories(
        ${YR_INSTALL_PATH}/runtime/sdk/cpp/include/faas
        ${YR_INSTALL_PATH}/runtime/sdk/cpp/include
    )

    add_executable(main main.cpp)
    target_link_libraries(main yr-api)
    set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${BINARY_DIR})
    ```

2. 注册 python 函数为 openYuanrong 函数

    先拷贝 `calculator.py` 文件到所有节点您创建的代码包目录下。再注册名为 `0-yr-mypython` 的 openYuanrong 函数，对应函数 URN 为 `sn:cn:yrk:12345678901234561234567890123456:function:0-yr-mypython:$latest`。
    
    使用 curl 工具调用 [注册函数 API](../api/function_service/register_function.md) 动态完成注册。
    
    ```bash
    # 替换 /opt/mycode/cpp-invoke-python 为您的代码包目录
    META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为 http://{主节点 ip}:31182>
    curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -d '{"name":"0-yr-mypython","runtime":"python3.9","kind":"yrlib","cpu":500,"memory":500,"timeout":60,"storageType":"local","codePath":"/opt/mycode/cpp-invoke-python"}'
    ```

3. 运行 C++ 主程序测试

    在 `cpp-invoke-python/build` 目录下，执行如下命令构建：

    ```bash
    cmake ..
    make
    ```

    成功构建将在该目录下生成可执行文件 `main` ，执行如下命令运行主程序。

    ```bash
    ./main
    # 输出 4 和 1
    ```

(cross-language-cpp-invoke-java)=

## C++ 程序中调用 Java 无状态和有状态函数

我们在 C++ 主程序中使用 [PyFunction](../api/distributed_programming/zh_cn/Cpp/PyFunction.rst) 和 [PyInstanceClass::FactoryCreate](../api/distributed_programming/zh_cn/Cpp/PyInstanceClass-FactoryCreate.rst) API 调用 Java 无状态和有状态函数。

1. 准备示例工程

    在 openYuanrong 集群所有节点新建目录例如：`/opt/mycode/cpp-invoke-java`，用于存放 Java 无状态函数和有状态函数代码包。
    
    新建工程目录 `cpp-invoke-java` 及文件如下：
   - main.cpp：C++ 主程序，调用 Java 无状态函数 `square` 和有状态函数 `Counter`。
   - CMakeLists.txt：CMake 的配置文件，用于编译构建 C++ 工程。
   - build：空目录，用于存放 C++ 编译构建中生成的文件。
   - pom.xml：maven 配置文件，用于编译构建 Java 工程。
   - Counter.java：定义 Java 有状态函数 `Counter`。
   - Square.java: 定义 Java 无状态函数 `square`。

    ```bash
    cpp-invoke-java
    ├── cpp-project
    │   ├── main.cpp
    │   ├── CMakeLists.txt
    │   └── build
    └── java-project
        ├── pom.xml
        └── src
            └── main
                └── java
                    └── com
                        └── yuanrong
                            └── demo
                                ├── Counter.java
                                └── Square.java
    ```

    :::{dropdown} main.cpp 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```cpp
    #include <iostream>
    #include "yr/yr.h"

    int main(int argc, char **argv)
    {
        // 我们将在后续步骤中注册java函数，生成如下URN
        std::string javaFunctionUrn = "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-myjava:$latest";

        YR::Config conf;
        YR::Init(conf, argc, argv);

        // 调用java无状态函数square
        auto resFutureSquare = YR::JavaFunction<int>("com.yuanrong.demo.Square", "square").SetUrn(javaFunctionUrn).Invoke(2);
        auto resSquare = *YR::Get(resFutureSquare);
        std::cout << resSquare << std::endl;

        // 调用java有状态函数Counter的add和get方法
        auto instanceHandler = YR::JavaInstanceClass::FactoryCreate("com.yuanrong.demo.Counter");
        auto instance = YR::Instance(instanceHandler).SetUrn(javaFunctionUrn).Invoke(1);
        auto resFutureAdd = instance.JavaFunction<void>("add").Invoke();
        // YR::Get(resFutureAdd);

        auto resFutureGet = instance.JavaFunction<int>("get").Invoke();
        auto resGet = *YR::Get(resFutureGet);
        std::cout << resGet << std::endl;

        // 销毁函数实例
        instance.Terminate();
        YR::Finalize();
    }
    ```

    :::
    :::{dropdown} CMakeLists.txt 文件内容，**须对应修改 YR_INSTALL_PATH 为您的 openYuanrong 安装路径**
    :chevron: down-up
    :icon: chevron-down

    ```cmake
    cmake_minimum_required(VERSION 3.16.1)
    project(main CXX)

    set(CMAKE_CXX_STANDARD 17)
    set(SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR})
    set(BINARY_DIR ${SOURCE_DIR}/build)

    # 替换 YR_INSTALL_PATH 的值为 openYuanrong 安装路径，可通过 yr version 命令查看
    set(YR_INSTALL_PATH "/usr/local/lib/python3.9/site-packages/yr/inner")
    link_directories(${YR_INSTALL_PATH}/runtime/sdk/cpp/lib)
    include_directories(
        ${YR_INSTALL_PATH}/runtime/sdk/cpp/include/faas
        ${YR_INSTALL_PATH}/runtime/sdk/cpp/include
    )

    add_executable(main main.cpp)
    target_link_libraries(main yr-api)
    set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${BINARY_DIR})
    ```

    :::
    :::{dropdown} pom.xml 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```xml
    <?xml version="1.0" encoding="UTF-8"?>

    <project xmlns="http://maven.apache.org/POM/4.0.0"
            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
        <modelVersion>4.0.0</modelVersion>

        <groupId>com.yuanrong.demo</groupId>
        <artifactId>calculator</artifactId>
        <version>1.0.0</version>

        <properties>
            <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
            <maven.compiler.source>1.8</maven.compiler.source>
            <maven.compiler.target>1.8</maven.compiler.target>
        </properties>

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

    :::
    :::{dropdown} Square.java 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```java
    package com.yuanrong.demo;

    public class Square {
        public static int square(int x) {
            return x * x;
        }
    }
    ```

    :::
    :::{dropdown} Counter.java 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```java
    package com.yuanrong.demo;

    public class Counter {
        private int count = 0;

        public Counter(int count) {
            this.count = count;
        }

        public void add() {
            this.count += 1;
        }

        public int get() {
            return this.count;
        }
    }
    ```

    :::

2. 编译构建 Java 工程

    在 `cpp-invoke-java/java-project` 目录下，执行如下命令构建：

    ```bash
    mvn clean package
    ```

    成功构建将在 `cpp-invoke-java/java-project/target` 目录下生成 `calculator-1.0.0.jar` 文件，拷贝文件到所有节点您创建的代码包目录例如：`/opt/mycode/cpp-invoke-java`。

3. 注册 Java 函数为 openYuanrong 函数

    注册名为 `0-yr-myjava` 的 openYuanrong 函数，对应函数 URN 为 `sn:cn:yrk:12345678901234561234567890123456:function:0-yr-myjava:$latest`。
    
    使用 curl 工具调用 [注册函数 API](../api/function_service/register_function.md) 动态完成注册。
    
    ```bash
    # 替换 /opt/mycode/cpp-invoke-java 为您的代码包目录
    META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为 http://{主节点 ip}:31182>
    curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -d '{"name":"0-yr-myjava","runtime":"java1.8","kind":"yrlib","cpu":500,"memory":500,"timeout":60,"storageType":"local","codePath":"/opt/mycode/cpp-invoke-java"}'
    ```

4. 运行 C++ 主程序测试

    在 `cpp-invoke-python/cpp-project/build` 目录下，执行如下命令构建。

    ```bash
    cmake ..
    make
    ```

    成功构建将在该目录下生成可执行文件 `main` ，执行如下命令运行主程序。

    ```bash
    ./main
    # 输出 4 和 2
    ```

(cross-language-java-invoke-cpp)=

## Java 程序中调用 C++ 无状态和有状态函数

我们将 C++ 无状态与有状态函数编译成动态库，在 Java 主程序中使用 [CppFunction](../api/distributed_programming/zh_cn/Java/cpp-function.md) 和 [CppInstanceClass](../api/distributed_programming/zh_cn/Java/cpp_instance_class.md) API 调用。

1. 准备示例工程

    在 openYuanrong 集群所有节点新建目录例如：`/opt/mycode/java-invoke-cpp`，用于存放 C++ 无状态函数和有状态函数代码包。

    新建工程目录 `java-invoke-cpp` 及文件如下：
   - calculator.cpp：定义 C++ 无状态函数 `Square` 和有状态函数 `Counter`。
   - CMakeLists.txt：CMake 的配置文件，用于编译构建 C++ 工程。
   - build：空目录，用于存放 C++ 编译构建生成的文件。
   - pom.xml：maven 配置文件，用于编译构建 Java 工程。
   - Main.java：java 主程序，调用 C++ 无状态函数 `Square` 和有状态函数 `Counter`。

    ```bash
    java-invoke-cpp
    ├── cpp-project
    │   ├── calculator.cpp
    │   ├── CMakeLists.txt
    │   └── build
    └── java-project
        ├── pom.xml
        └── src
            └── main
                └── java
                    └── com
                        └── yuanrong
                            └── demo
                                └── Main.java
    ```

    :::{dropdown} calculator.cpp 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```cpp
    #include "yr/yr.h"
    int Square(int x)
    {
        return x * x;
    }

    // 定义无状态函数 Square
    YR_INVOKE(Square)

    class Counter {
    public:
        Counter() {}
        Counter(int init) { count = init; }

        static Counter *FactoryCreate(int init) {
            return new Counter(init);
        }

        void Add() {
            count += 1;
        }

        int Get() {
            return count;
        }

        YR_STATE(count);

    public:
        int count;
    };

    // 定义有状态函数 Counter
    YR_INVOKE(Counter::FactoryCreate, &Counter::Add, &Counter::Get);
    ```

    :::
    :::{dropdown} CMakeLists.txt 文件内容，**须对应修改 YR_INSTALL_PATH 为您的 openYuanrong 安装路径**
    :chevron: down-up
    :icon: chevron-down

    ```cmake
    cmake_minimum_required(VERSION 3.16.1)
    project(calculator CXX)

    set(CMAKE_CXX_STANDARD 17)
    set(SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR})
    set(BINARY_DIR ${SOURCE_DIR}/build)
    set(BUILD_SHARED_LIBS ON)

    # 替换 YR_INSTALL_PATH 的值为 openYuanrong 安装路径，可通过 yr version 命令查看
    set(YR_INSTALL_PATH "/usr/local/lib/python3.9/site-packages/yr/inner")
    link_directories(${YR_INSTALL_PATH}/runtime/sdk/cpp/lib)
    include_directories(
        ${YR_INSTALL_PATH}/runtime/sdk/cpp/include
    )

    add_library(calculator SHARED calculator.cpp)
    set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${BINARY_DIR})
    ```

    :::
    :::{dropdown} Main.java 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```java
    package com.yuanrong.demo;

    import org.yuanrong.api.YR;
    import org.yuanrong.runtime.client.ObjectRef;
    import org.yuanrong.call.CppInstanceHandler;
    import org.yuanrong.function.CppFunction;
    import org.yuanrong.function.CppInstanceMethod;
    import org.yuanrong.function.CppInstanceClass;

    public class Main {
        public static void main(String[] args) throws Exception {
            YR.init();
            // 我们将在后续步骤中注册C++函数，生成如下URN
            String cppFunctionUrn = "sn:cn:yrk:12345678901234561234567890123456:function:0-yr-mycpp:$latest";

            try {
                // 调用无状态函数Square
                ObjectRef ref = YR.function(CppFunction.of("Square", int.class))
                    .setUrn(cppFunctionUrn)
                    .invoke(2);
                int res = (int) YR.get(ref, 30);
                System.out.println(res);

                // 调用有状态函数Counter的Add和Get方法
                CppInstanceHandler cppInstance = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
                    .setUrn(cppFunctionUrn)
                    .invoke(1);
                ObjectRef addRef = cppInstance.function(CppInstanceMethod.of("Add", void.class)).invoke();
                YR.get(addRef, 30);

                ObjectRef getRef = cppInstance.function(CppInstanceMethod.of("Get", int.class)).invoke();
                res = (int) YR.get(getRef, 30);

                System.out.println(res);
                cppInstance.terminate();
            } finally {
                YR.Finalize();
            }
        }
    }
    ```

    :::
    :::{dropdown} pom.xml 文件内容，**请参考 [安装 Java SDK](install-yuanrong-java-sdk) 并配置依赖 distributed-actortask-sdk**
    :chevron: down-up
    :icon: chevron-down

    ```xml
    <?xml version="1.0" encoding="UTF-8"?>
    
    <project xmlns="http://maven.apache.org/POM/4.0.0"
            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
        <modelVersion>4.0.0</modelVersion>
    
        <groupId>com.yuanrong.demo</groupId>
        <artifactId>yrdemo</artifactId>
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
                <version>9.9.9</version>
            </dependency>
            <dependency>
                <!-- 修改版本号为您实际使用版本 -->
                <groupId>org.yuanrong</groupId>
                <artifactId>faas-function-sdk</artifactId>
                <version>9.9.9</version>
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
                                <mainClass>com.yuanrong.demo.Main</mainClass>
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

    :::

2. 编译构建 C++ 函数为动态库

    在 `java-invoke-cpp/cpp-project/build` 目录下，执行如下命令构建：

    ```bash
    cmake ..
    make
    ```

    成功构建将在该目录下生成动态库文件 `libcalculator.so` ，拷贝文件到所有节点您创建的代码包目录例如：`/opt/mycode/java-invoke-cpp`。

3. 注册 C++ 函数为 openYuanrong 函数

    注册名为 `0-yr-mycpp` 的 openYuanrong 函数，对应函数 URN 为 `sn:cn:yrk:12345678901234561234567890123456:function:0-yr-mycpp:$latest`。
    
    使用 curl 工具调用 [注册函数 API](../api/function_service/register_function.md) 动态完成注册。
    
    ```bash
    # 替换 /opt/mycode/java-invoke-cpp 为您的代码包目录
    META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为 http://{主节点 ip}:31182>
    curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -d '{"name":"0-yr-mycpp","runtime":"cpp11","kind":"yrlib","cpu":500,"memory":500,"timeout":60,"storageType":"local","codePath":"/opt/mycode/java-invoke-cpp"}'
    ```

4. 运行 Java 主程序测试

    在 `java-invoke-cpp/java-project` 目录下执行如下命令编译打包：

    ```bash
    mvn clean package
    ```

    成功构建将在 `java-invoke-cpp/java-project/target` 目录下生成 `yrdemo-1.0.0.jar` 文件，执行如下命令运行主程序。

    ```bash
    java -jar yrdemo-1.0.0.jar
    # 输出 4 和 2
    ```
