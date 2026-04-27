# CppInstanceHandler importHandler

包名：`org.yuanrong.call`。

## 接口说明

### public void importHandler(Map<String, String> input) throws YRException

[CppInstanceHandler](CppInstanceHandler.md) 类的成员方法，用户通过该方法可导入句柄信息，可以从数据库等持久化工具获取并反序列化导入信息。

```java

CppInstanceHandler cppInstanceHandler = new CppInstanceHandler();
cppInstanceHandler.importHandler(in);
// 输入参数是通过从数据库或其他持久化存储中检索和反序列化获得。
```

- 参数：

   - **input** - 通过 kv 键值对存放的 CppInstanceHandler 的句柄信息。

- 抛出：

   - **YRException** - 导入信息失败抛出该异常，具体错误详见错误信息描述。
