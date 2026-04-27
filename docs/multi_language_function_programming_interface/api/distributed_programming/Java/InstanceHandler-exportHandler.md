# InstanceHandler exportHandler

package: `org.yuanrong.call`.

## Interface description

### public Map<String, String> exportHandler() throws YRException

The member method of the [InstanceHandler](InstanceHandler.md) class allows users to obtain handle information, which can be serialized and stored in a database or other persistence tools.

When the tenant context is enabled, handle information can also be obtained through this method and used across tenants.

```java

InstanceHandler instHandler = YR.instance(MyYRApp::new).invoke();
Map<String, String> out = instHandler.exportHandler();
// Serialize out and store in a database or other persistence tool.
```

- Returns:

    Map<String, String> type, The handle information of the [InstanceHandler](InstanceHandler.md) is stored in the kv key-value pair.

- Throws:

   - **YRException** - Unified exception types thrown.
