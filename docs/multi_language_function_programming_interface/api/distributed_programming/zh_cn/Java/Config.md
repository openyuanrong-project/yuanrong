# Config

包名：`org.yuanrong`。

## public class Config

类 Config 是 openYuanrong 的初始化数据结构，用于存放初始化 openYuanrong 系统时需要的 IP，端口，URN 等基础信息。 Config 实例是 init 接口的入参，除 functionURN，serverAddress，dataSystemAddress，cppFunctionURN（表格前四）为必须配置并支持通过构造函数配置外，其余参数皆为缺省并通过 setter 设置，具体接口详见末尾表格。

```java

 Config conf = new Config("sn:cn:yrk:default:function:0-${serviceName}-${}fun:$latest", "serverAddressIP", "dataSystemAddressIP", "cppFunctionURN", false);
 conf.setRecycleTime(100);
```

### 接口说明

#### public Config(String functionURN, String serverAddress, String dataSystemAddress, String cppFunctionURN)

Config 的构造函数。

- 参数：

   - **functionURN** - 部署函数返回的 functionURN。
   - **serverAddress** - 集群 ip（openYuanrong 集群主节点）。
   - **dataSystemAddress** - 数据系统 ip（openYuanrong 集群主节点）。
   - **cppFunctionURN** - 部署 cpp 函数返回的 functionID。

#### public Config()

Config 的构造函数。

#### public Config(String functionURN, String serverAddress, String dataSystemAddress, String cppFunctionURN, boolean isDriver)

Config 的构造函数。

- 参数：

   - **functionURN** - 部署函数返回的 functionURN。
   - **serverAddress** - 集群 ip（openYuanrong 集群主节点）。
   - **dataSystemAddress** - 数据系统 ip（openYuanrong 集群主节点）。
   - **cppFunctionURN** - 部署 cpp 函数返回的 functionID。
   - **isDriver** - 云上/云外。

#### public Config(String functionUrn, String serverAddr, int serverAddressPort, String dataSystemAddr, int dataSystemAddressPort, String cppFunctionUrn)

Config 的构造函数。

- 参数：

   - **functionURN** - 部署函数返回的 functionURN。
   - **serverAddr** - 集群 ip（openYuanrong 集群主节点）。
   - **serverAddressPort** - 集群端口号。
   - **dataSystemAddress** - 数据系统 ip（openYuanrong 集群主节点）。
   - **dataSystemAddressPort** - 数据系统端口号。
   - **cppFunctionURN** - 部署 cpp 函数返回的 functionID。

#### public Config(String functionUrn, String serverAddr, int serverAddressPort, String dataSystemAddr, int dataSystemAddressPort, String cppFunctionUrn, boolean isDriver)

Config 的构造函数。

- 参数：

   - **functionURN** - 部署函数返回的 functionURN。
   - **serverAddr** - 集群 ip（openYuanrong 集群主节点）。
   - **serverAddressPort** - 集群端口号。
   - **dataSystemAddress** - 数据系统 ip（openYuanrong 集群主节点）。
   - **dataSystemAddressPort** - 数据系统端口号。
   - **cppFunctionURN** - 部署 cpp 函数返回的 functionID。
   - **isDriver** - 云上/云外。

### 私有成员

``` java

private ArrayList<String> loadPaths = new ArrayList<>()
```

指定 so 路径，默认不填则由 services.yaml 指定。

``` java

private String functionURN = Utils.getEnvWithDefualtValue("YRFUNCID", DEFAULT_FUNC_URN, "config-")
```

部署函数返回的 functionURN。可以通过 YRFUNCID 环境变量设置，当前 functionURN 生成逻辑为 ``sn:cn:yrk:default:function:0-{ServiceName}-{FunctionName}:$latest``。

``` java

private String cppFunctionURN = DEFAULT_CPP_URN
```

部署 cpp 函数返回的 functionID。默认 ``sn:cn:yrk:default:function:0-defaultservice-cpp:$lates``。

``` java

private String serverAddress = ""
```

集群 ip（openYuanrong 集群主节点）。

``` java

private String dataSystemAddress = ""
```

数据系统 ip（openYuanrong 集群主节点）。

``` java

private int rpcTimeout = DEFAULT_TIMEOUT
```

rpc 超时时间（秒）。默认 ``30 * 60``。

``` java

private int serverAddressPort = DEFAULT_SERVER_PORT
```

集群端口号，默认 ``31222``。

``` java

private int dataSystemAddressPort = DEFAULT_DS_PORT
```

数据系统端口号，默认 ``31222``。

``` java

private String ns = ""
```

当用户使用具名有状态函数时，``ns`` 是有状态函数的命名空间。默认值为空字符串。如果 ``ns`` 不是空字符串，则有状态函数名称为 ``namespace-name``。

``` java

private int threadPoolSize = 0
```

线程池大小，仅在 ParallelFor 中使用。

``` java

private boolean isDriver = true
```

是否为 driver 程序，决定 api 在云上或者本地被使用，默认 ``true``。

``` java

private int recycleTime = DEFAULT_RECYCLETIME
```

回收时间（秒），默认 ``10``，取值要求大于 0。

``` java

private boolean enableMetrics = false
```

是否开启指标采集上报，默认 ``false``。

``` java

private boolean enableMTLS = false
```

是否开启云外客户端双向认证，默认 ``false``。

``` java

private String certificateFilePath
```

客户端证书文件路径。

``` java

private String privateKeyPath
```

客户端私钥文件路径。

``` java

private String verifyFilePath
```

服务端证书文件路径。

``` java

private String serverName
```

服务端 name。

``` java

private String logDir = DEFAULT_LOG_DIR
```

