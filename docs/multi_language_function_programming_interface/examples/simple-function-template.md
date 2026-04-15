# 函数编程简单示例

本节通过一些示例介绍如何使用有状态函数、无状态函数开发应用，您可以基于这些示例作为基础工程快速上手。

:::{note}
运行示例需先[安装 openYuanrong](../../deploy/installation.md) SDK、命令行工具 yr，并已在[主机上](../../deploy/deploy_processes/index.md)部署 openYuanrong。
:::

(example-project-stateless-function)=

## 无状态函数示例工程

:::::{tab-set}

::::{tab-item} Python

1. 准备示例工程

    新建一个工作目录 `python-stateless-function`，包含文件结构如下。其中，example.py 为应用代码。

    ```bash
    python-stateless-function
    └── example.py
    ```

    :::{dropdown} example.py 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```python
    import yr
    
    # Define stateless function
    @yr.invoke
    def say_hello(name):
        return 'hello, ' + name


    # Init only once
    yr.init()
    
    # Asynchronously invoke stateless functions in parallel
    results_ref = [say_hello.invoke('yuanrong') for i in range(3)]
    print(yr.get(results_ref))
    
    # Release environmental resources
    yr.finalize()
    ```
    
    :::

2. 运行程序

    在 `python-stateless-function` 目录下运行程序，输出如下：

    ```bash
    python example.py
    # ['hello, yuanrong', 'hello, yuanrong', 'hello, yuanrong']
    ```

::::

::::{tab-item} C++

:::{note}
示例运行依赖 make、g++ 及 cmake。
:::

1. 准备示例工程

    新建一个工作目录 `cpp-stateless-function`，包含文件结构如下。其中，build 目录用于存放编译构建中生成的文件，CMakeLists.txt 为 CMake 构建系统使用的配置文件，example.cpp 为应用代码。

    ```bash
    cpp-stateless-function
    ├── build
    ├── CMakeLists.txt
    └── src
        └── example.cpp
    ```

    :::{dropdown} example.cpp 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```cpp
    #include <iostream>
    #include "yr/yr.h"
    
    // Define stateless function
    std::string SayHello(std::string name)
    {
        return "hello, " + name;
    }
    YR_INVOKE(SayHello)


    int main(int argc, char *argv[])
    {
        // Init only once
        YR::Init(YR::Config{}, argc, argv);
    
        // Asynchronously invoke stateless functions in parallel
        std::vector<YR::ObjectRef<std::string>> results_ref;
        for (int i = 0; i < 3; i++) {
            auto result_ref = YR::Function(SayHello).Invoke(std::string("yuanrong"));
            results_ref.emplace_back(result_ref);
        }
    
        for (auto result : YR::Get(results_ref)) {
            std::cout << *result << std::endl;
        }
    
        // Release environmental resources
        YR::Finalize();
        return 0;
    }
    ```
    
    :::
    :::{dropdown} CMakeLists.txt 文件内容，**需对应修改 YR_INSTALL_PATH 为您的 openYuanrong 安装路径**
    :chevron: down-up
    :icon: chevron-down
    
    ```cmake
    cmake_minimum_required(VERSION 3.16.1)
    # 指定项目名称，例如：cpp-stateless-function
    project(cpp-stateless-function LANGUAGES C CXX)
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
    
    # 生成可执行文件 example，修改 example.cpp 为您对应的源码文件
    add_executable(example src/example.cpp)
    target_link_libraries(example yr-api)
    
    # 生成动态库文件 example-dll，修改 example.cpp 为您对应的源码文件
    add_library(example-dll SHARED src/example.cpp)
    target_link_libraries(example-dll yr-api)
    ```
    
    :::

2. 编译构建

    在 `cpp-stateless-function/build` 目录下，执行如下命令构建应用：

    ```bash
    cmake ..
    make
    ```

    成功构建将在该目录下生成 driver 程序 `example` 和包含无状态函数定义的动态库 `libexample-dll.so`。

    :::{note}
    无状态函数可能在 openYuanrong 集群的任意节点执行。因此，运行 driver 程序前，需要在其他节点相同路径下拷贝一份 `libexample-dll.so` 文件。
    :::

3. 运行程序

    在 `cpp-stateless-function/build` 目录下运行程序，输出如下：

    ```bash
    ./example --codePath=$(pwd)
    # hello, yuanrong
    # hello, yuanrong
    # hello, yuanrong
    ```

::::
::::{tab-item} Java

:::{note}
示例运行依赖 java 8 和 maven。
:::

