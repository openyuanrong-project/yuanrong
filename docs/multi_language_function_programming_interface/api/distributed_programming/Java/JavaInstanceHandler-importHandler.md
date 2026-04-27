# JavaInstanceHandler importHandler

package: `org.yuanrong.call`.

## Interface description

### public void importHandler(Map<String, String> input) throws YRException

The member method of the [JavaInstanceHandler](JavaInstanceHandler.md) class allows users to import handle information, which can be obtained and deserialized from persistent tools such as databases.

When the tenant context is enabled, this method can also be used to import handle information from other tenant contexts.

```java

JavaInstanceHandler javaInstanceHandler = new JavaInstanceHandler();
javaInstanceHandler.importHandler(in);
// The input parameter is obtained by retrieving and deserializing from a database or other persistent storage.
```

- Parameters:

   - **input** - Store the handle information of [JavaInstanceHandler](JavaInstanceHandler.md) through the kv key-value pair.

- Throws:

   - **YRException** - Unified exception types thrown.
