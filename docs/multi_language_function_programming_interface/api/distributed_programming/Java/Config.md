# Config

package: `org.yuanrong`.

## public class Config

The Config class is the initialization data structure of openYuanrong, used to store basic information such as IP, port, and URN needed when initializing the openYuanrong system.

The Config instance is the input parameter of the init interface. Except for functionURN, serverAddress, dataSystemAddress and cppFunctionURN (the top four in the table), which are mandatory configurations and supported through the constructor, the rest of the parameters are default and set through setters. For specific interfaces, please refer to the table at the end.

```java

 Config conf = new Config("sn:cn:yrk:default:function:0-${serviceName}-${}fun:$latest", "serverAddressIP", "dataSystemAddressIP", "cppFunctionURN", false);
 conf.setRecycleTime(100);
```

### Interface description

#### public Config(String functionURN, String serverAddress, String dataSystemAddress, String cppFunctionURN)

The constructor of Config.

- Parameters:

   - **functionURN** - The functionURN returned by the deployment function.
   - **serverAddress** - Cluster IP (openYuanrong cluster master node).
   - **dataSystemAddress** - Data system IP (openYuanrong cluster master node).
   - **cppFunctionURN** - The functionID returned by the deployment cpp function.

#### public Config()

The constructor of Config.

#### public Config(String functionURN, String serverAddress, String dataSystemAddress, String cppFunctionURN, boolean isDriver)

The constructor of Config.

- Parameters:

   - **functionURN** - The functionURN returned by the deployment function.
   - **serverAddress** - Cluster IP (openYuanrong cluster master node).
   - **dataSystemAddress** - Data system IP (openYuanrong cluster master node).
   - **cppFunctionURN** - The functionID returned by the deployment cpp function.
   - **isDriver** - On cloud or off cloud.

#### public Config(String functionUrn, String serverAddr, int serverAddressPort, String dataSystemAddr, int dataSystemAddressPort, String cppFunctionUrn)

The constructor of Config.

- Parameters:

   - **functionURN** - The functionURN returned by the deployment function.
   - **serverAddr** - Cluster IP (openYuanrong cluster master node).
   - **serverAddressPort** - Cluster port number.
   - **dataSystemAddress** - Data system IP (openYuanrong cluster master node).
   - **dataSystemAddressPort** - DataSystem port number.
   - **cppFunctionURN** - The functionID returned by the deployment cpp function.

#### public Config(String functionUrn, String serverAddr, int serverAddressPort, String dataSystemAddr, int dataSystemAddressPort, String cppFunctionUrn, boolean isDriver)

The constructor of Config.

- Parameters:

   - **functionURN** - The functionURN returned by the deployment function.
   - **serverAddr** - Cluster IP (openYuanrong cluster master node).
   - **serverAddressPort** - Cluster port number.
   - **dataSystemAddress** - Data system IP (openYuanrong cluster master node).
   - **dataSystemAddressPort** - DataSystem port number.
   - **cppFunctionURN** - The functionID returned by the deployment cpp function.
   - **isDriver** - On cloud or off cloud.

### Private Members

``` java

private ArrayList<String> loadPaths = new ArrayList<>()
```

Specify the so path.

If not specified, it is specified by services.yaml.

``` java

private String functionURN = Utils.getEnvWithDefualtValue("YRFUNCID", DEFAULT_FUNC_URN, "config-")
```

The functionURN returned by the deployment function.

It can be set through the YRFUNCID environment variable. The current functionURN generation logic is ``sn:cn:yrk:default:function:0-{ServiceName}-{FunctionName}:$latest``.

``` java

private String cppFunctionURN = DEFAULT_CPP_URN
```

The functionID returned by the deployment cpp function. Default is ``sn:cn:yrk:default:function:0-defaultservice-cpp:$lates``.

``` java

private String serverAddress = ""
```

Cluster IP (openYuanrongcluster master node).

``` java

private String dataSystemAddress = ""
```

Data system IP (openYuanrong cluster master node).

``` java

private int rpcTimeout = DEFAULT_TIMEOUT
```

rpc timeout, default ``30 * 60``.

``` java

private int serverAddressPort = DEFAULT_SERVER_PORT
```

Cluster port number, default ``31222``.

``` java

private int dataSystemAddressPort = DEFAULT_DS_PORT
```

