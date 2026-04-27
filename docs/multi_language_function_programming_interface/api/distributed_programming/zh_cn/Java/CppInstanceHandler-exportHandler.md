# CppInstanceHandler exportHandler

包名：`org.yuanrong.call`。

## 接口说明

### public Map<String, String> exportHandler() throws YRException

获取实例句柄信息。

[CppInstanceHandler](CppInstanceHandler.md)类的成员方法，用户通过该方法可获取句柄信息，可以将其序列化后存入数据库等持久化工具。

```java

CppInstanceHandler cppInstanceHandler = YR.instance(CppInstanceClass.of("Counter", "FactoryCreate"))
    .setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest")
    .invoke(1);
Map<String, String> out = cppInstanceHandler.exportHandler();
// 序列化后存储到数据库或其他持久化工具中。
```

- 返回：

    通过 kv 键值对存放的 [CppInstanceHandler](CppInstanceHandler.md) 的句柄信息

- 抛出：

   - **YRException** - 导出信息失败抛出 YRException 异常，具体错误详见错误信息描述。
