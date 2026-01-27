# 迁移 Springboot 容器微服务到openYuanrong

openYuanrong兼容主流 Java 微服务框架，支持微服务作为openYuanrong 函数统一部署运行，享受 Serverless 带来的弹性、免运维等优势。

## 方案介绍

在一个 SpringBooot 微服务工程中，引入openYuanrong提供的适配器 SDK 作为依赖并修改部分构建配置，打包部署为openYuanrong 函数服务，实现代码接近零改动迁移至openYuanrong平台。

## 准备工作

1. 部署openYuanrong集群（[在 Kubernetes 上部署](../deploy/deploy_on_k8s/index.md)）并在 k8s 集群主节点使用 kubectl 工具获取openYuanrong服务端点

    meta service 服务，它负责函数及资源池的管理等功能

    ```bash
    echo "http://$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'):$(kubectl get svc meta-service -o jsonpath='{.spec.ports[0].nodePort}')"
    ```

    frontend 服务，它负责接入流量，承担函数调用等功能

    ```bash
    echo "http://$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'):$(kubectl get svc faas-frontend-lb -o jsonpath='{.spec.ports[0].nodePort}')"
    ```

2. 安装 MinIO 客户端（[安装及配置](../reference/development-tools.md)），用于上传代码包到openYuanrong集群中的 MinIO 服务。
3. 安装openYuanrong微服务 Adapter SDK

    [下载openYuanrong发布包](../reference/releases.md) openyuanrong-*.tar.gz 并依次执行如下命令：

    ```bash
    tar -xzf openyuanrong-*.tar.gz
    cd openyuanrong
    WORKSPACE=$(pwd)

    cd ${WORKSPACE}/data_system/sdk/
    mvn install:install-file -Dfile=datasystem-${VERSION}_${ARCH}.jar -DartifactId=datasystem -DgroupId=org.yuanrong.datasystem -Dversion=${VERSION} -Dpackaging=jar

    cd ${WORKSPACE}/runtime/sdk/java/
    mvn install:install-file -Dfile=faas-function-sdk-${VERSION}.jar -DartifactId=faas-function-sdk -DgroupId=org.yuanrong -Dversion=${VERSION} -Dpackaging=jar

    cd ${WORKSPACE}/spring-adapter/sdk/
    bash install-adapter.sh -v ${VERSION}
    ```

## 实现流程

### 创建 SpringBooot 微服务工程

创建一个 SpringBooot 微服务工程，或基于您已有 Springboot 工程参考修改，目录结构如下，包含 pom.xml，zip_file.xml，Application.java 及 MyController.java 四个文件。

- pom.xml：maven 配置文件，增加openYuanrong microservice-function-yuanrong 依赖和 build plugin 配置。
- zip_file.xml：代码打包配置。
- Application.java：微服务主类。
- MyController.java：处理 web 请求，实现接口 `/rest/hello`。

``` bash
microservice-demo
├── pom.xml
└── src
    └── main
        ├── assembly
        │   └── zip_file.xml
        └── java
            └── com
                └── example
                    └── microservicedemo
                        ├── Application.java
                        └── MyController.java
```

:::{dropdown} pom.xml 文件内容，**须对应修改 microservice-function-yuanrong 版本号为您的安装版本**
:chevron: down-up
:icon: chevron-down