DataSystem port number, default ``31222``.

``` java

private String ns = ""
```

When a user uses a named stateful function, ``ns`` is the namespace of the stateful function.

The default value is an empty string. If ``ns`` is not an empty string, the stateful function name is ``namespace-name``.

``` java

private int threadPoolSize = 0
```

Thread pool size, used only in ParallelFor.

``` java

private boolean isDriver = true
```

Is driver or not, determines whether the API is used on the cloud or locally, default is ``true``.

``` java

private int recycleTime = DEFAULT_RECYCLETIME
```

Recycling time, default is ``10``, value must be greater than 0.

``` java

private boolean enableMetrics = false
```

Whether to enable metric collection and reporting.

The default value is ``false``.

``` java

private boolean enableMTLS = false
```

Whether to enable two-way authentication for external cloud clients, default is ``false``.

``` java

private String certificateFilePath
```

Client certificate file path.

``` java

private String privateKeyPath
```

Client private key file path.

``` java

private String verifyFilePath
```

Server certificate file path.

``` java

private String serverName
```

Server name.

``` java

private String logDir = DEFAULT_LOG_DIR
```

Input directory for client logs.

Default is the current directory. If the directory does not exist, create the directory and generate the log file in the directory.

``` java

private int maxLogSizeMb = 0
```

Maximum size of individual client log files, in MB, default value is ``0`` (if the default value is ``0``, it will eventually be set to ``40``).

When the size limit is exceeded, the log file will be rolled over and split. The system checks for rolling over every 30 seconds. A negative value will throw an exception.

``` java

private int maxLogFileNum = 0
```

The maximum number of files to be retained after client log rolling, default value ``0`` (if the default value is ``0``, it will be set to ``20``), and the oldest logs will be deleted when the number exceeds the limit, checked and deleted every 30 seconds.

``` java

private String logLevel = ""
```

Log level, with values of ``DEBUG``, ``INFO``, ``WARN``, and ``ERROR``.

If not set or set to an invalid value, it will be set to the default value ``INFO``.

``` java

private int maxTaskInstanceNum = -1
```

Limit the maximum number of instances that can be created for stateless functions.

The valid range is 1 to 65536. If not configured, the default value is ``-1``, which means no limit. If an invalid value is entered, the Init interface will throw an exception.

``` java

private int maxConcurrencyCreateNum = 100
```

Set the maximum number of concurrent stateless instance creations.

The value must be greater than 0. The default value is ``100``.

``` java

private boolean enableDisConvCallStack = true
```

If true, enables distributed collection call stack, and exceptions thrown by user code will be accurately located.

Users can use the distributed collection call stack to locate exceptions and errors that occur in the distributed call chain in a simpler way, just like in a single process.

``` java

private boolean isThreadLocal = false
```

Whether to enable multi-cluster mode.

The default is ``false``. If isThreadLocal is true, call `YR.init` in different threads and set different cluster addresses. The runtime Java SDK can connect to different clusters.

``` java

private Map<String, String> customEnvs = new HashMap<>()
```

Used to set custom environment variables for the runtime.

Currently, only ``LD_LIBRARY_PATH`` is supported.

``` java

private List<String> codePath
```

Code path.

``` java

private boolean enableDsEncrypt = false
```

If ``true``, enable data system tls authentication, else not.

``` java

private String dsPublicKeyContextPath = ""
```

The path of worker public key for data system tls authentication, if enableDsEncrypt is ``true`` and the dsPublicKeyContextPath is empty, an exception will be thrown.

``` java

private String runtimePublicKeyContextPath = ""
```

The path of client public key for data system tls authentication, if enableDsEncrypt is ``true`` and the runtimePublicKeyContextPath is empty, an exception will be thrown.

``` java

private String runtimePrivateKeyContextPath = ""
```

The path of client private key for data system tls authentication, if enableDsEncrypt is ``true`` and the runtimePrivateKeyContextPath is empty, an exception will be thrown.

#### Sample code

``` java

Config conf = new Config("sn:cn:yrk:default:function:0-${serviceName}-${}fun:$latest", "serverAddressIP", "dataSystemAddressIP", "cppFunctionURN", false);
conf.setRecycleTime(100);
```

### Interface table

| Field Name | getter interface | setter interface |
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