1. 准备示例工程

    新建一个工作目录 `java-stateless-function`，包含文件结构如下。其中，pom.xml 为 maven 配置文件，Greeter.java 和 Main.java 为应用代码。

    ```bash
    java-stateless-function
    ├── pom.xml
    └── src
        └── main
            └── java
                └── com
                    └── yuanrong
                        └── example
                            ├── Greeter.java
                            └── Main.java
    ```

    :::{dropdown} pom.xml 文件内容，**请参考[安装 Java SDK](install-yuanrong-java-sdk) 并配置依赖 yr-api-sdk**
    :chevron: down-up
    :icon: chevron-down

    ```xml
    <?xml version="1.0" encoding="UTF-8"?>

    <project xmlns="http://maven.apache.org/POM/4.0.0"
            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
        <modelVersion>4.0.0</modelVersion>

        <groupId>org.yuanrong.example</groupId>
        <artifactId>example</artifactId>
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
                                <mainClass>org.yuanrong.example.Main</mainClass>
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
    :::{dropdown} Greeter.java 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```java
    package org.yuanrong.example;

    public class Greeter {
        // Define stateless function
        public static String sayHello(String name) {
            return "hello, " + name;
        }
    }
    ```

    :::
    :::
    :::{dropdown} Main.java 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```java
    package org.yuanrong.example;

    import org.yuanrong.Config;
    import org.yuanrong.api.YR;
    import org.yuanrong.runtime.client.ObjectRef;
    import org.yuanrong.exception.YRException;

    import java.util.HashMap;
    import java.util.ArrayList;
    import java.util.List;

    public class Main {
        public static void main(String[] args) throws YRException {
            // Init only once
            YR.init(new Config());

            // Asynchronously invoke stateless functions in parallel
            List<ObjectRef> resultsRef = new ArrayList<>();
            for (int i = 0; i < 3; i++) {
                ObjectRef resultRef = YR.function(Greeter::sayHello).invoke("yuanrong");
                resultsRef.add(resultRef);
            }

            for (ObjectRef resultRef : resultsRef) {
                System.out.println(YR.get(resultRef, 30));
            }

            // Release environmental resources
            YR.Finalize();
        }
    }
    ```

    :::

2. 编译构建

    在 `java-stateless-function` 目录下，执行如下命令构建应用：

    ```bash
    mvn clean package
    ```

    成功构建将在 `java-stateless-function/target` 目录下生成 jar 包 `example-1.0.0.jar`。

    :::{note}
    无状态函数可能在 openYuanrong 集群的任意节点执行。因此，运行程序前，需要在其他节点相同路径下拷贝一份 `example-1.0.0.jar` 文件。
    :::

3. 运行程序

    在 `java-stateless-function` 目录下运行程序，输出如下：

    ```bash
    java -Dyr.codePath=$(pwd)/target -cp target/example-1.0.0.jar org.yuanrong.example.Main
    # hello, yuanrong
    # hello, yuanrong
    # hello, yuanrong
    ```

::::
:::::

(example-project-stateful-function)=

## 有状态函数示例工程

:::::{tab-set}
::::{tab-item} Python

1. 准备示例工程

    新建一个工作目录 `python-stateful-function`，包含文件结构如下。其中，example.py 为应用代码。

    ```bash
    python-stateful-function
    └── example.py
    ```

    :::{dropdown} example.py 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```python
    # example.py
    import yr
    
    # Define stateful function
    @yr.instance
    class Object:
        def __init__(self):
            self.value = 0
    
        def save(self, value):
            self.value = value
    
        def get(self):
            return self.value


    # Init only once
    yr.init()
    
    # Create three stateful function instances
    objs = [Object.invoke() for i in range(3)]
    
    # Asynchronously invoke stateful functions in parallel
    [obj.save.invoke(9) for obj in objs]
    results_ref = [obj.get.invoke() for obj in objs]
    print(yr.get(results_ref))
    
    # Destroy stateful function instance
    [obj.terminate() for obj in objs]
    
    # Release environmental resources
    yr.finalize()
    ```
    
    :::

2. 运行程序

    在 `python-stateful-function` 目录下运行程序，输出如下：

    ```bash
    python example.py
    # [9, 9, 9]
    ```

::::
::::{tab-item} C++

:::{note}
示例运行依赖 make、g++ 及 cmake。
:::

1. 准备示例工程

    新建一个工作目录 `cpp-stateful-function`，包含文件结构如下。其中，build 目录用于存放编译构建中生成的文件，CMakeLists.txt 为 CMake 构建系统使用的配置文件，example.cpp 为应用代码。

    ```bash
    cpp-stateless-function
    ├── build
    ├── CMakeLists.txt
    └── src
        └── example.cpp
    ```

    :::{dropdown} example.cpp 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```cpp
    #include <iostream>
    #include "yr/yr.h"

    // Define stateful function
    class Object {
    public:
        Object() {}
        Object(int value) { this->value = value; }

        static Object *FactoryCreate(int value) {
            return new Object(value);
        }

        void Save(int value) {
            this->value = value;
        }

        int Get() {
            return value;
        }

        YR_STATE(value);

    private:
        int value;
    };
    YR_INVOKE(Object::FactoryCreate, &Object::Save, &Object::Get);

    int main(int argc, char *argv[])
    {
        // Init only once
        YR::Init(YR::Config{}, argc, argv);

        // Create three stateful function instances
        std::vector<YR::NamedInstance<Object>> objects;
        for (int i = 0; i < 3; i++) {
            auto obj = YR::Instance(Object::FactoryCreate).Invoke(0);
            objects.emplace_back(obj);
        }

        // Asynchronously invoke stateful functions in parallel
        std::vector<YR::ObjectRef<int>> results_ref;
        for (auto obj : objects) {
            obj.Function(&Object::Save).Invoke(9);
            auto result_ref = obj.Function(&Object::Get).Invoke();
            results_ref.emplace_back(result_ref);
        }

        for (auto result : YR::Get(results_ref)) {
            std::cout << *result << std::endl;
        }

        // Destroy stateful function instance
        for (auto obj : objects) {
            obj.Terminate();
        }

        // Release environmental resources
        YR::Finalize();
        return 0;
    }
    ```

    :::
    :::{dropdown} CMakeLists.txt 文件内容，**需对应修改 YR_INSTALL_PATH 为您的 openYuanrong 安装路径**
    :chevron: down-up
    :icon: chevron-down

    ```cmake
    cmake_minimum_required(VERSION 3.16.1)
    # 指定项目名称，例如：cpp-stateful-function
    project(cpp-stateful-function LANGUAGES C CXX)
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

    # 生成可执行文件 example，修改 example.cpp 为您对应的源码文件
    add_executable(example src/example.cpp)
    target_link_libraries(example yr-api)

    # 生成动态库文件 example-dll，修改 example.cpp 为您对应的源码文件
    add_library(example-dll SHARED src/example.cpp)
    target_link_libraries(example-dll yr-api)
    ```

    :::

2. 编译构建

    在 `cpp-stateful-function/build` 目录下，执行如下命令构建应用：

    ```bash
    cmake ..
    make
    ```

    成功构建将在该目录下生成 driver 程序 `example` 和包含有状态函数定义的动态库 `libexample-dll.so`。

    :::{note}
    有状态函数可能在 openYuanrong 集群的任意节点执行。因此，运行 driver 程序前，需要在其他节点相同路径下拷贝一份 `libexample-dll.so` 文件。
    :::

3. 运行程序

    在 `cpp-stateful-function/build` 目录下运行程序，输出如下：

    ```bash
    ./example --codePath=$(pwd)
    # 9
    # 9
    # 9
    ```

::::
::::{tab-item} Java

:::{note}
示例运行依赖 java 8 和 maven。
:::

1. 准备示例工程

    新建一个工作目录 `java-stateful-function`，包含文件结构如下。其中，pom.xml 为 maven 配置文件，Object.java 和 Main.java 为应用代码。

    ```bash
    java-stateless-function
    ├── pom.xml
    └── src
        └── main
            └── java
                └── com
                    └── yuanrong
                        └── example
                            ├── Object.java
                            └── Main.java
    ```

    :::{dropdown} pom.xml 文件内容，**请参考[安装 Java SDK](install-yuanrong-java-sdk) 并配置依赖 yr-api-sdk**
    :chevron: down-up
    :icon: chevron-down

    ```xml
    <?xml version="1.0" encoding="UTF-8"?>

    <project xmlns="http://maven.apache.org/POM/4.0.0"
            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
        <modelVersion>4.0.0</modelVersion>

        <groupId>org.yuanrong.example</groupId>
        <artifactId>example</artifactId>
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
                                <mainClass>org.yuanrong.example.Main</mainClass>
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
    :::{dropdown} Object.java 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```java
    package org.yuanrong.example;

    // Define stateful function
    public class Object {
        private int value = 0;

        public Object(int value) {
            this.value = value;
        }

        public void save(int value) {
            this.value = value;
        }

        public int get() {
            return this.value;
        }
    }
    ```

    :::
    :::
    :::{dropdown} Main.java 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```java
    package org.yuanrong.example;

    import org.yuanrong.Config;
    import org.yuanrong.api.YR;
    import org.yuanrong.runtime.client.ObjectRef;
    import org.yuanrong.call.InstanceHandler;
    import org.yuanrong.exception.YRException;

    import java.util.HashMap;
    import java.util.ArrayList;
    import java.util.List;

    public class Main {
        public static void main(String[] args) throws YRException {
            // Init only once
            YR.init(new Config());

            // Create three stateful function instances
            List<InstanceHandler> objects = new ArrayList<>();
            for (int i = 0; i < 3; i++) {
                InstanceHandler obj = YR.instance(Object::new).invoke(0);
                objects.add(obj);
            }

            // Asynchronously invoke stateful functions in parallel
            List<ObjectRef> resultsRef = new ArrayList<>();
            for (InstanceHandler obj : objects) {
                obj.function(Object::save).invoke(9);
                ObjectRef resultRef = obj.function(Object::get).invoke();
                resultsRef.add(resultRef);
            }

            for (ObjectRef resultRef : resultsRef) {
                System.out.println(YR.get(resultRef, 30));
            }

            // Destroy stateful function instance
            for (InstanceHandler obj : objects) {
                obj.terminate();
            }

            // Release environmental resources
            YR.Finalize();
        }
    }
    ```

    :::

