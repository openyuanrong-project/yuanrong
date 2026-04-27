# init

## YR.init()

包名：`org.yuanrong.api`。

### 接口说明

#### public static ClientInfo init() throws YRException

openYuanrong 初始化接口。

- 抛出:

   - **YRException** - openYuanrong 集群初始化失败，网络连接异常等。

- 返回:

    [ClientInfo](#public-class-clientinfo): openYuanrong 客户端信息。

#### public static ClientInfo init(Config conf) throws YRException

openYuanrong 初始化接口，用于配置参数，参数说明详见数据结构 [Config](Config.md)。

```java

Config conf = new Config("sn:cn:yrk:default:function:0-${serviceName}-${}fun:$latest", "serverAddressIP", "dataSystemAddressIP", "cppFunctionURN");
ClientInfo jobid = YR.init(conf);
```

- 参数：

   - **conf** - openYuanrong 初始化参数配置。

- 抛出：

   - **YRException** - openYuanrong 集群初始化失败，配置错误或网络连接异常等。

- 返回：

    [ClientInfo](#public-class-clientinfo): openYuanrong 客户端信息。

## public class ClientInfo

包名：`org.yuanrong.api`。

用于存储 openYuanrong 客户端信息。

### Private Members

``` java

private String jobID
```

init() 接口返回, 用于跟踪和管理当前 job。

### 接口表格

| 字段名称 | getter 接口 | setter 接口 |
| -------- | ----------- | ----------- |
| jobID | getJobID| setJobID |
