# CppInstanceMethod

包名：`org.yuanrong.function`。

## public class CppInstanceMethod&lt;R\>

用于调用成员函数的辅助类。

### 接口说明

#### public static CppInstanceMethod\<Object\> of(String methodName)

Java 函数调用 C++ 类实例成员函数。

- 参数：

   - **methodName** - methodName C++实例方法名称。

- 返回：

    CppInstanceMethod 类型的实例对象。

#### public static &lt;R&gt; CppInstanceMethod&lt;R&gt; of(String methodName, Class&lt;R&gt; returnType)

Java 函数调用 C++ 类实例成员函数。

```java

ObjectRef<int> res = cppInstance.function(CppInstanceMethod.of("Add", int.class)).invoke(1);
```

- 参数：

   - **methodName** - C++实例方法名称。
   - **returnType** - C++类实例方法的返回值类型。
   - **&lt;R&gt;** - 实例方法的返回值类型。

- 返回：

    CppInstanceMethod 实例。
