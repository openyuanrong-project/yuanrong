# CppInstanceHandler importHandler

package: `org.yuanrong.call`.

## Interface description

### public void importHandler(Map<String, String> input) throws YRException

The member method of the [CppInstanceHandler](CppInstanceHandler.md) class allows users to import handle information, which can be obtained from a database or other persistent tools and deserialized.

```java

CppInstanceHandler cppInstanceHandler = new CppInstanceHandler();
cppInstanceHandler.importHandler(in);
// The input parameter is obtained by retrieving and deserializing from a database or other persistent storage.
```

- Parameters：

   - **input** - The handle information of [CppInstanceHandler](CppInstanceHandler.md) is stored in the kv key-value pair.

- Throws：

   - **YRException** - The import information failed and threw an YRException exception. For specific errors, see the error message description.