```xml
<?xml version="1.0" encoding="UTF-8"?>

<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>2.6.7</version>
        <relativePath/>
    </parent>

    <groupId>com.example</groupId>
    <artifactId>microservice-example</artifactId>
    <version>1.0.0</version>

    <properties>
        <package.finalName>microservice-demo</package.finalName>
        <package.outputDirectory>target</package.outputDirectory>
    </properties>

    <dependencies>
        <dependency>
            <groupId>org.yuanrong.m2s</groupId>
            <artifactId>microservice-function-yuanrong</artifactId>
            <!-- 修改版本号为您实际使用版本 -->
            <version>${VERSION}</version>
        </dependency>
        <dependency>
            <!-- 排除 spring-boot-starter-logging 避免和 log4j 冲突-->
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
                <exclusions>
                    <exclusion>
                        <groupId>org.springframework.boot</groupId>
                        <artifactId>spring-boot-starter-logging</artifactId>
                    </exclusion>
                </exclusions>
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
    <baseDirectory></baseDirectory>
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
            <fileMode>0660</fileMode>
            <directoryMode>0760</directoryMode>
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
:::{dropdown} Application.java 文件内容
:chevron: down-up
:icon: chevron-down

```java
package com.example.microservicedemo;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class Application {

    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }

}
```

:::
:::{dropdown} MyController.java 文件内容
:chevron: down-up
:icon: chevron-down

```java
package com.example.microservicedemo;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RequestMapping("/rest")
@RestController
public class MyController {
    @GetMapping(value = "/hello")
    public String helloWorld(String name) {
        return "Hello World, " + name;
    }
}
```

:::

### 打包微服务

在工程 microservice-demo 目录下，执行 maven 打包命令。

```bash
mvn clean package
```

部署包 microservice-demo.zip 生成在 target 目录下。

### 部署微服务

参考“准备工作” 安装 MinIO 步骤中的常用命令，上传 microservice-demo.zip 到openYuanrong集群中的 Minio 服务。

在 `microservice-demo` 目录下新建 create_func.json 文件，内容如下（**根据实际情况替换 s3CodePath 中的字段**），作为注册函数服务的请求参数，参数含义详见[API 说明](../multi_language_function_programming_interface/api/function_service/register_function.md)

```json
{
    "name": "0@microservice@demo",
     "handler": "org.yuanrong.m2s.function.Handler.handleRestRequest",
    "runtime": "java8",
    "description": "",
    "cpu": 600,
    "memory": 512,
    "timeout": 30,
    "concurrentNum": "100",
    "minInstance": "1",
    "maxInstance": "100",
    "environment": {
        "spring_start_class": "com.example.microservicedemo.Application"
    },
    "layers": [],
    "kind": "faas",
    "storageType": "s3",
    "s3CodePath": {
        "bucketId": "this-bucket",
        "objectId": "microservice-demo.zip",
        "bucketUrl": "http://x.x.x.x:9000/"
    },
    "codePath": "",
    "schedulePolicies": [
    ],
    "extendedHandler": {
         "initializer": "org.yuanrong.m2s.function.Handler.initializer"
    },
    "extendedTimeout": {
        "initializer": 600
    },
    "customResources": {}
}
```

使用 curl 工具注册微服务为 函数服务

```shell
META_SERVICE_ENDPOINT=<“准备工作”步骤中获取的 meta service 服务端点>
curl -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -H 'Content-Type: application/json' -H 'x-storage-type: local' -d @create_func.json
```

结果返回格式如下，记录 `functionVersionUrn` 字段的值用于调用，这里对应 `sn:cn:yrk:12345678901234561234567890123456:function:0@microservice@demo:latest`

```bash
{"code":0,"message":"SUCCESS","function":{"id":"sn:cn:yrk:12345678901234561234567890123456:function:0@microservice@demo:latest","createTime":"2025-02-28 07:36:09.088 UTC","updateTime":"","functionUrn":"sn:cn:yrk:12345678901234561234567890123456:function:0@microservice@demo","name":"0@microservice@demo","tenantId":"12345678901234561234567890123456","businessId":"yrk","productId":"","reversedConcurrency":0,"description":"","tag":null,"functionVersionUrn":"sn:cn:yrk:12345678901234561234567890123456:function:0@microservice@demo:latest","revisionId":"20250228073609088","codeSize":0,"codeSha256":"","bucketId":"","objectId":"","handler":"org.yuanrong.m2s.function.Handler.handleRestRequest","layers":null,"cpu":2000,"memory":2048,"runtime":"java8","timeout":30,"versionNumber":"latest","versionDesc":"latest","environment":{"spring_start_class":"com.example.microservicedemo.Application"},"customResources":null,"statefulFlag":0,"lastModified":"","Published":"2025-02-28 07:36:09.088 UTC","minInstance":1,"maxInstance":100,"concurrentNum":100,"funcLayer":[],"status":"","instanceNum":0,"device":{},"created":""}}
```

### 调用微服务

#### 准备资源池

运行微服务前，需要先在openYuanrong集群中初始化一个和服务资源（cpu、memory 等）配置（见前一步骤创建的 create_func.json 文件）匹配的资源池。

在 `microservice-demo` 目录下新建 create_pool.json 文件，内容如下，作为创建资源池的请求参数，参数含义详见 [API 说明](../deploy/deploy_on_k8s/api/create_pod_pool.md)

```json
{
    "pools": [
        {
            "id": "pool-600-512",
            "size": 1,
            "group": "rg1",
            "reuse": false,
            "image": "",
            "init_image": "",
            "labels": {
            },
            "environment": {},
            "volumes": "",
            "volume_mounts": "",
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

使用 curl 工具创建资源池

```shell
META_SERVICE_ENDPOINT=<“准备工作”步骤中获取的 meta service 服务端点>
curl -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/podpools -H 'Content-Type: application/json' -d @create_pool.json
```

结果返回如下：

```bash
{"code":0,"message":"","result":{"failed_pools":null}}
```

#### 服务调用

在 `microservice-demo` 目录下新建 invoke_func.json 文件，内容如下，作为调用函数服务的请求参数，参数含义详见[API 说明](../multi_language_function_programming_interface/api/function_service/function_invocation.md)

```json
{
  "isBase64Encoded": false,
  "httpMethod": "GET",
  "body": "",
  "path":"/rest/hello",
  "headers": {
      "Content-Type": "application/json"
  },
  "pathParameters":{},
  "queryStringParameters": {
    "name": "yuanrong"
  }
}
```

使用 curl 工具调用微服务

```shell
FUNCTION_VERSION_URN=<“部署微服务”步骤记录的 functionVersionUrn>
FRONTEND_ENDPONT=<“准备工作”步骤中获取的 frontend 服务地址>
curl -X POST -i ${FRONTEND_ENDPONT}/serverless/v1/functions/${FUNCTION_VERSION_URN}/invocations -H 'Content-Type: application/json' -d @invoke_func.json
```

结果输出：

```json
{
    "body": "Hello World, yuanrong",
    "headers": {
      "Content-Length": "21",
      "Content-Language": "en",
      "Content-Type": "text/plain"
    },
    "statusCode": 200,
    "isBase64Encoded": false
}
```