2. 编译构建

    在 `java-stateful-function` 目录下，执行如下命令构建应用：

    ```bash
    mvn clean package
    ```

    成功构建将在 `java-stateful-function/target` 目录下生成 jar 包 `example-1.0.0.jar`。

    :::{note}
    无状态函数可能在 openYuanrong 集群的任意节点执行。因此，运行程序前，需要在其他节点相同路径下拷贝一份 `example-1.0.0.jar` 文件。
    :::

3. 运行程序

    在 `java-stateful-function` 目录下运行程序，输出如下：

    ```bash
    java -Dyr.codePath=$(pwd)/target -cp target/example-1.0.0.jar org.yuanrong.example.Main
    # 9
    # 9
    # 9
    ```

::::
:::::

(example-project-function-service)=

## 函数服务示例工程

### 在主机集群中运行函数服务

参考[配置支持函数服务](deploy-processes-config-support-faas)确保您部署的 openYuanrong 集群支持运行函数服务。在所有节点创建相同的代码包目录，例如 `/opt/mycode/service`，用于存放构建生成的可执行函数代码。

:::::{tab-set}
::::{tab-item} Python

1. 准备示例代码

    新建如下内容的文件 `demo.py`，拷贝文件到所有节点的代码包目录下。

    ```python
    import datetime

    # 函数执行入口，每次请求都会执行，其中event参数为JSON格式
    def handler(event, context):
        print("received request,event content:", event)

        response = ""
        try:
            name = event.get("name")
            # 获取配置的环境变量，环境变量在注册和更新函数时设置
            show_date = context.getUserData("show_date")
            if show_date is not None:
                response = "hello " + name + ",today is " + datetime.date.today().strftime('%Y-%m-%d')
            else:
                response = "hello " + name
        except Exception as e:
            print(e)
            response = "please enter your name,for example:{'name':'yuanrong'}"

        return response

    # 函数初始化入口，函数实例启动时执行一次
    def init(context):
        print("function instance initialization completed")

    # 函数退出入口，函数实例被销毁时执行一次
    def pre_stop():
        print("function instance is being destroyed")
    ```

