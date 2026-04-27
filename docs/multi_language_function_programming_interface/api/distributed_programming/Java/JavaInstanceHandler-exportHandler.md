# JavaInstanceHandler exportHandler

package: `org.yuanrong.call`.

## Interface description

### public Map<String, String> exportHandler() throws YRException

The member method of the [JavaInstanceHandler](JavaInstanceHandler.md) class allows users to obtain handle information, which can be serialized and stored in a database or other persistence tools.

When the tenant context is enabled, handle information can also be obtained through this method and used across tenants.

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
Map<String, String> out = javaInstanceHandler.exportHandler();
// Serialize out and store in a database or other persistence tool.
```

- Returns:

    Map<String, String> type, storing the handle information of JavaInstanceHandler through kv key-value pairs.

- Throws:

   - **YRException** - Unified exception types thrown.
