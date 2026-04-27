# InstanceHandler importHandler

包名：`org.yuanrong.call`。

## 接口说明

### public void importHandler(Map<String, String> input) throws YRException

[InstanceHandler](InstanceHandler.md) 类的成员方法，用户通过该方法可导入句柄信息，可以从数据库等持久化工具获取并反序列化导入信息。

```java

InstanceHandler instanceHandler = new InstanceHandler();
instanceHandler.importHandler(in);
// The input parameter is obtained by retrieving and deserializing from a database or other persistent storage.
```

- 参数：

   - **input** - 通过 kv 键值对存放的 [InstanceHandler](InstanceHandler.md) 的句柄信息。

- 抛出：

   - **YRException** - 统一抛出的异常类型。
