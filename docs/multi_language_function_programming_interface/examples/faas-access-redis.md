# 函数服务连接后端存储

服务类应用开发中，业务经常需要连接后端服务比如 MySQL，Redis 等存储数据，或者连接 Kafka，RabbitMQ 等传递消息。基于 openYuanrong 函数服务函数开发应用，可以使用后端服务原生提供的连接方式如 客户端 SDK 等与服务端建立连接。

## 方案介绍

基于 openYuanrong 函数服务使用 Java 开发一个应用，它提供一个 REST API，连接 Redis 集群读写数据。函数中实现两个方法，初始化方法用于和 Redis 集群建链，执行方法处理外部请求。

需要注意是，之所以在函数服务的初始化方法中处理建链、创建连接池等逻辑，是因为初始化方法仅在函数实例启动时执行一次，创建的连接可在处理请求的执行方法中复用，避免反复建链带来的性能等问题。

## 准备工作

1. [在 K8s 上部署 openYuanrong 集群](../../deploy/deploy_on_k8s/index.md)在 K8s 集群主节点使用 kubectl 工具获取openYuanrong服务端点。

    meta service 服务，它负责函数及资源池的管理等功能。

    ```bash
    echo "http://$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'):$(kubectl get svc meta-service -o jsonpath='{.spec.ports[0].nodePort}')"
    ```

    frontend 服务，它负责接入流量，承担函数调用等功能。

    ```bash
    echo "http://$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'):$(kubectl get svc frontend-lb -o jsonpath='{.spec.ports[0].nodePort}')"
    ```