2. 函数注册及调用

    使用 curl 工具注册函数，参数含义详见 [API 说明](../api/function_service/register_function.md)：

    ```bash
    # 替换 /opt/mycode/service 为您的代码包目录
    META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为：http://{主节点 IP}:31182>
    curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -d '{"name":"0@myService@python-demo","runtime":"python3.9","handler":"demo.handler","environment":{"show_date":"true"},"extendedHandler":{"initializer":"demo.init","pre_stop":"demo.pre_stop"},"extendedTimeout":{"initializer":30,"pre_stop":10},"kind":"faas","cpu":600,"memory":512,"timeout":60,"storageType":"local","codePath":"/opt/mycode/service"}'
    ```

    结果返回格式如下，记录 `functionVersionUrn` 字段的值用于调用，这里对应 `sn:cn:yrk:default:function:0@myService@python-demo:latest`。

    ```bash
    {"code":0,"message":"SUCCESS","function":{"id":"sn:cn:yrk:default:function:0@myService@python-demo:latest","createTime":"2025-05-29 07:09:34.154 UTC","updateTime":"","functionUrn":"sn:cn:yrk:default:function:0@myService@python-demo","name":"0@myService@python-demo","tenantId":"default","businessId":"yrk","productId":"","reversedConcurrency":0,"description":"","tag":null,"functionVersionUrn":"sn:cn:yrk:default:function:0@myService@python-demo:latest","revisionId":"20250529070934154","codeSize":0,"codeSha256":"","bucketId":"","objectId":"","handler":"demo.handler","layers":null,"cpu":600,"memory":512,"runtime":"python3.9","timeout":60,"versionNumber":"latest","versionDesc":"latest","environment":{"show_date":"true"},"customResources":null,"statefulFlag":0,"lastModified":"","Published":"2025-05-29 07:09:34.154 UTC","minInstance":0,"maxInstance":100,"concurrentNum":100,"funcLayer":[],"status":"","instanceNum":0,"device":{},"created":""}}
    ```

    使用 curl 工具调用函数，参数含义详见 [API 说明](../api/function_service/function_invocation.md)：

    ```bash
    FRONTEND_ENDPOINT=<frontend 组件的终端节点，默认为：http://{主节点 ip}:8888`>
    FUNCTION_VERSION_URN=<前一步骤记录的 functionVersionUrn>
    curl -H "Content-type: application/json" -X POST -i ${FRONTEND_ENDPOINT}/serverless/v1/functions/${FUNCTION_VERSION_URN}/invocations -d '{"name":"yuanrong"}'
    ```

    结果输出：

    ```bash
    HTTP/1.1 200 OK
    Content-Type: application/json
    X-Billing-Duration: this is billing duration TODO
    X-Inner-Code: 0
    X-Invoke-Summary:
    X-Log-Result: dGhpcyBpcyB1c2VyIGxvZyBUT0RP
    Date: Tue, 20 May 2025 02:03:09 GMT
    Content-Length: 36
    
    "hello yuanrong,today is 2025-05-20"
    ```

::::
::::{tab-item} C++

1. 准备示例工程

    新建一个工作目录 `service-cpp-demo`，包含文件结构如下。其中，build 目录用于存放编译构建中生成的文件。CMakeLists.txt 为 CMake 构建系统使用的配置文件。demo.cpp 为函数代码文件，我们使用了开源 [nlohmann/json](https://github.com/nlohmann/json){target="_blank"} 库解析 JSON。

    ```bash
    service-cpp-demo
    ├── build
    ├── demo.cpp
    └── CMakeLists.txt
    ```

    :::{dropdown} demo.cpp 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```cpp
    #include <string>
    #include <cstdlib>
    #include <ctime>
    #include <nlohmann/json.hpp>

    #include "Runtime.h"
    #include "Function.h"
    #include "yr/yr.h"

    // 函数执行入口，每次请求都会执行，其中event参数要求为JSON格式，返回类型可自定义
    std::string HandleRequest(const std::string &event, Function::Context &context) {
        std::cout << "received request,event content:" << event << std::endl;

        std::string response = "";
        try {
            nlohmann::json jsonData = nlohmann::json::parse(event);
            // 读取 JSON 数据并输出
            std::string name = jsonData["name"];
            response += "hello ";
            response += name;

            std::string showDate = context.GetUserData("show_date");
            if (showDate != "") {
                time_t now = time(0);
                tm *ltm = localtime(&now);

                std::stringstream timeStr;
                timeStr << ltm->tm_year + 1900 << "-";
                    timeStr << ltm->tm_mon + 1 << "-";
                    timeStr << ltm->tm_mday;

                response += ",today is ";
                response += timeStr.str();
            }
        } catch (const std::exception& e) {
            std::cout << "JSON parsing error:" << e.what() << std::endl;
            response = "please enter your name,for example:{'name':'yuanrong'}";
        }
        return response;
    }

    // 函数初始化入口，函数实例启动时执行一次
    void Initializer(Function::Context &context) {
        std::cout << "function instance initialization completed" << std::endl;
        return;
    }

    // 函数退出入口，函数实例被销毁时执行一次
    void PreStop(Function::Context &context) {
        std::cout << "function instance is being destroyed" << std::endl;
        return;
    }

    int main(int argc, char *argv[])
    {
        Function::Runtime rt;
        rt.RegisterHandler(HandleRequest);
        rt.RegisterInitializerFunction(Initializer);
        rt.RegisterPreStopFunction(PreStop);
        rt.Start(argc, argv);
        return 0;
    }
    ```

    :::
    :::{dropdown} CMakeLists.txt 文件内容，**需对应修改 YR_INSTALL_PATH 的值为 openYuanrong 实际安装路径**
    :chevron: down-up
    :icon: chevron-down

    ```cmake
    cmake_minimum_required(VERSION 3.16.1)
    project(demo CXX)

    set(CMAKE_CXX_STANDARD 17)
    set(SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR})
    set(BINARY_DIR ${SOURCE_DIR}/build)

    # 替换 YR_INSTALL_PATH 的值为 openYuanrong 实际安装路径
    set(YR_INSTALL_PATH "/usr/local/lib/python3.9/site-packages/yr")
    link_directories(${YR_INSTALL_PATH}/cpp/lib)
    include_directories(
        ${YR_INSTALL_PATH}/cpp/include/faas
        ${YR_INSTALL_PATH}/cpp/include
    )

    add_executable(demo demo.cpp)
    target_link_libraries(demo functionsdk)
    set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${BINARY_DIR})
    ```

    :::

2. 编译构建

    下载 [nlohmann/json](https://github.com/nlohmann/json){target="_blank"} 例如 json-3.12.0.tar.gz 版本到所有openYuanrong节点，依次执行如下命令安装。

    ```bash
    tar -zxvf json-3.12.0.tar.gz
    cd json-3.12.0
    mkdir build
    cd build
    cmake ..
    make
    make install
    ```

    在 `service-cpp-demo/build` 目录下，执行如下命令构建函数：

    ```bash
    cmake ..
    make
    ```

    成功构建将在该目录下生成二进制文件 `demo` ，拷贝文件到所有节点的代码包目录下。

3. 函数注册及调用

    使用 curl 工具注册函数，参数含义详见 [API 说明](../api/function_service/register_function.md)：

    ```bash
    # 替换 /opt/mycode/service 为您的代码包目录
    META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为：http://{主节点 ip}:31182>
    curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -d '{"name":"0@myService@cpp-demo","runtime":"posix-custom-runtime","handler":"demo","environment":{"show_date":"true"},"kind":"faas","cpu":600,"memory":512,"timeout":60,"storageType":"local","codePath":"/opt/mycode/service"}'
    ```

    结果返回格式如下，记录 `functionVersionUrn` 字段的值用于调用，这里对应 `sn:cn:yrk:default:function:0@myService@cpp-demo:latest`

    ```bash
    {"code":0,"message":"SUCCESS","function":{"id":"sn:cn:yrk:default:function:0@myService@cpp-demo:latest","createTime":"2025-05-20 03:43:19.117 UTC","updateTime":"","functionUrn":"sn:cn:yrk:default:function:0@myService@cpp-demo","name":"0@myService@cpp-demo","tenantId":"default","businessId":"yrk","productId":"","reversedConcurrency":0,"description":"","tag":null,"functionVersionUrn":"sn:cn:yrk:default:function:0@myService@cpp-demo:latest","revisionId":"20250520034319117","codeSize":0,"codeSha256":"","bucketId":"","objectId":"","handler":"demo","layers":null,"cpu":600,"memory":512,"runtime":"posix-custom-runtime","timeout":60,"versionNumber":"latest","versionDesc":"latest","environment":{"show_date":"true"},"customResources":null,"statefulFlag":0,"lastModified":"","Published":"2025-05-20 03:43:19.117 UTC","minInstance":0,"maxInstance":100,"concurrentNum":100,"funcLayer":[],"status":"","instanceNum":0,"device":{},"created":""}}
    ```

    使用 curl 工具调用函数，参数含义详见 [API 说明](../api/function_service/function_invocation.md)：

    ```bash
    FRONTEND_ENDPOINT=<frontend 组件的终端节点，默认为：http://{主节点 ip}:8888>
    FUNCTION_VERSION_URN=<前一步骤记录的 functionVersionUrn>
    curl -H "Content-type: application/json" -X POST -i ${FRONTEND_ENDPOINT}/serverless/v1/functions/${FUNCTION_VERSION_URN}/invocations -d '{"name":"yuanrong"}'
    ```

    结果输出：

    ```bash
    HTTP/1.1 200 OK
    Content-Type: application/json
    X-Billing-Duration: this is billing duration TODO
    X-Inner-Code: 0
    X-Invoke-Summary:
    X-Log-Result: this is user log TODO
    Date: Tue, 20 May 2025 03:43:57 GMT
    Content-Length: 35
    
    "hello yuanrong,today is 2025-5-20"
    ```

::::
::::{tab-item} Java

1. 准备示例工程

    新建一个工作目录 `service-java-demo`，包含文件结构如下。其中，pom.xml 为 maven 配置文件，引入 openYuanrong SDK 等依赖。zip_file.xml 为打包配置。Demo.java 为函数代码文件。

    ```bash
    service-java-demo
    ├── pom.xml
    └── src
        └── main
            ├── assembly
            │   └── zip_file.xml
            └── java
                └── com
                    └── yuanrong
                        └── demo
                            └── Demo.java
    ```

    :::{dropdown} Demo.java 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```java
    package org.yuanrong.demo;

    import com.services.runtime.Context;
    import com.google.gson.JsonObject;
    import java.time.LocalDate;

    public class Demo {
        // 函数执行入口，每次请求都会执行，其中intput参数及函数返回类型可自定义
        public String handler(JsonObject event, Context context) {
            System.out.println("received request,event content:" + event);

            String response = "";
            try {
                String name = event.get("name").getAsString();
                // 获取配置的环境变量，环境变量在注册和更新函数时设置
                String showDate = context.getUserData("show_date");
                if (showDate != null) {
                    response = "hello " + name + ",today is " + LocalDate.now();
                } else {
                    response = "hello " + name;
                }
            } catch(Exception e) {
                e.printStackTrace();
                response = "please enter your name,for example:{'name':'yuanrong'}";
            }
            return response;
        }

        // 函数初始化入口，函数实例启动时执行一次
        public void initializer(Context context) {
            System.out.println("function instance initialization completed");
        }
    }
    ```

    :::
    :::{dropdown} pom.xml 文件内容，**请参考[安装 Java SDK](install-yuanrong-java-sdk) 并配置依赖 faas-function-sdk**
    :chevron: down-up
    :icon: chevron-down

    ```xml
    <?xml version="1.0" encoding="UTF-8"?>
    <project xmlns="http://maven.apache.org/POM/4.0.0"
            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
        <modelVersion>4.0.0</modelVersion>

        <groupId>org.yuanrong.demo</groupId>
        <artifactId>demo</artifactId>
        <version>1.0.0</version>

        <properties>
            <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
            <maven.compiler.source>1.8</maven.compiler.source>
            <maven.compiler.target>1.8</maven.compiler.target>
            <maven.build.timestamp.format>yyyyMMddHHmmss</maven.build.timestamp.format>
            <package.finalName>demo</package.finalName>
            <package.outputDirectory>target</package.outputDirectory>
        </properties>

        <dependencies>
            <dependency>
                <!-- 修改版本号为您实际使用版本 -->
                <groupId>org.yuanrong</groupId>
                <artifactId>faas-function-sdk</artifactId>
                <version>1.0.0</version>
            </dependency>
            <dependency>
                <groupId>com.google.code.gson</groupId>
                <artifactId>gson</artifactId>
                <version>2.9.0</version>
            </dependency>
        </dependencies>

        <build>
            <plugins>
                <!-- 配置如下打包方式 -->
                <plugin>
                    <groupId>org.apache.maven.plugins</groupId>
                    <artifactId>maven-assembly-plugin</artifactId>
                    <version>2.2</version>
                    <configuration>
                        <archive>
                            <manifest>
                                <addDefaultImplementationEntries>true</addDefaultImplementationEntries>
                            </manifest>
                        </archive>
                        <appendAssemblyId>false</appendAssemblyId>
                    </configuration>
                    <executions>
                        <execution>
                            <id>auto-deploy</id>
                            <phase>package</phase>
                            <goals>
                                <goal>single</goal>
                            </goals>
                            <configuration>
                                <descriptors>
                                    <descriptor>src/main/assembly/zip_file.xml</descriptor>
                                </descriptors>
                                <finalName>${package.finalName}</finalName>
                                <outputDirectory>${package.outputDirectory}</outputDirectory>
                                <archiverConfig>
                                    <directoryMode>0700</directoryMode>
                                    <fileMode>0600</fileMode>
                                    <defaultDirectoryMode>0700</defaultDirectoryMode>
                                </archiverConfig>
                            </configuration>
                        </execution>
                    </executions>
                </plugin>
            </plugins>
        </build>
    </project>
    ```

    :::
    :::{dropdown} zip_file.xml 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```xml
    <?xml version="1.0" encoding="UTF-8"?>

    <assembly xmlns="http://maven.apache.org/plugins/maven-assembly-plugin/assembly/1.1.2">

        <id>auto-deploy</id>
        <baseDirectory/>
        <formats>
            <format>zip</format>
        </formats>

        <fileSets>
            <fileSet>
                <directory>src/main/resources/</directory>
                <outputDirectory>config</outputDirectory>
                <includes>
                    <include>**</include>
                </includes>
                <fileMode>0600</fileMode>
                <directoryMode>0700</directoryMode>
            </fileSet>

        </fileSets>

        <dependencySets>
            <dependencySet>
                <outputDirectory>lib</outputDirectory>
                <scope>runtime</scope>
            </dependencySet>
        </dependencySets>
    </assembly>
    ```

    :::

2. 编译构建

    在 `service-java-demo` 目录下，执行如下命令构建：

    ```bash
    mvn clean package
    ```

    成功构建将在 `service-java-demo/target` 目录下生成压缩包 `demo.zip` ，拷贝文件到所有节点的代码包目录下并且使用命令 `unzip demo.zip` 解压。

3. 函数注册及调用

    使用 curl 工具注册函数，参数含义详见 [API 说明](../api/function_service/register_function.md)：

    ```bash
    # 替换 /opt/mycode/service 为您的代码包目录
    META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为：http://{主节点 ip}:31182>
    curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -d '{"name":"0@myService@java-demo","runtime":"java8","handler":"org.yuanrong.demo.Demo::handler","environment":{"show_date":"true"},"extendedHandler":{"initializer":"org.yuanrong.demo.Demo::initializer"},"extendedTimeout":{"initializer":30},"kind":"faas","cpu":600,"memory":512,"timeout":60,"storageType":"local","codePath":"/opt/mycode/service"}'
    ```

    结果返回格式如下，记录 `functionVersionUrn` 字段的值用于调用，这里对应 `sn:cn:yrk:default:function:0@myService@java-demo:latest`

    ```bash
    {"code":0,"message":"SUCCESS","function":{"id":"sn:cn:yrk:default:function:0@myService@java-demo:latest","createTime":"2025-05-20 06:26:42.396 UTC","updateTime":"","functionUrn":"sn:cn:yrk:default:function:0@myService@java-demo","name":"0@myService@java-demo","tenantId":"default","businessId":"yrk","productId":"","reversedConcurrency":0,"description":"","tag":null,"functionVersionUrn":"sn:cn:yrk:default:function:0@myService@java-demo:latest","revisionId":"20250520062642396","codeSize":0,"codeSha256":"","bucketId":"","objectId":"","handler":"org.yuanrong.demo.Demo::handler","layers":null,"cpu":600,"memory":512,"runtime":"java8","timeout":60,"versionNumber":"latest","versionDesc":"latest","environment":{"show_date":"true"},"customResources":null,"statefulFlag":0,"lastModified":"","Published":"2025-05-20 06:26:42.396 UTC","minInstance":0,"maxInstance":100,"concurrentNum":100,"funcLayer":[],"status":"","instanceNum":0,"device":{},"created":""}}
    ```

    使用 curl 工具调用函数，参数含义详见 [API 说明](../api/function_service/function_invocation.md)：

    ```bash
    FRONTEND_ENDPOINT=<frontend 组件的终端节点，默认为：http://{主节点 ip}:8888>
    FUNCTION_VERSION_URN=<前一步骤记录的 functionVersionUrn>
    curl -H "Content-type: application/json" -X POST -i ${FRONTEND_ENDPOINT}/serverless/v1/functions/${FUNCTION_VERSION_URN}/invocations -d '{"name":"yuanrong"}'
    ```

    结果输出：

    ```bash
    HTTP/1.1 200 OK
    Content-Type: application/json
    X-Billing-Duration: this is billing duration TODO
    X-Inner-Code: 0
    X-Invoke-Summary:
    X-Log-Result: dGhpcyBpcyB1c2VyIGxvZyBUT0RP
    Date: Tue, 20 May 2025 06:27:49 GMT
    Content-Length: 36
    
    "hello yuanrong,today is 2025-05-20"
    ```

::::
:::::

(example-project-function-K8s-service)=

### 在 K8s 集群中运行函数服务

K8s 中运行函数服务需要先创建资源池，如果您在部署 openYuanrong 时已经创建，可跳过这个步骤。首先创建 create_pool.json 文件，内容如下:

```json
{
    "pools": [
        {
            "id": "pool-600-512",
            "size": 1,
            "group": "rg1",
            "reuse": false,
            "resources": {
                "limits": {
                    "cpu": "600m",
                    "memory": "512Mi"
                },
                "requests": {
                    "cpu": "600m",
                    "memory": "512Mi"
                }
            }
        }
    ]
}
```

使用 curl 工具创建资源池，参数含义详见[API 说明](../../deploy/deploy_on_k8s/api/create_pod_pool.md)：

```bash
META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为：http://{主节点 IP}:31182>
curl -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/podpools -H 'Content-Type: application/json' -d @create_pool.json
```

返回结果格式如下。
    
```json
{
  "code":0,
  "message": "",
  "result": {
    "failed_pools": null
  }
}
```

:::::{tab-set}
::::{tab-item} Python

1. 准备示例代码

    新建如下内容文件 demo.py。

    ```python
    import datetime
    # 函数执行入口，每次请求都会执行，其中event参数为JSON格式
    def handler(event, context):
        print("received request,event content:", event)

        response = ""
        try:
            name = event.get("name")
            # 获取配置的环境变量，环境变量在注册和更新函数时设置
            show_date = context.getUserData("show_date")
            if show_date is not None:
                response = "hello " + name + ",today is " + datetime.date.today().strftime('%Y-%m-%d')
            else:
                response = "hello " + name
        except Exception as e:
            print(e)
            response = "please enter your name,for example:{'name':'yuanrong'}"
    
        return response

    # 函数初始化入口，函数实例启动时执行一次
    def init(context):
        print("function instance initialization completed")
    
    # 函数退出入口，函数实例被销毁时执行一次
    def pre_stop():
        print("function instance is being destroyed")    
    ```

    打包成demo.zip并且使用[MinIO Client](tools-minio-client)上传代码包到 openYuanrong 集群中的 MinIO 服务 。
        
    ```bash
    zip demo.zip demo.py
    mc mb mys3/demo-bucket
    mc cp ./demo.zip mys3/demo-bucket/demo.zip
    ```

2. 函数注册及调用

    新建如下内容文件 create_func.json。

    ```json
     {
        "name": "0@myService@python-demo",
        "handler": "demo.my_handler",
        "extendedHandler":{"initializer":"demo.init","pre_stop":"demo.pre_stop"},
          "environment":{"show_date":"true"},
          "runtime": "python3.9",
          "cpu": 600,
          "memory": 512,
          "timeout": 30,
          "currentNum": "100",
          "kind" : "faas",
          "storageType": "s3",
          "s3CodePath": {
            "bucketId": "demo-bucket",
            "objectId": "demo.zip",
            "bucketUrl": "http://{Your MinIO Address:30110}"
          }  
     }
    ```
    
    通过 curl 工具注册函数，参数含义详见 [API 说明](../api/function_service/register_function.md)：
        
    ```bash 
    META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为：http://{主节点 IP}:31182>
    curl -X POST -i ${META_SERVCICE_ENDPOINT}/serverless/v1/functions \
            -H 'Content-Type: application/json' -H 'x-storage-type: local' \
            -d @create_func.json
        
    ```
    
    结果格式如下， 记录 ` functionVersionUrn ` 字段的值用于调用， 这里对应 ` sn:cn:yrk:default:function:0@myService@python-demo:latest `
        
    ```json
    {"code":0,"message":"SUCCESS","function":{"id":"sn:cn:yrk:default:function:0@myService@python-demo:latest","createTime":"2026-01-20 01:57:36.938 UTC","updateTime":"","functionUrn":"sn:cn:yrk:default:function:0@myService@python-demo","name":"0@myService@python-demo","tenantId":"default","businessId":"yrk","productId":"","reversedConcurrency":0,"description":"","tag":null,"functionVersionUrn":"sn:cn:yrk:default:function:0@myService@python-demo:latest","revisionId":"20260120015736938","codeSize":0,"extendedHandler":{"initializer":"demo.init","pre_stop":"demo.pre_stop"},"codeSha256":"","bucketId":"demo-bucket","objectId":"demo.zip","handler":"demo.my_handler","layers":null,"cpu":600,"memory":512,"runtime":"python3.9","timeout":30,"versionNumber":"latest","versionDesc":"latest","environment":{"show_date":"true"},"customResources":null,"statefulFlag":0,"lastModified":"","Published":"2016-01-20 01:57:36.936 UTC","minInstance":1,"maxInstance":100,"concurrentNum":100,"funcLayer":[],"status":"","instanceNum":0,"device":{},"created":""}}
    ```

    使用 curl 工具调用函数，参数含义详见 [API 说明](../api/function_service/function_invocation.md)： 
    
    ```bash
    FRONTEND_ENDPOINT=<frontend 组件的终端节点，默认为：http://{主节点 ip}:8888`>
    #设置 functionVersionUrn 环境变量，根据之前创建函数返回的functionVersionUrn
    curl -H "Content-type: application/json" -X POST \
         -i ${FRONTEND_ENDPOINT}/serverless/v1/functions/${functionVersionUrn}/invocations \
         -d {\"name\":\"openyuanrong\"}
    ```
    
    结果输出: "hello yuanrong,today is 2026-01-20"

::::
::::{tab-item} C++

1. 准备示例工程

    新建一个工作目录 `service-cpp-demo`，包含文件结构如下。其中，build 目录用于存放编译构建中生成的文件。CMakeLists.txt 为 CMake 构建系统使用的配置文件。demo.cpp 为函数代码文件，我们使用了开源 [nlohmann/json](https://github.com/nlohmann/json){target="_blank"} 库解析 JSON。

    ```bash
    service-cpp-demo
    ├── build
    ├── demo.cpp
    └── CMakeLists.txt
    ```

    :::{dropdown} demo.cpp 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```cpp
    #include <string>
    #include <cstdlib>
    #include <ctime>
    #include <nlohmann/json.hpp>

    #include "Runtime.h"
    #include "Function.h"
    #include "yr/yr.h"

    // 函数执行入口，每次请求都会执行，其中event参数要求为JSON格式，返回类型可自定义
    std::string HandleRequest(const std::string &event, Function::Context &context) {
        std::cout << "received request,event content:" << event << std::endl;

        std::string response = "";
        try {
            nlohmann::json jsonData = nlohmann::json::parse(event);
            // 读取 JSON 数据并输出
            std::string name = jsonData["name"];
            response += "hello ";
            response += name;

            std::string showDate = context.GetUserData("show_date");
            if (showDate != "") {
                time_t now = time(0);
                tm *ltm = localtime(&now);

                std::stringstream timeStr;
                timeStr << ltm->tm_year + 1900 << "-";
                    timeStr << ltm->tm_mon + 1 << "-";
                    timeStr << ltm->tm_mday;

                response += ",today is ";
                response += timeStr.str();
            }
        } catch (const std::exception& e) {
            std::cout << "JSON parsing error:" << e.what() << std::endl;
            response = "please enter your name,for example:{'name':'yuanrong'}";
        }
        return response;
    }

    // 函数初始化入口，函数实例启动时执行一次
    void Initializer(Function::Context &context) {
        std::cout << "function instance initialization completed" << std::endl;
        return;
    }

    // 函数退出入口，函数实例被销毁时执行一次
    void PreStop(Function::Context &context) {
        std::cout << "function instance is being destroyed" << std::endl;
        return;
    }

    int main(int argc, char *argv[])
    {
        Function::Runtime rt;
        rt.RegisterHandler(HandleRequest);
        rt.RegisterInitializerFunction(Initializer);
        rt.RegisterPreStopFunction(PreStop);
        rt.Start(argc, argv);
        return 0;
    }
    ```

    :::
    :::{dropdown} CMakeLists.txt 文件内容，**需对应修改 YR_INSTALL_PATH 的值为 openYuanrong 实际安装路径**
    :chevron: down-up
    :icon: chevron-down

    ```cmake
    cmake_minimum_required(VERSION 3.16.1)
    project(demo CXX)

    set(CMAKE_CXX_STANDARD 17)
    set(SOURCE_DIR ${CMAKE_CURRENT_SOURCE_DIR})
    set(BINARY_DIR ${SOURCE_DIR}/build)

    # 替换 YR_INSTALL_PATH 的值为 openYuanrong 实际安装路径
    set(YR_INSTALL_PATH "/usr/local/lib/python3.9/site-packages/yr")
    link_directories(${YR_INSTALL_PATH}/cpp/lib)
    include_directories(
        ${YR_INSTALL_PATH}/cpp/include/faas
        ${YR_INSTALL_PATH}/cpp/include
    )

    add_executable(demo demo.cpp)
    target_link_libraries(demo functionsdk)
    set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${BINARY_DIR})
    ```

    :::

2. 编译构建

    下载 [nlohmann/json](https://github.com/nlohmann/json){target="_blank"} 例如 json-3.12.0.tar.gz 版本到所有openYuanrong节点，依次执行如下命令安装。

    ```bash
    tar -zxvf json-3.12.0.tar.gz
    cd json-3.12.0
    mkdir build
    cd build
    cmake ..
    make
    make install
    ```

    在 `service-cpp-demo/build` 目录下，执行如下命令构建函数：

    ```bash
    cmake ..
    make
    ```

    成功构建将在该目录下生成二进制文件 `demo` ，打包成demo.zip 并且使用 [Minio 客户端](../../reference/development-tools.md)上传到MinIO。
        
    ```bash
    zip demo.zip demo
    mc mb mys3/demo-bucket
    mc cp ./demo.zip mys3/demo-bucket/demo.zip
    ```

3. 函数注册及调用

    新建如下内容文件 create_func.json。

    ```json
    {
        "name": "0@myService@cpp-demo",
        "runtime": "posix-custom-runtime",
        "handler": "demo",
        "environment": {
            "show_date": "true"
        },
        "kind": "faas",
        "cpu": 600,
        "memory": 512,
        "timeout": 60,
        "storageType": "s3",
        "s3CodePath": {
            "bucketId": "demo-bucket",
            "objectId": "demo.zip",
            "bucketUrl": "http://{Your MinIO Address:9000}"
        }
    }   
    ```

    使用 curl 工具注册函数，参数含义详见 [API 说明](../api/function_service/register_function.md)：

    ```bash
    # 替换 /opt/mycode/service 为您的代码包目录
    META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为：http://{主节点 ip}:31182>
    curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -H 'Content-Type: application/json' \
         -H 'x-storage-type: local' \
         -d @create_func.json
    ```

    结果返回格式如下，记录 `functionVersionUrn` 字段的值用于调用，这里对应 `sn:cn:yrk:12345678901234561234567890123456:function:0@myService@cpp-demo:latest`

    ```bash
    {"code":0,"message":"SUCCESS","function":{"id":"sn:cn:yrk:12345678901234561234567890123456:function:0@myService@cpp-demo:latest","createTime":"2025-05-20 03:43:19.117 UTC","updateTime":"","functionUrn":"sn:cn:yrk:12345678901234561234567890123456:function:0@myService@cpp-demo","name":"0@myService@cpp-demo","tenantId":"12345678901234561234567890123456","businessId":"yrk","productId":"","reversedConcurrency":0,"description":"","tag":null,"functionVersionUrn":"sn:cn:yrk:12345678901234561234567890123456:function:0@myService@cpp-demo:latest","revisionId":"20250520034319117","codeSize":0,"codeSha256":"","bucketId":"demo-bucket","objectId":"demo.zip","handler":"demo","layers":null,"cpu":600,"memory":512,"runtime":"posix-custom-runtime","timeout":60,"versionNumber":"latest","versionDesc":"latest","environment":{"show_date":"true"},"customResources":null,"statefulFlag":0,"lastModified":"","Published":"2025-05-20 03:43:19.117 UTC","minInstance":0,"maxInstance":100,"concurrentNum":100,"funcLayer":[],"status":"","instanceNum":0,"device":{},"created":""}}
    ```

    使用 curl 工具调用函数，参数含义详见 [API 说明](../api/function_service/function_invocation.md)：

    ```bash
    FRONTEND_ENDPOINT=<frontend 组件的终端节点，默认为：http://{主节点 ip}:8888>
    FUNCTION_VERSION_URN=<前一步骤记录的 functionVersionUrn>
    curl -H "Content-type: application/json" -X POST -i ${FRONTEND_ENDPOINT}/serverless/v1/functions/${FUNCTION_VERSION_URN}/invocations -d '{"name":"yuanrong"}'
    ```

    结果输出：

    ```bash
    HTTP/1.1 200 OK
    Content-Type: application/json
    X-Billing-Duration: this is billing duration TODO
    X-Inner-Code: 0
    X-Invoke-Summary:
    X-Log-Result: this is user log TODO
    Date: Tue, 26 Jan 2026 11:31:57 GMT
    Content-Length: 35
    
    "hello yuanrong,today is 2026-1-26"
    ```

::::

::::{tab-item} Java

1. 准备示例工程

    新建一个工作目录 `service-java-demo`，包含文件结构如下。其中，pom.xml 为 maven 配置文件，引入 openYuanrong SDK 等依赖。zip_file.xml 为打包配置。Demo.java 为函数代码文件。

    ```bash
    service-java-demo
    ├── pom.xml
    └── src
        └── main
            ├── assembly
            │   └── zip_file.xml
            └── java
                └── com
                    └── yuanrong
                        └── demo
                            └── Demo.java
    ```

    :::{dropdown} Demo.java 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```java
    package org.yuanrong.demo;

    import com.services.runtime.Context;
    import com.google.gson.JsonObject;
    import java.time.LocalDate;

    public class Demo {
        // 函数执行入口，每次请求都会执行，其中intput参数及函数返回类型可自定义
        public String handler(JsonObject event, Context context) {
            System.out.println("received request,event content:" + event);

            String response = "";
            try {
                String name = event.get("name").getAsString();
                // 获取配置的环境变量，环境变量在注册和更新函数时设置
                String showDate = context.getUserData("show_date");
                if (showDate != null) {
                    response = "hello " + name + ",today is " + LocalDate.now();
                } else {
                    response = "hello " + name;
                }
            } catch(Exception e) {
                e.printStackTrace();
                response = "please enter your name,for example:{'name':'yuanrong'}";
            }
            return response;
        }

        // 函数初始化入口，函数实例启动时执行一次
        public void initializer(Context context) {
            System.out.println("function instance initialization completed");
        }
    }
    ```

    :::
    :::{dropdown} pom.xml 文件内容，**请参考[安装 Java SDK](install-yuanrong-java-sdk) 并配置依赖 faas-function-sdk**
    :chevron: down-up
    :icon: chevron-down

    ```xml
    <?xml version="1.0" encoding="UTF-8"?>
    <project xmlns="http://maven.apache.org/POM/4.0.0"
            xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
            xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
        <modelVersion>4.0.0</modelVersion>

        <groupId>org.yuanrong.demo</groupId>
        <artifactId>demo</artifactId>
        <version>1.0.0</version>

        <properties>
            <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
            <maven.compiler.source>1.8</maven.compiler.source>
            <maven.compiler.target>1.8</maven.compiler.target>
            <maven.build.timestamp.format>yyyyMMddHHmmss</maven.build.timestamp.format>
            <package.finalName>demo</package.finalName>
            <package.outputDirectory>target</package.outputDirectory>
        </properties>

        <dependencies>
            <dependency>
                <!-- 修改版本号为您实际使用版本 -->
                <groupId>org.yuanrong</groupId>
                <artifactId>faas-function-sdk</artifactId>
                <version>1.0.0</version>
            </dependency>
            <dependency>
                <groupId>com.google.code.gson</groupId>
                <artifactId>gson</artifactId>
                <version>2.9.0</version>
            </dependency>
        </dependencies>

        <build>
            <plugins>
                <!-- 配置如下打包方式 -->
                <plugin>
                    <groupId>org.apache.maven.plugins</groupId>
                    <artifactId>maven-assembly-plugin</artifactId>
                    <version>2.2</version>
                    <configuration>
                        <archive>
                            <manifest>
                                <addDefaultImplementationEntries>true</addDefaultImplementationEntries>
                            </manifest>
                        </archive>
                        <appendAssemblyId>false</appendAssemblyId>
                    </configuration>
                    <executions>
                        <execution>
                            <id>auto-deploy</id>
                            <phase>package</phase>
                            <goals>
                                <goal>single</goal>
                            </goals>
                            <configuration>
                                <descriptors>
                                    <descriptor>src/main/assembly/zip_file.xml</descriptor>
                                </descriptors>
                                <finalName>${package.finalName}</finalName>
                                <outputDirectory>${package.outputDirectory}</outputDirectory>
                                <archiverConfig>
                                    <directoryMode>0700</directoryMode>
                                    <fileMode>0600</fileMode>
                                    <defaultDirectoryMode>0700</defaultDirectoryMode>
                                </archiverConfig>
                            </configuration>
                        </execution>
                    </executions>
                </plugin>
            </plugins>
        </build>
    </project>
    ```

    :::
    :::{dropdown} zip_file.xml 文件内容
    :chevron: down-up
    :icon: chevron-down

    ```xml
    <?xml version="1.0" encoding="UTF-8"?>

    <assembly xmlns="http://maven.apache.org/plugins/maven-assembly-plugin/assembly/1.1.2">

        <id>auto-deploy</id>
        <baseDirectory/>
        <formats>
            <format>zip</format>
        </formats>

        <fileSets>
            <fileSet>
                <directory>src/main/resources/</directory>
                <outputDirectory>config</outputDirectory>
                <includes>
                    <include>**</include>
                </includes>
                <fileMode>0600</fileMode>
                <directoryMode>0700</directoryMode>
            </fileSet>

        </fileSets>

        <dependencySets>
            <dependencySet>
                <outputDirectory>lib</outputDirectory>
                <scope>runtime</scope>
            </dependencySet>
        </dependencySets>
    </assembly>
    ```

    :::

2. 编译构建

    在 `service-java-demo` 目录下，执行如下命令构建：

    ```bash
    mvn clean package
    ```

    成功构建将在 `service-java-demo/target` 目录下生成压缩包 `demo.zip` ，并且使用 [Minio 客户端](../../reference/development-tools.md)上传到MinIO。
        
    ```bash
    mc mb mys3/demo-bucket
    mc cp ./demo.zip mys3/demo-bucket/demo.zip
    ```

3. 函数注册及调用

    使用 curl 工具注册函数，参数含义详见 [API 说明](../api/function_service/register_function.md)：

    新建如下内容文件 create_func.json。

    ```json
    {
        "name": "0@myService@java-demo",
        "runtime": "java8",
        "handler": "org.yuanrong.demo.Demo::handler",
        "environment": {
            "show_date": "true"
        },
        "extendedHandler": {
            "initializer": "org.yuanrong.demo.Demo::initializer"
        },
        "extendedTimeout": {
            "initializer": 30
        },
        "kind": "faas",
        "cpu": 600,
        "memory": 512,
        "timeout": 60,
        "storageType": "s3",
        "s3CodePath": {
            "bucketId": "demo-bucket",
            "objectId": "demo.zip",
            "bucketUrl": "http://{Your MinIO Address:9000}"
        }
    }
    ```

    ```bash
    # 替换 /opt/mycode/service 为您的代码包目录
    META_SERVICE_ENDPOINT=<meta service 组件的服务端点，默认为：http://{主节点 ip}:31182>
    curl -H "Content-type: application/json" -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -H 'Content-Type: application/json' \
         -H 'x-storage-type: local' \
         -d @create_func.json
    ```

    结果返回格式如下，记录 `functionVersionUrn` 字段的值用    于调用，这里对应 `sn:cn:yrk:12345678901234561234567890123456:function:0@myService@java-demo:latest`

    ```bash
    {"code":0,"message":"SUCCESS","function":{"id":"sn:cn:yrk:12345678901234561234567890123456:function:0@myService@java-demo:latest","createTime":"2025-05-20 06:26:42.396 UTC","updateTime":"","functionUrn":"sn:cn:yrk:12345678901234561234567890123456:function:0@myService@java-demo","name":"0@myService@java-demo","tenantId":"12345678901234561234567890123456","businessId":"yrk","productId":"","reversedConcurrency":0,"description":"","tag":null,"functionVersionUrn":"sn:cn:yrk:12345678901234561234567890123456:function:0@myService@java-demo:latest","revisionId":"20250520062642396","codeSize":0,"codeSha256":"","bucketId":"demo-bucket","objectId":"demo.zip","handler":"org.yuanrong.demo.Demo::handler","layers":null,"cpu":600,"memory":512,"runtime":"java8","timeout":60,"versionNumber":"latest","versionDesc":"latest","environment":{"show_date":"true"},"customResources":null,"statefulFlag":0,"lastModified":"","Published":"2025-05-20 06:26:42.396 UTC","minInstance":0,"maxInstance":100,"concurrentNum":100,"funcLayer":[],"status":"","instanceNum":0,"device":{},"created":""}}
    ```

    使用 curl 工具调用函数，参数含义详见 [API 说明](../api/function_service/function_invocation.md)：

    ```bash
    FRONTEND_ENDPOINT=<frontend 组件的终端节点，默认为：http://{主节点 ip}:8888>
    FUNCTION_VERSION_URN=<前一步骤记录的 functionVersionUrn>
    curl -H "Content-type: application/json" -X POST -i ${FRONTEND_ENDPOINT}/serverless/v1/functions/${FUNCTION_VERSION_URN}/invocations -d '{"name":"yuanrong"}'
    ```

    结果输出：

    ```bash
    HTTP/1.1 200 OK
    Content-Type: application/json
    X-Billing-Duration: this is billing duration TODO
    X-Inner-Code: 0
    X-Invoke-Summary:
    X-Log-Result: dGhpcyBpcyB1c2VyIGxvZyBUT0RP
    Date: Tue, 26 Jan 2026 11:30:49 GMT
    Content-Length: 36
    
    "hello yuanrong,today is 2026-1-26"
    ```

::::
:::::