客户端日志的输入目录，默认是当前目录。若目录不存在则创建目录并在该目录下生成日志文件。

``` java

private int maxLogSizeMb = 0
```

客户端日志的单个文件最大体积，单位 MB，默认 ``0`` （若为默认 ``0``，则最终会被置为``40``）。

超出大小限制后滚动切分，每 30 秒检查滚动一次，负数会抛出异常。

``` java

private int maxLogFileNum = 0
```

客户端日志滚动后最多保留的文件数，默认 ``0`` (若为默认 ``0``，则最终会被置为 ``20``)，超出数量限制后从最旧的日志开始删除，每 30 秒检查删除一次。

``` java

private String logLevel = ""
```

日志级别，取值范围 ``DEBUG``、``INFO``、``WARN`` 和 ``ERROR``。

若未设置或者设置非法值将会被设置为默认值 ``INFO``。

``` java

private int maxTaskInstanceNum = -1
```

限制无状态函数创建的最大实例数量，合理取值范围是 1~65536。未配置时默认 ``-1``，表示不限制数量。如果输入异常值，`Init` 接口会抛出异常。

``` java

private int maxConcurrencyCreateNum = 100
```

设置无状态实例的最大并发创建数量，必须大于 0，默认 ``100``。

``` java

private boolean enableDisConvCallStack = true
```

如果为 ``true``，则启用分布式集合调用栈，用户代码抛出的异常将被准确定位。用户可以使用分布式集合调用栈以更简单的方式定位分布式调用链中发生的异常和错误，就像在单个进程中一样。

``` java

private boolean isThreadLocal = false
```

是否开启多集群模式，默认 ``false``。如果 isThreadLocal 为 ``true``，在不同的线程中调用 `YR.init`，设置不同的集群地址，runtime java sdk 可以连接到不同的集群。

``` java

private Map<String, String> customEnvs = new HashMap<>()
```

用于给 runtime 设置自定义环境变量，当前只支持设置 ``LD_LIBRARY_PATH``。

``` java

private List<String> codePath
```

代码路径。

``` java

private boolean enableDsEncrypt = false
```

设置为 ``true`` 时，启用数据系统的 TLS 认证，否则不启用。

``` java

private String dsPublicKeyContextPath = ""
```

数据系统 TLS 认证的 worker 公钥路径。enableDsEncrypt 设置为 ``true`` 时，该参数不能为空，否则将抛出异常。

``` java

private String runtimePublicKeyContextPath = ""
```

数据系统 TLS 认证的客户端公钥路径。enableDsEncrypt 设置为 ``true`` 时，该参数不能为空，否则将抛出异常。

``` java

private String runtimePrivateKeyContextPath = ""
```

数据系统 TLS 认证的客户端私钥路径。enableDsEncrypt 设置为 ``true`` 时，该参数不能为空，否则将抛出异常。

#### 示例代码

``` java

Config conf = new Config("sn:cn:yrk:default:function:0-${serviceName}-${}fun:$latest", "serverAddressIP", "dataSystemAddressIP", "cppFunctionURN", false);
conf.setRecycleTime(100);
```

### 接口表格

| 字段名称 | getter 接口 | setter 接口 |
| -------- | ----------- | ----------- |
| loadPaths | getLoadPaths | setLoadPaths |
| functionURN | getFunctionURN | setFunctionURN |
| cppFunctionURN | getCppFunctionURN | setCppFunctionURN |
| serverAddress | getServerAddress | setServerAddress |
| dataSystemAddress | getDataSystemAddress | setDataSystemAddress |
| rpcTimeout | getRpcTimeout | setRpcTimeout |
| serverAddressPort | getServerAddressPort | setServerAddressPort |
| dataSystemAddressPort | getDataSystemAddressPort | setDataSystemAddressPort |
| ns | getNs | setNs |
| threadPoolSize | getThreadPoolSize | setThreadPoolSize |
| isDriver | isDriver | setDriver |
| recycleTime | getRecycleTime | setRecycleTime |
| enableMetrics | isEnableMetrics | setEnableMetrics |
| enableMTLS | isEnableMTLS | setEnableMTLS |
| certificateFilePath | getCertificateFilePath | setCertificateFilePath |
| privateKeyPath | getPrivateKeyPath | setPrivateKeyPath |
| verifyFilePath | getVerifyFilePath | setVerifyFilePath |
| serverName | getServerName | setServerName |
| logDir | getLogDir | setLogDir |
| maxLogSizeMb | getMaxLogSizeMb | setMaxLogSizeMb |
| maxLogFileNum | getMaxLogFileNum | setMaxLogFileNum |
| logLevel | getLogLevel | setLogLevel |
| maxTaskInstanceNum | getMaxTaskInstanceNum | setMaxTaskInstanceNum |
| maxConcurrencyCreateNum | getMaxConcurrencyCreateNum | setMaxConcurrencyCreateNum |
| enableDisConvCallStack | isEnableDisConvCallStack | setEnableDisConvCallStack |
| isThreadLocal | isThreadLocal | setThreadLocal |
| customEnvs | getCustomEnvs | setCustomEnvs |
| codePath | getCodePath | setCodePath |
| enableDsEncrypt | isEnableDsEncrypt | setEnableDsEncrypt |
| dsPublicKeyContextPath | getDsPublicKeyContextPath | setDsPublicKeyContextPath |
| runtimePublicKeyContextPath | getRuntimePublicKeyContextPath | setRuntimePublicKeyContextPath |
| runtimePrivateKeyContextPath | getRuntimePrivateKeyContextPath | setRuntimePrivateKeyContextPath |
