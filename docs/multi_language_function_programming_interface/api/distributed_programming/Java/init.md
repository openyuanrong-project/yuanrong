# init

## YR.init()

package: `org.yuanrong.api`.

### public static ClientInfo init() throws YRException

The openYuanrong initialization interface.

- Throws:

   - **YRException** - The openYuanrong cluster initialization failed due to network connection issues.

- Returns:

    [ClientInfo](#public-class-clientinfo): openYuanrong client info.

### public static ClientInfo init(Config conf) throws YRException

The openYuanrong initialization interface is used to configure parameters.

For parameter descriptions, see the data structure [Config](Config.md).

```java

Config conf = new Config("sn:cn:yrk:default:function:0-${serviceName}-${}fun:$latest", "serverAddressIP", "dataSystemAddressIP", "cppFunctionURN");
ClientInfo jobid = YR.init(conf);
```

- Parameters:

   - **conf** - Initialization parameter configuration of Yuanrong.

- Throws:

   - **YRException** - The openYuanrong cluster initialization failed due to configuration errors or network connection issues.

- Returns:

    [ClientInfo](#public-class-clientinfo): openYuanrong client info.

## public class ClientInfo

package: `org.yuanrong.api`.

Use to storage openYuanrong client info.

### Private Members

``` java

private String jobID
```

Returned through the init() interface, used for tracking and managing the current job.

### Interface table

| Field Name | getter interface | setter interface |
| -------- | ----------- | ----------- |
| jobID | getJobID| setJobID |