2. 安装函数服务 [Java SDK](../../deploy/installation.md#java-sdk) 中的 faas-function-sdk 与 [MinIO Client](tools-minio-client)。MinIO Client 用于上传代码包到 openYuanrong 集群中的 MinIO 服务。
3. 已部署 Redis 集群。

## 实现流程

### 创建函数服务应用工程

创建一个 Java 语言的函数服务应用工程，目录结构如下，包含 pom.xml，zip_file.xml，FaaSHandler.java，Scenario.java，Utils.java 五个文件。

- pom.xml：maven 配置文件，引入openYuanrong SDK、Redis 等依赖和 build plugin 配置。
- zip_file.xml：代码打包配置。
- FaaSHandler.java：核心业务处理类，包括 `initializer` 初始化方法和 `handler` 执行方法。
- Scenario.java：格式化外部请求参数。
- Utils.java：Redis 连接配置。

``` bash
access-reids
├── pom.xml
└── src
    └── main
        ├── assembly
        │   └── zip_file.xml
        └── java
            └── com
                └── yuanrong
                    └── demo
                        ├── FaaSHandler.java
                        ├── Scenario.java
                        └── Utils.java
```

:::{dropdown} pom.xml 文件内容，**须对应修改 faas-function-sdk 版本号为您的安装版本**
:chevron: down-up
:icon: chevron-down

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>org.example</groupId>
    <artifactId>access-redis</artifactId>
    <version>1.0.0</version>

    <properties>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <maven.compiler.source>1.8</maven.compiler.source>
        <maven.compiler.target>1.8</maven.compiler.target>
        <maven.build.timestamp.format>yyyyMMddHHmmss</maven.build.timestamp.format>
        <package.finalName>access-redis</package.finalName>
        <package.outputDirectory>target</package.outputDirectory>
    </properties>

    <dependencies>
        <dependency>
            <groupId>org.apache.commons</groupId>
            <artifactId>commons-pool2</artifactId>
            <version>2.12.0</version>
        </dependency>
        <dependency>
            <groupId>redis.clients</groupId>
            <artifactId>jedis</artifactId>
            <version>5.2.0</version>
            <!-- 排除 slf4j-api 避免和 openYuanrong依赖的log4j 冲突-->
            <exclusions>
                <exclusion>
                    <groupId>org.slf4j</groupId>
                    <artifactId>slf4j-api</artifactId>
                </exclusion>
            </exclusions>
        </dependency>
        <dependency>
            <groupId>org.yuanrong</groupId>
            <artifactId>faas-function-sdk</artifactId>
            <!-- 修改版本号为您实际使用版本 -->
            <version>${VERSION}</version>
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
:::{dropdown} FaaSHandler.java 文件内容
:chevron: down-up
:icon: chevron-down

```java
package com.yuanrong.demo;

import org.yuanrong.services.runtime.Context;
import org.yuanrong.services.runtime.RuntimeLogger;

import com.google.gson.Gson;
import com.google.gson.JsonObject;

import org.apache.commons.pool2.impl.GenericObjectPoolConfig;
import redis.clients.jedis.Connection;
import redis.clients.jedis.HostAndPort;
import redis.clients.jedis.JedisCluster;

import java.time.Duration;
import java.util.Set;
import java.util.HashSet;

public class FaaSHandler {
    // 可复用的 Redis 集群客户端
    private JedisCluster jedisCluster = null;

    // 初始化方法，实例启动时执行一次，用于建立 redis 连接池
    public void initializer(Context context) {
        // 使用openYuanrong内置日志模块打印日志
        RuntimeLogger log = context.getLogger();

        // 从环境变量中获取 redis 集群地址，地址以逗号分开，例如：10.x.x.1:6379,10.x.x.2:6379,10.x.x.3:6379
        String redisClusterNodes = context.getUserData("redis_ip_address");
        log.info("redis cluster address:" + redisClusterNodes);

        // 从环境变量中获取 redis 密码
        String pwd = context.getUserData("redis_password");
        Set<HostAndPort> jedisClusterNodes = new HashSet<>();
        for (String clusterNodeStr : redisClusterNodes.split(",")) {
            String[] nodeInfo = clusterNodeStr.split(":");
            jedisClusterNodes.add(new HostAndPort(nodeInfo[0], Integer.parseInt(nodeInfo[1])));
        }
        GenericObjectPoolConfig<Connection> config = Utils.getJedisPoolConfig();
        jedisCluster = new JedisCluster(jedisClusterNodes, Utils.REDIS_CONNECT_TIMEOUT,
                Utils.REDIS_READ_TIMEOUT, 3,
                pwd, "faas-redis-client", config);
    }

    // 业务处理方法，每次外部请求都会触发执行，body 参数及返回值类型可自定义
    public JsonObject handler(String body, Context context){
        RuntimeLogger log = context.getLogger();
        log.info("request body:" + body);

        Gson gson = new Gson();
        Scenario s = gson.fromJson(body, Scenario.class);
        if (s == null) {
            log.info("parse object failed");
        }
        int size = s.getRecords().size();
        if (size <= 0 || size > 1000) {
            log.error("Records number must [1,1000],current:" + size);
            return new Gson().fromJson("{\"success\":false,\"message\":\"records must range from 1 to 1000\"}", JsonObject.class);
        }

        long  startTime = System.currentTimeMillis();
        for(String record:s.getRecords()){
            try {
                jedisCluster.lpush(s.getScenarioName(), record);
            }catch (Exception e){
                log.error("Redis error:" + e.getMessage());
                return new Gson().fromJson("{\"success\":false,\"message\":\""+e.getMessage()+"\"}", JsonObject.class);
            }
        }
        long endTime = System.currentTimeMillis();
        log.info("write records cost:" + (endTime - startTime));

        long count = jedisCluster.llen(s.getScenarioName());
        log.info("total records:" + count);
        for (long i = 0; i < count; i++) {
            String record = jedisCluster.lpop(s.getScenarioName());
            log.info("Pop record:" + record);
        }
        log.info("all is ok");
        return new Gson().fromJson("{\"success\":true, \"message\":\"\"}", JsonObject.class);
    }
}
```

:::
:::{dropdown} Scenario.java 文件内容
:chevron: down-up
:icon: chevron-down

```java
package com.yuanrong.demo;

import java.util.ArrayList;

public class Scenario {
    private String scenarioName;
    private ArrayList<String> records;

    public Scenario(String scenarioName, ArrayList<String> records) {
        super();
        this.scenarioName = scenarioName;
        this.records = records;
    }

    public Scenario() {
        super();
    }

    @Override
    public String toString() {
        return "scenarioName:" + scenarioName + ";records:" + records;
    }

    public void setScenarioName(String scenarioName) {
        this.scenarioName = scenarioName;
    }

    public String getScenarioName() {
        return scenarioName;
    }

    public void setRecords(ArrayList<String> records) {
        this.records = records;
    }

    public ArrayList<String> getRecords() {
        return records;
    }
}
```

:::
:::{dropdown} Utils.java 文件内容
:chevron: down-up
:icon: chevron-down

```java
package com.yuanrong.demo;

import org.apache.commons.pool2.impl.GenericObjectPoolConfig;

import java.time.Duration;

public class Utils {
    public static final Integer REDIS_CONNECT_TIMEOUT = 3000;
    public static final Integer REDIS_READ_TIMEOUT = 2000;
    public static final Integer REDIS_DEFAULT_MIN_IDLE = 5;
    public static final Integer REDIS_DEFAULT_MAX_TOTAL = 50;
    public static final Integer REDIS_DEFAULT_MAX_IDLE = 50;

    public static <T> GenericObjectPoolConfig<T> getJedisPoolConfig() {
        GenericObjectPoolConfig<T> config = new GenericObjectPoolConfig<>();
        config.setTestOnBorrow(false);
        config.setTestOnReturn(false);
        config.setTestOnCreate(false);
        config.setTestWhileIdle(true);
        config.setBlockWhenExhausted(true);
        config.setMaxWait(Duration.ofSeconds(3));
        int minIdle = getEnvValueInt("redis_minIdle", REDIS_DEFAULT_MIN_IDLE);
        int maxTotal = getEnvValueInt("redis_maxTotal", REDIS_DEFAULT_MAX_TOTAL);
        int maxIdle = getEnvValueInt("redis_maxIdle", REDIS_DEFAULT_MAX_IDLE);
        config.setMinIdle(minIdle);
        config.setMaxTotal(maxTotal);
        config.setMaxIdle(maxIdle);
        config.setTimeBetweenEvictionRuns(Duration.ofSeconds(600));
        return config;
    }

    public static int getEnvValueInt(String key, int defaultValue) {
        String envValue = System.getenv(key);
        if (envValue != null && envValue.length() != 0) {
            return Integer.parseInt(envValue);
        }
        return defaultValue;
    }
}
```

:::

### 打包函数服务应用

在工程 `access-redis` 目录下，执行 maven 打包命令。

```bash
mvn clean package
```

部署包 access-redis.zip 生成在 target 目录下。

### 部署函数服务应用

参考“准备工作” 安装 MinIO Client 步骤中的常用命令，上传部署包 access-redis.zip 到 openYuanrong 集群中的 Minio 服务。

在 `access-redis` 目录下新建 create_func.json 文件，内容如下，作为注册函数服务函数的请求参数，参数含义详见[API 说明](../api/function_service/register_function.md)

:::{Note}

根据实际情况替换以下字段内容：

- environment：函数使用的环境变量，包括 Redis 集群节点地址及密码
- s3CodePath：代码包在 minio 上存放的位置

:::

```json
{
    "name": "0@faas@access-redis",
     "handler": "com.yuanrong.demo.FaaSHandler.handler",
    "runtime": "java8",
    "description": "",
    "cpu": 600,
    "memory": 512,
    "timeout": 30,
    "concurrentNum": "100",
    "minInstance": "1",
    "maxInstance": "100",
    "environment": {
        "redis_ip_address": "10.x.x.1:6379,10.x.x.2:6379,10.x.x.3:6379",
        "redis_password":"xxx"
    },
    "layers": [],
    "kind": "faas",
    "storageType": "s3",
    "s3CodePath": {
        "bucketId": "code",
        "objectId": "access-redis.zip",
        "bucketUrl": "http://x.x.x.x:30110/"
    },
    "codePath": "",
    "schedulePolicies": [
    ],
    "extendedHandler": {
         "initializer": "com.yuanrong.demo.FaasHandler.initializer"
    },
    "extendedTimeout": {
        "initializer": 600
    },
    "customResources": {}
}
```

使用 curl 工具注册函数服务。

```shell
META_SERVICE_ENDPOINT=<“准备工作”步骤中获取的 meta service 服务端点>
curl -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/functions -H 'Content-Type: application/json' -H 'x-storage-type: local' -d @create_func.json
```

结果返回格式如下，记录 `functionVersionUrn` 字段的值用于调用，这里对应 `sn:cn:yrk:12345678901234561234567890123456:function:0@faas@access-redis:latest`

```bash
{"code":0,"message":"SUCCESS","function":{"id":"sn:cn:yrk:12345678901234561234567890123456:function:0@faas@access-redis:latest","createTime":"2025-03-11 07:23:08.339 UTC","updateTime":"","functionUrn":"sn:cn:yrk:12345678901234561234567890123456:function:0@faas@access-redis","name":"0@faas@access-redis","tenantId":"12345678901234561234567890123456","businessId":"yrk","productId":"","reversedConcurrency":0,"description":"","tag":null,"functionVersionUrn":"sn:cn:yrk:12345678901234561234567890123456:function:0@faas@access-redis:latest","revisionId":"20250311072308339","codeSize":0,"codeSha256":"","bucketId":"code","objectId":"access-redis.zip","handler":"com.yuanrong.demo.FaaSHandler.handler","layers":null,"cpu":600,"memory":512,"runtime":"java8","timeout":30,"versionNumber":"latest","versionDesc":"latest","environment":{"redis_ip_address":"10.x.x.1:6379,10.x.x.2:6379,10.x.x.3:6379","redis_password":"xxx"},"customResources":null,"statefulFlag":0,"lastModified":"","Published":"2025-03-11 07:23:08.339 UTC","minInstance":1,"maxInstance":100,"concurrentNum":100,"funcLayer":[],"status":"","instanceNum":0,"device":{},"created":""}}
```

### 调用函数服务应用

#### 准备资源池

运行应用前，需要先在openYuanrong集群中初始化一个和应用资源（cpu、memory 等）配置（见前一步骤创建的 create_func.json 文件）匹配的资源池。

在 `access-redis` 目录下新建 create_pool.json 文件，内容如下，作为创建资源池的请求参数，参数含义详见 [API 说明](../../deploy/deploy_on_k8s/api/create_pod_pool.md)

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

使用 curl 工具创建资源池。

```shell
META_SERVICE_ENDPOINT=<“准备工作”步骤中获取的 meta service 服务端点>
curl -X POST -i ${META_SERVICE_ENDPOINT}/serverless/v1/podpools -H 'Content-Type: application/json' -d @create_pool.json
```

结果返回如下：

```bash
{"code":0,"message":"","result":{"failed_pools":null}}
```

#### 调用应用

使用 curl 工具调用应用。

```shell
FUNCTION_VERSION_URN=<“部署微服务”步骤记录的 functionVersionUrn>
FRONTEND_ENDPONT=<“准备工作”步骤中获取的 frontend 服务端点>
curl -X POST -i ${FRONTEND_ENDPONT}/serverless/v1/functions/${FUNCTION_VERSION_URN}/invocations -H 'Content-Type: application/json' -d "{\"scenarioName\":\"snap-up\",\"records\":[\"phone-238,watch-862\",\"phone-858,watch-354\"]}"
```

结果输出：

```json
{
    "success": true,
    "message": ""
}
```
