# JavaInstanceHandler importHandler

包名：`org.yuanrong.call`。

## 接口说明

### public void importHandler(Map<String, String> input) throws YRException

[JavaInstanceHandler](JavaInstanceHandler.md) 类的成员方法，用户通过该方法可导入句柄信息，可以从数据库等持久化工具获取并反序列化导入信息。

```java

JavaInstanceHandler javaInstanceHandler = new JavaInstanceHandler();
javaInstanceHandler.importHandler(in);
// 输入参数是通过从数据库或其他持久化存储中检索并反序列化得到。
```

- 参数：

   - **input** - 通过 kv 键值对存放的 [JavaInstanceHandler](JavaInstanceHandler.md) 的句柄信息。

- 抛出：

   - **YRException** - 统一抛出的异常类型。
