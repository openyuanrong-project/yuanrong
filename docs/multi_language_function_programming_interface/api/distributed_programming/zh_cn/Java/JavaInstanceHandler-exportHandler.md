# JavaInstanceHandler exportHandler

包名：`org.yuanrong.call`。

## 接口说明

### public Map<String, String> exportHandler() throws YRException

[JavaInstanceHandler](JavaInstanceHandler.md) 类的成员方法，用户通过该方法可获取句柄信息，可以将其序列化后存入数据库等持久化工具。

```java

JavaInstanceHandler javaInstanceHandler = YR.instance(JavaInstanceClass.of("com.example.YrlibHandler$MyYRApp")).setUrn("sn:cn:yrk:default:function:0-opc-opc:$latest").invoke();
Map<String, String> out = javaInstanceHandler.exportHandler();
// 序列化后存储到数据库或其他持久化工具中。
```

- 返回：

    Map<String, String> 类型，通过 kv 键值对存放的 JavaInstanceHandler 的句柄信息。

- 抛出：

   - **YRException** - 统一抛出的异常类型。
