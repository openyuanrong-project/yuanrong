# CppInstanceHandler exportHandler

package: `org.yuanrong.call`.

## Interface description

### public Map<String, String> exportHandler() throws YRException

Obtain instance handle information.

[CppInstanceHandler](CppInstanceHandler.md) class member method. Users can obtain handle information through this method, which can be serialized and stored in a database or other persistent tools.

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
Map<String, String> out = cppInstanceHandler.exportHandler();
// Serialize out and store in a database or other persistence tool.
```

- Returns:

    The handle information of [CppInstanceHandler](CppInstanceHandler.md) is stored in the kv key-value pair.

- Throws:

   - **YRException** - The export information failed and threw an YRException exception. For specific errors, see the error message description.
