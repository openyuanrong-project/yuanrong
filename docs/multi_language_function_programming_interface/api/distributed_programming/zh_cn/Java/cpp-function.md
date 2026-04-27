# CppFunction

包名：`org.yuanrong.function`。

## public class CppFunction&lt;R\>

Java 函数调用 c++ 函数。

### 接口说明

#### public static CppFunction\<Object\> of(String functionName)

Java 函数调用 c++ 函数。

- 参数：

   - **functionName** - C++ 方法名称。

- 返回：

    CppFunction 实例。

#### public static &lt;R&gt; CppFunction&lt;R&gt; of(String functionName, Class&lt;R&gt; returnType)

Java 函数调用 c++ 函数。

```java

ObjectRef res = YR.function(CppFunction.of("PlusOne", int.class)).invoke(1);
```

- 参数：

   - **&lt;R&gt;** - 返回值类型。
   - **functionName** - C++ 方法名称。
   - **returnType** - C++ 方法的返回值类型。

- 返回：

    CppFunction 实例。
