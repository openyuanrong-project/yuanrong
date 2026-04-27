# CppInstanceClass

包名：`org.yuanrong.function`。

## public class CppInstanceClass

用于创建新的 C++ 实例类的辅助类。

### 接口说明

#### public static CppInstanceClass of(String className, String initFunctionName)

Java 函数创建 C++ 类实例。

```java

CppInstanceClass cppInstance = CppInstanceClass.of("Counter", "FactoryCreate");
```

- 参数：

   - **className** - C++ 实例的类名。
   - **initFunctionName** – C++ 类的初始化函数。

- 返回：

    CppInstanceClass 实例。
